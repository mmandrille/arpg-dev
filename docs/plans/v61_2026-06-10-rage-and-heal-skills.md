# v61 Plan — Rage and Heal Skills

Status: Complete
Goal: Add Rage and Heal as first-tier data-driven active skills with authoritative effects and client presentation.
Architecture: Replace projectile-only skill assumptions with a closed declarative `effects` list. Keep existing `cast_skill_intent` shape: projectile skills use target/direction, Rage is self-targeted, and Heal resolves its area center from target entity or cast direction. Server owns mana, cooldowns, stat buffs, healing, and expiry; the client renders selected skills, scale effects, and existing heal popups.
Tech stack: Shared JSON/schema, Go sim, Python bot, Godot client, SDD docs.

## Baseline and Shortcut Decision

Builds on v60 content manifest loading and v59 skill catalog data. The skill catalog remains loaded through `shared/content/content_libraries.v0.json`.

Godot plugin research decision: reject external plugins. Rage presentation can reuse the local player/equipment root scale path, and Heal can reuse `player_healed` floating text.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Allow closed effect-based non-projectile skill definitions. |
| Modify | `shared/rules/skills.v0.json` | Add Rage and Heal content. |
| Modify | `shared/assets/skill_presentations.v0.schema.json` | Allow non-projectile skill presentation rows. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add Rage and Heal labels/summaries. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Allow skill casts without projectiles and skill heals without potion item ids. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Keep recent-event schema aligned with state deltas. |
| Modify | `server/internal/game/rules.go` | Decode and validate supported skill kinds/effects. |
| Modify | `server/internal/game/sim.go` | Apply active stat buffs, Heal area targets, expiry, and max-HP sync. |
| Modify | `server/internal/game/handlers.go` | Dispatch skill casts by kind/effect. |
| Modify | `server/internal/game/game_test.go` | Cover catalog, Rage, Heal, and Magic Bolt regression. |
| Modify | `client/scripts/skills_panel.gd` | Render/select/spend multiple first-row skills. |
| Modify | `client/scripts/skill_bar.gd` | Show selected/right-click skill instead of first catalog id. |
| Modify | `client/scripts/main.gd` | Apply Rage visual scale events and selected skill-bar state. |
| Modify | `client/tests/*.gd` | Update skill loader, panel, bar, golden/client tests. |
| Create | `tools/bot/scenarios/39_rage_and_heal_skills.json` | Add protocol proof for Rage and Heal. |
| Modify/Create | `docs/as-built`, `PROGRESS.md` | Record shipped behavior. |

## Task 1 — Shared Catalog and Schemas

Files:
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/assets/skill_presentations.v0.schema.json`
- Modify: `shared/assets/skill_presentations.v0.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `client/tests/test_skill_rules_loader.gd`
- Modify: `client/tests/test_golden.gd`

- [x] Step 1.1: Add closed skill kinds/targeting/effects for projectile attacks, self stat-percent buffs, and area percent heals.
- [x] Step 1.2: Add `rage` and `heal` skill rows at tier 1, columns adjacent to Magic Bolt, with requested rank scaling and mana costs.
- [x] Step 1.3: Add non-projectile presentations and update loader/golden expectations.
- [x] Step 1.4: Update v8 event schemas so `player_healed` can come from skills and `skill_cast` does not require a projectile field.

```bash
make validate-shared
```

## Task 2 — Server Rule Validation and Effects

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add Go structs and validation for supported skill effects only.
- [x] Step 2.2: Split cast handling by skill kind while preserving Magic Bolt projectile behavior.
- [x] Step 2.3: Add runtime Rage buff state, stat-percent application, visual scale events, expiry, and max-HP clamp/sync.
- [x] Step 2.4: Add Heal area resolution, allied-player filtering, missing-HP clamp, `player_healed` events, and deterministic ordering.
- [x] Step 2.5: Add focused Go tests for Rage, Heal, validation, and Magic Bolt regression.

```bash
cd server && go test ./internal/game/...
```

## Task 3 — Bot Scenario Proof

Files:
- Create: `tools/bot/scenarios/39_rage_and_heal_skills.json`
- Modify: `tools/bot/run.py` and `tools/bot/test_protocol.py` only if new assertion helpers are necessary.

- [x] Step 3.1: Add a protocol scenario that allocates stats/skill points, learns Rage and Heal, casts both, and verifies emitted events/cooldowns/heal amounts.
- [x] Step 3.2: Reuse existing generic event assertions where possible; extend bot helpers only for missing observable checks.

```bash
make bot
```

## Task 4 — Client Multi-Skill UI and Presentation

Files:
- Modify: `client/scripts/skills_panel.gd`
- Modify: `client/scripts/skill_bar.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_skills_panel.gd`
- Modify: `client/tests/test_skill_bar.gd`
- Modify: `client/tests/test_coop_client.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 4.1: Render all first-row skills with selectable/spendable blocks and bot hooks by skill id.
- [x] Step 4.2: Make the skill bar track the selected/right-click skill and preserve Magic Bolt slot behavior.
- [x] Step 4.3: Apply Rage scale start/end events to the local player visual root, including equipped gear, while retaining Heal's existing `player_healed` popup path.
- [x] Step 4.4: Update unit/client-bot tests for the multi-skill catalog.

```bash
make client-unit
```

## Task 5 — Lifecycle Docs and CI

Files:
- Create: `docs/as-built/v61_rage-and-heal-skills.md`
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v61_2026-06-10-rage-and-heal-skills.md`

- [x] Step 5.1: Record as-built behavior, limitations, and verification.
- [x] Step 5.2: Update `PROGRESS.md` with v61 completion and next-review cadence unchanged.
- [x] Step 5.3: Mark plan checkboxes complete after verification.

```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `make bot`
- [x] `make client-unit`
- [x] `make ci`

## Deferred Scope

Explicit ground-position targeting, buff icons/timers in HUD, party-specific teams beyond allied player entities, and authored visual FX beyond scale/heal text are deferred.
