package database

import (
	"fmt"

	"gorm.io/gorm"
)

func migrateAccountUserForeignKey(db *gorm.DB) error {
	const migration = `
DO $$
DECLARE
    fk_record RECORD;
    user_id_attnum SMALLINT;
BEGIN
    SELECT attnum
    INTO user_id_attnum
    FROM pg_attribute
    WHERE attrelid = 'accounts'::regclass
      AND attname = 'user_id'
      AND NOT attisdropped;

    IF user_id_attnum IS NULL THEN
        RAISE EXCEPTION 'accounts.user_id column does not exist';
    END IF;

    -- تمام Foreign Keyهای موجود بین accounts.user_id و users.id
    -- حذف می‌شوند تا Constraint قدیمی CASCADE باقی نماند.
    FOR fk_record IN
        SELECT conname
        FROM pg_constraint
        WHERE contype = 'f'
          AND conrelid = 'accounts'::regclass
          AND confrelid = 'users'::regclass
          AND conkey = ARRAY[user_id_attnum]
    LOOP
        EXECUTE format(
            'ALTER TABLE accounts DROP CONSTRAINT %I',
            fk_record.conname
        );
    END LOOP;

    -- Constraint جدید با رفتار RESTRICT ساخته می‌شود.
    ALTER TABLE accounts
    ADD CONSTRAINT fk_accounts_user_restrict
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT;
END
$$;
`

	if err := db.Exec(migration).Error; err != nil {
		return fmt.Errorf(
			"migrate accounts user foreign key to RESTRICT: %w",
			err,
		)
	}

	return nil
}
