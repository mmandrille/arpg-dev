# v438 Spec — Armor Tier-3

**Status:** Approved  
**Date:** 2026-07-04  
**Codename:** `armor-tier3`

## Purpose

Replace procedural fallback armor GLBs (head/chest/gloves/boots/amulet/ring) with CC0 external meshes.

## Non-goals

- Belt slot remains procedural `gen_glb` (no suitable CC0 belt mesh found headlessly).

## Acceptance criteria

1. Runtime GLBs for helm, chest, gloves, boots, amulet, ring updated with manifest provenance.
2. `make validate-assets` passes.
