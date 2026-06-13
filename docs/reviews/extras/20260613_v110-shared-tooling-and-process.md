# arpg-dev - Shared Tooling and Process Review at slice **v110**

**Date:** 2026-06-13
**Scope:** Shared JSON, Python tooling, maintainability ratchet, SDD process.
**Overview:** [`../20260613_v110-overview.md`](../20260613_v110-overview.md)

## Summary

Shared validation remains a strong guardrail. v110 added one new main-config tuning field and kept it
schema-backed. The process issue this cycle was not missing tests, but stale maintainability
baseline metadata after externally landed v109 work.

## Findings

[Strength] `main_config.v0.json` continues to absorb global gameplay tuning that should not be
hardcoded in Go tests or store logic.

[Strength] `make validate-shared` caught the new field through schema validation, and the v110 plan
records the maintenance exception instead of hiding the baseline update.

[Med] `.maintainability/file-size-baseline.tsv` needed a catch-up for post-v109 files. This was
acceptable once documented, but repeated baseline bumps without extraction would erode the ratchet.

[Med] `tools/bot/run.py` is now 5092 lines. Upcoming market and elite-aura bot proof should reuse
existing step types or add focused modules rather than expanding the central runner heavily.

[Low] v109 landed without spec/plan files. The as-built note repairs discoverability, but future
external-agent work should still follow SDD before implementation when possible.

## Recommendations

1. Keep v111/v112 specs explicit about what is not included, especially market UI and elite VFX.
2. If market purchase needs new bot helpers, place them in focused support modules where practical.
3. Treat the v110 baseline update as a one-time catch-up after v109, not a default finish tactic.
4. Continue deriving gameplay test expectations from shared rules instead of copying tuning values.
