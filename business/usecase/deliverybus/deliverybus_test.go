package deliverybus_test

import (
	"context"
	"errors"
	"net/mail"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/usecase/deliverybus"
)

// =============================================================================
// Mocks

type targetQuerierMock struct {
	targets         []campaignbus.Target
	queryErr        error
	changeStatusErr error
	changedStatus   campaignbus.TargetStatus
}

func (m *targetQuerierMock) QueryDue(_ context.Context) ([]campaignbus.Target, error) {
	return m.targets, m.queryErr
}

func (m *targetQuerierMock) ChangeStatus(_ context.Context, _ campaignbus.Target, s campaignbus.TargetStatus) error {
	m.changedStatus = s
	return m.changeStatusErr
}

type campaignQuerierMock struct {
	campaigns map[uuid.UUID]campaignbus.Campaign
	queryErr  error
}

func (m *campaignQuerierMock) QueryByID(_ context.Context, id uuid.UUID) (campaignbus.Campaign, error) {
	if m.queryErr != nil {
		return campaignbus.Campaign{}, m.queryErr
	}
	c, ok := m.campaigns[id]
	if !ok {
		return campaignbus.Campaign{}, campaignbus.ErrCampaignNotFound
	}
	return c, nil
}

type employeeQuerierMock struct {
	employees map[uuid.UUID]employeebus.Employee
	queryErr  error
}

func (m *employeeQuerierMock) QueryByID(_ context.Context, id uuid.UUID) (employeebus.Employee, error) {
	if m.queryErr != nil {
		return employeebus.Employee{}, m.queryErr
	}
	e, ok := m.employees[id]
	if !ok {
		return employeebus.Employee{}, employeebus.ErrNotFound
	}
	return e, nil
}

type messageRendererMock struct {
	messages map[uuid.UUID]messagebus.Message
	queryErr error
}

func (m *messageRendererMock) QueryByID(_ context.Context, id uuid.UUID) (messagebus.Message, error) {
	if m.queryErr != nil {
		return messagebus.Message{}, m.queryErr
	}
	msg, ok := m.messages[id]
	if !ok {
		return messagebus.Message{}, messagebus.ErrNotFound
	}
	return msg, nil
}

type delivererMock struct {
	sendErr  error
	sentTo   string
	sentBody string
}

func (m *delivererMock) Send(_ context.Context, to, _, _, body string, _ map[string][]byte) error {
	m.sentTo = to
	m.sentBody = body
	return m.sendErr
}

type eventPublisherMock struct {
	published  []eventbus.NewEvent
	publishErr error
}

func (m *eventPublisherMock) Publish(_ context.Context, e eventbus.NewEvent) error {
	m.published = append(m.published, e)
	return m.publishErr
}

type attachmentProviderMock struct {
	attachments map[uuid.UUID]attachmentbus.Attachment
	queryErr    error
}

func (m *attachmentProviderMock) QueryByID(_ context.Context, id uuid.UUID) (attachmentbus.Attachment, error) {
	if m.queryErr != nil {
		return attachmentbus.Attachment{}, m.queryErr
	}
	a, ok := m.attachments[id]
	if !ok {
		return attachmentbus.Attachment{}, attachmentbus.ErrNotFound
	}
	return a, nil
}

type renderProviderMock struct {
	content   []byte
	renderErr error
	succeedN  int // succeed N times before returning renderErr; 0 = always apply renderErr
	callCount int
}

func (m *renderProviderMock) Render(_ context.Context, _ attachmentbus.Attachment, _ uuid.UUID) ([]byte, error) {
	m.callCount++
	if m.renderErr != nil && (m.succeedN == 0 || m.callCount > m.succeedN) {
		return nil, m.renderErr
	}
	return m.content, nil
}

// =============================================================================

func Test_Delivery(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testSendMail(), "sendmail")
}

// =============================================================================

func testSendMail() []unittest.Table {
	empID := uuid.New()
	campID := uuid.New()
	msgID := uuid.New()
	htmlBodyID := uuid.New()

	htmlContent := []byte(`<html><body><p>Phishing test</p></body></html>`)

	emp := employeebus.Employee{
		ID:        empID,
		FirstName: name.MustParse("Test"),
		LastName:  name.MustParse("User"),
		Email:     mail.Address{Address: "victim@example.com"},
	}

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: campID,
		EmployeeID: empID,
		Status:     campaignbus.Pending,
	}

	phishDomain := domain.MustParse("https://phish.example.com")

	activeCamp := campaignbus.Campaign{
		ID:        campID,
		Status:    campaignbus.Active,
		MessageID: &msgID,
		Type:      campaignbus.EmailCampaign,
		Domain:    phishDomain,
	}

	msgHTMLBody := messagebus.Message{
		ID:         msgID,
		Type:       messagebus.EmailMessage,
		HtmlBodyID: uuid.NullUUID{UUID: htmlBodyID, Valid: true},
		FromEmail:  mail.Address{Address: "phisher@example.com"},
	}

	htmlAatch := attachmentbus.Attachment{
		ID:          htmlBodyID,
		Type:        attachmentbus.Html,
		Label:       label.MustParse("HTML body"),
		ContentPath: file.MustParse("/content/html.html"),
	}

	cqOK := &campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: activeCamp}}
	eqOK := &employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}}
	mrOK := &messageRendererMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgHTMLBody}}
	apOK := &attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{htmlBodyID: htmlAatch}}

	// Successful send
	epOK := &eventPublisherMock{}
	tqOK := &targetQuerierMock{targets: []campaignbus.Target{target}}
	delivererOK := &delivererMock{}
	busOK := deliverybus.NewBusiness(tqOK, cqOK, eqOK, mrOK, apOK, delivererOK, epOK, &renderProviderMock{content: htmlContent})

	// No targets
	epNoTargets := &eventPublisherMock{}
	busNoTargets := deliverybus.NewBusiness(
		&targetQuerierMock{targets: nil},
		cqOK, eqOK, mrOK, apOK, &delivererMock{}, epNoTargets,
		&renderProviderMock{content: htmlContent},
	)

	// Campaign has no message
	cqNoMsg := &campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{
		campID: {ID: campID, Status: campaignbus.Active, Type: campaignbus.EmailCampaign, Domain: phishDomain},
	}}
	busNoMsg := deliverybus.NewBusiness(
		&targetQuerierMock{targets: []campaignbus.Target{target}},
		cqNoMsg, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
		&renderProviderMock{content: htmlContent},
	)

	// Campaign not active
	cqInactive := &campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{
		campID: {ID: campID, Status: campaignbus.Draft, MessageID: &msgID, Type: campaignbus.EmailCampaign},
	}}
	busInactive := deliverybus.NewBusiness(
		&targetQuerierMock{targets: []campaignbus.Target{target}},
		cqInactive, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
		&renderProviderMock{content: htmlContent},
	)

	// Max campaign — no email sent
	cqMax := &campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{
		campID: {ID: campID, Status: campaignbus.Active, MessageID: &msgID, Type: campaignbus.MaxCampaign},
	}}
	epMax := &eventPublisherMock{}
	busMax := deliverybus.NewBusiness(
		&targetQuerierMock{targets: []campaignbus.Target{target}},
		cqMax, eqOK, mrOK, apOK, &delivererMock{}, epMax,
		&renderProviderMock{content: htmlContent},
	)

	return []unittest.Table{
		{
			Name:    "empty-targets-no-errors",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				errs := busNoTargets.SendMail(ctx)
				return len(errs)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "successful-send-changes-status-to-sent",
			ExpResp: campaignbus.Sent.String(),
			ExcFunc: func(ctx context.Context) any {
				busOK.SendMail(ctx)
				return tqOK.changedStatus.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "successful-send-delivers-to-employee",
			ExpResp: "victim@example.com",
			ExcFunc: func(ctx context.Context) any {
				busOK.SendMail(ctx)
				return delivererOK.sentTo
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "successful-send-publishes-message-sent-event",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busOK.SendMail(ctx)
				for _, e := range epOK.published {
					if e.Type == eventbus.MessageSent {
						return true
					}
				}
				return false
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-message-returns-error-in-list",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				errs := busNoMsg.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "inactive-campaign-returns-error-in-list",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				errs := busInactive.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "max-campaign-skipped-no-errors",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				errs := busMax.SendMail(ctx)
				return len(errs)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				cqErr := &campaignQuerierMock{queryErr: errors.New("db error")}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqErr, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				eqErr := &employeeQuerierMock{queryErr: errors.New("employee not found")}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqErr, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-no-html-body-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				msgNoHTML := messagebus.Message{
					ID:        msgID,
					Type:      messagebus.EmailMessage,
					FromEmail: mail.Address{Address: "from@example.com"},
					// No HtmlBodyID
				}
				mrNoHTML := &messageRendererMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgNoHTML}}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrNoHTML, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "target-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{queryErr: errors.New("query error")},
					cqOK, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				apErr := &attachmentProviderMock{queryErr: errors.New("attachment error")}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrOK, apErr, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-not-html-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				pngAatch := attachmentbus.Attachment{ID: htmlBodyID, Type: attachmentbus.Png, Label: label.MustParse("PNG file"), ContentPath: file.MustParse("/content/file.png")}
				apPng := &attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{htmlBodyID: pngAatch}}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrOK, apPng, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "render-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{renderErr: errors.New("render failed")},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "deliverer-send-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrOK, apOK,
					&delivererMock{sendErr: errors.New("smtp failed")},
					&eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-with-extra-attachments",
			ExpResp: "victim@example.com",
			ExcFunc: func(ctx context.Context) any {
				extraAttID := uuid.New()
				extraAatch := attachmentbus.Attachment{
					ID:          extraAttID,
					Type:        attachmentbus.Png,
					Label:       label.MustParse("Extra att"),
					ContentPath: file.MustParse("/content/extra.png"),
				}
				msgWithAtts := messagebus.Message{
					ID:            msgID,
					Type:          messagebus.EmailMessage,
					FromEmail:     mail.Address{Address: "phisher@example.com"},
					HtmlBodyID:    uuid.NullUUID{UUID: htmlBodyID, Valid: true},
					AttachmentIDs: []uuid.UUID{extraAttID},
				}
				mrWithAtts := &messageRendererMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAtts}}
				apWithAtts := &attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
					htmlBodyID: htmlAatch,
					extraAttID: extraAatch,
				}}
				delivererAtts := &delivererMock{}
				busAtts := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrWithAtts, apWithAtts, delivererAtts, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				busAtts.SendMail(ctx)
				return delivererAtts.sentTo
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "body-without-body-tag-appends-pixel",
			ExpResp: "victim@example.com",
			ExcFunc: func(ctx context.Context) any {
				// HTML without </body> — pixel appended at end
				noBodyHTML := []byte("<p>No body tag here</p>")
				delivererNoBody := &delivererMock{}
				busNoBody := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrOK, apOK, delivererNoBody, &eventPublisherMock{},
					&renderProviderMock{content: noBodyHTML},
				)
				busNoBody.SendMail(ctx)
				return delivererNoBody.sentTo
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "empty-domain-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				cqNoDomain := &campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{
					campID: {ID: campID, Status: campaignbus.Active, MessageID: &msgID, Type: campaignbus.EmailCampaign},
				}}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqNoDomain, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "extra-attachment-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				extraAttID := uuid.New()
				msgWithAtts := messagebus.Message{
					ID:            msgID,
					Type:          messagebus.EmailMessage,
					FromEmail:     mail.Address{Address: "phisher@example.com"},
					HtmlBodyID:    uuid.NullUUID{UUID: htmlBodyID, Valid: true},
					AttachmentIDs: []uuid.UUID{extraAttID},
				}
				mrWithAtts := &messageRendererMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAtts}}
				apMissingExtra := &attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
					htmlBodyID: htmlAatch,
					// extraAttID intentionally absent → ErrNotFound
				}}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrWithAtts, apMissingExtra, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "extra-attachment-render-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				extraAttID := uuid.New()
				extraAatch := attachmentbus.Attachment{
					ID:          extraAttID,
					Type:        attachmentbus.Png,
					Label:       label.MustParse("Extra att"),
					ContentPath: file.MustParse("/content/extra.png"),
				}
				msgWithAtts := messagebus.Message{
					ID:            msgID,
					Type:          messagebus.EmailMessage,
					FromEmail:     mail.Address{Address: "phisher@example.com"},
					HtmlBodyID:    uuid.NullUUID{UUID: htmlBodyID, Valid: true},
					AttachmentIDs: []uuid.UUID{extraAttID},
				}
				mrWithAtts := &messageRendererMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAtts}}
				apWithExtra := &attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
					htmlBodyID: htmlAatch,
					extraAttID: extraAatch,
				}}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrWithAtts, apWithExtra, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent, renderErr: errors.New("render failed"), succeedN: 1},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				mrErr := &messageRendererMock{queryErr: errors.New("message db error")}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrErr, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "change-status-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				tqStatusErr := &targetQuerierMock{
					targets:         []campaignbus.Target{target},
					changeStatusErr: errors.New("status update failed"),
				}
				busErr := deliverybus.NewBusiness(
					tqStatusErr, cqOK, eqOK, mrOK, apOK, &delivererMock{}, &eventPublisherMock{},
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "event-publish-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				epErr := &eventPublisherMock{publishErr: errors.New("publish failed")}
				busErr := deliverybus.NewBusiness(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					cqOK, eqOK, mrOK, apOK, &delivererMock{}, epErr,
					&renderProviderMock{content: htmlContent},
				)
				errs := busErr.SendMail(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
