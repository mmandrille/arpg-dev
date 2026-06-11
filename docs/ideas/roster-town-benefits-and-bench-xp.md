# Idea: roster town benefits and bench XP

Status: idea / deferred

This note captures a possible account-roster progression system to consider later. It is not an
approved slice, spec, ADR, or implementation plan, and should not be treated as the next slice.

## Concept

When the player enters a game with one character, the account's other characters can appear in town
as inactive roster helpers. While the active character earns monster XP, eligible inactive
characters receive a small bounded fraction of that XP. Those town characters can also provide
small class-specific services that improve every 10 levels.

The goal is to make the whole account roster feel alive without letting a high-level character
power-level fresh characters for free.

## Bench XP rules to explore

- Grant only a small fraction of active-character monster XP to inactive same-account characters.
- Require inactive characters to be in a valid roster/town state, not dead or otherwise unavailable.
- Cap XP by level gap so a level 99 character killing level 100 dungeon monsters cannot rapidly
  level a fresh character.
- Consider caps based on monster depth, inactive character level, active character level, or daily /
  session totals.
- Prefer XP that helps neglected characters keep up a little, not XP that replaces playing them.
- Keep all XP awards server-authoritative, durable, replay-safe, and private to the owning account.

## Town class benefit examples

Benefits should be useful but modest. They should scale by level bracket, likely every 10 levels,
and should never replace full itemization, skills, vendors, or dungeon rewards.

| Class | Possible town benefit |
|-------|-----------------------|
| Paladin | Grants a temporary armor, block, or resistance blessing before dungeon entry. |
| Rogue | Provides limited keys or lockpicking support for closed chests. |
| Sorcerer | Offers identify, arcane reroll, mana blessing, or teleport-related convenience. |
| Barbarian | Sharpens weapons or grants a short physical damage / stun-resist preparation buff. |

## Suggested first slice shape

- Start with display-only inactive roster characters in town, using existing character data.
- Add one interactable roster helper for one class before building the full matrix.
- Keep the first benefit temporary, server-owned, and easy to prove in a protocol bot scenario.
- Defer bench XP until the town helper interaction loop is proven, or split bench XP into a separate
  slice if persistence and progression risk is too high.

## Likely future slices

1. `town-roster-helpers`: show inactive account characters in town and allow talking to one helper.
2. `town-class-benefits`: add class-specific helper services with 10-level scaling brackets.
3. `bench-xp-sharing`: add bounded inactive-character XP sharing from active monster kills.

## Acceptance criteria ideas

- Inactive same-account characters can be represented as town helpers without joining combat.
- Talking to a helper exposes a class-specific benefit derived from that character's class and
  level bracket.
- A benefit is applied by the server, reflected in protocol state, and persists or expires according
  to explicit rules.
- Bench XP, if added, is capped enough that high-level farming cannot skip early progression for
  fresh characters.
- Existing co-op XP sharing remains separate from account bench XP.

## Open questions

- Should inactive characters appear only in solo town, or also in co-op town?
- Can a character be both listed as a future mercenary and used as a town helper?
- Should helper benefits cost gold/resources, have cooldowns, or be free account perks?
- Should every class get exactly one benefit, or can classes unlock multiple town services later?
- Should bench XP be granted immediately per kill, summarized at session end, or claimed in town?
- Does bench XP apply to all inactive characters, one selected protege, or only characters near the
  active character's level?
- How should this interact with future passive trees, respecs, death, hardcore rules, and market
  value for low-level items?

## Non-goals for the first version

- No offline progress or idle-game loop.
- No replacing direct character play with account-wide leveling.
- No class balance tuning for all possible helper benefits in one slice.
- No mercenary combat AI, follower behavior, or hired-character economy.
- No client-only XP or buff calculation.
