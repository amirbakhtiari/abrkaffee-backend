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

func migrateOrderTables(db *gorm.DB) error {
	queries := []string{
		// ۱. جدول شمارنده روزانه سفارش‌ها
		`CREATE TABLE IF NOT EXISTS order_daily_sequences (
			order_date DATE PRIMARY KEY,
			last_number BIGINT NOT NULL DEFAULT 0,
			CONSTRAINT chk_order_daily_sequences_number CHECK (last_number >= 0)
		);`,

		// ۲. سربرگ سفارش
		`CREATE TABLE IF NOT EXISTS orders (
			id BIGSERIAL PRIMARY KEY,
			idempotency_key UUID,
			order_date DATE NOT NULL,
			daily_number BIGINT NOT NULL,
			table_id BIGINT NOT NULL REFERENCES tables(id) ON DELETE RESTRICT,
			customer_phone VARCHAR(20),
			customer_name VARCHAR(100),
			customer_note TEXT,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			payment_status VARCHAR(20) NOT NULL DEFAULT 'UNPAID',
			total_amount BIGINT NOT NULL DEFAULT 0,
			version INT NOT NULL DEFAULT 1,
			registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			settled_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,

			CONSTRAINT chk_orders_daily_number CHECK (daily_number > 0),
			CONSTRAINT chk_orders_total_amount CHECK (total_amount >= 0),
			CONSTRAINT uq_orders_daily_number UNIQUE (order_date, daily_number)
		);`,

		// ایندکس‌های سربرگ سفارش
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_idempotency_key 
		ON orders (idempotency_key) 
		WHERE idempotency_key IS NOT NULL;`,

		`CREATE INDEX IF NOT EXISTS idx_orders_table_status ON orders (table_id, status);`,
		`CREATE INDEX IF NOT EXISTS idx_orders_date ON orders (order_date);`,

		// ۳. اقلام سفارش (با اسنپ‌شات کامل کاتالوگ و مالی)
		`CREATE TABLE IF NOT EXISTS order_items (
			id BIGSERIAL PRIMARY KEY,
			order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
			category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
			price_announcement_item_id BIGINT,
			
			product_name VARCHAR(255) NOT NULL,
			product_code VARCHAR(64) NOT NULL,
			category_title VARCHAR(100) NOT NULL,
			
			unit_price BIGINT NOT NULL,
			quantity INT NOT NULL,
			total_price BIGINT NOT NULL,
			notes VARCHAR(255),
			
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			
			CONSTRAINT chk_order_items_quantity CHECK (quantity > 0),
			CONSTRAINT chk_order_items_unit_price CHECK (unit_price >= 0),
			CONSTRAINT chk_order_items_total_price CHECK (total_price >= 0)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);`,

		// ۴. تراکنش‌های پرداخت
		`CREATE TABLE IF NOT EXISTS order_payments (
			id BIGSERIAL PRIMARY KEY,
			order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
			amount BIGINT NOT NULL,
			payment_method VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			transaction_reference VARCHAR(100),
			paid_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			
			CONSTRAINT chk_order_payments_amount CHECK (amount > 0),
			CONSTRAINT chk_order_payments_method CHECK (payment_method IN ('CASH', 'POS', 'ONLINE')),
			CONSTRAINT chk_order_payments_status CHECK (status IN ('PENDING', 'SUCCESSFUL', 'FAILED', 'REFUNDED'))
		);`,

		`CREATE INDEX IF NOT EXISTS idx_order_payments_order_id ON order_payments (order_id);`,

		// ۵. تاریخچه تغییرات وضعیت (Audit Trail)
		`CREATE TABLE IF NOT EXISTS order_status_histories (
			id BIGSERIAL PRIMARY KEY,
			order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
			from_status VARCHAR(20),
			to_status VARCHAR(20) NOT NULL,
			changed_by BIGINT,
			note TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_order_status_histories_order ON order_status_histories (order_id);`,
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			return fmt.Errorf("failed executing order migration query: %w", err)
		}
	}

	return nil
}
