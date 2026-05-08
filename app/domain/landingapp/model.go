package landingapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/resolverbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Landing struct {
	ID           uuid.UUID `json:"id"`
	Label        string    `json:"label"`
	HasContent   bool      `json:"has_content"`
	RequiredVars []string  `json:"required_vars"`
}

type NewLanding struct {
	Label string `json:"label"`
}

type UpdateLanding struct {
	Label *string `json:"label"`
}

type RenderLanding struct {
	TargetID string `json:"target_id"`
}

type RenderedLanding struct {
	Content string `json:"content"`
}

func toBusNewLanding(req NewLanding) (landingbus.NewLanding, error) {
	var fieldErrors errs.FieldErrors

	lbl, err := label.Parse(req.Label)
	if err != nil {
		fieldErrors.Add("label", err)
	}

	if len(fieldErrors) > 0 {
		return landingbus.NewLanding{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return landingbus.NewLanding{
		Label: lbl,
	}, nil
}

func toBusUpdateLanding(req UpdateLanding) (landingbus.UpdateLanding, error) {
	var fieldErrors errs.FieldErrors

	var lbl *label.Label
	if req.Label != nil {
		parsed, err := label.Parse(*req.Label)
		if err != nil {
			fieldErrors.Add("label", err)
		}
		lbl = &parsed
	}

	if len(fieldErrors) > 0 {
		return landingbus.UpdateLanding{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return landingbus.UpdateLanding{
		Label: lbl,
	}, nil
}

func toBusRenderScope(req RenderLanding) (resolverbus.Scope, error) {
	var fieldErrors errs.FieldErrors
	var scope resolverbus.Scope

	if req.TargetID != "" {
		id, err := uuid.Parse(req.TargetID)
		if err != nil {
			fieldErrors.Add("target_id", err)
		}
		scope.TargetID = id
	}

	if len(fieldErrors) > 0 {
		return resolverbus.Scope{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	return scope, nil
}

func toAppLanding(landing landingbus.Landing) Landing {
	return Landing{
		ID:           landing.ID,
		Label:        landing.Label.String(),
		HasContent:   landing.ContentPath.Valid(),
		RequiredVars: landing.RequiredVars,
	}
}

func toAppLandings(landings []landingbus.Landing) []Landing {
	items := make([]Landing, len(landings))
	for i, landing := range landings {
		items[i] = toAppLanding(landing)
	}

	return items
}
