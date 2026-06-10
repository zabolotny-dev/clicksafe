package maxaccountbus_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/maxaccountbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
)

func Test_MaxAccount(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_MaxAccount")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, query(db.BusDomain, sd), "query")
	unittest.Run(t, count(db.BusDomain, sd), "count")
	unittest.Run(t, queryByID(db.BusDomain, sd), "querybyid")
	unittest.Run(t, update(db.BusDomain, sd), "update")
	unittest.Run(t, deleteAcc(db.BusDomain, sd), "delete")
	unittest.Run(t, startStop(db.BusDomain, sd), "startstop")
	unittest.Run(t, adapterOps(db.BusDomain), "adapter")
	unittest.Run(t, statusOps(), "status")
}

// =============================================================================

type seedData struct {
	Accounts []maxaccountbus.Account
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	accs, err := maxaccountbus.TestSeedAccounts(ctx, 3, busDomain.MaxAccount)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding accounts: %w", err)
	}

	return seedData{Accounts: accs}, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	label := sd.Accounts[0].Label
	phone := sd.Accounts[0].Phone

	return []unittest.Table{
		{
			Name:    "all-returns-seeded-accounts",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				accs, err := busDomain.MaxAccount.Query(ctx, maxaccountbus.QueryFilter{}, page.MustParse("1", "10"))
				return err == nil && len(accs) >= len(sd.Accounts)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "filter-by-label-returns-one",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				accs, err := busDomain.MaxAccount.Query(ctx, maxaccountbus.QueryFilter{Label: &label}, page.MustParse("1", "10"))
				return err == nil && len(accs) == 1 && accs[0].Label == label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "filter-by-phone-returns-one",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				accs, err := busDomain.MaxAccount.Query(ctx, maxaccountbus.QueryFilter{Phone: &phone}, page.MustParse("1", "10"))
				return err == nil && len(accs) == 1 && accs[0].Phone == phone
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "pagination-page-two-empty",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				accs, err := busDomain.MaxAccount.Query(ctx, maxaccountbus.QueryFilter{}, page.MustParse("2", "100"))
				return err == nil && len(accs) == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func count(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	label := sd.Accounts[1].Label
	absent := "nonexistent-xyz"

	return []unittest.Table{
		{
			Name:    "total-count-gte-seeded",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				n, err := busDomain.MaxAccount.Count(ctx, maxaccountbus.QueryFilter{})
				return err == nil && n >= len(sd.Accounts)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "filter-by-label-count-one",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				n, err := busDomain.MaxAccount.Count(ctx, maxaccountbus.QueryFilter{Label: &label})
				return err == nil && n == 1
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "filter-absent-label-count-zero",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				n, err := busDomain.MaxAccount.Count(ctx, maxaccountbus.QueryFilter{Label: &absent})
				return err == nil && n == 0
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func queryByID(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	acc := sd.Accounts[0]

	return []unittest.Table{
		{
			Name:    "existing-id-returns-account",
			ExpResp: acc.ID,
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.MaxAccount.QueryByID(ctx, acc.ID)
				if err != nil {
					return err
				}
				return result.ID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "unknown-id-returns-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.QueryByID(ctx, uuid.New())
				return errors.Is(err, maxaccountbus.ErrAccountNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func update(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	acc := sd.Accounts[1]

	return []unittest.Table{
		{
			Name:    "update-label-persisted",
			ExpResp: "Updated Label",
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.MaxAccount.Update(ctx, acc, maxaccountbus.UpdateAccount{Label: "Updated Label"})
				if err != nil {
					return err
				}
				return result.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-after-update-reflects-new-label",
			ExpResp: "Updated Label",
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.MaxAccount.QueryByID(ctx, acc.ID)
				if err != nil {
					return err
				}
				return result.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-nonexistent-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				ghost := maxaccountbus.Account{ID: uuid.New()}
				_, err := busDomain.MaxAccount.Update(ctx, ghost, maxaccountbus.UpdateAccount{Label: "Ghost"})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func deleteAcc(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	acc := sd.Accounts[2]

	return []unittest.Table{
		{
			Name:    "delete-existing-no-error",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.MaxAccount.Delete(ctx, acc)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-after-delete-returns-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.QueryByID(ctx, acc.ID)
				return errors.Is(err, maxaccountbus.ErrAccountNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "delete-nonexistent-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				ghost := maxaccountbus.Account{ID: uuid.New()}
				err := busDomain.MaxAccount.Delete(ctx, ghost)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func startStop(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	acc := sd.Accounts[0]

	return []unittest.Table{
		{
			// no-op adapter returns Account{} with empty status → Upsert fails on ParseStatus
			// covers adapter call + storer.Upsert + error return paths in Start
			Name:    "start-no-op-adapter-upsert-fails",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.Start(ctx, acc)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			// same for Stop
			Name:    "stop-no-op-adapter-upsert-fails",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.Stop(ctx, acc)
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func adapterOps(busDomain dbtest.BusDomain) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "begin-login-no-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.BeginLogin(ctx, maxaccountbus.BeginLogin{
					Phone: "+79001234567",
					Label: "Test Max Account",
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "begin-login-label-set-from-request",
			ExpResp: "My Label",
			ExcFunc: func(ctx context.Context) any {
				attempt, _ := busDomain.MaxAccount.BeginLogin(ctx, maxaccountbus.BeginLogin{
					Phone: "+79009999999",
					Label: "My Label",
				})
				return attempt.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "confirm-login-no-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.ConfirmLogin(ctx, maxaccountbus.ConfirmLogin{
					AttemptID: "test-attempt",
					Code:      "12345",
					Password:  "password",
					Label:     "Test Account",
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "confirm-login-label-propagated-to-attempt",
			ExpResp: "Propagated label",
			ExcFunc: func(ctx context.Context) any {
				// no-op: Account==nil, firstNonBlank(localLabel, attempt.Label) used
				lr, _ := busDomain.MaxAccount.ConfirmLogin(ctx, maxaccountbus.ConfirmLogin{
					Label: "Propagated label",
				})
				return lr.Attempt.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "confirm-login-empty-labels-uses-default",
			ExpResp: "Max account",
			ExcFunc: func(ctx context.Context) any {
				// both localLabel="" and attempt.Label="" and attempt.Phone="" → "Max account"
				lr, _ := busDomain.MaxAccount.ConfirmLogin(ctx, maxaccountbus.ConfirmLogin{})
				return lr.Attempt.Label
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "confirm-password-no-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.MaxAccount.ConfirmPassword(ctx, maxaccountbus.ConfirmPassword{
					AttemptID: "test-attempt",
					Password:  "password",
					Label:     "Test Account",
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func statusOps() []unittest.Table {
	return []unittest.Table{
		{
			Name:    "active-string",
			ExpResp: "ACTIVE",
			ExcFunc: func(ctx context.Context) any {
				return maxaccountbus.StatusActive.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "connected-string",
			ExpResp: "CONNECTED",
			ExcFunc: func(ctx context.Context) any {
				return maxaccountbus.StatusConnected.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "disconnected-string",
			ExpResp: "DISCONNECTED",
			ExcFunc: func(ctx context.Context) any {
				return maxaccountbus.StatusDisconnected.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "error-string",
			ExpResp: "ERROR",
			ExcFunc: func(ctx context.Context) any {
				return maxaccountbus.StatusError.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "pending-login-string",
			ExpResp: "PENDING_LOGIN",
			ExcFunc: func(ctx context.Context) any {
				return maxaccountbus.StatusPendingLogin.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "parse-valid-status",
			ExpResp: "CONNECTED",
			ExcFunc: func(ctx context.Context) any {
				s, _ := maxaccountbus.ParseStatus("CONNECTED")
				return s.String()
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "parse-invalid-returns-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := maxaccountbus.ParseStatus("UNKNOWN_STATUS")
				return errors.Is(err, maxaccountbus.ErrInvalidStatus)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}
