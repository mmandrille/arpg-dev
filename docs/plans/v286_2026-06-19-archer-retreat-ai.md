# v286 Plan — Archer Retreat AI

Status: Implemented
Goal: Give `dungeon_archer` a data-driven preferred minimum range and deterministic retreat movement
when the player rushes inside that range.
Architecture: Extend monster rules with ranged-only `preferred_min_range`. Keep retreat search in
`monster_ranged_positioning.go`; `sim.go` should only ask for a retreat goal before short-circuiting
because a ranged monster is already in attack range.
Tech stack: Go sim/rules, shared JSON/schema validation, protocol bot scenario.

## Baseline and shortcut decision

Builds on v52 ranged monster AI and v273 blocked-shot repositioning. The existing archer attack
range, projectile speed, hit chance, cooldown, spawn guarantees, and client bow marker stay as-is.

Asset/plugin decision:

- Adopt: current server-authoritative archer movement and projectile path.
- Borrow: existing ranged clear-shot, pathfinding, navigation budget, and bot distance assertion
  patterns.
- Reject: external assets, plugins, cover seeking, predictive leading, and client-only movement
  cheats.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.json` | Add `preferred_min_range` to `dungeon_archer`. |
| Modify | `shared/rules/monsters.v0.schema.json` | Allow the new ranged movement field. |
| Modify | `server/internal/game/rules.go` | Load and validate the field. |
| Modify | `server/internal/game/monster_ranged_positioning.go` | Search deterministic retreat goals. |
| Modify | `server/internal/game/sim.go` | Ask ranged monsters to retreat before accepting in-range standoff. |
| Modify | `shared/rules/worlds.v0.json` | Add a compact archer-retreat lab. |
| Modify | `server/internal/game/ranged_monster_positioning_test.go` | Add close-start retreat coverage. |
| Create | `tools/bot/scenarios/94_archer_retreat_ai.json` | Protocol bot proof for the authored lab. |
| Create during finish | `docs/as-built/v286_archer-retreat-ai.md` | Record proof and commands. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `server/internal/game/sim.go` — add only the retreat hook call.
- [x] `server/internal/game/rules.go` — add only the new field and validation.

Decision:

- [x] Extract helper/module: keep retreat search in existing focused `monster_ranged_positioning.go`.
- [x] Defer extraction with rationale: no broader AI coordinator extraction is needed for this slice.

Verification:

```bash
make maintainability
```

## Task 1 — Data contract

Files:

- Modify: `server/internal/game/rules.go`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`

- [x] Step 1.1: Add `PreferredMinRange float64` / `preferred_min_range`.
- [x] Step 1.2: Validate it is only valid on ranged monsters.
- [x] Step 1.3: Validate it is positive and below `attack_range`.
- [x] Step 1.4: Set `dungeon_archer.preferred_min_range` from rules.

Verify:

```bash
make validate-shared
```

## Task 2 — Retreat movement

Files:

- Modify: `server/internal/game/monster_ranged_positioning.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Add a retreat-goal search that runs only when the player is closer than
  `preferred_min_range`.
- [x] Step 2.2: Candidate goals must be reachable, unblocked, farther from the player than the
  current monster position, within ranged attack reach, and clear-shot.
- [x] Step 2.3: Cache the selected path using existing monster navigation cache helpers.
- [x] Step 2.4: Keep existing blocked-shot closer reposition fallback unchanged when no retreat goal
  applies.

Verify:

```bash
(cd server && go test ./internal/game -run 'TestRangedMonster' -count=1)
```

## Task 3 — Focused proof

Files:

- Modify: `shared/rules/worlds.v0.json`
- Modify: `server/internal/game/ranged_monster_positioning_test.go`
- Create: `tools/bot/scenarios/94_archer_retreat_ai.json`

- [x] Step 3.1: Add a close-start archer lab with enough open space for a valid backpedal.
- [x] Step 3.2: Add a Go test proving the archer starts inside preferred range and moves back out
  toward a valid ranged standoff.
- [x] Step 3.3: Keep the existing blocked-shot Go test green.
- [x] Step 3.4: Add a protocol bot scenario that observes the archer move from its close-start lab.

Verify:

```bash
(cd server && go test ./internal/game -run 'TestRangedMonster' -count=1)
make bot scenario=archer_retreat_ai
```

## Task 4 — Docs and lifecycle

Files:

- Existing: `docs/specs/v286_spec-archer-retreat-ai.md`
- Existing: `docs/plans/v286_2026-06-19-archer-retreat-ai.md`
- Create during finish: `docs/as-built/v286_archer-retreat-ai.md`
- Modify during finish: `PROGRESS.md`

- [x] Step 4.1: Record focused checks and bot proof in the as-built note.
- [x] Step 4.2: Update lifecycle/current status during finish.

## Task 5 — Final verification

- [x] `make validate-shared`
- [x] `(cd server && go test ./internal/game -run 'TestRangedMonster' -count=1)`
- [x] `make bot scenario=archer_retreat_ai`
- [x] `make maintainability`
