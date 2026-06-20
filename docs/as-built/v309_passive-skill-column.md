# v309 As-Built — Passive Skill Column

Date: 2026-06-20
Spec: [`docs/specs/v309_spec-passive-skill-column.md`](../specs/v309_spec-passive-skill-column.md)
Plan: [`docs/plans/v309_2026-06-20-passive-skill-column.md`](../plans/v309_2026-06-20-passive-skill-column.md)

## What shipped

- Added a right-side passive column for every class with one-rank row 1, row 2, and row 3 nodes.
- Added `passive_stat_bonus` as a shared skill kind with server-authoritative stat bonuses.
- Added 15 class-styled passive skills, icon presentations, and English/Spanish text keys:
  - Sorcerer: Arcane Focus, Mana Weaving, Spell Dynamo.
  - Barbarian: Iron Hide, Battle Tempo, Crushing Force.
  - Paladin: Vigilant Guard, Faithful Bulwark, Consecrated Vitality.
  - Rogue: Quick Hands, Killer Instinct, Evasive Footwork.
  - Ranger: Trail Sense, Precision Draw, Deadeye.
- Applied learned passive bonuses to derived stats, passive stat totals used by skill modifiers, and stat breakdown rows with `passive_skill` source kind.
- Updated skill-panel tooltips to show passive stat effects and verified code-native icon metadata.
- Kept class-foundation scenario coverage scoped to scenario-relevant skills; passive stat-only nodes are covered by shared validation and focused Go tests.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestPassiveSkillColumn|TestClassSkillAccessGatesSpendabilityAndLearning'
make client-unit
make maintainability
.venv/bin/pytest tools -q
make ci
```

All focused checks passed in the isolated worktree. Final `make ci` passed on `main` in 13m28s after transfer.

## Boundaries

- No external assets or plugins were adopted; logos use existing code-native skill icon shapes.
- No new active abilities, protocol message types, bot scenarios, or respec changes shipped.
- Broader final active/passive skill tree expansion remains deferred.
