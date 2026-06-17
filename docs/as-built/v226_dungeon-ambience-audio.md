# v226 As-Built - Dungeon Ambience Audio

Date: 2026-06-16

## What shipped

- Added generated town and dungeon ambience loops to `ClientAudioController`.
- Added ambience debug state for `ambient_zone` and `ambient_active`, exposed through the client bot
  state.
- Kept boss music higher priority than ambience: boss phases pause ambience, and ambience resumes
  when boss music stops.
- Routed current-level changes through `ClientAudioBridge` so `main.gd` only owns a narrow
  integration call.
- Added reusable `assert_audio_state` client-bot support for this and future audio slices.

## Proof

```bash
make client-unit
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
passed after the selected v226-v232 feature queue completed.

Manual visual/audio proof, if desired:

```bash
make bot-visual scenario=28_boss_phase_readability.json
```

## Scope limits

- No server, protocol, shared gameplay-rule, replay, HP, damage, loot, movement, boss AI, or dungeon
  generation changes shipped.
- No production audio assets, external audio packs/plugins, positional audio, authored stems,
  per-biome ambience, captions, or advanced audio settings shipped.
