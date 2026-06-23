package loadtest_test

import (
	"context"
	"fmt"
	"net/mail"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/apitest"
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/domain/sessionbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/seed"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/domain"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/login"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/password"
)

// UUIDs из seed.sql — детерминированы и не меняются между запусками.
const (
	seedDepartmentID = "e5270921-2a1f-4bb0-8a19-482470eb0001"
	seedLandingID    = "e5270921-2a1f-4bb0-8a19-482470eb0006" // Microsoft Login Page
	seedEducationID  = "e5270921-2a1f-4bb0-8a19-482470eb0007" // Security Warning Education
	seedMessageID    = "e5270921-2a1f-4bb0-8a19-482470eb0020" // EMAIL: Установка обновления безопасности ИСиР
)

const targetCount = 200

type loadSeedData struct {
	apitest.SeedData
	PhishingTokens []string
	EmployeeIDs    []uuid.UUID
}

func insertSeedData(db *dbtest.Database) (loadSeedData, error) {
	ctx := context.Background()

	if err := seed.Run(ctx, db.DB, seed.Config{
		Scenario:          seed.ScenarioDemo,
		AttachmentRootDir: db.TmpDir,
	}); err != nil {
		return loadSeedData{}, fmt.Errorf("seed: %w", err)
	}

	adm, err := db.BusDomain.Admin.Save(ctx, adminbus.NewAdmin{
		FirstName: name.MustParse("Нагрузочный"),
		LastName:  name.MustParse("Тест"),
		Login:     login.MustParse("loadtest.admin"),
		Password:  password.MustParse("LoadTest!2024"),
	})
	if err != nil {
		return loadSeedData{}, fmt.Errorf("creating admin: %w", err)
	}

	sess, err := db.BusDomain.Session.Create(ctx, sessionbus.NewSession{
		AdminID:   adm.ID,
		IPAddress: netip.MustParseAddr("127.0.0.1"),
		UserAgent: "loadtest",
	})
	if err != nil {
		return loadSeedData{}, fmt.Errorf("creating session: %w", err)
	}

	deptID     := uuid.MustParse(seedDepartmentID)
	landingID  := uuid.MustParse(seedLandingID)
	educID     := uuid.MustParse(seedEducationID)
	messageID  := uuid.MustParse(seedMessageID)
	dom        := domain.MustParse("http://localhost")
	now        := time.Now().UTC()

	dateRange, err := date.ParseNull(now, now.Add(30*24*time.Hour))
	if err != nil {
		return loadSeedData{}, fmt.Errorf("date range: %w", err)
	}

	firstNames := []string{
		"Александр", "Алексей", "Андрей", "Антон", "Артём",
		"Борис", "Василий", "Виктор", "Владимир", "Дмитрий",
		"Евгений", "Иван", "Кирилл", "Максим", "Михаил",
		"Никита", "Павел", "Роман", "Сергей", "Тимур",
		"Анна", "Елена", "Екатерина", "Мария", "Наталья",
	}

	lastNames := []string{
		"Иванов", "Петров", "Сидоров", "Козлов", "Смирнов",
		"Попов", "Новиков", "Волков", "Соколов", "Морозов",
		"Лебедев", "Семёнов", "Фёдоров", "Давыдов", "Зайцев",
		"Мельников", "Захаров", "Орлов", "Беляев", "Ермаков",
	}

	// создаём сотрудников для нагрузочного теста
	employees := make([]employeebus.Employee, targetCount)
	for i := range targetCount {
		emp, err := db.BusDomain.Employee.Save(ctx, employeebus.NewEmployee{
			DepartmentID: &deptID,
			FirstName:    name.MustParse(firstNames[i%len(firstNames)]),
			LastName:     name.MustParse(lastNames[i%len(lastNames)]),
			Email:        mail.Address{Address: fmt.Sprintf("loadtest%d@company.com", i+1)},
		})
		if err != nil {
			return loadSeedData{}, fmt.Errorf("creating employee %d: %w", i+1, err)
		}
		employees[i] = emp
	}

	cmp, err := db.BusDomain.Campaign.Save(ctx, campaignbus.NewCampaign{
		Label: label.MustParse("Нагрузочный тест"),
	})
	if err != nil {
		return loadSeedData{}, fmt.Errorf("creating campaign: %w", err)
	}

	cmp, err = db.BusDomain.Campaign.Update(ctx, cmp, campaignbus.UpdateCampaign{
		MessageID:   &messageID,
		LandingID:   &landingID,
		EducationID: &educID,
		Domain:      &dom,
		DateRange:   &dateRange,
	})
	if err != nil {
		return loadSeedData{}, fmt.Errorf("updating campaign: %w", err)
	}

	// создаём цель для каждого сотрудника
	targets := make([]campaignbus.Target, targetCount)
	for i, emp := range employees {
		t, err := db.BusDomain.Target.Save(ctx, campaignbus.NewTarget{
			EmployeeID: emp.ID,
			CampaignID: cmp.ID,
		})
		if err != nil {
			return loadSeedData{}, fmt.Errorf("creating target %d: %w", i+1, err)
		}
		targets[i] = t
	}

	if err := db.BusDomain.Target.AutoDistribute(ctx, cmp.ID); err != nil {
		return loadSeedData{}, fmt.Errorf("distributing targets: %w", err)
	}

	if _, err := db.BusDomain.Campaign.Start(ctx, cmp); err != nil {
		return loadSeedData{}, fmt.Errorf("starting campaign: %w", err)
	}

	tokens := make([]string, targetCount)
	for i, t := range targets {
		tokens[i] = t.Token
	}

	empIDs := make([]uuid.UUID, targetCount)
	for i, e := range employees {
		empIDs[i] = e.ID
	}

	return loadSeedData{
		SeedData: apitest.SeedData{
			Admins: []apitest.SeedAdmin{{
				Admin:     adm,
				RawToken:  sess.Token,
				CSRFToken: sess.CSRFToken,
			}},
			Campaigns: []campaignbus.Campaign{cmp},
		},
		PhishingTokens: tokens,
		EmployeeIDs:    empIDs,
	}, nil
}
