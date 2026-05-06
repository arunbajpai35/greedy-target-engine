package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed sql/init.sql sql/seed.sql
var files embed.FS

func Apply(ctx context.Context, db *sql.DB, withSeed bool) error {
	if err := exec(ctx, db, "sql/init.sql"); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	if withSeed {
		if err := exec(ctx, db, "sql/seed.sql"); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
	}
	return nil
}

func exec(ctx context.Context, db *sql.DB, name string) error {
	b, err := files.ReadFile(name)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(b))
	return err
}
