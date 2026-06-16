# v225 Spec - Boss Audio Readability

Status: Implemented (2026-06-16)
Date: 2026-06-16
Codename: boss-audio-readability

## Purpose

Build on v224 client audio foundation by making boss phase audio more useful during the existing Cave Warden fight. Boss telegraph, active, recovery, summon, and line/ranged-style phase starts should produce distinct generated cue families and update a lightweight boss-music intensity state so players can hear when the fight changes pace.

## Non-goals

- Production audio assets, imported packs, plugins, licensing research, or an audio asset manifest.
- Server, shared rules, protocol schema, replay, boss AI, boss damage, HP, loot, movement, or timing changes.
- Positional/3D audio, adaptive music stems, accessibility captions, advanced audio settings, or per-biome ambience.
- Full authored per-skill sound design beyond the existing v224 generic skill cue routing.

## Acceptance Criteria

- `boss_phase_started` events route their existing `pattern_id` and `phase_kind` metadata into the client audio controller.
- Generated boss phase cues are distinguishable by semantic debug state and tone routing:
  - `telegraph` phases use a windup cue,
  - `active` phases use a danger cue,
  - `recovery` phases use a release cue,
  - summon patterns use a summon cue,
  - line/ranged patterns such as `stone_lance` use a ranged cue.
- Boss music remains client-only and starts from boss phase cues as before, but records a normalized `boss_music_intensity` and `boss_music_layer` that change by phase kind/pattern.
- Muted SFX/music settings still update debug state without attempting audible playback.
- The existing client boss phase readability scenario remains green:

```bash
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

- `make client-unit` covers cue classification, music intensity/layer state, mute behavior, and regressions for the v224 cue methods.

## Scope and Likely Files

- Client:
  - `client/scripts/client_audio_controller.gd` - classify boss phase cues and intensity from existing event metadata.
  - `client/scripts/client_audio_bridge.gd` - pass boss event metadata through the narrow bridge.
  - `client/scripts/main.gd` - replace the generic boss phase call with metadata-aware routing without growing the file.
  - `client/tests/test_client_audio_controller.gd` - focused unit proof for the new cue and music state behavior.
- Docs:
  - `docs/plans/v225_2026-06-16-boss-audio-readability.md`
  - `docs/as-built/v225_boss-audio-readability.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

## Test and Bot Proof

- `make client-unit`
- `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- `make maintainability`
- Batch close-out: `make ci`

The bot cannot hear headless audio. Unit tests own audio classification and state, while the existing boss client scenario proves the same boss phase event stream still reaches the client after the metadata-aware routing change.

## Client Asset / Plugin Decision

- Reject external audio packs/plugins. v224 already established code-generated `AudioStreamWAV` placeholders and the next useful step is better semantic routing, not imported sound design.
- Borrow existing v224 `ClientAudioController` / `ClientAudioBridge` patterns and existing server-authored boss phase event metadata.

## Open Questions and Risks

- No blocking questions.
- Risk: `client/scripts/main.gd` is a grandfathered over-limit file already at its allowed baseline growth. This slice must keep the change to a same-line replacement or shrink nearby code; audio behavior stays in focused controller/bridge files.
