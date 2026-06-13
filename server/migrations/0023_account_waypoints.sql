-- 0023_account_waypoints: account-bound teleporter discoveries.

CREATE TABLE IF NOT EXISTS account_waypoints (
    account_id     TEXT NOT NULL REFERENCES accounts(id),
    level          INTEGER NOT NULL,
    discovered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, level)
);

INSERT INTO account_waypoints (account_id, level, discovered_at)
SELECT c.account_id, cw.level, MIN(cw.discovered_at)
FROM character_waypoints cw
JOIN characters c ON c.id = cw.character_id
GROUP BY c.account_id, cw.level
ON CONFLICT (account_id, level) DO NOTHING;
