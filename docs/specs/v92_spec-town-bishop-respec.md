# v92 Spec — Town Bishop Respec

Status: Complete
Date: 2026-06-12
Codename: town-bishop-respec

## Purpose

Add a red town bishop NPC that gives players a town service distinct from merchants:
clicking the bishop restores the local character to full HP and mana, and opens a small service
menu with one action, `Respec`. The respec costs 250 gold and is server-authoritative: it refunds
all allocated stat points and all spent skill points, resets base stats to the character class
baseline, clears learned skill ranks and skill cooldowns, and lets the player rebuild through the
existing stat and skill panels.

## Non-goals

- No quest, dialog tree, blessing buff, resurrection, or long-term service progression.
- No partial respec, per-skill refund, free first respec, confirmation modal, or localized copy
  beyond the minimal in-client service labels needed for this slice.
- No production bishop art. The client may use an in-repo primitive/model composition, but it must
  be visually distinct from existing merchant presentation and predominantly red.
- No protocol compatibility preservation beyond updating the current schemas, bot, and client
  together.

## Acceptance Criteria

- Town and relevant town lab worlds include a ready `town_bishop` interactable.
- Clicking or actioning `town_bishop` restores the interacting player to current max HP and max
  mana and emits a bishop service event with current respec cost and affordability.
- The client renders the bishop as a non-merchant red NPC/model and opens a compact service menu
  with only `Respec` for now.
- Sending the respec action with at least 250 gold deducts 250 gold, resets class base stats,
  refunds all level-earned stat points, refunds spent skill ranks to unspent skill points, clears
  skill cooldowns, updates HP/mana caps, fills HP/mana, and emits authoritative progression,
  skill, cooldown, gold, resource, and service events.
- Sending the respec action with insufficient gold is rejected with `not_enough_gold` and mutates
  no progression, skills, resources, or gold.
- Respec cost is data-driven in shared rules/config, not hardcoded into tests as an unrelated
  tuning lock.
- A protocol bot scenario proves healing, successful paid respec, and rejected unaffordable
  respec through the same WebSocket protocol as the real client.
- A client bot or headless client assertion proves the bishop can be clicked and that the Respec
  menu control exists without making the client authoritative.

## Scope And Likely Files

- Shared rules/schemas: `shared/rules/main_config.v0.json`, `shared/rules/interactables.v0.json`,
  `shared/rules/worlds.v0.json`, matching schemas as needed.
- Protocol schemas: current envelope/message/session snapshot/state delta schemas for the new
  respec intent and bishop events.
- Server sim: input handler registration, bishop service activation, paid respec mutation,
  progression/resource/gold/cooldown changes, and focused Go tests.
- Client: interactable presentation in `client/scripts/main.gd`, small bishop service panel/menu,
  click action wiring for `bishop_respec_intent`, and focused headless/client-bot proof.
- Bot/tools: scenario JSON and Python helpers/assertions for bishop events and progression reset.
- Lifecycle docs: this spec, the v92 plan, `PROGRESS.md`, and v92 as-built.

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... ./internal/inputdecode/...`
- `make bot scenario=town_bishop_respec`
- `make client-unit`
- `make bot-client scenario=town_bishop_respec_panel HEADLESS=1`
- `make maintainability`
- `make ci`

## Client Shortcut Decision

The plan must record the Godot plugin/asset adoption checklist. Default decision for this slice:
reject external plugins and borrow no third-party asset, because the requested model can be a small
in-repo primitive composition and no reusable UI framework is needed for a one-option service menu.

## Open Questions And Risks

- No blocking questions. The conservative default is full class-baseline reset with all level-earned
  points refunded, paid by character gold, and no production art.
- Respec touches progression, skills, cooldowns, HP/mana caps, and persistence; the server must emit
  all related views in one authoritative tick so the client and bot do not infer state locally.
