# v298 As-Built: Hit Stop / Impact Flash

Date: 2026-06-19
Spec: [`docs/specs/v298_spec-hit-stop-impact-flash.md`](../specs/v298_spec-hit-stop-impact-flash.md)
Plan: [`docs/plans/v298_2026-06-19-hit-stop-impact-flash.md`](../plans/v298_2026-06-19-hit-stop-impact-flash.md)

## What Shipped

- Extended `ModelReactionController` so hit and death reactions immediately brighten the affected
  model, hold briefly, then continue through the existing lean/restore or death lean.
- Added `impact_feedback_count` to reaction debug state so bot scenarios can assert that the client
  presented an impact without brittle frame-perfect color checks.
- Extended `wait_entity_reaction` / `assert_entity_reaction` matching with `impact_feedback_min`.
- Added `79_hit_stop_impact_flash`, which equips the control-lab training bow, retries until an
  authoritative `monster_damaged` event, then waits for the monster reaction state to report hit
  impact feedback.
- Preserved server authority and existing reaction behavior: no protocol, sim timing, damage,
  health, loot, or global scene pause changes.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_animation.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=79_hit_stop_impact_flash HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. `test_animation.gd` still prints the existing Godot
ObjectDB/resource warnings after its PASS line but exits 0. The local bot script printed the same
post-pass `cleanup_account.py` missing-`httpx` warning in this environment, but every scenario
returned success.

## Manual Visual Command

```bash
make bot-visual scenario=79_hit_stop_impact_flash
```

## Deferred

- Movement acceleration smoothing remains the next selected Movement / Combat Fluidity slice.
- Command retarget grace and melee lunge remain later selected slices in this feature queue.
- Camera shake, global hit stop, production VFX assets, and data-driven VFX catalogs remain out of
  scope.
