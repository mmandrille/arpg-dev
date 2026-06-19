# v278 Plan - Biped Monster Rigged Animations

Status: Complete
Goal: Add deterministic skinned rigs and bone-driven walk/attack clips to the melee and ranged biped monster visuals.
Architecture: Keep the slice presentation-only. Reuse the existing local static GLBs, append simple
rigid biped skin joints through deterministic Python tooling, then update Godot animation libraries
and tests to prove bones move without changing server combat or protocol behavior.
Tech stack: Python asset tooling, GLB manifest validation, Godot scenes/animations, client tests, docs.

## Baseline and shortcut decision

Builds on v277. Adopt the existing committed dark-purple biped and crocodile archer GLBs; borrow the
v275 rigging approach; reject external assets/plugins and Blender-dependent work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `tools/assets/rig_monster_glbs.py` | Append biped monster skin joints to static source GLBs |
| Add | `tools/assets/test_rig_monster_glbs.py` | Prove rigged monster GLBs expose expected joints |
| Modify | `assets/manifests/assets.v0.json` | Declare required biped skin joints and updated hashes |
| Modify | `client/assets/monsters/purple_fantasy/dark_purple_monster.glb` | Rigged melee biped runtime |
| Modify | `client/assets/monsters/archer/crocodile_archer.glb` | Rigged ranged biped runtime |
| Modify | `client/animations/monster_dark_purple_anims.tres` | Add bone-driven attack/walk clips |
| Modify | `client/animations/monster_crocodile_archer_anims.tres` | Add bone-driven attack/walk clips |
| Modify | `client/tests/test_animation.gd` | Assert biped skeletons and bone animation |
| Modify | `docs/specs/v278_spec-biped-monster-rigged-animations.md` | Mark complete once verified |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row |
| Modify | `PROGRESS.md` | Update current status after slice |
| Add | `docs/as-built/v278_biped-monster-rigged-animations.md` | Record shipped behavior and proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/tests/test_animation.gd` is under target; keep additions surgical and do not grow broad
  coordinator behavior.
- [x] Other grandfathered files: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Add focused asset tooling and tests rather than growing existing generators.

Verification:
```bash
make maintainability
```

Result: attempted during close-out. It failed on pre-existing unrelated files not touched by this
slice (`client/scripts/main.gd`, `client/tests/test_item_visuals.gd`, and
`skills/showme/scripts/visual_capture.gd`). Per the autoloop continuation decision, that ratchet
debt is deferred to the post-loop `$refactor` pass.

## Task 1 - Rigged monster asset tooling

Files:
- Create: `tools/assets/rig_monster_glbs.py`
- Create: `tools/assets/test_rig_monster_glbs.py`
- Modify: `client/assets/monsters/purple_fantasy/dark_purple_monster.glb`
- Modify: `client/assets/monsters/archer/crocodile_archer.glb`
- Modify: `assets/manifests/assets.v0.json`

- [x] Add deterministic biped-monster rigging for the two selected monster assets.
- [x] Regenerate the two runtime GLBs from committed source assets.
- [x] Update manifest hashes and required skin-joint declarations.

```bash
.venv/bin/python -m pytest tools/assets/test_rig_monster_glbs.py tools/assets/test_validate_assets.py -q
make validate-assets
```

## Task 2 - Bone-driven biped clips

Files:
- Modify: `client/animations/monster_dark_purple_anims.tres`
- Modify: `client/animations/monster_crocodile_archer_anims.tres`
- Modify: `client/tests/test_animation.gd`

- [x] Add `attack` to the melee biped animation library.
- [x] Update biped `walk` and `attack` clips to rotate leg and arm bones.
- [x] Preserve existing scene scale/yaw corrections and the crocodile archer marker.
- [x] Extend the animation smoke test to verify biped skeletons and bone movement.

```bash
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

## Task 3 - Lifecycle docs

Files:
- Modify: `docs/specs/v278_spec-biped-monster-rigged-animations.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v278_biped-monster-rigged-animations.md`

- [x] Mark the spec complete.
- [x] Record focused verification and batch-CI-pending status.
- [x] Add as-built notes and lifecycle row.

```bash
rg -n "v278|biped-monster-rigged-animations|Latest completed slice" PROGRESS.md docs/progress/slice-lifecycle.md docs/as-built/v278_biped-monster-rigged-animations.md
```

## Final verification

- [x] `.venv/bin/python -m pytest tools/assets/test_rig_monster_glbs.py tools/assets/test_validate_assets.py -q`
- [x] `make validate-assets`
- [x] `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- [x] `GODOT=/opt/homebrew/bin/godot make client-unit`
- [x] `make maintainability` attempted; blocked by pre-existing unrelated ratchet debt deferred to `$refactor`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
