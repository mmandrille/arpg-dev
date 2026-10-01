# v502 As-Built — Skill-specific impact VFX

- **Date:** 2026-10-01
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc` (detached worktree HEAD; no commits)
- **Prerequisite:** Coordinator-transferred, checksum-verified v501 change set is present in this worktree. Preserve it during integration.
- **Status:** Integrated; focused checks, `make client-unit`, the Magic Bolt client bot gate, and combined batch `make ci` passed (11m41s, 2026-10-01).
- **Spec:** [`v502_spec-skill-specific-vfx.md`](../specs/v502_spec-skill-specific-vfx.md) · **Plan:** [`v502_2026-10-01-skill-specific-vfx.md`](../plans/v502_2026-10-01-skill-specific-vfx.md)

## Delivered

- Added schema-backed impact presets and mappings for Magic Bolt, Ice Shard, and Lightning. Basic/unmapped hits keep `hit_spark`; death keeps `death_burst`, independent of `skill_id`.
- Added an allowlisted capture directive to the existing `skill_visual` replay path. It applies hit state before diagnostics/capture, focuses the target through replay-only camera settings, validates capture names, and uses the existing viewport writer.
- Moved the replay camera/capture orchestration and capture-only presentation diagnostics out of `main.gd` into `SkillVisualCaptureRuntime`. This keeps `main.gd` within the repository file-size ratchet while preserving the replay path.
- Moved only the `skill_visual_lab` soft target from `(24,5)` to `(27,5)`. The one-hit XP dummies at `(10,5)` and `(10,6)` had been within Lightning's configured first-chain range (`18 × 0.8 = 14.4` tiles), producing an unrelated level-up and off-screen death burst over the capture. At `(27,5)` both are outside chain range; the showcase caster stages at `(18,5)`, still within Lightning's 18-tile projectile range. A Python regression check derives these limits from shared skill data.
- Increased Lightning's catalog particle count and size range after the clean fixture capture showed the original burst too small. It remains short-lived and yellow-white; at performance quality the final capture reports 12 particles.

No server simulation, protocol, gameplay tuning, production monster placement, XP reward, external asset, plugin, or asset-pipeline changes were made. The only world-data adjustment is the visual-test lab fixture.

## Renderer evidence

All images use the actual Metal / Forward+ renderer at 1920×1080 with the performance quality preset. Each fixture records a successful, unblocked `monster_damaged` event for the intended skill, an on-screen emitting effect, and an incremented impact-feedback count.

| Skill | Command | PNG and fixture | Inspection |
|---|---|---|---|
| Magic Bolt | `ARPG_SKILL_VISUAL_SKILL_ID=magic_bolt make bot-visual scenario=skill_visual` | `.artifacts/bot-captures/v502_skill_magic_bolt.png` and `.json` | Blue burst visible; manifest: 12 particles, on-screen, captured 10:33:19 at the original lab target position. |
| Ice Shard | `ARPG_SKILL_VISUAL_SKILL_ID=ice_shard make bot-visual scenario=skill_visual` | `.artifacts/bot-captures/v502_skill_ice_shard.png` and `.json` | Cyan-white separated shatter visible; manifest: 15 particles, on-screen, captured 10:39:29 at the original lab target position. |
| Lightning | `ARPG_SKILL_VISUAL_SKILL_ID=lightning make bot-visual scenario=skill_visual` | `.artifacts/bot-captures/v502_skill_lightning.png` and `.json` | Final clean yellow-white burst is visibly distinct; manifest: 12 particles, lifetime 0.24s, on-screen at `(27.4,1,5)`, captured 10:52:07. The replay manifest reports `status=passed` and `replay_match=true`. |

Magic Bolt and Ice Shard captures predate the isolated lab-target adjustment; their VFX, successful hit mapping, camera focus, and quality tier are unchanged by that Lightning-only fixture correction. Lightning's final capture replaces the earlier contaminated image. Captures prove these replay-path visuals only; they are not performance or GPU-cost evidence.

The final Lightning `make bot-visual` process exited 0. Godot printed existing RID/ObjectDB/resource leak warnings during shutdown; the capture was saved and the replay completed successfully.

## Verification

| Check | Result |
|---|---|
| `.venv/bin/pytest tools/bot/test_skill_visual.py -q` | PASS — 17 tests, including fixture cast-range and chain-isolation checks. |
| `make validate-shared` | PASS — 2,210 checks and `validate_codemap.py`. |
| `make maintainability` | PASS after extracting the capture-only replay hook; file-size and extraction-coupling ratchets passed, dashboard 247/250 lines. |
| `git diff --check` | PASS. |
| `make client-unit` | PASS — completed on the final source, including the Combat VFX test. |
| `make bot-client SCENARIO=client_skill_points_and_magic_bolt HEADLESS=1` | PASS — 1/1. |
| `make ci` | PASS — combined batch gate, 11m41s on 2026-10-01. |
| `make ci-full` | NOT RUN — extended matrix was not requested or required by a concrete risk. |

## Security and data boundaries

The visual replay setting is opt-in and consumed only by the visual-replay path. Capture selection requires an allowlisted skill, the configured event and target, a successful unblocked hit, and a catalog mapping. Filenames pass the existing `BotFrameCapture.valid_name` check before the writer is called. This adds no remote input, command execution, authentication, persistence, or authorization path.

## Handoff manifest

v502-owned modified files:

- `client/scripts/combat_vfx.gd`
- `client/scripts/main.gd`
- `client/scripts/player_camera_controller.gd`
- `client/tests/test_combat_vfx.gd`
- `docs/CODEMAP.md` (shared index; preserve v501 entries)
- `shared/assets/vfx_presentation.v0.json`
- `shared/assets/vfx_presentation.v0.schema.json`
- `shared/rules/worlds.v0.json` (`skill_visual_lab` fixture position only)
- `tools/bot/scenarios/44_skill_visual.json`
- `tools/bot/skill_visual_runtime.py`
- `tools/bot/test_skill_visual.py`

v502-owned added files:

- `client/scripts/skill_visual_capture.gd`
- `client/scripts/skill_visual_capture_runtime.gd`
- `docs/as-built/v502_skill-specific-vfx.md`
- `docs/plans/v502_2026-10-01-skill-specific-vfx.md`
- `docs/specs/v502_spec-skill-specific-vfx.md`

No v502 files are deleted. No commits or pushes were made. The following coordinator-transferred v501 files are present and must not be treated as v502 changes or recopied over the integrated result:

- `client/scripts/gameplay_feedback_presentation.gd`
- `client/scripts/model_reaction_controller.gd`
- `client/scripts/monster_death_presentation_loader.gd`
- `client/tests/test_monster_death_presentation.gd`
- `docs/as-built/v501_monster-death-dissolve.md`
- `docs/plans/v501_2026-10-01-monster-death-dissolve.md`
- `docs/specs/v501_spec-monster-death-dissolve.md`
- `shared/assets/monster_death_presentation.v0.json`
- `shared/assets/monster_death_presentation.v0.schema.json`
- `tools/bot/scenarios/client/103_combat_input_flow_polish.json`
- The shared `docs/CODEMAP.md` v501 row and death-presentation entries.

The 18 tracked `.glb.import` sidecars regenerated by Godot were restored to the base revision before handoff. The ignored `.artifacts/` evidence to preserve includes the six `v502_skill_{magic_bolt,ice_shard,lightning}.{png,json}` files and `.artifacts/bot-runs/20261001T135134Z-visual.json` for the final Lightning run. `.venv/`, `.pytest_cache/`, `arpg_tools.egg-info/`, `client/.godot/`, and ignored Godot `*.gd.uid` sidecars are generated local state, not source changes.
