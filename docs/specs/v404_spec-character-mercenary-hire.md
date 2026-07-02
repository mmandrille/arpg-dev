# v404 Spec — Character Mercenary Hire

Status: Approved
Date: 2026-07-02
Codename: `character-mercenary-hire`

## Purpose

Replace the door-shaped mercenary board and fixed `mercenary_guard`/`mercenary_scout` hires with
same-account **character mercenaries**: the active hero opens a merc-style board, picks another
**alive** character on the account, pays `10 × active hero level` gold, and receives a companion
spawned from that character's real equipped gear and level-capped stats.

Level cap (no power boost): `effective_level = min(source_level, active_level)`. Combat stats scale
only when the source character is higher level than the active hero:
`scaled = source_value × (effective_level / source_level)`. When the source is lower or equal,
stats stay at the source character's true values.

Mercenaries do not use skills. They attack with equipped weapons (melee or ranged, like monsters:
follow owner, aggro monsters). Merc death emits `mercenary_lost`; the source character is
unchanged. Re-hire spends gold again.

## Non-goals

- Cross-account or cross-player mercenary listings (ADR-0010 future scope)
- Scaling mercenaries **up** when source level is below active hero level
- Mercenary skills, hotbar, mana, loot, XP, or potion use
- Durable mercenary roster persistence across sessions
- Production mercenary art (code-native board + existing class visuals)
- Full paper-doll equipment overlay on hired merc visual (class model is enough for v403)

## Acceptance criteria

- `town_mercenary_board` renders a dedicated merc-style board node (not door fallback).
- `mercenary_board_opened` lists same-account alive characters excluding the active hero, each with
  `character_id`, name, class, level, `price = mercenary_hire_cost_gold_per_level × active_level`,
  and `affordable`.
- Hire intent includes `mercenary_character_id`; invalid/dead/active/missing candidates reject.
- Hired companion uses source equipped items for attack mode and combat stats at capped level.
- Ranged mercenaries (bow) fire projectiles; melee mercenaries (hammer, etc.) use melee attacks.
- `mercenary_lost` on merc death; source character remains alive in store.
- Re-hire after loss works at current price formula.
- Hiring replaces prior board merc (existing v206 behavior).

## Scope and likely files

- Shared: `main_config.v0.json`, protocol v8 extensions (`mercenary_character_id` on action intent,
  `mercenary_candidates` on board event, `source_character_id` on hire/lost).
- Server: `mercenary_character.go`, `mercenary_hiring.go`, `companion_ai.go`, companion ranged helper,
  `session_loop.go` roster load, `types.go`.
- Client: `town_node_factory.gd`, `mercenary_panel.gd`, `mercenary_panel_bridge.gd`, `main.gd`.
- Bot: new protocol + client scenarios with two-character preflight.

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestMercenary|TestCharacterMercenary' -count=1
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot scenario=97_character_mercenary_hire
make bot-client scenario=70_character_mercenary_picker_ui HEADLESS=1
```

## Asset and plugin decision

- Adopt: market-board primitive pattern, existing companion AI/HUD/panel, character class models.
- Borrow: monster ranged projectile path for mercenary ranged attacks.
- Reject: external assets/plugins.

## Open questions

Resolved: same-account only; no scale-up; melee+ranged in one slice.
