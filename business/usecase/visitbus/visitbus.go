package visitbus

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
)

type targetQuerier interface {
	QueryByToken(ctx context.Context, token string) (campaignbus.Target, error)
	ChangeStatus(ctx context.Context, t campaignbus.Target, s campaignbus.TargetStatus) error
}

type campaignQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error)
}

type landingRenderer interface {
	QueryByID(ctx context.Context, id uuid.UUID) (landingbus.Landing, error)
}

type eventPublisher interface {
	Publish(ctx context.Context, e eventbus.NewEvent) error
}

type AttachmentProvider interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type RenderProvider interface {
	Render(ctx context.Context, atch attachmentbus.Attachment, targetID uuid.UUID) ([]byte, error)
}

type TargetData struct {
	IpAddress netip.Addr
	UserAgent string
	Referer   string
	Token     string
}

type Business struct {
	targetQuerier   targetQuerier
	campaignQuerier campaignQuerier
	landingRenderer landingRenderer
	eventPub        eventPublisher
	attachmentBus   AttachmentProvider
	renderBus       RenderProvider
}

func NewBusiness(t targetQuerier, c campaignQuerier, l landingRenderer, ep eventPublisher, a AttachmentProvider, r RenderProvider) *Business {
	return &Business{
		targetQuerier:   t,
		campaignQuerier: c,
		landingRenderer: l,
		eventPub:        ep,
		attachmentBus:   a,
		renderBus:       r,
	}
}

func (b *Business) Serve(ctx context.Context, td TargetData) (string, error) {
	target, err := b.targetQuerier.QueryByToken(ctx, td.Token)
	if err != nil {
		return "", fmt.Errorf("serve: query target: %w", err)
	}

	cmp, err := b.campaignQuerier.QueryByID(ctx, target.CampaignID)
	if err != nil {
		return "", fmt.Errorf("serve: query campaign: %w", err)
	}

	if cmp.Status != campaignbus.Active {
		return "", fmt.Errorf("serve: campaign is not active")
	}

	if cmp.LandingID == nil {
		return "", fmt.Errorf("serve: campaign has no landing page")
	}

	landing, err := b.landingRenderer.QueryByID(ctx, *cmp.LandingID)
	if err != nil {
		return "", fmt.Errorf("serve: query landing: %w", err)
	}

	if !landing.HtmlBodyID.Valid {
		return "", fmt.Errorf("serve: landing[%s] has no HTML body", landing.ID)
	}

	atch, err := b.attachmentBus.QueryByID(ctx, landing.HtmlBodyID.UUID)
	if err != nil {
		return "", fmt.Errorf("serve: query attachment: %w", err)
	}

	html, err := b.renderBus.Render(ctx, atch, target.ID)
	if err != nil {
		return "", fmt.Errorf("serve: render landing: %w", err)
	}

	if err := b.eventPub.Publish(ctx, eventbus.NewEvent{
		CampaignID: target.CampaignID,
		EmployeeID: target.EmployeeID,
		Type:       eventbus.LinkOpened,
		IPAddress:  td.IpAddress,
		UserAgent:  td.UserAgent,
		Referer:    td.Referer,
	}); err != nil {
		return "", fmt.Errorf("serve: publish event: %w", err)
	}

	if target.Status == campaignbus.Sent || target.Status == campaignbus.Pending || target.Status == campaignbus.Opened {
		if err := b.targetQuerier.ChangeStatus(ctx, target, campaignbus.Clicked); err != nil {
			return "", fmt.Errorf("serve: change target status: %w", err)
		}
	}

	return string(html), nil
}

func (b *Business) TrackOpen(ctx context.Context, td TargetData) error {
	target, err := b.targetQuerier.QueryByToken(ctx, td.Token)
	if err != nil {
		return fmt.Errorf("trackopen: query target: %w", err)
	}

	cmp, err := b.campaignQuerier.QueryByID(ctx, target.CampaignID)
	if err != nil {
		return fmt.Errorf("trackopen: query campaign: %w", err)
	}

	if cmp.Status != campaignbus.Active {
		return fmt.Errorf("trackopen: campaign is not active")
	}

	if err := b.eventPub.Publish(ctx, eventbus.NewEvent{
		CampaignID: target.CampaignID,
		EmployeeID: target.EmployeeID,
		Type:       eventbus.EmailOpened,
		IPAddress:  td.IpAddress,
		UserAgent:  td.UserAgent,
		Referer:    td.Referer,
	}); err != nil {
		return fmt.Errorf("trackopen: publish event: %w", err)
	}

	if target.Status == campaignbus.Sent {
		if err := b.targetQuerier.ChangeStatus(ctx, target, campaignbus.Opened); err != nil {
			return fmt.Errorf("trackopen: change target status: %w", err)
		}
	}

	return nil
}

func (b *Business) Submit(ctx context.Context, data TargetData) (string, error) {
	target, err := b.targetQuerier.QueryByToken(ctx, data.Token)
	if err != nil {
		return "", fmt.Errorf("submit: %w", err)
	}

	cmp, err := b.campaignQuerier.QueryByID(ctx, target.CampaignID)
	if err != nil {
		return "", fmt.Errorf("submit: query campaign: %w", err)
	}

	if cmp.Status != campaignbus.Active {
		return "", fmt.Errorf("submit: campaign is not active")
	}

	if cmp.EducationID == nil {
		return "", fmt.Errorf("submit: campaign has no education")
	}

	education, err := b.landingRenderer.QueryByID(ctx, *cmp.EducationID)
	if err != nil {
		return "", fmt.Errorf("submit: query education: %w", err)
	}

	if !education.HtmlBodyID.Valid {
		return "", fmt.Errorf("submit: education[%s] has no HTML body", education.ID)
	}

	atch, err := b.attachmentBus.QueryByID(ctx, education.HtmlBodyID.UUID)
	if err != nil {
		return "", fmt.Errorf("submit: query attachment: %w", err)
	}

	html, err := b.renderBus.Render(ctx, atch, target.ID)
	if err != nil {
		return "", fmt.Errorf("submit: render landing: %w", err)
	}

	if err := b.eventPub.Publish(ctx, eventbus.NewEvent{
		CampaignID: target.CampaignID,
		EmployeeID: target.EmployeeID,
		Type:       eventbus.DataSent,
		IPAddress:  data.IpAddress,
		UserAgent:  data.UserAgent,
		Referer:    data.Referer,
	}); err != nil {
		return "", fmt.Errorf("submit: publish event: %w", err)
	}

	if target.Status == campaignbus.Clicked {
		if err := b.targetQuerier.ChangeStatus(ctx, target, campaignbus.Submitted); err != nil {
			return "", fmt.Errorf("submit: change target status: %w", err)
		}
	}

	return string(html), nil
}
