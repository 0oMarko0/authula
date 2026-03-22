package configmanager

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/0oMarko0/authula/migrations"
)

func configManagerMigrationsForProvider(provider string) []migrations.Migration {
	return migrations.ForProvider(provider, migrations.ProviderVariants{
		"sqlite":   func() []migrations.Migration { return []migrations.Migration{configManagerSQLiteInitial()} },
		"postgres": func() []migrations.Migration { return []migrations.Migration{configManagerPostgresInitial()} },
		"mysql":    func() []migrations.Migration { return []migrations.Migration{configManagerMySQLInitial()} },
	})
}

func configManagerSQLiteInitial() migrations.Migration {
	return migrations.Migration{
		Version: "20260128000000_config_manager_initial",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return migrations.ExecStatements(
				ctx,
				tx,
				`CREATE TABLE IF NOT EXISTS authula_auth_settings (
  config_version INTEGER PRIMARY KEY NOT NULL,
  key VARCHAR(255) NOT NULL UNIQUE,
  value TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_auth_settings_config_version ON authula_auth_settings(config_version);`,
			)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return migrations.ExecStatements(ctx, tx, `DROP TABLE IF EXISTS authula_auth_settings;`)
		},
	}
}

func configManagerPostgresInitial() migrations.Migration {
	return migrations.Migration{
		Version: "20260128000000_config_manager_initial",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return migrations.ExecStatements(
				ctx,
				tx,
				`CREATE OR REPLACE FUNCTION config_manager_update_updated_at_column_func()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE SEQUENCE IF NOT EXISTS config_version_seq START WITH 1 INCREMENT BY 1;`,
				`CREATE TABLE IF NOT EXISTS authula_auth_settings (
  config_version BIGINT PRIMARY KEY DEFAULT nextval('config_version_seq'),
  key VARCHAR(255) UNIQUE NOT NULL,
  value JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_auth_settings_config_version ON authula_auth_settings(config_version);`,
				`DROP TRIGGER IF EXISTS update_authula_auth_settings_updated_at_trigger ON authula_auth_settings;`,
				`CREATE TRIGGER update_authula_auth_settings_updated_at_trigger
  BEFORE UPDATE ON authula_auth_settings
  FOR EACH ROW
  EXECUTE FUNCTION config_manager_update_updated_at_column_func();`,
			)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return migrations.ExecStatements(
				ctx,
				tx,
				`DROP TRIGGER IF EXISTS update_authula_auth_settings_updated_at_trigger ON authula_auth_settings;`,
				`DROP TABLE IF EXISTS authula_auth_settings;`,
				`DROP SEQUENCE IF EXISTS config_version_seq;`,
			)
		},
	}
}

func configManagerMySQLInitial() migrations.Migration {
	return migrations.Migration{
		Version: "20260128000000_config_manager_initial",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return migrations.ExecStatements(
				ctx,
				tx,
				`CREATE TABLE IF NOT EXISTS authula_auth_settings (
  config_version BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  key VARCHAR(255) NOT NULL UNIQUE,
  value JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_authula_auth_settings_key (key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
			)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return migrations.ExecStatements(ctx, tx, `DROP TABLE IF EXISTS authula_auth_settings;`)
		},
	}
}
