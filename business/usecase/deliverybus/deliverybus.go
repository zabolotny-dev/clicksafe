package deliverybus

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"regexp"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/eventbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

var closeBodyTag = regexp.MustCompile(`(?i)</body>`)

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
}

type Deliverer interface {
	Send(ctx context.Context, to, from, subject, body string, attachments map[string][]byte) error
}

type EventPublisher interface {
	Publish(ctx context.Context, e eventbus.NewEvent) error
}

type AttachmentProvider interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type RenderProvider interface {
	Render(ctx context.Context, atch attachmentbus.Attachment, targetID uuid.UUID) ([]byte, error)
}

type Business struct {
	targetQuerier      targetQuerier
	campaignQuerier    campaignQuerier
	employeeQuerier    employeeQuerier
	messageRenderer    messageRenderer
	attachmentProvider AttachmentProvider
	deliverer          Deliverer
	eventPub           EventPublisher
	renderProvider     RenderProvider
}

func NewBusiness(t targetQuerier, c campaignQuerier, e employeeQuerier, r messageRenderer, ap AttachmentProvider, d Deliverer, ep EventPublisher, rp RenderProvider) *Business {
	return &Business{targetQuerier: t, campaignQuerier: c, employeeQuerier: e, messageRenderer: r, attachmentProvider: ap, deliverer: d, eventPub: ep, renderProvider: rp}
}

func (b *Business) SendMail(ctx context.Context) []error {
	var errs []error
	targets, err := b.targetQuerier.QueryDue(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("sendmail: %w", err))
	}

	if len(targets) == 0 {
		return errs
	}

	for _, t := range targets {
		if err := b.processTarget(ctx, t); err != nil {
			errs = append(errs, fmt.Errorf("sendmail: target[%s]: %w", t.ID, err))
			b.eventPub.Publish(ctx, eventbus.NewEvent{
				CampaignID: t.CampaignID,
				EmployeeID: t.EmployeeID,
				Type:       eventbus.DeliveryFailed,
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

	if !msg.HtmlBodyID.Valid {
		return fmt.Errorf("processtarget: message[%s] has no html body", msg.ID)
	}

	atch, err := b.attachmentProvider.QueryByID(ctx, msg.HtmlBodyID.UUID)
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if atch.Type != attachmentbus.Html {
		return fmt.Errorf("processtarget: attachment[%s] is not html", atch.ID)
	}

	rendered, err := b.renderProvider.Render(ctx, atch, t.ID)
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	html, err := addOpenTrackingPixel(string(rendered), cmp, t)
	if err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	attachments := make(map[string][]byte)
	for _, a := range msg.AttachmentIDs {
		atch, err := b.attachmentProvider.QueryByID(ctx, a)
		if err != nil {
			return fmt.Errorf("processtarget: %w", err)
		}
		atchr, err := b.renderProvider.Render(ctx, atch, t.ID)
		if err != nil {
			return fmt.Errorf("processtarget: %w", err)
		}
		attachments[atch.Label.String()+atch.Type.String()] = atchr
	}

	if err := b.deliverer.Send(ctx, emp.Email.Address, msg.FromEmail.Address, msg.Subject.String(), html, attachments); err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if err := b.targetQuerier.ChangeStatus(ctx, t, campaignbus.Sent); err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	if err := b.eventPub.Publish(ctx, eventbus.NewEvent{
		CampaignID: t.CampaignID,
		EmployeeID: t.EmployeeID,
		Type:       eventbus.MessageSent,
	}); err != nil {
		return fmt.Errorf("processtarget: %w", err)
	}

	return nil
}

func addOpenTrackingPixel(body string, cmp campaignbus.Campaign, t campaignbus.Target) (string, error) {
	if cmp.Domain.IsEmpty() {
		return "", campaignbus.ErrDomainRequired
	}

	pixelURL := cmp.Domain.String() + "/" + url.PathEscape(t.Token) + ".gif"
	pixel := fmt.Sprintf(
		`<img src="%s" width="1" height="1" alt="" style="display:none;width:1px;height:1px;border:0;outline:0;" />`,
		html.EscapeString(pixelURL),
	)

	closeBodyMatches := closeBodyTag.FindAllStringIndex(body, -1)
	if len(closeBodyMatches) == 0 {
		return body + "\n" + pixel, nil
	}

	closeBodyIdx := closeBodyMatches[len(closeBodyMatches)-1][0]
	return body[:closeBodyIdx] + pixel + "\n" + body[closeBodyIdx:], nil
}
