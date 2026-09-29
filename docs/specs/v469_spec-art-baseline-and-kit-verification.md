# v469 — Art baseline and KayKit verification (ADR-0018 P0)

- **Status:** Draft
- **Date:** 2026-09-28
- **Codename:** `art-baseline-kit-verification`
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md), phase P0
- **Baseline:** v468 real-body first-person view (`98105010`) plus follow-up fixes through `7d792642`

---

## Purpose

ADR-0018 moves the game's visuals to the KayKit family. Before any kit asset enters the runtime,
this slice produces the evidence the later phases depend on. It has four deliverables:

1. **A working visual-regression harness, plus a committed "before" baseline.**
   - **Root cause of the empty screenshot folders:** `tools/test_regen_screenshots.py:62` runs
     `regen_screenshots --dry-run` without `--out-dir`. `regen_screenshots.py` creates the
     timestamped directory and each job's parent folder *before* it checks `dry_run`. Every
     `pytest tools/` run therefore leaves an empty `skeleton/` + `gear/` tree under
     `.artifacts/screenshots/`.
   - There is no evidence the batch has ever completed a real run: no `index.json`, no `latest`
     symlink.
   - The existing suites cover only classes, gear, icons and item assets. None shows a dungeon
     room, town, monsters or lighting, which are exactly what P1 and P2 change.
2. **A read-only kit inspector.** A stdlib Python tool that reports the structure of KayKit
   `.glb`/`.gltf` files: meshes, skins and joints, animations, triangle counts, textures and extents.
3. **A wall-grid audit.** A deterministic, report-only Go test. It measures whether
   server-generated wall rectangles snap to a common tile size, which decides ADR-0018 D6.
4. **A findings document** that answers every item on the ADR-0018 P0 checklist with evidence.
   ADR-0018 is then updated with the resolved decisions and the final D8 budgets.

Nothing a player sees changes in this slice.

## Non-goals

- No KayKit file under `client/`. No manifest, `item_visuals`, `gear_sockets` or catalog changes.
  Packs are staged only in gitignored `.artifacts/kaykit/`.
- No renderer, lighting, environment, material or VFX changes (P1, P5).
- No dungeon renderer changes (P2) and no server generation or snap-rule change. The audit only
  reports; if a snap rule is needed, it is its own spec'd slice.
- No rig, animation, socket or tint implementation (P3). No monster replacement or deletion of
  unconfirmed or unused GLBs (P4).
- No budget enforcement in `validate_assets.py`. P0 measures budgets and records them in ADR-0018.
  The first slice that imports kit files adds the budget data file and enforcement.
- No new Python or Godot dependencies (see the asset/plugin decision below).

## Acceptance criteria

### A. Harness repair and scene suite

- **A1.** `regen_screenshots --dry-run` creates no files or directories. The dry-run test passes an
  explicit `--out-dir` under `tmp_path` and asserts that the directory is still empty or absent
  afterward.
- **A2.** A real `make regen-screenshots` run completes for every suite. It prints `N/N captures ok`,
  writes `index.json` with `ok: true` for every job, and points `.artifacts/screenshots/latest` at
  the run.
  - If some existing focus is broken, fix it or record it as an explicit known failure in the
    as-built with its error.
  - Silently skipping a job is not allowed.
- **A3.** A new `scenes` suite in `tools/showme/screenshot_catalog.py` captures these existing
  showme focuses: `town`, `monsters`, `chests`, `stairs`, `eye-view`, `heal-rain`.
- **A4.** A new `dungeon-room` focus captures a representative dungeon room under the **real**
  runtime dungeon lighting. That means the biome palette from `dungeon_depth_lighting.gd` and torch
  lights from `dungeon_torch_lights.gd`, not the capture script's own lights. It also shows the
  current wall, floor, column, water and overlay presentation.
  - It reuses or extends `client/scripts/surface_material_room_capture.gd`.
  - It is dispatched from `render_focus.py` to that dedicated script. It is **not** added to the
    grandfathered `visual_capture.gd`.
  - The `scenes` suite includes one capture per `biome_palettes` entry in
    `shared/rules/dungeon_generation.v0.json`, at level `-min_depth`. Depths are discovered from the
    data, not hardcoded.
- **A5.** A curated "before" baseline is committed under `docs/as-built/assets/v469/`.
  - At most 16 PNGs, each ≤ 400 KB.
  - Coverage: at least one of each class, gear, dungeon-room (per captured biome), town, monsters,
    eye-view.
  - Filenames match their `index.json` slugs.
  - The v469 as-built links them as the ADR-0018 D9 "before" reference.

### B. Kit staging and inspector

- **B1.** The agent downloads the packs under the ADR-0018 D2 sourcing rules. Before the batch
  starts, it states each filename, source and size in chat.
  - Packs: Adventurers, Character Animations, Skeletons, Dungeon Remastered.
  - Optional: one candidate secondary VFX sprite pack (for example Kenney Particle Pack), inspected
    for P5.
  - Source: the official KayKit distribution. The exact URL is recorded at download time.
  - Packs are extracted to `.artifacts/kaykit/<pack>/`.
  - The findings doc records each archive's name, pack version, source URL and sha256.
- **B2.** `tools/assets/inspect_kit.py <path...>` accepts `.glb` and `.gltf` files and directories.
  It writes JSON (and a Markdown summary) to `.artifacts/kaykit/report/`. For each file it reports:
  - the node tree, including mesh node names and parent/child structure
  - per-mesh primitive count and triangle count
  - material names
  - embedded or referenced image dimensions
  - skins with ordered joint names
  - animations: name, duration, and which joints each targets
  - accessor-derived bounding extents
- **B3.** GLB/glTF parsing is shared, not copied a third time. `validate_assets.py` and
  `skills/3dmodel/scripts/create_model_probe.py` both have ad-hoc GLB JSON parsing today.
  - The inspector imports a small reader module, `tools/assets/glb_reader.py`.
  - `validate_assets.py` switches to that reader in the same slice.
  - Under the extraction-independence rule, the reader must be importable and tested directly.

### C. Wall-grid audit

- **C1.** A Go test in `server/internal/game/` (for example `dungeon_wall_grid_audit_test.go`) is
  skipped unless `ARPG_WALL_GRID_AUDIT_OUT` is set.
  - It generates levels across a fixed seed list and a depth range taken from rules: normal floors,
    boss floors, and every floor profile.
  - It writes a JSON report. For each candidate tile size, the report gives the fraction of wall
    rectangle edges (position ± size/2) that land on the grid, within a small epsilon.
  - Candidates: 0.5, 1.0, 2.0, and the kit tile size measured in B2 once known.
  - The report also includes a histogram of fractional remainders.
- **C2.** A make target (for example `make wall-grid-audit`) runs it and writes to
  `.artifacts/wall-grid-audit.json`.
- **C3.** The audit is deterministic (same seeds give the same report) and uses only the generator's
  public entry path. It passes `make lint-determinism`. It changes no generation code and no golden.

### D. Findings and ADR update

- **D1.** `docs/researchs/v469_kaykit-p0-findings.md` answers every ADR-0018 P0 checklist item with
  evidence taken from the B2 and C reports:

| # | Question | Pass condition / required output |
|---|----------|----------------------------------|
| 1 | License | Each pack's license text quoted by reference (file path in the archive) and confirmed CC0; versions and sha256 recorded |
| 2 | Contents & class mapping | Table: 5 classes → kit character file. Table: `item_visuals` families (main_hand types, off_hand, head) → kit mesh, with an explicit **gap list** of families that have no kit equivalent |
| 3 | Shared rig | Character joint-name sets are compared against the animation pack's rig (set equality, or an explicit diff). Kit bones are chosen for `right_hand_socket`, `off_hand_socket`, `head_socket`, `chest_socket`, and the `chest_view` first-person camera anchor. Kit clip names are mapped to logical states: idle, walk, attack, attack_off_hand, attack_2h, attack_ranged, attack_staff, hit, death, with a gap list |
| 4 | Mesh split & tint mechanism | Per character, the mesh-part list. Decision between ADR-0018 D5 mechanism 1 (per-part material override) and 2 (region-mask shader). Region table for chest / gloves / boots / belt stating what each tint can actually reach |
| 5 | Tile size vs walls | Measured kit floor/wall module extents. Audit alignment fractions. **Branch outcome:** client-only P2, or a server snap-rule slice first, with the reason |
| 6 | Budgets | Measured max triangles and texture sizes per D8 category. Final budget table |

- **D2.** One disposable Godot probe project under `.artifacts/kaykit/probe/` (never committed)
  loads one kit character, plays one kit animation clip, and attaches one kit weapon to the chosen
  hand bone. The findings doc embeds or links a screenshot. This tests the riskiest P3 assumption
  early. It does not modify `client/`.
- **D3.** ADR-0018 is updated:
  - Every checklist item is marked resolved, with a link to the findings.
  - D1's class mapping, D5's tint mechanism and granularity, D6's branch outcome and D8's budget
    table are replaced with the measured values.
  - The Context description of the screenshot harness is updated to post-repair state.

## Scope and files likely touched

| Area | Files |
|------|-------|
| Screenshot harness | `tools/showme/regen_screenshots.py`, `tools/showme/screenshot_catalog.py`, `tools/test_regen_screenshots.py`, `skills/showme/scripts/render_focus.py`, `skills/showme/SKILL.md` |
| Dungeon-room capture | `client/scripts/surface_material_room_capture.gd` (extend, or split into a focused capture script); a registered unit test if the node-render component-test pattern applies |
| Kit tooling | new `tools/assets/glb_reader.py`, new `tools/assets/inspect_kit.py`, new `tools/assets/test_glb_reader.py`, new `tools/assets/test_inspect_kit.py`, `tools/assets/validate_assets.py` (switch to the reader) |
| Server audit | new `server/internal/game/dungeon_wall_grid_audit_test.go`; make target in `make/*.mk` |
| Docs | `docs/researchs/v469_kaykit-p0-findings.md`, `docs/adr/0018-art-direction-and-kit-based-visuals.md`, `docs/as-built/v469_art-baseline-kit-verification.md` (+ `assets/v469/*.png`), `docs/plans/v469_*`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/CODEMAP.md` |
| Not touched | `client/assets/`, `assets/manifests/`, `shared/` (all), protocol, server generation code, goldens |

All new files must stay ≤ 600 lines. `visual_capture.gd` (baseline 1,237) must not grow.

### Client asset/plugin decision

- **Adopt:**
  - existing showme focuses, `render_focus.py` and `screenshot_catalog.py`
  - `surface_material_room_capture.gd` as the dungeon-room base
  - the runtime `dungeon_depth_lighting.gd` / `dungeon_torch_lights.gd` builders for capture
    lighting
- **Borrow:** the GLB chunk parsing already in `validate_assets.py` and `create_model_probe.py`,
  consolidated into `glb_reader.py`.
- **Reject:**
  - new Python dependencies such as `pygltflib` or `trimesh` (stdlib JSON/struct is enough to read
    structure)
  - Godot addons
  - importing kit files into `client/` in this slice (ADR-0018 gates that on these findings)

## Test and bot proof

- **Python:**
  - dry-run writes nothing (A1)
  - `scenes` suite discovery and counts match `suite_summary` (A3)
  - `glb_reader` parses a deterministic fixture GLB from `gen_glb.py`: nodes, a skin with named
    joints, triangle count, extents
  - `inspect_kit` produces the expected report keys for that fixture and for a `.gltf` + `.bin`
    fixture
  - `validate_assets` behavior is unchanged after moving to the reader (the existing
    `test_validate_assets.py` stays green)
- **Go:**
  - the audit test is skipped by default, so `make test-go` is unaffected
  - running it twice with the env var set produces byte-identical reports
- **Visual:** the real `make regen-screenshots` run (A2) and the committed baseline (A5) are the
  proof. Headless CI cannot render Forward+, so the as-built records the run command, machine and
  Godot version.
- **Gates:** `make ci` green, `make maintainability` (record the grandfathered count and line-total
  trend), `make validate-assets`, `make lint-determinism`.
- **Bot scenarios:** none. There is no gameplay or protocol change.

## Open questions and risks

- **Rig compatibility is the biggest unknown.** Older KayKit Adventurers releases may predate the
  rig used by the Character Animations pack.
  - If check 3 finds mismatched joint sets, P3's premise changes: a newer Adventurers release, a
    retarget step, or a different character source.
  - The findings doc must say this plainly. It must not paper over it.
- **Pack formats.** Packs may ship `.gltf` + `.bin` or FBX alongside or instead of `.glb`. The
  inspector supports `.gltf`. FBX is out of scope, and FBX-only content is recorded as a gap.
- **Screenshot run time and flakiness.** The full batch is 200+ captures in a real window (the macOS
  headless renderer produces no pixels). The plan should allow `SUITE=` subsets for iteration. It
  should also record the full-run duration so later slices know what a baseline refresh costs.
- **Godot version drift.** PROGRESS.md pins Godot 4.6.3, but the local `godot` is 4.7.2. Captures
  and the probe record the exact version used. P0 does not re-pin; it only reports the drift.
- **Baseline PNGs in git** (owner approved committing them, 2026-09-28). Committing about 16 small PNGs is a deliberate exception to "artifacts
  are gitignored". It exists so before/after comparisons survive across machines and sessions. If
  the owner prefers not to commit images, A5 falls back to documenting the capture command and
  commit hash only, and the before/after gate becomes weaker.
- **Official channel.** The agent downloads (owner authorization, ADR-0018 D2). If a pack's
  official page requires a login, a checkout flow (including "$0 / name your price"), or a CAPTCHA,
  the agent stops. The owner then downloads that archive into `.artifacts/kaykit/` by hand. The
  agent never creates accounts or completes checkout on the owner's behalf.
