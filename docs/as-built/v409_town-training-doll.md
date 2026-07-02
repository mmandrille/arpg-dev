# v409 As-built — Town Training Doll

**Slice:** v409 `town-training-doll`  
**Date:** 2026-07-02

## What it proves

- Town level 0 and `town_training_doll_lab` spawn a shared `town_training_doll` training target with procedural silhouette presentation.
- At session create the doll mirrors the host's defensive stats (`max_hp`, armor, block); it takes basic attacks/skills like a monster but grants no loot/XP.
- On 0 HP the doll plays `monster_killed` presentation, stays untargetable, and revives at spawn with full HP after `revive_delay_ticks` (30 ticks ≈ 3s) via `training_doll_revived`.
- Authoritative combat events against the doll include `damage_breakdown` formula lines; the client opens a scrollable combat damage log panel on hits and closes with X.
- Extended protocol bot `118_town_training_doll` and client bot `91_training_damage_log` cover the contracts.

## Key decisions

- Reused monster combat pipeline with `training_target` + `revive_delay_ticks` monster def flags.
- Protocol stayed on v8 with additive `damage_breakdown` / `training_doll_revived` events.
- Adopted in-repo silhouette (`training_doll_visual.gd`); rejected `monster_dummy` GLB for this target.
- Client wiring extracted to `training_damage_log_bridge.gd` and `bot_training_damage_log_actions.gd` for maintainability.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/... -run TrainingDoll`
- `make bot scenario=118_town_training_doll`
- `make bot-client SCENARIO=91_training_damage_log HEADLESS=1`
- `make maintainability`

## Visual check

- `make bot-visual scenario=118_town_training_doll`

## Deferred

- Co-op per-peer log filtering beyond `source_entity_id` rows.
- Miss/block log rows with breakdown (hits/skills own breakdown today).
- Attack_missed / attack_blocked panel open behavior.
