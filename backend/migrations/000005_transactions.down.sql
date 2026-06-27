DROP TRIGGER IF EXISTS set_transactions_updated_at ON transactions;
DROP INDEX IF EXISTS idx_transactions_transfer_pair_id;
DROP INDEX IF EXISTS idx_transactions_date;
DROP INDEX IF EXISTS idx_transactions_category_id;
DROP INDEX IF EXISTS idx_transactions_account_id;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP TABLE IF EXISTS transactions;