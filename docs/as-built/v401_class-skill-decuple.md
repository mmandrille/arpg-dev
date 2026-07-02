# v401 As-built — Class Skill Decuple

## Proved

- All five classes now have exactly **5 actives + 1 mobility + 4 passives** (rogue completed in same slice).
- Removed `split_arrow` and `arcane_orb`; renamed `ligthing` → `lightning` with rank alias migration.
- Added tier-4 passives (level 15): `arcane_reservoir`, `unstoppable_heart`, `oathbound_resolve`, `wildborn_endurance`.
- Added actives: `skullcrusher`, `consecrated_smite`, `shadowstep`, `eviscerate`.
- `validate_skills.py` enforces decuple shape per class.

## Verification

- `make validate-shared`
- `go test ./internal/game/... -run 'TestThirdClass|TestLoadRules|TestRangerSkillRules'`
- `godot --headless --path client --script res://tests/test_skills_panel.gd`
- `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_skill_visual.py -q`
