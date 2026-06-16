# v203 Plan - Upgrade Pity Counter

Status: Complete
Goal: Add item-owned upgrade pity progress that guarantees success after a configured failure streak.
Architecture: Keep pity as rolled item metadata (`upgrade_pity.failures`) so it follows inventory,
stash, and market item movement without a schema migration. The server/store remains authoritative;
the client only previews current progress from returned item metadata.
Tech stack: Shared JSON rules/schema, Go store + HTTP route threading, Godot blacksmith UI, client
bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v197 upgrade success chance and v202 resource consumption. Asset/plugin decision:
reject external assets/plugins; reuse the existing blacksmith panel and bot scenario.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Configure pity failure threshold. |
| Modify | `shared/rules/main_config.v0.schema.json` | Validate pity threshold. |
| Modify | `server/internal/game/rules.go` | Load and validate pity threshold. |
| Modify | `server/internal/store/interfaces.go` | Add threshold parameter to upgrade store contract. |
| Modify | `server/internal/store/repos.go` | Persist pity failures and force success at threshold. |
| Add | `server/internal/store/upgrade_pity.go` | Parse and write item-owned pity metadata. |
| Modify | `server/internal/store/store_test.go` or `upgrade_chance_test.go` | Cover fail/fail/guaranteed-success behavior. |
| Modify | `server/internal/http/account_stash.go` | Pass configured threshold to the store. |
| Modify | `server/internal/http/auth_session_test.go` | Keep HTTP upgrade route deterministic with new response metadata. |
| Modify | `server/internal/replay/replay_test.go` | Update fake store interface. |
| Modify | `client/scripts/blacksmith_panel.gd` | Preview/debug pity count, threshold, and guarantee state. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match blacksmith pity debug expectations. |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate pity expectations in blacksmith bot steps. |
| Add | `client/tests/test_blacksmith_panel.gd` | Focused blacksmith pity UI unit coverage. |
| Modify | `scripts/client_smoke.sh` | Run the focused blacksmith panel test. |
| Modify | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Assert default pity fields. |
| Add | `docs/as-built/v203_upgrade-pity-counter.md` | Record proof. |
| Modify | `PROGRESS.md` | Update current status after completion. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v203 lifecycle row. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` is not planned for this slice.
- [x] `client/scripts/bot_scenario_runner.gd` may grow but remains within grandfather allowance.
- [x] `client/scripts/blacksmith_panel.gd` must stay at or below 600 lines; trim equivalent lines or extract helper code.
- [x] Did every touched grandfathered file stay at or below its baseline/allowance?

Verification:
```bash
make maintainability
```

## Task 1 - Shared config

Files:
- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `item_upgrade_pity_failure_threshold`.
- [x] Step 1.2: Validate the threshold as non-negative.

```bash
make validate-shared
```

## Task 2 - Store authority

Files:
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Add: `server/internal/store/upgrade_pity.go`
- Modify: `server/internal/store/upgrade_chance_test.go`
- Modify: `server/internal/replay/replay_test.go`

- [x] Step 2.1: Read/write `upgrade_pity.failures` in rolled item metadata.
- [x] Step 2.2: Increment failures on accepted failed attempts.
- [x] Step 2.3: Guarantee success when the pre-attempt failure count meets the threshold.
- [x] Step 2.4: Reset failures on success.
- [x] Step 2.5: Cover fail/fail/guaranteed-success in a deterministic store test.

```bash
cd server && go test ./internal/store -run 'Upgrade' -count=1
```

## Task 3 - HTTP/client presentation

Files:
- Modify: `server/internal/http/account_stash.go`
- Modify: `server/internal/http/auth_session_test.go`
- Modify: `client/scripts/blacksmith_panel.gd`
- Add: `client/tests/test_blacksmith_panel.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 3.1: Pass the configured pity threshold through HTTP upgrade calls.
- [x] Step 3.2: Expose blacksmith debug fields for current failures, threshold, and guaranteed state.
- [x] Step 3.3: Show pity progress in the staged-item preview.
- [x] Step 3.4: Add focused client unit coverage.

```bash
cd server && go test ./internal/http -run 'Upgrade' -count=1
make client-unit
```

## Task 4 - Bot proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Step 4.1: Extend blacksmith bot matching and validation for pity fields.
- [x] Step 4.2: Assert default pity threshold/count/guarantee state in the blacksmith scenario.

```bash
SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 5 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v203_upgrade-pity-counter.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Step 5.1: Mark the spec complete after verification.
- [x] Step 5.2: Add as-built proof notes.
- [x] Step 5.3: Update progress status and lifecycle row.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/store -run 'Upgrade' -count=1`
- [x] `cd server && go test ./internal/http -run 'Upgrade' -count=1`
- [x] `make client-unit`
- [x] `SCENARIO=blacksmith_upgrade_ui HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make ci`
