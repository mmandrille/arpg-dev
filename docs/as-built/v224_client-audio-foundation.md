# v224 As-Built - Client Audio Foundation

Date: 2026-06-16

## What shipped

- Added `ClientAudioController`, a client-only Godot audio node that generates placeholder SFX and
  looped boss music from code instead of imported audio assets.
- Added persistent `master_volume`, `music_volume`, and `sfx_volume` settings with clamped load,
  save, and immediate settings-panel slider application.
- Routed existing client input and authoritative events to distinguishable semantic cues for
  movement, attacks, skills, heal/potion use, damage, kills, boss kills, and boss phase starts.
- Added `ClientAudioBridge` so `main.gd` only performs narrow wiring and stayed within its
  grandfathered maintainability baseline.
- Added Godot unit coverage for audio settings and controller cue/music state, and registered those
  tests in the client unit harness.

## Proof

```bash
make maintainability
make client-unit
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

All focused checks passed on 2026-06-16 during `$autoloop`.

Manual visual/audio proof, if desired:

```bash
make bot-visual scenario=28_boss_phase_readability.json
```

## Scope limits

- No server, protocol, shared gameplay-rule, replay, HP, damage, loot, movement, or skill outcome
  changes shipped.
- No production audio assets, external audio packs/plugins, positional audio, per-biome music,
  authored per-skill sound design, accessibility captions, or advanced audio settings shipped.
