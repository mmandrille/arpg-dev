# v218 Rogue Duelist Mark - Implementation Plan

Goal: make Rogue stronger in 1v1 through a damage-amplifying mark, dash stun, execute passive, and
standard DEX-to-critical-damage scaling.

Architecture: keep gameplay tuning in shared JSON, keep outcomes server-authoritative and
deterministic, and avoid protocol schema changes by reusing existing combat and skill-effect events.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add mark, dash stun, and `executioner` passive data. |
| Modify | `shared/rules/skills.v0.schema.json` | Validate passive execute and new Rogue payload fields. |
| Modify | `shared/rules/character_progression.v0.json` | Add DEX scaling to `crit_damage`. |
| Modify | `server/internal/game/rules.go` | Parse and validate passive execute payloads. |
| Modify | `server/internal/game/rogue_rules.go` | Validate mark and dash-stun Rogue payload fields. |
| Modify | `server/internal/game/rogue_skills.go` | Store/advance marks and apply mark bonus to poison DOTs. |
| Modify | `server/internal/game/sim.go` / `sim_players.go` | Persist marks and apply mark/execute to player damage paths. |
| Modify | `server/internal/game/rogue_skills_test.go` | Add focused mark, dash stun, and execute passive tests. |
| Modify | `server/internal/game/game_test.go` | Guard DEX critical-damage scaling through existing rules tests. |
| Modify | `shared/assets/skill_presentations.v0.json`, `tools/bot/skill_demo.py` | Surface passive metadata in skill tooling. |
| Modify/Add | `client/scripts/player_status_effect_markers.gd`, `client/scripts/rogue_mark_effect.gd`, `client/scripts/main.gd` | Show a red skull over monsters with `rogue_mark`. |
| Modify | `client/tests/test_status_effect_presentation.gd` | Guard the Rogue mark skull cue and Dash stun cue. |
| Modify | `tools/bot/scenarios/47_rogue_class_foundation.json` | Include `executioner` in Rogue proof. |
| Add | `docs/as-built/v218_rogue-duelist-mark.md` | Record proof and scope limits. |

## Tasks

- [x] Write the spec and plan.
- [x] Extend shared rules and schemas.
- [x] Implement server mark, dash stun, passive execute, and DEX crit-damage behavior.
- [x] Update focused tests and Rogue bot scenario.
- [x] Run targeted verification.
- [x] Write as-built notes.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestRogue|TestLoadRules|TestCritDamageUsesDexterityAsStandardDerivedStat|TestDerivedStats|TestEffectiveAttackSpeedUsesWeaponAndItemPercent' -count=1
.venv/bin/pytest tools/bot/test_skill_demo.py tools/bot/test_protocol.py::test_load_scenarios_discovers_rogue_class_foundation tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q
make bot scenario=rogue_class_foundation
godot --headless --path client --script res://tests/test_status_effect_presentation.gd
make client-unit
```
