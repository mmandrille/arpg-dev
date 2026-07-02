# v409 Spec — Town Training Doll

Status: Approved for implementation
Date: 2026-07-02
Codename: `town-training-doll`
Baseline: v408 `weapon-elemental-procs` complete

## Purpose

Add a **Training Doll** to town (`level 0`): a human-silhouette target that accepts basic attacks and
skills. At session creation the doll mirrors the host player's **defensive** profile (`max_hp`,
`armor`, `block_percent`, resistances). Hits use the authoritative combat pipeline, show floating
damage numbers, and open a scrollable **Combat Damage Log** panel listing each attack with
server-authored formula breakdown lines. When HP reaches 0 the doll plays a death reaction, then
revives at full HP after **3 seconds** (data-driven). No XP, loot, or retaliation.

## Non-goals

- Live re-sync when the player reallocates stats or re-equips mid-session
- General town PvP or combat against non-doll entities
- Production imported character art (code-native silhouette first)
- Draggable window / layout persistence for the log panel
- Per-player dolls or per-player filtered logs in co-op (shared doll; rows tagged by `source_entity_id`)
- Changing lab `training_dummy_*` scenarios

## Acceptance criteria

- [ ] `dungeon_levels` town (`level 0`) contains one `town_training_doll` entity visible in `make play`
- [ ] Doll renders as a human silhouette (code-native mesh; adopt in-repo procedural art)
- [ ] Doll defensive stats mirror the host player's effective values at session creation (rule-derived tests, not hardcoded pins)
- [ ] Basic attacks and at least one active skill can target the doll in town
- [ ] Floating combat text appears on hits (existing presentation path)
- [ ] Combat events against the doll include `damage_breakdown` lines (hit → roll → armor → resistance → final)
- [ ] First hit opens the Combat Damage Log; subsequent hits append scrollable rows; top-right **X** closes; next hit re-opens
- [ ] Doll death emits `monster_killed` with no XP/loot; after 3s revive at full HP on spawn tile; not targetable while down
- [ ] Town combat remains impossible against non-training entities (only doll placed in town)
- [ ] Extended bot `118_town_training_doll` and client `91_training_damage_log` pass

## Scope and files

| Area | Paths |
|------|-------|
| Rules | `shared/rules/monsters.v0.json`, `monsters.v0.schema.json`, `worlds.v0.json`, `shared/assets/monster_visuals.v0.json` |
| Protocol | `shared/protocol/state_delta.v8.schema.json` (additive `damage_breakdown`) |
| Server | `server/internal/game/training_doll.go`, `combat_breakdown.go`, `sim.go`, `types.go`, `rules.go`, `damage_types.go` |
| Bot | `tools/bot/scenarios/118_town_training_doll.json`, `runtime_assertions` / `combat_event_matches` |
| Client | `training_damage_log_panel.gd`, `training_doll_visual.gd`, `combat_breakdown_format.gd`, `main.gd` |
| Docs | `PROGRESS.md`, `docs/as-built/v409_town-training-doll.md` |

## Test and bot proof

- Go: `go test ./internal/game/... -run TrainingDoll`
- Protocol: `make bot scenario=118_town_training_doll`
- Client: `make bot-client SCENARIO=91_training_damage_log HEADLESS=1`
- Visual: `make bot-visual scenario=118_town_training_doll`

## Asset decision

**Adopt** in-repo procedural human silhouette (`CapsuleMesh` stack, unshaded neutral tone) — same
pattern as `town_ambient_life.gd`. **Reject** external GLB and `monster_dummy` crate visual for this def.

## Resolved questions

| # | Decision |
|---|----------|
| Q-1 | Mirror defensive stats only at session create |
| Q-2 | Death presentation + 3s full-HP revive (`revive_delay_ticks: 30` at 10 Hz) |
| Q-3 | Re-open log panel on next hit after close |
| Q-4 | One shared town doll; log rows use `source_entity_id` |
| Q-5 | Full pipeline breakdown lines on combat events |
| Q-6 | Fixed right-side panel (~360px), not draggable |
| Q-7 | Host snapshot at session create for shared doll stats |

## Risks

- Protocol additive field `damage_breakdown` on combat events — backward-compatible on v8 schema
- `main.gd` touch — extract panel wiring to keep coordinator stable
- Determinism: revive scheduling uses tick count only
