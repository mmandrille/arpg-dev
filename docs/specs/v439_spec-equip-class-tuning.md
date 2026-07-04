# v439 Spec — Equip Class Tuning

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `equip-class-tuning`

## Purpose

Per-class `local_transform` overrides in `item_visuals.v0.json` and class-aware resolution in `EquipmentVisualResolver`.

## Acceptance criteria

1. Optional `class_transforms` on item visual entries (schema-backed).
2. `resolver.set_character_class()` merges class overrides when mounting gear.
3. `make validate-shared` and `make client-unit` pass.
