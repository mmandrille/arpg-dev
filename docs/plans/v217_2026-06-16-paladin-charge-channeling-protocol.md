# v217 Plan: Paladin Charge Channeling Protocol

Spec: `docs/specs/v217_spec-paladin-charge-channeling-protocol.md`

## Architecture

Add an additive `channel_skill_intent` protocol path. The backend owns active channel state, movement, mana drain, collision, and Charge line impact each tick. The client only sends start/update/stop direction intents. Charge tuning remains in `shared/rules/skills.v0.json`, with cooldown still `none`.

## Tasks

- [x] Extend shared message and delta schemas for `channel_skill_intent` and channel lifecycle events.
- [x] Add data-driven Charge channel mana drain to `skills.v0.json` and schema-backed Go rule loading.
- [x] Add server input decode and handler support for channel start/update/stop.
- [x] Add server active channel tick processing for Charge movement, mana drain, path impacts, and end reasons.
- [x] Add client right-click channel helpers and payload tests.
- [x] Add bot channel path action and update `paladin_class_foundation` to one curved held Charge path.
- [x] Add focused tests and update as-built/progress docs.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/inputdecode ./internal/game -run 'TestDecodeChannelSkillIntent|TestCharge'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make client-unit`
- `make bot scenario=paladin_class_foundation`
- Final: `make ci`
