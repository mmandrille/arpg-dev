# v188 Plan — Live Rare Combat Affixes

Status: Complete - make ci green on 2026-06-15
Goal: Make rare hit, crit, evade, and attack-speed affixes live in authoritative combat.
Architecture: Shared item-template data owns roll eligibility and bounded values. The Go sim owns
stat aggregation and combat resolution. Existing combat events represent hit, crit, and miss, so no
protocol version bump is required.

## Baseline and shortcut decision

Builds on v187 rarity roll pools and v31 combat stat effects. Godot plugin decision: reject new
client/plugin work; existing character stats panel and combat feedback are sufficient.

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/game_test.go`
- [x] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Touched grandfathered files must stay at or below baseline + allowance.

Verification:
```bash
make maintainability
```

## Task 1 — Shared rare affix data

Files:
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/golden/item_rolls.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add supported roll stat ids for `hit_chance`, `crit_chance`, and `evade_chance`.
- [x] Step 1.2: Add rare-gated roll candidates to a deterministic equipment template.
- [x] Step 1.3: Constrain new values as percent-point integers in schema/validation.
```bash
make validate-shared
```

## Task 2 — Server stat aggregation and combat

Files:
- Modify: `server/internal/game/sim.go`
- Modify/Create: focused `server/internal/game/*_test.go`

- [x] Step 2.1: Aggregate base/rolled hit, crit, and evade equipment stats.
- [x] Step 2.2: Clamp hit/crit/evade to [0, 1] after converting percent points.
- [x] Step 2.3: Resolve defender evade after attacker hit and before block using existing miss outcome.
- [x] Step 2.4: Add focused tests for hit, crit, evade, and attack-speed continuity.
```bash
cd server && go test ./internal/game/... -run 'RareCombatAffix|CombatStat|AttackSpeed'
```

## Task 3 — Bot proof

Files:
- Create: `tools/bot/scenarios/79_live_rare_combat_affixes.json`

- [x] Step 3.1: Add a compact protocol scenario that equips deterministic rolled gear.
- [x] Step 3.2: Assert stat breakdowns include rare affix roll sources.
- [x] Step 3.3: Prove player hit/crit and defender evade combat outcomes in focused Go tests; keep
  the protocol scenario focused on deterministic stat-breakdown evidence.
```bash
make bot scenario=79_live_rare_combat_affixes.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v188_live-rare-combat-affixes.md`
- Modify: `docs/specs/v188_spec-live-rare-combat-affixes.md`
- Modify: `docs/plans/v188_2026-06-15-live-rare-combat-affixes.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark spec/plan complete after verification.
- [x] Step 4.2: Add as-built summary and PROGRESS lifecycle row.
- [x] Step 4.3: Record deferred tuning and affix-name scope.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'RareCombatAffix|CombatStat|AttackSpeed'`
- [x] `make bot scenario=79_live_rare_combat_affixes.json`
- [x] `make ci`
