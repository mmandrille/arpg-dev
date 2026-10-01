# v510 — Ranger affinity passive (detached worktree handoff)

- **Status:** Integrated; focused server/shared/client/bot/replay checks and combined batch `make ci` passed (11m41s, 2026-10-01).
- **Base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`; v501–v509 integrated working state was transferred into this worktree before implementation.
- **Approved choice:** Deadeye uses effective Dexterity; bounded coefficients are owned by shared skill rules.

## Behavior

Deadeye retains its existing critical chance bonus and adds a conditional bonus to direct Ranger ranged damage. The server counts positive, active Ranger affinity rolls on equipped gear, caps the count using the shared rule, and computes the percentage from effective Dexterity. The current percentage is projected as `ranged_damage_bonus_percent` in both v8 character-progression derived-stat schemas. Generic damage stats remain unchanged.

The Ranger bow basic projectile and explicitly tagged direct-hit Ranger skill projectiles receive the scaled raw damage range once. Persistent projectiles keep that range from launch; piercing and volley-family skills calculate it once at cast resolution. Melee, other classes' projectiles, pets, DoT, elemental follow-ons, and untagged skills do not use this helper. The skill-specific range and magic scaling run before the affinity multiplier; existing hit, armor, resistance, and crit resolution follows it. No new RNG is used.

The new Ranger quiver is a class-specialist belt item, so it can pair with the existing two-handed affinity bow. Both are placed near the player in `skill_progression_lab`. Existing item presentation and armor-look data cover the quiver; no new external asset or dependency was added. The skills tooltip reads rule coefficients and the server's current bonus, while the stats panel labels the bonus as conditional ranged damage. A shared golden drives Go and GDScript formula tests.

## Focused verification

| Check | Result | Evidence |
|---|---|---|
| Shared validation and CODEMAP | PASS | 2,231 shared checks plus `codemap ok`. The log explicitly validates both v8 schemas and `session_snapshot.json` / `state_delta.json` examples, along with the new skill tags, item, and Ranger golden. |
| `go test ./internal/game -run '^TestRangerAffinity' -count=1` | PASS | Golden, invalid rule bounds, equip/reversal, exclusions, Dex threshold, count cap, basic and skill projectile snapshots, piercing multi-hit, and same-input replay. |
| `go test ./internal/game -count=1` | PASS | Full game package rerun on final server/source state after the class-specialist catalog count was updated for the new quiver. |
| Scenario audit and catalog selection tests | PASS | Five movement-audit tests and four focused protocol catalog tests. |
| `make lint-determinism`, `make maintainability`, `git diff --check` | PASS | No new deterministic-map-order or size-ratchet violation. |
| `make client-unit` | PASS | Clean full rerun passed, including the Ranger GDScript golden, skill tooltip, character stats panel, and existing client tests. An earlier run was intentionally interrupted for v508's renewed runner reservation (exit 130); it was not a test failure. |
| `make bot scenario=affinity_passive_ranger` | PASS | Run with `-o pyproject.toml` to reuse the existing Python environment without reinstalling tools. Runtime progression checked 0→3→6→3→0; Piercing Shot produced a server `monster_damaged` event; the bot's `/state`, reconnect, and replay checks passed. |
| Recorded `make replay` | PASS | Session `sess_01M3W6WVSZZQKSWF9YZ455KNVP`, seed `v510_affinity_passive_ranger`: 11 inputs, 9 recorded events, 9 derived events; exact authoritative event match. |
| v451 Barbarian and Rogue affinity protocol bots | PASS | Both existing extended scenarios passed sequentially on the final v510 rules. |

The new protocol scenario is extended tier. It allocates Deadeye after equipping the bow, checks 0→3→6 percent as one then two active affinities contribute, observes a server `monster_damaged` event from Piercing Shot, and checks 6→3→0 after unequipping. Its fixed seed is `v510_affinity_passive_ranger` in `skill_progression_lab`. The level-10 fixture preserves effective Dexterity 17; level 15 would trigger the Ranger's level-growth floor of 22 and change the expected percentage. The retained [bot manifest](../../.artifacts/v510/bot-recorded-manifest.json) and [replay log](../../.artifacts/v510/replay.log) are local ignored evidence. Combined `make ci`, lifecycle closeout, and commits belong to the coordinator.

## Handoff boundaries

This worktree includes the transferred v501–v509 batch state. The v510-specific paths and full changed/untracked manifest follow below. No branch, commit, push, full CI, or coordinator checkout edit was made here. The 18 generated tracked Godot import sidecars from client runs were restored. Ignored `.artifacts/v510/` contains validation, server, client, bot, and replay logs plus the retained replay manifest and incomplete Python bootstrap; the local `.venv` link points to the coordinator's existing environment for tooling and does not change committed source.

**Complete checkout inventory:** [handoff-status.txt](assets/v510/handoff-status.txt) records all 164 changed and untracked paths, including the transferred v501–v509 batch state. The v510-specific edits are listed here so the coordinator can compare them with the integrated result:

```text
M  client/scripts/character_stats_breakdown.gd
M  client/scripts/character_stats_panel.gd
M  client/scripts/skill_passive_tooltip.gd
M  client/scripts/skills_panel.gd
M  docs/CODEMAP.md
M  docs/progress/scenario-movement-audit.tsv
M  scripts/client_smoke.sh
M  server/internal/game/class_specialist_gear_test.go
M  server/internal/game/derived_stats.go
M  server/internal/game/passive_skill_rules.go
M  server/internal/game/ranger_skills.go
M  server/internal/game/rules.go
M  server/internal/game/sim.go
M  server/internal/game/skill_damage_burst.go
M  server/internal/game/skill_rules_validation.go
M  shared/assets/armor_look.v0.json
M  shared/assets/item_presentations.v0.json
M  shared/protocol/examples/session_snapshot.json
M  shared/protocol/examples/state_delta.json
M  shared/protocol/session_snapshot.v8.schema.json
M  shared/protocol/state_delta.v8.schema.json
M  shared/rules/item_templates.v0.json
M  shared/rules/skills.v0.json
M  shared/rules/skills.v0.schema.json
M  shared/rules/worlds.v0.json
M  tools/validate_shared.py
?? client/scripts/ranger_affinity_damage.gd
?? client/tests/test_ranger_affinity_damage.gd
?? docs/as-built/assets/v510/handoff-status.txt
?? docs/as-built/v510_ranger-affinity-passive.md
?? docs/plans/v510_2026-10-01-ranger-affinity-passive.md
?? docs/specs/v510_spec-ranger-affinity-passive.md
?? server/internal/game/ranger_affinity_damage.go
?? server/internal/game/ranger_affinity_damage_test.go
?? shared/golden/ranger_affinity_damage.json
?? shared/golden/ranger_affinity_damage.v0.schema.json
?? tools/bot/scenarios/125_affinity_passive_ranger.json
?? tools/validate_ranger_affinity.py
```

`docs/CODEMAP.md`, `scripts/client_smoke.sh`, `server/internal/game/rules.go`, `server/internal/game/sim.go`, both v8 schemas, and the scenario audit also contain transferred earlier-slice edits. The scenario audit includes the inherited v506 living-town row, which was missing from the transferred table and is required by its audit test. Compare these shared paths rather than replacing them wholesale during integration. Ignored evidence under `.artifacts/v510/` comprises `validate-shared.log`, `game-package.log`, `client-unit-final.log`, `bot-ranger-final.log`, `bot-recorded.log`, `bot-recorded-manifest.json`, `replay.log`, both v451 bot logs, and setup/interrupted-run logs. `.godot/` is a local import cache. The recorded session remains in this worktree's test database for replay inspection.
