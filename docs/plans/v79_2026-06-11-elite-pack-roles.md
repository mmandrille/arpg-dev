# v79 Plan — Elite Pack Roles

Status: Complete
Goal: Add deterministic pack leader and role-composition foundations without changing client protocol.
Architecture: Shared rules describe monster roles and pack composition constraints. Dungeon generation assigns planned pack members using those constraints and keeps leader state internal to generated monsters for tests/future slices.

## Task 1 — Shared Role Rules

- [x] Add generated monster `pack_role` values.
- [x] Add dungeon pack composition and elite chance tuning.
- [x] Validate role references and composition bounds.

## Task 2 — Role-Aware Pack Generation

- [x] Build pack definitions per pack size rather than one global random list.
- [x] Ensure each pack has a frontline member and respects max ranged members.
- [x] Mark at most one planned member per elite pack as internal leader.
- [x] Bias leader rarity toward champion while preserving normal rarity rolls for the rest.

## Task 3 — Proof

- [x] Add server tests for role composition, leader bounds, and determinism.
- [x] Run the existing pack aggro protocol scenario.
- [x] Update lifecycle docs and `PROGRESS.md`.

## Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestDungeonMonsterGeneration|TestDungeonMonsterGenerationCreatesDeterministicPacks|TestDungeonMonsterGenerationCanForceElitePackLeaders'`
- [x] `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`
- [x] `make ci`
