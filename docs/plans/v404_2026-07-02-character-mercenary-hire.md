# v404 Plan — Character Mercenary Hire

Date: 2026-07-02
Spec: [`docs/specs/v404_spec-character-mercenary-hire.md`](../specs/v404_spec-character-mercenary-hire.md)

## Tasks

- [x] Shared: `mercenary_hire_cost_gold_per_level` + protocol v8 extensions
- [x] Server: character roster load, hire flow, melee/ranged companion combat
- [x] Client: mercenary board node, candidate picker panel, class-model companion visual
- [x] Bot: protocol `97_character_mercenary_hire` + client `70_character_mercenary_picker_ui`
- [x] Update CI pack mercenary client scenario `47_mercenary_roster_ui`

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestMercenary|TestCharacterMercenary|TestCompanionStance' -count=1
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot scenario=character_mercenary_hire
make bot-client SCENARIO=mercenary_roster_ui HEADLESS=1
```
