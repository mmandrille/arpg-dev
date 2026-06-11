# ADR-0014: Core Progression and Endgame Design Rules

- **Status:** Proposed
- **Date:** 2026-06-11
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, progression, itemization, economy, endgame, co-op, pvp, trade

---

## Context

The project now has durable character progression, stats, skills, rolled equipment, stash storage,
gold, vendors, co-op foundations, boss floors, and future ADRs for item upgrades, mystery seller
offers, and the player market.

Future slices need product-level rules that keep tuning and feature work aligned. Without explicit
direction, it is easy for a looter ARPG to collapse into a single-axis game where gear score is all
that matters, or into the opposite problem: too many currencies, upgrade paths, and opaque systems
that make players feel lost.

This ADR records the core progression and endgame design rules that future specs should reference
before adding itemization, skills, economy, co-op, PvP, or market features. These are also agent
challenge rules: if the project owner requests a direction that appears to violate one of these
rules, the agent should pause, name the conflict, and ask the owner to argue why the exception is
worth it before implementing or writing the spec that way.

---

## Decisions

### D1 - Character build power is multi-source

**Decision:** Stats, active skills, passive skills, and gear must all matter. No single exceptional
item should make any character with any stat/skill choices equivalent to a deliberately built
character.

**Design rule:** Gear may create breakthroughs, but character investment must remain visible in
damage, survival, resource sustain, utility, and build access. Item requirements, stat scaling,
passives, skill synergies, and resource costs are valid tools to prevent "loot alone solves the
character" tuning.

**Rejected:** pure gear-score progression where leveling and build choices become placeholders.

### D2 - Every level should keep loot hope alive

**Decision:** On any dungeon level, the player should be able to believe that the next monster,
chest, boss, vendor refresh, or mystery seller outcome might produce a useful item for the current
build or for another character build.

**Design rule:** Loot tables should support current-build upgrades, off-build discoveries, account
stash value, and trade value. A level may be inefficient for a specific target, but it should not
feel categorically pointless once the player outgrows a narrow band.

### D3 - Gold and resources must stay valuable

**Decision:** Gold and future resources should matter because they pay for meaningful actions:
upgrades, vendor purchases, mystery offers, rerolls, stash services, trade participation, and other
town or endgame systems.

**Design rule:** Sinks should be durable enough that advanced players still care about drops and
trade, but costs must be legible. A resource should exist only when it creates a clear decision the
player can understand.

### D4 - Avoid unnecessary resource complexity

**Decision:** Do not add multi-resource systems unless each resource has a distinct, player-visible
purpose.

**Design rule:** Prefer a small set of broadly useful currencies/materials over many narrow tokens.
If a feature proposes a new resource, the spec must explain why gold, an existing material, item
ownership, or dungeon depth cannot carry the decision cleanly.

### D5 - Unique items change behavior, not just numbers

**Decision:** Unique items should customize the experience by changing how skills, passives, or
build loops work in meaningful ways.

**Design rule:** A unique item can have strong stats, but its reason to exist should be a build
identity hook: altering a skill shape, enabling a new resource loop, changing survival posture,
supporting a co-op role, or opening a tradeoff that normal affixes do not offer.

### D6 - Progression should feel effectively endless

**Decision:** The game should let players believe there is always another target: deeper dungeon
levels, better rolls, item upgrades, alternate builds, market goals, boss challenges, or co-op/PvP
mastery.

**Design rule:** Endless progression does not require infinite power inflation. Horizontal goals,
rare build-enabling items, account stash projects, trade goals, and harder content can carry the
feeling without making old systems meaningless.

### D7 - Death is important but should not be instant spike failure

**Decision:** A fair death should usually be the result of multiple errors, ignored warnings,
resource mismanagement, bad positioning, bad build preparation, or failed recovery choices. A
character should not die from a single unreactable spike of damage.

**Design rule:** Bosses and dangerous monsters may punish hard, but damage needs readable buildup,
telegraphs, cadence, mitigation paths, or recovery windows. One-shot mechanics are only acceptable
as explicitly signaled challenge rules, not as normal combat tuning.

### D8 - Passive skills matter for survival

**Decision:** Passive skills must be a major part of survivability, not only a way to increase
damage.

**Design rule:** Future passive trees should include meaningful health, mitigation, resistance,
recovery, mobility, resource sustain, and risk-conversion choices. Survival passives should enable
distinct play styles rather than only adding flat durability.

### D9 - Endgame content exists at all levels

**Decision:** "Endgame" should not mean only maximum-level content. Players should encounter
endgame-like goals throughout progression: hard optional fights, trade-worthy drops, upgrade
materials, build-defining items, market opportunities, and account-level goals at many depths.

**Design rule:** Low and mid levels can still produce valuable items when those items have build,
trade, upgrade, twink, or account-stash value. Future content should avoid making early progression
feel like disposable tutorial space.

### D10 - Co-op changes the experience

**Decision:** Co-op should be fun and should change how the game plays compared with solo play.

**Design rule:** Co-op should create reasons to coordinate: complementary skills, support roles,
shared danger, positioning, revive/recovery opportunities, party rewards, and encounters that are
more interesting with multiple players. Co-op should not become mandatory for basic progression,
but it should feel meaningfully different and desirable.

### D11 - PvP is skill-based, with builds still mattering

**Decision:** PvP should reward player execution, timing, positioning, and matchup knowledge, while
still respecting character builds.

**Design rule:** Gear and build choices can create advantages and identities, but PvP should avoid
unanswerable stat-check outcomes. PvP-specific scaling, caps, telegraphs, cooldowns, arenas, or
rulesets may be required to keep skill expression visible.

### D12 - The player market is the real long-term endgame

**Decision:** ADR-0011's player market is the real endgame direction. Even the most advanced and
powerful players should still need, want, produce, or trade valuable items.

**Design rule:** Itemization, upgrades, uniques, resources, and account stash goals should feed
trade value instead of ending it. Advanced players should have reasons to seek niche items,
perfected rolls, off-build pieces, upgrade materials, market arbitrage, and multi-item offers.

---

## Spec Requirements For Future Slices

Any future spec touching progression, combat balance, itemization, economy, co-op, PvP, upgrades,
unique items, or market systems should answer:

- Which of these design rules does the slice advance?
- Does it preserve stats/skills/passives as meaningful alongside gear?
- Does it create or preserve reasons to keep hunting, spending, or trading?
- Does it add a new resource, and if so, why is that resource necessary?
- Could the feature create unfair spike deaths, one-axis stat checks, or mandatory co-op?
- How will the feature remain valuable before and after the current character's immediate upgrade
  path is exhausted?

### Agent challenge rule

When a user request, proposed spec, plan, or implementation decision conflicts with this ADR, the
agent must challenge the decision before proceeding. The challenge should be concrete and brief:

- Identify the exact ADR-0014 rule at risk.
- Explain the likely player-facing cost.
- Ask the owner to confirm the exception and explain the reason.
- If the owner confirms, document the exception in the spec or plan with the rationale.

The agent should still follow the owner's final decision after the conflict is explicit and the
rationale is recorded. Silent drift away from these rules is not acceptable.

---

## Relationship To Existing ADRs

- [ADR-0008](0008-world-structure-and-dungeon-progression.md) defines the infinite dungeon and
  co-op world model; this ADR defines how progression should feel inside that world.
- [ADR-0009](0009-boss-floors-and-timing-mechanics.md) already favors telegraphed timing
  mechanics; this ADR strengthens the rule that deaths should be readable and recoverable until
  the player compounds mistakes.
- [ADR-0011](0011-player-market-and-multi-item-trade-offers.md) is elevated here as the long-term
  endgame economy direction.
- [ADR-0012](0012-item-upgrades-and-item-levels.md) should use these rules to keep upgrades
  meaningful without turning into opaque multi-currency complexity.
- [ADR-0013](0013-mystery-seller-and-unidentified-item-offers.md) supports loot hope and gold
  sinks, but must stay tuned so blind offers are exciting rather than mandatory or punishing.

---

## Open Design Questions

- What exact power budget should come from character stats, active skills, passive skills, and gear
  at early, mid, and advanced progression?
- What resource set is sufficient for upgrades and trade without overwhelming players?
- Which unique-item effects are allowed to alter skill behavior, cooldowns, targeting, resource
  flow, or party roles?
- What PvP-specific scaling or constraints are required once friendly fire/PvP enters scope?
- How should low-level and mid-level items retain endgame trade value without flooding the market
  with noise?
- Which death or recovery penalties make death important without making progress feel brittle?

---

## Non-Goals For Current Slices

This ADR does not implement formulas, passive trees, PvP rules, unique item catalogs, market
routes, upgrade resources, economy sinks, or encounter balance. It records product direction so
future specs can make coherent implementation choices.

---

## Consequences

- Future balance work should test more than raw DPS or gear score; specs should include survival,
  resource, and build-fit proof where relevant.
- Agents must challenge owner requests that appear to violate these rules, and specs/plans must
  record accepted exceptions instead of burying them in implementation details.
- Unique items and passives need effect systems capable of behavior changes, not only stat
  modifiers.
- Economy specs must justify new resources and should prefer legible sinks tied to durable player
  goals.
- Market and trade design should be considered a core loop, not a late optional feature layered on
  after itemization is complete.
- PvP and co-op will likely need dedicated rules rather than blindly reusing solo combat tuning.
