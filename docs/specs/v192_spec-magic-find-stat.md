# v192 Spec - Magic Find Stat

Status: Approved
Date: 2026-06-15
Codename: magic-find-stat

## Purpose

Add Magic Find as a gear stat that gives loot-oriented builds a visible reason to keep hunting and
equipping off-DPS items. Equipped `magic_find_percent` should be visible in derived stats and should
server-authoritatively bias monster equipment roll rarity upward without changing item ownership or
shop stock generation.

## Non-goals

- No final loot economy tuning or endgame Magic Find cap.
- No Magic Find effect on fixed drops, gold amount, upgrade resources, shop stock, mystery seller,
  chest content, or unique/set special rules.
- No new UI panel; existing stat breakdown/inventory/client labels should render the new stat.
- No protocol version bump unless required by existing schemas.

## Acceptance Criteria

- `magic_find_percent` is a valid rollable equipment stat and can appear on a deterministic item
  used by the protocol bot proof.
- Equipped Magic Find appears in `derived_stats.magic_find_percent` and stat breakdowns.
- Monster-dropped rolled equipment uses the player's current Magic Find to bias random rarity
  selection toward non-common rarity; focused Go tests prove the adjusted rarity weights and the
  zero-Magic-Find baseline path.
- Generated shop offers still use the baseline rarity weights and do not inherit player Magic Find.
- Bot scenario `81_magic_find_stat.json` equips a Magic Find item and proves the derived stat and
  stat breakdown over protocol.

## Scope and Files

- Shared: item-template schema/rules and protocol schemas if derived stats require it.
- Server: item roll rarity selection, derived stats/stat breakdowns, focused tests.
- Client/tools: stat label, bot scenario/assertion support if required.
- Docs: plan, as-built, `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'MagicFind|ItemRollsGolden|ShopGeneratedOfferGolden' -count=1`
- `make bot scenario=81_magic_find_stat.json`
- `make ci`

## Open Questions and Risks

- The first formula is intentionally modest and deterministic: Magic Find adds relative weight to
  magic-or-higher random rarity entries for monster item-template rolls only. Final economy tuning
  remains deferred.
