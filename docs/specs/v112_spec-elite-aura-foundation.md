# v112 Spec: Elite Aura Foundation

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `elite-aura-foundation`

## Purpose

Give generated elite packs a small server-authoritative command aura foundation. When a pack leader
is alive and nearby, same-pack followers deal increased outgoing monster damage. The aura tuning
lives in shared dungeon-generation rules, so later slices can add more aura types, presentation, or
protocol labels without hardcoding balance into Go.

## Non-goals

- No client VFX, monster nameplates, aura icons, aura labels, or protocol schema changes.
- No multiple aura rolls, random aura type selection, resist aura, speed aura, healing aura, or
  debuff aura.
- No static/lab monster aura behavior; this applies only to generated dungeon packs that already
  have `pack_id` / leader metadata.
- No item, loot, XP, or monster-rarity formula changes.

## Acceptance criteria

- `shared/rules/dungeon_generation.v0.json` defines an `elite_aura` object under
  `monster_placement` with id, radius, and damage bonus percent.
- The dungeon-generation schema and Go rule loader validate the aura config.
- Generated monster `pack_id` and `packLeader` metadata is preserved on live monster entities.
- A same-pack non-leader monster receives the aura damage bonus only when its pack leader is alive
  and within the configured radius.
- The aura does not affect pack leaders, monsters from other packs, monsters without pack metadata,
  monsters whose leader is dead, or monsters outside radius.
- Focused Go tests cover aura application/rejection cases and pack metadata transfer into live
  generated monsters.

## Scope and likely files

- Rules: `shared/rules/dungeon_generation.v0.json`,
  `shared/rules/dungeon_generation.v0.schema.json`, `server/internal/game/rules.go`
- Sim: `server/internal/game/sim.go`, new focused aura helper/test file under
  `server/internal/game/`
- Lifecycle docs: `PROGRESS.md`, `docs/as-built/v112_elite-aura-foundation.md`

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestEliteAura|TestGeneratedDungeonMonstersPreservePackMetadata' -count=1`
- `make maintainability`
- `make test-go`
- `make ci`

Protocol bot proof is deferred for this foundation because no realtime protocol or client-visible
aura state changes in v112. The deterministic combat contract is owned by focused Go tests now; a
future presentation/protocol slice should add a bot scenario once aura state or visuals are exposed
through snapshots/events.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Does the leader buff itself? | No. Leaders already receive champion/rarity scaling; v112 buffs followers only. |
| Q-2 | How many aura types are added? | One data-backed `elite_command` damage aura to keep the slice small. |
| R-1 | Aura implementation can accidentally affect static monsters. | Require pack metadata and a living same-pack leader. |
| R-2 | Combat tuning can become hardcoded. | Radius and damage bonus live in shared rules and schema validation. |

## ADR alignment

- ADR-0001: keeps combat outcomes server-authoritative and deterministic.
- v79 elite-pack roles: builds on generated pack leader metadata with actual live combat impact.
