CREATE TABLE travel_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    currency VARCHAR(3) NOT NULL DEFAULT 'COP',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE travel_group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES travel_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, user_id)
);

CREATE TABLE travel_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES travel_groups(id) ON DELETE CASCADE,
    paid_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'COP',
    description VARCHAR(255) NOT NULL,
    split_method VARCHAR(20) NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE travel_expense_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID NOT NULL REFERENCES travel_expenses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    UNIQUE (expense_id, user_id)
);

CREATE TABLE travel_settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES travel_groups(id) ON DELETE CASCADE,
    from_user UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'COP',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ
);

CREATE INDEX idx_travel_groups_created_by ON travel_groups(created_by);
CREATE INDEX idx_travel_group_members_group_id ON travel_group_members(group_id);
CREATE INDEX idx_travel_group_members_user_id ON travel_group_members(user_id);
CREATE INDEX idx_travel_expenses_group_id ON travel_expenses(group_id);
CREATE INDEX idx_travel_expenses_paid_by ON travel_expenses(paid_by);
CREATE INDEX idx_travel_expenses_date ON travel_expenses(group_id, date DESC);
CREATE INDEX idx_travel_expense_shares_expense_id ON travel_expense_shares(expense_id);
CREATE INDEX idx_travel_expense_shares_user_id ON travel_expense_shares(user_id);
CREATE INDEX idx_travel_settlements_group_id ON travel_settlements(group_id);
CREATE INDEX idx_travel_settlements_from_user ON travel_settlements(from_user);
CREATE INDEX idx_travel_settlements_to_user ON travel_settlements(to_user);

CREATE TRIGGER set_travel_groups_updated_at
BEFORE UPDATE ON travel_groups
FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();