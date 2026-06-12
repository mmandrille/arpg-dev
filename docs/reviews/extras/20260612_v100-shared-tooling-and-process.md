# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v100**

**Date:** 2026-06-12
**Scope:** Shared JSON contracts, Python bot tooling, validators, asset pipeline, Make CI, SDD lifecycle.
**Baseline:** `main` @ `0a249ea` (`fix: polish rogue combat presentation`); worktree clean at review start.
**Stats:** 99 specs, 99 plans, 70 protocol files, 70 golden files, 34 rule files, 81 bot scenario files.
**Overview:** [`../20260612_v100-overview.md`](../20260612_v100-overview.md)

---

## Summary

The shared/tooling layer remains the project’s strongest safety net. Shared validation, protocol bots, client bots, replay, and SDD docs are all active and current.

The next requested feature batch should lean on that strength. Damage types, resistances, and undead immunity are exactly the sort of gameplay behavior that should be schema-backed, validator-checked, and proven by bot scenarios rather than embedded only in Go code.

## 1. Architecture

[Strength] Shared data already centralizes gameplay contracts. `tools/validate_shared.py:1` validates shared protocol, rules, and golden fixtures, and `server/internal/game/rules.go:13` consumes the same rule set server-side.

[Strength] The bot suite is broad: 81 files under `tools/bot/scenarios`, including current rogue skill proof at `tools/bot/scenarios/47_rogue_class_foundation.json`.

[Med] Damage typing has no schema-level owner yet. Adding `damage_type` and `resistances` should update JSON schemas, validator cross-checks, and tests in the same slice.

## 2. Technical

[Med] `tools/bot/run.py:1` remains a 5037-line orchestration file. New bot assertions for damage/resistance should either reuse existing combat-event helpers or land in the smallest possible focused helper.

[Med] `tools/validate_shared.py:1` is 3074 lines. Damage-type validation should be isolated enough that future elements can be added without spreading checks across the whole file.

[Low] The asset generator is now under the ratchet target, so the undead skeleton model can likely be added without growing a new oversized tool file. The plan still needs to check asset manifest validation.

## 3. Maintainability

[Strength] The SDD lifecycle is current through v99, and the review gate caught the need to pause before v100 gameplay work.

[Med] The next two user ideas are dependent. Damage types and resistances should ship first; undead poison immunity should consume that new contract instead of inventing a parallel immunity flag.

## 4. Documentation

[Strength] Specs/plans are complete through v99 and `PROGRESS.md` is the canonical lifecycle file.

[Med] Review findings should feed directly into the upcoming specs: canonical type names, schema ranges, bot proof, and visual/art adoption notes.

## Top shared/tooling/process recommendations

1. Add a canonical shared damage-type enum with `force` fallback.
2. Add monster resistance schema and validator checks for known damage types and valid ranges.
3. Reuse bot combat-event assertions to prove resistant, weak, and immune targets.
4. Keep undead as a second slice that consumes v100’s resistance contract.
5. Run `make validate-shared`, focused Go tests, bot proof, and final `make ci` for both slices.

*Evidence: `tools/validate_shared.py:1`, `tools/bot/run.py:1`, `tools/bot/scenarios/47_rogue_class_foundation.json`, shared file counts, and current SDD lifecycle metadata.*
