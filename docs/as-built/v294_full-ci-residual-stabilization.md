# v294 As-Built: Full-CI Residual Stabilization

## Summary

v294 restored the full-CI baseline after v293 by stabilizing residual non-feature failures in the
elite-minion Go test, mercenary protocol scenarios, broad client bot scenarios, and click-driven bot
event matching.

## What Changed

- Stabilized `TestEliteMinionFollowsLeaderWithoutPassiveAggro` by clearing generated lab walls,
  selecting reachable leader/minion geometry, and asserting movement toward the production
  elite-minion follow goal.
- Made mercenary protocol scenarios tolerate v289 offer variants while still proving hire board
  behavior and death-loss cleanup through live companion assertions.
- Fixed client bot click waits so entity selector keys are only used to choose the clicked target,
  not to over-filter pending events emitted by combat interactions.
- Let `monster_def_id` event expectations match generic, source, or target monster fields.
- Updated older boss and mercenary client scenarios so broad UI/readability checks follow the
  current variable boss and mercenary-offer contracts.
- Kept `66_boss_telegraph_decals.json` pinned to the deterministic `boss_floor_gate` seed so its
  Cave Warden line, circle, and cone decal proof remains shape-specific.
- Drained one tick after the Crypt Matron kill in `second_boss_template` so the live recording
  observes the same summoned-add tail event that replay derives.
- Improved replay event-count mismatch diagnostics to identify the first extra or missing event.

## Verification

Focused gates run during stabilization:

```sh
go test ./internal/game -run 'TestEliteMinionFollowsLeaderWithoutPassiveAggro|TestEliteMinionAssistsLeaderTarget|TestEliteMinionDoesNotAttackWithoutLeaderEngagement|TestMercenary' -count=1
BOT_ADDR=:18082 BOT_BASE_URL=http://localhost:18082 make bot scenario=mercenary_hiring_board
BOT_ADDR=:18083 BOT_BASE_URL=http://localhost:18083 make bot scenario=mercenary_death_loss
BOT_ADDR=:18097 BOT_BASE_URL=http://localhost:18097 make bot scenario=second_boss_template
godot --headless --path client --script res://tests/test_client_bot.gd
HEADLESS=1 make bot-client
make ci
```

Previously failing client bot scenarios were also rerun individually with isolated `BOT_ADDR` /
`BOT_BASE_URL` ports:

```sh
make bot-client scenario=09_character_stats_panel.json
make bot-client scenario=11_combat_feedback.json
make bot-client scenario=12_model_reaction_polish.json
make bot-client scenario=19_skill_points_and_magic_bolt.json
make bot-client scenario=26_boss_health_bar_ui.json
make bot-client scenario=28_boss_phase_readability.json
make bot-client scenario=31_combat_threat_readability.json
make bot-client scenario=33_unique_burn_effect_live.json
make bot-client scenario=47_mercenary_roster_ui.json
make bot-client scenario=64_mercenary_combat_stats.json
make bot-client scenario=66_boss_telegraph_decals.json
```

## Deferred

- No gameplay tuning, balance work, new boss content, or new mercenary content shipped in this
  slice.
- The next session should run the due engineering review/refactor handoff before starting another
  feature autoloop.
