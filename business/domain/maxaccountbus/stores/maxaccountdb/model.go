package maxaccountdb

import (
	"fmt"

	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus/stores/maxaccountdb/sqlc"
)

func toBusAccount(dbAcc sqlc.MaxAccount) (maxaccountbus.Account, error) {
	status, err := maxaccountbus.ParseStatus(dbAcc.Status)
	if err != nil {
		return maxaccountbus.Account{}, fmt.Errorf("parse status: %w", err)
	}

	return maxaccountbus.Account{
		ID:        dbAcc.ID,
		AdapterID: dbAcc.AdapterID,
		Phone:     dbAcc.PhoneNumber,
		Label:     dbAcc.Label,
		Status:    status,
		MaxUserID: dbAcc.MaxUserID.String,
		LastError: dbAcc.LastError.String,
		CreatedAt: dbAcc.CreatedAt.Time,
		UpdatedAt: dbAcc.UpdatedAt.Time,
	}, nil
}

func toBusAccounts(dbAccs []sqlc.MaxAccount) ([]maxaccountbus.Account, error) {
	accounts := make([]maxaccountbus.Account, len(dbAccs))
	for i, dbAcc := range dbAccs {
		acc, err := toBusAccount(dbAcc)
		if err != nil {
			return nil, err
		}
		accounts[i] = acc
	}
	return accounts, nil
}
