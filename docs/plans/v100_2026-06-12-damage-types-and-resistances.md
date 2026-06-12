# v100 Plan - Damage Types and Resistances

Status: Ready for implementation
Goal: Add server-authoritative damage types and monster resistances with a bot-proven lightning resistance/weakness contract.
Architecture: Damage type is shared data on skills and weapons, defaulting to `force` when omitted. Monster resistances live in monster rules as fractions where positive values reduce damage and negative values increase it. The Go sim applies resistance after armor mitigation and before minimum-damage clamping, then emits `damage_type` on authoritative combat events. The client remains presentation-only and ignores the new field unless future slices render it.
Tech stack: Shared JSON/schema, Go sim/rules/tests, protocol event schema, Python bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v99 `rogue-skill-mechanics` plus the v100 engineering review gate. No client UI, camera, inventory presentation, or art is in scope, so the Godot plugin adoption checklist is not required for this slice.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add explicit skill damage types for Magic Bolt, Ice Shard, Ligthing, and Poison Stab. |
| Modify | `shared/rules/skills.v0.schema.json` | Allow canonical skill damage types. |
| Modify | `shared/rules/items.v0.json` | Add optional/fallback weapon damage type where needed. |
| Modify | `shared/rules/items.v0.schema.json` | Allow `damage_type` on weapon damage ranges. |
| Modify | `shared/rules/monsters.v0.json` | Add lightning resistance/weakness on bat and wolf, plus lab targets. |
| Modify | `shared/rules/monsters.v0.schema.json` | Allow resistance maps. |
| Modify | `shared/rules/worlds.v0.json` | Add compact bot lab world. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Add optional `damage_type` to combat events. |
| Modify | `server/internal/game/types.go` | Add event and rule fields for damage type/resistances. |
| Modify/Create | `server/internal/game/*damage*.go`, `server/internal/game/*damage*_test.go` | Apply and test resistance math without growing monoliths unnecessarily. |
| Modify | `server/internal/game/rules.go` | Validate damage types and resistance ranges. |
| Modify | `server/internal/game/sim.go`, `server/internal/game/handlers.go`, `server/internal/game/rogue_skills.go` | Thread damage type into combat call sites and events. |
| Add | `tools/bot/scenarios/48_damage_types_and_resistances.json` | End-to-end bot proof. |
| Modify | `tools/bot/run.py` | Only if existing assertions cannot match damage type/relative damage. |
| Modify | `docs/specs/v100_spec-damage-types-and-resistances.md` | Mark complete at close-out. |
| Add | `docs/as-built/v100_damage-types-and-resistances.md` | Summarize shipped behavior. |
| Modify | `PROGRESS.md` | Lifecycle row, status, and deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` (not touched)
- [x] `server/internal/game/game_test.go` (not touched)
- [x] `tools/bot/run.py` (minimal matcher-only change)
- [x] `tools/validate_shared.py` (not touched)
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `server/internal/game/sim.go`, `server/internal/game/rules.go`, `server/internal/game/types.go`, `server/internal/game/handlers.go`

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: over-limit edits were kept to field/call-site plumbing while new damage math/tests live in focused files.

Verification:
```bash
make maintainability
```

## Task 1 - Shared contract and rules

Files:
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/items.v0.json`
- Modify: `shared/rules/items.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`

- [x] Add damage-type enum support: `force`, `cold`, `poison`, `lightning`.
- [x] Add `damage_type` to relevant skill damage and weapon damage schema/rules.
- [x] Add monster `resistances` schema/rules with valid range `-1.0` to `1.0`.
- [x] Add `damage_type` to combat event schema.
```bash
make validate-shared
```

## Task 2 - Server rules and combat math

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/rules.go`
- Create: `server/internal/game/damage_types.go`
- Create: `server/internal/game/damage_types_test.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/rogue_skills.go`

- [x] Add typed fields and validation for canonical damage types and resistances.
- [x] Add focused helper for defaulting damage type and applying monster resistance.
- [x] Thread damage type through basic attacks, projectile skills, chain hits, cold shards, poison ticks, Rogue cone, and dash damage.
- [x] Emit `damage_type` on combat events.
- [x] Add focused Go tests for neutral/resistant/weak targets and fallback `force`.
```bash
cd server && go test ./internal/game/... -run 'TestDamageType|TestResistance'
```

## Task 3 - Bot scenario

Files:
- Modify: `shared/rules/worlds.v0.json`
- Add: `tools/bot/scenarios/48_damage_types_and_resistances.json`
- Modify: `tools/bot/run.py` only if needed

- [x] Add a compact lab world with neutral, flying/resistant, and quadruped/weak targets.
- [x] Add bot steps proving lightning events include `damage_type: lightning` and the weak target receives higher damage than the resistant target.
- [x] Reuse existing assertions where possible; add minimal assertion support only if necessary.
```bash
make bot scenario=damage_types_and_resistances
```

## Task 4 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v100_spec-damage-types-and-resistances.md`
- Add: `docs/as-built/v100_damage-types-and-resistances.md`
- Modify: `PROGRESS.md`

- [x] Mark the spec complete.
- [x] Add as-built notes with rule semantics, bot proof, and deferred undead scope.
- [x] Update `PROGRESS.md` latest completed slice, lifecycle table, CI gate, and open gaps/deferred scope.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestDamageType|TestResistance'`
- [x] `make bot scenario=damage_types_and_resistances`
- [x] `make ci`
