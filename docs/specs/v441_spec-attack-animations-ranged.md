# v441 Spec — Attack Animations Ranged

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `attack-animations-ranged`

## Purpose

Bow string-pull and staff cast wind-up clips driven by `attack_presentation` data.

## Acceptance criteria

1. `attack_ranged` and `attack_staff` clips in `character_anims.tres`.
2. Ranger bow / sorcerer staff select ranged clips via loader.
3. `make client-unit` passes.
