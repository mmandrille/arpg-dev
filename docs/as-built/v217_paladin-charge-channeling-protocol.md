# v217 As-Built — Paladin Charge Channeling Protocol

Date: 2026-06-16

## Shipped

- Added additive `channel_skill_intent` protocol support with `start`, `update`, and `stop` phases.
- Added server-owned Charge channel state that moves per tick, turns on update intents, drains data-driven mana, and ends on release, blocked movement, or insufficient mana.
- Kept Charge cooldown-free and moved its channel speed/mana tuning into `shared/rules/skills.v0.json`.
- Reused existing Charge line impact for damage, stun, and push while preventing repeat impacts on the same monster during one channel.
- Wired Godot right-click Charge to send channel start/update/stop messages.
- Added a bot `channel_skill_path` action and changed `paladin_class_foundation` to one longer curved held Charge path.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/inputdecode ./internal/game -run 'TestDecodeChannelSkillIntent|TestPaladinCharge|TestSorcererTeleport|TestBarbarianLeap|TestRangerDisengage'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make client-unit`
- `make bot scenario=paladin_class_foundation`
- `make ci`

## Visual Verification

- `make bot-visual scenario=paladin_class_foundation`

## Notes

- Existing one-shot casts of Paladin Charge now reject with `use_channel_skill_intent`; other mobility skills keep `cast_skill_intent`.
- Channel mana uses `channel_mana_per_10_seconds` so fractional per-tick drains stay deterministic without hardcoding player-facing tuning in code.
