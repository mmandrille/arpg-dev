# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v90**

**Date:** 2026-06-12
**Scope:** Shared JSON contracts, Python bots/validators, Make CI, SDD lifecycle, review cadence.
**Baseline:** `main` @ `4a5afb5` (`chore: tune up lighting`); review pre-task cleaned CI blockers before this report landed.
**Stats:** 854 repo files from `rg --files`. Large tooling files: `tools/bot/run.py` 4915 lines and `tools/validate_shared.py` 3110 lines.
**Overview:** [`../20260612_v90-overview.md`](../20260612_v90-overview.md)

---

## Summary

Shared contracts and tooling remain strong. Schema validation, Python protocol bots, replay, client bots, and Godot headless smoke give the project a real end-to-end gate. The review cleanup restored that gate to green before proceeding.

The next shared-contract need is localization: visible content needs stable text keys, English source entries, Spanish entries, and validation/fallback behavior so future skills, monsters, quests, and panels can add text without code changes.

## 1. Architecture

[Strength] Shared validation is still a meaningful content gate. `tools/validate_shared.py` owns schema validation plus cross-file consistency checks, and `make validate-shared` passed with 727 checks.

[Strength] Protocol bots and client bots cover complementary risks: server authority/replay on one side, presentation and UI on the other.

[Med] Localization should become a shared data contract, not a client-only convenience. Skill ids, monster ids, stats, menus, and future quests need text-key conventions that validation can check.

## 2. Technical

[Strength] `make ci` is green after blocker cleanup. The final run passed shared validation, Go test/vet, Python tests, protocol bots plus replay, Godot client bots, and headless smoke.

[Med] `tools/bot/run.py` still carries orchestration, assertions, special-case scenario code, timing, reporting, and process entrypoint in one 4915-line file. The new 15-second elapsed budget is pragmatic, but a per-scenario budget field would be cleaner for naturally long scenarios.

[Med] `tools/validate_shared.py` remains broad at 3110 lines. The localization catalog should either land in a focused validation module or at minimum keep its validation block isolated.

## 3. Maintainability

[Strength] The SDD lifecycle is still useful: review cadence forced blocker cleanup before feature work resumed.

[Med] The project should treat "split and shrink touched files" as a standing rule. This is especially important for localization because it can otherwise thread through every UI surface.

## Top shared/tooling/process recommendations

1. Define `shared/i18n/en.json` as the source catalog and validate key uniqueness/shape.
2. Add Spanish as `shared/i18n/es.json` with key-by-key English fallback.
3. Add tooling that fails when known visible catalog ids lack English text.
4. Keep protocol bot timing as metadata rather than growing global constants.
5. Split new bot/validator functionality into focused helpers when possible.

*Evidence: direct reads of `shared/`, `tools/`, `scripts/ci.sh`, `make/agents.mk`, `PROGRESS.md`, and successful verification listed in the overview.*
