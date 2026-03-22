package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

// CorePluginID identifies the core migrations set.
const CorePluginID = "core"

// CoreMigrationSet returns the programmatic migrations for the core schema.
func CoreMigrationSet(provider string) (MigrationSet, error) {
	switch provider {
	case "sqlite":
		return MigrationSet{
			PluginID: CorePluginID,
			Migrations: []Migration{
				coreSQLiteInitial(),
				coreSQLiteAddMetadata(),
			},
		}, nil
	case "postgres":
		return MigrationSet{
			PluginID: CorePluginID,
			Migrations: []Migration{
				corePostgresInitial(),
				corePostgresAddMetadata(),
			},
		}, nil
	case "mysql":
		return MigrationSet{
			PluginID: CorePluginID,
			Migrations: []Migration{
				coreMySQLInitial(),
				coreMySQLAddMetadata(),
			},
		}, nil
	default:
		return MigrationSet{}, fmt.Errorf("unsupported database provider: %s", provider)
	}
}

func coreSQLiteInitial() Migration {
	return Migration{
		Version: "20260126000000_core_initial",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(
				ctx,
				tx,
				`PRAGMA foreign_keys = ON;`,
				`CREATE TABLE IF NOT EXISTS authula_users (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  email_verified BOOLEAN DEFAULT 0,
  image TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,
				`CREATE TABLE IF NOT EXISTS authula_accounts (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  account_id VARCHAR(255) NOT NULL,
  provider_id VARCHAR(255) NOT NULL,
  access_token TEXT,
  refresh_token TEXT,
  id_token TEXT,
  access_token_expires_at TIMESTAMP,
  refresh_token_expires_at TIMESTAMP,
  scope TEXT,
  password TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE,
  UNIQUE(provider_id, account_id)
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_accounts_user_id ON authula_accounts(user_id);`,
				`CREATE TABLE IF NOT EXISTS authula_sessions (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  token VARCHAR(255) UNIQUE NOT NULL,
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_sessions_user_id ON authula_sessions(user_id);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_sessions_expires_at ON authula_sessions(expires_at);`,
				`CREATE TABLE IF NOT EXISTS authula_verifications (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255),
  identifier VARCHAR(255) NOT NULL,
  token VARCHAR(255) UNIQUE NOT NULL,
  type VARCHAR(50) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_user_id ON authula_verifications(user_id);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_identifier ON authula_verifications(identifier);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_type ON authula_verifications(type);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_expires_at ON authula_verifications(expires_at);`,
			)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(
				ctx,
				tx,
				`DROP TABLE IF EXISTS authula_verifications;`,
				`DROP TABLE IF EXISTS authula_sessions;`,
				`DROP TABLE IF EXISTS authula_accounts;`,
				`DROP TABLE IF EXISTS authula_users;`,
			)
		},
	}
}

func coreSQLiteAddMetadata() Migration {
	return Migration{
		Version: "20260127000000_core_add_metadata",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(ctx, tx, `ALTER TABLE authula_users ADD COLUMN metadata JSON NOT NULL DEFAULT '{}';`)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(ctx, tx, `ALTER TABLE authula_users DROP COLUMN metadata;`)
		},
	}
}

func corePostgresInitial() Migration {
	return Migration{
		Version: "20260126000000_core_initial",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(
				ctx,
				tx,
				`CREATE EXTENSION IF NOT EXISTS pgcrypto;`,
				`CREATE OR REPLACE FUNCTION core_set_updated_at_fn() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TABLE IF NOT EXISTS authula_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  email_verified BOOLEAN DEFAULT FALSE,
  image TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);`,
				`DROP TRIGGER IF EXISTS update_authula_users_updated_at_trigger ON authula_users;`,
				`CREATE TRIGGER update_authula_users_updated_at_trigger
  BEFORE UPDATE ON authula_users
  FOR EACH ROW
  EXECUTE FUNCTION core_set_updated_at_fn();`,
				`CREATE TABLE IF NOT EXISTS authula_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  account_id VARCHAR(255) NOT NULL,
  provider_id VARCHAR(255) NOT NULL,
  access_token TEXT,
  refresh_token TEXT,
  id_token TEXT,
  access_token_expires_at TIMESTAMP,
  refresh_token_expires_at TIMESTAMP,
  scope TEXT,
  password TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_authula_accounts_user FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE,
  CONSTRAINT unique_authula_provider_account UNIQUE(account_id, provider_id)
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_accounts_user_id ON authula_accounts(user_id);`,
				`DROP TRIGGER IF EXISTS update_authula_accounts_updated_at_trigger ON authula_accounts;`,
				`CREATE TRIGGER update_authula_accounts_updated_at_trigger
  BEFORE UPDATE ON authula_accounts
  FOR EACH ROW
  EXECUTE FUNCTION core_set_updated_at_fn();`,
				`CREATE TABLE IF NOT EXISTS authula_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  token VARCHAR(255) UNIQUE NOT NULL,
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_authula_sessions_user FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_sessions_user_id ON authula_sessions(user_id);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_sessions_expires_at ON authula_sessions(expires_at);`,
				`DROP TRIGGER IF EXISTS update_authula_sessions_updated_at_trigger ON authula_sessions;`,
				`CREATE TRIGGER update_authula_sessions_updated_at_trigger
  BEFORE UPDATE ON authula_sessions
  FOR EACH ROW
  EXECUTE FUNCTION core_set_updated_at_fn();`,
				`CREATE TABLE IF NOT EXISTS authula_verifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID,
  identifier VARCHAR(255) NOT NULL,
  token VARCHAR(255) UNIQUE NOT NULL,
  type VARCHAR(50) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_authula_verifications_user FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE
);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_user_id ON authula_verifications(user_id);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_identifier ON authula_verifications(identifier);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_type ON authula_verifications(type);`,
				`CREATE INDEX IF NOT EXISTS idx_authula_verifications_expires_at ON authula_verifications(expires_at);`,
				`DROP TRIGGER IF EXISTS update_authula_verifications_updated_at_trigger ON authula_verifications;`,
				`CREATE TRIGGER update_authula_verifications_updated_at_trigger
  BEFORE UPDATE ON authula_verifications
  FOR EACH ROW
  EXECUTE FUNCTION core_set_updated_at_fn();`,
				`CREATE OR REPLACE FUNCTION cleanup_expired_records_fn() RETURNS void AS $$
BEGIN
  DELETE FROM authula_sessions WHERE expires_at < NOW();
  DELETE FROM authula_verifications WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;`,
			)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(
				ctx,
				tx,
				`DROP TRIGGER IF EXISTS update_authula_verifications_updated_at_trigger ON authula_verifications;`,
				`DROP TRIGGER IF EXISTS update_authula_sessions_updated_at_trigger ON authula_sessions;`,
				`DROP TRIGGER IF EXISTS update_authula_accounts_updated_at_trigger ON authula_accounts;`,
				`DROP TRIGGER IF EXISTS update_authula_users_updated_at_trigger ON authula_users;`,
				`DROP TABLE IF EXISTS authula_verifications CASCADE;`,
				`DROP TABLE IF EXISTS authula_sessions CASCADE;`,
				`DROP TABLE IF EXISTS authula_accounts CASCADE;`,
				`DROP TABLE IF EXISTS authula_users CASCADE;`,
				`DROP FUNCTION IF EXISTS cleanup_expired_records_fn();`,
				`DROP FUNCTION IF EXISTS core_set_updated_at_fn();`,
			)
		},
	}
}

func corePostgresAddMetadata() Migration {
	return Migration{
		Version: "20260127000000_core_add_metadata",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(ctx, tx, `ALTER TABLE authula_users ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}'::JSONB;`)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(ctx, tx, `ALTER TABLE authula_users DROP COLUMN metadata;`)
		},
	}
}

func coreMySQLInitial() Migration {
	return Migration{
		Version: "20260126000000_core_initial",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(
				ctx,
				tx,
				`CREATE TABLE IF NOT EXISTS authula_users (
  id BINARY(16) NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  email_verified TINYINT(1) DEFAULT 0,
  image LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
				`CREATE TABLE IF NOT EXISTS authula_accounts (
  id BINARY(16) NOT NULL PRIMARY KEY,
  user_id BINARY(16) NOT NULL,
  account_id VARCHAR(255) NOT NULL,
  provider_id VARCHAR(255) NOT NULL,
  access_token LONGTEXT,
  refresh_token LONGTEXT,
  id_token LONGTEXT,
  access_token_expires_at TIMESTAMP NULL,
  refresh_token_expires_at TIMESTAMP NULL,
  scope TEXT,
  password TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_authula_accounts_user FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE,
  CONSTRAINT unique_authula_provider_account UNIQUE(account_id, provider_id),
  INDEX idx_authula_accounts_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
				`CREATE TABLE IF NOT EXISTS authula_sessions (
  id BINARY(16) NOT NULL PRIMARY KEY,
  user_id BINARY(16) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  token VARCHAR(255) UNIQUE NOT NULL,
  ip_address VARCHAR(45),
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_authula_sessions_user FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE,
  UNIQUE INDEX idx_authula_sessions_token (token),
  INDEX idx_authula_sessions_user_id (user_id),
  INDEX idx_authula_sessions_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
				`CREATE TABLE IF NOT EXISTS authula_verifications (
  id BINARY(16) NOT NULL PRIMARY KEY,
  user_id BINARY(16) NULL,
  identifier VARCHAR(255) NOT NULL,
  token VARCHAR(255) UNIQUE NOT NULL,
  type VARCHAR(50) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_authula_verifications_user FOREIGN KEY (user_id) REFERENCES authula_users(id) ON DELETE CASCADE,
  INDEX idx_authula_verifications_user_id (user_id),
  INDEX idx_authula_verifications_identifier (identifier),
  INDEX idx_authula_verifications_token (token),
  INDEX idx_authula_verifications_type (type),
  INDEX idx_authula_verifications_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
				`DROP PROCEDURE IF EXISTS cleanup_sessions;`,
				`DROP PROCEDURE IF EXISTS cleanup_verifications;`,
				`DROP PROCEDURE IF EXISTS cleanup_all_expired;`,
				`CREATE PROCEDURE cleanup_sessions()
BEGIN
  DELETE FROM authula_sessions WHERE expires_at < NOW();
END;`,
				`CREATE PROCEDURE cleanup_verifications()
BEGIN
  DELETE FROM authula_verifications WHERE expires_at < NOW();
END;`,
				`CREATE PROCEDURE cleanup_all_expired()
BEGIN
  CALL cleanup_sessions();
  CALL cleanup_verifications();
END;`,
			)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(
				ctx,
				tx,
				`DROP PROCEDURE IF EXISTS cleanup_all_expired;`,
				`DROP PROCEDURE IF EXISTS cleanup_verifications;`,
				`DROP PROCEDURE IF EXISTS cleanup_sessions;`,
				`DROP TABLE IF EXISTS authula_verifications;`,
				`DROP TABLE IF EXISTS authula_sessions;`,
				`DROP TABLE IF EXISTS authula_accounts;`,
				`DROP TABLE IF EXISTS authula_users;`,
			)
		},
	}
}

func coreMySQLAddMetadata() Migration {
	return Migration{
		Version: "20260127000000_core_add_metadata",
		Up: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(ctx, tx, `ALTER TABLE authula_users ADD COLUMN metadata JSON;`)
		},
		Down: func(ctx context.Context, tx bun.Tx) error {
			return ExecStatements(ctx, tx, `ALTER TABLE authula_users DROP COLUMN metadata;`)
		},
	}
}
