# v48 Plan — Co-op Proximity XP And Logarithmic Monster Scaling

Status: Complete - `make ci` green
Goal: Make co-op combat reward nearby party members and scale monster HP/damage with active same-level party size.
Architecture: The Go sim remains authoritative for XP eligibility, progression, monster HP, and monster damage. Shared combat rules define the radius and logarithmic scale formula. Realtime fanout and persistence will route private progression changes by an internal owner marker, not by adding protocol fields. The protocol stays v6 unless implementation proves otherwise.
Tech stack: Shared JSON rules/goldens, Go sim/realtime/replay tests, Python protocol bot scenario, lifecycle docs.

## Baseline and shortcut decision

Baseline is v47 `shop-stock-lifecycle` on `main`. Reuse:

- v30 generated monster rarity scaling and golden style.
- v33/v38 co-op session membership, actor-tagged inputs, recipient-scoped snapshots/deltas, and N-member replay.
- v46 real Join Game proof as regression coverage, but no client UI change is expected.
## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Created | `docs/specs/v48_spec-coop-rewards-and-scaling.md` | Slice spec |
| Create | `docs/plans/v48_2026-06-09-coop-rewards-and-scaling.md` | This implementation plan |
| Modify | `PROGRESS.md` | Lifecycle update when v48 ships |
| Modify | `shared/rules/combat.v0.json` | Co-op XP and party challenge defaults |
| Modify | `shared/rules/combat.v0.schema.json` | Validate co-op combat rule block |
| Create | `shared/golden/coop_rewards_and_scaling.json` | Multiplier/eligibility expectations |
| Create | `shared/golden/coop_rewards_and_scaling.v0.schema.json` | Golden fixture schema |
| Modify | `server/internal/game/rules.go` | Parse/validate co-op combat rules |
| Modify | `server/internal/game/types.go` | Internal private change owner marker |
| Modify | `server/internal/game/sim.go` | XP sharing and party challenge scaling |
| Modify | `server/internal/game/game_test.go` | Focused sim/golden tests |
| Modify | `server/internal/realtime/session_loop.go` | Fanout/persist private changes by owner |
| Modify | `server/internal/realtime/session_loop_test.go` | Private routing tests |
| Modify | `server/internal/replay/replay_test.go` | Replay proof for shared XP |
| Modify | `tools/bot/run.py` | Co-op rewards/scaling scenario runner |
| Modify | `tools/bot/test_protocol.py` | Scenario discovery/helper tests if needed |
| Create | `tools/bot/scenarios/34_coop_rewards_and_scaling.json` | Protocol bot proof |

## Task 1 — Shared Rules And Golden

Files:
- Modify: `shared/rules/combat.v0.json`
- Modify: `shared/rules/combat.v0.schema.json`
- Create: `shared/golden/coop_rewards_and_scaling.json`
- Create: `shared/golden/coop_rewards_and_scaling.v0.schema.json`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 1.1: Add a `coop` block to combat rules with XP share radius `10.0`, full-XP sharing enabled, dead/disconnected exclusion, party challenge `per_double_bonus: 0.25`, and `max_bonus: 0.50`.
- [x] Step 1.2: Extend the combat schema with strict validation for the co-op rule block.
- [x] Step 1.3: Add a golden fixture covering 1/2/3/4/5-player multipliers and core XP eligibility defaults.
- [x] Step 1.4: Parse co-op combat rules in Go and apply defaults only through shared data.
- [x] Step 1.5: Add validation for positive radius, non-negative bonuses, `max_bonus >= per_double_bonus`, and enabled boolean fields.
- [x] Step 1.6: Add Go tests that assert multiplier values from the golden and reject invalid rule data.

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestRules|TestCoop'
```

## Task 2 — Sim XP Sharing

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add an internal, non-marshaled owner field to `game.Change` for private changes that belong to a player other than the tick actor.
- [x] Step 2.2: Replace killing-blow-only XP with stable player-id ordered eligibility over alive connected players on the killed monster's level and within the shared-rule radius.
- [x] Step 2.3: Award full XP once to every eligible player, including the killer, and preserve existing level/stat/skill point behavior.
- [x] Step 2.4: Ensure multi-recipient XP changes carry the correct private owner for progression and skill progression updates.
- [x] Step 2.5: Keep loot pickup ownership unchanged.
- [x] Step 2.6: Add Go tests for nearby share, killer no-duplicate, out-of-radius exclusion, different-level exclusion, disconnected exclusion, dead-player exclusion, recipient level-up, solo unchanged, and loot unchanged.

```bash
cd server && go test ./internal/game/... -run 'TestCoop.*Experience|TestCoop.*Loot|TestCharacterProgression'
```

## Task 3 — Monster Party Challenge Scaling

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Add deterministic party-count and multiplier helpers using alive connected players on the monster's level.
- [x] Step 3.2: Apply HP/max-HP party scaling at monster spawn time after existing rarity/template scaling.
- [x] Step 3.3: Apply damage scaling at monster attack resolution time using current same-level alive connected count.
- [x] Step 3.4: Compose party scaling with v30 rarity scaling and document/pin deterministic rounding as nearest integer with minimum `1`.
- [x] Step 3.5: Add Go tests for 1/2/3/4/5-player multiplier values, HP spawn scaling, attack damage scaling, late-join damage scaling without HP retro-scaling, and rarity composition.

```bash
cd server && go test ./internal/game/... -run 'TestCoop.*Scaling|TestMonsterRarity'
```

## Task 4 — Realtime Fanout And Persistence

Files:
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `server/internal/realtime/session_loop_test.go`

- [x] Step 4.1: Update private-change filtering so an explicit change owner takes precedence over `TickResult.ActorPlayerID`.
- [x] Step 4.2: Filter private XP/progression events by event entity/player id so one client does not receive another player's progression events.
- [x] Step 4.3: Persist private changes under the owner member's account/character when the owner marker is present.
- [x] Step 4.4: Keep existing actor-scoped inventory, shop, gold, teleporter, hotbar, and equipment behavior unchanged.
- [x] Step 4.5: Add unit tests proving multi-recipient progression changes fan out only to each owner and persist to the correct member.

```bash
cd server && go test ./internal/realtime/...
```

## Task 5 — Replay Proof

Files:
- Modify: `server/internal/replay/replay_test.go`
- Audit: `server/internal/replay/replay.go`

- [x] Step 5.1: Confirm replay reconstruction carries actor ids and member start snapshots needed for co-op shared XP.
- [x] Step 5.2: Add a replay test that records a co-op kill with shared XP and verifies derived events/final progression match recorded outcomes.
- [x] Step 5.3: Only touch replay code if private-owner result routing is lost during reconstruction.

```bash
cd server && go test ./internal/replay/...
```

## Task 6 — Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py` if new loader metadata is needed
- Create: `tools/bot/scenarios/34_coop_rewards_and_scaling.json`

- [x] Step 6.1: Add scenario metadata for `coop_rewards_and_scaling` with `world_id: "dungeon_levels"`, stable seed, and at least two peers.
- [x] Step 6.2: Extend the existing co-op bot helpers or add a focused runner for the new scenario.
- [x] Step 6.3: Drive host and guest into the same dungeon level, position both within XP share radius, kill a monster, and assert both receive XP.
- [x] Step 6.4: Prove an out-of-radius or different-level peer does not receive XP.
- [x] Step 6.5: Pin monster attack damage scaling in Go sim tests; the live protocol bot stays focused on XP/exclusion/persistence to avoid brittle dungeon passive timing.
- [x] Step 6.6: Verify reconnect or fresh state reflects durable progression for both rewarded characters.
- [x] Step 6.7: Keep `23_true_coop_session` and `27_session_browser_uncapped_coop` green.

```bash
make bot scenario=34_coop_rewards_and_scaling.json
make bot scenario=23_true_coop_session.json
make bot scenario=27_session_browser_uncapped_coop.json
```

## Task 7 — Lifecycle Docs And CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v48_2026-06-09-coop-rewards-and-scaling.md`

- [x] Step 7.1: Add v48 to the slice numbering note and lifecycle table when implementation finishes.
- [x] Step 7.2: Add a concise v48 summary under "What each slice proved".
- [x] Step 7.3: Update the scripted scenario catalog with `coop_rewards_and_scaling`.
- [x] Step 7.4: Move any newly deferred co-op reward/scaling scope to Open gaps.
- [x] Step 7.5: Keep this plan's checkboxes accurate during execution.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `cd server && go test ./internal/realtime/...`
- [x] `cd server && go test ./internal/replay/...`
- [x] `make bot scenario=34_coop_rewards_and_scaling.json`
- [x] `make bot scenario=23_true_coop_session.json`
- [x] `make bot scenario=27_session_browser_uncapped_coop.json`
- [x] `make ci`

## Deferred scope

- Loot allocation and shared gold remain unchanged.
- No party bonus beyond HP/damage scaling.
- Existing live monsters are not retroactively HP-scaled when players join, leave, die, or change levels.
- No client UI changes unless implementation discovers a missing debug assertion that cannot be tested through protocol state.
