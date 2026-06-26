# v345 — Monster melee windup cue

`dungeon_mob` now uses `attack_windup_ticks: 8`. Server delays melee damage and emits `monster_attack_windup`; client `MonsterMeleeWindupMarker` shows a fading ring and triggers the attack clip.

Verification:
- `cd server && go test ./internal/game/... -run TestDungeonMobMeleeWindupDelaysDamage`
- `godot --headless --path client --script res://tests/test_look_and_feel_polish.gd`

Visual: `make bot-visual scenario=01_click_to_kill`
