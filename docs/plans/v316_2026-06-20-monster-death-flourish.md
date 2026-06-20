# v316 Plan - Monster Death Flourish

Status: Complete
Goal: Add an assertable code-native death flourish to terminal model reactions.
Architecture: Keep the flourish inside `ModelReactionController.enter_death()`, the existing
client-only death reaction path. Server event authority and entity lifecycle remain unchanged.
Tech stack: Godot 4 GDScript presentation helper and existing animation unit test.

## Baseline and Shortcut Decision

Builds on v315. Asset/plugin decision: adopt generated mesh/material reaction effects; reject
external VFX assets, particle plugins, and audio changes.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/model_reaction_controller.gd` | Add death flourish child node |
| Modify | `client/tests/test_animation.gd` | Assert death flourish in terminal reaction test |
| Create | `docs/as-built/v316_monster-death-flourish.md` | Completion proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] None

Decision:
- [x] Keep changes in the existing focused reaction helper; no large coordinator touched.

Verification:
```bash
make maintainability
```

## Tasks

- [x] Add death flourish mesh children and metadata.
- [x] Assert flourish creation in existing animation/reaction test.
- [x] Update lifecycle docs and as-built proof.

## Verification

- [x] `godot --headless --path client --script res://tests/test_animation.gd`
- [x] `make client-unit`
- [x] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
