# v208 Spec: Companion Stance Command

Status: Complete
Date: 2026-06-16
Codename: companion-stance-command

## Purpose

Let players command owned companions into a small set of server-authoritative combat stances. Companions should default to the current assist behavior, but the protocol should support switching all active owned companions between `assist`, `defend`, and `passive` so future client UI can present real commands without adding new simulation semantics.

## Baseline

Builds on v206/v207:

- Hired mercenaries and summoned/revived companions are represented as owned `companion` entities.
- Companion AI currently follows the owner and attacks nearby monsters within a fixed assist radius.
- Entity views already expose `owner_id` and `target_id` for companions, and bot scenarios can assert entity state over the shared protocol.

Asset/plugin decision: reject external assets/plugins and new UI art. This slice changes authoritative simulation and protocol contracts only; a Godot command panel/button treatment is deferred.

## Stances

- `assist`: default/current behavior. The companion may target valid monsters near itself.
- `defend`: the companion may target valid monsters near its owner, then move to engage.
- `passive`: the companion follows its owner but never selects or attacks a target.

The command applies to all living companions owned by the acting player on the active level. This keeps the first command contract useful without requiring target-selection UI.

## Non-goals

- No Godot stance controls or visual stance selector.
- No per-companion targeting, hold-position behavior, retreat command, or leash tuning.
- No mercenary death/loss, gear snapshot refresh, loot/XP/potion behavior, or pricing/listing changes.
- No persistence of stance across logout/session reconstruction beyond the current simulation entity state.

## Acceptance Criteria

- Clients can send `companion_command_intent` with `stance` equal to `assist`, `defend`, or `passive`.
- Invalid stances are rejected with a stable reason, and a command with no active owned companion is rejected.
- New companions default to `assist`.
- Entity snapshots and deltas expose `companion_stance` for companion entities.
- Successful commands emit `companion_stance_changed` with the player id, stance, and affected companion count.
- `passive` clears companion targeting and prevents attacks while preserving follow behavior.
- `defend` targets monsters near the owner rather than monsters that are only near the companion.
- A protocol bot scenario proves the command against a hired mercenary.
- Focused Go tests cover decode/handler validation and stance-specific AI behavior.

## Scope and Likely Files

- Server simulation: add companion stance state, command handler, defaulting, event emission, and AI target filtering.
- Input decoding: add `companion_command_intent` payload decoding.
- Shared protocol: update message, snapshot, and delta schemas for the intent, entity field, and event.
- Bot runner: add a small `set_companion_stance` action and selector support for `companion_stance`.
- Scenario: add `tools/bot/scenarios/89_companion_stance_command.json`.
- Docs: update plan/as-built/progress/lifecycle/scenario catalog.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/inputdecode ./internal/game -run 'TestDecodeCompanionCommandIntent|TestCompanionStance'`
- `make bot scenario=companion_stance_command`
- `make bot scenario=mercenary_hiring_board`
- `make maintainability`
- Final `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: protocol-only command is not yet clickable in the Godot UI. This is intentional so the command contract and AI semantics are correct before adding UI controls.
- Risk: `defend` could be tuned in several ways. This slice defines it conservatively as "attack monsters near the owner" and leaves richer leash/formation tuning to later slices.
