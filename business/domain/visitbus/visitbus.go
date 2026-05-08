package visitbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/types/event"
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
	Render(ctx context.Context, landing landingbus.Landing, scope resolverbus.Scope) (string, error)
}

type eventPublisher interface {
	Publish(ctx context.Context, e eventbus.NewEvent) error
}

type Business struct {
	targetQuerier   targetQuerier
	campaignQuerier campaignQuerier
	landingRenderer landingRenderer
	eventPub        eventPublisher
}

func NewBusiness(t targetQuerier, c campaignQuerier, l landingRenderer, ep eventPublisher) *Business {
	return &Business{
		targetQuerier:   t,
		campaignQuerier: c,
		landingRenderer: l,
		eventPub:        ep,
	}
}

func (b *Business) Serve(ctx context.Context, token string) (string, error) {
	target, err := b.targetQuerier.QueryByToken(ctx, token)
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

	html, err := b.landingRenderer.Render(ctx, landing, resolverbus.Scope{
		TargetID: target.ID,
	})
	if err != nil {
		return "", fmt.Errorf("serve: render landing: %w", err)
	}

	if err := b.eventPub.Publish(ctx, eventbus.NewEvent{
		CampaignID: target.CampaignID,
		EmployeeID: target.EmployeeID,
		Type:       event.LINK_OPENED,
	}); err != nil {
		return "", fmt.Errorf("serve: publish event: %w", err)
	}

	if target.Status == campaignbus.Sent {
		if err := b.targetQuerier.ChangeStatus(ctx, target, campaignbus.Opened); err != nil {
			return "", fmt.Errorf("serve: change target status: %w", err)
		}
	}

	return html, nil
}
