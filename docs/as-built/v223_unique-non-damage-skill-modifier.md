# v223 As-Built - Unique Non-Damage Skill Modifier

Date: 2026-06-16

## What shipped

- Added the named unique `bloodbound_sigil` / `Bloodbound Sigil`, based on `cave_ring`, with fixed
  `max_hp`, `max_mana`, level 5 requirements, and the existing live `blood_price` effect.
- Extended named unique payload coverage so Bloodbound Sigil has stable template, display name,
  fixed stats, requirements, rarity, and effect IDs.
- Added a server behavior proof that equips the named item package and casts Magic Bolt at zero mana,
  paying the missing skill cost with HP through Blood Price.
- Added a protocol bot scenario that opens the purple town unique chest, takes Bloodbound Sigil, and
  asserts the rolled unique inventory payload exposes `effect_ids: ["blood_price"]`.
- Extended the client unique chest proof so the stash panel presents the Bloodbound Sigil row with
  the Blood Price summary.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'NamedUnique|BloodPrice|UniqueTestChest' -count=1
make bot scenario=unique_non_damage_skill_modifier
make bot-client scenario=unique_chest_client_proof
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
remains deferred until the selected feature queue is complete.

Manual visual proof, if desired:

```bash
make bot-visual scenario=unique_chest_client_proof
```

## Scope limits

- No new unique-effect hook, skill formula, active skill, resource type, protocol field, or client UI
  system shipped.
- No balancing pass, market restrictions, binding rules, production art, or production audio shipped.
