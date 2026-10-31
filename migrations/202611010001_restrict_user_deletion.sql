-- Prevent deleting a user that still has an account.
-- Run this migration after checking the existing FK and before deploying the application change.
DO $$
DECLARE
    fk_name text;
BEGIN
    SELECT tc.constraint_name
      INTO fk_name
      FROM information_schema.table_constraints tc
      JOIN information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
       AND tc.constraint_schema = kcu.constraint_schema
     WHERE tc.constraint_type = 'FOREIGN KEY'
       AND tc.table_schema = current_schema()
       AND tc.table_name = 'accounts'
       AND kcu.column_name = 'user_id'
     LIMIT 1;

    IF fk_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE accounts DROP CONSTRAINT %I', fk_name);
    END IF;

    ALTER TABLE accounts
      ADD CONSTRAINT fk_accounts_user_id
      FOREIGN KEY (user_id) REFERENCES users(id)
      ON UPDATE CASCADE
      ON DELETE RESTRICT;
END $$;
