# v69 As-Built — Character Class Foundation

Date: 2026-06-11

## What Shipped

- Characters now persist a stable `character_class` id, defaulting legacy and omitted create requests to `barbarian`.
- `POST /v0/characters`, `GET /v0/characters`, create responses, and rename responses include `character_class`; invalid class ids are rejected.
- Shared character progression rules define `barbarian`, `sorcerer`, and `paladin` class entries with class-specific starting stats.
- Session-start progression reads the persisted character class before creating the durable progression row.

## Key Decisions

- `barbarian` keeps the previous baseline `5/5/5/5` start so existing omitted-class bot, smoke, and skill proofs remain stable.
- Sorcerer and paladin already prove divergent stat starts; player-facing class selection is deferred to the client slice.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/store ./internal/http`

## Deferred

- Class skill restrictions and class-required weapons move to v70.
- Create/select UI class blocks, tooltips, and sprites move to v71.
