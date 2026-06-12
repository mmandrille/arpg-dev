# v99 Plan — Rogue Skill Mechanics

Status: Complete
Goal: Make Rogue's `Poison Stab` and `Dash` fully usable and prove them in the Rogue foundation bot scenario.
Architecture: Shared skill data owns tuning for poison percent/duration and dash range/damage/cooldown. The Go sim owns all movement, damage, DOT ticks, and events. The Godot client remains presentation-only, consuming existing or minimally extended event payloads. Python bot proof drives the same protocol path as the real client.
Tech stack: shared JSON/schema, Go authoritative sim, Python protocol bot, Godot client presentation, SDD docs.

## Baseline And Shortcut Decision

Builds on v98 Rogue class foundation plus the post-v98 fix that made `poison_stab` and `dash` visible in the skill tree.

Godot plugin adoption: reject. This slice does not add UI frameworks, art packs, camera tooling, or client authority. Any client work is small event presentation inside the existing in-repo client.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Schema for poison DOT and dash payloads |
| Modify | `shared/rules/skills.v0.json` | Rogue skill tuning |
| Modify | `server/internal/game/rules.go` | Rule structs and validation |
| Modify | `server/internal/game/handlers.go` | Dispatch Rogue-specific skill casts |
| Modify | `server/internal/game/sim.go` | Deterministic Rogue state integration |
| Create | `server/internal/game/rogue_skills.go` | Deterministic dash, poison ticks, off-hand damage helpers |
| Create | `server/internal/game/rogue_skills_test.go` | Rogue skill unit coverage |
| Modify | `tools/bot/scenarios/47_rogue_class_foundation.json` | End-to-end Rogue proof |
| Modify | `tools/bot/run.py` | Bot action/assertion support only if existing steps are insufficient |
| Modify | `tools/bot/test_protocol.py` | Bot unit tests for new scenario/assertions |
| Modify | `PROGRESS.md` | Lifecycle closeout |
| Create | `docs/as-built/v99_rogue-skill-mechanics.md` | As-built summary |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/inventory_panel.gd` had pre-existing committed drift.

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [x] `server/internal/game/rogue_skills.go` isolates Rogue dash, poison DOT, and off-hand timing
  helpers from the existing simulation file.
- [x] Documented maintenance exception: update `client/scripts/inventory_panel.gd` baseline to the
  current committed line count surfaced by `make maintainability`; this slice does not edit the
  inventory panel.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Skill Contracts

Files:
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`

- [x] Step 1.1: Add schema-backed `poison` payload for DOT percent per second, duration ticks, rank scaling, and magic duration scaling.
- [x] Step 1.2: Add schema-backed `dash` payload for range, range per rank, damage percent, magic damage scaling, and movement-through-targets behavior.
- [x] Step 1.3: Update `poison_stab` and `dash` data to use the new payloads and intended tuning.
```bash
make validate-shared
```

## Task 2 — Server Authority

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/sim.go`
- Modify/Create: `server/internal/game/*test*.go`

- [x] Step 2.1: Load and validate Rogue poison/dash rule payloads.
- [x] Step 2.2: Implement Poison Stab as immediate weapon damage plus deterministic one-second poison ticks.
- [x] Step 2.3: Implement Dash as server movement along a line, damaging crossed monsters by a rule-derived percent.
- [x] Step 2.4: Emit skill, damage, poison tick, and cooldown events through existing event/state_delta paths.
- [x] Step 2.5: Add focused Go tests for Poison Stab DOT, Dash movement/damage, and off-hand cadence.
```bash
cd server && go test ./internal/game -run 'TestRogue'
```

## Task 3 — Bot Proof

Files:
- Modify: `tools/bot/scenarios/47_rogue_class_foundation.json`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 3.1: Extend the Rogue foundation scenario to learn/cast Dash and Poison Stab.
- [x] Step 3.2: Prove Dash moves through an enemy and damages it.
- [x] Step 3.3: Prove poison damage ticks after Poison Stab.
- [x] Step 3.4: Prove the Rogue lands at least three attacks after dash/poison, with at least two main-hand and one off-hand attack events.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=rogue_class_foundation
```

## Task 4 — Client Presentation

Files:
- Modify: `client/scripts/main.gd` if new event presentation is needed
- Modify/Create: `client/tests/*.gd` if client behavior changes

- [x] Step 4.1: Reuse existing skill/damage presentation where possible; no client event mapping change was needed.
- [x] Step 4.2: Add focused client coverage only if client code changes; skipped because this slice did not change client code.
```bash
make client-unit
```

## Task 5 — Lifecycle Docs And CI

Files:
- Modify: `docs/plans/v99_2026-06-12-rogue-skill-mechanics.md`
- Modify: `docs/specs/v99_spec-rogue-skill-mechanics.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v99_rogue-skill-mechanics.md`

- [x] Step 5.1: Mark completed plan tasks.
- [x] Step 5.2: Update `PROGRESS.md` and add as-built notes.
- [x] Step 5.3: Run final gates.
```bash
make maintainability
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestRogue|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=rogue_class_foundation`
- [x] `make client-unit` if client code changes; skipped because no client code changed
- [x] `make ci`

## Deferred Scope

- Rich dash/poison visual effects beyond existing damage/skill event feedback.
- Further Rogue active skills or skill-tree branches.
- Global dual-wield rebalance beyond proving Rogue off-hand attacks in the scenario.
