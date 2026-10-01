# v506 As-built — Living town

- **Base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Status:** Integrated; focused checks and combined batch `make ci` passed (11m41s, 2026-10-01). Resident-memory sampling remains unverified.
- **Scope:** Two presentation-only residents in level-0 town dressing. No server, protocol, world-preset, interaction, playable-class, or asset-manifest changes.

## Implementation

`shared/assets/town_presentation.v0.json` owns an idle ranger at `(2.5, 16.0)` and a rogue patrol along `(19.0, 4.5) → (21.5, 5.0) → (21.5, 8.5) → (19.0, 8.5)`. The schema bounds actor count and validates activity-specific fields. Clearance, route, facing, speed, pause and phase remain data-owned. `TownDressing.build()` adds the new `TownAmbientLife` layer to the shared live and preview path; it uses existing registered KayKit character assets and the cached `kaykit_hero` `Idle_A` / `Walking_A` clips. The patrol advances from a deterministic town-local clock and changes locomotion only at path/pause transitions.

The residents reuse the existing character animation-player/model hierarchy without equipment hooks. They have no collision/trigger nodes, gameplay groups, interactable identity, server state, or navigation effect, and are removed with town dressing on level change. No new manifest entries, external assets, runtime downloads, or plugins were adopted. `docs/CODEMAP.md` has one Town services row update; `PROGRESS.md` and the lifecycle log were left to the coordinator.

The clearance test validates complete patrol segments against every gameplay anchor, prop, service path, gate approach, and fence; it also includes an unsafe vendor-overlap fixture. The Godot test checks deterministic route progression/pause/cycle, existing looping clips, presentation-only node structure, disabled/empty configuration, and town lifecycle.

## Asset and security decisions

- **Adopt:** Existing registered KayKit Adventurers bodies and `Idle_A` / `Walking_A` clips.
- **Borrow:** Existing class presentation lookup, idle stance, animation controller/library, and town-dressing path.
- **Reject:** New packs, runtime fetches, external plugins, and a separate art family.
- Security-router assessment: presentation-only UX is outside its review scope; no security-relevant surface was introduced.

## Verification

| Check | Result |
|---|---|
| `make validate-shared` | PASS — 2,208 checks; codemap validation passed. |
| `make validate-assets` | PASS — 451 checks. |
| `.venv/bin/pytest -q tools/test_town_dressing.py` | PASS — 12 tests, including continuous route clearance and unsafe fixture. |
| `godot --headless --path client --script res://tests/test_town_ambient_life.gd` | PASS — 73 checks, including the final catalog-derived idle-position assertion. |
| `make maintainability` | PASS — file-size, extraction-coupling, and progress-dashboard checks. |
| `git diff --check` | PASS. |
| `make bot-visual scenario=town_living_life` | PASS — one visible client scenario; it captured the town at ticks 83 and 105 while the scripted player camera moved, then opened the town vendor. |
| `make bot-client SCENARIO=15_town_vendor_shop_panel HEADLESS=1` | PASS — vendor shop panel scenario. |
| `make regen-screenshots OUT=.artifacts/v506/matched/before/captures SUITE="scenes"` | PASS — 15/15 base captures. |
| `make regen-screenshots OUT=.artifacts/v506/matched/after/captures SUITE="scenes"` | PASS — 15/15 candidate captures; normal/max plaza, vendor, gate, west, and north views inspected. |

The visible bot run opened a real Godot window and recorded player-camera frames. It is scripted proof, not a hand-played `make play` session. The saved frames show residents rendered at those moments and the vendor approach remains legible; stills do not establish continuous animation quality or subjective clickability. The deterministic route test and successful vendor assertions cover those behaviors separately. The scene suite also includes unrelated baseline views; the slice-specific views are `town-play-plaza-normal.png`, `town-play-plaza-max.png`, `town-play-vendor-normal.png`, `town-play-gate-max.png`, `town-play-west-max.png`, and `town-play-north-max.png` under each capture folder.

## Matched render-route comparison

The v500 route fixture ran windowed at 1280×720 with `BOT_STEP_DELAY=0`; host/runtime were Apple M4 Pro, Godot 4.7.2 (`ed1daf0bf`), Forward+ / Metal 4.0, vsync on. Three serial runs per tier were collected on the base source first and then three per tier on the candidate. Every route passed. Each run had 23 steady one-second samples after first-spawn plus two settle samples, complete frame batches, and no dropped intervals. Per-run raw logs and the exact parser summary are under `.artifacts/v506/matched/`; see [render-route-summary.md](../../.artifacts/v506/matched/render-route-summary.md).

Values below are medians of each run's steady p95 (or draw-call p50/p95). The permitted delta is `max(0.5 ms, 5%)` for frame/process p95 and `max(10 calls, 5%)` for draw calls.

| Tier / metric | Base | Candidate | Delta | Limit | Result |
|---|---:|---:|---:|---:|---|
| Balanced frame interval p95 | 15.458 ms | 15.539 ms | +0.081 ms | 0.773 ms | Pass |
| Balanced process p95 | 21.040 ms | 20.150 ms | −0.890 ms | 1.052 ms | Pass |
| Balanced draw calls p50 | 400 | 420 | +20 | 20 | Pass at threshold |
| Balanced draw calls p95 | 431 | 450 | +19 | 21.55 | Pass |
| Performance frame interval p95 | 15.292 ms | 15.490 ms | +0.198 ms | 0.765 ms | Pass |
| Performance process p95 | 21.220 ms | 20.160 ms | −1.060 ms | 1.061 ms | Pass |
| Performance draw calls p50 | 202 | 210 | +8 | 10.1 | Pass |
| Performance draw calls p95 | 234 | 236 | +2 | 11.7 | Pass |

Primitives increased: Balanced p50/p95 `60,802 / 63,918 → 70,458 / 79,458` and Performance `35,651 / 37,050 → 43,257 / 52,435`. The approved gate has no primitive-count limit, but this is a material increase for two residents. The route does not report RSS, so resident-memory acceptance and low-end-device impact remain unverified. No agent-controlled timing job overlapped these runs; OS background activity was not controlled. Baseline preflight load average was 2.34 with Virtualization and Spotlight indexing active, so small timing differences remain host-specific.

The earlier baseline logs in `.artifacts/v506/perf/` overlapped v505 timed controls; they remain preserved as exploratory data and are excluded from the matched result. Godot generated three default import keys in 18 existing `.glb.import` sidecars during setup; all 18 were inspected and restored to base after the runtime checks.

## Complete source manifest

Modified tracked paths:

- `client/scripts/town_dressing.gd`
- `docs/CODEMAP.md`
- `shared/assets/town_presentation.v0.json`
- `shared/assets/town_presentation.v0.schema.json`
- `tools/test_town_dressing.py`

New source/documentation paths:

- `client/scripts/town_ambient_life.gd`
- `client/tests/test_town_ambient_life.gd`
- `docs/specs/v506_spec-living-town.md`
- `docs/plans/v506_2026-10-01-living-town.md`
- `docs/as-built/v506_living-town.md`
- `tools/bot/scenarios/client/town_living_life.json`

Ignored evidence is retained in `.artifacts/v506/before/captures/`, `.artifacts/v506/perf/` (contended exploratory only), `.artifacts/v506/matched/before/`, `.artifacts/v506/matched/after/`, `.artifacts/v506/matched/render-route-summary.md`, `.artifacts/v506/candidate-source/`, and `.artifacts/bot-captures/v506_town_living_{start,camera_move}.{png,json}`. Godot's `.artifacts/showme/godot.log`, `client/.godot/`, and the repo `.venv/` are also ignored local/runtime state. No commits, pushes, full CI, or canonical progress edits were made.
