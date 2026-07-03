# v414 As-Built — Class Survival Skills

Date: 2026-07-03
Spec: [`docs/specs/v414_spec-class-survival-skills.md`](../specs/v414_spec-class-survival-skills.md)
Plan: [`docs/plans/v414_2026-07-03-class-survival-skills.md`](../plans/v414_2026-07-03-class-survival-skills.md)

## What shipped

- New `survival_autocast` skill kind and **Survival** branch (column 6) with one skill per class.
- **Second Wind**, **Arcane Barrier**, **Divine Protection**, **Evasive Stance**, **Spectral Path** — level 10, rank paid by skill point, 120s CD, 0 mana.
- PvE lethal auto-proc: floors HP to 1, activates effect; **PvP never procs**.
- Server module `survival_skills.go`: proc gate, per-class effects, validation.
- Generalized immunity/outgoing-damage/phasing/redirect/evade hooks in sim.
- Skills panel widened for column 6; i18n and skill presentation entries.

## Proof

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run TestSurvival -count=1
make bot scenario=class_survival_skills
make client-unit
```

## Boundaries

- Lethal proc timing and per-class effect math covered by Go tests; bot scenario proves rank-1 spend only.
- Extended-only bot scenario; not in CI pack.
- Survival branch label/badge in panel deferred (width + data wiring only).
- Stronger bot lethal-proc proof deferred (needs lab damage setup).
