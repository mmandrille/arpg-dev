# v182 Plan — Companion AI Foundation

Status: Complete
Goal: Add a server-owned companion entity foundation with follow, targeting, melee attack, and bot proof.
Architecture: Companions are authoritative sim entities, not client pets. The server emits `companion` entity views over the existing snapshot/delta path, owns target acquisition and attacks, and reuses deterministic sorted entity ordering. Class skills, revive, rank scaling, and elite minion reuse remain later slices.
Tech stack: Go sim, shared protocol schemas, shared world JSON, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v181 set item foundation and ADR-0010's server-authoritative companion direction. Godot plugin adoption check read on 2026-06-15: reject for v182 because this is server simulation behavior and existing rendering can consume protocol entities; no external client plugin should own gameplay AI.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/sim.go` | Companion kind, AI loop, target selection, melee damage, entity view |
| Modify | `server/internal/game/types.go` | Protocol view comments/fields if needed |
| Create | `server/internal/game/companion_ai_test.go` | Focused server companion tests |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Allow companion entity type |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow companion entity type |
| Modify | `shared/rules/worlds.v0.json` | Add compact companion AI lab |
| Modify | `tools/bot/run.py` | Add minimal scenario actions/assertions if existing ones cannot prove companion behavior |
| Modify | `tools/bot/runtime_assertions.py` | Add matching runtime assertion if needed |
| Modify | `tools/bot/test_protocol.py` | Bot helper unit coverage if needed |
| Create | `tools/bot/scenarios/73_companion_ai_foundation.json` | Protocol bot proof |
| Modify | `docs/specs/v182_spec-companion-ai-foundation.md` | Mark final status |
| Create | `docs/as-built/v182_companion-ai-foundation.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle closeout |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `client/scripts/stash_panel.gd`, `server/internal/game/game_test.go`, `server/internal/game/types.go`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: companion helpers live in `server/internal/game/companion_ai.go`; no bot runner changes were needed.

Verification:
```bash
make maintainability
```

## Task 1 — Shared protocol and lab data

Files:
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/rules/worlds.v0.json`

- [x] Step 1.1: Allow `companion` as an entity `type` in snapshot and delta schemas.
- [x] Step 1.2: Add a compact lab world with a player, one companion, one soft monster target, and simple bounds.
```bash
make validate-shared
```

## Task 2 — Server companion foundation

Files:
- Modify: `server/internal/game/sim.go`
- Create: `server/internal/game/companion_ai_test.go`

- [x] Step 2.1: Add a companion entity kind and lab world loading support for companion entities with owner, monster definition, HP, damage, cooldown, and spawn position.
- [x] Step 2.2: Add deterministic companion follow movement toward an owner-adjacent follow slot.
- [x] Step 2.3: Add deterministic companion target acquisition and melee attack against living monsters.
- [x] Step 2.4: Emit entity updates and combat events for companion movement/damage.
- [x] Step 2.5: Cover identity, follow, and attack behavior with focused Go tests.
```bash
cd server && go test ./internal/game/... -run Companion
```

## Task 3 — Bot proof

Files:
- Create: `tools/bot/scenarios/73_companion_ai_foundation.json`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/runtime_assertions.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 3.1: Add minimal companion-aware scenario assertions/actions only if existing entity and combat assertions cannot prove the slice.
- [x] Step 3.2: Add protocol bot scenario proving the companion exists, follows after owner movement, and damages the lab monster.
- [x] Step 3.3: Add bot unit coverage for any new assertion helpers.
```bash
make bot scenario=73_companion_ai_foundation.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Modify: `docs/specs/v182_spec-companion-ai-foundation.md`
- Modify: `docs/plans/v182_2026-06-15-companion-ai-foundation.md`
- Create: `docs/as-built/v182_companion-ai-foundation.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec and plan complete.
- [x] Step 4.2: Add v182 lifecycle/as-built updates and defer Ranger, revive, rank scaling, elite minion reuse, persistence, UI, and mercenaries.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run Companion`
- [x] `make bot scenario=73_companion_ai_foundation.json`
- [x] `make ci`

## Deferred scope

- v183 Ranger black wolf skill.
- v184 Sorcerer revive skill.
- v185 data-driven companion rank scaling and limits.
- v186 elite minions using companion-follow/assist behavior.
- Mercenary hiring, persistence, gear snapshots, commands, UI, XP/loot/potion behavior.
