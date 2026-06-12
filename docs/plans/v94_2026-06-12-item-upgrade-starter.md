# v94 Plan — Item Upgrade Starter

Status: Ready for implementation
Goal: Add a guaranteed account-stash item upgrade that spends stash gold and increments one existing rolled stat.
Architecture: Tuning lives in `main_config.v0.json`. The HTTP layer validates ownership through the authenticated account and delegates mutation to a store transaction. The store locks account stash gold and the target stash item, spends gold, mutates the rolled JSON payload, and returns the updated item.
Tech stack: shared JSON/schema, Go rules loader, Go store, Go HTTP, lifecycle docs.

## Baseline and shortcut decision

Builds on v50 account stash, v68/v93 market ownership, and ADR-0012. No Godot plugin adoption applies because this slice has no client UI/art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Upgrade cost and max level |
| Modify | `shared/rules/main_config.v0.schema.json` | Config validation |
| Modify | `server/internal/game/rules.go` | Load config fields |
| Modify | `tools/validate_shared.py` | Cross-check config bounds |
| Modify | `server/internal/store/interfaces.go` | Upgrade repository method |
| Modify | `server/internal/store/repos.go` | Transactional stash item upgrade |
| Modify | `server/internal/store/store_test.go` | Store upgrade tests |
| Create | `server/internal/http/account_stash.go` | Upgrade HTTP route |
| Modify | `server/internal/http/server.go` | Register account-stash route |
| Modify | `server/internal/http/auth_session_test.go` | HTTP upgrade tests |
| Modify | `server/internal/replay/replay_test.go` | Fake repo method |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Create | `docs/as-built/v94_item-upgrade-starter.md` | As-built summary |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/store/repos.go`
- [x] `server/internal/store/store_test.go`
- [x] `server/internal/http/auth_session_test.go`
- [x] `tools/validate_shared.py`

Decision:
- [x] Defer extraction with rationale: v94 is narrow and touches existing over-limit test/repo files
  for focused coverage; the v93 baseline repair means maintainability must still pass.
- [x] Documented maintenance exception: v94 extends the same over-limit store/http/validator files.
  Baseline was updated to current line counts; future market/upgrade work should split these files
  before adding more behavior.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Config

- [x] Add `item_upgrade_cost_gold` and `item_upgrade_max_level`.
- [x] Load and validate both fields.

```bash
make validate-shared
```

## Task 2 — Store Upgrade

- [x] Add a repository method that locks stash gold and item rows.
- [x] Reject missing items, insufficient gold, max-level items, and no numeric stat to upgrade.
- [x] Increment the first deterministic numeric rolled stat and `item_level`.
- [x] Add store tests.

```bash
cd server && go test ./internal/store -run TestAccountStashItemUpgrade -count=1
```

## Task 3 — HTTP Route

- [x] Add `POST /v0/account-stash/items/{stash_item_id}/upgrade`.
- [x] Register the route and map store errors to HTTP responses.
- [x] Add authenticated route tests.

```bash
cd server && go test ./internal/http -run TestAccountStashItemUpgrade -count=1
```

## Task 4 — Lifecycle Docs And CI

- [x] Update plan checkboxes, `PROGRESS.md`, and as-built.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make maintainability`
- [x] `cd server && go test ./internal/store -run TestAccountStashItemUpgrade -count=1`
- [x] `cd server && go test ./internal/http -run TestAccountStashItemUpgrade -count=1`
- [x] `make test-go`
- [x] `make ci`

## Deferred scope

Advanced resource costs, failure chances, recipe tiers, random affix addition, blacksmith UI/NPC,
upgrade audit history, and market restrictions for upgraded items remain deferred.
