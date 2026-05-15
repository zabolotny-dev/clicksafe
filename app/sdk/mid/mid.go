package mid

import (
	"context"
	"errors"

	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
)

type ctxKey int

const (
	departmentKey ctxKey = iota + 1
	employeeKey
	messageKey
	campaignKey
	targetKey
	landingKey
	attachmentKey
	sessionKey
)

func setDepartment(ctx context.Context, d departmentbus.Department) context.Context {
	return context.WithValue(ctx, departmentKey, d)
}

func GetDepartment(ctx context.Context) (departmentbus.Department, error) {
	d, ok := ctx.Value(departmentKey).(departmentbus.Department)
	if !ok {
		return departmentbus.Department{}, errors.New("department not found in context")
	}
	return d, nil
}

func setEmployee(ctx context.Context, e employeebus.Employee) context.Context {
	return context.WithValue(ctx, employeeKey, e)
}

func GetEmployee(ctx context.Context) (employeebus.Employee, error) {
	e, ok := ctx.Value(employeeKey).(employeebus.Employee)
	if !ok {
		return employeebus.Employee{}, errors.New("employee not found in context")
	}
	return e, nil
}

func setMessage(ctx context.Context, m messagebus.Message) context.Context {
	return context.WithValue(ctx, messageKey, m)
}

func GetMessage(ctx context.Context) (messagebus.Message, error) {
	m, ok := ctx.Value(messageKey).(messagebus.Message)
	if !ok {
		return messagebus.Message{}, errors.New("message not found in context")
	}
	return m, nil
}

func setLanding(ctx context.Context, landing landingbus.Landing) context.Context {
	return context.WithValue(ctx, landingKey, landing)
}

func GetLanding(ctx context.Context) (landingbus.Landing, error) {
	landing, ok := ctx.Value(landingKey).(landingbus.Landing)
	if !ok {
		return landingbus.Landing{}, errors.New("landing not found in context")
	}
	return landing, nil
}

func setCampaign(ctx context.Context, c campaignbus.Campaign) context.Context {
	return context.WithValue(ctx, campaignKey, c)
}

func GetCampaign(ctx context.Context) (campaignbus.Campaign, error) {
	c, ok := ctx.Value(campaignKey).(campaignbus.Campaign)
	if !ok {
		return campaignbus.Campaign{}, errors.New("campaign not found in context")
	}
	return c, nil
}

func setTarget(ctx context.Context, t campaignbus.Target) context.Context {
	return context.WithValue(ctx, targetKey, t)
}

func GetTarget(ctx context.Context) (campaignbus.Target, error) {
	t, ok := ctx.Value(targetKey).(campaignbus.Target)
	if !ok {
		return campaignbus.Target{}, errors.New("target not found in context")
	}
	return t, nil
}

func setAttachment(ctx context.Context, atch attachmentbus.Attachment) context.Context {
	return context.WithValue(ctx, attachmentKey, atch)
}

func GetAttachment(ctx context.Context) (attachmentbus.Attachment, error) {
	atch, ok := ctx.Value(attachmentKey).(attachmentbus.Attachment)
	if !ok {
		return attachmentbus.Attachment{}, errors.New("attachment not found in context")
	}
	return atch, nil
}

func setSession(ctx context.Context, s sessionbus.Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func GetSession(ctx context.Context) (sessionbus.Session, error) {
	s, ok := ctx.Value(sessionKey).(sessionbus.Session)
	if !ok {
		return sessionbus.Session{}, errors.New("session not found in context")
	}
	return s, nil
}
