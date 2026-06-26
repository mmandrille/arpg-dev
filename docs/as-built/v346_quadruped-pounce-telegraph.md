# v346 — Quadruped pounce telegraph

Extends v345 windup to `dungeon_wolf` pounce attacks: larger gold ring and `attack_style: pounce` on `monster_attack_windup`.

Verification:
- `cd server && go test ./internal/game/... -run TestWolfPounceWindup`
- `make bot-visual scenario=01_click_to_kill` (wolf-heavy labs)
