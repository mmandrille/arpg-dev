# v206 Spec: Mercenary Hiring Board

Status: Complete
Date: 2026-06-15
Codename: mercenary-hiring-board

## Purpose

Add the first server-authored mercenary hiring service. A player can interact with a mercenary board, pay a configured gold cost, and receive one owned `mercenary_guard` companion that reuses the existing companion AI and companion HUD/entity presentation.

This is intentionally the smallest vertical hire path: it proves the authoritative contract, gold spend, companion spawn, and bot flow before adding richer roster UI or player-character-derived mercenary snapshots.

## Baseline

Builds on v205 and the existing companion/mercenary foundation:

- v182 introduced server-owned companion entities, follow/assist AI, and companion combat events.
- v198 added the `mercenary_guard` archetype, visual metadata, i18n, and a protocol proof for authored mercenary companion behavior.
- ADR-0010 frames hired mercenaries as derived combat copies and explicitly leaves listing, pricing, snapshots, death rules, and roster depth for future slices.

Asset/plugin decision: reject external assets/plugins. The board uses a rule-authored interactable and existing primitive town-service presentation patterns; the hired mercenary uses the existing `mercenary_guard` companion visual.

## Non-goals

- No player-character mercenary export/listing/snapshot persistence.
- No player-set pricing, offer browser, market integration, durable hire records, death recovery, insurance, reputation, or roster management.
- No Godot mercenary roster panel or rich hiring UI; a client UI slice follows separately.
- No companion stance commands, equipment, inventory, loot, XP, potion use, or mercenary leveling.
- No multiple simultaneous hired mercenaries; hiring replaces any existing hired mercenary from this service.

## Acceptance Criteria

- Shared config defines a non-negative `mercenary_hire_cost_gold` and validation rejects invalid values.
- Shared interactable rules define `town_mercenary_board` as a ready service interactable with service `mercenary`.
- The server treats `action_intent` on `town_mercenary_board` as a hire attempt only when the board is present/in range and the player has enough gold.
- On success, the server subtracts the configured gold, spawns one owned `mercenary_guard` companion near the player, emits `mercenary_hired`, and updates the player's gold in the authoritative state.
- Re-hiring from the board prunes/replaces the prior hired mercenary for that owner rather than stacking unlimited hired guards.
- On insufficient gold or invalid/out-of-range board target, the server rejects the intent and spawns no companion.
- A focused protocol bot scenario opens/hits the board path, hires a mercenary, observes the companion, moves the owner, and proves the mercenary can damage a target.

## Scope and Likely Files

- Shared rules/schema: `shared/rules/main_config.v0.json`, `shared/rules/main_config.v0.schema.json`, `shared/rules/interactables.v0.json`, `shared/rules/interactables.v0.schema.json`, `shared/rules/worlds.v0.json`.
- Protocol schemas: `shared/protocol/state_delta.v8.schema.json`, `shared/protocol/session_snapshot.v8.schema.json`.
- Server: `server/internal/game/rules.go`, a focused mercenary service helper/test file, `server/internal/game/interactables.go`, `server/internal/game/types.go`.
- Bot: `tools/bot/scenarios/88_mercenary_hiring_board.json` using existing action steps.
- Client: optional primitive recognition for `town_mercenary_board` only if default rendering or smoke requires it; no panel.
- Docs: v206 plan/as-built/lifecycle updates.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestMercenaryHiring|TestMercenaryFoundation|TestCompanion'`
- `make bot scenario=mercenary_hiring_board`
- `make bot scenario=mercenary_foundation`
- `make maintainability`
- Final `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: using `action_intent` makes the first hire interaction coarse. This is intentional for the thin board slice; the roster UI slice can add richer browsing/selection if needed.
- Risk: pricing can become economy-sensitive. This slice uses a single configurable cost in `main_config`; richer price formulas remain deferred.
