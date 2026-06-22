CREATE TABLE settlement_batches
(
    id           text primary key,
    recipient_id text        not null,
    currency     text        not null,
    total_amount bigint      not null check ( total_amount > 0 ),
    item_count   integer     not null check ( item_count > 0 ),
    status       text        not null check
        ( status in ('DRAFT', 'READY', 'APPROVED', 'PROCESSING', 'PAID', 'FAILED', 'CANCELED')),
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now()
);
-- 수취인, 통화, 상태별 정산 묶음 조회를 빠르게 하기 위한 복합 인덱스다.
CREATE INDEX idx_settlement_batches_recipient_currency_status
    ON settlement_batches (recipient_id, currency, status);

CREATE TABLE settlement_items
(
    batch_id        text   not null references settlement_batches (id),
    ledger_entry_id text   not null references ledger_entries (id),
    amount          bigint not null check ( amount > 0 ),
    currency        text   not null,
    primary key (batch_id, ledger_entry_id)
);

CREATE UNIQUE INDEX idx_settlement_items_ledger_entry_id
    ON settlement_items (ledger_entry_id);
