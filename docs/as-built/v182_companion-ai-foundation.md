# v182 As-Built — Companion AI foundation

Date: 2026-06-15

## What shipped

- Added a server-owned `companion` entity type to v8 snapshot and delta schemas.
- Added preset support for lab-authored companions that reference existing monster definitions while carrying `owner_id`, HP, target, movement, and melee attack behavior as server state.
- Added focused companion AI in `server/internal/game/companion_ai.go`: owner follow, hostile monster target acquisition, owner-ignoring movement collision, melee damage, and normal combat events sourced from the companion.
- Added `companion_ai_lab` and protocol bot scenario `73_companion_ai_foundation.json`, proving a test companion appears, follows after owner movement, and damages a lab monster.

## Key decisions

- The v182 lab companion uses `combat_lab_crit_attacker` for deterministic damage proof. Ranger wolf visuals and skill behavior are deferred to v183.
- Companion tuning remains intentionally small and local to the foundation proof. v185 owns data-driven rank scaling, limits, HP, damage, and attack stat rules.
- The existing bot assertion system could prove companion identity, movement, and companion-sourced combat events without new bot runner actions.
- No Godot plugin was adopted because this slice is authoritative server simulation; the existing renderer consumes server entity views.

## Verification

```bash
make maintainability
make validate-shared
cd server && go test ./internal/game/... -run Companion -count=1
make bot scenario=73_companion_ai_foundation.json
make ci
```

Manual visual proof:

```bash
make bot-visual scenario=73_companion_ai_foundation.json
```

## Deferred

- v183 Ranger black wolf summon skill and black wolf presentation.
- v184 Sorcerer revive skill and corpse/dead-monster targeting.
- v185 data-driven companion rank scaling, stats, and limits.
- v186 elite minions reusing companion-follow/assist behavior.
- Mercenary hiring, persistence, equipment snapshots, commands, UI, XP/loot/potion behavior, and session persistence.
