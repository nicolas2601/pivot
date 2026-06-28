DROP TRIGGER IF EXISTS set_goals_updated_at ON goals;
DROP INDEX IF EXISTS idx_goals_user_completed;
DROP INDEX IF EXISTS idx_goals_user_id;
DROP TABLE IF EXISTS goals;
