CREATE TABLE merchants
(
    id         TEXT PRIMARY KEY,
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_merchants_email ON merchants (email);

CREATE TABLE invoices
(
    id          TEXT PRIMARY KEY,
    merchant_id TEXT        NOT NULL REFERENCES merchants (id),
    amount      BIGINT      NOT NULL,
    currency    TEXT        NOT NULL,
    status      TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invoices_merchant_id ON invoices (merchant_id);
CREATE INDEX idx_invoices_status ON invoices (status);

CREATE TABLE payments
(
    id               TEXT PRIMARY KEY,
    invoice_id       TEXT        NOT NULL REFERENCES invoices (id),
    amount           BIGINT      NOT NULL,
    currency         TEXT        NOT NULL,
    status           TEXT        NOT NULL,
    transaction_hash TEXT,
    finalized_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_invoice_id ON payments (invoice_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE UNIQUE INDEX idx_payments_transaction_hash
    ON payments (transaction_hash) WHERE transaction_hash IS NOT NULL;
