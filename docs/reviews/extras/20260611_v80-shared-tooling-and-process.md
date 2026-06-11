# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v80**

**Date:** 2026-06-11
**Scope:** Shared JSON contracts, Python bots/validators, Make CI, SDD lifecycle, review cadence.
**Baseline:** `main` @ `4dcab9e` (`feat: v80: combat threat readability`); worktree was clean at review start.
**Stats:** 811 tracked repo files from `rg --files`; 43 protocol bot scenarios, 31 client bot scenarios, 32 shared rule JSON files, 70 files under `shared/protocol`, and 70 files under `shared/golden`.
**Overview:** [`../20260611_v80-overview.md`](../20260611_v80-overview.md)

---

## Summary

The shared/tooling layer is doing its job: schemas and cross-checks catch data drift, protocol bots prove server behavior, and client bots prove presentation. The main pressure is scale in `tools/bot/run.py` and `tools/validate_shared.py`, both of which remain effective but heavy.

Checks run for this refresh: `make validate-shared` passed with 720 checks, `.venv/bin/pytest tools` passed with 67 tests, and `make client-unit` passed on the v80 baseline. Direct `make client-smoke` failed only when it reached the server-dependent login smoke without a live server. A later full `make ci` attempt ran after unrelated uncommitted v81 skill changes appeared; it passed phases 1-8, then failed in phase 9 because `client/tests/test_skill_rules_loader.gd:20` still expects 3 manifest-loaded skills while the dirty worktree loads 4.

## 1. Architecture

[Strength] Shared validation is a real contract gate, not a thin JSON parser. `tools/validate_shared.py:2` documents schema validation plus cross-consistency guards, and combat lab cross-checks ensure proof monsters and templates resolve at `tools/validate_shared.py:2521`.

[Strength] The scenario split is useful: 43 protocol scenarios under `tools/bot/scenarios` exercise authoritative flows, while 31 client scenarios under `tools/bot/scenarios/client` exercise presentation and UI behavior through Godot.

[Strength] CI orchestration keeps the slice proof end-to-end. The final phases in `scripts/ci.sh:150`, `scripts/ci.sh:176`, and `scripts/ci.sh:184` cover protocol bot plus replay, client bots, and headless smoke.

## 2. Technical

[Med] `tools/bot/run.py` is 4910 lines and still combines scenario loading, orchestration, assertions, co-op special cases, report output, and process entrypoint. The top-level dispatch special-cases co-op and shared-loot scenarios from `tools/bot/run.py:4757` through `tools/bot/run.py:4848`.

[Med] `tools/validate_shared.py` is 3110 lines and combines schema validation, content manifest checks, world checks, combat golden checks, and many domain cross-checks. It is valuable, but new domains should extract local helper modules before this becomes harder to navigate.

[Low] `make client-smoke` is easy to run without the needed server because the target itself does not start one. The failure mode is clear, but the review docs and local workflow should prefer `make ci` for full smoke or `make bot-client` for server-managed client scenarios.

## 3. Maintainability

[Strength] The SDD habit is strong: v80 has a spec, plan, as-built, committed review cadence, and bot/client proof paths connected to the slice lifecycle in `PROGRESS.md`.

[Low] `PROGRESS.md` correctly says v80 is the last engineering review and v90 is next. Because this request is an ad hoc refresh of the already-current v80 review, cadence fields do not need to change.

## 4. Documentation

[Strength] The review cadence remains useful. The v70 and v80 review sets make recurring risks visible, and the surviving realtime fanout issue is small enough to become a focused `/next` maintenance slice.

## Top 5 shared/tooling/process refactors

1. Split `tools/bot/run.py` into protocol driver, assertion helpers, and special co-op scenario modules.
2. Split `tools/validate_shared.py` by domain: rules, worlds, content manifests, protocol examples, and goldens.
3. Document when to use `make client-smoke` directly versus `make ci` or `make bot-client`.
4. Keep new gameplay-visible changes paired with both protocol and client bot proof when presentation matters.
5. Feed the fanout, client-payload, and tooling-size findings into `$next` before starting v81.

*Evidence: direct reads of `shared/`, `tools/`, `scripts/ci.sh`, `make/agents.mk`, `PROGRESS.md`, and current test output listed above.*
