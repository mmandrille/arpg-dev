# v69 Plan — Character Class Foundation

Status: Complete
Goal: Persist class identity and use it for fresh starting stats.
Architecture: Class ids become server-authoritative character metadata. Shared progression rules define valid class ids and starting stats; session-start progression asks the store for the character class before creating default progression. Existing characters default to `barbarian`.
Tech stack: Shared JSON/schema, Go store/API/session, Go tests, lifecycle docs.

## Baseline and Shortcut Decision

Builds on v68. Godot plugin decision: reject for this slice because v69 has no client UI or art work; class visuals are deferred to v71.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/character_progression.v0.json` | Class catalog and class-specific base stats |
| Modify | `shared/rules/character_progression.v0.schema.json` | Validate class catalog |
| Modify | `tools/validate_shared.py` | Cross-check valid class ids and starting stats |
| Add | `server/migrations/0017_character_class.sql` | Persist class id and backfill legacy rows |
| Modify | `server/internal/store/*` | Character model/repo class fields |
| Modify | `server/internal/http/character.go` | Class create validation and response payload |
| Modify | `server/internal/http/session.go` | Session-start class-specific defaults |
| Modify | `server/internal/store/store_test.go` | Persistence coverage |
| Modify | `server/internal/http/auth_session_test.go` | API coverage |
| Modify | `docs/as-built/v69_character-class-foundation.md`, `PROGRESS.md` | Close-out docs |

## Task 1 — Shared Rules

Files:
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/character_progression.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Add class entries for `barbarian`, `sorcerer`, `paladin` with distinct base stats.
- [x] Validate all class ids are present and every class declares `str`, `dex`, `vit`, `magic`.

```bash
make validate-shared
```

## Task 2 — Persistence and API

Files:
- Add: `server/migrations/0017_character_class.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/http/character.go`

- [x] Add `character_class` to characters with default `barbarian`.
- [x] Thread class id through create/list/get/rename responses.
- [x] Reject unknown `character_class` values in create requests.

```bash
cd server && go test ./internal/store ./internal/http
```

## Task 3 — Class Starting Stats

Files:
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/store/store_test.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Resolve the persisted class before creating character progression.
- [x] Use class-specific base stats when no progression row exists.
- [x] Prove two classes receive different durable starting stat rows.

```bash
cd server && go test ./internal/store ./internal/http
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v69_character-class-foundation.md`
- Modify: `docs/specs/v69_spec-character-class-foundation.md`
- Modify: `docs/plans/v69_2026-06-11-character-class-foundation.md`
- Modify: `PROGRESS.md`

- [x] Mark plan tasks complete as implementation lands.
- [x] Add as-built summary and lifecycle row.

```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/store ./internal/http`
- [x] `make test-go`
- [x] `make ci`
