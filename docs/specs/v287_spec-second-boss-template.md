# v287 Spec: Second Boss Template

Status: Implemented
Date: 2026-06-19
Codename: `second-boss-template`

## Purpose

Add one additional boss template to the existing boss-floor system and make boss-template pool
selection deterministic so generated boss floors can visibly use more than Cave Warden.

The first new template is `crypt_matron`: a data-authored boss using existing monster visuals,
existing boss patterns, existing boss reward tables, and existing boss-floor gating behavior.

## Non-goals

- Do not add new boss pattern shapes, new telegraph rendering, new boss arenas, or new reward tables.
- Do not rebalance the full boss economy, dungeon cadence, or boss-floor layout.
- Do not add multi-boss floors, weighted random decks, or production boss art.
- Do not change existing Cave Warden proof seeds.

## Acceptance Criteria

- `shared/rules/boss_templates.v0.json` contains a second boss template named `crypt_matron`.
- `crypt_matron` references valid existing monster, pattern, loot, and visual data.
- `shared/rules/dungeon_generation.v0.json` includes both `cave_warden` and `crypt_matron` in the
  boss template pool.
- Boss-floor generation selects a template deterministically from the pool using seed and level.
- Existing `boss_floor_gate`-style seeds continue selecting `cave_warden`.
- A dedicated seed selects `crypt_matron` and exposes its boss template id, base monster def,
  visual model, visual tint, and scale in the generated boss entity.
- Protocol bot proof observes `crypt_matron`, at least one authored pattern from its deck, boss kill,
  and exit unlock.

## Scope And Likely Files

- `shared/rules/boss_templates.v0.json` adds `crypt_matron`.
- `shared/rules/dungeon_generation.v0.json` adds `crypt_matron` to the pool.
- `server/internal/game/dungeon_gen.go` changes boss template selection from first-entry to
  deterministic seed/level selection.
- `server/internal/game/game_test.go` or a focused boss test adds selection coverage.
- `tools/bot/scenarios/95_second_boss_template.json` proves the new template through protocol.
- Existing client boss health bar and portrait fallback should continue to render generic boss titles
  without a client data change.

## Test And Bot Proof

Focused checks:

```bash
(cd server && go test ./internal/game -run 'TestBoss' -count=1)
make validate-shared
make bot scenario=second_boss_template
make maintainability
```

Visual verification command for humans/agents:

```bash
make bot-visual scenario=second_boss_template
```

## Asset And Plugin Decision

- Adopt: existing boss pattern definitions, boss floor layout, boss loot tables, boss visual model
  identifiers, and generic boss health bar/portrait fallback.
- Borrow: existing dungeon RNG helpers and boss proof scenarios.
- Reject: new art assets, external plugins, new boss arenas, new pattern engine work, and new reward
  economies.

## Outcome

- `crypt_matron` ships as a second data-authored boss template using the existing skeleton model,
  summon-bat opener, boss-floor gate, and boss loot table.
- Deterministic pool selection keeps known Cave Warden proof seeds stable while seed
  `second_boss_template` selects Crypt Matron.
