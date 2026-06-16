-- 0027_account_resource_wallet: account-wide material/resource balances plus
-- immutable session-start resource snapshots for replay.

CREATE TABLE IF NOT EXISTS account_resource_wallet (
    account_id   TEXT NOT NULL REFERENCES accounts(id),
    resource_id  TEXT NOT NULL,
    amount       INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, resource_id),
    CHECK (amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_account_resource_wallet_account
    ON account_resource_wallet(account_id, resource_id);

CREATE TABLE IF NOT EXISTS session_start_account_resource_wallet (
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    account_id  TEXT NOT NULL REFERENCES accounts(id),
    resource_id TEXT NOT NULL,
    amount      INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (session_id, account_id, resource_id),
    CHECK (amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_session_start_account_resource_wallet_account
    ON session_start_account_resource_wallet(session_id, account_id, resource_id);
