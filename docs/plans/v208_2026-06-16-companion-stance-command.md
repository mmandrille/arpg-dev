# v208 Plan - Companion Stance Command

Status: Complete
Goal: Add a server-authoritative companion stance command with `assist`, `defend`, and `passive` behavior, visible in protocol state and proven by tests/bot.
Architecture: Keep stance semantics in the simulation. Clients send one command intent for all active owned companions; entity views and events expose the result.
Tech stack: Go simulation/input decoding, shared JSON schemas, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v206/v207 companion and mercenary work. Asset/plugin decision: reject external assets/plugins and new client UI assets; this slice is protocol/simulation first and defers clickable Godot controls.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/internal/game/companion_stance.go` | Stance constants, validation, command handler, and helper logic. |
| Add | `server/internal/game/companion_stance_test.go` | Focused AI and handler behavior coverage. |
| Modify | `server/internal/game/companion_ai.go` | Apply stance filtering to target selection and expose stance in entity views. |
| Modify | `server/internal/game/sim.go` | Add entity/input fields for stance state and intent payload. |
| Modify | `server/internal/game/types.go` | Add protocol fields for entity stance and stance event payload. |
| Modify | `server/internal/game/handlers.go` | Register `companion_command_intent`. |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode and whitelist the new intent. |
| Modify | `server/internal/inputdecode/inputdecode_test.go` | Cover valid/invalid stance payloads. |
| Modify | `shared/protocol/messages.v0.schema.json` | Add command intent schema. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Add `companion_stance` entity field and event field. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Add `companion_stance` entity field and stance event requirements. |
| Modify | `tools/bot/run.py` | Add `set_companion_stance` action and stance entity selector. |
| Add | `tools/bot/scenarios/89_companion_stance_command.json` | Protocol bot proof. |
| Modify | lifecycle docs | Spec/plan/as-built/progress/scenario catalog updates. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/types.go`
- [x] `server/internal/game/handlers.go`
- [x] `server/internal/inputdecode/inputdecode.go`
- [x] `tools/bot/run.py`

Decision:
- [x] Keep simulation command logic in new `companion_stance.go`.
- [x] Touch large files only for data shape, registry, and schema-adjacent decode wiring.
- [x] Keep bot additions to one action branch and one generic selector key.

Verification:
```bash
make maintainability
```

## Task 1 - Simulation Contract

Files:
- Add: `server/internal/game/companion_stance.go`
- Add: `server/internal/game/companion_stance_test.go`
- Modify: `server/internal/game/companion_ai.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/handlers.go`

- [x] Step 1.1: Add stance constants/defaulting and an input payload.
- [x] Step 1.2: Implement `companion_command_intent` validation, companion updates, entity deltas, and `companion_stance_changed`.
- [x] Step 1.3: Make AI honor `assist`, `defend`, and `passive` target rules.
- [x] Step 1.4: Expose `companion_stance` on companion entity views.
```bash
cd server && go test ./internal/game -run 'TestCompanionStance'
```

## Task 2 - Decode and Shared Protocol

Files:
- Modify: `server/internal/inputdecode/inputdecode.go`
- Modify: `server/internal/inputdecode/inputdecode_test.go`
- Modify: `shared/protocol/messages.v0.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`

- [x] Step 2.1: Decode `companion_command_intent` from the wire.
- [x] Step 2.2: Add schema coverage for the command payload, entity field, and stance event.
- [x] Step 2.3: Validate shared schemas.
```bash
make validate-shared
cd server && go test ./internal/inputdecode -run 'TestDecodeCompanionCommandIntent'
```

## Task 3 - Bot Proof

Files:
- Modify: `tools/bot/run.py`
- Add: `tools/bot/scenarios/89_companion_stance_command.json`
- Modify: `docs/progress/scenario-catalog.md`

- [x] Step 3.1: Add `set_companion_stance` to the protocol bot.
- [x] Step 3.2: Let entity-count selectors match `companion_stance`.
- [x] Step 3.3: Add a scenario that hires a mercenary, switches stances, and proves stance state plus combat behavior.
```bash
make bot scenario=companion_stance_command
make bot scenario=mercenary_hiring_board
```

## Task 4 - Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v208_spec-companion-stance-command.md`
- Modify: `docs/plans/v208_2026-06-16-companion-stance-command.md`
- Add: `docs/as-built/v208_companion-stance-command.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 4.1: Mark spec/plan complete and write as-built proof.
- [x] Step 4.2: Update progress/lifecycle docs after verification.
```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/inputdecode ./internal/game -run 'TestDecodeCompanionCommandIntent|TestCompanionStance'`
- [x] `make bot scenario=companion_stance_command`
- [x] `make bot scenario=mercenary_hiring_board`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

Godot stance controls, per-companion commands, hold-position/retreat behavior, durable stance persistence, mercenary death/loss rules, mercenary gear snapshot refresh, loot/XP/potion behavior, and listing/pricing models remain deferred.
