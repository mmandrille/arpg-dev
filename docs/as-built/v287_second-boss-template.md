# v287 As Built: Second Boss Template

Date: 2026-06-19
Spec: [`docs/specs/v287_spec-second-boss-template.md`](../specs/v287_spec-second-boss-template.md)
Plan: [`docs/plans/v287_2026-06-19-second-boss-template.md`](../plans/v287_2026-06-19-second-boss-template.md)

## What shipped

- Added `crypt_matron` as a second boss template using the existing `dungeon_undead` base monster,
  skeleton visual, boss loot table, enrage rules, and authored boss pattern deck.
- Added `crypt_matron` to the boss-floor template pool while preserving Cave Warden data.
- Boss-floor generation now selects from the configured template pool deterministically by seed and
  absolute dungeon level.
- Added focused Go coverage proving known Cave Warden seeds remain stable and
  `second_boss_template` selects Crypt Matron with the expected boss entity fields.
- Relaxed shared boss visual validation from the old Cave Warden placeholder to a schema-owned
  allowlist covering existing monster scenes plus the existing humanoid boss special case.
- Added protocol bot scenario `second_boss_template`, which proves the Crypt Matron spawn, summon-bat
  pattern, boss kill, and exit unlock.

## Proof

Focused verification:

```bash
(cd server && go test ./internal/game -run 'TestBoss' -count=1)
make validate-shared
make bot scenario=second_boss_template
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: deferred until the end of the selected autoloop queue.

## Manual visual command

```bash
make bot-visual scenario=second_boss_template
```

## Deferred

- Weighted boss pools, multi-boss floors, new boss arenas, new pattern primitives, and production boss
  art remain deferred.
- Additional boss templates beyond Cave Warden and Crypt Matron remain future content work.
