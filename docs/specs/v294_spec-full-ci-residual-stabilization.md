# v294 Spec: Full-CI Residual Stabilization

## Status

Complete after final `make ci` proof.

## Goal

Restore a green full-CI baseline after v293 by fixing residual non-feature regressions in broad Go,
protocol bot, and client bot gates so the repository is ready for the next review/refactor and
feature autoloop.

## Background

The v293 focused gates passed, but the full `make ci` attempt on 2026-06-19 exposed residual
failures outside the bishop badge-cost slice:

- `TestEliteMinionFollowsLeaderWithoutPassiveAggro` used brittle synthetic geometry.
- Protocol mercenary scenarios still assumed the original guard-only offer set after v289 added
  mercenary offer variants.
- Several older client bot scenarios pinned stale boss/mercenary labels or could not match
  click-driven combat events after presentation/protocol growth.
- `second_boss_template` could finish before observing one authoritative tail event emitted by
  summoned adds during replay reconstruction.

## Requirements

- Keep the slice stabilization-only: no new gameplay feature, tuning rebalance, or asset pipeline.
- Preserve server authority and existing gameplay behavior while making tests/scenarios describe the
  current contracts.
- Keep mercenary proofs compatible with v289 offer variants without weakening the hire/death-loss
  behavior being tested.
- Keep boss UI proofs compatible with a variable boss pool while preserving a Cave Warden-specific
  decal proof where shape coverage depends on that template.
- Keep replay diagnostics actionable when event counts differ.
- Restore full `make ci` to green before committing.

## Non-Goals

- New boss templates, mercenary offers, combat tuning, or UI features.
- Reworking the bot framework beyond the narrow event-matching fix needed by existing scenarios.
- Running the next engineering review or refactor pass inside this slice.

## External Adoption Decision

- Adopt: none.
- Borrow: none.
- Reject: external assets/plugins. This slice only stabilizes in-repo tests, scenarios, and bot
  helpers.
