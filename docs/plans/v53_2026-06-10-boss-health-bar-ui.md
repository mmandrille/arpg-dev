# v53 Plan: Boss Health Bar UI

Spec: `docs/specs/v53_spec-boss-health-bar-ui.md`
Date: 2026-06-10

## Adoption Checklist

- Adopt: existing Godot `Control`/`CanvasLayer` UI patterns and the current client entity debug-state pattern.
- Borrow: sizing/color language from `monster_health_bar.gd` and the existing HUD palette.

## Tasks

1. Add `client/scripts/boss_health_bar.gd`.
   - Build a stable top-center `Control` with title, hp text, and fill bar.
   - Provide `show_boss(...)`, `hide_boss()`, and `get_debug_state()`.
   - Clamp hp/max hp and ratio defensively.

2. Wire the boss bar in `client/scripts/main.gd`.
   - Preload and store a `BossHealthBar`.
   - Add it to the existing HUD scene graph during `_build_scene()`.
   - Scan current live boss monster records after entity updates/removals and level clears.
   - Hide on level change, teardown, entity remove, and boss death.
   - Add `boss_health_bar` to `get_bot_state()`.

3. Extend `client/scripts/bot_scenario_runner.gd`.
   - Add `wait_boss_health_bar` and `assert_boss_health_bar`.
   - Match optional expectations for visible, boss id, template id, title, hp/max hp, and ratio ranges.
   - Validate that boss bar steps include at least one useful expectation and `timeout_s` for waits.

4. Add/adjust client unit tests.
   - Cover boss bar display/update/hide behavior.
   - Cover bot-runner validation and matching for the new step type.

5. Add the client scenario `tools/bot/scenarios/client/26_boss_health_bar_ui.json`.
   - Start in `dungeon_levels`.
   - Descend to level `-5`.
   - Wait for the boss health bar and assert `boss_template_id == "cave_warden"`.

6. Verification and finish.
   - Run `CLIENT_UNIT_ONLY=1 make client-smoke`.
   - Run the new bot client scenario headless.
   - Run `make ci`.
   - Update `PROGRESS.md`.
   - Commit as `feat: v53: boss health bar UI`.
