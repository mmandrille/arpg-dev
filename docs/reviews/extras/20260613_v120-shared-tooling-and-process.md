# v120 Shared, Tooling, and Process Review

Date: 2026-06-13
Scope: `shared/`, `tools/`, SDD docs, maintainability process

## Summary

The shared-data direction is healthy: rules are validated, protocol/bot scenarios cover important
flows, and v120 began removing accidental test locks on tunable data. Tooling remains effective but
large. The next process wins should come from splitting validator/bot domains and continuing the
tuning-friendly assertion audit.

## Strengths

- Shared validation now cross-checks enabled unique item metadata and effect references tightly
  enough to support enabling the full catalog.
- The protocol bot suite continues to provide deterministic proofs for economy, elite, unique, and
  persistence-sensitive behavior.
- The maintainability ratchet is active and caught by the normal closeout commands.
- The SDD trail remains complete for v118-v120, including as-built summaries and review cadence.

## Findings

### Medium: `tools/validate_shared.py` should be split by catalog

`tools/validate_shared.py` is about 3.3k lines and validates many unrelated shared domains. It is
still a useful single entrypoint, but new catalog validation should not keep expanding the same
implementation file.

Recommendation: keep `tools/validate_shared.py` as the CLI wrapper, but move unique/effect, skill,
item-template, and economy validation into importable modules.

### Medium: `tools/bot/run.py` remains a multi-domain runner

`tools/bot/run.py` is about 5.1k lines and contains action dispatch, inventory/equipment helpers,
scenario setup, and domain-specific assertions. The scenario coverage is valuable, but changes in
one domain still require reviewing a very large tool file.

Recommendation: preserve the scenario JSON shape while extracting domain action/assertion modules.

### Low: tuning-friendly tests need a repo-wide follow-through

v120 converted one high-signal Godot test. The backlog item remains larger: Go tests, GDScript
tests, Python tests, and bot scenarios should be audited for values copied from `shared/rules`.

Recommendation: classify exact values as contract/golden, rule-derived fixture, semantic/range
assertion, or accidental pin. Convert accidental pins first.

## Verification Reviewed

- `make maintainability` passed during v120.
- `make ci` passed during v120 before the review files were added.
