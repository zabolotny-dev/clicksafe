package visitbus_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/database"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/usecase/visitbus"
)

// =============================================================================
// Mocks

type targetQuerierMock struct {
	target         campaignbus.Target
	queryErr       error
	changeStatusErr error
	statusChanged  campaignbus.TargetStatus
}

func (m *targetQuerierMock) QueryByToken(_ context.Context, _ string) (campaignbus.Target, error) {
	return m.target, m.queryErr
}

func (m *targetQuerierMock) ChangeStatus(_ context.Context, _ campaignbus.Target, s campaignbus.TargetStatus) error {
	m.statusChanged = s
	return m.changeStatusErr
}

type campaignQuerierMock struct {
	campaign campaignbus.Campaign
	queryErr error
}

func (m *campaignQuerierMock) QueryByID(_ context.Context, _ uuid.UUID) (campaignbus.Campaign, error) {
	return m.campaign, m.queryErr
}

type landingRendererMock struct {
	landing  landingbus.Landing
	queryErr error
}

func (m *landingRendererMock) QueryByID(_ context.Context, _ uuid.UUID) (landingbus.Landing, error) {
	return m.landing, m.queryErr
}

type eventPublisherMock struct {
	publishErr error
	published  []eventbus.NewEvent
}

func (m *eventPublisherMock) Publish(_ context.Context, e eventbus.NewEvent) error {
	m.published = append(m.published, e)
	return m.publishErr
}

type attachmentProviderMock struct {
	attachment attachmentbus.Attachment
	queryErr   error
}

func (m *attachmentProviderMock) QueryByID(_ context.Context, _ uuid.UUID) (attachmentbus.Attachment, error) {
	return m.attachment, m.queryErr
}

type renderProviderMock struct {
	content   []byte
	renderErr error
}

func (m *renderProviderMock) Render(_ context.Context, _ attachmentbus.Attachment, _ uuid.UUID) ([]byte, error) {
	return m.content, m.renderErr
}

// =============================================================================

func buildBus(
	tq *targetQuerierMock,
	cq *campaignQuerierMock,
	lr *landingRendererMock,
	ep *eventPublisherMock,
	ap *attachmentProviderMock,
	rp *renderProviderMock,
) *visitbus.Business {
	return visitbus.NewBusiness(tq, cq, lr, ep, ap, rp, database.Noop)
}

func Test_Visit(t *testing.T) {
	t.Parallel()

	unittest.Run(t, testServe(), "serve")
	unittest.Run(t, testTrackOpen(), "trackopen")
	unittest.Run(t, testSubmit(), "submit")
}

// =============================================================================

func testServe() []unittest.Table {
	campID := uuid.New()
	landID := uuid.New()
	htmlBodyID := uuid.New()
	targetID := uuid.New()
	ip := netip.MustParseAddr("1.2.3.4")

	htmlContent := []byte("<html><body>Phish</body></html>")

	target := campaignbus.Target{ID: targetID, CampaignID: campID, Status: campaignbus.Sent}
	activeCamp := campaignbus.Campaign{ID: campID, Status: campaignbus.Active, LandingID: &landID}
	landing := landingbus.Landing{ID: landID, HtmlBodyID: uuid.NullUUID{UUID: htmlBodyID, Valid: true}}
	atch := attachmentbus.Attachment{ID: htmlBodyID, Type: attachmentbus.Html, ContentPath: file.MustParse("/content/landing.html")}

	td := visitbus.TargetData{Token: "tok", IpAddress: ip, UserAgent: "UA", Referer: "ref"}

	// Success path
	ep := &eventPublisherMock{}
	tq := &targetQuerierMock{target: target}
	busOK := buildBus(tq, &campaignQuerierMock{campaign: activeCamp}, &landingRendererMock{landing: landing}, ep, &attachmentProviderMock{attachment: atch}, &renderProviderMock{content: htmlContent})

	// Token query error
	busTokenErr := buildBus(
		&targetQuerierMock{queryErr: campaignbus.ErrTargetNotFound},
		&campaignQuerierMock{campaign: activeCamp},
		&landingRendererMock{landing: landing},
		&eventPublisherMock{},
		&attachmentProviderMock{attachment: atch},
		&renderProviderMock{content: htmlContent},
	)

	// Campaign not active
	inactiveCamp := campaignbus.Campaign{ID: campID, Status: campaignbus.Draft, LandingID: &landID}
	busInactive := buildBus(
		&targetQuerierMock{target: target},
		&campaignQuerierMock{campaign: inactiveCamp},
		&landingRendererMock{landing: landing},
		&eventPublisherMock{},
		&attachmentProviderMock{attachment: atch},
		&renderProviderMock{content: htmlContent},
	)

	// No landing
	campNoLanding := campaignbus.Campaign{ID: campID, Status: campaignbus.Active}
	busNoLanding := buildBus(
		&targetQuerierMock{target: target},
		&campaignQuerierMock{campaign: campNoLanding},
		&landingRendererMock{landing: landing},
		&eventPublisherMock{},
		&attachmentProviderMock{attachment: atch},
		&renderProviderMock{content: htmlContent},
	)

	// Landing no HTML body
	landingNoHTML := landingbus.Landing{ID: landID}
	busLandingNoHTML := buildBus(
		&targetQuerierMock{target: target},
		&campaignQuerierMock{campaign: activeCamp},
		&landingRendererMock{landing: landingNoHTML},
		&eventPublisherMock{},
		&attachmentProviderMock{attachment: atch},
		&renderProviderMock{content: htmlContent},
	)

	return []unittest.Table{
		{
			Name:    "success-returns-html",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				html, err := busOK.Serve(ctx, td)
				if err != nil {
					return err
				}
				return len(html) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "success-publishes-event",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busOK.Serve(ctx, td)
				if err != nil {
					return err
				}
				return len(ep.published) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "success-changes-status-to-clicked",
			ExpResp: campaignbus.Clicked.String(),
			ExcFunc: func(ctx context.Context) any {
				_, err := busOK.Serve(ctx, td)
				if err != nil {
					return err
				}
				return tq.statusChanged.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "token-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busTokenErr.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "inactive-campaign-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busInactive.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-landing-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoLanding.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "landing-no-html-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busLandingNoHTML.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busAttErr := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{campaign: activeCamp},
					&landingRendererMock{landing: landing},
					&eventPublisherMock{},
					&attachmentProviderMock{queryErr: errors.New("attachment not found")},
					&renderProviderMock{content: htmlContent},
				)
				_, err := busAttErr.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "render-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busRenderErr := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{campaign: activeCamp},
					&landingRendererMock{landing: landing},
					&eventPublisherMock{},
					&attachmentProviderMock{attachment: atch},
					&renderProviderMock{renderErr: errors.New("render failed")},
				)
				_, err := busRenderErr.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busCampErr := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{queryErr: errors.New("db error")},
					&landingRendererMock{},
					&eventPublisherMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
				)
				_, err := busCampErr.Serve(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testTrackOpen() []unittest.Table {
	campID := uuid.New()
	targetID := uuid.New()
	ip := netip.MustParseAddr("1.2.3.4")

	target := campaignbus.Target{ID: targetID, CampaignID: campID, Status: campaignbus.Sent}
	activeCamp := campaignbus.Campaign{ID: campID, Status: campaignbus.Active}
	inactiveCamp := campaignbus.Campaign{ID: campID, Status: campaignbus.Draft}

	td := visitbus.TargetData{Token: "tok", IpAddress: ip}

	ep := &eventPublisherMock{}
	tq := &targetQuerierMock{target: target}
	busOK := buildBus(tq, &campaignQuerierMock{campaign: activeCamp}, &landingRendererMock{}, ep, &attachmentProviderMock{}, &renderProviderMock{})

	busInactive := buildBus(
		&targetQuerierMock{target: target},
		&campaignQuerierMock{campaign: inactiveCamp},
		&landingRendererMock{},
		&eventPublisherMock{},
		&attachmentProviderMock{},
		&renderProviderMock{},
	)

	busTokenErr := buildBus(
		&targetQuerierMock{queryErr: errors.New("not found")},
		&campaignQuerierMock{},
		&landingRendererMock{},
		&eventPublisherMock{},
		&attachmentProviderMock{},
		&renderProviderMock{},
	)

	return []unittest.Table{
		{
			Name:    "success-publishes-email-opened",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				if err := busOK.TrackOpen(ctx, td); err != nil {
					return err
				}
				for _, e := range ep.published {
					if e.Type == eventbus.EmailOpened {
						return true
					}
				}
				return false
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "success-changes-status-to-opened",
			ExpResp: campaignbus.Opened.String(),
			ExcFunc: func(ctx context.Context) any {
				if err := busOK.TrackOpen(ctx, td); err != nil {
					return err
				}
				return tq.statusChanged.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "inactive-campaign-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busInactive.TrackOpen(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "token-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busTokenErr.TrackOpen(ctx, td) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busCampErr := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{queryErr: errors.New("db error")},
					&landingRendererMock{},
					&eventPublisherMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
				)
				return busCampErr.TrackOpen(ctx, td) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "non-sent-target-no-status-change",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				clickedTarget := campaignbus.Target{ID: uuid.New(), CampaignID: campID, Status: campaignbus.Clicked}
				tqClicked := &targetQuerierMock{target: clickedTarget}
				busClicked := buildBus(tqClicked, &campaignQuerierMock{campaign: activeCamp}, &landingRendererMock{}, &eventPublisherMock{}, &attachmentProviderMock{}, &renderProviderMock{})
				err := busClicked.TrackOpen(ctx, td)
				// Clicked target should not change status to Opened
				return err == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func testSubmit() []unittest.Table {
	campID := uuid.New()
	eduID := uuid.New()
	htmlBodyID := uuid.New()
	ip := netip.MustParseAddr("1.2.3.4")

	htmlContent := []byte("<html><body>Education</body></html>")

	target := campaignbus.Target{ID: uuid.New(), CampaignID: campID, Status: campaignbus.Clicked}
	activeCamp := campaignbus.Campaign{ID: campID, Status: campaignbus.Active, EducationID: &eduID}
	education := landingbus.Landing{ID: eduID, HtmlBodyID: uuid.NullUUID{UUID: htmlBodyID, Valid: true}, Label: label.MustParse("Education page")}
	atch := attachmentbus.Attachment{ID: htmlBodyID, Type: attachmentbus.Html, ContentPath: file.MustParse("/content/edu.html")}

	td := visitbus.TargetData{Token: "tok", IpAddress: ip}

	ep := &eventPublisherMock{}
	tq := &targetQuerierMock{target: target}
	busOK := buildBus(tq, &campaignQuerierMock{campaign: activeCamp}, &landingRendererMock{landing: education}, ep, &attachmentProviderMock{attachment: atch}, &renderProviderMock{content: htmlContent})

	campNoEdu := campaignbus.Campaign{ID: campID, Status: campaignbus.Active}
	busNoEdu := buildBus(
		&targetQuerierMock{target: target},
		&campaignQuerierMock{campaign: campNoEdu},
		&landingRendererMock{},
		&eventPublisherMock{},
		&attachmentProviderMock{},
		&renderProviderMock{},
	)

	eduNoHTML := landingbus.Landing{ID: eduID, Label: label.MustParse("Edu no html")}
	busEduNoHTML := buildBus(
		&targetQuerierMock{target: target},
		&campaignQuerierMock{campaign: activeCamp},
		&landingRendererMock{landing: eduNoHTML},
		&eventPublisherMock{},
		&attachmentProviderMock{},
		&renderProviderMock{},
	)

	return []unittest.Table{
		{
			Name:    "success-returns-html",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				html, err := busOK.Submit(ctx, td)
				if err != nil {
					return err
				}
				return len(html) > 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "success-changes-status-to-submitted",
			ExpResp: campaignbus.Submitted.String(),
			ExcFunc: func(ctx context.Context) any {
				_, err := busOK.Submit(ctx, td)
				if err != nil {
					return err
				}
				return tq.statusChanged.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "no-education-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busNoEdu.Submit(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "education-no-html-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busEduNoHTML.Submit(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "campaign-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busCampErr := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{queryErr: errors.New("db error")},
					&landingRendererMock{},
					&eventPublisherMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
				)
				_, err := busCampErr.Submit(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "inactive-campaign-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				inactiveCampNoEdu := campaignbus.Campaign{ID: campID, Status: campaignbus.Draft, EducationID: &eduID}
				busInactiveSubmit := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{campaign: inactiveCampNoEdu},
					&landingRendererMock{},
					&eventPublisherMock{},
					&attachmentProviderMock{},
					&renderProviderMock{},
				)
				_, err := busInactiveSubmit.Submit(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "attachment-query-error-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				busAttErr := buildBus(
					&targetQuerierMock{target: target},
					&campaignQuerierMock{campaign: activeCamp},
					&landingRendererMock{landing: education},
					&eventPublisherMock{},
					&attachmentProviderMock{queryErr: errors.New("attachment error")},
					&renderProviderMock{content: htmlContent},
				)
				_, err := busAttErr.Submit(ctx, td)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
