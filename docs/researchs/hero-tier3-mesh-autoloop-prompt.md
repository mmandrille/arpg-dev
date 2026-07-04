# Autoloop prompt — Hero Tier-3 mesh swap (one slice per class)

Copy everything inside the fenced block below and paste it as your `/autoloop` command.

```text
/autoloop Execute four Tier-3 hero body mesh swap slices (barbarian, rogue, sorcerer, ranger) using the v429 paladin proof as the template. Paladin is already done — do not re-slice it.

## Goal

For each remaining class, swap the static source GLB under assets/characters/<class>/ for a better external/CC0 mesh (Poly Pizza or equivalent), re-rig with tools/assets/rig_hero_glbs.py, update manifest provenance, and prove equipped gear + animation bones in client. One committed slice per class (v430–v433).

## Non-negotiable class theming

Each hero must read as its class at a glance — silhouette and height matter as much as texture. Do not pick five generic humanoids at the same height.

| Class | Visual identity | Target height (rig) | Mesh selection cues |
|-------|-----------------|---------------------|---------------------|
| **Barbarian** | Biggest, strongest, widest shoulders | **~1.95–2.00 m** (`HERO_TARGET_HEIGHTS`) | Goliath, berserker, tribal warrior; heavy limbs, bare or fur-trimmed armor |
| **Paladin** | *(done v429)* | ~1.85 m | Armored knight / crusader — reference slice only |
| **Rogue** | Shortest, lean, sneaky | **~1.68–1.72 m** | Assassin, hood, light leather; narrow frame; pair with existing `idle_stance` lean (~4°) |
| **Sorcerer** | Slender, robed caster | **~1.78–1.82 m** | Mage, wizard, staff-ready posture; robes/cloak volume |
| **Ranger** | Medium, agile archer | **~1.80–1.84 m** | Hood, quiver/bow silhouette; keep ranger rest-pose behavior in rig_hero_glbs |

After rigging, sanity-check in showme (`render_focus.py --focus gear --class-id <class>`) that the class is visually distinct from the others side-by-side (`--focus classes`).

## Per-slice checklist (mirror v429)

1. Read docs/researchs/agent-visual-improvement-playbook.md and docs/as-built/v429_paladin-ai-mesh-swap.md.
2. Archive current source as `<name>_legacy.glb` before replacing.
3. Probe candidate: `python3 skills/3dmodel/scripts/create_model_probe.py --model <path> --key <class>_tier3`.
4. Replace source GLB; tune `HERO_TARGET_HEIGHTS["<class>"]` if needed (not uniform 1.85 for all).
5. `python3 tools/assets/rig_hero_glbs.py` → update `assets/manifests/assets.v0.json` (origin URL, license, sha256).
6. `cd client && godot --headless --import`; commit extracted textures/import sidecars.
7. Class README under `assets/characters/<class>/` with swap + AI handoff steps.
8. Extended client bot scenario (new `ci_tier: extended` unless replacing an existing proof): class + starter gear equip + `visual_model: character`.
9. `make validate-assets`, bone contract test, client scenario, showme gear capture.
10. Spec, plan, as-built, PROGRESS.md + slice-lifecycle row; commit `feat: vN: <class> tier-3 mesh swap`.

## Constraints

- No new git branches. No CI pack promotion unless budget-neutral demotion elsewhere.
- Static source GLB only (no embedded skin/animations — rig tool rejects skinned sources).
- Record adopt/borrow/reject + license in each spec; reject runtime network fetch.
- Clean up `_model_probe` sandbox files before validate-assets.
- Do not duplicate paladin work; barbarian → rogue → sorcerer → ranger order is fine.

## Current sources (replace these)

- barbarian: assets/characters/barbarian/goliath_barbarian.glb
- rogue: assets/characters/rogue/assasine.glb
- sorcerer: assets/characters/sorcerer/mage.glb
- ranger: assets/characters/ranger/green_hood.glb

Run post-loop `make ci` once after all four slices land.
```

## Reference

- Template slice: v429 paladin — [`docs/as-built/v429_paladin-ai-mesh-swap.md`](../as-built/v429_paladin-ai-mesh-swap.md)
- Playbook: [`agent-visual-improvement-playbook.md`](agent-visual-improvement-playbook.md)
- Rig heights: `tools/assets/rig_hero_glbs.py` → `HERO_TARGET_HEIGHTS`
- Presentation lean/scale: `shared/assets/class_presentations.v0.json` → `idle_stance`
