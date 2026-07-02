# v408 Plan — Weapon Elemental Procs

Status: Complete
Goal: Add data-driven weapon elemental on-hit procs and mercenary parity.
Architecture: Extend weapon elemental module with proc helpers; config in main_config; reuse slow/root/burn/poison infrastructure.
Tech stack: Go sim, shared JSON rules, Python protocol bot.

## Task 1 — Shared config

- [x] Step 1.1: Add `weapon_elemental_procs` to `main_config.v0.json` + schema
```bash
make validate-shared
```

## Task 2 — Server proc application

- [x] Step 2.1: Implement proc rolls and status application after elemental hits
- [x] Step 2.2: Monster attack-speed slow from freeze
- [x] Step 2.3: Mercenary elemental hits + procs
```bash
cd server && go test ./internal/game/... -run 'WeaponElemental' -count=1
```

## Task 3 — Bot + docs

- [x] Step 3.1: `117_weapon_elemental_procs.json` (`ci_tier: extended`)
- [x] Step 3.2: PROGRESS, lifecycle, as-built
```bash
make bot scenario=117_weapon_elemental_procs
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] Focused Go tests + bot scenario
