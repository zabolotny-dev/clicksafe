package seed

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ScenarioDemo    = "demo"
	ScenarioMinimal = "minimal"

	dataDir = "data"
)

type Config struct {
	Scenario          string
	AttachmentRootDir string
}

//go:embed data/*.sql data/*.html data/*.txt
var seedFS embed.FS

func Run(ctx context.Context, db *pgxpool.Pool, cfg Config) error {
	cfg = DefaultConfig(cfg)

	if err := execSQL(ctx, db, "minimal.sql"); err != nil {
		return err
	}

	switch cfg.Scenario {
	case ScenarioMinimal:
		return nil

	case ScenarioDemo:
		if err := copyAssets(cfg.AttachmentRootDir); err != nil {
			return err
		}
		if err := execSQL(ctx, db, "seed.sql"); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown seed scenario %q", cfg.Scenario)
	}
}

func DefaultConfig(cfg Config) Config {
	cfg.Scenario = strings.ToLower(strings.TrimSpace(cfg.Scenario))
	if cfg.Scenario == "" || cfg.Scenario == "fixture" || cfg.Scenario == "fixtures" {
		cfg.Scenario = ScenarioDemo
	}
	if cfg.AttachmentRootDir == "" {
		cfg.AttachmentRootDir = "./private/attachment"
	}
	return cfg
}

func execSQL(ctx context.Context, db *pgxpool.Pool, name string) (err error) {
	content, err := fs.ReadFile(seedFS, path.Join(dataDir, name))
	if err != nil {
		return fmt.Errorf("read seed sql %q: %w", name, err)
	}

	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire db connection: %w", err)
	}
	defer conn.Release()

	defer func() {
		if err != nil {
			_, _ = conn.Conn().PgConn().Exec(ctx, "ROLLBACK").ReadAll()
		}
	}()

	sql := "BEGIN;\n" + string(content) + "\nCOMMIT;"
	if _, err = conn.Conn().PgConn().Exec(ctx, sql).ReadAll(); err != nil {
		return fmt.Errorf("exec seed sql %q: %w", name, err)
	}

	return nil
}

func copyAssets(attachmentRootDir string) error {
	if err := os.MkdirAll(attachmentRootDir, os.ModePerm); err != nil {
		return fmt.Errorf("create attachment dir: %w", err)
	}

	for srcName, dstName := range assets {
		dst := filepath.Join(attachmentRootDir, dstName)

		content, err := fs.ReadFile(seedFS, path.Join(dataDir, srcName))
		if err != nil {
			return fmt.Errorf("read seed asset %q: %w", srcName, err)
		}

		if err := os.WriteFile(dst, content, 0644); err != nil {
			return fmt.Errorf("write attachment asset %q: %w", dst, err)
		}
	}

	return nil
}

var assets = map[string]string{
	"landing.html":                  "e5270921-2a1f-4bb0-8a19-482470eb0004.html",
	"education.html":                "e5270921-2a1f-4bb0-8a19-482470eb0005.html",
	"msg_with_link.txt":             "e5270921-2a1f-4bb0-8a19-482470eb0008.txt",
	"msg_without_link.txt":          "e5270921-2a1f-4bb0-8a19-482470eb0009.txt",
	"max_education.txt":             "e5270921-2a1f-4bb0-8a19-482470eb0012.txt",
	"max_dit_crypto.txt":            "e5270921-2a1f-4bb0-8a19-482470eb0013.txt",
	"max_password.txt":              "e5270921-2a1f-4bb0-8a19-482470eb0014.txt",
	"email_dit_crypto.html":         "e5270921-2a1f-4bb0-8a19-482470eb0015.html",
	"landing_dit_crypto.html":       "e5270921-2a1f-4bb0-8a19-482470eb0016.html",
	"max_hr_vacation.txt":           "e5270921-2a1f-4bb0-8a19-482470eb0022.txt",
	"max_education_hr_vacation.txt": "e5270921-2a1f-4bb0-8a19-482470eb0023.txt",
	"email_hr_vacation.html":        "e5270921-2a1f-4bb0-8a19-482470eb0024.html",
	"landing_hr_vacation.html":      "e5270921-2a1f-4bb0-8a19-482470eb0025.html",
	"education_hr_vacation.html":    "e5270921-2a1f-4bb0-8a19-482470eb0026.html",
}
