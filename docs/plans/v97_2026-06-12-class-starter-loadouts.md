# v97 Plan — Class starter loadouts

Status: Complete
Goal: Newly created heroes receive durable class starter equipment plus one health and one mana potion.
Architecture: Starter loadouts are seeded in the server character-creation path using shared item-template data and durable `character_item_instances`. Rolled starter weapons use existing roll payload semantics so sessions, persistence, stats, bot scenarios, and the client observe them through current snapshot fields. `skill_damage_percent` becomes a schema-backed item stat consumed by skill damage resolution; no protocol schema bump is required because existing item/stat payload maps carry the value.
Tech stack: shared JSON rules/assets, Go HTTP/store/game tests, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v96 `town-presentation-polish` and reuses v69-v71 class identity, v23/v28 rolled equipment persistence, and v44+ skill damage paths. Godot plugin adoption: reject for this slice because no new client UI/art system is needed; placeholder item visuals/presentations are already manifest-driven and sufficient for protocol-visible starter gear.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/item_templates.v0.json` | Add starter weapon/shield templates and caster stat rolls. |
| Modify | `shared/rules/item_templates.v0.schema.json` | Allow starter item types and `skill_damage_percent`. |
| Modify | `shared/assets/item_visuals.v0.json` | Map starter template ids to existing placeholder equipment assets. |
| Modify | `shared/assets/item_presentations.v0.json` | Add icon/ground presentation for starter template ids. |
| Modify/Create | `server/internal/http/*` | Seed class starter items on explicit character creation. |
| Modify | `server/internal/game/sim.go` | Apply equipped `skill_damage_percent` to skill damage. |
| Modify/Create | `server/internal/http/*_test.go`, `server/internal/game/*_test.go` | Focused server proof. |
| Modify/Create | `tools/bot/scenarios/*.json`, `tools/bot/test_protocol.py` | Protocol bot proof for starter loadout. |
| Add | `docs/as-built/v97_class-starter-loadouts.md` | As-built summary. |
| Modify | `PROGRESS.md` | Lifecycle close-out. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/http/auth_session_test.go`
- [x] `tools/bot/test_protocol.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `server/internal/store/repos.go` only if needed

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: changes are narrow additions to established creation, test, and damage paths; splitting these legacy files now would add more risk than the slice warrants.

Verification:
```bash
make maintainability
```

## Task 1 — Shared starter templates

Files:
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/assets/item_visuals.v0.json`
- Modify: `shared/assets/item_presentations.v0.json`

- [x] Step 1.1: Add `starter_paladin_sword`, `starter_paladin_shield`, `starter_sorcerer_staff`, and `starter_barbarian_axe` templates with common-compatible base stats.
- [x] Step 1.2: Add `axe`, `staff`, and `skill_damage_percent` schema support.
- [x] Step 1.3: Add placeholder visual and presentation entries for starter template ids.
```bash
make validate-shared
```

## Task 2 — Server starter seeding

Files:
- Modify/Create: `server/internal/http/character.go` or adjacent helper
- Inspect: `server/internal/http/session.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 2.1: Seed durable starter items immediately after successful explicit character creation.
- [x] Step 2.2: Keep session creation as a load-only path for explicit created-character items; do not backfill compatibility defaults.
- [x] Step 2.3: Ensure seeding is idempotent when a character already has items.
- [x] Step 2.4: Add tests for barbarian, sorcerer, and paladin starter equipment/potions.
```bash
cd server && go test ./internal/http -run 'TestCharacterClassSeedsSessionStartProgression|TestCreatedCharactersReceiveClassStarterLoadouts'
```

## Task 3 — Skill damage item stat

Files:
- Modify: `server/internal/game/sim.go`
- Modify/Create: `server/internal/game/*_test.go`

- [x] Step 3.1: Sum equipped base and rolled `skill_damage_percent` from item templates.
- [x] Step 3.2: Apply the percent to projectile/cold/chain skill damage ranges before mitigation.
- [x] Step 3.3: Add a deterministic test proving an equipped staff stat raises magic-bolt damage.
```bash
cd server && go test ./internal/game -run TestStarterStaffAddsMaxManaAndSkillDamage
```

## Task 4 — Bot scenario

Files:
- Add/Modify: `tools/bot/scenarios/46_class_starter_loadout.json`
- Modify: `tools/bot/test_protocol.py` if discovery assertions are needed

- [x] Step 4.1: Add a sorcerer scenario that creates a sorcerer, starts a session, and asserts equipped staff, empty offhand, red potion, and blue potion.
- [x] Step 4.2: Add scenario discovery/unit assertions if needed.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=46_class_starter_loadout
```

## Task 5 — Lifecycle docs and CI

Files:
- Modify: `docs/specs/v97_spec-class-starter-loadouts.md`
- Modify: `docs/plans/v97_2026-06-12-class-starter-loadouts.md`
- Add: `docs/as-built/v97_class-starter-loadouts.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and add as-built notes.
- [x] Step 5.2: Update `PROGRESS.md` current status and lifecycle row.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/http -run 'TestCharacterClassSeedsSessionStartProgression|TestCreatedCharactersReceiveClassStarterLoadouts'`
- [x] `cd server && go test ./internal/game -run TestStarterStaffAddsMaxManaAndSkillDamage`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=46_class_starter_loadout`
- [x] `make ci`

## Deferred scope

- Existing-character backfill, including login-created compatibility default characters.
- Dedicated axe/staff/shield models and richer item icons.
- Displaying `skill_damage_percent` as a top-level derived character stat.
