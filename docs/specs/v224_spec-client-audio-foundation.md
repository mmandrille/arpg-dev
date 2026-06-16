# v224 Spec — Client Audio Foundation

Status: Implemented (2026-06-16)
Date: 2026-06-16
Codename: client-audio-foundation

## Purpose

Add a first playable audio layer to the Godot client so common on-screen actions are distinguishable by sound. The slice adds persistent audio volume settings, code-generated placeholder SFX cues for movement, attacks, potion/heal use, skill casts, damage, kills, and boss phases, plus a simple looped boss-floor music bed.

## Non-goals

- Production-quality sound design, external audio packs, recorded WAV/OGG assets, asset-manifest audio pipelines, and licensing/provenance research.
- Server, protocol, or shared gameplay-rule changes.
- Per-skill bespoke authored music/SFX beyond distinct placeholder cue families.
- Positional/3D audio, device selection, accessibility audio captions, and full settings remapping.

## Acceptance Criteria

- Settings persist and reload `master_volume`, `music_volume`, and `sfx_volume` from `user://settings.json` with safe defaults and clamped values.
- The settings panel exposes volume controls and immediately applies changes to the client audio controller.
- Existing local input and authoritative events trigger distinguishable placeholder sounds:
  - movement start,
  - local basic attack result/miss,
  - skill cast/effect cue,
  - potion or heal cue,
  - player/monster damage,
  - monster/boss kill,
  - boss phase start.
- Boss-floor presence or boss phase activity starts a looped code-generated music bed; leaving the boss context stops it.
- Audio playback is client-only presentation and does not affect deterministic simulation, protocol, replay, HP, damage, loot, movement, or skill outcomes.
- Headless/unit proof covers settings parsing/saving/clamping, settings panel sync, audio-controller cue routing, and boss music state.

## Scope and Likely Files

- Client:
  - `client/scripts/client_audio_controller.gd` — code-generated placeholder tones and music loop, volume application, cue routing.
  - `client/scripts/client_settings.gd` — persisted volume values and setters.
  - `client/scripts/settings_panel.gd` — volume sliders.
  - `client/scripts/main.gd` — wire settings changes and existing events/input to audio cues.
  - `client/tests/test_client_audio_controller.gd` — audio controller behavior.
  - `client/tests/test_audio_settings.gd` or an existing settings test — settings and panel proof.
- Shared i18n:
  - `shared/i18n/en.json`
  - `shared/i18n/es.json`
- Docs:
  - `docs/plans/v224_2026-06-16-client-audio-foundation.md`
  - `docs/as-built/v224_client-audio-foundation.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

## Test and Bot Proof

- `make client-unit` must pass.
- `make maintainability` must pass.
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1` should remain green and is the recommended visual/audio-context scenario for manual non-headless verification:

```bash
make bot-visual scenario=28_boss_phase_readability.json
```

The bot cannot hear audio in headless mode, so unit tests assert controller state and cue dispatch while the boss phase client scenario proves the same event stream still reaches the client.

## Client Asset / Plugin Decision

- Reject external audio packs/plugins for this slice. The project has no existing audio asset pipeline, and the requested sounds "not need to be great"; code-generated Godot `AudioStreamWAV` tones are enough to prove routing, settings, and boss music without adding licensing or import risk.
- Borrow existing settings-panel and event-routing patterns from `client_settings.gd`, `settings_panel.gd`, and `main.gd`.

## Open Questions and Risks

- No blocking questions. The conservative default is a small client-only foundation with placeholder tones and production audio deferred.
- Risk: `client/scripts/main.gd` is a grandfathered large file. Keep additions to minimal wiring and place audio behavior in a new focused controller.
