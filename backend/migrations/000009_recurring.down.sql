DROP INDEX IF EXISTS idx_recurring_runs_status;
DROP INDEX IF EXISTS idx_recurring_runs_rule_id;
DROP INDEX IF EXISTS idx_recurring_runs_user_id;
DROP TABLE IF EXISTS recurring_runs;
DROP TRIGGER IF EXISTS set_recurring_rules_updated_at ON recurring_rules;
DROP INDEX IF EXISTS idx_recurring_rules_due;
DROP INDEX IF EXISTS idx_recurring_rules_user_active;
DROP INDEX IF EXISTS idx_recurring_rules_user_id;
DROP TABLE IF EXISTS recurring_rules;
