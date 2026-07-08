-- 001_initial_schema.sql
-- Initial database schema for AI Finance Tracker
-- Applied automatically by GORM AutoMigrate, but kept here for reference
-- and for manual production deployments.

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE IF NOT EXISTS users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               VARCHAR(255) UNIQUE NOT NULL,
    name                VARCHAR(255) NOT NULL,
    password_hash       TEXT,
    profile_picture     TEXT,
    is_email_verified   BOOLEAN DEFAULT FALSE,
    email_verify_token  TEXT,
    reset_token         TEXT,
    reset_token_expiry  TIMESTAMPTZ,
    google_id           VARCHAR(255),
    timezone            VARCHAR(100) DEFAULT 'Asia/Kolkata',
    currency            VARCHAR(10)  DEFAULT 'INR',
    preferred_language  VARCHAR(10)  DEFAULT 'en',
    created_at          TIMESTAMPTZ  DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);

-- Sessions
CREATE TABLE IF NOT EXISTS sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token        TEXT,
    refresh_token       TEXT UNIQUE,
    expires_at          TIMESTAMPTZ NOT NULL,
    refresh_expires_at  TIMESTAMPTZ NOT NULL,
    ip_address          VARCHAR(45),
    user_agent          TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id       ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_access_token  ON sessions(access_token);

-- Expenses
CREATE TABLE IF NOT EXISTS expenses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount          NUMERIC(12,2) NOT NULL,
    currency        VARCHAR(10) DEFAULT 'INR',
    category        VARCHAR(50)  NOT NULL,
    merchant        VARCHAR(255),
    description     TEXT,
    notes           TEXT,
    date            TIMESTAMPTZ  NOT NULL,
    expense_type    VARCHAR(20)  DEFAULT 'spend',
    payment_method  VARCHAR(20),
    image_url       TEXT,
    ocr_data        JSONB,
    is_duplicate    BOOLEAN DEFAULT FALSE,
    duplicate_of    UUID,
    is_favorite     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_expenses_user_date     ON expenses(user_id, date);
CREATE INDEX IF NOT EXISTS idx_expenses_user_category ON expenses(user_id, category);
CREATE INDEX IF NOT EXISTS idx_expenses_user_merchant ON expenses(user_id, merchant);

-- Incomes
CREATE TABLE IF NOT EXISTS incomes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount          NUMERIC(12,2) NOT NULL,
    currency        VARCHAR(10)  DEFAULT 'INR',
    source          VARCHAR(255),
    category        VARCHAR(100),
    description     TEXT,
    notes           TEXT,
    date            TIMESTAMPTZ  NOT NULL,
    payment_method  VARCHAR(20),
    is_taxable      BOOLEAN DEFAULT FALSE,
    tax_amount      NUMERIC(12,2) DEFAULT 0,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_incomes_user_date ON incomes(user_id, date);

-- Budgets
CREATE TABLE IF NOT EXISTS budgets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category    VARCHAR(50)  NOT NULL,
    amount      NUMERIC(12,2) NOT NULL,
    currency    VARCHAR(10)  DEFAULT 'INR',
    period      VARCHAR(20)  DEFAULT 'monthly',
    month       SMALLINT     DEFAULT 0,
    year        SMALLINT     NOT NULL,
    alert_at    NUMERIC(5,2) DEFAULT 80,
    description TEXT,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_budgets_user_category ON budgets(user_id, category);

-- Tags
CREATE TABLE IF NOT EXISTS tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    color       VARCHAR(20)  DEFAULT '#6366f1',
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Expense-Tag join
CREATE TABLE IF NOT EXISTS expense_tags (
    expense_id  UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    tag_id      UUID NOT NULL REFERENCES tags(id)     ON DELETE CASCADE,
    PRIMARY KEY (expense_id, tag_id)
);

-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    amount              NUMERIC(12,2) NOT NULL,
    currency            VARCHAR(10)  DEFAULT 'INR',
    billing_cycle       VARCHAR(20)  DEFAULT 'monthly',
    next_billing_date   TIMESTAMPTZ  NOT NULL,
    category            VARCHAR(50),
    payment_method      VARCHAR(20),
    notes               TEXT,
    is_active           BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ  DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_subs_user_billing ON subscriptions(user_id, next_billing_date);

-- Goals
CREATE TABLE IF NOT EXISTS goals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    target_amount   NUMERIC(12,2) NOT NULL,
    current_amount  NUMERIC(12,2) DEFAULT 0,
    currency        VARCHAR(10)   DEFAULT 'INR',
    category        VARCHAR(100),
    target_date     TIMESTAMPTZ   NOT NULL,
    priority        SMALLINT      DEFAULT 3,
    status          VARCHAR(20)   DEFAULT 'active',
    created_at      TIMESTAMPTZ   DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    message     TEXT,
    type        VARCHAR(50),
    is_read     BOOLEAN DEFAULT FALSE,
    data        JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notifs_user_read ON notifications(user_id, is_read);

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action      VARCHAR(20)  NOT NULL,
    entity      VARCHAR(100) NOT NULL,
    entity_id   UUID,
    old_data    JSONB,
    new_data    JSONB,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_user_entity ON audit_logs(user_id, entity);

-- Merchant Mappings
CREATE TABLE IF NOT EXISTS merchant_mappings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant    VARCHAR(255) UNIQUE NOT NULL,
    category    VARCHAR(50)  NOT NULL,
    aliases     TEXT[],
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Seed well-known Indian merchants
INSERT INTO merchant_mappings (merchant, category, aliases) VALUES
    ('Swiggy',      'food',           ARRAY['swiggy food', 'swiggypop']),
    ('Zomato',      'food',           ARRAY['zomato order', 'zomato gold']),
    ('Dominos',     'food',           ARRAY['domino''s pizza', 'dominos pizza']),
    ('McDonalds',   'food',           ARRAY['mcdonald''s', 'mcd']),
    ('Starbucks',   'food',           ARRAY['starbucks coffee']),
    ('BigBasket',   'food',           ARRAY['bigbasket', 'bb daily']),
    ('Blinkit',     'food',           ARRAY['blinkit', 'grofers']),
    ('Amazon',      'shopping',       ARRAY['amazon.in', 'amazon prime']),
    ('Flipkart',    'shopping',       ARRAY['flipkart.com', 'fk']),
    ('Myntra',      'shopping',       ARRAY['myntra.com']),
    ('Uber',        'travel',         ARRAY['uber cab', 'uber auto', 'ubereats']),
    ('Rapido',      'travel',         ARRAY['rapido cab', 'rapido bike']),
    ('Ola',         'travel',         ARRAY['ola cabs', 'ola electric']),
    ('Netflix',     'subscription',   ARRAY['netflix.com']),
    ('Spotify',     'subscription',   ARRAY['spotify premium']),
    ('Amazon Prime','subscription',   ARRAY['prime video', 'amazon prime']),
    ('Hotstar',     'subscription',   ARRAY['disney+ hotstar', 'hotstar premium']),
    ('ChatGPT',     'subscription',   ARRAY['openai', 'chatgpt plus']),
    ('Apollo',      'health',         ARRAY['apollo pharmacy', 'apollo 247']),
    ('PharmEasy',   'health',         ARRAY['pharmeasy.in']),
    ('1mg',         'health',         ARRAY['tata 1mg'])
ON CONFLICT (merchant) DO NOTHING;

-- Recurring Expenses
CREATE TABLE IF NOT EXISTS recurring_expenses (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant         VARCHAR(255),
    amount           NUMERIC(12,2),
    currency         VARCHAR(10)  DEFAULT 'INR',
    category         VARCHAR(50),
    frequency        VARCHAR(20),
    last_occurrence  TIMESTAMPTZ,
    next_occurrence  TIMESTAMPTZ,
    confidence       NUMERIC(4,3) DEFAULT 0,
    is_approved      BOOLEAN DEFAULT FALSE,
    is_active        BOOLEAN DEFAULT TRUE,
    created_at       TIMESTAMPTZ  DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

-- Financial Health Scores
CREATE TABLE IF NOT EXISTS financial_health_scores (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score                NUMERIC(5,2)  DEFAULT 0,
    income_score         NUMERIC(5,2)  DEFAULT 0,
    savings_score        NUMERIC(5,2)  DEFAULT 0,
    expense_ratio        NUMERIC(5,2)  DEFAULT 0,
    budget_health        NUMERIC(5,2)  DEFAULT 0,
    debt_health          NUMERIC(5,2)  DEFAULT 0,
    subscription_health  NUMERIC(5,2)  DEFAULT 0,
    insights             JSONB,
    created_at           TIMESTAMPTZ   DEFAULT NOW(),
    updated_at           TIMESTAMPTZ   DEFAULT NOW()
);
