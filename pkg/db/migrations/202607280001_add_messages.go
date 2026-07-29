package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addMessages() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202607280001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE TABLE IF NOT EXISTS messages (
				id            VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
				adapter_name  VARCHAR(255) NOT NULL,
				kind          VARCHAR(100) NOT NULL,
				resource_id   VARCHAR(255) NOT NULL,
				payload       JSONB NOT NULL,
				status        VARCHAR(20) NOT NULL DEFAULT 'pending',
				created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				claimed_at    TIMESTAMPTZ NULL,
				completed_at  TIMESTAMPTZ NULL,
				error_message TEXT NULL
			);`).Error; err != nil {
				return err
			}

			if err := tx.Exec(
				"CREATE INDEX IF NOT EXISTS idx_messages_adapter_status " +
					"ON messages (adapter_name, status) WHERE status = 'pending';",
			).Error; err != nil {
				return err
			}

			if err := tx.Exec(
				"CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_dedup " +
					"ON messages (adapter_name, resource_id) " +
					"WHERE status IN ('pending', 'claimed');",
			).Error; err != nil {
				return err
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("DROP TABLE IF EXISTS messages;").Error
		},
	}
}
