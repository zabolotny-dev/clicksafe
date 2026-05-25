package messagebus

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type messageStoreStub struct {
	saved Message
}

func (s *messageStoreStub) Save(_ context.Context, msg Message) error {
	s.saved = msg
	return nil
}

func (s *messageStoreStub) Update(_ context.Context, msg Message) error {
	s.saved = msg
	return nil
}

func (s *messageStoreStub) Delete(_ context.Context, _ Message) error { return nil }

func (s *messageStoreStub) QueryByID(_ context.Context, _ uuid.UUID) (Message, error) {
	return Message{}, ErrNotFound
}

func (s *messageStoreStub) Query(_ context.Context, _ QueryFilter, _ order.By, _ page.Page) ([]Message, error) {
	return nil, nil
}

func (s *messageStoreStub) Count(_ context.Context, _ QueryFilter) (int, error) { return 0, nil }

type attachmentQuerierStub struct {
	attachments map[uuid.UUID]attachmentbus.Attachment
}

func (s attachmentQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	attachment, ok := s.attachments[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return attachment, nil
}

func TestSaveDefaultsToEmailAndRequiresSender(t *testing.T) {
	business := NewBusiness(&messageStoreStub{}, attachmentQuerierStub{})

	_, err := business.Save(context.Background(), NewMessage{
		Label: label.MustParse("Quarterly notice"),
	})
	if !errors.Is(err, ErrFromEmailRequired) {
		t.Fatalf("Save error = %v, want %v", err, ErrFromEmailRequired)
	}
}

func TestSaveMaxRequiresAccountAndTextBody(t *testing.T) {
	business := NewBusiness(&messageStoreStub{}, attachmentQuerierStub{})

	_, err := business.Save(context.Background(), NewMessage{
		Type:  MaxMessage,
		Label: label.MustParse("Max notice"),
	})
	if !errors.Is(err, ErrMaxAccountRequired) {
		t.Fatalf("Save error = %v, want %v", err, ErrMaxAccountRequired)
	}

	accountID := uuid.New()
	_, err = business.Save(context.Background(), NewMessage{
		Type:         MaxMessage,
		Label:        label.MustParse("Max notice"),
		MaxAccountID: uuid.NullUUID{UUID: accountID, Valid: true},
	})
	if !errors.Is(err, ErrTextBodyRequired) {
		t.Fatalf("Save error = %v, want %v", err, ErrTextBodyRequired)
	}
}

func TestSaveMaxAcceptsTxtBody(t *testing.T) {
	textBodyID := uuid.New()
	store := &messageStoreStub{}
	business := NewBusiness(store, attachmentQuerierStub{
		attachments: map[uuid.UUID]attachmentbus.Attachment{
			textBodyID: {
				ID:    textBodyID,
				Type:  attachmentbus.Txt,
				Label: label.MustParse("Max text"),
			},
		},
	})

	message, err := business.Save(context.Background(), NewMessage{
		Type:         MaxMessage,
		Label:        label.MustParse("Max notice"),
		MaxAccountID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: textBodyID, Valid: true},
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if message.Type != MaxMessage {
		t.Fatalf("message type = %s, want %s", message.Type.String(), MaxMessage.String())
	}
	if store.saved.TextBodyID.UUID != textBodyID {
		t.Fatalf("saved text body = %s, want %s", store.saved.TextBodyID.UUID, textBodyID)
	}
}

func TestSaveMaxRejectsNonTxtBody(t *testing.T) {
	bodyID := uuid.New()
	business := NewBusiness(&messageStoreStub{}, attachmentQuerierStub{
		attachments: map[uuid.UUID]attachmentbus.Attachment{
			bodyID: {
				ID:    bodyID,
				Type:  attachmentbus.Html,
				Label: label.MustParse("HTML body"),
			},
		},
	})

	_, err := business.Save(context.Background(), NewMessage{
		Type:         MaxMessage,
		Label:        label.MustParse("Max notice"),
		MaxAccountID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: bodyID, Valid: true},
	})
	if !errors.Is(err, ErrInvalidAttachment) {
		t.Fatalf("Save error = %v, want %v", err, ErrInvalidAttachment)
	}
}

func TestSaveMaxRejectsHTMLAttachment(t *testing.T) {
	textBodyID := uuid.New()
	htmlAttachmentID := uuid.New()
	business := NewBusiness(&messageStoreStub{}, attachmentQuerierStub{
		attachments: map[uuid.UUID]attachmentbus.Attachment{
			textBodyID: {
				ID:    textBodyID,
				Type:  attachmentbus.Txt,
				Label: label.MustParse("Max text"),
			},
			htmlAttachmentID: {
				ID:    htmlAttachmentID,
				Type:  attachmentbus.Html,
				Label: label.MustParse("HTML attachment"),
			},
		},
	})

	_, err := business.Save(context.Background(), NewMessage{
		Type:          MaxMessage,
		Label:         label.MustParse("Max notice"),
		MaxAccountID:  uuid.NullUUID{UUID: uuid.New(), Valid: true},
		TextBodyID:    uuid.NullUUID{UUID: textBodyID, Valid: true},
		AttachmentIDs: []uuid.UUID{htmlAttachmentID},
	})
	if !errors.Is(err, ErrMaxHTMLAttachment) {
		t.Fatalf("Save error = %v, want %v", err, ErrMaxHTMLAttachment)
	}
}

func TestUpdateToMaxRejectsExistingHTMLAttachment(t *testing.T) {
	textBodyID := uuid.New()
	htmlAttachmentID := uuid.New()
	business := NewBusiness(&messageStoreStub{}, attachmentQuerierStub{
		attachments: map[uuid.UUID]attachmentbus.Attachment{
			textBodyID: {
				ID:    textBodyID,
				Type:  attachmentbus.Txt,
				Label: label.MustParse("Max text"),
			},
			htmlAttachmentID: {
				ID:    htmlAttachmentID,
				Type:  attachmentbus.Html,
				Label: label.MustParse("HTML attachment"),
			},
		},
	})
	msgType := MaxMessage

	_, err := business.Update(context.Background(), Message{
		ID:            uuid.New(),
		Type:          EmailMessage,
		Label:         label.MustParse("Email notice"),
		TextBodyID:    uuid.NullUUID{UUID: textBodyID, Valid: true},
		MaxAccountID:  uuid.NullUUID{UUID: uuid.New(), Valid: true},
		AttachmentIDs: []uuid.UUID{htmlAttachmentID},
	}, UpdateMessage{
		Type: &msgType,
	})
	if !errors.Is(err, ErrMaxHTMLAttachment) {
		t.Fatalf("Update error = %v, want %v", err, ErrMaxHTMLAttachment)
	}
}
