# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v70**

**Date:** 2026-06-11
**Scope:** `shared/`, `tools/`, CI/make scripts, SDD docs, ADR/process metadata
**Baseline:** `main` @ `ec1f6ef` — `feat: v70: class skill and item gates` (clean tree at review start)
**Stats:** 30 shared rule JSON files, 70 golden JSON files, 36 protocol schema files, 70 specs, 70 plans, 70 as-built files after this review set, 41 protocol bot scenarios plus 30 client scenarios.
**Overview:** [`../20260611_v70-overview.md`](../20260611_v70-overview.md)

---

## Summary

The shared/tooling layer continued to mature since v60. The skills-only content-library manifest
landed, class rules now participate in shared validation, and CI validation expanded from 606 checks
at v60 to 682 checks at v70. SDD cadence remains strong: v70 has spec, plan, as-built, bot proof,
and a green full CI gate.

The main risks are scale and duplicated payload assumptions. `validate_shared.py` is now 3,075 LOC,
`tools/bot/run.py` is 4,901 LOC, and both the Godot client and Python bot still index
`equipped_update.slot` directly.

## 1. Architecture

- **[Strength] Rules-as-data remains enforced.** Class-required weapons are schema-backed in
  `shared/rules/items.v0.schema.json:25` and cross-checked in `tools/validate_shared.py:836`.
- **[Strength] The content manifest is real, not just design intent.** Backend and client loaders
  now consume `shared/content/content_libraries.v0.json`; the Python validator checks the manifest
  before skill merge validation.
- **[Med] Protocol versions still have no retirement policy.** Validator checks keep v4-v8 schema
  sets live at `tools/validate_shared.py:313`; no archive/deprecation rule exists before v9+.
- **[Med] Shared class presentation is not data-backed yet.** v70 deliberately deferred sprites and
  class UI. v71 should avoid hardcoding class presentation in only the Godot panel if it can be a
  small shared asset catalog.

## 2. Technical

- **[Strength] Validation caught up with class gameplay.** `class_required` must reference known
  classes, only equippable items may use it, and class weapons must be stronger than `rusty_sword`
  (`tools/validate_shared.py:836`).
- **[Strength] Protocol bot coverage expanded.** The new `paladin_heal_skill` scenario passed in
  full CI, and class-specific character creation supports Magic Bolt/Rage/Heal proofs.
- **[Med] Python bot delta application mirrors the client slot hazard.**
  `tools/bot/run.py:2350` indexes `c["slot"]` for `equipped_update`. Make it as defensive as the
  recommended Godot fix.
- **[High] `validate_shared.py` remains a god-script.** `cross_checks` starts at
  `tools/validate_shared.py:196`; new class, item, protocol, content, and formula checks all keep
  accumulating in one function.

## 3. Maintainability

- **[Strength] SDD lifecycle is disciplined.** There are 70 specs and 70 plans, and every recent
  slice has direct spec/plan/as-built links in `PROGRESS.md`.
- **[Med] Bot scenario runner scale is becoming a maintenance concern.** `tools/bot/run.py` is
  4,901 LOC and owns HTTP setup, scenario interpretation, state tracking, replay checks, and client
  support logic.
- **[Med] `PROGRESS.md` remains useful but long.** Continue adding short steering entries rather
  than expanding narrative backlog.

## 4. Documentation

- **[Strength] ADR-0014 gives future progression/item/economy slices product guardrails.**
- **[Med] Manifest outcome should be promoted into durable architecture guidance.** v60 shipped the
  skills manifest; future item/class manifests need a short ADR/as-built rule saying when to add a
  manifest entry versus when to change runtime contracts.

## Top 5 shared/tooling/process improvements

1. Split `validate_shared.py` cross-check domains into helper modules.
2. Guard `equipped_update.slot` in both Python bot and Godot client.
3. Define protocol version retirement/archive policy before adding v9.
4. Add a small shared class presentation catalog in v71 if class sprites are introduced.
5. Split `tools/bot/run.py` state tracking and scenario actions when the next bot feature lands.

*Evidence: direct reads of shared schemas/rules/goldens, validator and bot code, ADR-0001,
ADR-0014, v70 docs, and CI output; `make validate-shared`, `make test-py`, `make bot`, and
`make ci` passed.*
