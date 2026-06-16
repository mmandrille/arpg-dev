# v214 Spec: Paladin Charge Channel

Status: Draft
Date: 2026-06-16
Codename: paladin-charge-channel

## Purpose

Turn Paladin `charge` from a one-shot aimed mobility cast into a held-button channel: while the player keeps right click pressed and has enough mana, the Paladin keeps charging in the pointer direction, visibly moves at a fast run, can turn as the pointer direction changes, and displaces plus stuns monsters crossed by the charge lane.

## Baseline

Builds on v211 Charge. The current compatibility path keeps direction-only `cast_skill_intent` usable for bots and existing clients by performing one server-approved charge segment with line impact. Asset/plugin decision: adopt existing code-native character motion and bot visual infrastructure; reject outside assets/plugins for this contract change.

## Non-goals

- No production VFX/audio beyond fast walking presentation.
- No target-lock behavior for Charge.
- No hardcoded balance constants in gameplay code.

## Acceptance Criteria

- Charge has data-driven channel tuning: movement speed, mana-per-second, stun duration, lane width, and monster displacement.
- Client right-click press starts the Charge channel in the current pointer direction; release stops it.
- While the channel is active, moving the pointer changes the charge direction without ending the skill.
- Server owns channel movement, mana drain, collision-safe displacement, monster stun, and channel stop when mana is insufficient.
- Charge has no cooldown; mana drain is the limiting resource for the channel.
- Charge does not require or accept a monster target while channeling.
- Bot coverage can hold one Charge activation long enough to trace a curved or circular trajectory from the starting point through multiple enemies before release.
- `paladin_class_foundation` ends with one held Charge demonstration that curves through enemies and asserts channel start, channel stop, stun, and displacement events.

## Protocol Shape

Add a channel-control input that is explicit about start/stop phases, for example:

- `channel_skill_intent`: `{ "skill_id": "charge", "phase": "start", "direction": { "x": 1, "y": 0 } }`
- `channel_skill_intent`: `{ "skill_id": "charge", "phase": "update", "direction": { "x": 0, "y": 1 } }`
- `channel_skill_intent`: `{ "skill_id": "charge", "phase": "stop" }`

The exact name can change during planning, but the contract must distinguish held-button start, direction updates, and release, and must not overload target-based `cast_skill_intent`.

## Scope and Likely Files

- Shared protocol schemas for the new channel-control intent.
- Shared skill rules/schema for Charge channel speed, mana drain, lane width, and push tuning.
- Server sim state for active skill channels and per-tick mana/movement/effects.
- Client sustained right-click input and bot scenario channel actions, including a path action that sends direction updates over one held channel.
- Focused Go tests, protocol validation tests, and client unit coverage where practical.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestPaladinCharge|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=paladin_class_foundation`
- Visual manual check: `make bot-visual scenario=paladin_class_foundation`

## Open Questions and Risks

- Decide whether repeated hits on the same monster during one channel are allowed or require a per-target lockout window.
