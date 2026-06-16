# v220 Plan - Mercenary Death Loss

Status: Ready for implementation
Goal: make hired mercenaries die, leave active companion state, and be rehired through the existing board.
Architecture: keep loss server-authoritative by applying monster combat resolution to engaged
companions and centralizing lethal hired-mercenary removal in a small helper. Reuse existing
mercenary hiring metadata for the loss event payload. Avoid client UI changes because v219 already
opens and renders the companion interface from authoritative entity state.
Tech stack: Go sim, shared world preset JSON, Python protocol bot scenario, lifecycle docs.

## Baseline and Shortcut Decision

Builds on v206 hired mercenary spawning, v208 companion commands, and v219 companion panel/roster
presentation. Asset/plugin decision: borrow existing combat lab monster rules and mercenary board
scenario patterns; reject new art, external assets, plugins, or a new recovery UI.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/sim.go` | Route monster melee attacks to an engaged companion when appropriate. |
| Add | `server/internal/game/monster_companion_combat.go` | Resolve monster damage against companions and emit hired mercenary loss. |
| Modify | `server/internal/game/mercenary_hiring_test.go` | Cover lethal loss payload/removal and rehire after loss. |
| Modify | `shared/rules/worlds.v0.json` | Add a focused mercenary loss lab using existing combat monster data. |
| Add | `tools/bot/scenarios/90_mercenary_death_loss.json` | Protocol proof for hire, loss, and rehire. |
| Add | `docs/as-built/v220_mercenary-death-loss.md` | Record proof and deferred scope. |
| Modify | `PROGRESS.md` | Advance current status after focused verification. |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 - Server Companion Damage and Loss

Files:
- Modify: `server/internal/game/sim.go`
- Add: `server/internal/game/monster_companion_combat.go`
- Modify: `server/internal/game/mercenary_hiring_test.go`

- [x] Add a deterministic monster attack target helper that prefers an engaged owned companion
  when it is in the monster's attack range.
- [x] Apply monster combat resolution to companion targets and remove hired mercenaries on lethal
  damage.
- [x] Emit `companion_damaged` / `companion_killed` combat events plus `mercenary_lost` for hired
  guard death.
- [x] Add focused Go coverage for loss payload/removal and rehire after loss.

```bash
cd server && go test ./internal/game/... -run 'TestMercenary|TestMonsterDamagesCompanion'
```

## Task 2 - Shared Lab and Bot Scenario

Files:
- Modify: `shared/rules/worlds.v0.json`
- Add: `tools/bot/scenarios/90_mercenary_death_loss.json`

- [x] Add `mercenary_death_loss_lab` using the existing mercenary board and `combat_lab_crit_attacker`.
- [x] Add a bot scenario that hires, waits for `mercenary_lost`, asserts zero living hired guards,
  then hires a replacement.

```bash
make validate-shared
make bot scenario=mercenary_death_loss
```

## Task 3 - Lifecycle Docs

Files:
- Add: `docs/as-built/v220_mercenary-death-loss.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/specs/v220_spec-mercenary-death-loss.md`

- [x] Mark the spec complete and record focused verification.
- [x] Update current status and lifecycle row.

```bash
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestMercenary|TestMonsterDamagesCompanion'`
- [x] `make bot scenario=mercenary_death_loss`
- [x] `make maintainability`

Final batch `make ci` is deferred to the enclosing `$autoloop` gate.
