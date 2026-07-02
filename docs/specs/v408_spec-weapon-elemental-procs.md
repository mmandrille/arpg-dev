# v408 Spec — Weapon Elemental Procs

Status: Complete
Date: 2026-07-02
Codename: `weapon-elemental-procs`
Baseline: v407 `weapon-elemental-damage` complete

## Purpose

After a successful weapon elemental hit (v407), roll seeded on-hit procs:
- **Cold (10%):** 25% movement + attack slow for 3s
- **Fire (10%):** burn DOT — 10% of combined hit damage per second for 10s
- **Lightning (5%):** stun for 3s
- **Poison (10%):** poison DOT — % of elemental damage dealt, rogue-style replace/refresh

Mercenary basic attacks apply elemental hits and procs from their equipped weapon.

## Acceptance criteria

- [x] Proc tuning lives in `shared/rules/main_config.v0.json`
- [x] Procs roll per successful elemental hit using session seeded RNG
- [x] Cold/fire/lightning/poison procs match approved defaults
- [x] Mercenary companion attacks apply elemental hits + procs
- [x] Focused Go tests + extended bot scenario pass
