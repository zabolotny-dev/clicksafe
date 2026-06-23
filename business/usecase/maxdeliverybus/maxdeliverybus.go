package maxdeliverybus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/maxadapter"
)

var ErrDeliveryNotFound = errors.New("max delivery not found")

const consumerName = "clicksafe-api"

type targetQuerier interface {
	QueryDue(ctx context.Context) ([]campaignbus.Target, error)
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Target, error)
	ChangeStatus(ctx context.Context, t campaignbus.Target, s campaignbus.TargetStatus) error
}

type campaignQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error)
}

type employeeQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error)
}

type messageQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (messagebus.Message, error)
}

type maxAccountQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (maxaccountbus.Account, error)
}

type attachmentProvider interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type renderProvider interface {
	Render(ctx context.Context, atch attachmentbus.Attachment, targetID uuid.UUID) ([]byte, error)
}

type eventPublisher interface {
	Publish(ctx context.Context, e eventbus.NewEvent) error
}

type Delivery struct {
	ID                uuid.UUID
	TargetID          uuid.UUID
	CampaignID        uuid.UUID
	EmployeeID        uuid.UUID
	MaxAccountID      uuid.UUID
	AdapterAccountID  uuid.UUID
	ChatID            string
	MessageID         string
	SentAt            time.Time
	ReadAt            *time.Time
	RepliedAt         *time.Time
	EducationSentAt   *time.Time
	IncomingMessageID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type store interface {
	SaveSent(ctx context.Context, delivery Delivery) error
	QueryByID(ctx context.Context, id uuid.UUID) (Delivery, error)
	QueryByMessage(ctx context.Context, accountID uuid.UUID, chatID string, messageID string) (Delivery, error)
	QueryReply(ctx context.Context, accountID uuid.UUID, chatID string, replyToMessageID string) (Delivery, error)
	QueryLatestUnreadByChat(ctx context.Context, accountID uuid.UUID, chatID string) (Delivery, error)
	MarkRead(ctx context.Context, id uuid.UUID, at time.Time) (bool, error)
	MarkReplied(ctx context.Context, id uuid.UUID, incomingMessageID string, at time.Time) (bool, error)
	MarkEducationSent(ctx context.Context, id uuid.UUID, at time.Time) error
	IsProcessed(ctx context.Context, seq int64) (bool, error)
	MarkProcessed(ctx context.Context, seq int64) error
}

type adapter interface {
	SendMessage(ctx context.Context, req maxadapter.SendMessageRequest) (maxadapter.SendMessageResponse, error)
	SubscribeEvents(ctx context.Context, consumer string, fromSeq int64, handle func(maxadapter.AdapterEvent) error) error
	AckEvents(ctx context.Context, consumer string, upToSeq int64) (int64, error)
}

type Business struct {
	targets     targetQuerier
	campaigns   campaignQuerier
	employees   employeeQuerier
	messages    messageQuerier
	accounts    maxAccountQuerier
	attachments attachmentProvider
	renderer    renderProvider
	events      eventPublisher
	adapter     adapter
	store       store
	runInTx     database.TxRunner
}

func NewBusiness(targets targetQuerier, campaigns campaignQuerier, employees employeeQuerier, messages messageQuerier, accounts maxAccountQuerier, attachments attachmentProvider, renderer renderProvider, events eventPublisher, adapter adapter, store store, runInTx database.TxRunner) *Business {
	return &Business{
		targets:     targets,
		campaigns:   campaigns,
		employees:   employees,
		messages:    messages,
		accounts:    accounts,
		attachments: attachments,
		renderer:    renderer,
		events:      events,
		adapter:     adapter,
		store:       store,
		runInTx:     runInTx,
	}
}

func (b *Business) SendDue(ctx context.Context) []error {
	var errs []error

	targets, err := b.targets.QueryDue(ctx)
	if err != nil {
		return []error{fmt.Errorf("sendmax: %w", err)}
	}

	for _, target := range targets {
		campaign, err := b.campaigns.QueryByID(ctx, target.CampaignID)
		if err != nil {
			errs = append(errs, fmt.Errorf("sendmax: target[%s]: query campaign: %w", target.ID, err))
			continue
		}
		if campaign.Type != campaignbus.MaxCampaign {
			continue
		}

		if err := b.processTarget(ctx, target, campaign); err != nil {
			errs = append(errs, fmt.Errorf("sendmax: target[%s]: %w", target.ID, err))
			_ = b.targets.ChangeStatus(ctx, target, campaignbus.Failed)
			_ = b.events.Publish(ctx, eventbus.NewEvent{
				CampaignID: target.CampaignID,
				EmployeeID: target.EmployeeID,
				Type:       eventbus.DeliveryFailed,
			})
		}
	}

	return errs
}

func (b *Business) processTarget(ctx context.Context, target campaignbus.Target, campaign campaignbus.Campaign) error {
	if campaign.MessageID == nil {
		return campaignbus.ErrMessageRequired
	}

	message, err := b.messages.QueryByID(ctx, *campaign.MessageID)
	if err != nil {
		return fmt.Errorf("query message: %w", err)
	}
	if message.Type != messagebus.MaxMessage {
		return campaignbus.ErrMessageTypeMismatch
	}

	account, err := b.accounts.QueryByID(ctx, message.MaxAccountID.UUID)
	if err != nil {
		return fmt.Errorf("query max account: %w", err)
	}

	employee, err := b.employees.QueryByID(ctx, target.EmployeeID)
	if err != nil {
		return fmt.Errorf("query employee: %w", err)
	}
	if employee.Phone.String() == "" {
		return campaignbus.ErrTargetPhoneRequired
	}

	text, err := b.renderText(ctx, message.TextBodyID.UUID, target.ID)
	if err != nil {
		return err
	}

	attachments, err := b.renderAttachments(ctx, message.AttachmentIDs, target.ID)
	if err != nil {
		return err
	}

	response, err := b.adapter.SendMessage(ctx, maxadapter.SendMessageRequest{
		AccountID: account.AdapterID,
		Recipient: maxadapter.MessageRecipient{
			Kind:  maxadapter.RecipientKindPhone,
			Value: employee.Phone.String(),
		},
		Text:            text,
		ClientRequestID: "max-target-" + target.ID.String(),
		Attachments:     attachments,
		Notify:          true,
	})
	if err != nil {
		return fmt.Errorf("send adapter message: %w", err)
	}

	now := time.Now().UTC()
	delivery := Delivery{
		ID:               uuid.New(),
		TargetID:         target.ID,
		CampaignID:       target.CampaignID,
		EmployeeID:       target.EmployeeID,
		MaxAccountID:     account.ID,
		AdapterAccountID: account.AdapterID,
		ChatID:           response.ChatID,
		MessageID:        response.MessageID,
		SentAt:           now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	return b.runInTx(ctx, func(ctx context.Context) error {
		if err := b.store.SaveSent(ctx, delivery); err != nil {
			return fmt.Errorf("save delivery: %w", err)
		}
		if err := b.targets.ChangeStatus(ctx, target, campaignbus.Sent); err != nil {
			return fmt.Errorf("change target status: %w", err)
		}
		return b.events.Publish(ctx, eventbus.NewEvent{
			CampaignID: target.CampaignID,
			EmployeeID: target.EmployeeID,
			Type:       eventbus.MessageSent,
		})
	})
}

func (b *Business) ConsumeEvents(ctx context.Context) error {
	return b.adapter.SubscribeEvents(ctx, consumerName, 0, func(event maxadapter.AdapterEvent) error {
		processed, err := b.store.IsProcessed(ctx, event.Seq)
		if err != nil {
			return err
		}
		if !processed {
			if err := b.processAdapterEvent(ctx, event); err != nil {
				return err
			}
			if err := b.store.MarkProcessed(ctx, event.Seq); err != nil {
				return err
			}
		}

		_, err = b.adapter.AckEvents(ctx, consumerName, event.Seq)
		return err
	})
}

func (b *Business) processAdapterEvent(ctx context.Context, event maxadapter.AdapterEvent) error {
	switch event.Type {
	case maxadapter.AdapterEventMessageRead:
		return b.processRead(ctx, event)
	case maxadapter.AdapterEventMessageReplied, maxadapter.AdapterEventMessageReceived:
		return b.processReply(ctx, event)
	default:
		return nil
	}
}

func (b *Business) processRead(ctx context.Context, event maxadapter.AdapterEvent) error {
	delivery, err := b.findReadDelivery(ctx, event)
	if err != nil {
		if errors.Is(err, ErrDeliveryNotFound) {
			return nil
		}
		return err
	}

	return b.runInTx(ctx, func(ctx context.Context) error {
		changed, err := b.store.MarkRead(ctx, delivery.ID, time.Now().UTC())
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}

		target, err := b.targets.QueryByID(ctx, delivery.TargetID)
		if err != nil {
			return err
		}
		if target.Status == campaignbus.Sent {
			if err := b.targets.ChangeStatus(ctx, target, campaignbus.Opened); err != nil {
				return err
			}
		}

		return b.events.Publish(ctx, eventbus.NewEvent{
			CampaignID: delivery.CampaignID,
			EmployeeID: delivery.EmployeeID,
			Type:       eventbus.MessageRead,
		})
	})
}

func (b *Business) findReadDelivery(ctx context.Context, event maxadapter.AdapterEvent) (Delivery, error) {
	if event.MessageID != "" {
		delivery, err := b.store.QueryByMessage(ctx, event.AccountID, event.ChatID, event.MessageID)
		if err == nil || !errors.Is(err, ErrDeliveryNotFound) {
			return delivery, err
		}
	}

	if event.ChatID == "" {
		return Delivery{}, ErrDeliveryNotFound
	}
	return b.store.QueryLatestUnreadByChat(ctx, event.AccountID, event.ChatID)
}

func (b *Business) processReply(ctx context.Context, event maxadapter.AdapterEvent) error {
	delivery, err := b.findReplyDelivery(ctx, event)
	if err != nil {
		if errors.Is(err, ErrDeliveryNotFound) {
			return nil
		}
		return err
	}

	if err := b.runInTx(ctx, func(ctx context.Context) error {
		now := time.Now().UTC()
		changed, err := b.store.MarkReplied(ctx, delivery.ID, event.MessageID, now)
		if err != nil {
			return err
		}

		if !changed {
			return nil
		}

		target, err := b.targets.QueryByID(ctx, delivery.TargetID)
		if err != nil {
			return err
		}
		if target.Status == campaignbus.Sent || target.Status == campaignbus.Opened || target.Status == campaignbus.Clicked {
			if err := b.targets.ChangeStatus(ctx, target, campaignbus.Replied); err != nil {
				return err
			}
		}

		return b.events.Publish(ctx, eventbus.NewEvent{
			CampaignID: delivery.CampaignID,
			EmployeeID: delivery.EmployeeID,
			Type:       eventbus.MessageReplied,
		})
	}); err != nil {
		return err
	}

	updated, err := b.store.QueryByID(ctx, delivery.ID)
	if err != nil {
		return err
	}
	if updated.EducationSentAt != nil {
		return nil
	}

	return b.sendEducation(ctx, updated, event.MessageID)
}

func (b *Business) findReplyDelivery(ctx context.Context, event maxadapter.AdapterEvent) (Delivery, error) {
	return b.store.QueryReply(ctx, event.AccountID, event.ChatID, event.ReplyToMessageID)
}

func (b *Business) sendEducation(ctx context.Context, delivery Delivery, replyToMessageID string) error {
	campaign, err := b.campaigns.QueryByID(ctx, delivery.CampaignID)
	if err != nil {
		return err
	}
	if !campaign.MaxEducationTextID.Valid {
		return campaignbus.ErrMaxEducationRequired
	}

	text, err := b.renderText(ctx, campaign.MaxEducationTextID.UUID, delivery.TargetID)
	if err != nil {
		return err
	}

	_, err = b.adapter.SendMessage(ctx, maxadapter.SendMessageRequest{
		AccountID: delivery.AdapterAccountID,
		Recipient: maxadapter.MessageRecipient{
			Kind:  maxadapter.RecipientKindChatID,
			Value: delivery.ChatID,
		},
		Text:             text,
		ClientRequestID:  "max-education-" + delivery.TargetID.String(),
		Notify:           true,
		ReplyToMessageID: replyToMessageID,
	})
	if err != nil {
		return err
	}

	return b.store.MarkEducationSent(ctx, delivery.ID, time.Now().UTC())
}

func (b *Business) renderText(ctx context.Context, attachmentID uuid.UUID, targetID uuid.UUID) (string, error) {
	attachment, err := b.attachments.QueryByID(ctx, attachmentID)
	if err != nil {
		return "", fmt.Errorf("query text body: %w", err)
	}
	if attachment.Type != attachmentbus.Txt {
		return "", messagebus.ErrInvalidAttachment
	}

	rendered, err := b.renderer.Render(ctx, attachment, targetID)
	if err != nil {
		return "", fmt.Errorf("render text body: %w", err)
	}

	return string(rendered), nil
}

func (b *Business) renderAttachments(ctx context.Context, ids []uuid.UUID, targetID uuid.UUID) ([]maxadapter.MessageAttachment, error) {
	attachments := make([]maxadapter.MessageAttachment, 0, len(ids))
	for _, id := range ids {
		attachment, err := b.attachments.QueryByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("query attachment: %w", err)
		}

		rendered, err := b.renderer.Render(ctx, attachment, targetID)
		if err != nil {
			return nil, fmt.Errorf("render attachment: %w", err)
		}

		ext := attachment.Type.String()
		filename := attachment.Label.String() + ext

		attachments = append(attachments, maxadapter.MessageAttachment{
			Filename:  filepath.Base(filename),
			MimeType:  attachment.Type.MIMEType(),
			Extension: strings.TrimPrefix(ext, "."),
			Size:      int64(len(rendered)),
			Kind:      attachmentKind(attachment.Type),
			Reader:    bytes.NewReader(rendered),
		})
	}

	return attachments, nil
}

func attachmentKind(attachmentType attachmentbus.AttachmentType) maxadapter.AttachmentKind {
	if attachmentType.IsImage() {
		return maxadapter.AttachmentKindPhoto
	}
	if attachmentType.IsAudio() {
		return maxadapter.AttachmentKindAudio
	}
	return maxadapter.AttachmentKindFile
}
