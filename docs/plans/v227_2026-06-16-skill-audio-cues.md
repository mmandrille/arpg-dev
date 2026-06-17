# v227 Plan - Skill Audio Cues

Status: Ready for implementation
Goal: Give active skills distinct generated cue families while preserving client-only presentation.
Architecture: Existing server-authored `skill_cast` events drive local audio; the client maps skill
IDs to cue families and exposes debug state for bot proof. No protocol or gameplay changes.
Tech stack: Godot GDScript client, client bot scenario proof, SDD docs.

## Baseline and shortcut decision

Reuse v226 `ClientAudioController` generated waveform and `assert_audio_state` support. Asset/plugin
decision: adopt existing code-generated cue synthesis; reject external audio assets, packs, and
plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/client_audio_controller.gd` | Map skill IDs to cue families and generated waveforms |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Let bot assertions check skill/cue debug fields |
| Modify | `client/tests/test_client_audio_controller.gd` | Unit proof for skill cue mapping |
| Modify | `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json` | Live client proof after `skill_cast` |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v227_skill-audio-cues.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: no grandfathered coordinator was touched.

Verification:
```bash
make maintainability
```

## Task 1 - Skill cue mapping

Files:
- Modify: `client/scripts/client_audio_controller.gd`
- Modify: `client/tests/test_client_audio_controller.gd`

- [x] Step 1.1: Add skill cue families for projectile, buff, movement, heal, revival, and generic
  fallback while recording `last_skill_id`.
```bash
make client-unit
```

## Task 2 - Bot proof

Files:
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Modify: `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json`

- [x] Step 2.1: Extend `assert_audio_state` for cue/skill fields and assert the Rage cast uses a
  buff cue in the live client scenario.
```bash
make bot-client scenario=19_skill_points_and_magic_bolt.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v227_skill-audio-cues.md`

- [x] Step 3.1: Record v227 as complete with focused proof and note the final batch CI is pending.
```bash
make maintainability
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=19_skill_points_and_magic_bolt.json HEADLESS=1`
- [x] Batch-level `make ci` passed after the selected v226-v232 `$autoloop` queue.
