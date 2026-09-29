# v469 Plan — Art baseline and KayKit verification (ADR-0018 P0)

Status: Implemented 2026-09-28. See the [as-built](../as-built/v469_art-baseline-kit-verification.md).

**Execution deviations:**
- **No git commits.** The execute skill forbids commits unless the owner asks, so every per-task
  "Commit …" step is left for the owner.
- **`make` could not run.** It is blocked by the unaccepted Xcode license, so each target's
  underlying command ran directly. `go test` ran with `CGO_ENABLED=0`, because clang is also
  blocked.
- **`inspect-kit` target location.** It went into `make/shared.mk`, next to the other asset targets,
  instead of `make/tools.mk`.
- **Task 3 additions** (A2 allows local fixes):
  - `render_focus.py` now fails a capture on any GDScript `SCRIPT ERROR`.
  - The `town` and `stairs` focuses were repaired. They had been silently broken since `ba083f77`.
  - The `visual_capture.gd` baseline was lowered from 1237 to 1235.
- **Task 6 change.** The audit records generation failures instead of aborting, because 4 of 200
  seed/level pairs fail generation. A separate task was filed for that bug.
- **Task 7 sources.**
  - The three GitHub 1.0 packs were downloaded.
  - The itch.io Adventurers 2.0 and Character Animations 1.1 packs need the owner, because itch.io
    uses a "name your price" dialog.
  - Adventurers 1.0 embeds its clips, so the probe did not need the animation pack.

Goal: build the evidence ADR-0018 needs before any kit asset enters the runtime:
- a working screenshot harness and a committed "before" baseline
- a KayKit structure report
- a wall-grid alignment report
- resolved P0 checklist answers

Architecture:
- This slice is tooling plus docs only. It changes nothing in `client/assets/`,
  `assets/manifests/`, `shared/`, protocol, or server generation code.
- Screenshot fixes stay in the Python harness and a dedicated capture script. The grandfathered
  `visual_capture.gd` is not touched.
- GLB parsing is consolidated into one stdlib reader, used by both the validator and the new
  inspector.
- The wall-grid audit is a package-internal Go test, skipped unless an env var is set, because
  `generatedDungeonLevel` is unexported.

Tech stack: Python (tools, pytest), GDScript (capture script), Go (audit test), Godot 4 real-window
rendering for captures.

Spec: [`docs/specs/v469_spec-art-baseline-and-kit-verification.md`](../specs/v469_spec-art-baseline-and-kit-verification.md)
ADR: [`docs/adr/0018-art-direction-and-kit-based-visuals.md`](../adr/0018-art-direction-and-kit-based-visuals.md)

## Spec review (2026-09-28)

**Verified against code:**
- **Root cause of the empty screenshot directories:** `tools/test_regen_screenshots.py:62` runs a
  dry-run without `--out-dir`. `regen_screenshots.py` calls `out_dir.mkdir` and each job's
  `output.parent.mkdir` before checking `dry_run`.
- **The room capture uses its own lights.** `client/scripts/surface_material_room_capture.gd`
  (105 lines) builds `_make_key_light` / `_make_fill_light` instead of using the runtime lighting.
  - `DungeonDepthLighting.apply_for_level(level, directional, world_env, factory)` is static.
  - `DungeonTorchLights.new(parent, null, factory, renderer).sync(level, walls, true)` accepts a
    null fog overlay.
  - So both runtime builders can drive the capture directly.
- **Biome palettes** are a list with `min_depth` / `max_depth` (`shared/rules/dungeon_generation.v0.json`).
  Spec A4 was tightened to one capture per palette.
- **Wall shape:** `wallObstacle.pos` is the rectangle's center (`perimeterWalls`,
  `server/internal/game/dungeon_gen.go:198`), so edges are `pos ± size/2`.
  `GenerateDungeonLevel(seed, levelNum, rules)` returns the unexported `generatedDungeonLevel`,
  so the audit must be an in-package test.
- **Existing parser bug to fix during the move:** `validate_assets.py` has two ad-hoc GLB JSON
  parsers. `parse_glb_non_unit_node_scales` returns `["?", [0.0]]` on exception. That is a list of
  two unrelated items, not a list of `(name, scale)` tuples. It gets fixed when moved to the reader
  (Task 4).

**Gaps and drift that don't block this slice:**
- **Stale PROGRESS.md.** It still says "latest v467". v468 shipped in commits `98105010` through
  `7d792642`, but has no plan or as-built. This slice records v469 honestly and flags the missing
  v468 closeout. It does **not** invent a v468 as-built.
- **Godot version.** The local Godot is 4.7.2; PROGRESS pins 4.6.3. The spec now records this as a
  risk: record the version, don't re-pin.
- **No bot scenarios.** No gameplay, protocol, world, or replay changes, as stated in the spec.

## Baseline and shortcut decision

Reuses:
- `tools/showme/regen_screenshots.py` and `tools/showme/screenshot_catalog.py`
- `skills/showme/scripts/render_focus.py`
- `surface_material_room_capture.gd`, `GroundWallFactory`, `WallRenderer`,
  `DungeonSurfaceDetailPresentation`, `DungeonDepthLighting`, `DungeonTorchLights`
- the GLB parsing in `validate_assets.py` and `skills/3dmodel/scripts/create_model_probe.py`
- `LoadRules` plus `GenerateDungeonLevel`

**Asset/plugin decision:**

| Decision | What |
|----------|------|
| Adopt | The in-repo harness and runtime lighting builders listed above |
| Borrow | Existing GLB parsing, consolidated into `tools/assets/glb_reader.py` |
| Reject | New Python deps (`pygltflib`, `trimesh`), Godot addons, and any kit file under `client/` |

KayKit packs are staged only in gitignored `.artifacts/kaykit/`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/showme/regen_screenshots.py` | Dry-run writes nothing; mkdir only on real runs |
| Modify | `tools/test_regen_screenshots.py` | Dry-run uses `tmp_path`; asserts no writes; `scenes` suite coverage |
| Modify | `tools/showme/screenshot_catalog.py` | New `scenes` suite: showme scene focuses plus `dungeon-room` per biome palette |
| Modify | `skills/showme/scripts/render_focus.py` | New `dungeon-room` focus dispatched to the dedicated capture script with `--level` |
| Modify | `client/scripts/surface_material_room_capture.gd` | `--level` arg; runtime depth lighting and torches instead of bespoke lights |
| Modify | `skills/showme/SKILL.md` | Document the `scenes` suite and the `dungeon-room` focus |
| Create | `tools/assets/glb_reader.py` | Stdlib `.glb`/`.gltf` reader: JSON, buffers, accessors, node tree, skins, animations, images |
| Create | `tools/assets/test_glb_reader.py` | Direct unit tests on `gen_glb.py` fixtures plus a `.gltf`+`.bin` fixture |
| Modify | `tools/assets/validate_assets.py` | Use `glb_reader`; fix the malformed exception return |
| Create | `tools/assets/inspect_kit.py` | Kit report (JSON + Markdown) to `.artifacts/kaykit/report/` |
| Create | `tools/assets/test_inspect_kit.py` | Report keys and values on fixtures |
| Modify | `make/tools.mk` | `inspect-kit` target |
| Create | `server/internal/game/dungeon_wall_grid_audit_test.go` | Env-gated deterministic alignment report |
| Modify | `make/server.mk` | `wall-grid-audit` target |
| Create | `docs/as-built/assets/v469/*.png` | Curated "before" baseline (≤ 16, each ≤ 400 KB) plus the probe capture |
| Create | `docs/researchs/v469_kaykit-p0-findings.md` | P0 checklist answers with evidence |
| Modify | `docs/adr/0018-art-direction-and-kit-based-visuals.md` | Resolved checklist, measured values |
| Create | `docs/as-built/v469_art-baseline-kit-verification.md` | As-built with baseline links |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/CODEMAP.md` | Lifecycle, and new files indexed (`test_validate_codemap.py`) |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`: not touched
- [x] `server/internal/game/game_test.go`: not touched
- [x] `tools/bot/run.py`: not touched
- [x] `tools/validate_shared.py`: not touched
- [x] Other over-limit files from `.maintainability/file-size-baseline.tsv`: none. `visual_capture.gd`
  (1,237) is explicitly avoided.
- [x] Every new file ≤ 600 lines (`glb_reader.py` is the likeliest to grow; keep reporting in
  `inspect_kit.py`)

Decision:
- [x] Extract a focused module as part of this slice: `glb_reader.py`, replacing the two parsers in
  `validate_assets.py`. It is importable and tested directly, with no `globals()` laundering.

Verification:
```bash
make maintainability
```

## Task 1 — Dry-run writes nothing

Files: `tools/showme/regen_screenshots.py`, `tools/test_regen_screenshots.py`

- [x] 1.1: In `main()`, create `out_dir` only when `not args.dry_run`. In the job loop, call
  `output.parent.mkdir` only when `not args.dry_run`.
- [x] 1.2: Change `test_regen_screenshots_dry_run` to pass `--out-dir str(tmp_path / "shots")`.
  Assert that the directory does not exist afterward. Keep the existing `--focus` stdout assertions.
- [x] 1.3: Add a regression test: listing `.artifacts/screenshots/` before and after a dry-run with
  **no** `--out-dir` gives the same entries.
```bash
.venv/bin/pytest tools/test_regen_screenshots.py -v
```
- [x] 1.4: Commit `fix(showme): regen-screenshots dry-run no longer creates output dirs`.

The ~69 stale empty directories are local and gitignored, and are not touched by code. The executor
may delete **empty** `.artifacts/screenshots/2026*` directories after checking that each is empty
(`find … -type d -empty`).

## Task 2 — `scenes` suite and `dungeon-room` focus

Files: `client/scripts/surface_material_room_capture.gd`, `skills/showme/scripts/render_focus.py`,
`tools/showme/screenshot_catalog.py`, `tools/test_regen_screenshots.py`, `skills/showme/SKILL.md`

- [x] 2.1: Capture script:
  - Parse `--level <int>` (default `-1`).
  - Replace `_make_key_light` and `_make_fill_light` with a `DirectionalLight3D` and a
    `WorldEnvironment`, configured by `DungeonDepthLighting.apply_for_level(level, light, env, factory)`.
  - Add `DungeonTorchLights.new(world, null, factory, renderer).sync(level, walls, true)`.
  - Pass `level` to `renderer.set_level`, `factory.make_ground_node`, and the surface-detail sync.
  - Keep the camera and sample layout. The file stays well under 600 lines.
- [x] 2.2: In `render_focus.py`:
  - Add a `dungeon-room` choice and a `--level` argument.
  - When the focus is `dungeon-room`, launch `client/scripts/surface_material_room_capture.gd`
    instead of `visual_capture.gd`. Use the same godot flags, a real window, and forward `--output`
    and `--level`.
  - Default size 900×620 (the script's own size).
- [x] 2.3: In `screenshot_catalog.py`, add a `scenes` `SuiteSpec` with these jobs:
  - `town`, `monsters`, `chests`, `stairs`, `eye-view`, `heal-rain` (existing focuses, no extra args)
  - `dungeon-room` once per entry in `biome_palettes`, with `--level -<min_depth>` and slug
    `dungeon-room-<palette id>`
  - Palettes are read from `shared/rules/dungeon_generation.v0.json`. Nothing is hardcoded.
- [x] 2.4: Tests:
  - the `scenes` suite is in `DEFAULT_SUITES`
  - the number of `dungeon-room` jobs equals `len(biome_palettes)`
  - slugs are unique
  - a dry-run shows `--focus dungeon-room` and `--level`
- [x] 2.5: Smoke one capture in a real window, then look at the PNG. It must show floor, walls,
  torches, and biome-tinted lighting.
```bash
.venv/bin/pytest tools/test_regen_screenshots.py tools/test_showme.py -v
python3 skills/showme/scripts/render_focus.py --focus dungeon-room --level -1 --output .artifacts/showme/v469-dungeon-room.png
```
- [x] 2.6: Update the suites table in `skills/showme/SKILL.md`. Commit
  `feat(showme): scenes suite + runtime-lit dungeon-room capture`.

## Task 3 — Full real run and committed baseline

Files: `docs/as-built/assets/v469/*.png`

- [x] 3.1: Run the full batch and record wall-clock duration, Godot version, and machine:
```bash
time make regen-screenshots
godot --version
```
- [x] 3.2: Acceptance:
  - the run prints `N/N captures ok`
  - `index.json` has every job with `ok: true`
  - `.artifacts/screenshots/latest` resolves to the run
  - If a pre-existing focus fails, fix it when the fix is local and small. Otherwise record it in
    the as-built as a known failure with its stderr, and re-run with `SUITE=` for the rest. Never
    drop a job silently.
- [x] 3.3: Curate at most 16 PNGs into `docs/as-built/assets/v469/`, keeping the `index.json` slugs
  as filenames:
  - `skeleton` × 5 classes, or one `classes` shot if clearer
  - 2 `gear`
  - every `dungeon-room-*`
  - `town`, `monsters`, `eye-view`
- [x] 3.4: Enforce ≤ 400 KB each. If a file is larger, downscale with `sips -Z 900 <file>`
  (macOS built-in).
```bash
find docs/as-built/assets/v469 -name '*.png' -size +400k
```
- [x] 3.5: Commit `docs(v469): before-baseline screenshots for ADR-0018`.

## Task 4 — `glb_reader.py` extraction

Files: `tools/assets/glb_reader.py`, `tools/assets/test_glb_reader.py`, `tools/assets/validate_assets.py`

- [x] 4.1: Write `glb_reader.py` (stdlib only). `load_gltf(path) -> Gltf` handles:
  - `.glb`: JSON chunk plus BIN chunk
  - `.gltf`: JSON plus `uri` buffers, relative files or base64 `data:` URIs

  Helpers:
  - `accessor_count`
  - `accessor_min_max`
  - `mesh_triangle_count` (indices count / 3, or POSITION count / 3; mode 4 only, other modes
    reported)
  - `node_tree` (roots, children, names)
  - `skin_joint_names(skin_index)` (ordered)
  - `animation_summaries` (name, duration from input accessor max, target node names, paths)
  - `image_dimensions` (PNG IHDR, JPEG SOF0/SOF2 from bufferView or file)
  - `skin_joint_name_set()`
  - `non_unit_node_scales(tolerance)`
- [x] 4.2: Tests import `glb_reader` directly:
  - `gen_glb.monster_skeleton_glb()` (skinned, 7 joints) written to `tmp_path`
  - `monster_dummy_glb()`: triangle count equals the cube geometry
  - a hand-built minimal `.gltf` + `.bin` with a tiny embedded PNG, for URI buffers and image
    dimensions
- [x] 4.3: `validate_assets.py`:
  - Replace `parse_glb_skin_joint_names` and `parse_glb_non_unit_node_scales` bodies with reader
    calls. Keep the function names and signatures so callers are unchanged.
  - On exception, return `None` or `[("?", [0.0])]` (the malformed return is fixed).
```bash
.venv/bin/pytest tools/assets/test_glb_reader.py tools/assets/test_validate_assets.py -v
make validate-assets
```
- [x] 4.4: Commit `refactor(assets): shared stdlib glb_reader; validate_assets uses it`.

## Task 5 — `inspect_kit.py`

Files: `tools/assets/inspect_kit.py`, `tools/assets/test_inspect_kit.py`, `make/tools.mk`

- [x] 5.1: CLI:
  `python -m tools.assets.inspect_kit <path...> [--out .artifacts/kaykit/report]`
  - Recurses into directories for `*.glb` and `*.gltf`.
  - Lists `.fbx` / `.obj` files it finds as `unsupported` (so they are reported as gaps).
- [x] 5.2: Each file's report entry contains:
  - relative path and sha256
  - node tree
  - `meshes: [{name, node_names, primitives, triangles, materials}]`
  - `total_triangles`
  - `materials`
  - `images: [{name/uri, width, height, mime}]`
  - `skins: [{name, joints:[…]}]`
  - `animations: [{name, duration_s, targets:[…]}]`
  - `extents: {min, max, size}` (POSITION accessor bounds with node transforms applied)
- [x] 5.3: Outputs:
  - `report.json`
  - `report.md`, with a table per directory plus a **rig comparison** section: the joint-name
    sets of all skinned files, grouped by identical set, with a diff between groups
- [x] 5.4: Tests: run on the Task 4 fixtures and assert the keys, triangle counts, joint list, and
  rig-group count. `make inspect-kit KIT=<path>` wraps the CLI.
```bash
.venv/bin/pytest tools/assets/test_inspect_kit.py -v
```
- [x] 5.5: Commit `feat(assets): read-only kit inspector for ADR-0018 P0`.

## Task 6 — Wall-grid audit

Files: `server/internal/game/dungeon_wall_grid_audit_test.go`, `make/server.mk`

- [x] 6.1: `TestDungeonWallGridAudit`:
  - Call `t.Skip` unless `ARPG_WALL_GRID_AUDIT_OUT` is set.
  - Load `LoadRules(filepath.Join("..","..","..","shared","rules"))`.
  - **Seeds:** a fixed slice of 20 literal strings.
  - **Levels:** `-1` through `-maxDepth`. `maxDepth` is the larger of:
    - the largest non-null biome `max_depth`
    - the largest floor-profile `min_depth`
    - `-(BossFloor.FirstLevel) + BossFloor.Cadence`

    This covers every biome, every profile, and at least two boss floors.
  - For each wall, compute edges `pos.X ± size.X/2` and `pos.Y ± size.Y/2`, broken down by
    `source` and `kind`.
  - **Candidate tile sizes:** `[0.5, 1.0, 2.0]`, plus optional extra sizes from
    `ARPG_WALL_GRID_EXTRA_TILES` (comma-separated), so Task 8 can feed the measured kit tile size.
  - For each candidate, report the aligned fraction (epsilon `1e-6`) overall and per source.
  - Also report a histogram of `edge mod 1.0` remainders, rounded to 0.05, as a sorted slice.
  - Write sorted, indented JSON to the env path.
- [x] 6.2: Determinism:
  - Iterate slices only. Where a map is used for aggregation, emit it through sorted keys.
  - No `time.Now()`. The file is a `_test.go`, but it still passes `make lint-determinism`.
- [x] 6.3: Add a `wall-grid-audit` make target that runs:
  `ARPG_WALL_GRID_AUDIT_OUT=$(ROOT)/.artifacts/wall-grid-audit.json go test ./internal/game/ -run TestDungeonWallGridAudit -count=1`
- [x] 6.4: Verify:
  - running it twice gives an identical file
  - the default `make test-go` shows the test as skipped
```bash
make wall-grid-audit && cp .artifacts/wall-grid-audit.json /tmp/a.json && make wall-grid-audit && cmp /tmp/a.json .artifacts/wall-grid-audit.json
cd server && go test ./internal/game/ -run TestDungeonWallGridAudit -v | grep -i skip
make lint-determinism
```
- [x] 6.5: Commit `test(game): env-gated dungeon wall grid alignment audit`.

## Task 7 — Stage KayKit packs (no repo changes)

- [x] 7.1: Find each pack's **official** distribution page (the KayKit / Kay Lousberg site, its
  itch.io page, or its official GitHub org): Adventurers, Character Animations, Skeletons, Dungeon
  Remastered. Optionally one secondary VFX sprite pack (for example Kenney Particle Pack) for P5.
- [x] 7.2: Before downloading, state each filename, source URL, and size in chat (ADR-0018 D2).
- [x] 7.3: Stop and hand off to the owner if a page requires a login, an account, a checkout
  (including "$0 / name your price"), or a CAPTCHA. The owner drops those archives into
  `.artifacts/kaykit/`.
- [x] 7.4: Extract to `.artifacts/kaykit/<pack>/`. Record the archive name, version, source URL,
  sha256, and the path of the license file inside the archive.
- [x] 7.5: Run the inspector. Re-run the wall audit with the measured kit floor/wall module size.
  The size comes from `extents` of the floor tile and straight wall pieces.
```bash
shasum -a 256 .artifacts/kaykit/*.zip
make inspect-kit KIT=.artifacts/kaykit
ARPG_WALL_GRID_EXTRA_TILES=<kit_tile> make wall-grid-audit
```

## Task 8 — Disposable Godot rig probe

- [x] 8.1: Create `.artifacts/kaykit/probe/project.godot`: a minimal Forward+ project, never
  committed. Copy in one Adventurer character (the paladin candidate), the animation library file
  that matches its rig, and one sword.
- [x] 8.2: Write a probe `SceneTree` script that:
  - instances the character
  - finds its `Skeleton3D` and the `AnimationPlayer` (or builds one from the kit library)
  - plays a walk or attack clip and advances ~0.4 s
  - attaches the sword with a `BoneAttachment3D` on the hand bone chosen in the findings
  - adds a camera and a light, then saves a PNG in a real window
- [x] 8.3: If the clip doesn't bind (the rig doesn't match), still save the capture, record the
  error, and write that result into the findings. Don't work around it.
- [x] 8.4: Copy the capture to `docs/as-built/assets/v469/kaykit-rig-probe.png` (≤ 400 KB).
```bash
godot --path .artifacts/kaykit/probe --import
godot --path .artifacts/kaykit/probe --script res://probe.gd
```

## Task 9 — Findings and ADR update

Files: `docs/researchs/v469_kaykit-p0-findings.md`, `docs/adr/0018-art-direction-and-kit-based-visuals.md`

- [x] 9.1: Findings doc, one section per checklist item 1–6, in the format spec D1 requires. Each
  claim cites `report.json`, `report.md`, `wall-grid-audit.json` excerpts, or the probe capture.
  The doc must include:
  - the class mapping table
  - the `item_visuals` family gap list (derived from `shared/assets/item_visuals.v0.json` slots:
    28 main_hand, 5 off_hand, 5 head)
  - the socket and camera-anchor bone choices
  - the clip-to-state map with gaps
  - the per-character mesh-part list and the chosen tint mechanism
  - the region reach table
  - the kit tile size and the D6 branch outcome
  - the measured budget table
- [x] 9.2: ADR-0018:
  - mark checklist items resolved with links
  - replace D1 mapping, D5 mechanism and granularity, D6 branch, and D8 table values with the
    measured ones
  - update the Context screenshot-harness bullet to the post-repair state
  - if check 3 fails (rig mismatch), change the status to note that P3's premise needs revision.
    Do not paper over it.
- [x] 9.3: Commit `docs(v469): ADR-0018 P0 findings and resolved decisions`.

## Task 10 — Lifecycle docs and CI

Files: `docs/as-built/v469_art-baseline-kit-verification.md`, `PROGRESS.md`,
`docs/progress/slice-lifecycle.md`, `docs/CODEMAP.md`

- [x] 10.1: The as-built covers:
  - what was proven
  - the full-run duration and Godot version
  - known capture failures, if any
  - a linked baseline gallery
  - the findings summary
  - the next phase (P1, or the P2 server snap slice if D6 branched)
- [x] 10.2: `PROGRESS.md`:
  - latest v469
  - next: ADR-0018 P1 render baseline
  - an open-gap row noting that the v468 plan and as-built are missing (not fabricated)
  - the Godot 4.6.3 vs 4.7.2 drift
- [x] 10.3: `slice-lifecycle.md` gets a v469 row. `CODEMAP.md` lists the new tools, capture
  changes, and the audit test.
```bash
.venv/bin/pytest tools/test_validate_codemap.py -v
make ci
```
- [x] 10.4: Commit `docs(v469): as-built, progress, codemap`.

## Final verification

- [x] `make maintainability` (record the grandfathered count and line-total trend in the as-built)
- [x] `make validate-assets`
- [x] `make test-py` and `make test-go` (the audit test shows as skipped)
- [x] `make lint-determinism`
- [x] `make ci`
- [x] `git status` shows nothing under `client/assets/`, `assets/manifests/` or `shared/`

## Deferred (explicitly out of this slice)

- Kit import, manifest entries, budget data file and validator enforcement: first importing slice
  (P2 or P3).
- Render baseline: P1.
- Dungeon auto-tiler and any server snap rule: P2, or a dedicated server slice.
- Rig swap, clip catalog, socket remap and armor tints: P3.
- Monster replacement and unconfirmed or unused GLB purge: P4.
- VFX: P5.
- The v468 plan/as-built closeout (flagged, not done here).
- Re-pinning Godot.
