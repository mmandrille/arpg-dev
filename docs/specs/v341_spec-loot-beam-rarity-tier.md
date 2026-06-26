# v341 Spec — Loot Beam Rarity Tier

Status: Complete
Date: 2026-06-25
Codename: `loot-beam-rarity-tier`

## Purpose

Strengthen ground-loot readability with tiered rarity glow intensity, rare+ pickup beams, and white highlight for currency labels so gold text is not confused with unique loot colors.

## Acceptance Criteria

- Glow intensity scales by rarity; rare/unique/set gain a vertical pickup beam.
- Highlighted currency loot labels render white.
- Loot node and label filter unit tests cover beam + currency highlight behavior.
