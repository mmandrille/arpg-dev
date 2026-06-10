# v53 — Boss health bar UI

**Proves:** The Godot client now has a dedicated screen-anchored boss health bar driven by existing
authoritative boss entity state.

- Added an in-repo `BossHealthBar` control that displays boss title, hp/max hp, and a clamped fill
  ratio without changing server combat, boss generation, or protocol schemas.
- `main.gd` syncs the bar from live `is_boss` monster records, deterministically chooses the active
  live boss, derives `Cave Warden` from `boss_template_id`, and hides on death, remove, level clear,
  or gameplay teardown.
- Existing world-space monster health bars remain unchanged; the boss bar is an additional
  top-center HUD element for boss readability.
- `get_bot_state()` exposes `boss_health_bar` debug state, and `BotScenarioRunner` supports
  `wait_boss_health_bar` / `assert_boss_health_bar` steps with visibility, id/template/title,
  hp/max hp, and ratio expectations.
- Client bot scenario `26_boss_health_bar_ui.json` descends to the first boss floor and proves the
  live `cave_warden` bar is visible.
