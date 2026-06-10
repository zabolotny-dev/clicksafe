package adminbus_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

func Test_Admin(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Admin")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, query(db.BusDomain, sd), "query")
	unittest.Run(t, create(db.BusDomain), "create")
	unittest.Run(t, update(db.BusDomain, sd), "update")
	unittest.Run(t, authenticate(db.BusDomain, sd), "authenticate")
	unittest.Run(t, querylist(db.BusDomain, sd), "querylist")
	unittest.Run(t, counttest(db.BusDomain), "count")
}

// =============================================================================

type seedData struct {
	Admins    []adminbus.Admin
	Passwords []string
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	admins, passwords, err := adminbus.TestSeedAdmins(ctx, 2, busDomain.Admin)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding admins: %w", err)
	}

	return seedData{Admins: admins, Passwords: passwords}, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "byid",
			ExpResp: sd.Admins[0].ID.String(),
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Admin.QueryByID(ctx, sd.Admins[0].ID)
				if err != nil {
					return err
				}
				return resp.ID.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "bylogin",
			ExpResp: sd.Admins[0].Login.String(),
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Admin.QueryByLogin(ctx, sd.Admins[0].Login)
				if err != nil {
					return err
				}
				return resp.Login.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain) []unittest.Table {
	return []unittest.Table{
		{
			Name: "basic",
			ExpResp: adminbus.Admin{
				FirstName: name.MustParse("Create"),
				LastName:  name.MustParse("Admin"),
				Login:     login.MustParse("create.admin"),
			},
			ExcFunc: func(ctx context.Context) any {
				na := adminbus.NewAdmin{
					FirstName: name.MustParse("Create"),
					LastName:  name.MustParse("Admin"),
					Login:     login.MustParse("create.admin"),
					Password:  password.MustParse("Password123!@#sec"),
				}
				resp, err := busDomain.Admin.Save(ctx, na)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(adminbus.Admin)
				if !ok {
					return "error occurred"
				}
				expResp := exp.(adminbus.Admin)
				expResp.ID = gotResp.ID
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.CreatedAt = gotResp.CreatedAt
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	newFirst := name.MustParse("Updated")

	return []unittest.Table{
		{
			Name: "basic",
			ExpResp: adminbus.Admin{
				ID:        sd.Admins[0].ID,
				FirstName: newFirst,
				LastName:  sd.Admins[0].LastName,
				Login:     sd.Admins[0].Login,
				CreatedAt: sd.Admins[0].CreatedAt,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Admin.Update(ctx, sd.Admins[0].ID, adminbus.UpdateAdmin{
					FirstName: &newFirst,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(adminbus.Admin)
				if !ok {
					return "error occurred"
				}
				expResp := exp.(adminbus.Admin)
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.CreatedAt = gotResp.CreatedAt
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func authenticate(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "valid-credentials",
			ExpResp: sd.Admins[0].Login.String(),
			ExcFunc: func(ctx context.Context) any {
				adm, err := busDomain.Admin.Authenticate(ctx, sd.Admins[0].Login, password.MustParse(sd.Passwords[0]))
				if err != nil {
					return err
				}
				return adm.Login.String()
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "invalid-password",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Admin.Authenticate(ctx, sd.Admins[0].Login, password.MustParse("WrongPassword123!"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "not-found-login",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Admin.QueryByID(ctx, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func querylist(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "all",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				result, err := busDomain.Admin.Query(ctx, adminbus.AdminQueryFilter{})
				if err != nil {
					return err
				}
				return len(result) >= len(sd.Admins)
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Stub-based unit tests for error paths

type adminStorerStub struct {
	data     map[uuid.UUID]adminbus.Admin
	loginMap map[string]adminbus.Admin
	saveErr  error
	qErr     error
}

func newAdminStorerStub() *adminStorerStub {
	return &adminStorerStub{
		data:     make(map[uuid.UUID]adminbus.Admin),
		loginMap: make(map[string]adminbus.Admin),
	}
}

func (s *adminStorerStub) Save(_ context.Context, adm adminbus.Admin) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[adm.ID] = adm
	s.loginMap[adm.Login.String()] = adm
	return nil
}

func (s *adminStorerStub) Update(_ context.Context, adm adminbus.Admin) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[adm.ID] = adm
	s.loginMap[adm.Login.String()] = adm
	return nil
}

func (s *adminStorerStub) QueryByID(_ context.Context, id uuid.UUID) (adminbus.Admin, error) {
	if s.qErr != nil {
		return adminbus.Admin{}, s.qErr
	}
	adm, ok := s.data[id]
	if !ok {
		return adminbus.Admin{}, adminbus.ErrNotFound
	}
	return adm, nil
}

func (s *adminStorerStub) QueryByLogin(_ context.Context, lgn login.Login) (adminbus.Admin, error) {
	if s.qErr != nil {
		return adminbus.Admin{}, s.qErr
	}
	adm, ok := s.loginMap[lgn.String()]
	if !ok {
		return adminbus.Admin{}, adminbus.ErrNotFound
	}
	return adm, nil
}

func (s *adminStorerStub) Query(_ context.Context, _ adminbus.AdminQueryFilter) ([]adminbus.Admin, error) {
	if s.qErr != nil {
		return nil, s.qErr
	}
	var result []adminbus.Admin
	for _, a := range s.data {
		result = append(result, a)
	}
	return result, nil
}

func (s *adminStorerStub) Count(_ context.Context, _ adminbus.AdminQueryFilter) (int, error) {
	if s.qErr != nil {
		return 0, s.qErr
	}
	return len(s.data), nil
}

type hasherStub struct {
	generateErr error
}

func (h *hasherStub) Generate(p string) (string, error) {
	if h.generateErr != nil {
		return "", h.generateErr
	}
	return "hashed:" + p, nil
}

func (h *hasherStub) Compare(p, encodedHash string) (bool, error) {
	return encodedHash == "hashed:"+p, nil
}

func Test_Admin_UnitErrors(t *testing.T) {
	t.Parallel()
	unittest.Run(t, adminErrorPaths(), "error-paths")
}

func adminErrorPaths() []unittest.Table {
	dbErr := errors.New("db error")
	hashErr := errors.New("hash error")

	admID := uuid.New()
	lgn := login.MustParse("err.admin")

	origAdm := adminbus.Admin{
		ID:           admID,
		FirstName:    name.MustParse("Orig"),
		LastName:     name.MustParse("Admin"),
		Login:        lgn,
		PasswordHash: "hashed:oldpass",
	}

	// Save - storer error
	storeSaveErr := newAdminStorerStub()
	storeSaveErr.saveErr = dbErr
	busSaveErr := adminbus.NewBusiness(&hasherStub{}, storeSaveErr)

	// Save - hasher error
	storeOK := newAdminStorerStub()
	busHashErr := adminbus.NewBusiness(&hasherStub{generateErr: hashErr}, storeOK)

	// Update - query not found
	busUpdateNotFound := adminbus.NewBusiness(&hasherStub{}, newAdminStorerStub())

	// Update - hasher error on password change
	storeForUpdate := newAdminStorerStub()
	storeForUpdate.data[admID] = origAdm
	busUpdateHashErr := adminbus.NewBusiness(&hasherStub{generateErr: hashErr}, storeForUpdate)

	// Update - storer error
	storeUpdateErr := newAdminStorerStub()
	storeUpdateErr.data[admID] = origAdm
	storeUpdateErr.saveErr = dbErr
	busUpdateStoreErr := adminbus.NewBusiness(&hasherStub{}, storeUpdateErr)

	// Query error
	storeQErr := newAdminStorerStub()
	storeQErr.qErr = dbErr
	busQErr := adminbus.NewBusiness(&hasherStub{}, storeQErr)

	// Authenticate - login not found
	busAuthNotFound := adminbus.NewBusiness(&hasherStub{}, newAdminStorerStub())

	newPass := password.MustParse("NewPass123456!")
	newFirst := name.MustParse("NewFirst")

	return []unittest.Table{
		{
			Name:    "save-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busSaveErr.Save(ctx, adminbus.NewAdmin{
					Login:    login.MustParse("save.err"),
					Password: password.MustParse("Pass123456789!"),
				})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "save-hasher-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busHashErr.Save(ctx, adminbus.NewAdmin{
					Login:    login.MustParse("hash.err"),
					Password: password.MustParse("Pass123456789!"),
				})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-admin-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busUpdateNotFound.Update(ctx, uuid.New(), adminbus.UpdateAdmin{FirstName: &newFirst})
				return errors.Is(err, adminbus.ErrNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-hasher-error-on-password",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busUpdateHashErr.Update(ctx, admID, adminbus.UpdateAdmin{Password: &newPass})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busUpdateStoreErr.Update(ctx, admID, adminbus.UpdateAdmin{FirstName: &newFirst})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-list-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQErr.Query(ctx, adminbus.AdminQueryFilter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-count-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQErr.Count(ctx, adminbus.AdminQueryFilter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-bylogin-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busAuthNotFound.QueryByLogin(ctx, login.MustParse("ghost.user"))
				return errors.Is(err, adminbus.ErrNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "authenticate-login-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busAuthNotFound.Authenticate(ctx, login.MustParse("no.user.here"), password.MustParse("Pass12345678!"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func counttest(busDomain dbtest.BusDomain) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "total",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				n, err := busDomain.Admin.Count(ctx, adminbus.AdminQueryFilter{})
				if err != nil {
					return err
				}
				return n >= 0
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
