# v404 — character-mercenary-hire

## Proved

- Same-account character mercenaries replace fixed guard/scout hires.
- Hire cost uses `mercenary_hire_cost_gold_per_level × active hero level`.
- Level cap scales **down only**: `effective_level = min(source, active)`; no power boost when source is lower level.
- Board lists `mercenary_candidates`; hire intent sends `mercenary_character_id`.
- Melee and ranged companions follow owner, aggro monsters, and use equipped weapon mode.
- `mercenary_lost` on death; source character unchanged.
- Client merc-style board node, candidate picker UI, and class-model companion visuals.

## Key files

- `server/internal/game/mercenary_character.go`, `mercenary_hiring.go`, `companion_ranged.go`
- `server/internal/realtime/mercenary_roster.go`
- `client/scripts/mercenary_panel.gd`, `town_node_factory.gd`, `main.gd`
- `tools/bot/scenarios/97_character_mercenary_hire.json`
- `tools/bot/scenarios/client/70_character_mercenary_picker_ui.json`

## Deferred

- Cross-account mercenary listings (ADR-0010)
- Durable merc persistence across sessions
- Full equipment paper-doll on hired merc visual
