package deliverybus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/types/event"
)

type targetQuerier interface {
	QueryDue(ctx context.Context) ([]campaignbus.Target, error)
	ChangeStatus(ctx context.Context, t campaignbus.Target, s campaignbus.TargetStatus) error
}

type campaignQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (campaignbus.Campaign, error)
}

type employeeQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error)
}
type messageRenderer interface {
	QueryByID(ctx context.Context, id uuid.UUID) (messagebus.Message, error)
	Render(ctx context.Context, msg messagebus.Message, scope resolverbus.Scope) (string, error)
}

type Deliverer interface {
	Send(ctx context.Context, to, from, subject, body string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, e eventbus.NewEvent) error
}

type Business struct {
	targetQuerier   targetQuerier
	campaignQuerier campaignQuerier
	employeeQuerier employeeQuerier
	messageRenderer messageRenderer
	deliverer       Deliverer
	eventPub        EventPublisher
}

func NewBusiness(t targetQuerier, c campaignQuerier, e employeeQuerier, r messageRenderer, d Deliverer, ep EventPublisher) *Business {
	return &Business{targetQuerier: t, campaignQuerier: c, employeeQuerier: e, messageRenderer: r, deliverer: d, eventPub: ep}
}

func (b *Business) SendMail(ctx context.Context) []error {
	var errs []error
	targets, err := b.targetQuerier.QueryDue(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("sendmail: %w", err))
		return errs
	}

	if len(targets) == 0 {
		return nil
	}

	for _, t := range targets {
		if err := b.processTarget(ctx, t); err != nil {
			errs = append(errs, fmt.Errorf("sendmail: target[%s]: %w", t.ID, err))
			b.eventPub.Publish(ctx, eventbus.NewEvent{
				CampaignID: t.CampaignID,
				EmployeeID: t.EmployeeID,
				Type:       event.DeliveryFailed,
			})
		}
	}

	return errs
}

func (b *Business) processTarget(ctx context.Context, t campaignbus.Target) error {
	cmp, err := b.campaignQuerier.QueryByID(ctx, t.CampaignID)
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if cmp.MessageID == nil {
		return fmt.Errorf("processtarget: campaign[%s] has no message", cmp.ID)
	}

	if cmp.Status != campaignbus.Active {
		return fmt.Errorf("processtarget: campaign[%s] is not active: %s", cmp.ID, cmp.Status)
	}

	emp, err := b.employeeQuerier.QueryByID(ctx, t.EmployeeID)
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	msg, err := b.messageRenderer.QueryByID(ctx, *cmp.MessageID)
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	html, err := b.messageRenderer.Render(ctx, msg, resolverbus.Scope{
		EmployeeID: emp.ID,
	})
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if err := b.deliverer.Send(ctx, emp.Email.Address, msg.FromEmail.Address, msg.Subject.String(), html); err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if err := b.targetQuerier.ChangeStatus(ctx, t, campaignbus.Sent); err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if err := b.eventPub.Publish(ctx, eventbus.NewEvent{
		CampaignID: t.CampaignID,
		EmployeeID: t.EmployeeID,
		Type:       event.MessageSent,
	}); err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	return nil
}
