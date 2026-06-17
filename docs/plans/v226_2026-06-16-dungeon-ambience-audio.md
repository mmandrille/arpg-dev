# v226 Plan - Dungeon Ambience Audio

Status: Ready for implementation
Goal: Add generated town/dungeon ambience loops that yield to boss music.
Architecture: This is a client-only presentation slice. Existing authoritative level and boss
state drive audio selection; no server, protocol, or shared rule changes are needed.
Tech stack: Godot GDScript client, client bot scenario proof, SDD docs.

## Baseline and shortcut decision

Reuse v224-v225 `ClientAudioController` and `ClientAudioBridge`. Asset/plugin decision: adopt
existing code-generated waveform loops; reject external audio assets, packs, and plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/client_audio_controller.gd` | Own ambience stream state and boss-priority behavior |
| Modify | `client/scripts/client_audio_bridge.gd` | Provide narrow helper for level/ambience updates |
| Modify | `client/scripts/main.gd` | Call bridge from current-level changes without owning audio logic |
| Modify | `client/tests/test_client_audio_controller.gd` | Unit proof for ambience zones and boss priority |
| Modify | `tools/bot/scenarios/client/28_boss_phase_readability.json` | Headless proof of boss ambience/music state |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v226_dungeon-ambience-audio.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: main should only receive a narrow bridge call, so splitting
  the presentation router is riskier than the small integration touch.

Verification:
```bash
make maintainability
```

## Task 1 - Ambience controller state

Files:
- Modify: `client/scripts/client_audio_controller.gd`
- Modify: `client/tests/test_client_audio_controller.gd`

- [x] Step 1.1: Add generated ambience streams for `town` and `dungeon`, with debug state for
  `ambient_zone`, `ambient_active`, and boss-priority pause/resume.
```bash
make client-unit
```

## Task 2 - Level wiring and bot proof

Files:
- Modify: `client/scripts/client_audio_bridge.gd`
- Modify: `client/scripts/main.gd`
- Modify: `tools/bot/scenarios/client/28_boss_phase_readability.json`

- [x] Step 2.1: Route current-level changes through the audio bridge and expose bot assertions for
  ambience state using existing client debug surfaces.
```bash
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v226_dungeon-ambience-audio.md`

- [x] Step 3.1: Record v226 as complete with focused proof and note the final batch CI is pending.
```bash
make maintainability
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] Batch-level `make ci` is deferred to `$autoloop` after the selected queue commits.
