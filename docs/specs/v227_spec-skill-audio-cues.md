# v227 Spec: Skill Audio Cues

Status: Complete
Date: 2026-06-16
Codename: skill-audio-cues

## Purpose

Make active skills sound more distinct by mapping existing `skill_cast` events to generated cue
families such as projectile, movement, buff, heal, and revival.

## Baseline

Builds on v226 dungeon ambience audio and v225 boss audio readability. The client already receives
authoritative `skill_cast` events and routes local-player casts to `ClientAudioController.play_skill`.
Asset/plugin decision: adopt the existing code-generated audio controller and bot audio assertion
surface; reject external audio assets, audio packs, and plugins.

## Non-goals

- No production SFX, imported samples, positional audio, cooldown timing changes, or animation
  changes.
- No server, protocol, shared skill rules, mana cost, cooldown, damage, or skill outcome changes.
- No per-rank or per-class authored sound library.

## Acceptance Criteria

- `ClientAudioController` maps known skill IDs to distinct generated cue families.
- The controller records the last skill ID and cue family for headless assertions, even when muted.
- Existing movement and heal classifications keep working.
- At least one live client-bot skill scenario asserts a skill-specific audio cue after `skill_cast`.
- Unknown skills continue to fall back to the generic skill cue.

## Scope and Likely Files

- `client/scripts/client_audio_controller.gd`
- `client/scripts/bot_assertion_handlers.gd`
- `client/tests/test_client_audio_controller.gd`
- `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json`
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v227_skill-audio-cues.md`

## Test and Bot Proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=19_skill_points_and_magic_bolt.json HEADLESS=1`

Manual visual/audio proof, if desired:

- `make bot-visual scenario=19_skill_points_and_magic_bolt.json`

## Open Questions and Risks

- No blocking questions. This is presentation-only and uses existing authoritative events.
