# v226 Spec: Dungeon Ambience Audio

Status: Complete
Date: 2026-06-16
Codename: dungeon-ambience-audio

## Purpose

Add client-only generated ambience loops so town, dungeon, and boss spaces feel distinct while
preserving the server-authoritative gameplay boundary.

## Baseline

Builds on v225 boss audio readability. The client already owns code-generated SFX and boss music in
`ClientAudioController`, with settings persistence and headless debug state. Asset/plugin decision:
adopt the existing code-generated audio controller and bridge; reject external audio assets, audio
packs, and plugins for this slice.

## Non-goals

- No production audio assets, imported stems, positional audio, biome-specific authored tracks, or
  protocol/server changes.
- No new audio settings beyond the existing master/music/SFX controls.
- No gameplay, HP, damage, loot, movement, or dungeon generation changes.

## Acceptance Criteria

- The client starts a generated ambient loop for town/world level `0`.
- The client starts a different generated ambient loop for dungeon levels below `0`.
- Boss music remains higher priority than ambience while active and ambience resumes when boss music
  stops.
- Muted music settings still update ambience debug state without playing audible audio.
- A headless client-bot scenario can assert the current ambience zone and loop state.

## Scope and Likely Files

- `client/scripts/client_audio_controller.gd`
- `client/scripts/client_audio_bridge.gd`
- `client/scripts/main.gd`
- `client/tests/test_client_audio_controller.gd`
- `tools/bot/scenarios/client/28_boss_phase_readability.json` or a focused client scenario
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v226_dungeon-ambience-audio.md`

## Test and Bot Proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`

Manual visual/audio proof, if desired:

- `make bot-visual scenario=28_boss_phase_readability.json`

## Open Questions and Risks

- No blocking questions. The slice uses existing level state and boss events; if a precise level
  transition hook is missing, the implementation should add a narrow client bridge helper rather
  than change protocol.
