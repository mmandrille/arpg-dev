# Spec: `class-skill-build-branches`

Status: Complete  
Date: 2026-07-03  
Codename: class-skill-build-branches  
Slice batch: v420–v424 (one class per execution slice)

## Purpose

Add **30 new active skills** (six per class: three slot-A forks + three slot-B capstones) so each
class supports multiple divergent build paths. All skills reuse **existing** `kind` and effect
payloads from other classes — no new Go/GDScript skill-engine primitives.

Implementation splits into five vertical slices:

| Slice | Class | New skill IDs |
|-------|-------|---------------|
| v420 | barbarian | `rampage`, `shatter_strike`, `battle_roar`, `worldbreaker`, `blood_frenzy`, `gore_strike` |
| v421 | sorcerer | `glacial_lance`, `chain_storm`, `renewing_light`, `inferno`, `arcane_overload`, `arcane_renewal` |
| v422 | paladin | `blessed_recovery`, `avenging_light`, `bulwark_aura`, `divine_hammer`, `sacred_ground`, `righteous_fury` |
| v423 | rogue | `blade_dance`, `venom_spray`, `shadow_veil`, `death_blossom`, `assassinate`, `killing_mark` |
| v424 | ranger | `alpha_call`, `pinning_volley`, `hunters_volley`, `pack_master`, `meteor_shot`, `arrow_storm` |

Per-class active count rises from **7 → 13** (still 1 mobility, 4 passives, 1 survival).

## Non-goals

- New skill-engine kinds or protocol/schema version bump.
- Combined volley+root on one projectile (`pinning_volley` uses volley only).
- Production VFX/audio (borrow existing presentation families).
- Full combat balance pass; respec UX; CI pack promotion (extended bot labs only).

## Acceptance criteria (batch)

- [ ] All 30 skills in `skills.v0.json` with tree placement, prerequisites, synergies, i18n, presentations.
- [ ] `validate_skills.py` expects 13 actives per class.
- [ ] Focused Go tests prove one representative skill per class (minimum); barbarian bleed capstone in bot lab.
- [ ] Extended bot scenario `class_build_branches_lab` casts at least one new skill per class (or class-specific labs).
- [ ] Skill tree layout renders without false stacks (v419 resolver).
- [ ] `make validate-shared` green after each slice; final autoloop `make ci` green.

## Client asset decision

**Borrow** existing icon shapes and effect visuals from sibling skills in each branch.

## Skill catalog

See plan files per slice for exact tuning copied from template skills.
