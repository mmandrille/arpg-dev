# v185 Spec: Companion Rank Scaling and Limits

Status: Complete
Date: 2026-06-15

## Goal

Make companion quantity and core rank scaling data-driven so summon/revive behavior is controlled by shared rules rather than hardcoded slice constants.

## Requirements

- Companion skill payloads declare active limits as data.
- Revive active limit is `1 + floor((rank - 1) / 3)`.
- Ranger wolf remains one active companion through data.
- Revived monster HP/damage scaling remains rule-driven at 50% rank 1 and +10% per rank.
- Companion limit validation rejects invalid limit tuning.
- Recasting beyond the active limit removes oldest same-owner/same-skill companions deterministically.
- Bot proof shows rank 4 Revive supports two active companions and rank-scaled companion HP.

## Non-Goals

- No Ranger multi-wolf scaling yet.
- No UI for companion count or commands.
- No persistence, equipment, inventory, or leveling for companions.
- No mercenaries.

## Presentation Addendum

- Companion quantity scaling is reflected by the generic top-left companion row for all local-player companions.
- The row must remain data-driven by active companion entity state rather than skill-specific constants, so Revive rank scaling and future mercenary companions appear without bespoke UI branches.
- Revive gains a targeting affordance: when Revive is selected and learned, hovering a valid dead monster corpse highlights it and reveals a corpse name/status label.
- Revived companions last 60 seconds at rank 1, plus 10 seconds per additional rank, and the companion row shows their remaining duration with a cooldown-style strip on the companion block.
