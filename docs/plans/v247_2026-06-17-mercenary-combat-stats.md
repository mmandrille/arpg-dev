# v247 Plan - Mercenary Combat Stats

Status: Complete
Goal: Show server-owned combat stats for hired mercenaries in the existing panel.
Architecture: Add a compact companion-only `combat_stats` protocol view, copy it through the
client companion sync, and extend the existing stats card renderer.
Tech stack: Go sim/protocol, Godot UI/client bot, shared schema, docs.

## Baseline and Asset Decision

Builds on v239 mercenary stats card and v245-v246 client panel patterns.

Asset/plugin decision:
- Adopt existing text-first stats-card UI.
- Borrow authoritative companion combat fields from the server entity.
- Reject external assets/plugins and icon work.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/types.go` | Add optional companion combat stat view |
| Add | `server/internal/game/companion_combat_stats_view.go` | Build compact combat stats from companion state |
| Modify | `server/internal/game/sim.go` | Attach combat stats to companion entity views |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow `combat_stats` in entity schema |
| Modify | `client/scripts/main.gd` | Preserve `combat_stats` and pass it to companion UI state |
| Modify | `client/scripts/mercenary_panel.gd` | Render combat-stat lines |
| Modify | `client/tests/test_mercenary_panel.gd` | Prove card rendering |
| Add | `tools/bot/scenarios/client/64_mercenary_combat_stats.json` | Client proof |
| Modify | `server/internal/store/repos.go` | Delete resource-wallet snapshots during stale-session cleanup |
| Add | `docs/as-built/v247_mercenary-combat-stats.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines, except grandfathered baselines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/types.go`
- [x] `server/internal/game/sim.go`
- [x] `client/scripts/main.gd`

Decision:
- [x] Keep `sim.go` and `main.gd` edits line-neutral.
- [x] Put server combat-stat construction in a new helper file.

Verification:
```bash
make maintainability
```

## Task 1 - Server/protocol combat stats

Files:
- Modify: `server/internal/game/types.go`
- Add: `server/internal/game/companion_combat_stats_view.go`
- Modify: `server/internal/game/sim.go`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `server/internal/game/companion_ai_test.go`

- [x] Add optional companion `combat_stats` fields for damage min/max, cooldown, armor, block,
  hit chance, and crit chance.
- [x] Populate the view from actual companion state, with shared monster rules as fallback.
- [x] Prove the fixed guard view mirrors `shared/rules/monsters.v0.json`.

```bash
cd server && go test ./internal/game -run 'MercenaryFoundation|MercenaryHiring' -count=1
make validate-shared
```

## Task 2 - Client panel and bot proof

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/mercenary_panel.gd`
- Modify: `client/tests/test_mercenary_panel.gd`
- Add: `tools/bot/scenarios/client/64_mercenary_combat_stats.json`

- [x] Preserve companion `combat_stats` in client entity records and companion UI state.
- [x] Render combat lines in the mercenary stats card when present.
- [x] Add focused Godot and bot coverage.

```bash
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot-client scenario=64_mercenary_combat_stats.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v247_mercenary-combat-stats.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `cd server && go test ./internal/game -run 'MercenaryFoundation|MercenaryHiring' -count=1`
- [x] `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- [x] `make bot-client scenario=64_mercenary_combat_stats.json HEADLESS=1`
- [x] `make validate-shared`
- [x] `make maintainability`
