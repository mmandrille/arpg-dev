# v220 Spec - Mercenary Death Loss

Codename: `mercenary-death-loss`
Status: Complete
Baseline: v219 `companion-stance-ui`

## Purpose

Make hired mercenaries losable in combat. A hired guard that takes lethal monster damage should be
removed from the active companion roster, emit an authoritative `mercenary_lost` event, and require
the player to hire a replacement through the existing mercenary board flow.

This closes the first death/loss rule for v206-v219 mercenaries while keeping hiring, gold spend,
and companion state server-owned.

## Non-goals

- No durable roster, gear snapshot, market listing, or offline mercenary persistence.
- No refund, revive, cooldown, insurance, or recovery UI for a lost hire.
- No changes to player death, corpse, gold persistence, or companion stance semantics.
- No production art, new client UI, external assets, or plugins.
- No broad monster aggro rewrite beyond the minimal ability for monsters to hit an engaged
  companion.

## Adopt / Borrow / Reject

- **Adopt:** Existing hired mercenary spawn metadata (`mercenary_hire`, `fixed:mercenary_guard`) and
  companion entity roster updates.
- **Borrow:** Existing `combat_lab_crit_attacker` rules and mercenary hiring lab patterns for a
  deterministic bot proof.
- **Reject:** New art/plugin dependencies, new durable storage tables, and a separate loss recovery
  interface.

## Acceptance Criteria

1. Monster attacks can damage an owned companion that is engaged in front of the player.
2. When a hired mercenary reaches zero HP from monster damage, the server emits `mercenary_lost`
   with the lost companion id, killer source id, `service=mercenary`, `offer_id=fixed:mercenary_guard`,
   and `monster_def_id=mercenary_guard`.
3. The lost hired mercenary is removed from active entity state, so the companion roster no longer
   contains a living `mercenary_guard`.
4. The same player can rehire from the existing mercenary board after loss, spending gold through
   the existing `mercenary_hired` path and spawning a new guard.
5. Focused Go tests cover lethal companion damage, loss event payload, entity removal, and rehire.
6. A protocol bot scenario proves hire -> combat loss -> no active mercenary -> rehire.

## Scope and Files Likely Touched

- `server/internal/game/sim.go`
- `server/internal/game/monster_companion_combat.go`
- `server/internal/game/mercenary_hiring_test.go`
- `shared/rules/worlds.v0.json`
- `tools/bot/scenarios/90_mercenary_death_loss.json`
- `docs/as-built/v220_mercenary-death-loss.md`
- `PROGRESS.md`
- `docs/progress/slice-lifecycle.md`

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'TestMercenary|TestMonsterDamagesCompanion'`
- `make bot scenario=mercenary_death_loss`
- `make maintainability`

Manual visual check, if desired:

```bash
make bot-visual scenario=mercenary_death_loss
```

## Open Questions and Risks

- None. Loss is limited to active hired mercenary entities; durable mercenary rosters remain future
  ADR-0010 work.
