# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v398**

**Date:** 2026-07-01  
**Scope:** `shared/`, `tools/`, Makefile/CI, docs/SDD — v385–v398 vs v384 baseline (all ten v384 recommendations closed).  
**Baseline:** `main` @ `e0f402e4`. Worktree: **clean**.  
**Stats:** 220 bot scenarios (35 pack / ~180 extended / 4 perf probes); 73 protocol files; 40 rules; 70 goldens; 276 Go / 311 GDScript / 81 Python files; `validate_shared.py` 3,033 lines; `run.py` 4,301 lines.  
**Overview:** [`../20260701_v398-overview.md`](../20260701_v398-overview.md)

---

## Summary

Contract hygiene and the two-tier CI model remain a **strength**. Since v384 (all ten recommendations closed; ci-full green post-paydown), slices **v390–v398** added substantive shared/rules/content work (upgrade shards, codex compiler, item archetypes, `client_reconnect` tuning) without a protocol version bump — canonical wire remains **v8**.

Process quality **improved** vs the v370–v378 stub era: v390–v398 as-builts are usable regression guides. Remaining drift: **engineering review was overdue** (milestone ~v394; now v398), **v392–v394 lack formal plans**, `validate_shared.py` `cross_checks()` remains a ~2,800-line monolith, **five duplicate scenario IDs** (protocol + client pairs; two in CI pack), and the **tuning-friendly rule-test audit** backlog is still open.

## 1. Architecture

**[Strength]** Rules-as-data boundary held through v390–v398. Gameplay tuning lives in `shared/rules/` and `shared/assets/`; v398 explicitly avoided protocol/schema bumps.

**[Strength]** Protocol canonical set stable: `envelope.v8`, `messages.v8`, `session_snapshot.v8`, `state_delta.v8` in `tools/validate_shared.py`.

**[Strength]** New **content compiler** domain (`tools/content/build_codex.py` → `shared/content/codex_index.v0.json`) extends shared contracts without crossing the authoritative boundary.

**[Strength]** CI pack policy enforced: `tools/bot/ci_pack.json` — 22 protocol + 13 client = 35 pack members; `validate_ci_pack()` in CI step 7.

**[Med]** **Duplicate scenario IDs** across protocol and client trees (5 pairs). Two pack members duplicated:

| Scenario ID | Protocol | Client |
|-------------|----------|--------|
| `mystery_seller_core` | `37_mystery_seller_core.json` | `client/24_mystery_seller_core.json` |
| `quest_town_turn_in` | `97_quest_town_turn_in.json` | `client/75_quest_town_turn_in.json` |

Also duplicated (extended): `mystery_seller_paid_reroll`, `shop_stock_lifecycle`, `unique_burn_effect_live`.

**[Med]** v397 archetype rename is **breaking** for persisted DBs (`make db-reset` documented).

**[Low]** v398 reconnect has **no authoritative bot proof** for network partition — unit-tested only.

## 2. Technical

**[Strength]** `make ci` pipeline (`scripts/ci.sh`) — 11-step curated gate: maintainability → `validate-shared` → assets → determinism → Go → Python → server → protocol pack → client pack → headless smoke.

**[Strength]** `validate-shared` modular imports grew: `validate_fog_presentation.py`, `validate_item_naming.py`, `validate_main_config.py`, `validate_skills.py`, etc. Fog `point_light` schema gap from v384 **closed**.

**[Strength]** Bot runtime modularization: 38 modules under `tools/bot/` including `bot_types.py`, `bot_context.py`, domain runtimes. `run.py` grandfathered at 4,301 lines.

**[Strength]** `make validate-shared` — **1,818 checks passed** (2026-07-01 spot-check).

**[Med]** `validate_shared.py` still **3,033 lines**; `cross_checks()` ~lines 201–3017 — v337/v384 extraction recommendation only partially landed.

**[Med]** Codex semantic drift guarded by `tools/test_build_codex.py` (4 tests) but **no codex/rules cross_checks** in `validate_shared.py`.

**[Med]** **Tuning-friendly rule-test audit** remains open backlog (`PROGRESS.md` Testing/tooling; v120 scoped to one GDScript file).

## 3. Maintainability

**[Strength]** Maintainability ratchet green: 35 grandfathered files, 65,746 lines; extraction-coupling 0 `helpers=globals()` sites.

**[Strength]** CI pack curation discipline in `.cursor/rules/ci-pack-maintenance.mdc` and `docs/progress/scenario-catalog.md`. v390–v398 new proofs default **extended** — pack size stable at 35.

**[Med]** Grandfathered tooling coordinators:

| File | Lines (current / baseline) |
|------|---------------------------|
| `tools/validate_shared.py` | 3,033 / 3,038 |
| `tools/bot/run.py` | 4,301 / 4,284 |
| `tools/bot/test_protocol.py` | 1,434 (grandfathered) |

**[Low]** `skill_visual` scenario excluded from `SCENARIO=all` protocol runs — intentional but increases matrix mental load.

## 4. Documentation

**[Strength]** v390–v398 SDD trail healthy: specs + substantive as-builts with proof commands for all nine slices.

**[Strength]** Lifecycle index (`docs/progress/slice-lifecycle.md`) current through v398.

**[Med]** **Missing formal plans** for v392, v393, v394 (lifecycle plan column `—`).

**[Strength]** v398 spec documents spec-gate boundary and explicit non-goals (no merge-gate bot scenario).

**[Strength]** `docs/CODEMAP.md` includes game-codex domain row.

**[Low]** `docs/progress/scenario-catalog.md` client pack count slightly stale (~14 vs actual 13).

## Top 5 shared/tooling/process refactors

1. **[minor · Process]** **Land this v398 review** and run `$refactor` against the new overview — cadence was ~v394; 14 slices since v384.

2. **[High · Maint]** **Decompose `validate_shared.cross_checks()`** into importable domain modules (`validate_shared.py:201–3017`; partial extractions exist).

3. **[Med · Test]** **Systematic tuning-friendly rule-test audit** (Go + GDScript + Python + bot) — `PROGRESS.md` deferred item; policy in `CLAUDE.md`.

4. **[Med · Process]** **Resolve duplicate bot scenario IDs** — 5 protocol/client pairs; 2 in `ci_pack.json` (`mystery_seller_core`, `quest_town_turn_in`).

5. **[Med · SDD]** **Backfill plans for v392–v394** + add codex semantic cross-validation (`build_codex.py` output vs live rules).

*Evidence: `PROGRESS.md`, `tools/bot/ci_pack.json`, `scripts/ci.sh`, `.maintainability/file-size-baseline.tsv`, v390–v398 as-builts; comparison to [`20260629_v384-shared-tooling-and-process.md`](20260629_v384-shared-tooling-and-process.md).*
