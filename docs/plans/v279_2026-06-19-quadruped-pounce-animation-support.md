# v279 Plan - Quadruped Pounce Animation Support

Status: Complete
Goal: Add a deterministic quadruped skin and pounce-ready bone clips to the existing quadruped monster visual.
Architecture: Keep the slice client/asset-only. Add a focused quadruped rig injector for the fox
source GLB, update the manifest and animation library, and prove the Godot scene has readable
bone-driven walk/attack/pounce clips.
Tech stack: Python asset tooling, GLB manifest validation, Godot scenes/animations, client tests, docs.

## Baseline and shortcut decision

Builds on v278. Adopt the existing quadruped predator source asset; borrow the deterministic rigging
tool style from v278; reject external rigs/plugins and server behavior changes.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `tools/assets/rig_quadruped_monster_glbs.py` | Append quadruped skin joints to the fox GLB |
| Add | `tools/assets/test_rig_quadruped_monster_glbs.py` | Prove quadruped rig joints are present |
| Modify | `make/shared.mk` | Include quadruped rigging in `make gen-assets` |
| Modify | `assets/manifests/assets.v0.json` | Declare quadruped skin joints and updated hash |
| Modify | `client/assets/monsters/purple_fantasy/evil_fox_monster.glb` | Rigged quadruped runtime |
| Modify | `client/scenes/monster_quadruped.tscn` | Target imported skeleton for animation root |
| Modify | `client/animations/monster_quadruped_fox_anims.tres` | Add bone-driven walk/attack/pounce clips |
| Modify | `client/tests/test_animation.gd` | Assert quadruped skeleton and clips |
| Modify | `docs/specs/v279_spec-quadruped-pounce-animation-support.md` | Mark complete |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row |
| Modify | `PROGRESS.md` | Update current status |
| Add | `docs/as-built/v279_quadruped-pounce-animation-support.md` | Record proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Decision:
- [x] Add focused new asset tooling/tests and keep touched existing files below target.

Verification:
```bash
make maintainability
```

Expected known issue: `make maintainability` is currently blocked by unrelated pre-existing ratchet
debt and will be paid down in the post-loop `$refactor`.

## Task 1 - Quadruped rigged runtime

Files:
- Create: `tools/assets/rig_quadruped_monster_glbs.py`
- Create: `tools/assets/test_rig_quadruped_monster_glbs.py`
- Modify: `make/shared.mk`
- Modify: `assets/manifests/assets.v0.json`
- Modify: `client/assets/monsters/purple_fantasy/evil_fox_monster.glb`

- [x] Add deterministic quadruped skin injection for the fox source GLB.
- [x] Regenerate the runtime GLB and update manifest hash/required joints.

```bash
.venv/bin/python -m pytest tools/assets/test_rig_quadruped_monster_glbs.py tools/assets/test_validate_assets.py -q
make validate-assets
```

## Task 2 - Pounce-ready client clips

Files:
- Modify: `client/scenes/monster_quadruped.tscn`
- Modify: `client/animations/monster_quadruped_fox_anims.tres`
- Modify: `client/tests/test_animation.gd`

- [x] Add bone-driven `walk`, `attack`, and `pounce` clips while preserving scene transforms.
- [x] Extend animation smoke coverage for quadruped rig, attack, and pounce.

```bash
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

## Task 3 - Lifecycle docs

Files:
- Modify: `docs/specs/v279_spec-quadruped-pounce-animation-support.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v279_quadruped-pounce-animation-support.md`

- [x] Mark the spec complete.
- [x] Record focused verification and batch-CI-pending status.
- [x] Add as-built notes and lifecycle row.

## Final verification

- [x] `.venv/bin/python -m pytest tools/assets/test_rig_quadruped_monster_glbs.py tools/assets/test_validate_assets.py -q`
- [x] `make validate-assets`
- [x] `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- [x] `GODOT=/opt/homebrew/bin/godot make client-unit`
- [x] `make maintainability` not re-run for v279; known unrelated ratchet debt deferred to `$refactor`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
