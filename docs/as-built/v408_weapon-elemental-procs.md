# v408 As-built — Weapon Elemental Procs

**Slice:** v408 `weapon-elemental-procs`  
**Date:** 2026-07-02

## What it proves

- Weapon elemental hits roll seeded on-hit procs from `main_config.weapon_elemental_procs`.
- Cold freeze applies 25% movement + attack slow for 3s; fire burn DOT ticks 10% of total hit damage/sec for 10s; lightning stuns 3s; poison DOT uses rogue-style replace/refresh at 25% of elemental damage per tick.
- Mercenary/companion basic attacks apply weapon elemental hits and procs from roster main-hand gear.
- Go tests cover each proc path with 100% proc fixtures; extended bot `117_weapon_elemental_procs` regression-tests the proc lab world and split damage.

## Verification

- `go test ./internal/game/... -run 'WeaponElemental'`
- `make bot scenario=117_weapon_elemental_procs`
- `make validate-shared`
