# v69 Spec — Character Class Foundation

Status: Complete
Date: 2026-06-11
Codename: character-class-foundation

## Purpose

Introduce authoritative character class identity for new and existing characters. A character can be a `barbarian`, `sorcerer`, or `paladin`; the selected class persists with the character, appears in character API responses, and seeds fresh character progression with class-specific starting stats.

## Non-goals

- No create-character class picker UI yet.
- No class sprites in the character list yet.
- No skill or item class restrictions yet.
- No migration UI for changing an existing character class after creation.

## Acceptance Criteria

- `POST /v0/characters` accepts optional `character_class`; omitted values create a valid default class.
- Invalid classes are rejected with a 400 response and do not create characters.
- `GET /v0/characters` and create/rename responses include `character_class`.
- Existing characters migrate to a deterministic default class.
- Fresh session-start progression uses the persisted character class starting stats.
- Store and HTTP tests prove at least two classes start with different base stats.

## Scope and Likely Files

- Shared rules: `shared/rules/character_progression.v0.json`, schema, validation.
- Server store/API: character model, migration, create/list/get SQL, HTTP request/response types.
- Session start: progression defaults selected from persisted class identity.
- Tests: store and HTTP tests for class persistence, validation, and starting stats.
- Docs: spec, plan, as-built, `PROGRESS.md` during finish.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/store ./internal/http`
- `make test-go`
- `make ci`

Bot proof is deferred because v69 does not expose player-facing class selection in the client; v71 will add client bot coverage for create/select UI.

## Open Questions and Risks

- Default class for legacy/omitted creation is `barbarian` to keep existing melee starts viable.
- Class names are stored as stable lowercase ids; display names remain a client concern until v71.
