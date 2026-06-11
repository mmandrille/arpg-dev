# v70 Spec — Class Skill and Item Gates

Status: Complete
Date: 2026-06-11
Codename: class-skill-and-item-gates

## Purpose

Make the persisted character class affect gameplay. Each class can learn and cast only its class skill, and each class receives one slightly stronger class-required weapon definition.

## Non-goals

- No create-character class picker UI; v71 owns the client selection flow.
- No class sprites or character-list presentation.
- No procedural class-locked rolled items.
- No passive skills or multiple skill trees per class.

## Acceptance Criteria

- Skill definitions use class ids: `barbarian`, `sorcerer`, and `paladin`.
- Barbarian can learn/cast `rage`, sorcerer can learn/cast `magic_bolt`, and paladin can learn/cast `heal`.
- Cross-class learn/cast attempts are rejected authoritatively and do not mutate ranks/resources.
- Fixed weapon definitions exist for each class and are stronger than baseline `rusty_sword`.
- Class-required weapons reject equip attempts by other classes and equip successfully for their class.
- Session-start/replay snapshots carry class identity so restrictions survive the authoritative boundary.

## Scope and Likely Files

- Shared rules/schemas: skills, item definitions, validation.
- Server sim: progression class state, skill and equipment checks.
- Store snapshots: persist class through session-start progression.
- Tests: Go unit coverage for class skill/item gates.
- Docs: spec, plan, as-built, `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game ./internal/store ./internal/http ./internal/realtime ./internal/replay`
- `make test-go`
- `make ci`

Protocol bot proof creates class-specific characters for Magic Bolt, Rage, and Heal paths. The Godot
class picker UI remains v71 scope.

## Open Questions and Risks

- `heal` becomes paladin-owned; existing heal-lab tests that use default barbarian must explicitly use paladin progression where needed.
