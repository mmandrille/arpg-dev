# Spec: `combat-threat-readability`

Status: Accepted
Date: 2026-06-11
Codename: `combat-threat-readability`
Slice: v80 - combat threat readability
Baseline: v79 `elite-pack-roles`

## Purpose

Pack aggro now makes nearby monsters join fights, but the client mostly shows that as movement after
the fact. This slice makes existing authoritative combat events more readable in the moment:
monsters that aggro show a short threat pop-up, and floating combat text exposes threat-relevant
variants consistently for miss, block, crit, heal, skill reject, and damage outcomes.

The change is presentation-only. The server remains authoritative for aggro, combat outcomes, and
all event metadata.

## Non-goals

- No server combat, aggro, AI, pack generation, or protocol/schema changes.
- No new monster intent, threat table, targeting UI, minimap marker, audio, or production VFX.
- No balance tuning.

## Acceptance criteria

1. Existing `monster_aggro` events produce floating text on the aggroing monster.
2. Aggro text is visually distinct from damage text and uses a stable bot-visible variant.
3. Existing damage-number variants remain intact for miss, block, crit, normal damage, heal, mana,
   and skill rejection text.
4. Floating text still respects the existing user setting that disables combat text.
5. Bot debug state exposes the aggro text through the existing damage-number list.
6. A focused client-bot scenario proves pack aggro creates at least one visible aggro text pop-up.
7. Existing protocol bot pack aggro proof still passes.

## Scope and likely files

- `client/scripts/main.gd` - map `monster_aggro` events to floating threat text.
- `client/scripts/damage_number.gd` - add display treatment for a threat variant.
- `client/tests/test_coop_client.gd` - unit proof that `monster_aggro` spawns the expected text and
  respects the setting.
- `client/tests/test_client_bot.gd` - scenario runner validation for the new client-bot proof if
  needed.
- `tools/bot/scenarios/client/31_combat_threat_readability.json` - client-bot proof.
- `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json` - existing protocol regression stays
  as server proof.
- `PROGRESS.md` and `docs/as-built/v80_combat-threat-readability.md` - lifecycle close-out.

## Test and bot proof

- `make client-unit`
- `SCENARIO=31_combat_threat_readability HEADLESS=1 ./scripts/bot_client_local.sh`
- `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`
- `make ci`

## Open questions and risks

- Decision: use the compact text `AGGRO` and variant `threat` for now. This is clear enough for
  headless proof and avoids adding localization/UI copy infrastructure.
- Risk: The aggro pop-up can disappear before a client-bot assertion if the scenario waits too
  long. Mitigation: wait on `monster_aggro` and immediately assert a `threat` damage number through
  `get_bot_state()`.
