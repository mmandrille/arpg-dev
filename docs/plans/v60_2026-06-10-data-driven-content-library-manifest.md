# v60 Plan - Data-Driven Content Library Manifest

Status: Ready for implementation
Goal: Add a skills-only content-library manifest that indexes skill rules and skill presentation files without changing gameplay behavior.
Architecture: The manifest is a deterministic loader index, not a runtime model. Go and Godot resolve manifest paths relative to the manifest file, merge skill rules/presentations in declared order, and keep stable skill IDs as the only gameplay/runtime identifiers. Validation gets focused manifest checks so this proof does not grow the existing loader and validator monoliths unnecessarily.
Tech stack: shared JSON schema/data, Go rules loader, Godot GDScript loader, Python shared validator, existing protocol/client bot regression proof.

## Baseline and shortcut decision

Baseline is v59 `data-driven-skill-catalog` on `main`, plus the v60 review recommendation and
`docs/researchs/data-driven-content-libraries.md`.

Godot shortcut decision: **reject external plugins/assets**. This slice only changes shared-data
loading for the existing skill UI; no UI/art shortcut is needed and no plugin may own skill
authority.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `shared/content/content_libraries.v0.json` | Skills-only manifest instance |
| Add | `shared/content/content_libraries.v0.schema.json` | Strict manifest schema |
| Add | `tools/content_manifest.py` | Focused manifest path/merge helpers |
| Add | `tools/test_validate_shared.py` | Manifest schema/merge unit tests |
| Modify | `tools/validate_shared.py` | Include `shared/content` schema/instance validation and manifest cross-checks |
| Add | `server/internal/game/rules_manifest.go` | Server manifest loader helpers |
| Modify | `server/internal/game/rules.go` | Load skill rules through manifest-listed files |
| Add | `server/internal/game/rules_manifest_test.go` | Focused manifest loader tests |
| Modify | `client/scripts/skill_rules_loader.gd` | Load skill rules/presentations through manifest-listed files |
| Add | `client/tests/test_skill_rules_loader.gd` | Focused Godot loader test |
| Modify | `scripts/client_smoke.sh` | Register the loader test in client-unit |
| Modify | `tools/bot/scenarios/32_skill_points_and_magic_bolt.json` | Stabilize existing protocol regression with a regen wait |
| Modify | `docs/specs/v60_spec-data-driven-content-library-manifest.md` | Mark complete at close-out |
| Modify | `docs/plans/v60_2026-06-10-data-driven-content-library-manifest.md` | Track task completion |
| Add | `docs/as-built/v60_data-driven-content-library-manifest.md` | Slice proof summary |
| Modify | `PROGRESS.md` | Lifecycle, numbering note, backlog/deferred updates |

## Task 1 - Shared manifest contract

Files:
- Add: `shared/content/content_libraries.v0.json`
- Add: `shared/content/content_libraries.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add a strict skills-only manifest schema with `version`, `rules.skills`, and
  `assets.skills.presentations`; reject unknown top-level groups.
- [x] Step 1.2: Add the repo manifest referencing the existing `shared/rules/skills.v0.json` and
  `shared/assets/skill_presentations.v0.json` by paths relative to the manifest file.
- [x] Step 1.3: Extend `make validate-shared` discovery to validate `shared/content/*.v0.json`
  against matching schemas.
- [x] Step 1.4: Add manifest cross-checks that the listed skill rules and presentations merge to the
  same skill IDs already covered by v59 validation.

```bash
make validate-shared
```

## Task 2 - Go rules loader manifest support

Files:
- Modify: `server/internal/game/rules.go`
- Add: `server/internal/game/rules_manifest_test.go`

- [x] Step 2.1: Add small manifest structs/helpers that resolve paths relative to
  `shared/content/content_libraries.v0.json`.
- [x] Step 2.2: Load `Rules.Skills` through the manifest in deterministic declared order.
- [x] Step 2.3: Reject duplicate skill IDs across listed files with a clear error.
- [x] Step 2.4: Keep `validateSkillRules` operating on the merged map, preserving all v59 behavior.
- [x] Step 2.5: Add focused tests for repo manifest loading, duplicate skill IDs, and missing listed
  files.

```bash
cd server && go test ./internal/game/... -run 'Manifest|LoadRules|Skill'
```

## Task 3 - Godot skill loader manifest support

Files:
- Modify: `client/scripts/skill_rules_loader.gd`
- Add: `client/tests/test_skill_rules_loader.gd`

- [x] Step 3.1: Load the content manifest before skill rules/presentations.
- [x] Step 3.2: Resolve listed paths relative to the manifest file and merge skills/presentations in
  declared order.
- [x] Step 3.3: Keep `skill_ids()` sorted and keep existing public loader methods stable.
- [x] Step 3.4: Emit `push_warning` diagnostics for missing manifest files, missing listed files,
  invalid JSON shape, or duplicate IDs.
- [x] Step 3.5: Add a focused headless test that verifies manifest-driven skill and presentation
  loading for Magic Bolt.

```bash
make client-unit
```

## Task 4 - Regression bot proof

Files:
- Existing protocol and client bot scenarios

- [x] Step 4.1: Run the existing Magic Bolt protocol bot scenario to prove gameplay output did not
  change.
- [x] Step 4.2: Run the existing Magic Bolt client bot scenario to prove the UI still resolves
  Magic Bolt metadata after manifest loading.

```bash
make bot scenario=32_skill_points_and_magic_bolt.json
SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh
```

## Task 5 - Lifecycle docs and final verification

Files:
- Modify: `docs/specs/v60_spec-data-driven-content-library-manifest.md`
- Modify: `docs/plans/v60_2026-06-10-data-driven-content-library-manifest.md`
- Add: `docs/as-built/v60_data-driven-content-library-manifest.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark plan checkboxes complete as implementation lands.
- [x] Step 5.2: Mark the spec complete.
- [x] Step 5.3: Add the v60 lifecycle row and current status updates in `PROGRESS.md`.
- [x] Step 5.4: Add v60 as-built notes, including deferred item/class manifest rollout.
- [x] Step 5.5: Run final CI.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'Manifest|LoadRules|Skill'`
- [x] `make client-unit`
- [x] `make bot scenario=32_skill_points_and_magic_bolt.json`
- [x] `SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh`
- [x] `make ci`

## Deferred scope

- Item, class, monster, loot, shop, and broader asset manifest rollout.
- New active skills or new skill capabilities.
- Protocol version changes.
- Runtime/persistence/protocol references to manifest file paths.
