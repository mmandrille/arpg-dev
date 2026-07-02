# v407 As-built — Weapon Elemental Damage

**Slice:** v407 `weapon-elemental-damage`  
**Date:** 2026-07-02

## What it proves

- Weapon rolls expose at most one `bonus_{cold,fire,lightning,poison}_damage` affix with ilvl-1 range 1–6.
- Physical basic attacks use `damage_type: force`; a second flat typed hit applies the rolled elemental amount after a successful physical hit.
- Monster resistances apply to the elemental portion only (cold-resistant lab target halves cold damage).
- Item summaries show "+X {Element} Damage" instead of folding elemental into physical range.
- Lab worlds can pin loot via `loot_preset` on world entities for deterministic bot proofs.

## Key decisions

- Reused existing stat keys and v100 resistance model; no protocol schema bump.
- Elemental affix rolls are mutually exclusive on generate/reroll.
- Mercenary/companion elemental hits and on-hit procs deferred to v408.

## Verification

- `go test ./internal/game/... -run 'WeaponElemental|ElementalAffix|ItemSummaryShowsElemental'`
- `make bot scenario=116_weapon_elemental_damage` (extended tier)
- `make validate-shared`

## Scope limits

- Basic attacks and player projectiles only; skills unchanged.
- No elemental procs (v408).
