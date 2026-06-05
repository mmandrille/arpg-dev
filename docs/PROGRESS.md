# Project progress & slice lifecycle

**Read this file at the start of every new task** before writing specs, plans, or code.
It is the canonical snapshot of what exists, what each slice proved, and what is still open.

Last updated: 2026-06-05

---

## Current status

| Field | Value |
|-------|-------|
| **Latest completed slice** | v6 — `visual-bot-scenario-runner` (discoverable bot scenarios + visual replay playlist) |
| **Active branch** | `feature/resume-authoritative-state` |
| **CI gate** | `make ci` green on 2026-06-05 |
| **Next slice** | TBD |

### Slice numbering note

ADR-0001 sometimes calls the first slice **v1**; repo docs call it **v0**
(`first-playable-vertical-slice`). This file uses **v0–v5** to match spec/plan filenames.

---

## Slice lifecycle

Slices are small, end-to-end proofs. Each ships: shared contracts → Go sim → Godot client →
Python bot/smoke → golden fixtures → `make ci` green.

```text
v0 first-playable ──► v2 equip-and-see-it ──► v3 animate-and-react ──► v4 take-a-hit ──► v5 resume-state ──► v6 visual-bot-scenarios
   (architecture)        (visual pipeline)         (skeletal anims)         (player damage)      (resume replay)      (visual replay playlist)
        │                      │                        │                        │                         │                         │
     main ✓                  main ✓                    main ✓                    main ✓              branch ✓                  branch ✓
```

| Slice | Codename | Status | Spec | Plan |
|-------|----------|--------|------|------|
| **v0** | `first-playable-vertical-slice` | Complete (on `main`) | [`spec-first-playable-vertical-slice.md`](specs/spec-first-playable-vertical-slice.md) | [`2026-06-05-first-playable-vertical-slice.md`](plans/2026-06-05-first-playable-vertical-slice.md) |
| **v2** | `equip-and-see-it` | Complete (on `main`) | [`spec-equip-and-see-it.md`](specs/spec-equip-and-see-it.md) | [`2026-06-05-equip-and-see-it.md`](plans/2026-06-05-equip-and-see-it.md) |
| **v3** | `animate-and-react` | Complete (on `main`) | [`spec-animate-and-react.md`](specs/spec-animate-and-react.md) | [`2026-06-05-animate-and-react.md`](plans/2026-06-05-animate-and-react.md) |
| **v4** | `take-a-hit` | Complete (on `main`) | [`spec-take-a-hit.md`](specs/spec-take-a-hit.md) | [`2026-06-05-take-a-hit.md`](plans/2026-06-05-take-a-hit.md) |
| **v5** | `resume-authoritative-state` | Complete (`make ci` green) | [`spec-resume-authoritative-state.md`](specs/spec-resume-authoritative-state.md) | [`2026-06-05-resume-authoritative-state.md`](plans/2026-06-05-resume-authoritative-state.md) |
| **v6** | `visual-bot-scenario-runner` | Complete (`make ci` green) | [`spec-visual-bot-scenario-runner.md`](specs/spec-visual-bot-scenario-runner.md) | [`2026-06-05-visual-bot-scenario-runner.md`](plans/2026-06-05-visual-bot-scenario-runner.md) |

---

## What each slice proved

### v0 — First playable vertical slice

**Proves:** ADR-0001 architecture end-to-end.

- Go authoritative server + Godot thin client over JSON WebSocket
- Dev auth, solo session create/resume, Postgres persistence
- Deterministic 20 Hz sim (move, attack, loot drop, pickup, equip)
- Seeded replay + Python protocol bot + headless Godot smoke
- `GET /v0/sessions/{id}/state` inspection for agents

**Key as-built decisions:** session-scoped inventory (not character-scoped); WebSocket
`?access_token=` fallback; monster corpse at `hp == 0`; combat always hits in v0 (no range gate).

### v2 — Equip and see it

**Proves:** ADR-0001 D7 Tier A + ADR-0006 asset pipeline contract.

- Shared `item_visuals.v0.json` + `assets.v0.json` → Godot mount on `right_hand_socket`
- Deterministic `gen_glb.py` runtime assets; `make validate-assets`
- Equipped `rusty_sword` visible on character; server authority unchanged
- Resume restores equipped weapon from persisted inventory

**Scope limit:** only `rusty_sword` has a visual mapping; other items deferred.

### v3 — Animate and react

**Proves:** ADR-0007 animation state model; rigged GLB → skeletal clips pipeline.

- Player: `idle` / `walk` / `attack` from client input/prediction; weapon on `hand_r` bone
- Monster: `hit` / `death` from authoritative `monster_damaged` / `monster_killed` events
- `AnimationController` priority machine: terminal > one-shot > locomotion
- Clips built by `client/tools/build_animations.gd` → committed `.tres` libraries
- **No server/protocol change** — client starts reading existing `state_delta.events`

### v4 — Take a hit

**Proves:** Bidirectional combat + player reactions on the same event-driven path as monsters.

- Per-monster optional `retaliation_damage` in `shared/rules/monsters.v0.json`
- Server emits `player_damaged` / `player_killed`; dead player intents rejected
- Client: player `hit` / `death` clips; input gated when `hp <= 0`
- Golden: `pinned_seed` `deadbeefdeadbeef` → `final_player_hp: 9` (one hit, one retaliation)
- Bot/smoke assert `hp < 10` on random seeds (not exact golden HP)
- Extras: `test_golden.gd` retaliation gate; `make bot-visual` for interactive inspection

**Explicit non-goals (still true):** no respawn, no healing, no monster attack anim on retaliate.

### v5 — Resume authoritative state

**Proves:** Same-session reconnect restores server-owned combat/world state through deterministic replay.

- WebSocket resume uses `replay.Reconstruct` when `session_inputs` exist.
- No `LoadInventory` on replay resume; inventory/equipped state comes from recorded pickup/equip inputs.
- Initial resume snapshot restores player HP, monster HP/death, inventory, equipped weapon, server tick, and ID continuity.
- Runner seeds historical message IDs and next sequence, so old intents reject as `duplicate`.
- Bot and Godot smoke assert real resumed monster death and reduced player HP.
- Extras: dead-player resume rejects gameplay intents; `/state` and WebSocket resume snapshot parity.

**Explicit non-goals (still true):** no character-scoped inventory, no respawn/healing/checkpoints, no protocol schema bump.

### v6 — Visual bot scenario runner

**Proves:** Bot scenarios are discoverable local artifacts and can be visually replayed without hardcoding Godot-only flows.

- `tools/bot/scenarios/*.json` defines declarative scenario steps and named assertions.
- `make bot` runs every discovered scenario through auth + WebSocket, then verifies `/state`, reconnect resume, and replay.
- `tools.bot.run --write-manifest` writes `.artifacts/bot-runs/*.json` with scenario/session metadata.
- Debug endpoint `GET /v0/sessions/{id}/replay/timeline` emits protocol-shaped snapshot/delta envelopes from deterministic replay.
- `make bot-visual` records all scenarios, verifies replay, then launches Godot with a visual replay playlist.
- Godot visual replay mode consumes the manifest and timeline envelopes through existing snapshot/delta render handlers.
- The visual replay client exits normally after the playlist completes; set `ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE=0` to keep it open.

**Explicit non-goals (still true):** no production replay browser, no durable artifact retention policy, no client presentation annotations beyond authoritative events.

---

## Architecture decisions (ADRs)

| ADR | Topic | Status |
|-----|-------|--------|
| [0001](adr/0001-technology-stack.md) | Foundational stack (Go server, Godot client, shared rules, replay, bot) | Accepted |
| [0006](adr/0006-asset-pipeline.md) | glTF-first assets, manifests, sockets, validation | Accepted; v3 as-built for rigged GLBs |
| [0007](adr/0007-animation-state-model.md) | Client-only animation; event-driven reactions | Accepted; v4 as-built for player reactions |

Anticipated but **not written:** netcode timing, Protobuf migration, production auth, multiplayer split
(see ADR-0001 follow-up list).

---

## Scripted vertical slice flow (bot + smoke)

Every slice keeps this loop working unless the spec explicitly changes it:

```text
dev-login → create session → move → attack training dummy → pick up loot → equip rusty_sword
```

After v4 the player **survives with reduced HP** (`hp < 10`). Monster dies; player may take retaliation
each successful hit. After v6 this flow lives in `tools/bot/scenarios/vertical_slice.json`; additional
scenario JSON files are automatically included in `make bot` and `make bot-visual`.

**Verify:**

```bash
make db-up && make server    # terminal 1
make bot                     # terminal 2 — all protocol bot scenarios
make client-smoke            # headless Godot gates + slice smoke
make ci                      # full suite
make bot-visual              # optional — record all bot scenarios and watch replay playlist in Godot
```

---

## Open gaps & deferred work

Do **not** assume these are the next slice — they are documented backlog items agents should know about.

### Recently closed

**Combat/world state now persists on same-session resume.** v5 replays recorded
inputs before the WebSocket `session_snapshot`, so monster death, player HP,
inventory, equipped state, and ID continuity are restored authoritatively.

### Other deferred items (from specs / ADRs)

| Area | Deferred item | Source |
|------|---------------|--------|
| Persistence | Cross-session **character-scoped** inventory | v0 as-built §10 |
| Combat | Attack range, miss chance, healing, armor, respawn | v0/v4 non-goals |
| Content | Visual mappings for items beyond `rusty_sword` | equip spec §4.9 |
| Assets | Blender export pipeline, texture budget, remote patcher | ADR-0006 |
| Platform | Production auth provider, dashboards, historical inspect API | v0 §8, ADR-0001 |
| Protocol | Protobuf / `godobuf` migration | ADR-0001 |
| Multiplayer | Matchmaking, multi-player sessions, split deployables | v0 non-goals, ADR-0001 |

---

## Starting a new task (agent checklist)

1. **Read this file** (`docs/PROGRESS.md`) — confirm baseline slice and open gaps.
2. **Read ADR-0001** and any feature-specific ADRs listed above.
3. **Spec first** — create or read `docs/specs/spec-<feature>.md` (SDD).
4. **Plan second** — create `docs/plans/<YYYY-MM-DD>-<feature>.md` with file map + verification commands.
5. **Branch** — `feature/<codename>` off latest integration branch (today: merge target TBD).
6. **Implement** shared → server → client → bot/smoke → docs; keep `make ci` green.
7. **Update this file** when the slice completes: new row in lifecycle table, summary, and any new gaps.

### Invariants (do not break)

- Go sim determinism: seeded RNG only, no wall-clock in `game/`, stable ordering.
- Shared rules are **data**; formulas evaluated in Go + GDScript from the same golden fixtures.
- Animation is client-only; new reactions need a **server event** first, then client mapping.
- Golden changes require Go tests **and** GDScript `test_golden.gd` / `validate_shared.py` updates.

---

## Repo map (quick reference)

```text
client/          Godot 4.6.3 — main.gd, animation_controller.gd, net_client.gd, smoke.gd
server/          Go — internal/game (sim), internal/realtime (WS), internal/store (Postgres)
shared/          protocol schemas, rules JSON, golden fixtures
tools/           bot, replay, validate_shared.py, assets/
assets/          manifests + gen scripts
docs/            ADRs, specs, plans, this file
```

**Agent entrypoints:** [`CLAUDE.md`](../CLAUDE.md) (commands + architecture), this file (progress),
[`README.md`](../README.md) (human onboarding).
