# v209 Spec: Sorcerer Teleport

Status: Complete
Date: 2026-06-16
Codename: sorcerer-teleport

## Purpose

Give Sorcerers a class mobility escape skill, `teleport`, that instantly relocates the player in an aimed direction without damaging monsters. This establishes the shared mobility skill contract that later class escape skills can reuse while keeping the first slice small and directly playable.

## Baseline

Builds on v208 companion stance command and the existing active-skill stack:

- `cast_skill_intent` already carries a `direction_or_target` payload.
- Rogue `dash` already proves server-authoritative instant movement and collision-safe endpoint resolution.
- Skill visuals are code-native and routed from shared skill metadata.

Asset/plugin decision: borrow the existing code-native skill visual and bot skill-visual replay paths; reject external assets/plugins and new art pipelines for this slice.

## Non-goals

- No Barbarian Leap, Paladin Charge, Ranger Disengage, or Rogue Dash changes in this slice.
- No client-only movement authority, protocol intent changes, or new schema version.
- No production VFX/audio; `teleport_blink` is a metadata visual id for existing placeholder rendering paths.
- No monster damage, stun, root, or invulnerability on Teleport.

## Acceptance Criteria

- `shared/rules/skills.v0.json` defines `teleport` as a Sorcerer mobility skill with data-driven range, mana cost, cooldown, and visual id.
- Shared skill schema and Go rules validation accept `kind: "mobility"` with a typed `mobility` block.
- A Sorcerer with Teleport learned can cast in a direction, spend mana, start cooldown, move to a collision-safe endpoint, and emit `skill_cast`.
- Teleport does not damage monsters it passes near.
- Skill visual/demo tooling can categorize and cast mobility skills without expecting a damage event.
- Focused validation and Go tests prove the rule and sim behavior.

## Scope and Likely Files

- Shared rules/schema: `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json`.
- Server sim: `server/internal/game/rules.go`, `server/internal/game/rogue_rules.go`, `server/internal/game/handlers.go`, new focused mobility helper/test.
- Bot/visual tooling: `tools/bot/skill_demo.py`, `tools/bot/skill_visual_runtime.py`.
- Docs: spec, plan, as-built, progress lifecycle.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestSorcererTeleport|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- `make maintainability`
- Final `make ci`

Visual manual check after the slice: `make bot-visual scenario=skill_visual ARPG_SKILL_VISUAL_SKILL_ID=teleport`.

## Open Questions and Risks

- No blocking questions.
- Risk: adding a generic mobility kind expands the skill contract in v209. This is intentional so later class escape slices remain data-driven rather than copy-pasting Rogue Dash semantics.
