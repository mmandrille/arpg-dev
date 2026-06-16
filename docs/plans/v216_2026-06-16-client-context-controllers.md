# v216 Plan — Client Context Controllers

Status: Ready for implementation
Goal: Prove the typed context-object extraction pattern by moving town node construction and boss visual coordination out of `main.gd`.
Architecture: This is a client-only maintainability slice that builds on v215 helper extraction. Stateless town mesh construction moves to a static factory. Boss visual state moves to a controller with a plain typed context object and narrow callables for tint/status refreshes that remain owned by `main.gd`.
Tech stack: Godot client, maintainability ratchet, headless Godot tests.

## Baseline and shortcut decision

Builds on v215 after `client-pure-factories` lands. Client assets/plugins decision: **adopt existing in-repo primitive meshes and `ChestPresentationScript`; reject external assets/plugins** because the slice preserves existing rendering.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/town_node_factory.gd` | Pure/static constructors for town, stair, teleporter, chest, merchant, and market board nodes. |
| Add | `client/scripts/boss_visuals_context.gd` | Plain data carrier for boss visual dependencies. |
| Add | `client/scripts/boss_visuals_controller.gd` | Boss health bar, phase, telegraph, and boss query coordination. |
| Modify | `client/scripts/main.gd` | Delegate town constructors and boss visual methods; remove extracted code. |
| Modify | `client/tests/test_factories.gd` | Direct preload tests for context/controller/factory. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `main.gd` baseline to actual post-extraction count. |
| Add | `docs/as-built/v216_client-context-controllers.md` | Record shipped proof and verification. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle consolidation after verification. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Extract town node factory

Files:
- Create: `client/scripts/town_node_factory.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Move pure `_make_*` town/interactable constructors and material helpers into `TownNodeFactory`.
- [x] Step 1.2: Replace `main.gd` call sites with static factory calls and keep hero corpse construction in `main.gd`.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 2 — Extract boss visual context and controller

Files:
- Create: `client/scripts/boss_visuals_context.gd`
- Create: `client/scripts/boss_visuals_controller.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Add `BossVisualsContext` with only live references needed by boss visuals.
- [x] Step 2.2: Move boss bar, phase display, telegraph marker, and boss query helpers into `BossVisualsController`.
- [x] Step 2.3: Keep model tint/status refresh implementation in `main.gd` behind explicit callables.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 3 — Client regression proof

Files:
- Modify: `client/tests/test_factories.gd`

- [x] Step 3.1: Extend direct preload tests for `TownNodeFactory`, `BossVisualsContext`, and `BossVisualsController`.
- [x] Step 3.2: Run existing client smoke proof.
```bash
make client-smoke
```

## Task 4 — Lifecycle docs and final verification

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Add: `docs/as-built/v216_client-context-controllers.md`
- Modify: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`

- [x] Step 4.1: Lower the maintainability baseline for `client/scripts/main.gd`.
- [x] Step 4.2: Record as-built proof and lifecycle status.
- [x] Step 4.3: Run final standalone gates.
```bash
make maintainability
make client-unit
make client-smoke
make ci
```

## Final verification

- [x] `godot --headless --path client --script res://tests/test_factories.gd`
- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make ci`

