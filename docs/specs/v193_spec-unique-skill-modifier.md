# v193 Spec - Unique Skill Modifier

Status: Approved
Date: 2026-06-15
Codename: unique-skill-modifier

## Purpose

Add the first unique effect that modifies one named skill rather than all attacks or generic skill
stats. The slice should prove unique items can carry authored skill-specific behavior while keeping
skill outcomes server-authoritative and data-driven.

## Non-goals

- No generic skill tree itemization system or broad unique-effect scripting language.
- No new client panel; existing unique chest/tooltips and event payloads should expose the effect.
- No protocol version bump unless current event/item payloads cannot represent the proof.
- No balance pass for all skills or unique effects.

## Acceptance Criteria

- A new ready unique effect declares a supported `on_skill_damage_roll` hook with a `skill_id` and
  damage bonus parameter.
- A named unique item carries the effect and appears in the debug unique chest payloads.
- When equipped, the effect increases only the configured skill's server-owned damage roll.
- Other skills and basic attacks do not receive the named-skill bonus.
- Go tests prove the configured skill bonus and non-target skill baseline.
- A protocol bot scenario proves the new named unique can be taken from the unique chest.

## Scope and Files

- Shared: `unique_effects.v0.json/schema`, `unique_items.v0.json`.
- Server: unique effect damage hook in the skill damage path, focused unique-effect tests.
- Bot: protocol unique-chest proof updated or new compact scenario.
- Docs: plan, as-built, `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'UniqueSkill|UniqueChest|UniqueItemValidation' -count=1`
- `make bot scenario=82_unique_skill_modifier.json`
- `make ci`

## Open Questions and Risks

- The first hook is intentionally narrow: additive percent damage for one configured skill before
  hit resolution. More complex modifiers should be separate slices with explicit schema fields.
