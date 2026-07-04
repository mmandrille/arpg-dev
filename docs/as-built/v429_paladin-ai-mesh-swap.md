# v429 As Built — Paladin AI Mesh Swap (Evaluation)

Date: 2026-07-04

## Shipped

- Swapped `assets/characters/paladin/knight.glb` to CC0 Poly Pizza **Warrior** (mastjie); archived prior mesh as `knight_legacy.glb`.
- Re-rigged runtime `client/assets/characters/paladin/paladin.glb` (~488 KB vs ~6.2 MB legacy runtime).
- Updated `character_paladin_v0` manifest provenance (CC0-1.0, Poly Pizza origin URL, new `sha256`).
- Paladin README documents AI/Meshy handoff at the same source path.

## Verification

```bash
make validate-assets
.venv/bin/pytest tools/assets/test_rig_hero_glbs.py::test_paladin_runtime_glb_has_required_bones -q
SCENARIO=92_paladin_tier3_visual HEADLESS=1 ./scripts/bot_client_local.sh
python3 skills/showme/scripts/render_focus.py --focus gear --class-id paladin
```

Showme capture: `.artifacts/showme/20260704-020213-gear.png`

Visual replay: `make bot-visual scenario=92_paladin_tier3_visual`

## Decision note

**Adopt** CC0 external mesh via Tier-3 pipeline (permanent paladin body). Other classes: use autoloop prompt in [`docs/researchs/hero-tier3-mesh-autoloop-prompt.md`](../researchs/hero-tier3-mesh-autoloop-prompt.md).
