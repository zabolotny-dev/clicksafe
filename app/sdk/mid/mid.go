package mid

import (
	"context"
	"errors"

	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus"
)

type ctxKey int

const (
	departmentKey ctxKey = iota + 1
	employeeKey
	messageKey
	campaignKey
	targetKey
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

func setTarget(ctx context.Context, t targetbus.Target) context.Context {
	return context.WithValue(ctx, targetKey, t)
}

func GetTarget(ctx context.Context) (targetbus.Target, error) {
	t, ok := ctx.Value(targetKey).(targetbus.Target)
	if !ok {
		return targetbus.Target{}, errors.New("target not found in context")
	}
	return t, nil
}
