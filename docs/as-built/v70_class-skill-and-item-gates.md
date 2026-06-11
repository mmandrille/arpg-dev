# v70 As-Built — Class Skill and Item Gates

Date: 2026-06-11

## What Shipped

- Skill rules now map each active skill to one class: barbarian uses `rage`, sorcerer uses `magic_bolt`, and paladin uses `heal`.
- The sim rejects cross-class skill point spends and cross-class casts with `skill_class_not_allowed` without mutating ranks, cooldowns, or resources.
- Fixed class weapons now exist for each class: `barbarian_axe`, `sorcerer_staff`, and `paladin_mace`; they are slightly stronger than `rusty_sword` and declare `class_required`.
- Equipment handling rejects wrong-class fixed weapons with `class_requirement_not_met`.
- Session-start progression snapshots carry `character_class`, and realtime session construction reconciles live sim state with the selected immutable character class.
- Protocol bot scenarios now create class-specific characters for class-owned skill proofs.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game ./internal/store ./internal/http ./internal/realtime ./internal/replay`
- `make test-py`
- `make bot`
- `make ci`

## Deferred

- Create-character class picker UI, class sprites, and character-list class icon presentation remain v71 scope.
- Procedural class-locked rolled items remain deferred; v70 only adds fixed class weapons.
