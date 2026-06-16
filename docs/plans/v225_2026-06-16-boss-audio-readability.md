# v225 Plan - Boss Audio Readability

Status: Implemented (2026-06-16)
Goal: Make existing boss phase events produce more informative generated client audio cues and boss-music intensity state.
Architecture: Keep audio client-only. Reuse v224 generated audio streams and route existing `boss_phase_started` metadata through `ClientAudioBridge` into `ClientAudioController`; no server, protocol, shared rule, or replay change is needed. Headless tests assert semantic cue and intensity state because bots cannot hear audio.
Tech stack: Godot GDScript client, Godot headless unit tests, client bot boss-phase scenario, SDD docs.

## Baseline and Shortcut Decision

Builds on v224 `client-audio-foundation` on `main`. Reuses existing boss phase events, boss readability client scenario, generated audio controller, and settings/mute behavior.

Asset/plugin decision: reject external audio packs/plugins and audio import pipeline work. Borrow v224 code-generated audio primitives and existing boss phase metadata (`pattern_id`, `phase_kind`) because the slice is about usefulness and routing, not production polish.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/client_audio_controller.gd` | Classify boss phase cues and intensity/layer state from event metadata. |
| Modify | `client/scripts/client_audio_bridge.gd` | Pass boss event metadata through narrow bridge helpers. |
| Modify | `client/scripts/main.gd` | Forward existing `boss_phase_started` event data without adding new authority or file growth. |
| Modify | `client/tests/test_client_audio_controller.gd` | Cover distinct boss cues, intensity/layer state, and muted state updates. |
| Add | `docs/specs/v225_spec-boss-audio-readability.md` | Slice spec. |
| Add | `docs/plans/v225_2026-06-16-boss-audio-readability.md` | This plan. |
| Add | `docs/as-built/v225_boss-audio-readability.md` | Shipped behavior and proof summary. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle close-out. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice: existing `client_audio_controller.gd` owns behavior; `main.gd` stays a wiring-only replacement.
- [ ] Defer extraction with rationale: n/a

Verification:

```bash
make maintainability
```

## Task 1 - Boss Audio Classification

Files:
- Modify: `client/scripts/client_audio_controller.gd`
- Modify: `client/tests/test_client_audio_controller.gd`

- [x] Step 1.1: Add debug fields for `boss_music_intensity`, `boss_music_layer`, `last_boss_pattern_id`, and `last_boss_phase_kind`.
- [x] Step 1.2: Change `play_boss_phase` to accept `pattern_id` and `phase_kind`, classify summon/ranged/telegraph/active/recovery cue names, and update intensity/layer before starting boss music.
- [x] Step 1.3: Extend cue frequency/wave routing for the new generated boss cue families.
- [x] Step 1.4: Add unit tests for distinct boss cues, intensity/layer state, and muted state updates.

```bash
make client-unit
```

## Task 2 - Boss Event Wiring

Files:
- Modify: `client/scripts/client_audio_bridge.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Add bridge overload/helper support for passing the boss phase event dictionary.
- [x] Step 2.2: Replace the generic boss phase call in `main.gd` with metadata-aware routing while keeping `main.gd` at or below its current maintainability allowance.
- [x] Step 2.3: Keep boss kill music stop behavior unchanged.

```bash
make client-unit
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 3 - Lifecycle Docs and Focused Verification

Files:
- Modify: `docs/specs/v225_spec-boss-audio-readability.md`
- Modify: `docs/plans/v225_2026-06-16-boss-audio-readability.md`
- Add: `docs/as-built/v225_boss-audio-readability.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 3.1: Mark spec/plan implemented and check off completed tasks.
- [x] Step 3.2: Add as-built proof and lifecycle row.
- [x] Step 3.3: Update `PROGRESS.md` latest slice, CI gate wording, next slice, and deferred audio scope.

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

Production audio assets, external packs/plugins, positional audio, authored adaptive music stems, accessibility captions, per-biome ambience, and advanced audio settings remain deferred.
