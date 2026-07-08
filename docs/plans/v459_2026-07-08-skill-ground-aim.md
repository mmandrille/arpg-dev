# v459 Plan — Skill Ground Aim

Status: Complete  
Goal: Directional skills always cast toward click/crosshair aim; server rejects homing `target_id` on `direction` skills.  
Architecture: Add `direction` targeting enum in shared rules; migrate active skills except `revive`; extract client payload builder to `skill_aim_input.gd`; server `skillCastDirectionWithRange` rejects `target_id` when targeting is `direction`; bot resolves `monster_def_id` to direction for direction skills.

## Baseline and shortcut decision

- Reuses v296 LMB sticky attack boundary and existing `cast_skill_intent` wire shape.
- Adopt in-repo aim/camera helpers; reject external plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/adr/0017-skill-aim-and-cast-resolution.md` | Aim rule + future travel/latency pointers |
| Modify | `shared/rules/skills.v0.schema.json` | `direction` enum + kind validation |
| Modify | `shared/rules/skills.v0.json` | Migrate targeting to `direction` (keep `revive`) |
| Modify | `server/internal/game/skill_rules_validation.go` | Accept `direction` for active skill kinds |
| Modify | `server/internal/game/sim.go` | Reject `target_id` on `direction` skills |
| Create | `server/internal/game/skill_aim_test.go` | Direction-only cast tests |
| Create | `client/scripts/skill_aim_input.gd` | Payload builder extracted from `main.gd` |
| Modify | `client/scripts/main.gd` | Wire RMB/hotbar to ground aim |
| Modify | `client/scripts/enemy_health_bar_visibility.gd` | Skip skill highlight for direction casts |
| Modify | `client/tests/test_coop_client.gd` | Ground-aim payload tests |
| Modify | `tools/bot/run.py` | `monster_def_id` → direction for `direction` skills |
| Create | `tools/bot/scenarios/skill_ground_aim_lab.json` | Protocol proof |
| Modify | `PROGRESS.md`, lifecycle, as-built | Close-out |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot files touched:
- [x] `client/scripts/main.gd` — extract `skill_aim_input.gd` (touch-to-shrink)
- [x] Other touched files under 600 lines

Decision:
- [x] Extract `skill_aim_input.gd` from `main.gd`

Verification:

```bash
make maintainability
```

## Task 1 — Shared contracts and ADR

- [x] Step 1.1: Add ADR-0017 and `direction` targeting enum; migrate `skills.v0.json`
- [x] Step 1.2: `make validate-shared`

## Task 2 — Server enforcement

- [x] Step 2.1: Update skill rules validation for `direction`
- [x] Step 2.2: Reject `target_id` in `skillCastDirectionWithRange` when targeting is `direction`
- [x] Step 2.3: Add skill aim tests in `skill_affixes_test.go`

```bash
cd server && go test ./internal/game/... -run SkillCast -count=1
```

## Task 3 — Client input

- [x] Step 3.1: Create `skill_aim_input.gd` and wire `main.gd` RMB + hotbar
- [x] Step 3.2: Update health-bar visibility for direction casts
- [x] Step 3.3: Extend `test_coop_client.gd`

```bash
make client-unit
```

## Task 4 — Bot proof

- [x] Step 4.1: Update `cast_skill` helper (extracted to `skill_cast_runtime.py`)
- [x] Step 4.2: Add `skill_ground_aim_lab.json` (`ci_tier: extended`)

```bash
make bot scenario=skill_ground_aim_lab
```

## Task 5 — Lifecycle docs

- [x] Update `PROGRESS.md`, lifecycle, as-built

## Final verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run SkillCast -count=1
make bot scenario=skill_ground_aim_lab
make client-unit
make maintainability
```
