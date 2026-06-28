CREATE TABLE recurring_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('expense', 'income')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'COP',
    description VARCHAR(255),
    notes TEXT,
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'biweekly', 'monthly', 'yearly')),
    interval_count INTEGER NOT NULL DEFAULT 1 CHECK (interval_count > 0),
    start_date DATE NOT NULL,
    end_date DATE,
    last_run_date DATE,
    next_run_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurring_rules_user_id ON recurring_rules(user_id);
CREATE INDEX idx_recurring_rules_user_active ON recurring_rules(user_id, is_active);
CREATE INDEX idx_recurring_rules_due ON recurring_rules(is_active, next_run_date)
    WHERE is_active = true;

CREATE TRIGGER set_recurring_rules_updated_at
BEFORE UPDATE ON recurring_rules
FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();

CREATE TABLE recurring_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recurring_rule_id UUID NOT NULL REFERENCES recurring_rules(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scheduled_date DATE NOT NULL,
    executed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'executed', 'skipped', 'failed')),
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One run per rule per scheduled date (idempotent generation).
    CONSTRAINT uq_recurring_runs_rule_date UNIQUE (recurring_rule_id, scheduled_date)
);

CREATE INDEX idx_recurring_runs_user_id ON recurring_runs(user_id);
CREATE INDEX idx_recurring_runs_rule_id ON recurring_runs(recurring_rule_id);
CREATE INDEX idx_recurring_runs_status ON recurring_runs(user_id, status, scheduled_date);
