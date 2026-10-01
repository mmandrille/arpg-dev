# v507 Spec — Class Silhouettes

- **Status:** Complete — baseline meets visual-readability criteria; no presentation-data tuning was justified, and combined batch `make ci` passed.
- **Date:** 2026-10-01
- **Codename:** `class-silhouettes`
- **Accepted batch:** v501–v510; reserved slice v507
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc` (detached worktree)
- **Dependency:** v503 Ground Gear Models must be integrated or its touched paths transferred and reconciled before v507 implementation if the surfaces overlap.

## Purpose

Make the five KayKit hero classes and their equipped looks easier to distinguish at the actual isometric play-camera size. Preserve the existing rig, socket fit, attack silhouettes, first-person body path, and server-owned equipment state. This is a client presentation slice: the same item and class still have the same gameplay effect.

## Baseline and boundaries

- `class_presentations.v0.json` maps Barbarian, Sorcerer, Paladin, Rogue, and Ranger to committed Adventurers 2.0 GLBs, all at model scale 0.8. `armor_look.v0.json` recolors body/arms/legs and toggles native headgear for equipped head items. `item_visuals.v0.json` owns weapon/off-hand mounts. `EquipmentVisualResolver` applies equipment on the local hero and first-person rig.
- Inspection of committed GLB mesh nodes: Knight has cape, helmet, and visor; Barbarian has a bear hat; Mage has cape and hat; Rogue has cape; Ranger has cape and quiver. Rogue and Ranger have **no native headgear**. The staged `Rogue_Hooded` noted in research is not a committed runtime asset. The v483 `gear` images use the same sword/shield on every class and obscure differences; they are a fit baseline, not a play-camera readability result.
- ADR-0018 D5 keeps torso/limb armor as tints, head as an existing accessory toggle, and weapons/shields as socket meshes. This slice may vary class model scale, stance, visibility of existing class-owned mesh parts, tint strength/palette, and existing equipped visual selection only through schema-backed presentation data. It will not create attached torso/limb armor or alter server equipment semantics.

## Acceptance criteria

1. **Visible class distinction.** Real-renderer, same-camera before/after captures include all five classes both without equipment and in a class-appropriate equipped set, at normal play zoom and one tighter battle view. At normal zoom, the main outline and class-specific accessory/weapon remain visibly different from adjacent classes when viewed in grayscale as well as color. The as-built names any class pair that remains ambiguous rather than claiming a full pass from isolated close-ups.
2. **Equipped silhouette.** Equipping and unequipping head, chest, glove, boot, main-hand, and off-hand items updates the intended current presentation with no stale nodes, hidden body parts, or clipped weapon/head accessories. Existing armor tint precedence and all equipment outcomes remain unchanged. Rogue/Ranger head items may still lack a world head mesh; this limit is explicit in the as-built.
3. **Animation and cameras.** Idle, walk, basic attack, hit, and death preserve the intended outline and hand grip for representative melee, bow, staff, and shield setups. First-person hands/weapon path remains usable; any adjusted model scale or accessory visibility must not move the camera into the head or conceal the attack.
4. **Data and asset integrity.** New or changed visual parameters live in schema-backed `shared/assets/` catalogs; all referenced mesh names and asset IDs resolve on the five committed GLBs and manifest. No runtime download, new art family, plugin, gameplay rule, protocol, server, or replay change.
5. **Focused proof.** Shared/asset validation, focused Godot presentation tests and fit probes, one live Godot client bot equipment scenario, and before/after real-renderer screenshots pass. Record a matched renderer cost comparison after v503 integration if a model or draw-bearing accessory is added or made persistently visible. Screenshots alone do not prove frame performance or combat feel.

## Likely files and checks

| Surface | Likely change |
|---|---|
| Shared presentation | `shared/assets/class_presentations.v0.json` and schema; `armor_look.v0.json` and schema only where needed for class-owned visible parts or equipped tint |
| Client | `client/scripts/class_presentations_loader.gd`, a focused class-silhouette presenter if required, `character_visual.gd`, `armor_look.gd`, and the local model-swap call sites only when needed; preserve `equipment_visuals.gd` ownership of equipment |
| Assets | Reuse committed `client/assets/characters/kaykit/*.glb`, existing rig-native equipment and `assets/manifests/assets.v0.json`; no vendoring planned |
| Tests and capture | `client/tests/test_armor_look.gd`, `client/tests/test_animation.gd` or a focused new test, `client/tests/equipped_gear_fit_probe.gd`, the existing client equipment bot scenario, `showme` gear/skeleton suites, and a play-camera capture when isolation is insufficient |
| Docs | `docs/CODEMAP.md` for new files; `docs/as-built/v507_class-silhouettes.md` and paired images after implementation; coordinator later updates lifecycle/PROGRESS |

The existing `bone_gear_sockets` Godot client bot scenario provides the live equip path. Run it as `make bot-visual scenario=97_bone_gear_sockets` for visual inspection, and use `make bot-client SCENARIO=97_bone_gear_sockets HEADLESS=1` for the automated equip assertion. Extend or add a focused scenario only if the chosen look needs a missing semantic assertion. Capture the five-class matrix with `make regen-screenshots SUITE="skeleton gear"`; add a normal-zoom play-camera comparison if the isolated focus overstates legibility. Inspect PNGs under `.artifacts/screenshots/latest/` and preserve selected paired images in the as-built.

## Asset and plugin decision

- **Adopt:** the five already committed KayKit Adventurers 2.0 hero GLBs and their native mesh parts, plus committed KayKit weapons/shields. Their manifest provenance and CC0 license are already recorded.
- **Borrow:** the existing class presentation loader, ArmorLook, equipment resolver, socket catalog, rig clips, screenshot suite, and fit probe.
- **Reject:** external assets/plugins, runtime downloads, uncommitted staged models, a second body rig, new separate armor meshes, and a new asset pipeline. Any later asset adoption needs its own provenance and ADR-0018 review.

## Dependencies and integration risks

- v503 Ground Gear Models may touch `item_presentations`, `equipment_display`, manifest/asset validation, showme captures, or item model tests. Recheck its integrated diff before touching these files; v507 must build on the final v503 mapping. Implementation is held until the coordinator gives that integrated state or transferred changes.
- Persistent class accessories or model-scale edits can change shadow/camera occlusion and first-person framing. Evaluate at the real play camera and during attack, not only in an isolated T-pose.
- `showme/visual_capture.gd` and `main.gd` are grandfathered over the 600-line target. Any edit must leave them no larger than their current baselines or extract a coherent helper; do not grow them casually.

## Non-goals

- New class models, headgear geometry for Rogue/Ranger, paper-doll redesign, equipment replication on remote players/mercenaries, first-person-only bespoke animations, gameplay balance, loot generation, protocol changes, or server changes.
- Claiming a measurable readability or performance gain from a static image alone. The as-built must separate visual judgment, automated proof, and performance measurement.

## Implementation and evidence outcome

The actual gameplay-camera baseline was captured for all five classes, empty and class-appropriately equipped, at normal and battle zoom in color and grayscale. Attack samples at 20%, 50%, and 80% were also reviewed. The current assets meet the visual-readability criteria, including grayscale distinction; Rogue and Ranger are the closest unequipped pair but remain distinguishable. Therefore no class-presentation or armor-look catalog tuning was made. The existing Rogue/Ranger native-headgear limitation remains.

See [the v507 as-built](../as-built/v507_class-silhouettes.md) for the capture matrix, class-by-class findings, test results, and the default-delay visible-bot timeout caveat. Capture evidence remains in ignored local artifact directories; no renderer-cost comparison was needed because no persistent geometry or accessory visibility changed.

## Planning questions

None blocked the review. Matched actual-camera captures resolved the tuning question without requiring new data fields.
