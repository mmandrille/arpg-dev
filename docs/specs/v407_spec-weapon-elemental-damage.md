# v407 Spec — Weapon Elemental Damage

Status: Complete
Date: 2026-07-02
Codename: `weapon-elemental-damage`
Baseline: v406 `bishop-force-drop` complete

## Purpose

Weapons can roll one elemental affix (`cold`, `fire`, `lightning`, or `poison`) shown as
**"+X {Element} Damage"**. Basic attacks deal **physical (`force`) damage** from the weapon's
`damage_min`/`damage_max` roll plus a **separate typed elemental hit** equal to the rolled bonus.
Monster resistances apply to the elemental portion. Elemental procs are deferred to v408.

## Non-goals

- Elemental on-hit procs (freeze, burn, stun, poison DOT) — v408
- Mercenary/companion elemental hits — v408
- Skill damage elemental split
- New damage types beyond existing five
- Affix grammar / procedural naming changes
- Protocol schema version bump (reuse `damage_type` on separate `monster_damaged` events)

## Acceptance criteria

- [x] Weapon templates roll at most one `bonus_{element}_damage` affix; ilvl-1 range is 1–6 (v389 scaling applies at higher ilvl)
- [x] Elemental bonus is excluded from physical `damage_min`/`damage_max`; summary shows "+X Cold Damage" (etc.)
- [x] Successful basic attacks emit `monster_damaged` with `damage_type: force` for physical and a second event with the elemental type and flat amount
- [x] Elemental damage respects monster resistances (v100 combat lab pattern)
- [x] Lab world can spawn loot with a fixed roll payload for deterministic bot proof
- [x] Focused Go tests + protocol bot scenario pass

## Scope and likely files

- `shared/rules/item_templates.v0.json` — elemental affix ilvl-1 max 6
- `shared/rules/worlds.v0.json` + schema — `loot_preset` on lab loot entities; `weapon_elemental_lab`
- `shared/rules/monsters.v0.json` — optional cold-resistant lab target
- `server/internal/game/weapon_elemental.go` — split damage helpers (new)
- `server/internal/game/sim.go`, `handlers.go`, `affix_names.go`, `item_rolls.go`, `shop.go`
- `tools/bot/scenarios/116_weapon_elemental_damage.json`
- Docs: plan, as-built, lifecycle

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'WeaponElemental|ElementalAffix' -count=1
make bot scenario=116_weapon_elemental_damage
```

## Open questions and risks

- None blocking; procs and mercenary parity land in v408.
