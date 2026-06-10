# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v60**

**Date:** 2026-06-10
**Scope:** `shared/`, `tools/`, `assets/`, root make/CI scripts, `docs/` SDD process and research notes
**Baseline:** `main` @ `7210fbf` - `feat: v59: data-driven skill catalog` (clean tree at review start)
**Stats:** `shared/` + `tools/` 33,644 LOC; 68 protocol schema/example files; 70 golden files; 59 specs; 59 plans; 58 as-built files; 38 protocol bot scenarios plus 28 client scenarios
**Overview:** [`../20260610_v60-overview.md`](../20260610_v60-overview.md)

---

## Summary

The shared/tooling layer is in a better state than the v53 review. The old health-regen naming risk
now has explicit cross-checks, golden regeneration exists, skill presentation metadata is split from
skill mechanics, and validation covers 606 consistency checks.

The next risk is scale. Protocol schemas have grown to v8 with all versions validated as live, and
`validate_shared.py` is now 2,861 LOC. A skills-only content-library manifest is a good next slice,
but it should be used to prove a small deterministic loader/index pattern before attempting item or
class library splitting.

## 1. Architecture

- **[Strength] Rules-as-data still holds.** `shared/rules/skills.v0.schema.json:22` allows only the
  current `projectile_attack` capability; new behavior types still require schema/code/test changes.
- **[Strength] Skill presentation metadata is separate from mechanics.**
  `shared/rules/skills.v0.json:4` defines Magic Bolt mechanics; `shared/assets/skill_presentations.v0.json:4`
  defines label/icon/projectile presentation. This is the right shape for library indexing.
- **[Med] The protocol surface keeps accumulating versions.** `shared/protocol` now contains 68
  files, with schema families through v8 still validated. No deprecation/archive policy exists.
- **[Med] Content manifests need a loader contract before broad rollout.**
  `docs/researchs/data-driven-content-libraries.md:107` lists deterministic order, duplicate-ID
  rejection, relative path resolution, schema validation, and path-free runtime state. Those should
  become executable checks in the first manifest slice.

## 2. Technical

- **[Strength] Shared validation is broad and green.** `make validate-shared` passed with 606 checks,
  including skill presentation coverage at `tools/validate_shared.py:1061` and health-regen naming
  guards at `tools/validate_shared.py:2809`.
- **[Strength] Golden regeneration exists.** `make/server.mk:15` defines `regen-golden`, resolving
  the v53 manual-golden hazard.
- **[High] `validate_shared.py` remains a god-script.** `tools/validate_shared.py:163` starts a
  `cross_checks` function that runs through thousands of lines of schema/data/formula/golden logic.
  Adding manifest validation here directly will worsen the problem unless it is extracted first.
- **[Med] Examples still validate against the current schema, not their authoring version.**
  `tools/validate_shared.py:60` maps protocol example instances to schemas centrally. That is fine
  for now, but protocol history will become harder to reason about as v9+ arrives.

## 3. Maintainability

- **[Strength] SDD cadence remains excellent.** There are 59 specs and 59 plans, and every completed
  recent slice links spec, plan, and as-built docs.
- **[Med] `PROGRESS.md` is carrying both lifecycle and backlog history.** It is disciplined, but
  long. The curated candidate table is useful; keep adding short steering entries rather than
  expanding narrative history.
- **[Med] The next manifest slice should be skills-only.** Skills have one small catalog, one
  presentation file, one client loader, server-side generic validation, and a focused bot/golden
  scenario. Items involve fixed definitions, templates, visuals, presentations, shop/loot refs, and
  duplicated icon rendering, so they are a larger follow-up.

## 4. Documentation

- **[Strength] The data-driven content note captures the right rule.**
  `docs/researchs/data-driven-content-libraries.md:15` says file-path dictionaries are indexes, not
  the gameplay model. That should be copied into the manifest spec.
- **[Med] The note is design intent, not an ADR.** After a manifest lands, write an as-built/ADR
  update so future content slices know whether to add data files, extend loader capabilities, or
  change runtime contracts.
- **[Strength] `CLAUDE.md` and `PROGRESS.md` now point agents at current invariants.** The v53
  staleness finding is resolved.

## Top 5 shared/tooling/process improvements

1. Implement a skills-only content library manifest with deterministic merge validation and no
   runtime path IDs.
2. Extract manifest/content checks into a small validator module instead of growing
   `validate_shared.py`.
3. Add protocol version retirement rules before v9/v10.
4. Keep item/class manifest rollout deferred until the skills manifest proves the loader contract.
5. Convert the data-driven content note into an ADR/as-built decision after the first manifest slice.

*Evidence: direct reads of `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json`,
`shared/assets/skill_presentations.v0.json`, `tools/validate_shared.py`, `make/server.mk`,
`docs/researchs/data-driven-content-libraries.md`; `make validate-shared` passed.*
