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
