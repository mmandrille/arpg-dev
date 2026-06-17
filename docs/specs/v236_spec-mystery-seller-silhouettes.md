# v236 Spec - Mystery Seller Silhouettes

Status: Approved for autoloop
Date: 2026-06-17
Codename: mystery-seller-silhouettes

## Purpose

Make mystery seller offers easier to scan by giving each concealed offer a safe silhouette cue. The
cue should communicate only the offer's already-visible equipment slot/category and must not reveal
item identity, rarity, stats, template, requirements, comparison data, or generated affixes.

## Non-goals

- No server/protocol changes, new asset pipeline, external art, icons, plugins, or item reveal state.
- No change to mystery offer generation, purchase, reroll, pricing, or eligibility.
- No hover preview, item comparison, requirement preview, or post-purchase identification flow.

## Acceptance Criteria

- Mystery offers draw a slot-derived silhouette in the existing shop offer cell instead of the
  generic question-mark-only glyph.
- Mystery offer detail/tooltip lines include a player-visible silhouette clue derived from slot or
  category.
- Debug state exposes the derived silhouette key for client tests while continuing to hide all
  identity fields.
- Concealed rows still report zero comparison, requirement, and equip-preview rows.
- A focused client unit test proves a ring mystery offer exposes the ring silhouette clue and still
  hides identity.
- A client bot scenario opens the mystery seller and asserts concealed mystery rows include the
  silhouette clue in their visible summary.

## Scope and Likely Files

- Client: `client/scripts/shop_panel.gd`.
- Unit tests: `client/tests/test_shop_panel.gd`.
- Bot/scenario: `tools/bot/scenarios/client/53_mystery_seller_silhouettes.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=53_mystery_seller_silhouettes.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. The implementation rejects outside assets/plugins and uses code-drawn
  shapes derived from safe metadata.
- Risk: shop panel and test files are already large. Keep the change local and avoid new shared
  bot runner code by using existing summary assertions.
