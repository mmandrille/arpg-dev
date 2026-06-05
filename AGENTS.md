# Repository Guidelines

## Project Structure & Module Organization

This repo contains a Godot client, Go server, shared contracts, and Python tooling. `client/` holds the Godot 4 project: GDScript in `client/scripts/`, scenes in `client/scenes/`, tests in `client/tests/`, and runtime assets in `client/assets/`. `server/` contains the authoritative Go server, with entrypoints in `server/cmd/` and packages in `server/internal/`. `shared/` stores versioned JSON schemas, protocol contracts, rules-as-data, and golden fixtures. `tools/` contains Python bots, replay helpers, validators, and asset generators. `assets/` is the source asset manifest area; `docs/` contains specs, ADRs, plans, and progress notes.

## Build, Test, and Development Commands

Use `make help` as the command index. Key workflows:

- `make tools`: create `.venv` and install Python tooling in editable mode.
- `make db-up`: start local Postgres from `docker-compose.yml`.
- `make server`: run the Go server against local Postgres.
- `make bot`: run the Python protocol bot against a running server.
- `make play`: start Postgres, server, and the interactive Godot client.
- `make test-go`: run `go test ./...` in `server/`.
- `make client-smoke`: run Godot headless smoke and golden checks.
- `make validate-shared` / `make validate-assets`: validate JSON contracts and assets.
- `make ci`: run the full local CI suite.

## Coding Style & Naming Conventions

Keep Go code `gofmt` formatted and package names short, lowercase, and domain-oriented. Preserve deterministic simulation rules in `server/internal/game/`: seeded RNG, stable ordering, and no wall-clock time in gameplay logic. Use snake_case for Python modules, pytest files, and GDScript files. JSON contracts are versioned with names like `*.v0.schema.json`; bump versions deliberately when wire formats change.

## Testing Guidelines

Add Go unit tests next to the package under test using `*_test.go`. Python tests use `pytest` under `tools/` with `test_*.py` names. Godot tests live in `client/tests/` and run through `make client-smoke`. Combat, loot, protocol, or equipment-visual changes should update shared golden fixtures and verify both Go and Godot consumers.

## Commit & Pull Request Guidelines

Recent history uses Conventional Commits, often scoped: `feat(client): ...`, `test(tools): ...`, `docs: ...`, `chore: ...`. Keep commits focused and describe user-visible or contract-level effects. PRs should include slice/spec context, linked issue or plan, test commands run, and screenshots or recordings for Godot-facing changes.

## Agent-Specific Instructions

Before a new feature slice, read `docs/PROGRESS.md`, then the relevant spec, plan, and ADRs. Keep the server authoritative: clients render, predict presentation, and send intents, but do not own HP, damage, loot, inventory, or replay-critical outcomes.
