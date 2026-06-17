# v247 Spec - Mercenary Combat Stats

Status: Complete
Date: 2026-06-17
Codename: mercenary-combat-stats

## Purpose

Make hired mercenaries easier to evaluate by showing their server-owned combat stats in the
mercenary panel. The existing stats card shows identity and health; this slice adds the combat
numbers that explain what the guard contributes: damage range, attack cooldown, armor, block,
hit chance, and crit chance.

## Non-goals

- No mercenary balance changes, gear, level scaling, durable roster details, offer variants, new AI,
  new companion command UI, or production portraits.
- No client-local combat-stat tuning. Values must come from the authoritative companion entity view,
  which is built from shared monster rules or companion scaling.

## Client Asset / Plugin Decision

- **Adopt:** Existing mercenary panel stats card and companion bar sync.
- **Borrow:** Server companion combat fields already used by companion attack simulation.
- **Reject:** External assets/plugins and new visual icons.

## Acceptance Criteria

- Companion entity views expose a compact `combat_stats` object for companions.
- Hired mercenary companion state carries `combat_stats` through the client companion sync.
- The mercenary panel stats card includes damage, attack cooldown, armor, block, hit chance, and
  crit chance when present.
- The stats card remains hidden when no companion is hired and keeps existing identity/HP/stance/id
  lines.
- Focused server tests prove the mercenary view mirrors shared monster rules.
- Focused client tests prove the panel renders combat-stat lines from companion state.
- A client bot scenario hires a mercenary and asserts the new combat-stat text.

## Scope and Likely Files

- Server/protocol: `server/internal/game/types.go`, new helper under `server/internal/game/`,
  `server/internal/game/sim.go`, `shared/protocol/session_snapshot.v8.schema.json`.
- Client: `client/scripts/main.gd`, `client/scripts/mercenary_panel.gd`,
  `client/scripts/bot_mercenary_panel_assertions.gd`, `client/scripts/bot_step_catalog.gd`.
- Tests: `server/internal/game/companion_ai_test.go`, `client/tests/test_mercenary_panel.gd`.
- Bot/scenario: `tools/bot/scenarios/client/64_mercenary_combat_stats.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `cd server && go test ./internal/game -run 'MercenaryFoundation|MercenaryHiring' -count=1`
- `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- `make bot-client scenario=64_mercenary_combat_stats.json HEADLESS=1`
- `make validate-shared`
- `make maintainability`

## Open Questions and Risks

- No blocking questions.
- `main.gd` and `sim.go` are at their ratchet limits, so implementation must keep those edits
  line-neutral and move helper logic into small files.
