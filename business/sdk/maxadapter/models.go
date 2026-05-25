package maxadapter

import (
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	maxadapterv1 "github.com/zabolotny-dev/clicksafe/adapters/max/gen/go/maxadapter/v1"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	Addr  string
	Token string
}

type RecipientKind string

const (
	RecipientKindChatID RecipientKind = "CHAT_ID"
	RecipientKindPhone  RecipientKind = "PHONE"
	RecipientKindUserID RecipientKind = "USER_ID"
)

type AttachmentKind string

const (
	AttachmentKindPhoto AttachmentKind = "PHOTO"
	AttachmentKindFile  AttachmentKind = "FILE"
	AttachmentKindVideo AttachmentKind = "VIDEO"
	AttachmentKindAudio AttachmentKind = "AUDIO"
)

type MessageRecipient struct {
	Kind  RecipientKind
	Value string
}

type MessageAttachment struct {
	Filename  string
	MimeType  string
	Extension string
	Size      int64
	Kind      AttachmentKind
	Reader    io.Reader
}

type SendMessageRequest struct {
	AccountID        uuid.UUID
	Recipient        MessageRecipient
	Text             string
	ClientRequestID  string
	Attachments      []MessageAttachment
	Notify           bool
	ReplyToMessageID string
}

type SendMessageResponse struct {
	ClientRequestID string
	AccountID       uuid.UUID
	ChatID          string
	MessageID       string
	SentAt          string
}

type AdapterEventType string

const (
	AdapterEventAccountConnected    AdapterEventType = "ACCOUNT_CONNECTED"
	AdapterEventAccountDisconnected AdapterEventType = "ACCOUNT_DISCONNECTED"
	AdapterEventMessageReceived     AdapterEventType = "MESSAGE_RECEIVED"
	AdapterEventMessageRead         AdapterEventType = "MESSAGE_READ"
	AdapterEventMessageReplied      AdapterEventType = "MESSAGE_REPLIED"
	AdapterEventRaw                 AdapterEventType = "RAW"
)

type AdapterEvent struct {
	Seq                int64
	Type               AdapterEventType
	AccountID          uuid.UUID
	ChatID             string
	MessageID          string
	SenderID           string
	Text               string
	ReplyToMessageID   string
	PayloadJSON        string
	OccurredAt         string
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if ok && st.Code() == codes.NotFound {
		return maxaccountbus.ErrAccountNotFound
	}
	return fmt.Errorf("%w: %v", maxaccountbus.ErrAdapterFailed, err)
}

func toLoginResult(resp *maxadapterv1.ConfirmLoginResponse) (maxaccountbus.LoginResult, error) {
	result := maxaccountbus.LoginResult{
		Attempt: toLoginAttempt(resp.GetAttempt()),
	}

	if resp.GetAccount() != nil && resp.GetAccount().GetId() != "" {
		account, err := toAccount(resp.GetAccount())
		if err != nil {
			return maxaccountbus.LoginResult{}, err
		}
		result.Account = &account
	}
	return result, nil
}

func toLoginAttempt(attempt *maxadapterv1.LoginAttempt) maxaccountbus.LoginAttempt {
	if attempt == nil {
		return maxaccountbus.LoginAttempt{}
	}
	return maxaccountbus.LoginAttempt{
		ID:        attempt.GetId(),
		Phone:     attempt.GetPhone(),
		Label:     attempt.GetLabel(),
		Status:    loginStatus(attempt.GetStatus()),
		Error:     attempt.GetError(),
		ExpiresAt: attempt.GetExpiresAt(),
	}
}

func toAccount(account *maxadapterv1.Account) (maxaccountbus.Account, error) {
	if account == nil {
		return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
	}

	adapterID, err := uuid.Parse(account.GetId())
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("parse adapter account id: %w", err)
	}

	status, err := maxaccountbus.ParseStatus(accountStatus(account.GetStatus()))
	if err != nil {
		return maxaccountbus.Account{}, err
	}

	return maxaccountbus.Account{
		AdapterID: adapterID,
		Phone:     account.GetPhone(),
		Label:     account.GetLabel(),
		Status:    status,
		MaxUserID: account.GetMaxUserId(),
		LastError: account.GetLastError(),
		CreatedAt: parseTime(account.GetCreatedAt()),
		UpdatedAt: parseTime(account.GetUpdatedAt()),
	}, nil
}

func accountStatus(status maxadapterv1.AccountStatus) string {
	switch status {
	case maxadapterv1.AccountStatus_ACCOUNT_STATUS_PENDING_LOGIN:
		return "PENDING_LOGIN"
	case maxadapterv1.AccountStatus_ACCOUNT_STATUS_ACTIVE:
		return "ACTIVE"
	case maxadapterv1.AccountStatus_ACCOUNT_STATUS_CONNECTED:
		return "CONNECTED"
	case maxadapterv1.AccountStatus_ACCOUNT_STATUS_DISCONNECTED:
		return "DISCONNECTED"
	case maxadapterv1.AccountStatus_ACCOUNT_STATUS_ERROR:
		return "ERROR"
	default:
		return "ERROR"
	}
}

func loginStatus(status maxadapterv1.LoginAttemptStatus) string {
	switch status {
	case maxadapterv1.LoginAttemptStatus_LOGIN_ATTEMPT_STATUS_CODE_REQUIRED:
		return "CODE_REQUIRED"
	case maxadapterv1.LoginAttemptStatus_LOGIN_ATTEMPT_STATUS_PASSWORD_REQUIRED:
		return "PASSWORD_REQUIRED"
	case maxadapterv1.LoginAttemptStatus_LOGIN_ATTEMPT_STATUS_COMPLETED:
		return "COMPLETED"
	case maxadapterv1.LoginAttemptStatus_LOGIN_ATTEMPT_STATUS_FAILED:
		return "FAILED"
	case maxadapterv1.LoginAttemptStatus_LOGIN_ATTEMPT_STATUS_EXPIRED:
		return "EXPIRED"
	default:
		return ""
	}
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func recipientProto(recipient MessageRecipient) *maxadapterv1.MessageRecipient {
	return &maxadapterv1.MessageRecipient{
		Kind:  recipientKindProto(recipient.Kind),
		Value: recipient.Value,
	}
}

func recipientKindProto(kind RecipientKind) maxadapterv1.RecipientKind {
	switch kind {
	case RecipientKindChatID:
		return maxadapterv1.RecipientKind_RECIPIENT_KIND_CHAT_ID
	case RecipientKindPhone:
		return maxadapterv1.RecipientKind_RECIPIENT_KIND_PHONE
	case RecipientKindUserID:
		return maxadapterv1.RecipientKind_RECIPIENT_KIND_USER_ID
	default:
		return maxadapterv1.RecipientKind_RECIPIENT_KIND_UNSPECIFIED
	}
}

func attachmentKindProto(kind AttachmentKind) maxadapterv1.AttachmentKind {
	switch kind {
	case AttachmentKindPhoto:
		return maxadapterv1.AttachmentKind_ATTACHMENT_KIND_PHOTO
	case AttachmentKindVideo:
		return maxadapterv1.AttachmentKind_ATTACHMENT_KIND_VIDEO
	case AttachmentKindAudio:
		return maxadapterv1.AttachmentKind_ATTACHMENT_KIND_AUDIO
	case AttachmentKindFile:
		return maxadapterv1.AttachmentKind_ATTACHMENT_KIND_FILE
	default:
		return maxadapterv1.AttachmentKind_ATTACHMENT_KIND_FILE
	}
}

func adapterEventType(event maxadapterv1.AdapterEventType) AdapterEventType {
	switch event {
	case maxadapterv1.AdapterEventType_ADAPTER_EVENT_TYPE_ACCOUNT_CONNECTED:
		return AdapterEventAccountConnected
	case maxadapterv1.AdapterEventType_ADAPTER_EVENT_TYPE_ACCOUNT_DISCONNECTED:
		return AdapterEventAccountDisconnected
	case maxadapterv1.AdapterEventType_ADAPTER_EVENT_TYPE_MESSAGE_RECEIVED:
		return AdapterEventMessageReceived
	case maxadapterv1.AdapterEventType_ADAPTER_EVENT_TYPE_MESSAGE_READ:
		return AdapterEventMessageRead
	case maxadapterv1.AdapterEventType_ADAPTER_EVENT_TYPE_MESSAGE_REPLIED:
		return AdapterEventMessageReplied
	case maxadapterv1.AdapterEventType_ADAPTER_EVENT_TYPE_RAW:
		return AdapterEventRaw
	default:
		return AdapterEventRaw
	}
}

func toAdapterEvent(event *maxadapterv1.AdapterEvent) (AdapterEvent, error) {
	if event == nil {
		return AdapterEvent{}, nil
	}

	var accountID uuid.UUID
	if event.GetAccountId() != "" {
		parsed, err := uuid.Parse(event.GetAccountId())
		if err != nil {
			return AdapterEvent{}, fmt.Errorf("parse event account id: %w", err)
		}
		accountID = parsed
	}

	return AdapterEvent{
		Seq:              event.GetSeq(),
		Type:             adapterEventType(event.GetType()),
		AccountID:        accountID,
		ChatID:           event.GetChatId(),
		MessageID:        event.GetMessageId(),
		SenderID:         event.GetSenderId(),
		Text:             event.GetText(),
		ReplyToMessageID: event.GetReplyToMessageId(),
		PayloadJSON:      event.GetPayloadJson(),
		OccurredAt:       event.GetOccurredAt(),
	}, nil
}
