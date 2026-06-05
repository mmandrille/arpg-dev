# Spec: `take-a-hit`

Status: Ready for implementation (2026-06-05)
Branch: `feature/take-a-hit` (to be created)
Slice: v4 — bidirectional combat + authoritative player damage reactions
Baseline: slice v3 `animate-and-react` (complete — `make ci` green)
Related: ADR-0001 (tech stack), ADR-0006 (asset pipeline), ADR-0007 (client
animation state model — player reactions extend the event-driven path proven on
monsters in v3).

## 1. Purpose

Prove **bidirectional combat** and the **authoritative-event-driven animation
path on the local player** — the complement to v3, which deliberately left
player damage out of scope because the sim was one-directional (player → monster
only).

After this slice:

- The **training dummy retaliates** when the player lands a successful hit.
  Damage and events are **server-authoritative**; the client reacts visually.
- The server emits **`player_damaged` / `player_killed`** in `state_delta.events`
  (same envelope as existing `monster_damaged` / `monster_killed`; no new message
  types).
- The **local player** plays `hit` (one-shot) and `death` (terminal) clips via
  the existing `AnimationController`, driven from those events — mirroring the
  monster flow from v3 and fulfilling ADR-0007's stated extension point.
- The scripted vertical-slice flow (bot + smoke + golden fixture) still completes
  end-to-end, but the player **takes real damage** along the way and survives
  with reduced HP.

The proof is the **server trigger → client clip** loop on the player entity, not
combat depth, AI, or art quality.

## 2. Non-goals

- **No new attack intents or player-targeting UI.** The player still only
  `attack_intent`s monsters; retaliation is automatic on the server.
- **No range checks or aggro AI.** Retaliation is tied to a successful player
  hit, not proximity or monster tick logic.
- **No monster attack animation.** The dummy does not play an `attack` clip when
  retaliating; only the player's `hit` reaction is in scope.
- **No player respawn or session reset on death.** A dead player persists at
  `hp == 0` (same policy as killed monsters in v3).
- **No healing, armor, damage reduction, invulnerability frames, or block.**
- **No new equippable items, loot tables, or skills.**
- **No `AnimationTree` / blend spaces.** Discrete clips + the existing priority
  state machine only.
- **No protocol envelope bump.** `state_delta` shape is unchanged; only new
  `event_type` string values appear in the existing `events` array.
- **Art quality is a non-goal.** Crude `spine` wobble/topple clips, same spirit
  as v3 monster reactions.

## 3. Files to create or modify

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.schema.json` | Add optional `retaliation_damage` range on monster defs |
| Modify | `shared/rules/monsters.v0.json` | `training_dummy` retaliates `{min:1,max:1}` per hit |
| Modify | `server/internal/game/rules.go` | Load `retaliation_damage`; expose on monster def |
| Modify | `server/internal/game/sim.go` | Retaliate on successful player hit; emit player events; reject intents when player dead |
| Modify | `server/internal/game/game_test.go` | Update slice golden HP; add player-death scenario test |
| Modify | `shared/golden/slice_outcome.v0.schema.json` | Add `pinned_seed` (golden applies only with this seed) |
| Modify | `shared/golden/slice_outcome.json` | Add `pinned_seed`; `final_player_hp` is `9` for that seed (1 hit × 1 retaliation) |
| Create | `shared/golden/retaliation_damage.json` | Cross-language fixture for retaliation damage roll |
| Create | `shared/golden/retaliation_damage.v0.schema.json` | Schema for the retaliation golden |
| Modify | `shared/protocol/examples/` | Example `state_delta` payload including `player_damaged` |
| Modify | `tools/validate_shared.py` | Validate new golden if needed |
| Modify | `client/tools/build_animations.gd` | Add `hit` + `death` clips to character library |
| Modify | `client/animations/character_anims.tres` | Regenerated artifact (`make gen-anims`) |
| Modify | `client/scripts/main.gd` | `PLAYER_EVENT_CLIPS`; drive `player_anim` from events; gate input when dead; resume death pose |
| Modify | `client/scripts/smoke.gd` | Assert player `hit` path + `hp < 10` from `/state` (random session seed) |
| Modify | `client/tests/test_animation.gd` | Character `hit`/`death` clips; controller terminal; snapshot player death pose harness |
| Modify | `tools/bot/run.py` | Assert `/state` player `hp < 10` after slice (bot uses random seed, not golden seed) |
| Modify | `docs/adr/0007-animation-state-model.md` | As-built: player reactions now event-driven |
| Modify | `docs/plans/v4_2026-06-05-take-a-hit.md` | Implementation plan |

**Additional files (sensible post-plan additions, not blocking acceptance):**

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/tests/test_golden.gd` | GDScript cross-language golden: consume `retaliation_damage.json` (mirrors `tools/validate_shared.py` + Go `TestRetaliationDamageGolden`) |
| Create | `scripts/bot_visual.sh` | Start local server + launch Godot with `ARPG_AUTOPLAY=1` for visible slice inspection |
| Modify | `make/agents.mk` | Expose `make bot-visual` target |
| Modify | `README.md` | Document `make bot-visual` usage |

**Out of scope for file changes:** `AnimationController` API (unchanged),
`EquipmentVisualResolver`, asset manifests/GLBs (rig joints unchanged), monster
scenes/clips.

## 4. Data shapes

### 4.1 Monster rules extension

`monsters.v0.json` gains an optional per-monster retaliation range. When present
and both bounds are ≥ 0, a successful player hit against that monster triggers
retaliation. When absent, the monster does not retaliate (forward-compatible for
future monster defs).

```json
{
  "version": 0,
  "monsters": {
    "training_dummy": {
      "name": "Training Dummy",
      "max_hp": 3,
      "loot_table": "basic_drop",
      "retaliation_damage": { "min": 1, "max": 1 }
    }
  }
}
```

Schema (`monsters.v0.schema.json`): add optional property
`retaliation_damage: { min: integer ≥ 0, max: integer ≥ min }` with
`additionalProperties: false` preserved on the monster object.

### 4.2 New authoritative events

Emitted in `state_delta.events` (same `event` def in
`state_delta.v0.schema.json` — `event_type` remains an open string):

| `event_type` | `entity_id` | When |
|--------------|-------------|------|
| `player_damaged` | Player entity id (e.g. `"1001"`) | Retaliation damage applied and player `hp > 0` after hit |
| `player_killed` | Player entity id | Retaliation (or any future player damage) reduces player `hp` to `0` |

Mirror the existing monster event naming (`monster_damaged`, `monster_killed`).
`correlation_id` is propagated from the triggering `attack_intent` (same as
monster damage today).

**Tick ordering inside `handleAttack` on a successful player hit:**

1. Apply damage to monster; emit `entity_update` for monster if HP changed.
2. Emit `monster_damaged` (unchanged).
3. If monster `hp == 0`: emit `monster_killed` + loot (unchanged).
4. **Retaliate** if attacker monster def has `retaliation_damage` (including on
   the killing blow — simultaneous exchange):
   - Roll retaliation damage from seeded RNG (separate draws after the player
     damage draws; document draw order in the implementation plan).
   - Reduce player `hp` (floor at 0).
   - Emit `entity_update` for player.
   - If player `hp == 0`: emit **`player_killed` only**.
   - Else: emit **`player_damaged` only**.

   Monster fatal hits remain unchanged (`monster_damaged` + `monster_killed` in
   the same tick). Player retaliation uses the asymmetric pattern above (§9 #2).

### 4.3 Character animation clips (Godot-authored)

Extend `character_anims.tres` via `build_animations.gd`:

| Clip | Loop | Bone(s) | Role |
|------|------|---------|------|
| `idle` | loop | `spine` | unchanged |
| `walk` | loop | `leg_l`, `leg_r` | unchanged |
| `attack` | one-shot | `arm_r` | unchanged (client-derived) |
| `hit` | one-shot | `spine` | quick backward wobble on authoritative damage |
| `death` | one-shot, terminal | `spine` | topple and hold (player `hp == 0` pose) |

Regenerate with `make gen-anims`. No GLB or manifest change.

### 4.4 Client event → clip map

Add alongside the existing monster map in `main.gd` (client-only; not in
`shared/`):

```gdscript
const PLAYER_EVENT_CLIPS := {
	"player_damaged": "hit",
	"player_killed": "death",
}
```

`_apply_delta` event loop:

1. If `event.entity_id == player_id` and `player_anim != null`: map through
   `PLAYER_EVENT_CLIPS` (same terminal vs one-shot rules as monsters).
2. Else: existing `MONSTER_EVENT_CLIPS` path on `entities[eid].controller`.

`AnimationController` is unchanged. Event mapping stays in `main.gd` / `smoke.gd`.

### 4.5 Dead player authority

When player `hp <= 0`:

- **Server:** reject `move_intent`, `attack_intent`, `pick_up_intent`, and
  `equip_intent` with reason `"player_dead"`. `client_ready` still acks.
- **Client:** stop sending movement and attack intents; stop locomotion drive
  (`set_locomotion(false)` is insufficient once terminal — `enter_terminal`
  already ignores locomotion). Track `player_hp` from authoritative entity
  updates in snapshot/delta.

## 5. Architecture and flow

### 5.1 Server: retaliation hook

Retaliation lives in `handleAttack` after the existing monster damage block.
Pseudocode:

```
on successful player hit against monster M:
  ... existing monster damage + events ...
  if rules.Monsters[M.def].RetaliationDamage configured:
    dmg := rollRetaliationDamage()
    player.hp -= dmg
    clamp player.hp >= 0
    emit entity_update(player)
    if player.hp == 0:
      emit player_killed
    else:
      emit player_damaged
```

Missed player attacks (`attack_missed`) do **not** retaliate.

### 5.2 Client: player reaction flow (authoritative-event-driven)

1. `_ready`: `player_anim` already exists (v3).
2. `_apply_delta`: after `changes`, process `events`:
   - Player events → `player_anim.play_one_shot("hit")` or
     `player_anim.enter_terminal("death")`.
   - Monster events → unchanged entity-controller path.
3. **Interaction with client-derived `attack`:** a `player_damaged` event may
   arrive while the local `attack` one-shot is playing. Controller policy:
   `play_one_shot("hit")` replaces the active one-shot (same as re-triggering
   any one-shot). No queueing in v4.
4. **Input gating:** when tracked `player_hp <= 0`, skip `_handle_input` move/
   attack sends (and pickup/equip if the smoke/manual client would send them).

### 5.3 Resume / snapshot consistency

Extend v3's snapshot rules to the **player**:

- On `session_snapshot`, read the player entity's `hp` from `entities[]`.
- If `hp <= 0` and `player_anim != null`: `enter_terminal("death")` immediately
  (no `recent_events` dependency — same rule as monsters).
- If `hp > 0`: do not force a clip; locomotion/idle derives from input as today.
- **v5 as-built:** WebSocket resume now replays recorded inputs before sending
  `session_snapshot`; monster death and reduced player HP are restored
  authoritatively. See `v5_spec-resume-authoritative-state.md`.
- **Player death resume:** covered by the v5 WebSocket integration test with a
  focused high-HP/lethal-retaliation rules fixture.

### 5.4 Scripted slice outcome (golden)

The cross-language golden is **seed-coupled**. `TestScriptedSliceMatchesGolden`
uses pinned seed `deadbeefdeadbeef` (unchanged from v0). With
`retaliation_damage: {1,1}` and current combat (`player_damage` 2–4, dummy
`max_hp` 3), that seed kills the dummy in **one** successful hit, so the player
takes **one** retaliation:

| Field | v0 value | v4 value (`pinned_seed`) |
|-------|----------|--------------------------|
| `pinned_seed` | _(absent)_ | `"deadbeefdeadbeef"` |
| `final_player_hp` | `10` | `9` (`10 - 1` retaliation) |
| `final_monster_hp` | `0` | `0` (unchanged) |
| inventory / equipped | unchanged | unchanged |

Other seeds produce different hit counts (typically 1–2 for this monster) and
therefore different surviving HP (e.g. `8` after two retaliations). That is
expected and deterministic per seed.

- **Go golden test:** asserts `final_player_hp` against `slice_outcome.json` for
  `pinned_seed` only.
- **Bot / smoke:** use a **random server seed**; assert `hp < 10` (damage taken)
  and at least one `player_damaged` observed — **not** an exact match to
  `final_player_hp`.

### 5.5 Player death scenario (test-only)

Proving `player_killed` requires more than three hits (monster dies first). A
**Go unit test** (not the main bot flow) scripts this:

1. `NewSim` with standard rules.
2. Test mutates the spawned monster entity to `hp = 999`, `max_hp = 999` (test
   code only — no new monster def in production data).
3. Loop `attack_intent` until player `hp == 0`.
4. Assert final player HP is 0, at least one `player_damaged` in the event
   stream, and exactly one `player_killed`.
5. Assert subsequent `move_intent` is rejected with `player_dead`.

No second monster, no debug cheat intent, no production rule change.

## 6. Asset constraints

- No GLB regeneration required.
- `character_anims.tres` remains a committed generated artifact; rebuild via
  `make gen-anims`.
- Clip motion may be crude; must target `Skeleton3D` bone poses (not root
  `Node3D` transforms) so the mounted weapon still behaves correctly during
  `hit` (weapon rides `hand_r` / arm chain).

## 7. Failure behavior

- Unknown `event_type`: ignored (unchanged).
- `player_damaged` / `player_killed` with wrong `entity_id`: ignored.
- Missing `hit` / `death` clip: `AnimationController` warns (`unknown_clip`),
  stays in current state; CI fails at scene test layer.
- Dead player intents: server `reject` with `player_dead`; client logs rejection
  if received (should not send after local gate).
- Retaliation on a monster with no `retaliation_damage`: no-op (no events).
- Repeated `player_killed` or `player_damaged` after terminal death: ignored by
  controller latch (same as monster).

## 8. Acceptance criteria

1. `training_dummy` declares `retaliation_damage`; schema validates.
2. Successful player hit against the dummy emits `player_damaged` (non-fatal) or
   `player_killed` (fatal); missed attacks do not.
3. Player `entity_update` in the same tick reflects new `hp`.
4. Dead player cannot move, attack, pick up, or equip (server rejects).
5. `character_anims.tres` includes `hit` and `death`; headless scene test passes.
6. Live play: each `player_damaged` plays `hit` once; `player_killed` latches
   `death` terminal pose.
7. Resume: snapshot with player `hp == 0` shows terminal `death` without delta
   replay.
8. Scripted slice golden: `pinned_seed` + `final_player_hp == 9`; Go golden test
   passes. Bot/smoke assert `hp < 10` (random seed).
9. Player death unit test passes (high-HP monster patch).
10. Replay verification still passes for a post-slice session recording (new
    events reproduce under seed).
11. `make ci` green including extended Godot smoke.
12. v3 behaviors preserved: monster hit/death, equip visual, bone-mounted weapon,
    client-derived locomotion/attack when alive.

## 9. Decisions (resolved)

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | How does the dummy damage the player? | **Retaliate on each successful player hit.** | Smallest authoritative trigger; no AI/range; bot flow exercises it naturally; deterministic with fixed damage. |
| 2 | Fatal hit event shape | **Emit `monster_damaged` + `monster_killed` on fatal player hit (unchanged); emit `player_killed` only on fatal retaliation (no paired `player_damaged`).** | Keeps monster semantics stable; avoids duplicate player events on the killing retaliation. Non-fatal retaliation emits `player_damaged` only. |
| 3 | Retaliation on killing blow? | **Yes.** | The dummy "hits back" even as it dies; every successful player hit retaliates once. |
| 4 | Damage source | **Per-monster `retaliation_damage` in `monsters.v0.json`.** | Future monsters can opt in/out without combat.v0 churn. |
| 5 | Retaliation damage value | **Fixed `{1,1}` on `training_dummy`.** | Predictable per-hit damage; golden `final_player_hp` is derived from `pinned_seed` hit count (`9` for `deadbeefdeadbeef`). |
| 6 | Player clip authorship | **`build_animations.gd` + committed `.tres`.** | Same pipeline as v3; no GLB animation tracks. |
| 7 | Event map location | **`PLAYER_EVENT_CLIPS` in `main.gd` (not `shared/`).** | Animation presentation is client-only per ADR-0007. |
| 8 | Player death in main bot flow? | **No — separate Go test.** | Monster dies before player in the scripted flow; bot proves damage-survive path only. |
| 9 | Cross-language golden? | **Yes — `retaliation_damage.json` for the roll formula.** | Mirrors `damage_formula.json`; keeps rules evaluator honest. |
| 10 | Protocol schema version | **No bump.** | `event_type` is already an open string; examples document new values. |
| 11 | Bot/smoke HP assertion | **`hp < 10`, not exact golden.** | Bot/smoke sessions use OS-random seeds; only the Go test uses `pinned_seed`. |

## 10. Testing plan

**Python (`make validate-shared`, `pytest tools`):** schema for monsters +
retaliation golden.

**Headless GDScript golden (`client/tests/test_golden.gd`, via `make client-smoke`):**
- Consumes `damage_formula.json`, **`retaliation_damage.json`**, and `loot_roll.json`
  from `shared/golden/`.
- For retaliation: asserts golden range matches `training_dummy.retaliation_damage`
  in `monsters.v0.json`, then verifies each case satisfies
  `min + (draw mod (max - min + 1))` — same formula checked in Python and Go.

**Go (`go test ./...`):**
- `TestScriptedSliceMatchesGolden` — `pinned_seed` + `final_player_hp: 9`.
- `TestRetaliationDamageGolden` — roll formula matches fixture.
- `TestPlayerKilledByRetaliation` — high-HP monster patch, death + reject.
- Existing replay tests still green on recordings made after implementation.

**Headless GDScript (`client/tests/test_animation.gd`):**
- Character scene has `hit` + `death` clips.
- Controller: `hit` ignored after `enter_terminal("death")` on player clips.
- Snapshot harness: player entity with `hp == 0` → `enter_terminal("death")`
  (acceptance #7; mirrors monster resume wiring test pattern).

**Extended slice smoke (`smoke.gd`):**
- During kill loop: assert at least one `player_damaged` and player `hit` clip.
- After equip verify: `/state` player `hp < 10` (not exact golden — random seed).
- Resume phase after v5: assert replay-restored monster `hp == 0`, player
  `hp < 10`, and mounted weapon state from the server snapshot.

**Python bot (`make bot`):** assert `/state` and WebSocket resume both report
player `hp < 10`, monster `hp == 0`, and equipped `rusty_sword`.

**Interactive visual inspection (`make bot-visual`):** optional developer workflow;
not part of CI. See §11.

## 11. As-Built Notes

- Retaliation RNG runs after loot drop on killing blows; event order is monster
  damage/kill/loot, then player `entity_update` + `player_damaged` or
  `player_killed`.
- `make ci` green on 2026-06-05; bot/smoke assert `hp < 10` on random seeds;
  Go golden uses `pinned_seed` `deadbeefdeadbeef` with `final_player_hp: 9`.

### Post-plan additions

**`client/tests/test_golden.gd` — retaliation golden on the client**

The plan only required Go + Python validation of `retaliation_damage.json`. During
implementation, `test_golden.gd` was extended to consume the new fixture alongside
`damage_formula.json` and `loot_roll.json`, following the cross-language pattern from
slice v2 (`v2_spec-equip-and-see-it` §4.8). It runs as the first gate in
`scripts/client_smoke.sh` (`make client-smoke` / CI step 7). This gives a third
language proof that the client reads the same shared rules the server uses.

**`make bot-visual` + `scripts/bot_visual.sh` — interactive slice inspection**

Headless `make bot` and `make client-smoke` prove the slice but are hard to watch.
`bot-visual` builds the server, waits for `/readyz`, imports Godot assets, then
launches the normal client with `ARPG_AUTOPLAY=1` so the scripted flow (move →
attack → pickup → equip, including player `hit` reactions) plays visibly. The
server shuts down when the Godot window closes.

```bash
make bot-visual
AUTOPLAY_STEP_DELAY=0.8 make bot-visual   # slower steps for inspection
GODOT=/path/to/godot make bot-visual      # override Godot binary
```

Wiring: `make/agents.mk` (`bot-visual` target) → `scripts/bot_visual.sh` → Godot
`client/` with `ARPG_BASE_URL`, `ARPG_DEV_TOKEN`, `ARPG_DEBUG_TOKEN`, and optional
`AUTOPLAY_STEP_DELAY`. Documented in `README.md`; not required for slice acceptance.
