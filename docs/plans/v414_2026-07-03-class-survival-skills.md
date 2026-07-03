# v414 Plan — Class survival skills

Status: Complete  
Goal: Ship five auto-proc Survival skills with PvE lethal trigger and PvP guard.  
Architecture: New `survival_autocast` kind + `survival_skills.go` lethal hook; reuse immunity/mark/reflect patterns; widen skills panel for column 6.  
Tech stack: shared JSON, Go sim, Python bot, Godot client (panel + presentations).

## Baseline and shortcut decision

Builds on v413 mark/bleed engine. **Borrow** holy-shield/sanctuary/heal VFX families.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` + schema | branch, kind, five skills |
| Create | `server/internal/game/survival_skills.go` | proc gate, effects, PvP guard |
| Modify | `server/internal/game/rules.go` | load/validate survival |
| Modify | `server/internal/game/sim.go` | immunity generalization, phasing, outgoing mult |
| Modify | `server/internal/game/unique_survival_effects.go` | lethal hook delegation |
| Create | `server/internal/game/survival_skills_test.go` | per-class + PvP guard |
| Modify | `client/scripts/skills_panel.gd` | panel width column 6 |
| Modify | `shared/i18n/en.json`, `skill_presentations.v0.json` | copy |
| Create | `tools/bot/scenarios/class_survival_skills.json` | extended PvE proc |
| Create | `docs/as-built/v414_class-survival-skills.md` | proof |

## Maintenance ratchet

Hotspot: `sim.go` — touch-to-shrink; survival logic in new file.

## Task 1 — Shared schema + skills data

- [x] Schema branch, survival_autocast, survival effect types
- [x] Five skills + i18n + presentations
```bash
make validate-shared
```

## Task 2 — Go survival engine

- [x] `survival_skills.go`: proc, effects, cooldown
- [x] PvP guard, HP floor, mana shield, redirect, phasing, immunity, outgoing mult
- [x] Reject manual cast for survival_autocast
```bash
cd server && go test ./internal/game -run TestSurvival -count=1
```

## Task 3 — Client panel

- [x] Widen skills panel; branch label for survival (width only; badge deferred)
```bash
make client-unit
```

## Task 4 — Bot scenario

- [x] `class_survival_skills.json` extended
```bash
make bot scenario=class_survival_skills
```

## Task 5 — Docs

- [x] as-built + lifecycle on finish

## Final verification

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game -run TestSurvival -count=1
make bot scenario=class_survival_skills
make client-unit
```
