# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v130**

**Date:** 2026-06-13
**Scope:** `shared/`, Python tools, bot scenarios, validation, Make/CI, SDD docs, and review cadence after v121-v130.
**Baseline:** `main` at `8db5538a` (`fix: deliver accepted market offers to stash`); worktree clean at review start.
**Stats:** 36 shared rule JSON files, 71 protocol files, 70 golden files, 62 bot scenario JSON files.
**Overview:** [`../20260613_v130-overview.md`](../20260613_v130-overview.md)

---

## Summary

The process did what the v120 review asked: skill validation was split out of the monolithic shared
validator, bot scenarios learned to derive skill caps from rules, and the town-service bridge
reduced `main.gd` pressure. The remaining tooling risk is concentrated in the still-large bot runner
and the remaining non-skill domains inside `tools/validate_shared.py`.

## 1. Architecture

[Strength] `tools/validate_shared.py` is now a CLI wrapper plus domain checks, and v126 extracted
skill-specific validation into `tools/validate_skills.py` (`tools/validate_shared.py:22`,
`tools/validate_shared.py:30`, `tools/validate_skills.py:21`).

[Strength] Bot scenario tuning improved. `tools/bot/run.py` resolves `max_rank: "from_rules"` from
`shared/rules/skills.v0.json`, which keeps skill-cap tuning out of scenario JSON
(`tools/bot/run.py:61`, `tools/bot/run.py:64`).

[Med] Validation extraction is only partially complete. `validate_shared.py` remains 3166 lines and
still owns many unrelated domains. The next likely split for unique item work is unique/effect
catalog validation, because the upcoming chest/test work will rely on those rules.

## 2. Technical

[Strength] v128-v130 kept market state durable through migrations and store tests. The audit table
has indexes by listing and actor (`server/migrations/0024_market_expiration_and_audit.sql:22`,
`server/migrations/0024_market_expiration_and_audit.sql:36`,
`server/migrations/0024_market_expiration_and_audit.sql:39`).

[Med] The protocol bot runner remains broad at 5148 lines. Runtime assertions are still a long
chain of domain cases (`tools/bot/run.py:3700`, `tools/bot/run.py:3800`). The next unique chest
scenario should avoid adding another large branch if a small assertion helper can carry the domain.

[Low] `PROGRESS.md` still lists `tuning-friendly-rule-tests` as an open candidate even though v120
and v125 shipped concrete pieces (`PROGRESS.md:1029`). This is not blocking, but the label should
be reframed as continuing audit work so `/next` does not treat the already-shipped slice name as
unstarted.

## 3. Maintainability

[Strength] The maintainability ratchet is still visible and current. The largest files are
grandfathered, and recent slices documented focused exceptions instead of silently growing them.

[Med] The bot and validator are the next best tooling split targets. Skill validation is a good
template: keep `make validate-shared` and scenario JSON stable while moving domain details behind
small importable modules.

[Low] Review cadence is healthy. The repo now has review sets at v53, v60, v70, v80, v90, v100,
v110, v120, and this v130 pre-task.

## 4. Documentation

[Strength] SDD traceability remains strong. Every v121-v130 completed slice has spec, plan, and
as-built entries in the lifecycle table.

[Med] The next feature batch should feed review findings directly into slice specs: unique test
chest, fixed unique packages, unique inspection, and validation split work are all clear candidates.

## Top 5 shared/tooling/process refactors

1. Split unique item/effect validation out of `validate_shared.py`.
2. Split bot runtime assertions into domain helpers before adding unique/chest-specific assertions.
3. Reword the stale `tuning-friendly-rule-tests` candidate in `PROGRESS.md` as ongoing audit work.
4. Keep market audit APIs internal until a specific player/admin inspection UI is specced.
5. Preserve review cadence at v140 after the unique-items batch.

*Evidence: line counts from `wc -l`; file counts from `find`; code references above.*
