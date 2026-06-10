# v60 As Built: Data-Driven Content Library Manifest

Date: 2026-06-10
Spec: [`docs/specs/v60_spec-data-driven-content-library-manifest.md`](../specs/v60_spec-data-driven-content-library-manifest.md)
Plan: [`docs/plans/v60_2026-06-10-data-driven-content-library-manifest.md`](../plans/v60_2026-06-10-data-driven-content-library-manifest.md)

## What shipped

- Added `shared/content/content_libraries.v0.json` as the first content-library manifest. It indexes
  the existing skill rules and skill presentation files while preserving stable runtime skill IDs.
- Added a strict manifest schema that rejects unknown top-level groups and keeps v60 skills-only:
  `rules.skills` plus `assets.skills.presentations`.
- Added focused Python manifest helpers and tests for relative path resolution, collection merging,
  unknown group rejection, and duplicate ID rejection.
- `make validate-shared` now validates `shared/content/*.v0.json` and checks that manifest-listed
  skill rules and presentations merge to the canonical v59 IDs.
- The Go rules loader now reads `Rules.Skills` through the content manifest, resolves paths relative
  to the manifest file, preserves declared file order, and rejects duplicate skill IDs.
- Godot `SkillRulesLoader` now loads both skill rules and skill presentations through the manifest,
  keeps sorted public skill IDs, and emits warnings for missing/bad manifest or listed files.
- Added focused Go and Godot loader tests, and registered the Godot loader test in `client-unit`.
- Stabilized the existing protocol Magic Bolt bot proof with a regen wait between melee and archer
  sweeps; gameplay behavior is unchanged.

## Proof

- `make validate-shared`
- `.venv/bin/python -m pytest -q tools/test_validate_shared.py`
- `cd server && go test ./internal/game/... -run 'Manifest|LoadRules|Skill'`
- `make client-unit`
- `SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh`
- `make bot scenario=32_skill_points_and_magic_bolt.json`
- `make ci`

## Deferred

- Item, class, monster, loot, shop, and broader asset/presentation manifest rollout.
- New active skills or new skill capability types.
- Protocol version changes.
- Runtime, persistence, protocol, replay, or golden references to manifest file paths.
