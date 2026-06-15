CREATE TABLE ledger_accounts
(
    id         TEXT PRIMARY KEY,
    type       TEXT        NOT NULL,
    owner_id   TEXT        NOT NULL,
    currency   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_accounts_owner_id ON ledger_accounts (owner_id);
CREATE INDEX idx_ledger_accounts_type ON ledger_accounts (type);

CREATE TABLE ledger_transactions
(
    id              TEXT PRIMARY KEY,
    reference_type  TEXT        NOT NULL,
    reference_id    TEXT        NOT NULL,
    idempotency_key TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_transactions_reference
    ON ledger_transactions (reference_type, reference_id);

CREATE UNIQUE INDEX idx_ledger_transactions_idempotency_key
    ON ledger_transactions (idempotency_key);

CREATE TABLE ledger_entries
(
    id             TEXT PRIMARY KEY,
    transaction_id TEXT        NOT NULL REFERENCES ledger_transactions (id),
    account_id     TEXT        NOT NULL REFERENCES ledger_accounts (id),
    direction      TEXT        NOT NULL,
    amount         BIGINT      NOT NULL,
    currency       TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_entries_transaction_id
    ON ledger_entries (transaction_id);

CREATE INDEX idx_ledger_entries_account_id
    ON ledger_entries (account_id);