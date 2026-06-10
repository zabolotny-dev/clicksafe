package messagebus

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

// =============================================================================
// Stubs

type messageStoreStub struct {
	data    map[uuid.UUID]Message
	saveErr error
	qErr    error
}

func newMessageStoreStub() *messageStoreStub {
	return &messageStoreStub{data: make(map[uuid.UUID]Message)}
}

func (s *messageStoreStub) Save(_ context.Context, msg Message) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[msg.ID] = msg
	return nil
}

func (s *messageStoreStub) Update(_ context.Context, msg Message) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[msg.ID] = msg
	return nil
}

func (s *messageStoreStub) Delete(_ context.Context, msg Message) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	delete(s.data, msg.ID)
	return nil
}

func (s *messageStoreStub) QueryByID(_ context.Context, id uuid.UUID) (Message, error) {
	msg, ok := s.data[id]
	if !ok {
		return Message{}, ErrNotFound
	}
	return msg, nil
}

func (s *messageStoreStub) Query(_ context.Context, _ QueryFilter, _ order.By, _ page.Page) ([]Message, error) {
	if s.qErr != nil {
		return nil, s.qErr
	}
	var result []Message
	for _, msg := range s.data {
		result = append(result, msg)
	}
	return result, nil
}

func (s *messageStoreStub) Count(_ context.Context, _ QueryFilter) (int, error) {
	if s.qErr != nil {
		return 0, s.qErr
	}
	return len(s.data), nil
}

type attachmentQuerierStub struct {
	attachments map[uuid.UUID]attachmentbus.Attachment
}

func newAttachmentQuerierStub() attachmentQuerierStub {
	return attachmentQuerierStub{attachments: make(map[uuid.UUID]attachmentbus.Attachment)}
}

func (s attachmentQuerierStub) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	attachment, ok := s.attachments[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return attachment, nil
}

// =============================================================================

func Test_Message(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testSaveEmail(), "save-email")
	unittest.Run(t, testSaveMax(), "save-max")
	unittest.Run(t, testUpdate(), "update")
	unittest.Run(t, testDelete(), "delete")
	unittest.Run(t, testMessageQueryByID(), "querybyid")
	unittest.Run(t, testMessageQuery(), "query")
	unittest.Run(t, testMessageCount(), "count")
	unittest.Run(t, testMessageSeedHelper(), "seed-helper")
	unittest.Run(t, testParseMessageType(), "parse-type")
	unittest.Run(t, testErrorTypes(), "error-types")
	unittest.Run(t, testUpdateBranches(), "update-branches")
}

func testParseMessageType() []unittest.Table {
	return []unittest.Table{
		{
			Name:    "valid-email-type",
			ExpResp: EmailMessage.String(),
			ExcFunc: func(ctx context.Context) any {
				t, err := ParseMessageType("EMAIL")
				if err != nil {
					return err
				}
				return t.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "valid-max-type",
			ExpResp: MaxMessage.String(),
			ExcFunc: func(ctx context.Context) any {
				t, err := ParseMessageType("MAX")
				if err != nil {
					return err
				}
				return t.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "invalid-type-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := ParseMessageType("UNKNOWN")
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testErrorTypes() []unittest.Table {
	return []unittest.Table{
		{
			Name:    "missing-attachments-error-message",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				e := &ErrMissingAttachments{IDs: []uuid.UUID{uuid.New()}}
				return len(e.Error()) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "duplicate-attachments-error-message",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				e := &ErrDuplicateAttachments{IDs: []uuid.UUID{uuid.New()}}
				return len(e.Error()) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "missing-attachments-triggered-by-save",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				aq := newAttachmentQuerierStub()
				// no attachments in stub - so attachment IDs won't be found
				bus := NewBusiness(newMessageStoreStub(), aq)
				missingID := uuid.New()
				_, err := bus.Save(ctx, NewMessage{
					Type:          EmailMessage,
					Label:         label.MustParse("Missing att"),
					FromEmail:     mail.Address{Address: "from@example.com"},
					AttachmentIDs: []uuid.UUID{missingID},
				})
				var missingErr *ErrMissingAttachments
				return errors.As(err, &missingErr)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testUpdateBranches() []unittest.Table {
	htmlID3 := uuid.New()
	aqHTML3 := newAttachmentQuerierStub()
	aqHTML3.attachments[htmlID3] = attachmentbus.Attachment{ID: htmlID3, Type: attachmentbus.Html, Label: label.MustParse("HTML body")}
	busHTML3 := NewBusiness(newMessageStoreStub(), aqHTML3)

	txtID4 := uuid.New()
	aqTxt4 := newAttachmentQuerierStub()
	aqTxt4.attachments[txtID4] = attachmentbus.Attachment{ID: txtID4, Type: attachmentbus.Txt, Label: label.MustParse("TXT body")}
	busUpdateMax := NewBusiness(newMessageStoreStub(), aqTxt4)

	accID := uuid.New()

	return []unittest.Table{
		{
			Name:    "update-html-body-to-valid",
			ExpResp: htmlID3,
			ExcFunc: func(ctx context.Context) any {
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Email msg"),
					FromEmail: mail.Address{Address: "from@example.com"},
				}
				newHTML := uuid.NullUUID{UUID: htmlID3, Valid: true}
				updated, err := busHTML3.Update(ctx, msg, UpdateMessage{HtmlBodyID: &newHTML})
				if err != nil {
					return err
				}
				return updated.HtmlBodyID.UUID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-max-message-success",
			ExpResp: MaxMessage.String(),
			ExcFunc: func(ctx context.Context) any {
				msg := Message{
					ID:           uuid.New(),
					Type:         MaxMessage,
					Label:        label.MustParse("Max msg"),
					MaxAccountID: uuid.NullUUID{UUID: accID, Valid: true},
					TextBodyID:   uuid.NullUUID{UUID: txtID4, Valid: true},
				}
				newLabel := label.MustParse("Updated max msg")
				updated, err := busUpdateMax.Update(ctx, msg, UpdateMessage{Label: &newLabel})
				if err != nil {
					return err
				}
				return updated.Type.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testMessageSeedHelper() []unittest.Table {
	store := newMessageStoreStub()
	bus := NewBusiness(store, newAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "seed-creates-n-email-messages",
			ExpResp: 2,
			ExcFunc: func(ctx context.Context) any {
				msgs, err := TestSeedEmailMessages(ctx, 2, bus)
				if err != nil {
					return err
				}
				return len(msgs)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

// =============================================================================

func testSaveEmail() []unittest.Table {
	busNoSender := NewBusiness(newMessageStoreStub(), newAttachmentQuerierStub())

	htmlID := uuid.New()
	aqHTML := newAttachmentQuerierStub()
	aqHTML.attachments[htmlID] = attachmentbus.Attachment{ID: htmlID, Type: attachmentbus.Html, Label: label.MustParse("HTML body")}
	storeOK := newMessageStoreStub()
	busOK := NewBusiness(storeOK, aqHTML)

	htmlID2 := uuid.New()
	aqWrongHTML := newAttachmentQuerierStub()
	aqWrongHTML.attachments[htmlID2] = attachmentbus.Attachment{ID: htmlID2, Type: attachmentbus.Png, Label: label.MustParse("PNG body")}
	busWrongHTML := NewBusiness(newMessageStoreStub(), aqWrongHTML)

	dupID := uuid.New()
	aqDup := newAttachmentQuerierStub()
	aqDup.attachments[dupID] = attachmentbus.Attachment{ID: dupID, Type: attachmentbus.Png, Label: label.MustParse("PNG att")}
	busDup := NewBusiness(newMessageStoreStub(), aqDup)

	fromEmail := mail.Address{Address: "sender@example.com"}

	return []unittest.Table{
		{
			Name:    "no-sender-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoSender.Save(ctx, NewMessage{Label: label.MustParse("Notice")})
				return errors.Is(err, ErrFromEmailRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "defaults-to-email-type",
			ExpResp: EmailMessage.String(),
			ExcFunc: func(ctx context.Context) any {
				msg, err := busOK.Save(ctx, NewMessage{
					Label:     label.MustParse("Notice"),
					FromEmail: fromEmail,
				})
				if err != nil {
					return err
				}
				return msg.Type.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "with-html-body-success",
			ExpResp: htmlID,
			ExcFunc: func(ctx context.Context) any {
				msg, err := busOK.Save(ctx, NewMessage{
					Label:      label.MustParse("Email msg"),
					FromEmail:  fromEmail,
					HtmlBodyID: uuid.NullUUID{UUID: htmlID, Valid: true},
				})
				if err != nil {
					return err
				}
				return msg.HtmlBodyID.UUID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "html-body-wrong-type-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busWrongHTML.Save(ctx, NewMessage{
					Label:      label.MustParse("Email msg"),
					FromEmail:  fromEmail,
					HtmlBodyID: uuid.NullUUID{UUID: htmlID2, Valid: true},
				})
				return errors.Is(err, ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "duplicate-attachment-ids-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDup.Save(ctx, NewMessage{
					Label:         label.MustParse("Email msg"),
					FromEmail:     fromEmail,
					AttachmentIDs: []uuid.UUID{dupID, dupID},
				})
				var dupErr *ErrDuplicateAttachments
				return errors.As(err, &dupErr)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testSaveMax() []unittest.Table {
	busNoAccount := NewBusiness(newMessageStoreStub(), newAttachmentQuerierStub())

	txtID := uuid.New()
	aqTxt := newAttachmentQuerierStub()
	aqTxt.attachments[txtID] = attachmentbus.Attachment{ID: txtID, Type: attachmentbus.Txt, Label: label.MustParse("Max text")}
	storeMaxOK := newMessageStoreStub()
	busMaxOK := NewBusiness(storeMaxOK, aqTxt)

	nonTxtID := uuid.New()
	aqNonTxt := newAttachmentQuerierStub()
	aqNonTxt.attachments[nonTxtID] = attachmentbus.Attachment{ID: nonTxtID, Type: attachmentbus.Html, Label: label.MustParse("HTML body")}
	busNonTxt := NewBusiness(newMessageStoreStub(), aqNonTxt)

	txtID2 := uuid.New()
	htmlAttID := uuid.New()
	aqHTMLAtt := newAttachmentQuerierStub()
	aqHTMLAtt.attachments[txtID2] = attachmentbus.Attachment{ID: txtID2, Type: attachmentbus.Txt, Label: label.MustParse("Max text")}
	aqHTMLAtt.attachments[htmlAttID] = attachmentbus.Attachment{ID: htmlAttID, Type: attachmentbus.Html, Label: label.MustParse("HTML att")}
	busHTMLAtt := NewBusiness(newMessageStoreStub(), aqHTMLAtt)

	accountID := uuid.New()

	return []unittest.Table{
		{
			Name:    "no-account-id-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoAccount.Save(ctx, NewMessage{Type: MaxMessage, Label: label.MustParse("Max notice")})
				return errors.Is(err, ErrMaxAccountRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-text-body-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoAccount.Save(ctx, NewMessage{
					Type:         MaxMessage,
					Label:        label.MustParse("Max notice"),
					MaxAccountID: uuid.NullUUID{UUID: accountID, Valid: true},
				})
				return errors.Is(err, ErrTextBodyRequired)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "accepts-txt-body",
			ExpResp: MaxMessage.String(),
			ExcFunc: func(ctx context.Context) any {
				msg, err := busMaxOK.Save(ctx, NewMessage{
					Type:         MaxMessage,
					Label:        label.MustParse("Max notice"),
					MaxAccountID: uuid.NullUUID{UUID: accountID, Valid: true},
					TextBodyID:   uuid.NullUUID{UUID: txtID, Valid: true},
				})
				if err != nil {
					return err
				}
				return msg.Type.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "rejects-non-txt-body",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNonTxt.Save(ctx, NewMessage{
					Type:         MaxMessage,
					Label:        label.MustParse("Max notice"),
					MaxAccountID: uuid.NullUUID{UUID: accountID, Valid: true},
					TextBodyID:   uuid.NullUUID{UUID: nonTxtID, Valid: true},
				})
				return errors.Is(err, ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "rejects-html-attachment",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busHTMLAtt.Save(ctx, NewMessage{
					Type:          MaxMessage,
					Label:         label.MustParse("Max notice"),
					MaxAccountID:  uuid.NullUUID{UUID: accountID, Valid: true},
					TextBodyID:    uuid.NullUUID{UUID: txtID2, Valid: true},
					AttachmentIDs: []uuid.UUID{htmlAttID},
				})
				return errors.Is(err, ErrMaxHTMLAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testUpdate() []unittest.Table {
	txtID3 := uuid.New()
	htmlBodyID3 := uuid.New()
	htmlAttID2 := uuid.New()

	aqUpd := newAttachmentQuerierStub()
	aqUpd.attachments[txtID3] = attachmentbus.Attachment{ID: txtID3, Type: attachmentbus.Txt, Label: label.MustParse("Max text")}
	aqUpd.attachments[htmlAttID2] = attachmentbus.Attachment{ID: htmlAttID2, Type: attachmentbus.Html, Label: label.MustParse("HTML att")}
	aqUpd.attachments[htmlBodyID3] = attachmentbus.Attachment{ID: htmlBodyID3, Type: attachmentbus.Html, Label: label.MustParse("HTML body")}

	busUpd := NewBusiness(newMessageStoreStub(), aqUpd)

	// Storer error
	storeErrStub := newMessageStoreStub()
	storeErrStub.saveErr = errors.New("db error")
	busStoreErr := NewBusiness(storeErrStub, aqUpd)

	newLabel := label.MustParse("Updated label")
	fromName, _ := label.ParseNull("Sender Name")
	subj, _ := subject.ParseNull("New Subject")
	msgType := MaxMessage
	accID := uuid.New()
	txtBodyID := uuid.New()

	aqWithTxt := newAttachmentQuerierStub()
	aqWithTxt.attachments[txtBodyID] = attachmentbus.Attachment{ID: txtBodyID, Type: attachmentbus.Txt, Label: label.MustParse("Txt")}
	busWithTxt := NewBusiness(newMessageStoreStub(), aqWithTxt)

	wrongHTMLID := uuid.New()
	aqWrongHTML := newAttachmentQuerierStub()
	aqWrongHTML.attachments[wrongHTMLID] = attachmentbus.Attachment{ID: wrongHTMLID, Type: attachmentbus.Png, Label: label.MustParse("PNG")}
	busWrongHTML := NewBusiness(newMessageStoreStub(), aqWrongHTML)

	return []unittest.Table{
		{
			Name:    "update-to-max-rejects-html-attachment",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busUpd.Update(ctx, Message{
					ID:            uuid.New(),
					Type:          EmailMessage,
					Label:         label.MustParse("Email notice"),
					TextBodyID:    uuid.NullUUID{UUID: txtID3, Valid: true},
					MaxAccountID:  uuid.NullUUID{UUID: uuid.New(), Valid: true},
					AttachmentIDs: []uuid.UUID{htmlAttID2},
				}, UpdateMessage{Type: &msgType})
				return errors.Is(err, ErrMaxHTMLAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-label-only",
			ExpResp: newLabel.String(),
			ExcFunc: func(ctx context.Context) any {
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Old label"),
					FromEmail: mail.Address{Address: "from@example.com"},
				}
				updated, err := busUpd.Update(ctx, msg, UpdateMessage{Label: &newLabel})
				if err != nil {
					return err
				}
				return updated.Label.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-from-email",
			ExpResp: "new@example.com",
			ExcFunc: func(ctx context.Context) any {
				newFrom := mail.Address{Address: "new@example.com"}
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Msg"),
					FromEmail: mail.Address{Address: "old@example.com"},
				}
				updated, err := busUpd.Update(ctx, msg, UpdateMessage{FromEmail: &newFrom})
				if err != nil {
					return err
				}
				return updated.FromEmail.Address
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-from-name-and-subject",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Msg"),
					FromEmail: mail.Address{Address: "from@example.com"},
				}
				updated, err := busUpd.Update(ctx, msg, UpdateMessage{
					FromName: &fromName,
					Subject:  &subj,
				})
				if err != nil {
					return err
				}
				return updated.FromName.String() == "Sender Name" && updated.Subject.String() == "New Subject"
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-text-body-and-max-account",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				nullTxt := uuid.NullUUID{UUID: txtBodyID, Valid: true}
				nullAcc := uuid.NullUUID{UUID: accID, Valid: true}
				msg := Message{
					ID:           uuid.New(),
					Type:         MaxMessage,
					Label:        label.MustParse("Max msg"),
					MaxAccountID: nullAcc,
					TextBodyID:   nullTxt,
				}
				updated, err := busWithTxt.Update(ctx, msg, UpdateMessage{
					TextBodyID:   &nullTxt,
					MaxAccountID: &nullAcc,
				})
				if err != nil {
					return err
				}
				return updated.TextBodyID.UUID == txtBodyID && updated.MaxAccountID.UUID == accID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-clear-html-body",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				existing := uuid.NullUUID{UUID: htmlBodyID3, Valid: true}
				nullBody := uuid.NullUUID{Valid: false}
				msg := Message{
					ID:         uuid.New(),
					Type:       EmailMessage,
					Label:      label.MustParse("Email msg"),
					FromEmail:  mail.Address{Address: "from@example.com"},
					HtmlBodyID: existing,
				}
				updated, err := busUpd.Update(ctx, msg, UpdateMessage{HtmlBodyID: &nullBody})
				if err != nil {
					return err
				}
				return updated.HtmlBodyID.Valid
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-html-body-wrong-type",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				pngBody := uuid.NullUUID{UUID: wrongHTMLID, Valid: true}
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Msg"),
					FromEmail: mail.Address{Address: "from@example.com"},
				}
				_, err := busWrongHTML.Update(ctx, msg, UpdateMessage{HtmlBodyID: &pngBody})
				return errors.Is(err, ErrInvalidAttachment)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-storer-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Msg"),
					FromEmail: mail.Address{Address: "from@example.com"},
				}
				_, err := busStoreErr.Update(ctx, msg, UpdateMessage{Label: &newLabel})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-set-html-body",
			ExpResp: htmlBodyID3,
			ExcFunc: func(ctx context.Context) any {
				htmlBody := uuid.NullUUID{UUID: htmlBodyID3, Valid: true}
				msg := Message{
					ID:        uuid.New(),
					Type:      EmailMessage,
					Label:     label.MustParse("Email msg"),
					FromEmail: mail.Address{Address: "from@example.com"},
				}
				updated, err := busUpd.Update(ctx, msg, UpdateMessage{HtmlBodyID: &htmlBody})
				if err != nil {
					return err
				}
				return updated.HtmlBodyID.UUID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testDelete() []unittest.Table {
	msg := Message{
		ID:    uuid.New(),
		Type:  EmailMessage,
		Label: label.MustParse("ToDelete"),
	}
	store := newMessageStoreStub()
	store.data[msg.ID] = msg
	bus := NewBusiness(store, newAttachmentQuerierStub())

	storeWithErr := newMessageStoreStub()
	storeWithErr.data[msg.ID] = msg
	storeWithErr.saveErr = errors.New("db error")
	busErr := NewBusiness(storeWithErr, newAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "removes-from-storer",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				if err := bus.Delete(ctx, msg); err != nil {
					return err
				}
				_, exists := store.data[msg.ID]
				return exists
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "delete-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busErr.Delete(ctx, msg) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testMessageQueryByID() []unittest.Table {
	msg := Message{
		ID:    uuid.New(),
		Type:  EmailMessage,
		Label: label.MustParse("Find me"),
	}
	storeFound := newMessageStoreStub()
	storeFound.data[msg.ID] = msg
	busFound := NewBusiness(storeFound, newAttachmentQuerierStub())
	busNotFound := NewBusiness(newMessageStoreStub(), newAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-message",
			ExpResp: msg.ID,
			ExcFunc: func(ctx context.Context) any {
				got, err := busFound.QueryByID(ctx, msg.ID)
				if err != nil {
					return err
				}
				return got.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "not-found-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNotFound.QueryByID(ctx, uuid.New())
				return errors.Is(err, ErrNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testMessageQuery() []unittest.Table {
	store := newMessageStoreStub()
	bus := NewBusiness(store, newAttachmentQuerierStub())

	for i := range 3 {
		store.data[uuid.New()] = Message{
			ID:    uuid.New(),
			Type:  EmailMessage,
			Label: label.MustParse(fmt.Sprintf("Message %d of test", i+1)),
		}
	}

	storeErr := newMessageStoreStub()
	storeErr.qErr = errors.New("db error")
	busQueryErr := NewBusiness(storeErr, newAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-all",
			ExpResp: 3,
			ExcFunc: func(ctx context.Context) any {
				result, err := bus.Query(ctx, QueryFilter{}, DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}
				return len(result)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQueryErr.Query(ctx, QueryFilter{}, DefaultOrderBy, page.MustParse("1", "10"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testMessageCount() []unittest.Table {
	store := newMessageStoreStub()
	bus := NewBusiness(store, newAttachmentQuerierStub())

	for range 5 {
		store.data[uuid.New()] = Message{ID: uuid.New(), Type: EmailMessage, Label: label.MustParse("Message cnt")}
	}

	storeErr := newMessageStoreStub()
	storeErr.qErr = errors.New("db error")
	busCountErr := NewBusiness(storeErr, newAttachmentQuerierStub())

	return []unittest.Table{
		{
			Name:    "returns-correct-count",
			ExpResp: 5,
			ExcFunc: func(ctx context.Context) any {
				count, err := bus.Count(ctx, QueryFilter{})
				if err != nil {
					return err
				}
				return count
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "count-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busCountErr.Count(ctx, QueryFilter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
