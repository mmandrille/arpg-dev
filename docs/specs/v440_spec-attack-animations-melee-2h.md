# v440 Spec — Attack Animations Melee 2H

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `attack-animations-melee-2h`

## Purpose

Data-driven melee clip selection including `attack_2h`; suppress off-hand visual when main weapon is two-handed.

## Acceptance criteria

1. `attack_presentation.v0.json` maps handedness to clip keys.
2. `attack_2h` clip in `character_anims.tres`.
3. Off-hand gear hidden when `occupies_hands` includes off hand.
4. `make client-unit` passes.
