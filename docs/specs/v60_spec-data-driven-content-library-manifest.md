# Spec: `data-driven-content-library-manifest`

Status: Complete
Date: 2026-06-10
Branch: `main`
Slice: v60 - data-driven content library manifest
Baseline: v59 `data-driven-skill-catalog`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - shared rules as data and server authority
- [`../researchs/data-driven-content-libraries.md`](../researchs/data-driven-content-libraries.md) - manifest direction and core rule
- [`../reviews/20260610_v60-overview.md`](../reviews/20260610_v60-overview.md) - review recommendation for a skills-only manifest
- [`v59_spec-data-driven-skill-catalog.md`](v59_spec-data-driven-skill-catalog.md) - current skill catalog baseline

## 1. Purpose

v59 made Magic Bolt catalog-driven, but loaders still know the concrete file names
`skills.v0.json` and `skill_presentations.v0.json`. This slice introduces the first
content-library manifest as a deterministic index for skill rules and skill presentation data.

The manifest proves the project rule from `docs/researchs/data-driven-content-libraries.md`:
file paths organize and load content, but stable gameplay IDs remain the public model. Runtime
contracts, persistence, protocol payloads, skill ranks, cooldowns, replay data, and UI state must
continue to refer to `magic_bolt`, not to manifest paths.

This is intentionally a no-behavior-change slice. The merged in-memory skill mechanics and
presentation maps should be equivalent to v59.

## 2. Non-Goals

- No new skills beyond Magic Bolt.
- No item, class, monster, shop, loot, or asset manifest rollout beyond the skills proof.
- No protocol schema bump and no new persisted fields.
- No file-path references in runtime state, protocol payloads, replay output, or goldens.
- No free-form formula language or new skill capability.
- No client UI redesign or external Godot plugin adoption.

## 3. Acceptance Criteria

1. A schema-backed shared content-library manifest exists and is validated by `make validate-shared`.
2. The manifest has strict top-level groups and rejects unknown groups.
3. The skills section can list one or more skill rule files in deterministic declared order.
4. The skill presentation section can list one or more presentation files in deterministic declared order.
5. Manifest paths resolve relative to the manifest file, not the process working directory.
6. The Go rules loader loads skills through the manifest and produces the same `Rules.Skills` map as v59.
7. The Go loader rejects duplicate skill IDs across manifest-listed skill files with a clear error.
8. The Go loader keeps stable skill IDs as map keys; file paths do not enter `SkillDef`, sim state, snapshots, deltas, replay, or goldens.
9. The Godot `SkillRulesLoader` loads skill rules and presentations through the manifest and continues to expose sorted skill IDs.
10. Godot loader diagnostics warn clearly when the manifest, a listed file, or a listed JSON shape is invalid or missing.
11. `tools/validate_shared.py` validates the manifest and cross-checks that manifest-listed skill files merge to the same IDs covered by skill presentations.
12. Focused tests cover duplicate skill IDs and unknown manifest groups rather than relying only on the full repo happy path.
13. Existing Magic Bolt protocol bot, client bot, replay, and golden behavior remain unchanged.
14. `make ci` passes.

## 4. Scope And Likely Files

```text
docs/specs/v60_spec-data-driven-content-library-manifest.md
docs/plans/v60_2026-06-10-data-driven-content-library-manifest.md
docs/as-built/v60_data-driven-content-library-manifest.md
PROGRESS.md

shared/content/content_libraries.v0.json
shared/content/content_libraries.v0.schema.json
shared/rules/skills.v0.json
shared/assets/skill_presentations.v0.json

server/internal/game/rules.go
server/internal/game/rules_manifest_test.go

client/scripts/skill_rules_loader.gd
client/tests/test_skill_rules_loader.gd

tools/validate_shared.py
tools/test_validate_shared.py or focused validator tests if an existing pattern exists
```

## 5. Test And Bot Proof

Required verification:

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'Manifest|LoadRules|Skill'
make client-unit
make bot scenario=32_skill_points_and_magic_bolt.json
SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh
make ci
```

Bot proof is regression proof, not new gameplay proof: the existing Magic Bolt scenarios must keep
passing because the manifest changes loading only.

## 6. Open Questions And Risks

| # | Question / Risk | Default / Mitigation |
|---|-----------------|----------------------|
| R-1 | The review calls the milestone review `v60`, while the next implementation file number is also `v60`. | Use `v60` for the implementation spec/plan because v59 is the latest completed slice and no `v60` spec/plan exists. The review remains a milestone review artifact. |
| R-2 | A generic manifest can grow into an item/class rollout. | Keep this slice skills-only; item/class rollout remains deferred until this loader contract is proven. |
| R-3 | Adding manifest support could expand `rules.go` or `validate_shared.py` monoliths. | Add small helper functions and focused tests; do not bury all checks in one long block. |
| R-4 | Strict manifest validation could break isolated temp-rule tests that only construct a rules directory. | Runtime repo loading should use the manifest. Focused unit tests may create minimal manifest fixtures alongside temp rules when they exercise manifest behavior. |
