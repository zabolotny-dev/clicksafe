package mid

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

func LoadDepartment(departmentBus *departmentbus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				dep, err := departmentBus.QueryByID(c.Request().Context(), productID)
				if err != nil {
					switch {
					case errors.Is(err, departmentbus.ErrNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "getbyid: departmentID[%s]: %s", productID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setDepartment(c.Request().Context(), dep),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}

func LoadEmployee(employeeBus *employeebus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				var err error
				employeeID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				employee, err := employeeBus.QueryByID(c.Request().Context(), employeeID)
				if err != nil {
					switch {
					case errors.Is(err, employeebus.ErrNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "querybyid: employeeID[%s]: %s", employeeID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setEmployee(c.Request().Context(), employee),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}

func LoadMessage(messageBus *messagebus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				messageID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				message, err := messageBus.QueryByID(c.Request().Context(), messageID)
				if err != nil {
					switch {
					case errors.Is(err, messagebus.ErrNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "querybyid: messageID[%s]: %s", messageID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setMessage(c.Request().Context(), message),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}

func LoadLanding(landingBus *landingbus.Business) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				landingID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				landing, err := landingBus.QueryByID(c.Request().Context(), landingID)
				if err != nil {
					switch {
					case errors.Is(err, landingbus.ErrNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "querybyid: landingID[%s]: %s", landingID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setLanding(c.Request().Context(), landing),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}

func LoadCampaign(campaignBus *campaignbus.CampaignBusiness) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				campaignID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				campaign, err := campaignBus.QueryByID(c.Request().Context(), campaignID)
				if err != nil {
					switch {
					case errors.Is(err, campaignbus.ErrCampaignNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "querybyid: campaignID[%s]: %s", campaignID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setCampaign(c.Request().Context(), campaign),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}

func LoadTarget(targetBus *campaignbus.TargetBusiness) echo.MiddlewareFunc {
	m := func(next echo.HandlerFunc) echo.HandlerFunc {
		h := func(c *echo.Context) error {
			id := c.Param("id")

			if id != "" {
				targetID, err := uuid.Parse(id)
				if err != nil {
					return errs.New(errs.InvalidArgument, errs.ErrInvalidID)
				}

				target, err := targetBus.QueryByID(c.Request().Context(), targetID)
				if err != nil {
					switch {
					case errors.Is(err, campaignbus.ErrTargetNotFound):
						return errs.New(errs.NotFound, err)
					default:
						return errs.Errorf(errs.InternalOnlyLog, "querybyid: targetID[%s]: %s", targetID, err)
					}
				}

				c.SetRequest(c.Request().WithContext(
					setTarget(c.Request().Context(), target),
				))
			}

			return next(c)
		}

		return h
	}

	return m
}
