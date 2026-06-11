# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v80**

**Date:** 2026-06-11
**Scope:** Shared JSON contracts, Python bots/validators, Make CI, SDD lifecycle and review cadence.
**Baseline:** `main` @ `2acb4c1` (`feat: v79: elite pack roles`) with v80 slice changes in the working tree.
**Stats:** 43 protocol scenarios, 31 client scenarios, 32 shared rule JSON files, 36 protocol JSON/schema files.
**Overview:** [`../20260611_v80-overview.md`](../20260611_v80-overview.md)

---

## Summary

The shared/tooling layer is doing its job: rule schemas caught pack-role structure, protocol bots
proved server behavior, and client bots proved presentation. The main pressure is scale in
`tools/bot/run.py` and `tools/validate_shared.py`.

## 1. Architecture

[Strength] Shared validation is a real contract gate, not a thin JSON parser. Evidence:
`tools/validate_shared.py:1` describes schema plus cross-consistency checks, and combat lab
cross-checks ensure required proof monsters/equipment resolve at `tools/validate_shared.py:2521`.

[Strength] Scenario coverage keeps expanding through public flows: protocol bot selection and
fresh-session checks run through `tools/bot/run.py:4748`, while client bot scenarios remain
separate display proofs.

## 2. Technical

[Med] `tools/bot/run.py` is 4910 lines and contains orchestration, scenario dispatch, assertions,
co-op special cases, and report output. Evidence: top-level scenario loop and special-case dispatch
span `tools/bot/run.py:4748` through `tools/bot/run.py:4848`.

[Med] `tools/validate_shared.py` is 3110 lines and combines schema validation, content manifest
checks, world checks, combat golden checks, and many domain cross-checks. It is effective, but new
domains should extract local helper modules before this becomes harder to navigate.

## 3. Maintainability

[Strength] The SDD habit is strong: v80 has a spec, plan, bot proof, as-built, and review cadence
work in one slice.

[Low] `PROGRESS.md` remains the canonical baseline, but recent out-of-order v76/v79 work means the
status table needs careful updates during v80 close-out. The slice lifecycle table should include
v79 and v80, while preserving the note that earlier v76-v78 main-config slices already existed.

## 4. Documentation

[Strength] The review cadence remains useful. The v70 review findings are still visible, and at
least one high-value finding, realtime fanout level snapshotting, remains actionable and small.

## Top 5 shared/tooling/process refactors

1. Split `tools/bot/run.py` into protocol driver, assertion helpers, and special co-op scenario modules.
2. Split `tools/validate_shared.py` by domain: rules, worlds, content manifests, protocol examples, goldens.
3. Add a lightweight `make bot-client scenario=<id>` wrapper if the script supports only env vars today.
4. Keep new gameplay-visible changes paired with both protocol and client bot proof when presentation matters.
5. After v80, update review cadence to v90 and feed the fanout/client-payload/tooling-size findings into `/next`.
