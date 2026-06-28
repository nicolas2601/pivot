CREATE TABLE goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    target_amount BIGINT NOT NULL CHECK (target_amount > 0),
    current_amount BIGINT NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'COP',
    deadline DATE,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    color VARCHAR(7),
    notes TEXT,
    is_completed BOOLEAN NOT NULL DEFAULT false,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_goals_current_lte_target CHECK (current_amount <= target_amount)
);

CREATE INDEX idx_goals_user_id ON goals(user_id);
CREATE INDEX idx_goals_user_completed ON goals(user_id, is_completed);

CREATE TRIGGER set_goals_updated_at
BEFORE UPDATE ON goals
FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
