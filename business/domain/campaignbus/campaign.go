package campaignbus

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Campaign struct {
	ID                 uuid.UUID
	Type               CampaignType
	MessageID          *uuid.UUID
	LandingID          *uuid.UUID
	EducationID        *uuid.UUID
	MaxEducationTextID uuid.NullUUID
	Label              label.Label
	Domain             domain.Domain
	Status             CampaignStatus
	DateRange          date.Null
	Attributes         map[string]string
}

type NewCampaign struct {
	Type               CampaignType
	MessageID          *uuid.UUID
	LandingID          *uuid.UUID
	EducationID        *uuid.UUID
	MaxEducationTextID uuid.NullUUID
	Label              label.Label
	Domain             domain.Domain
	DateRange          date.Null
	Attributes         map[string]string
}

type UpdateCampaign struct {
	Type               *CampaignType
	MessageID          *uuid.UUID
	LandingID          *uuid.UUID
	EducationID        *uuid.UUID
	MaxEducationTextID *uuid.NullUUID
	Label              *label.Label
	Domain             *domain.Domain
	DateRange          *date.Null
	Attributes         *map[string]string
}

type TargetMissingVars struct {
	TargetID   uuid.UUID
	EmployeeID uuid.UUID
	Vars       []string
}

type MessageQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (messagebus.Message, error)
}

type LandingQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (landingbus.Landing, error)
}

type EmployeeQuerier interface {
	QueryByID(ctx context.Context, id uuid.UUID) (employeebus.Employee, error)
}

type VarsValidator interface {
	Validate(ctx context.Context, campaign Campaign, targets []Target, requiredVars []string) ([]TargetMissingVars, error)
}

type AttachmentProvider interface {
	QueryByID(ctx context.Context, id uuid.UUID) (attachmentbus.Attachment, error)
}

type CampaignStorer interface {
	Save(ctx context.Context, campaign Campaign) error
	Update(ctx context.Context, campaign Campaign) error
	QueryByID(ctx context.Context, id uuid.UUID) (Campaign, error)
	Query(ctx context.Context, filter CampaignQueryFilter, orderBy order.By, page page.Page) ([]Campaign, error)
	Delete(ctx context.Context, campaign Campaign) error
	Count(ctx context.Context, filter CampaignQueryFilter) (int, error)
	QueryExpired(ctx context.Context) ([]Campaign, error)
}

func (b *CampaignBusiness) Save(ctx context.Context, campaign NewCampaign) (Campaign, error) {
	cmpType := campaign.Type
	if cmpType == (CampaignType{}) {
		cmpType = EmailCampaign
	}

	cmp := Campaign{
		ID:                 uuid.New(),
		Type:               cmpType,
		MessageID:          campaign.MessageID,
		LandingID:          campaign.LandingID,
		EducationID:        campaign.EducationID,
		MaxEducationTextID: campaign.MaxEducationTextID,
		Label:              campaign.Label,
		Domain:             campaign.Domain,
		Status:             Draft,
		DateRange:          campaign.DateRange,
		Attributes:         campaign.Attributes,
	}

	if err := b.campaignStorer.Save(ctx, cmp); err != nil {
		if errors.Is(err, ErrMessageNotFound) && campaign.MessageID != nil {
			return Campaign{}, fmt.Errorf("save: messageID[%s]: %w", *campaign.MessageID, err)
		}
		if errors.Is(err, ErrLandingNotFound) && campaign.LandingID != nil {
			return Campaign{}, fmt.Errorf("save: landingID[%s]: %w", *campaign.LandingID, err)
		}
		if errors.Is(err, ErrEducationNotFound) && campaign.EducationID != nil {
			return Campaign{}, fmt.Errorf("save: educationID[%s]: %w", *campaign.EducationID, err)
		}
		return Campaign{}, fmt.Errorf("save: %w", err)
	}

	return cmp, nil
}

func (b *CampaignBusiness) Update(ctx context.Context, cmp Campaign, upd UpdateCampaign) (Campaign, error) {
	if upd.Type != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change type in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.Type = *upd.Type
	}

	if upd.MessageID != nil {
		cmp.MessageID = upd.MessageID
	}

	if upd.LandingID != nil {
		cmp.LandingID = upd.LandingID
	}

	if upd.Label != nil {
		cmp.Label = *upd.Label
	}

	if upd.Domain != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change domain in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.Domain = *upd.Domain
	}

	if upd.DateRange != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change date in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.DateRange = *upd.DateRange
	}

	if upd.EducationID != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change education in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.EducationID = upd.EducationID
	}

	if upd.MaxEducationTextID != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change max education in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.MaxEducationTextID = *upd.MaxEducationTextID
	}

	if upd.Attributes != nil {
		if cmp.Status != Draft && cmp.Status != Paused {
			return Campaign{}, fmt.Errorf("update: %w: cannot change attributes in %s status", ErrCampaignLocked, cmp.Status)
		}
		cmp.Attributes = *upd.Attributes
	}

	if err := b.campaignStorer.Update(ctx, cmp); err != nil {
		if errors.Is(err, ErrMessageNotFound) && upd.MessageID != nil {
			return Campaign{}, fmt.Errorf("update: messageID[%s]: %w", *upd.MessageID, err)
		}
		if errors.Is(err, ErrLandingNotFound) && upd.LandingID != nil {
			return Campaign{}, fmt.Errorf("update: landingID[%s]: %w", *upd.LandingID, err)
		}
		if errors.Is(err, ErrEducationNotFound) && upd.EducationID != nil {
			return Campaign{}, fmt.Errorf("update: educationID[%s]: %w", *upd.EducationID, err)
		}
		return Campaign{}, fmt.Errorf("update: %w", err)
	}

	return cmp, nil
}

func (b *CampaignBusiness) QueryByID(ctx context.Context, id uuid.UUID) (Campaign, error) {
	cmp, err := b.campaignStorer.QueryByID(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("query: campaignID[%s]: %w", id, err)
	}

	return cmp, nil
}

func (b *CampaignBusiness) Query(ctx context.Context, filter CampaignQueryFilter, orderBy order.By, page page.Page) ([]Campaign, error) {
	camps, err := b.campaignStorer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return camps, nil
}

func (b *CampaignBusiness) Delete(ctx context.Context, campaign Campaign) error {
	if err := b.campaignStorer.Delete(ctx, campaign); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *CampaignBusiness) Start(ctx context.Context, campaign Campaign) (Campaign, error) {
	if !isValidCampaignTransition(campaign.Status, Active) {
		return Campaign{}, fmt.Errorf("start: %w: cannot move from %s to %s", ErrInvalidStatusTransition, campaign.Status, Active)
	}

	if !campaign.DateRange.Valid() {
		return Campaign{}, fmt.Errorf("start: %w", ErrDateRangeRequired)
	}

	if campaign.DateRange.Range().IsExpired(time.Now().UTC()) {
		return Campaign{}, fmt.Errorf("start: %w", ErrDateRangeExpired)
	}

	targets, err := b.targetStorer.Query(ctx, TargetQueryFilter{CampaignID: &campaign.ID})
	if err != nil {
		return Campaign{}, fmt.Errorf("start: query targets: %w", err)
	}

	if len(targets) == 0 {
		return Campaign{}, fmt.Errorf("start: %w", ErrTargetsRequired)
	}

	var noScheduleTargets []Target
	for _, t := range targets {
		if t.ScheduledAt == nil {
			noScheduleTargets = append(noScheduleTargets, t)
		}
	}

	if len(noScheduleTargets) > 0 {
		ids := make([]uuid.UUID, len(noScheduleTargets))
		for i, t := range noScheduleTargets {
			ids[i] = t.ID
		}
		return Campaign{}, fmt.Errorf("start: %w", &ErrUnscheduledTargets{TargetIDs: ids})
	}

	switch campaign.Type {
	case EmailCampaign:
		if err := b.validateEmailStart(ctx, campaign, targets); err != nil {
			return Campaign{}, err
		}
	case MaxCampaign:
		if err := b.validateMaxStart(ctx, campaign, targets); err != nil {
			return Campaign{}, err
		}
	default:
		return Campaign{}, fmt.Errorf("start: invalid campaign type: %s", campaign.Type.String())
	}

	return b.changeStatus(ctx, campaign, Active, "start")
}

func (b *CampaignBusiness) validateEmailStart(ctx context.Context, campaign Campaign, targets []Target) error {
	if campaign.MessageID == nil {
		return fmt.Errorf("start: %w", ErrMessageRequired)
	}

	if campaign.LandingID == nil {
		return fmt.Errorf("start: %w", ErrLandingRequired)
	}

	if campaign.EducationID == nil {
		return fmt.Errorf("start: %w", ErrEducationRequired)
	}

	if campaign.Domain.IsEmpty() {
		return fmt.Errorf("start: %w", ErrDomainRequired)
	}

	vars := make(map[string]struct{})
	message, err := b.messageProvider.QueryByID(ctx, *campaign.MessageID)
	if err != nil {
		return fmt.Errorf("start: query message: %w", err)
	}

	if message.Type != messagebus.EmailMessage {
		return fmt.Errorf("start: %w", ErrMessageTypeMismatch)
	}

	if !message.HtmlBodyID.Valid {
		return fmt.Errorf("start: %w", ErrMessageHTMLRequired)
	}

	landing, err := b.landingProvider.QueryByID(ctx, *campaign.LandingID)
	if err != nil {
		return fmt.Errorf("start: query landing: %w", err)
	}

	if !landing.HtmlBodyID.Valid {
		return fmt.Errorf("start: %w", ErrLandingHTMLRequired)
	}

	education, err := b.landingProvider.QueryByID(ctx, *campaign.EducationID)
	if err != nil {
		return fmt.Errorf("start: query education: %w", err)
	}

	if !education.HtmlBodyID.Valid {
		return fmt.Errorf("start: %w", ErrEducationHTMLRequired)
	}

	ids := slices.Concat(
		[]uuid.UUID{message.HtmlBodyID.UUID, landing.HtmlBodyID.UUID, education.HtmlBodyID.UUID},
		message.AttachmentIDs,
	)

	for _, id := range ids {
		att, err := b.attachmentProvider.QueryByID(ctx, id)
		if err != nil {
			return fmt.Errorf("start: query attachment: %w", err)
		}
		for _, v := range att.RequiredVars {
			vars[v] = struct{}{}
		}
	}

	var reqVars []string
	for k := range vars {
		reqVars = append(reqVars, k)
	}
	sort.Strings(reqVars)

	res, err := b.varsValidator.Validate(ctx, campaign, targets, reqVars)
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	if len(res) > 0 {
		return fmt.Errorf("start: %w", &ErrTargetsMissingVars{Targets: res})
	}

	return nil
}

func (b *CampaignBusiness) validateMaxStart(ctx context.Context, campaign Campaign, targets []Target) error {
	if campaign.MessageID == nil {
		return fmt.Errorf("start: %w", ErrMessageRequired)
	}

	if !campaign.MaxEducationTextID.Valid {
		return fmt.Errorf("start: %w", ErrMaxEducationRequired)
	}

	message, err := b.messageProvider.QueryByID(ctx, *campaign.MessageID)
	if err != nil {
		return fmt.Errorf("start: query message: %w", err)
	}

	if message.Type != messagebus.MaxMessage {
		return fmt.Errorf("start: %w", ErrMessageTypeMismatch)
	}

	if !message.MaxAccountID.Valid {
		return fmt.Errorf("start: %w", messagebus.ErrMaxAccountRequired)
	}

	if !message.TextBodyID.Valid {
		return fmt.Errorf("start: %w", messagebus.ErrTextBodyRequired)
	}

	textBody, err := b.attachmentProvider.QueryByID(ctx, message.TextBodyID.UUID)
	if err != nil {
		return fmt.Errorf("start: query max text: %w", err)
	}
	if textBody.Type != attachmentbus.Txt {
		return fmt.Errorf("start: %w", messagebus.ErrInvalidAttachment)
	}

	educationText, err := b.attachmentProvider.QueryByID(ctx, campaign.MaxEducationTextID.UUID)
	if err != nil {
		return fmt.Errorf("start: query max education: %w", err)
	}
	if educationText.Type != attachmentbus.Txt {
		return fmt.Errorf("start: %w", ErrMaxEducationTXTRequired)
	}

	vars := make(map[string]struct{})
	for _, v := range slices.Concat(textBody.RequiredVars, educationText.RequiredVars) {
		vars[v] = struct{}{}
	}

	for _, id := range message.AttachmentIDs {
		att, err := b.attachmentProvider.QueryByID(ctx, id)
		if err != nil {
			return fmt.Errorf("start: query attachment: %w", err)
		}
		if att.Type == attachmentbus.Html {
			return fmt.Errorf("start: %w", messagebus.ErrMaxHTMLAttachment)
		}
		for _, v := range att.RequiredVars {
			vars[v] = struct{}{}
		}
	}

	siteEnabled := campaign.LandingID != nil || campaign.EducationID != nil || !campaign.Domain.IsEmpty()
	if siteEnabled {
		if campaign.LandingID == nil {
			return fmt.Errorf("start: %w", ErrLandingRequired)
		}
		if campaign.EducationID == nil {
			return fmt.Errorf("start: %w", ErrEducationRequired)
		}
		if campaign.Domain.IsEmpty() {
			return fmt.Errorf("start: %w", ErrDomainRequired)
		}

		landing, err := b.landingProvider.QueryByID(ctx, *campaign.LandingID)
		if err != nil {
			return fmt.Errorf("start: query landing: %w", err)
		}
		if !landing.HtmlBodyID.Valid {
			return fmt.Errorf("start: %w", ErrLandingHTMLRequired)
		}
		landingBody, err := b.attachmentProvider.QueryByID(ctx, landing.HtmlBodyID.UUID)
		if err != nil {
			return fmt.Errorf("start: query landing body: %w", err)
		}
		for _, v := range landingBody.RequiredVars {
			vars[v] = struct{}{}
		}

		education, err := b.landingProvider.QueryByID(ctx, *campaign.EducationID)
		if err != nil {
			return fmt.Errorf("start: query education: %w", err)
		}
		if !education.HtmlBodyID.Valid {
			return fmt.Errorf("start: %w", ErrEducationHTMLRequired)
		}
		educationBody, err := b.attachmentProvider.QueryByID(ctx, education.HtmlBodyID.UUID)
		if err != nil {
			return fmt.Errorf("start: query education body: %w", err)
		}
		for _, v := range educationBody.RequiredVars {
			vars[v] = struct{}{}
		}
	}

	if !siteEnabled {
		if _, ok := vars["Target.Link"]; ok {
			return fmt.Errorf("start: %w", ErrLandingRequired)
		}
	}

	for _, target := range targets {
		employee, err := b.employeeProvider.QueryByID(ctx, target.EmployeeID)
		if err != nil {
			return fmt.Errorf("start: query employee: %w", err)
		}
		if employee.Phone.String() == "" {
			return fmt.Errorf("start: employeeID[%s]: %w", target.EmployeeID, ErrTargetPhoneRequired)
		}
	}

	var reqVars []string
	for k := range vars {
		reqVars = append(reqVars, k)
	}
	sort.Strings(reqVars)

	res, err := b.varsValidator.Validate(ctx, campaign, targets, reqVars)
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	if len(res) > 0 {
		return fmt.Errorf("start: %w", &ErrTargetsMissingVars{Targets: res})
	}

	return nil
}

func (b *CampaignBusiness) Pause(ctx context.Context, campaign Campaign) (Campaign, error) {
	return b.changeStatus(ctx, campaign, Paused, "pause")
}

func (b *CampaignBusiness) Cancel(ctx context.Context, campaign Campaign) (Campaign, error) {
	return b.changeStatus(ctx, campaign, Canceled, "cancel")
}

func (b *CampaignBusiness) Count(ctx context.Context, filter CampaignQueryFilter) (int, error) {
	count, err := b.campaignStorer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

func (b *CampaignBusiness) CompleteExpired(ctx context.Context) []error {
	var errs []error
	campaigns, err := b.campaignStorer.QueryExpired(ctx)

	if err != nil {
		errs = append(errs, fmt.Errorf("completeexpired: %w", err))
		return errs
	}

	if len(campaigns) == 0 {
		return nil
	}

	var readyCmp []Campaign
	for _, campaign := range campaigns {
		targetCount, err := b.targetStorer.Count(ctx, TargetQueryFilter{CampaignID: &campaign.ID, Status: &Pending})
		if err != nil {
			errs = append(errs, fmt.Errorf("completeexpired: campaignID[%s]: %w", campaign.ID, err))
			continue
		}
		if targetCount != 0 {
			continue
		}
		readyCmp = append(readyCmp, campaign)
	}

	for _, campaign := range readyCmp {
		if _, err := b.changeStatus(ctx, campaign, Completed, "completeexpired"); err != nil {
			errs = append(errs, fmt.Errorf("completeexpired: campaignID[%s]: %w", campaign.ID, err))
		}
	}

	return errs
}

func (b *CampaignBusiness) changeStatus(ctx context.Context, campaign Campaign, status CampaignStatus, op string) (Campaign, error) {
	if !isValidCampaignTransition(campaign.Status, status) {
		return Campaign{}, fmt.Errorf("%s: %w: cannot move from %s to %s", op, ErrInvalidStatusTransition, campaign.Status, status)
	}

	campaign.Status = status

	if err := b.campaignStorer.Update(ctx, campaign); err != nil {
		return Campaign{}, fmt.Errorf("%s: %w", op, err)
	}

	return campaign, nil
}

func isValidCampaignTransition(current, next CampaignStatus) bool {
	switch current {
	case Draft:
		return next == Active || next == Canceled
	case Active:
		return next == Paused || next == Canceled || next == Completed
	case Paused:
		return next == Active || next == Canceled
	default:
		return false
	}
}
