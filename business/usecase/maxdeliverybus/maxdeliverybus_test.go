package maxdeliverybus_test

import (
	"context"
	"errors"
	"net/mail"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/maxadapter"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
	"github.com/zabolotny-dev/clicksafe/business/usecase/maxdeliverybus"
)

// =============================================================================
// Mocks

type targetQuerierMock struct {
	targets       []campaignbus.Target
	queryDueErr   error
	queryByIDMap  map[uuid.UUID]campaignbus.Target
	changedStatus campaignbus.TargetStatus
	changeErr     error
}

func (m *targetQuerierMock) QueryDue(_ context.Context) ([]campaignbus.Target, error) {
	return m.targets, m.queryDueErr
}

func (m *targetQuerierMock) QueryByID(_ context.Context, id uuid.UUID) (campaignbus.Target, error) {
	if t, ok := m.queryByIDMap[id]; ok {
		return t, nil
	}
	return campaignbus.Target{}, campaignbus.ErrTargetNotFound
}

func (m *targetQuerierMock) ChangeStatus(_ context.Context, _ campaignbus.Target, s campaignbus.TargetStatus) error {
	m.changedStatus = s
	return m.changeErr
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

type messageQuerierMock struct {
	messages map[uuid.UUID]messagebus.Message
	queryErr error
}

func (m *messageQuerierMock) QueryByID(_ context.Context, id uuid.UUID) (messagebus.Message, error) {
	if m.queryErr != nil {
		return messagebus.Message{}, m.queryErr
	}
	msg, ok := m.messages[id]
	if !ok {
		return messagebus.Message{}, messagebus.ErrNotFound
	}
	return msg, nil
}

type maxAccountQuerierMock struct {
	accounts map[uuid.UUID]maxaccountbus.Account
	queryErr error
}

func (m *maxAccountQuerierMock) QueryByID(_ context.Context, id uuid.UUID) (maxaccountbus.Account, error) {
	if m.queryErr != nil {
		return maxaccountbus.Account{}, m.queryErr
	}
	acc, ok := m.accounts[id]
	if !ok {
		return maxaccountbus.Account{}, maxaccountbus.ErrAccountNotFound
	}
	return acc, nil
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

type eventPublisherMock struct {
	published  []eventbus.NewEvent
	publishErr error
}

func (m *eventPublisherMock) Publish(_ context.Context, e eventbus.NewEvent) error {
	m.published = append(m.published, e)
	return m.publishErr
}

type adapterMock struct {
	sendErr  error
	sentReqs []maxadapter.SendMessageRequest
	subErr   error
	ackErr   error
	events   []maxadapter.AdapterEvent
}

func (m *adapterMock) SendMessage(_ context.Context, req maxadapter.SendMessageRequest) (maxadapter.SendMessageResponse, error) {
	m.sentReqs = append(m.sentReqs, req)
	return maxadapter.SendMessageResponse{ChatID: "chat-123", MessageID: "msg-456"}, m.sendErr
}

func (m *adapterMock) SubscribeEvents(_ context.Context, _ string, _ int64, handle func(maxadapter.AdapterEvent) error) error {
	if m.subErr != nil {
		return m.subErr
	}
	for _, event := range m.events {
		if err := handle(event); err != nil {
			return err
		}
	}
	return nil
}

func (m *adapterMock) AckEvents(_ context.Context, _ string, _ int64) (int64, error) {
	return 0, m.ackErr
}

type deliveryStoreMock struct {
	deliveries                map[uuid.UUID]maxdeliverybus.Delivery
	processedSeqs             map[int64]bool
	saveErr                   error
	markReadResult            bool
	markReadErr               error
	markReplyResult           bool
	markReplyErr              error
	queryByMessageDelivery    *maxdeliverybus.Delivery
	queryByMessageErr         error
	queryReplyDelivery        *maxdeliverybus.Delivery
	queryReplyErr             error
	queryLatestUnreadDelivery *maxdeliverybus.Delivery
	queryLatestUnreadErr      error
	isProcessedErr            error
	markProcessedErr          error
	markEducationErr          error
}

func newDeliveryStoreMock() *deliveryStoreMock {
	return &deliveryStoreMock{
		deliveries:    make(map[uuid.UUID]maxdeliverybus.Delivery),
		processedSeqs: make(map[int64]bool),
	}
}

func (m *deliveryStoreMock) SaveSent(_ context.Context, d maxdeliverybus.Delivery) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.deliveries[d.ID] = d
	return nil
}

func (m *deliveryStoreMock) QueryByID(_ context.Context, id uuid.UUID) (maxdeliverybus.Delivery, error) {
	d, ok := m.deliveries[id]
	if !ok {
		return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
	}
	return d, nil
}

func (m *deliveryStoreMock) QueryByMessage(_ context.Context, _ uuid.UUID, _ string, _ string) (maxdeliverybus.Delivery, error) {
	if m.queryByMessageDelivery != nil {
		return *m.queryByMessageDelivery, nil
	}
	if m.queryByMessageErr != nil {
		return maxdeliverybus.Delivery{}, m.queryByMessageErr
	}
	return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
}

func (m *deliveryStoreMock) QueryReply(_ context.Context, _ uuid.UUID, _ string, _ string) (maxdeliverybus.Delivery, error) {
	if m.queryReplyDelivery != nil {
		return *m.queryReplyDelivery, nil
	}
	if m.queryReplyErr != nil {
		return maxdeliverybus.Delivery{}, m.queryReplyErr
	}
	return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
}

func (m *deliveryStoreMock) QueryLatestUnreadByChat(_ context.Context, _ uuid.UUID, _ string) (maxdeliverybus.Delivery, error) {
	if m.queryLatestUnreadDelivery != nil {
		return *m.queryLatestUnreadDelivery, nil
	}
	if m.queryLatestUnreadErr != nil {
		return maxdeliverybus.Delivery{}, m.queryLatestUnreadErr
	}
	return maxdeliverybus.Delivery{}, maxdeliverybus.ErrDeliveryNotFound
}

func (m *deliveryStoreMock) MarkRead(_ context.Context, id uuid.UUID, at time.Time) (bool, error) {
	if m.markReadErr != nil {
		return false, m.markReadErr
	}
	d, ok := m.deliveries[id]
	if !ok {
		return false, nil
	}
	d.ReadAt = &at
	m.deliveries[id] = d
	return m.markReadResult, nil
}

func (m *deliveryStoreMock) MarkReplied(_ context.Context, id uuid.UUID, msgID string, at time.Time) (bool, error) {
	if m.markReplyErr != nil {
		return false, m.markReplyErr
	}
	d, ok := m.deliveries[id]
	if !ok {
		return false, nil
	}
	d.RepliedAt = &at
	d.IncomingMessageID = msgID
	m.deliveries[id] = d
	return m.markReplyResult, nil
}

func (m *deliveryStoreMock) MarkEducationSent(_ context.Context, id uuid.UUID, at time.Time) error {
	if m.markEducationErr != nil {
		return m.markEducationErr
	}
	d, ok := m.deliveries[id]
	if !ok {
		return maxdeliverybus.ErrDeliveryNotFound
	}
	d.EducationSentAt = &at
	m.deliveries[id] = d
	return nil
}

func (m *deliveryStoreMock) IsProcessed(_ context.Context, seq int64) (bool, error) {
	if m.isProcessedErr != nil {
		return false, m.isProcessedErr
	}
	return m.processedSeqs[seq], nil
}

func (m *deliveryStoreMock) MarkProcessed(_ context.Context, seq int64) error {
	if m.markProcessedErr != nil {
		return m.markProcessedErr
	}
	m.processedSeqs[seq] = true
	return nil
}

// =============================================================================

func Test_MaxDelivery(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testSendDue(), "senddue")
	unittest.Run(t, testConsumeEvents(), "consume")
}

// =============================================================================

func testSendDue() []unittest.Table {
	campID := uuid.New()
	msgID := uuid.New()
	accID := uuid.New()
	adapterAccID := uuid.New()
	txtID := uuid.New()
	empID := uuid.New()
	imgID := uuid.New()

	ph, _ := phone.ParseNull("+79991234567")
	emp := employeebus.Employee{
		ID:        empID,
		FirstName: name.MustParse("Ivan"),
		LastName:  name.MustParse("Petrov"),
		Email:     mail.Address{Address: "ivan@example.com"},
		Phone:     ph,
	}

	acc := maxaccountbus.Account{
		ID:        accID,
		AdapterID: adapterAccID,
		Phone:     "+79001234567",
		Label:     "Test Account",
	}

	txtAatch := attachmentbus.Attachment{
		ID:          txtID,
		Type:        attachmentbus.Txt,
		Label:       label.MustParse("Text body"),
		ContentPath: file.MustParse("/content/text.txt"),
	}

	imgAatch := attachmentbus.Attachment{
		ID:          imgID,
		Type:        attachmentbus.Png,
		Label:       label.MustParse("Image"),
		ContentPath: file.MustParse("/content/img.png"),
	}

	msg := messagebus.Message{
		ID:           msgID,
		Type:         messagebus.MaxMessage,
		FromEmail:    mail.Address{Address: "from@example.com"},
		MaxAccountID: uuid.NullUUID{UUID: accID, Valid: true},
		TextBodyID:   uuid.NullUUID{UUID: txtID, Valid: true},
	}

	msgWithAtts := messagebus.Message{
		ID:            msgID,
		Type:          messagebus.MaxMessage,
		FromEmail:     mail.Address{Address: "from@example.com"},
		MaxAccountID:  uuid.NullUUID{UUID: accID, Valid: true},
		TextBodyID:    uuid.NullUUID{UUID: txtID, Valid: true},
		AttachmentIDs: []uuid.UUID{imgID},
	}

	maxCamp := campaignbus.Campaign{
		ID:        campID,
		Type:      campaignbus.MaxCampaign,
		Status:    campaignbus.Active,
		MessageID: &msgID,
	}

	target := campaignbus.Target{
		ID:         uuid.New(),
		CampaignID: campID,
		EmployeeID: empID,
		Status:     campaignbus.Pending,
	}

	// -------------------------------------------------------------------------
	// Shared OK instances

	adapterOK := &adapterMock{}
	storeOK := newDeliveryStoreMock()
	epOK := &eventPublisherMock{}
	tqOK := &targetQuerierMock{
		targets:      []campaignbus.Target{target},
		queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
	}

	busOK := maxdeliverybus.NewBusiness(
		tqOK,
		&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
		&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
		&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
		&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
		&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: txtAatch}},
		&renderProviderMock{content: []byte("Hello Ivan")},
		epOK,
		adapterOK,
		storeOK,
	)

	// -------------------------------------------------------------------------
	// No targets

	busNoTargets := maxdeliverybus.NewBusiness(
		&targetQuerierMock{targets: nil},
		&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{}},
		&employeeQuerierMock{},
		&messageQuerierMock{},
		&maxAccountQuerierMock{},
		&attachmentProviderMock{},
		&renderProviderMock{},
		&eventPublisherMock{},
		&adapterMock{},
		newDeliveryStoreMock(),
	)

	// -------------------------------------------------------------------------
	// Email campaign — should skip

	emailCamp := campaignbus.Campaign{
		ID:        campID,
		Type:      campaignbus.EmailCampaign,
		Status:    campaignbus.Active,
		MessageID: &msgID,
	}
	busEmailCamp := maxdeliverybus.NewBusiness(
		&targetQuerierMock{targets: []campaignbus.Target{target}},
		&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: emailCamp}},
		&employeeQuerierMock{},
		&messageQuerierMock{},
		&maxAccountQuerierMock{},
		&attachmentProviderMock{},
		&renderProviderMock{},
		&eventPublisherMock{},
		&adapterMock{},
		newDeliveryStoreMock(),
	)

	// -------------------------------------------------------------------------
	// Adapter send error

	adapterErr := &adapterMock{sendErr: errors.New("adapter unavailable")}
	tqErr := &targetQuerierMock{
		targets:      []campaignbus.Target{target},
		queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
	}
	busAdapterErr := maxdeliverybus.NewBusiness(
		tqErr,
		&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
		&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
		&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
		&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
		&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: txtAatch}},
		&renderProviderMock{content: []byte("Hello")},
		&eventPublisherMock{},
		adapterErr,
		newDeliveryStoreMock(),
	)

	newBusForTarget := func(tq *targetQuerierMock, cq *campaignQuerierMock, eq *employeeQuerierMock, mq *messageQuerierMock, aq *maxAccountQuerierMock, ap *attachmentProviderMock, rp *renderProviderMock, ep *eventPublisherMock, ad *adapterMock, st *deliveryStoreMock) *maxdeliverybus.Business {
		return maxdeliverybus.NewBusiness(tq, cq, eq, mq, aq, ap, rp, ep, ad, st)
	}

	return []unittest.Table{
		{
			Name:    "no-targets-returns-no-errors",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				return len(busNoTargets.SendDue(ctx))
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "email-campaign-skipped-no-errors",
			ExpResp: 0,
			ExcFunc: func(ctx context.Context) any {
				return len(busEmailCamp.SendDue(ctx))
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "successful-send-changes-status-to-sent",
			ExpResp: campaignbus.Sent.String(),
			ExcFunc: func(ctx context.Context) any {
				busOK.SendDue(ctx)
				return tqOK.changedStatus.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "successful-send-publishes-message-sent-event",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busOK.SendDue(ctx)
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
			Name:    "successful-send-stores-delivery",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busOK.SendDue(ctx)
				return len(storeOK.deliveries) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "adapter-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				errs := busAdapterErr.SendDue(ctx)
				return len(errs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "targets-query-due-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{queryDueErr: errors.New("db down")},
					&campaignQuerierMock{},
					&employeeQuerierMock{},
					&messageQuerierMock{},
					&maxAccountQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-query-error-adds-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{queryErr: errors.New("campaign db error")},
					&employeeQuerierMock{},
					&messageQuerierMock{},
					&maxAccountQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-message-id-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				campNoMsg := campaignbus.Campaign{ID: campID, Type: campaignbus.MaxCampaign, Status: campaignbus.Active}
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campNoMsg}},
					&employeeQuerierMock{},
					&messageQuerierMock{},
					&maxAccountQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{},
					&messageQuerierMock{queryErr: errors.New("msg db error")},
					&maxAccountQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-type-mismatch-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				emailMsg := messagebus.Message{ID: msgID, Type: messagebus.EmailMessage}
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: emailMsg}},
					&maxAccountQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "account-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{queryErr: errors.New("account not found")},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{queryErr: errors.New("employee not found")},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "employee-no-phone-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				empNoPhone := employeebus.Employee{
					ID:        empID,
					FirstName: name.MustParse("No"),
					LastName:  name.MustParse("Phone"),
					Email:     mail.Address{Address: "nophone@example.com"},
				}
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: empNoPhone}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "text-body-query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{queryErr: errors.New("attachment db error")},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "text-body-not-txt-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				htmlAsText := attachmentbus.Attachment{
					ID:          txtID,
					Type:        attachmentbus.Html,
					Label:       label.MustParse("Wrong type"),
					ContentPath: file.MustParse("/content/wrong.html"),
				}
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: htmlAsText}},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "text-render-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: txtAatch}},
					&renderProviderMock{renderErr: errors.New("render failed")},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-with-image-attachment-succeeds",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				tqImg := &targetQuerierMock{
					targets:      []campaignbus.Target{target},
					queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
				}
				bus := newBusForTarget(
					tqImg,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAtts}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
						txtID: txtAatch,
						imgID: imgAatch,
					}},
					&renderProviderMock{content: []byte("text content")},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				errs := bus.SendDue(ctx)
				return len(errs) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "extra-attachment-query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				apMissingImg := &attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
					txtID: txtAatch,
					// imgID absent → ErrNotFound
				}}
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAtts}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					apMissingImg,
					&renderProviderMock{content: []byte("text content")},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "extra-attachment-render-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAtts}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
						txtID: txtAatch,
						imgID: imgAatch,
					}},
					// succeedN=1: text render succeeds, image render fails
					&renderProviderMock{renderErr: errors.New("render failed"), succeedN: 1},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-with-audio-attachment-succeeds",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				audioID := uuid.New()
				audioAatch := attachmentbus.Attachment{
					ID:          audioID,
					Type:        attachmentbus.Mp3,
					Label:       label.MustParse("Audio file"),
					ContentPath: file.MustParse("/content/audio.mp3"),
				}
				msgWithAudio := messagebus.Message{
					ID:            msgID,
					Type:          messagebus.MaxMessage,
					FromEmail:     mail.Address{Address: "from@example.com"},
					MaxAccountID:  uuid.NullUUID{UUID: accID, Valid: true},
					TextBodyID:    uuid.NullUUID{UUID: txtID, Valid: true},
					AttachmentIDs: []uuid.UUID{audioID},
				}
				tqAudio := &targetQuerierMock{
					targets:      []campaignbus.Target{target},
					queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
				}
				bus := newBusForTarget(
					tqAudio,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithAudio}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
						txtID:   txtAatch,
						audioID: audioAatch,
					}},
					&renderProviderMock{content: []byte("audio content")},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-with-docx-attachment-uses-file-kind",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				docxID := uuid.New()
				docxAatch := attachmentbus.Attachment{
					ID:          docxID,
					Type:        attachmentbus.Docx,
					Label:       label.MustParse("Word doc"),
					ContentPath: file.MustParse("/content/doc.docx"),
				}
				msgWithDocx := messagebus.Message{
					ID:            msgID,
					Type:          messagebus.MaxMessage,
					FromEmail:     mail.Address{Address: "from@example.com"},
					MaxAccountID:  uuid.NullUUID{UUID: accID, Valid: true},
					TextBodyID:    uuid.NullUUID{UUID: txtID, Valid: true},
					AttachmentIDs: []uuid.UUID{docxID},
				}
				tqDocx := &targetQuerierMock{
					targets:      []campaignbus.Target{target},
					queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
				}
				bus := newBusForTarget(
					tqDocx,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msgWithDocx}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
						txtID:   txtAatch,
						docxID: docxAatch,
					}},
					&renderProviderMock{content: []byte("docx content")},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "save-delivery-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.saveErr = errors.New("save failed")
				bus := newBusForTarget(
					&targetQuerierMock{targets: []campaignbus.Target{target}},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: txtAatch}},
					&renderProviderMock{content: []byte("Hello")},
					&eventPublisherMock{},
					&adapterMock{},
					st,
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "change-status-after-save-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				tqChangeErr := &targetQuerierMock{
					targets:      []campaignbus.Target{target},
					queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
					changeErr:    errors.New("status update failed"),
				}
				bus := newBusForTarget(
					tqChangeErr,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: txtAatch}},
					&renderProviderMock{content: []byte("Hello")},
					&eventPublisherMock{},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "publish-event-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				tqForPub := &targetQuerierMock{
					targets:      []campaignbus.Target{target},
					queryByIDMap: map[uuid.UUID]campaignbus.Target{target.ID: target},
				}
				bus := newBusForTarget(
					tqForPub,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: maxCamp}},
					&employeeQuerierMock{employees: map[uuid.UUID]employeebus.Employee{empID: emp}},
					&messageQuerierMock{messages: map[uuid.UUID]messagebus.Message{msgID: msg}},
					&maxAccountQuerierMock{accounts: map[uuid.UUID]maxaccountbus.Account{accID: acc}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{txtID: txtAatch}},
					&renderProviderMock{content: []byte("Hello")},
					&eventPublisherMock{publishErr: errors.New("publish failed")},
					&adapterMock{},
					newDeliveryStoreMock(),
				)
				return len(bus.SendDue(ctx)) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

// =============================================================================

func testConsumeEvents() []unittest.Table {
	campID := uuid.New()
	msgID := uuid.New()
	accID := uuid.New()
	adapterAccID := uuid.New()
	txtID := uuid.New()
	eduTxtID := uuid.New()
	empID := uuid.New()
	deliveryID := uuid.New()
	targetID := uuid.New()
	accountID := uuid.New()

	txtAatch := attachmentbus.Attachment{
		ID:          txtID,
		Type:        attachmentbus.Txt,
		Label:       label.MustParse("Text body"),
		ContentPath: file.MustParse("/content/text.txt"),
	}

	eduTxtAatch := attachmentbus.Attachment{
		ID:          eduTxtID,
		Type:        attachmentbus.Txt,
		Label:       label.MustParse("Education text"),
		ContentPath: file.MustParse("/content/edu.txt"),
	}

	campWithEdu := campaignbus.Campaign{
		ID:                 campID,
		Type:               campaignbus.MaxCampaign,
		Status:             campaignbus.Active,
		MessageID:          &msgID,
		MaxEducationTextID: uuid.NullUUID{UUID: eduTxtID, Valid: true},
	}

	campNoEdu := campaignbus.Campaign{
		ID:       campID,
		Type:     campaignbus.MaxCampaign,
		Status:   campaignbus.Active,
		MessageID: &msgID,
	}

	baseDelivery := maxdeliverybus.Delivery{
		ID:               deliveryID,
		TargetID:         targetID,
		CampaignID:       campID,
		EmployeeID:       empID,
		MaxAccountID:     accID,
		AdapterAccountID: adapterAccID,
		ChatID:           "chat-abc",
		MessageID:        "msg-xyz",
		SentAt:           time.Now().UTC(),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	sentTarget := campaignbus.Target{
		ID:         targetID,
		CampaignID: campID,
		EmployeeID: empID,
		Status:     campaignbus.Sent,
	}

	openedTarget := campaignbus.Target{
		ID:         targetID,
		CampaignID: campID,
		EmployeeID: empID,
		Status:     campaignbus.Opened,
	}

	readEvent := maxadapter.AdapterEvent{
		Seq:       1,
		Type:      maxadapter.AdapterEventMessageRead,
		AccountID: accountID,
		ChatID:    "chat-abc",
		MessageID: "msg-xyz",
	}

	replyEvent := maxadapter.AdapterEvent{
		Seq:              2,
		Type:             maxadapter.AdapterEventMessageReplied,
		AccountID:        accountID,
		ChatID:           "chat-abc",
		MessageID:        "incoming-msg",
		ReplyToMessageID: "msg-xyz",
	}

	receivedEvent := maxadapter.AdapterEvent{
		Seq:              3,
		Type:             maxadapter.AdapterEventMessageReceived,
		AccountID:        accountID,
		ChatID:           "chat-abc",
		MessageID:        "incoming-msg-2",
		ReplyToMessageID: "msg-xyz",
	}

	unknownEvent := maxadapter.AdapterEvent{
		Seq:  4,
		Type: maxadapter.AdapterEventAccountConnected,
	}

	newBus := func(tq *targetQuerierMock, cq *campaignQuerierMock, ap *attachmentProviderMock, rp *renderProviderMock, ep *eventPublisherMock, ad *adapterMock, st *deliveryStoreMock) *maxdeliverybus.Business {
		return maxdeliverybus.NewBusiness(
			tq, cq,
			&employeeQuerierMock{},
			&messageQuerierMock{},
			&maxAccountQuerierMock{},
			ap, rp, ep, ad, st,
		)
	}

	return []unittest.Table{
		{
			Name:    "subscribe-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{subErr: errors.New("subscribe failed")},
					newDeliveryStoreMock(),
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "event-already-processed-skipped",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.processedSeqs[1] = true
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "is-processed-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.isProcessedErr = errors.New("db error")
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "mark-processed-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.markProcessedErr = errors.New("mark failed")
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{unknownEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "ack-events-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{
						events: []maxadapter.AdapterEvent{unknownEvent},
						ackErr: errors.New("ack failed"),
					},
					newDeliveryStoreMock(),
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unknown-event-type-ignored",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{unknownEvent}},
					newDeliveryStoreMock(),
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-delivery-not-found-skipped",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				evNoID := maxadapter.AdapterEvent{
					Seq:  5,
					Type: maxadapter.AdapterEventMessageRead,
					// ChatID and MessageID empty → ErrDeliveryNotFound
				}
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{evNoID}},
					newDeliveryStoreMock(),
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-find-delivery-by-message-id",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryByMessageDelivery = &baseDelivery
				st.markReadResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-fallback-to-chat",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				evNoMsgID := maxadapter.AdapterEvent{
					Seq:    6,
					Type:   maxadapter.AdapterEventMessageRead,
					ChatID: "chat-abc",
					// MessageID empty → skip QueryByMessage, use QueryLatestUnreadByChat
				}
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryLatestUnreadDelivery = &baseDelivery
				st.markReadResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: openedTarget},
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{evNoMsgID}},
					st,
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-not-changed-no-update",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryByMessageDelivery = &baseDelivery
				st.markReadResult = false
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-mark-read-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryByMessageDelivery = &baseDelivery
				st.markReadErr = errors.New("db error")
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-delivery-not-found-skipped",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					newDeliveryStoreMock(),
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-not-changed-education-already-sent-no-update",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				now := time.Now().UTC()
				deliveryEdu := baseDelivery
				deliveryEdu.EducationSentAt = &now
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = deliveryEdu
				st.queryReplyDelivery = &deliveryEdu
				st.markReplyResult = false
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campNoEdu}},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-changed-sets-target-replied",
			ExpResp: campaignbus.Replied.String(),
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campNoEdu}},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				bus.ConsumeEvents(ctx)
				return tq.changedStatus.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-received-same-as-replied",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{receivedEvent}},
					newDeliveryStoreMock(),
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-education-already-sent-skipped",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				now := time.Now().UTC()
				deliveryEduSent := baseDelivery
				deliveryEduSent.EducationSentAt = &now
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = deliveryEduSent
				st.queryReplyDelivery = &deliveryEduSent
				st.markReplyResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
				}
				adMock := &adapterMock{events: []maxadapter.AdapterEvent{replyEvent}}
				bus := newBus(
					tq,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campNoEdu}},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					adMock,
					st,
				)
				err := bus.ConsumeEvents(ctx)
				// education was already sent — adapter should not have sent a second message
				return err == nil && len(adMock.sentReqs) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-sends-education",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
				}
				adMock := &adapterMock{events: []maxadapter.AdapterEvent{replyEvent}}
				bus := newBus(
					tq,
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campWithEdu}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{
						txtID:    txtAatch,
						eduTxtID: eduTxtAatch,
					}},
					&renderProviderMock{content: []byte("Education text")},
					&eventPublisherMock{},
					adMock,
					st,
				)
				err := bus.ConsumeEvents(ctx)
				return err == nil && len(adMock.sentReqs) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-education-no-text-id-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = false // not changed → skip target update, go to sendEducation
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campNoEdu}},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "find-read-delivery-query-by-message-other-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.queryByMessageErr = errors.New("unexpected db error")
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-target-query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryByMessageDelivery = &baseDelivery
				st.markReadResult = true
				bus := newBus(
					&targetQuerierMock{}, // targetID not in map → ErrTargetNotFound
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-read-change-status-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryByMessageDelivery = &baseDelivery
				st.markReadResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
					changeErr:    errors.New("status update failed"),
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{readEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-mark-replied-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyErr = errors.New("db error")
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-changed-target-query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = true
				bus := newBus(
					&targetQuerierMock{}, // targetID not in map
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-changed-target-not-eligible-for-status-change",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				failedTarget := campaignbus.Target{
					ID:         targetID,
					CampaignID: campID,
					EmployeeID: empID,
					Status:     campaignbus.Failed,
				}
				now := time.Now().UTC()
				deliveryEdu := baseDelivery
				deliveryEdu.EducationSentAt = &now
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = deliveryEdu
				st.queryReplyDelivery = &deliveryEdu
				st.markReplyResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: failedTarget},
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-store-query-after-mark-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = false
				// deliveryID not in map → QueryByID returns ErrDeliveryNotFound
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-education-campaign-query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				now := time.Now()
				deliveryNoEdu := baseDelivery
				_ = now
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = deliveryNoEdu
				st.queryReplyDelivery = &deliveryNoEdu
				st.markReplyResult = false
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{queryErr: errors.New("campaign db error")},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-education-render-text-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				eduTxtAatchLocal := attachmentbus.Attachment{
					ID:          eduTxtID,
					Type:        attachmentbus.Txt,
					Label:       label.MustParse("Education text"),
					ContentPath: file.MustParse("/content/edu.txt"),
				}
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = false
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campWithEdu}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{eduTxtID: eduTxtAatchLocal}},
					&renderProviderMock{renderErr: errors.New("render failed")},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-changed-change-status-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				now := time.Now().UTC()
				deliveryEdu := baseDelivery
				deliveryEdu.EducationSentAt = &now
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = deliveryEdu
				st.queryReplyDelivery = &deliveryEdu
				st.markReplyResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
					changeErr:    errors.New("status update failed"),
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-find-delivery-other-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				st := newDeliveryStoreMock()
				st.queryReplyErr = errors.New("unexpected db error")
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "message-replied-changed-publish-event-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				now := time.Now().UTC()
				deliveryEdu := baseDelivery
				deliveryEdu.EducationSentAt = &now
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = deliveryEdu
				st.queryReplyDelivery = &deliveryEdu
				st.markReplyResult = true
				tq := &targetQuerierMock{
					queryByIDMap: map[uuid.UUID]campaignbus.Target{targetID: sentTarget},
				}
				bus := newBus(
					tq,
					&campaignQuerierMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
					&eventPublisherMock{publishErr: errors.New("publish failed")},
					&adapterMock{events: []maxadapter.AdapterEvent{replyEvent}},
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "send-education-adapter-send-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				eduTxtAatchLocal := attachmentbus.Attachment{
					ID:          eduTxtID,
					Type:        attachmentbus.Txt,
					Label:       label.MustParse("Education text"),
					ContentPath: file.MustParse("/content/edu.txt"),
				}
				st := newDeliveryStoreMock()
				st.deliveries[deliveryID] = baseDelivery
				st.queryReplyDelivery = &baseDelivery
				st.markReplyResult = false
				adEduErr := &adapterMock{
					sendErr: errors.New("adapter error"),
					events:  []maxadapter.AdapterEvent{replyEvent},
				}
				bus := newBus(
					&targetQuerierMock{},
					&campaignQuerierMock{campaigns: map[uuid.UUID]campaignbus.Campaign{campID: campWithEdu}},
					&attachmentProviderMock{attachments: map[uuid.UUID]attachmentbus.Attachment{eduTxtID: eduTxtAatchLocal}},
					&renderProviderMock{content: []byte("Education text")},
					&eventPublisherMock{},
					adEduErr,
					st,
				)
				return bus.ConsumeEvents(ctx) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
