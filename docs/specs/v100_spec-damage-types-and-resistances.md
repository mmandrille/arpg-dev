# v100 Spec - Damage Types and Resistances

Status: Complete
Date: 2026-06-12
Codename: `damage-types-and-resistances`

## Purpose

Add server-authoritative damage types and monster resistances so skills and weapons can carry an elemental/combat identity without hardcoded behavior. Existing untyped damage falls back to `force`. The first gameplay proof covers:

- `magic_bolt` as `force`.
- `ice_shard` as `cold`.
- `poison_stab` as `poison`.
- the existing `ligthing` skill id as canonical damage type `lightning`.
- flying enemies taking 50% less lightning damage.
- quadruped enemies taking 50% more lightning damage.

## Non-goals

- Do not rename the existing `ligthing` skill id, presentation key, or projectile ids in this slice.
- Do not add undead or poison immunity; that is the next selected slice.
- Do not add client VFX, UI filters, or type-specific damage number colors beyond tolerating/passing through event data.
- Do not rebalance final combat numbers beyond the explicit resistance examples.
- Do not introduce new damage types beyond `force`, `cold`, `poison`, and `lightning`.

## Acceptance Criteria

- Shared schemas allow `damage_type` on skill damage and weapon damage, and `resistances` on monsters.
- Missing `damage_type` resolves to `force` for existing skills/weapons.
- Shared validation rejects unknown damage types and out-of-range resistance values.
- Server combat applies monster resistance after armor mitigation and before minimum-damage clamping.
- Positive resistance reduces damage; negative resistance increases damage.
- `dungeon_bat` has `lightning: 0.5`; `dungeon_wolf` has `lightning: -0.5`.
- Authoritative combat events include `damage_type` for monster damage caused by basic attacks, projectile skills, chain hits, cold shards, poison ticks, and Rogue cone/dash damage where applicable.
- Focused Go tests prove `lightning` damage against a resistant flying target is lower than the same hit against a neutral target, and higher against a weak quadruped target.
- A protocol bot scenario proves the resistance contract through authoritative events without relying on client presentation.
- `make validate-shared`, focused Go tests, `make bot`, `make maintainability`, and `make ci` pass.

## Scope and Likely Files

- Shared rules:
  - `shared/rules/skills.v0.json`
  - `shared/rules/skills.v0.schema.json`
  - `shared/rules/items.v0.json`
  - `shared/rules/items.v0.schema.json`
  - `shared/rules/monsters.v0.json`
  - `shared/rules/monsters.v0.schema.json`
  - `shared/rules/worlds.v0.json`
- Protocol:
  - `shared/protocol/session_snapshot.v8.schema.json`
  - `server/internal/game/types.go`
- Server:
  - focused combat helper/test files under `server/internal/game/`
  - `server/internal/game/rules.go`
  - small call-site updates in `sim.go`, `handlers.go`, and Rogue skill helpers as needed
- Bot:
  - new focused scenario under `tools/bot/scenarios/`
  - `tools/bot/run.py` only if existing assertions cannot validate `damage_type` and relative damage
- Docs:
  - `docs/plans/v100_2026-06-12-damage-types-and-resistances.md`
  - `docs/as-built/v100_damage-types-and-resistances.md`
  - `PROGRESS.md`

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestDamageType|TestResistance'`
- `make bot scenario=damage_types_and_resistances` or the scenario name produced by the plan
- `make maintainability`
- `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: `sim.go` is over the maintainability baseline; the plan must either keep edits very small and offset growth through a focused helper, or document a ratchet decision.
- Risk: `tools/bot/run.py` is over the maintainability baseline; prefer existing assertion helpers or minimal changes.
- Risk: event schema changes must stay compatible with clients that ignore unknown fields.
