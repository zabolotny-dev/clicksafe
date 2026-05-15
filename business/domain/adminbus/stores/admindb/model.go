package admindb

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus/stores/admindb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

func toDBAdmin(adm adminbus.Admin) (sqlc.Admin, error) {
	return sqlc.Admin{
		ID:           adm.ID,
		FirstName:    adm.FirstName.String(),
		LastName:     adm.LastName.String(),
		Login:        adm.Login.String(),
		PasswordHash: adm.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: adm.CreatedAt, Valid: true},
	}, nil
}

func toBusAdmin(adm sqlc.Admin) (adminbus.Admin, error) {
	firstName, err := name.Parse(adm.FirstName)
	if err != nil {
		return adminbus.Admin{}, err
	}

	lastName, err := name.Parse(adm.LastName)
	if err != nil {
		return adminbus.Admin{}, err
	}

	login, err := login.Parse(adm.Login)
	if err != nil {
		return adminbus.Admin{}, err
	}

	return adminbus.Admin{
		ID:           adm.ID,
		FirstName:    firstName,
		LastName:     lastName,
		Login:        login,
		PasswordHash: adm.PasswordHash,
		CreatedAt:    adm.CreatedAt.Time,
	}, nil
}

func toBusAdmins(admins []sqlc.Admin) ([]adminbus.Admin, error) {
	busAdmins := make([]adminbus.Admin, len(admins))

	for i, adm := range admins {
		var err error
		busAdmins[i], err = toBusAdmin(adm)
		if err != nil {
			return nil, err
		}
	}

	return busAdmins, nil
}
