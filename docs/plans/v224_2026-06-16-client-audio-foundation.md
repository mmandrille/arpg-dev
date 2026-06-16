# v224 Plan — Client Audio Foundation

Status: Implemented (2026-06-16)
Goal: Add client-only generated audio cues, volume settings, and boss-floor music without changing gameplay authority.
Architecture: Godot owns all audio as presentation. `ClientAudioController` generates short placeholder streams at runtime, applies master/music/SFX volumes, and exposes semantic cue methods. `main.gd` only forwards existing local input and authoritative events to the controller; server/protocol/shared gameplay remain unchanged.
Tech stack: Godot GDScript client, shared i18n JSON, Godot headless unit tests, client bot boss-phase smoke.

## Baseline and Shortcut Decision

Builds on v223 with a clean `main` checkout. The slice reuses existing settings persistence, settings-panel construction, text catalog, boss-phase events, and local input hooks.

Asset/plugin decision: reject external audio assets and plugins for v224. Borrow existing Godot settings/event patterns and generate placeholder tones in code so the audio layer is playable, deterministic enough for tests, and free of licensing/import work.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/client_audio_controller.gd` | Generate placeholder SFX/music, apply volumes, route semantic cues. |
| Modify | `client/scripts/client_settings.gd` | Persist and clamp master/music/SFX volumes. |
| Modify | `client/scripts/settings_panel.gd` | Add volume sliders and sync helpers. |
| Modify | `client/scripts/main.gd` | Instantiate audio controller and forward existing input/events. |
| Add | `client/tests/test_client_audio_controller.gd` | Prove cue and music state behavior. |
| Add | `client/tests/test_audio_settings.gd` | Prove settings persistence and panel slider sync. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Settings labels for volume sliders. |
| Add | `docs/as-built/v224_client-audio-foundation.md` | Slice proof summary. |
| Modify | `docs/specs/v224_spec-client-audio-foundation.md`, `docs/plans/v224_2026-06-16-client-audio-foundation.md` | Mark implemented and completed tasks. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle close-out. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice: `client_audio_controller.gd` owns audio behavior.
- [ ] Defer extraction with rationale: n/a

Verification:

```bash
make maintainability
```

## Task 1 — Audio Controller

Files:
- Create: `client/scripts/client_audio_controller.gd`

- [x] Step 1.1: Add a focused `ClientAudioController` node that creates `AudioStreamPlayer` children for SFX and music, clamps volumes, generates short `AudioStreamWAV` tones/noise patterns, and exposes semantic cue methods.
- [x] Step 1.2: Add looped boss music generation and `start_boss_music()` / `stop_boss_music()` state.
- [x] Step 1.3: Add headless-safe introspection fields for last cue, cue count, and boss music active.

```bash
make client-unit
```

## Task 2 — Settings Persistence and Panel

Files:
- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/settings_panel.gd`
- Modify: `shared/i18n/en.json`
- Modify: `shared/i18n/es.json`
- Create: `client/tests/test_audio_settings.gd`

- [x] Step 2.1: Add default/clamped `master_volume`, `music_volume`, and `sfx_volume` fields, JSON load/save support, and setters.
- [x] Step 2.2: Add settings-panel sliders with translated labels and sync/update helpers.
- [x] Step 2.3: Add unit tests for missing/invalid/clamped settings, save/reload shape, and panel slider sync.

```bash
make client-unit
```

## Task 3 — Gameplay Audio Wiring

Files:
- Modify: `client/scripts/main.gd`
- Create: `client/tests/test_client_audio_controller.gd`

- [x] Step 3.1: Instantiate and apply settings to `ClientAudioController`.
- [x] Step 3.2: Wire settings-panel slider signals to settings save/apply and controller updates.
- [x] Step 3.3: Forward movement input, local attack outcomes, skill casts/effects, potion/heal events, damage/kills, boss kills, and boss phase starts to semantic audio cues.
- [x] Step 3.4: Start boss music when boss phase events begin or boss entities are active; stop it when no boss context remains.
- [x] Step 3.5: Add focused unit tests for cue routing helpers and boss music state where direct `main.gd` integration is too scene-heavy.

```bash
make client-unit
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 4 — Lifecycle Docs and Focused Verification

Files:
- Modify: `docs/specs/v224_spec-client-audio-foundation.md`
- Modify: `docs/plans/v224_2026-06-16-client-audio-foundation.md`
- Add: `docs/as-built/v224_client-audio-foundation.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 4.1: Mark spec/plan implemented.
- [x] Step 4.2: Add as-built proof and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md` latest slice, CI gate wording for autoloop focused verification, next slice, and open gaps/non-goals as needed.

```bash
make maintainability
make client-unit
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Final Verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] Batch-level `make ci`

## Deferred Scope

- Production audio assets, proper audio import manifests, positional audio, per-biome music, per-skill authored sound design, accessibility captions, and advanced audio settings remain deferred.
