# v191 Spec - Affix Name Grammar

Status: Approved
Date: 2026-06-15
Codename: affix-name-grammar

## Purpose

Rolled non-unique equipment should expose readable affix-style item names instead of only rarity
prefixes such as `Rare Cave Blade`. The first slice adds deterministic prefix/suffix words derived
from rolled stat families so inventory, stash, shop, market, and bot views can identify why a rare
item is interesting without inspecting raw stat payloads.

## Non-goals

- No procedural item ID changes, market compatibility layer, or persistence migration.
- No localization catalog rollout for generated affix words.
- No multi-affix grammar, legendary naming, or player-editable names.
- No unique or set item rename changes; named uniques and set pieces keep their authored names.

## Acceptance Criteria

- Rare or magic rolled items generated through the existing server roll path receive deterministic
  display names based on their rolled stat families.
- Names remain durable in `ItemRollPayload.display_name` and flow through existing inventory, stash,
  loot, shop, and bot views without a protocol schema change.
- Unique and set item display names remain authored by their existing unique/set payload logic.
- A focused Go test proves a rare skill-affix staff receives an affix grammar name.
- Protocol bot scenario `80_skill_affix_rolls.json` asserts the generated display name while
  preserving its stat-key proof.

## Scope and Files

- Server: `server/internal/game/shop.go`, focused item rarity/name tests.
- Bot: `tools/bot/scenarios/80_skill_affix_rolls.json`.
- Docs: this spec, the v191 plan, `PROGRESS.md`, and an as-built note.

## Test and Bot Proof

- `cd server && go test ./internal/game -run 'TestAffixName|TestSkillAffix' -count=1`
- `make bot scenario=80_skill_affix_rolls.json`
- `make ci`

## Open Questions and Risks

- The first grammar intentionally chooses one strongest stat family as a prefix. Rich multi-affix
  names and localized text remain future itemization work.
