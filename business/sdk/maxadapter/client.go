package maxadapter

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	maxadapterv1 "github.com/zabolotny-dev/clicksafe/adapters/max/gen/go/maxadapter/v1"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn  *grpc.ClientConn
	api   maxadapterv1.MaxAdapterServiceClient
	token string
}

func New(cfg Config) (*Client, error) {
	if cfg.Addr == "" {
		return nil, errors.New("max adapter addr is required")
	}

	conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("new max adapter client: %w", err)
	}

	return &Client{
		conn:  conn,
		api:   maxadapterv1.NewMaxAdapterServiceClient(conn),
		token: cfg.Token,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) BeginLogin(ctx context.Context, phone string) (maxaccountbus.LoginAttempt, error) {
	resp, err := c.api.BeginLogin(c.withToken(ctx), &maxadapterv1.BeginLoginRequest{
		Phone: phone,
	})
	if err != nil {
		return maxaccountbus.LoginAttempt{}, mapErr(err)
	}
	return toLoginAttempt(resp.GetAttempt()), nil
}

func (c *Client) ConfirmLogin(ctx context.Context, attemptID, code, password string) (maxaccountbus.LoginResult, error) {
	resp, err := c.api.ConfirmLogin(c.withToken(ctx), &maxadapterv1.ConfirmLoginRequest{
		LoginAttemptId: attemptID,
		Code:           code,
		Password:       password,
	})
	if err != nil {
		return maxaccountbus.LoginResult{}, mapErr(err)
	}
	return toLoginResult(resp)
}

func (c *Client) ConfirmPassword(ctx context.Context, attemptID, password string) (maxaccountbus.LoginResult, error) {
	resp, err := c.api.ConfirmPassword(c.withToken(ctx), &maxadapterv1.ConfirmPasswordRequest{
		LoginAttemptId: attemptID,
		Password:       password,
	})
	if err != nil {
		return maxaccountbus.LoginResult{}, mapErr(err)
	}
	return toLoginResult(resp)
}

func (c *Client) ListAccounts(ctx context.Context) ([]maxaccountbus.Account, error) {
	resp, err := c.api.ListAccounts(c.withToken(ctx), &maxadapterv1.ListAccountsRequest{})
	if err != nil {
		return nil, mapErr(err)
	}

	accounts := make([]maxaccountbus.Account, 0, len(resp.GetAccounts()))
	for _, account := range resp.GetAccounts() {
		converted, err := toAccount(account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, converted)
	}
	return accounts, nil
}

func (c *Client) StartAccount(ctx context.Context, adapterID uuid.UUID) (maxaccountbus.Account, error) {
	resp, err := c.api.StartAccount(c.withToken(ctx), &maxadapterv1.StartAccountRequest{
		AccountId: adapterID.String(),
	})
	if err != nil {
		return maxaccountbus.Account{}, mapErr(err)
	}
	return toAccount(resp)
}

func (c *Client) StopAccount(ctx context.Context, adapterID uuid.UUID) (maxaccountbus.Account, error) {
	resp, err := c.api.StopAccount(c.withToken(ctx), &maxadapterv1.StopAccountRequest{
		AccountId: adapterID.String(),
	})
	if err != nil {
		return maxaccountbus.Account{}, mapErr(err)
	}
	return toAccount(resp)
}

func (c *Client) DeleteAccount(ctx context.Context, adapterID uuid.UUID) error {
	_, err := c.api.DeleteAccount(c.withToken(ctx), &maxadapterv1.DeleteAccountRequest{
		AccountId: adapterID.String(),
	})
	return mapErr(err)
}

func (c *Client) SendMessage(ctx context.Context, req SendMessageRequest) (SendMessageResponse, error) {
	if req.AccountID == uuid.Nil {
		return SendMessageResponse{}, errors.New("account id is required")
	}
	if req.ClientRequestID == "" {
		return SendMessageResponse{}, errors.New("client request id is required")
	}
	if req.Recipient.Value == "" {
		return SendMessageResponse{}, errors.New("recipient is required")
	}

	descriptors := make([]*maxadapterv1.AttachmentDescriptor, len(req.Attachments))
	for i, attachment := range req.Attachments {
		if attachment.Reader == nil {
			return SendMessageResponse{}, fmt.Errorf("attachment[%d] reader is required", i)
		}
		descriptors[i] = &maxadapterv1.AttachmentDescriptor{
			Filename:  attachment.Filename,
			MimeType:  attachment.MimeType,
			Extension: attachment.Extension,
			Size:      attachment.Size,
			Kind:      attachmentKindProto(attachment.Kind),
		}
	}

	stream, err := c.api.SendMessage(c.withToken(ctx))
	if err != nil {
		return SendMessageResponse{}, mapErr(err)
	}

	err = stream.Send(&maxadapterv1.SendMessageRequest{
		Payload: &maxadapterv1.SendMessageRequest_Metadata{
			Metadata: &maxadapterv1.SendMessageMetadata{
				AccountId:        req.AccountID.String(),
				Recipient:        recipientProto(req.Recipient),
				Text:             req.Text,
				ClientRequestId:  req.ClientRequestID,
				Attachments:      descriptors,
				Notify:           req.Notify,
				ReplyToMessageId: req.ReplyToMessageID,
			},
		},
	})
	if err != nil {
		return SendMessageResponse{}, mapErr(err)
	}

	buf := make([]byte, 64*1024)
	for index, attachment := range req.Attachments {
		for {
			n, readErr := attachment.Reader.Read(buf)
			if n > 0 {
				chunk := append([]byte(nil), buf[:n]...)
				err := stream.Send(&maxadapterv1.SendMessageRequest{
					Payload: &maxadapterv1.SendMessageRequest_Chunk{
						Chunk: &maxadapterv1.AttachmentChunk{
							AttachmentIndex: int32(index),
							Data:            chunk,
						},
					},
				})
				if err != nil {
					return SendMessageResponse{}, mapErr(err)
				}
			}
			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				return SendMessageResponse{}, fmt.Errorf("read attachment[%d]: %w", index, readErr)
			}
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return SendMessageResponse{}, mapErr(err)
	}

	accountID, err := uuid.Parse(resp.GetAccountId())
	if err != nil {
		return SendMessageResponse{}, fmt.Errorf("parse sent account id: %w", err)
	}

	return SendMessageResponse{
		ClientRequestID: resp.GetClientRequestId(),
		AccountID:       accountID,
		ChatID:          resp.GetChatId(),
		MessageID:       resp.GetMessageId(),
		SentAt:          resp.GetSentAt(),
	}, nil
}

func (c *Client) SubscribeEvents(ctx context.Context, consumer string, fromSeq int64, handle func(AdapterEvent) error) error {
	if consumer == "" {
		return errors.New("consumer is required")
	}
	if handle == nil {
		return errors.New("event handler is required")
	}

	stream, err := c.api.SubscribeEvents(c.withToken(ctx), &maxadapterv1.SubscribeEventsRequest{
		Consumer: consumer,
		FromSeq:  fromSeq,
	})
	if err != nil {
		return mapErr(err)
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			return mapErr(err)
		}

		converted, err := toAdapterEvent(event)
		if err != nil {
			return err
		}

		if err := handle(converted); err != nil {
			return err
		}
	}
}

func (c *Client) AckEvents(ctx context.Context, consumer string, upToSeq int64) (int64, error) {
	if consumer == "" {
		return 0, errors.New("consumer is required")
	}

	resp, err := c.api.AckEvents(c.withToken(ctx), &maxadapterv1.AckEventsRequest{
		Consumer: consumer,
		UpToSeq:  upToSeq,
	})
	if err != nil {
		return 0, mapErr(err)
	}

	return resp.GetAcknowledgedSeq(), nil
}

func (c *Client) withToken(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}
