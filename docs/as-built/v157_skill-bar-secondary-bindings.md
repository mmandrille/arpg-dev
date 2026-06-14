# v157 As-built — Skill-bar secondary bindings

Date: 2026-06-14

## Shipped

- Expanded authoritative skill bindings from 8 to 16 slots.
- Slots 0-7 remain primary F1-F8 bindings.
- Slots 8-15 are secondary Shift+F1 through Shift+F8 bindings.
- `set_skill_bindings_intent`, snapshots, and deltas now carry 16 fixed `function_keys`.
- Character and session-start skill binding persistence now accepts slot indexes 0-15.
- The skills panel labels secondary assignments as `S-F1` through `S-F8`.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run TestSetSkillBindingsIntentUpdatesSnapshotAndDelta -count=1`
- `cd server && go test ./internal/store ./internal/replay -run 'TestCharacter|TestReplay' -count=1`
- `cd server && go test ./internal/inputdecode -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make client-unit`
- `make bot scenario=67_skill_secondary_bindings.json`

Visual/client verification command:

```bash
make bot-visual scenario=67_skill_secondary_bindings.json
```

## Notes

- Existing F1-F8 and right-click skill selection behavior is preserved.
- Existing short server-side binding payloads still normalize through the sim/store path, but current protocol examples use the full 16-slot contract.
- This slice does not add a full multi-slot cast bar; it only adds the secondary binding row and its authoritative persistence.
