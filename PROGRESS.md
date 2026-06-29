# Project progress

**Read this file at the start of every new task** — **Current status**, **Open gaps**, and
**Agent checklist** only. Do not read the full file when a section pointer is enough.

| Need | Read |
|------|------|
| Where we are / backlog | This file — sections below |
| Slice history & links | [`docs/progress/slice-lifecycle.md`](docs/progress/slice-lifecycle.md) |
| vN codename lookup | [`docs/progress/slice-codename-index.md`](docs/progress/slice-codename-index.md) |
| Bot scenario catalog | [`docs/progress/scenario-catalog.md`](docs/progress/scenario-catalog.md) |
| What a slice proved | [`docs/as-built/`](docs/as-built/) |
| Domain → files | [`docs/CODEMAP.md`](docs/CODEMAP.md) |

Per-slice as-built summaries live in [`docs/as-built/`](docs/as-built/). On `/finish`, update
`docs/as-built/vN_<codename>.md` and the lifecycle index — **never** add inline shipped prose here.

Last updated: 2026-06-28

---

## Current status

| Field | Value |
|-------|-------|
| **Latest completed slice** | v368 — remote-adaptive-smoothing |
| **Active branch** | `main` |
| **CI gate** | `make ci` green (2026-06-28) |
| **Next slice** | Run `/next` for new work |
| **Last engineering review** | v349 — [`docs/reviews/20260626_v349-overview.md`](docs/reviews/20260626_v349-overview.md) (2026-06-26; official cadence) |
| **Next engineering review** | Due now (~v359 milestone passed) |


### Periodic engineering reviews

Every **~10 completed slices**, pause for a repo-wide engineering review under [`docs/reviews/`](docs/reviews/).
Use the milestone slice number in filenames and headings (e.g. v50, v60, v70 — v60 is the latest pass).

**When to write:** after the milestone slice ships and `make ci-full` is green. Run `$review` first to
generate the fresh scorecard and ranked recommendations from the current baseline. Then run
`$refactor` pointed at that new review so every recommendation is classified and the minor,
verified architecture/maintainability/test/docs/process paydown commits land before `/next`
proposes the next feature batch.

**Minimum set** (follow the v53 pattern):

| File | Focus |
|------|-------|
| `docs/reviews/YYYYMMDD_vN-overview.md` | Executive summary, scorecard, cross-cutting themes |
| `docs/reviews/backend/YYYYMMDD_vN-backend.md` | Go server / `internal/game` |
| `docs/reviews/client/YYYYMMDD_vN-client.md` | Godot client |
| `docs/reviews/extras/YYYYMMDD_vN-shared-tooling-and-process.md` | `shared/`, `tools/`, SDD process |

Update **Last engineering review** / **Next engineering review** in the table above when a review lands.
Feed actionable findings into open gaps or the next slice briefs — reviews are input to `/next`, not shelfware.

---


## Architecture decisions (ADRs)

| ADR | Topic | Status |
|-----|-------|--------|
| [0001](docs/adr/0001-technology-stack.md) | Foundational stack (Go server, Godot client, shared rules, replay, bot) | Accepted |
| [0006](docs/adr/0006-asset-pipeline.md) | glTF-first assets, manifests, sockets, validation | Accepted; v3 as-built for rigged GLBs |
| [0007](docs/adr/0007-animation-state-model.md) | Client-only animation; event-driven reactions | Accepted; v4 as-built for player reactions |
| [0008](docs/adr/0008-world-structure-and-dungeon-progression.md) | World structure: infinite inverted-tower dungeon, multi-level Sim, character-scoped persistence, waypoints, co-op | Accepted |
| [0009](docs/adr/0009-boss-floors-and-timing-mechanics.md) | Boss floors, telegraphed timing mechanics, and progression gates | Proposed; v35 as-built covers first boss-floor gate |
| [0010](docs/adr/0010-mercenaries-from-player-characters.md) | Hired mercenaries derived from other players' characters | Proposed |
| [0011](docs/adr/0011-player-market-and-multi-item-trade-offers.md) | Player market listings and multi-item trade offers | Proposed |
| [0012](docs/adr/0012-item-upgrades-and-item-levels.md) | Item upgrades, item levels, and advanced dungeon resources | Proposed |
| [0013](docs/adr/0013-mystery-seller-and-unidentified-item-offers.md) | Mystery seller with expensive unidentified equipment offers | Proposed |
| [0014](docs/adr/0014-core-progression-and-endgame-design-rules.md) | Core progression, itemization, economy, endgame, co-op, and PvP design rules | Proposed |

Anticipated but **not written:** netcode timing, Protobuf migration, production auth, multiplayer split,
quest system design, NPC interaction protocol, character progression formulas
(see ADR-0001 follow-up list and ADR-0008 deferred items). Future mercenaries, player market,
item upgrades, and mystery seller economy are captured separately in ADR-0010, ADR-0011, ADR-0012,
and ADR-0013.

---


## Open gaps & deferred work

Do **not** assume these are the next slice — they are documented backlog items agents should know about.

### Active review follow-ups

- **v349 `$review` complete (official cadence at `82422216`).** Overview:
  [`docs/reviews/20260626_v349-overview.md`](docs/reviews/20260626_v349-overview.md). Failure inventory for `$refactor`:
  [`docs/reviews/20260626_v349-ci-full-failures.md`](docs/reviews/20260626_v349-ci-full-failures.md) (**15 extended scenarios** — 12 protocol + 3 client). All 15 recovered at v350; `make ci-full` now green.
- **v337 `$review` (ad hoc).** [`docs/reviews/20260625_v337-overview.md`](docs/reviews/20260625_v337-overview.md). Maintainability ratchet breach resolved; coordinator paydown items largely still open.
- **v334 `$refactor` paydown — landed:** CODEMAP inverse check, fog schema/guards, fog overlay baseline, ADR-0015, bot presentation debug extraction, dungeon generation rules split, movement-input presenter, wall-floor lab nav test + scenario 78 proof, sim tick context starter, item-visual probe extraction, fog validator unit tests.
- **v337 future-plan items:** `sim.go` phase-helper paydown via `simTickCtx`; quarantine `realtime/runner.go`; `validate_shared.py` validation-domain extraction; `main.gd` attack-move cluster extraction (9-line headroom).

### Other deferred items (from specs / ADRs)

| Area | Deferred item | Source |
|------|---------------|--------|
| Persistence | Player-facing old-session resume, delete/rename characters, class selection, visual customization, portraits, richer character detail panels, stash tabs/capacity upgrades, town stash delivery/market receipts, quest progress, durable quest turn-in completion/repeat limits/anti-farming, respec/refund, respawn/checkpoints, durable dungeon map snapshots, durable fog-of-war explored-map memory, durable buyback history, starter loadout backfill for existing or compatibility-default characters | v22/v24/v26/v39/v40/v41/v44/v45/v47/v50/v54/v59/v97/v253/v255/v291 non-goals, ADR-0008 deferred, ADR-0011, ADR-0014 |
| Combat | Basic-attack cooldown rebalance, animation-speed scaling, mana regeneration, respawn, richer spell systems, piercing/AoE/homing projectiles, debuffs/DOT/status effects, summons/traps/auras, richer ranged monster AI, richer quadruped pounce variants/polish beyond the v281 first pass, bat swarm behavior and true flying gameplay/pathing beyond the v280 dive first pass, additional ranged/rectangle boss pattern variants beyond the v282 first rectangle, elite archer packs, retreat/cover seeking, predictive leading, fog-aware monster awareness/aggro/scouting behavior, final ranged monster damage/range/cooldown balance, final combat balance across damage/HP/movement/rarity/depth, depth scaling beyond loot bands, offhand abilities/dual-wield, named elite packs/minions/aura modifiers, additional boss templates/pattern decks beyond the v287 Cave Warden/Crypt Matron set, additional enrage phases and summon variants beyond first wolf/bat adds, weighted/random boss pattern selection, final skill tree and active/passive ability catalog beyond the first passive column, additional active skills beyond Rage/Heal/Magic Bolt/Holy Shield/Arcane Barrage/Sanctuary, free-form skill formulas, class-locked skill trees, skill capability expansion beyond projectile/self-buff/area-heal/area-stat-buff/passive-stat-bonus, PvP/friendly fire | v0/v4/v12/v17/v21/v23/v26/v28/v29/v30/v31/v32/v35/v37/v39/v40/v44/v48/v52/v56/v57/v58/v59/v61/v72/v81/v159/v161/v171/v253/v254, v280-v283/v287 follow-up polish non-goals |
| Itemization | Affix grammar, procedural item names, special-effect execution, loot filters, crafting, richer gold sinks, Magic Find, additional unique/set catalogs beyond the first set package, unique items that change skill/build behavior, unique monster special drops, final item-level/depth progression, richer material wallet UI/stash material storage, shared blacksmith recipe catalog, recipe unlocks, per-recipe costs/success formulas, multi-resource recipes, item-owned levels, success-chance add/improve-roll upgrades, richer boss drop economy, richer dungeon drop economy, expanded shop depth economy bands, badge spending for stat/skill resources, quest reward source-depth metadata, item sorting/filtering, multi-cell item footprints, passive skill sources for inventory rows and equipment requirements, item auto-pickup | v23/v25/v26/v28/v29/v30/v35/v36/v39/v41/v42/v43/v47/v49/v51/v181/v202/v221/v222/v290/v292 non-goals, ADR-0009 deferred, ADR-0012, ADR-0013, ADR-0014 |
| Economy / trade | Gold/resource pricing beyond direct stash-gold listing prices, market restrictions for upgraded/bound/equipped/hotbar-assigned items, player-facing offer browser/cancel UI polish, market notification inbox/unread persistence/polling/realtime push beyond summary-refreshed board badges, clock/timer/daily mystery refresh, account-wide mystery stock, stash overflow delivery for purchases, mystery refunds/binding/special resale, final mystery price tuning against visible vendor prices, clock-based shop refresh, long-term market endgame loops for advanced players | v33/v38/v41/v42/v47/v51/v64/v68/v111/v128/v129/v130/v288 non-goals, ADR-0011, ADR-0012, ADR-0013, ADR-0014 |
| Content | Production item art/icons, production menu art/audio, production town/vendor/stash/mystery-seller/quest-giver art, production imported town building assets, collision-aware town decorations, ambient NPC movement, production dungeon art/lighting/sound, production fog/visibility art, production chest art/animation/audio, production archer attack animation, production monster art/VFX/audio, production boss art/VFX/audio, generalized ranged-monster equipment overlays, production combat/skill VFX/audio beyond code-native placeholders, production paper-doll art/model preview, colorblind/accessibility-safe rarity presentation, additional NPCs/vendors, quest-giver dialog/portrait/audio polish, mystery seller presentation polish, additional item families beyond current rules, full content-library manifest/index rollout beyond skills for items, classes, and broader presentation assets | v15/v20/v23/v24/v25/v28/v29/v30/v31/v32/v35/v36/v37/v39/v40/v41/v42/v43/v44/v45/v47/v50/v51/v52/v57/v58/v59/v60/v72/v81/v96/v97/v172/v225/v253/v255/v264/v273/v291 non-goals, ADR-0013 |
| Client presentation | Boss portraits, multi-boss layouts, exact authoritative boss countdown sync, production shape-specific telegraph decals/VFX/audio, production boss health bar art/audio, production dungeon fog lighting/art pass beyond code-native radial/LOS/organic masks, minimap routefinding/click-to-navigate/legend/filter UI, draggable titlebar migration for waypoint/menu windows, reset-layout UI, server/account-synced UI layout | v53/v57/v58/v73/v74/v75/v225/v253/v255/v263/v264 non-goals, ADR-0009 |
| Dungeon generation | Non-rectangular/polygon fog line-of-sight blocking beyond current rectangular wall/tall-obstacle and closed-door occlusion, full room/corridor PCG, rotated/polygon/destructible/secret obstacles, boss-floor obstacle generation, final biome/difficulty balance beyond first area-density formulas | v40/v252/v253/v254/v255/v260/v261/v262/v295/v296/v297/v298/v299/v300 non-goals |
| Client controls | Reliable full-scene headless modifier/mouse proof for `SHIFT+LMB` stationary attack; v37 covers the behavior with Godot unit helpers and protocol bot coverage instead | v37 deferred |
| Testing / tooling | Tuning-friendly rule tests: audit hardcoded values copied from `shared/rules/*.json` across Go/GDScript/Python/bot scenarios, classify each as contract/golden/accidental tuning pin, and convert accidental pins to rule-derived, semantic, range, or eventual assertions. Goal: balance changes such as `training_dummy.max_hp`, skill mana costs, monster cooldowns, loot weights, and generated population tuning should not require unrelated test edits; exact values remain only where a named golden or protocol/schema contract intentionally owns them. | v32 test-locking policy follow-up, v76/v77/v78 deferred |
| Settings | Controls remapping, accessibility options, language selection | v24/v224 non-goals; v351 shipped windowed/fullscreen/windowed-fullscreen display mode in `user://settings.json` |
| Assets | Blender export pipeline, texture budget, remote patcher | ADR-0006 |
| Platform | Production auth provider, dashboards, historical inspect API | v0 §8, ADR-0001 |
| Protocol | Protobuf / `godobuf` migration | ADR-0001 |
| Multiplayer | Matchmaking/lobby beyond backend-listed sessions, advanced active-session filtering/pagination/load-aware capacity controls, Steam lobby/invites, friend flows, richer party UI, chat/emotes/ready checks, richer party reward bonuses beyond full shared XP and HP/damage scaling, loot allocation, personal/hidden/reserved loot, shared/split gold, friendly fire/PvP, production remote-player art, load-aware capacity limits, split deployables / cross-process session ownership, co-op roles/encounters that change the solo experience, PvP rules that preserve skill expression while respecting builds | v0/v33/v38/v45/v46/v48/v49/v164 non-goals, ADR-0001, ADR-0014 |
| Companions / AI | Hired mercenaries derived from other players' characters, multi-offer picker UI, per-offer pricing, durable mercenary roster/recovery rules, pricing/listing model, gear snapshot refresh rules, limits per player/party, per-companion command UI, ranged mercenary AI, production mercenary variant art, mercenary loot/XP/potion behavior | v206-v208/v220/v289 non-goals, ADR-0010 |

### Curated autoloop candidates

These candidates were curated during `$autoloop 1` on 2026-06-10 and should be considered first by
the next autoloop pass unless code changes make them stale.

| Candidate | Status | Value | Size | Touch surfaces | Main risk / dependency |
|-----------|--------|-------|------|----------------|------------------------|
| `boss-phase-timer-ui` | Completed in v57 | Add boss phase/windup timing cues to the existing boss health bar. | S | client, bot, docs | Kept display-only from existing `boss_phase` state/events. |
| `boss-pattern-variety` | Completed in v58 | Add one more server-authored boss attack pattern so Cave Warden is less repetitive. | M | shared, server, bot, docs | Implemented deterministic deck-order cycling and server-owned circle hit shape. |
| `data-driven-content-library-manifest` | Completed in v60 | Introduce a manifest/index loader for skills first, preserving stable gameplay IDs and deterministic merge validation. | M | shared, server, client loader, validation, docs | Shipped as skills-only; item/class rollout remains deferred. |
| `mystery-seller-paid-reroll` | Completed in v64 | Let players spend gold to reroll concealed mystery seller stock. | M | shared/protocol, server, store, client, bot, docs | Shipped with a 50 gold server-owned reroll and deterministic stock replacement. |
| `stash-search-and-sorting` | Completed in v65 | Add search/sort controls to stash and bag views without changing item authority. | S/M | client, bot, docs | Shipped as display-only Godot controls with server-ID mutation safety. |
| `character-select-summaries` | Completed in v54 | Show level, gold, deepest depth, and status in character selection. | M | store, HTTP, client, bot, docs | Needs careful aggregate/query shape; rename/delete already exists. |
| `session-browser-filters` | Completed in v164 | Add Join Game search/filter/sort controls for listed sessions. | S/M | client, bot, docs | Shipped as display-only client controls using the existing listed-session endpoint. |
| `loot-label-filter-core` | Open | Add display-only loot label filtering/highlighting for rarity/category. | M | client, bot, docs | Presentation-only; avoid changing shared loot ownership. |
| `tuning-friendly-rule-tests` | Open | Make shared-rule balance tuning less brittle by replacing accidental hardcoded rule values in tests/scenarios with rule-derived or semantic assertions. | M | shared, server tests, client tests, bot scenarios, validation docs | Must preserve exact locks for schemas, replay determinism, persistence boundaries, and named goldens. |
| `client-boss-telegraph-polish` | Completed in v57 | Improve boss telegraph readability with a clearer in-world warning marker. | S/M | client, bot, docs | Reused in-repo primitive marker patterns; external plugins/assets rejected. |

---

## Starting a new task (agent checklist)

1. **Read this file** (`PROGRESS.md`) — confirm baseline slice and open gaps.
2. **Read ADR-0001** and any feature-specific ADRs listed above.
3. **Spec first** — create or read `docs/specs/vN_spec-<feature>.md` (SDD; `N` = next execution order).
4. **Plan second** — create `docs/plans/vN_<YYYY-MM-DD>-<feature>.md` with file map + verification commands.
5. **Branch** — stay on the current checkout; do not create branches (user creates them before development if needed).
6. **Implement** shared → server → client → bot/smoke → docs; keep `make ci` green.
7. **Update this file** when the slice completes: **Current status**, open gaps, and review cadence.
   Add the lifecycle row in [`docs/progress/slice-lifecycle.md`](docs/progress/slice-lifecycle.md).
   Write the as-built summary in `docs/as-built/vN_<codename>.md` — never inline shipped prose here.
8. **Engineering review cadence** — when the latest completed slice hits the next ~10-slice milestone
   (see **Next engineering review** above), write or refresh the review set under
   [`docs/reviews/`](docs/reviews/) first, then run `$refactor` against all review recommendations
   for scorecard-driven minor cleanup commits before piling on more feature slices.

### Invariants (do not break)

- Go sim determinism: seeded RNG only, no wall-clock in `game/`, stable ordering.
  **Enforced by CI gate:** `make lint-determinism` (step 3/9) — fails on `time.Now()`,
  `math/rand` import, or bare map range (key+value) in `sim.go` / `handlers.go`.
- New intents: register one entry in `handlers.go inputHandlers` map — do **not** edit
  `applyInput` in `sim.go`. The dispatcher is a registry lookup now.
- Shared rules are **data**; formulas evaluated in Go + GDScript from the same golden fixtures.
  After intentional formula changes: `make regen-golden` → `make ci` to keep goldens current.
- ADR-0014 progression/endgame rules are challenge rules. If a requested direction conflicts with
  stats/skills/passives mattering, loot hope, economy value, low complexity, meaningful uniques,
  endless progression, fair deaths, survival passives, all-level endgame, co-op differentiation,
  skill-based PvP, or market-as-endgame, pause and ask the owner to justify the exception before
  speccing or implementing it. Record accepted exceptions in the spec or plan.
- Animation is client-only; new reactions need a **server event** first, then client mapping.
- Golden changes require Go tests **and** GDScript `test_golden.gd` / `validate_shared.py` updates.
- GDScript shared data singletons: use `class_name Foo extends RefCounted` with `static var`
  and `ensure_loaded()` guard. Do **not** use Godot autoload for anything that headless tests
  `preload()` — autoload names are not resolvable at GDScript compile time without `--import`.

---

## Repo map (quick reference)

```text
client/          Godot 4.6.3 — main.gd, animation_controller.gd, net_client.gd, smoke.gd
server/          Go — internal/game (sim), internal/realtime (WS), internal/store (Postgres)
shared/          protocol schemas, rules JSON, golden fixtures
tools/           bot, replay, validate_shared.py, assets/
assets/          manifests + gen scripts
docs/            ADRs, specs, plans, as-built, progress archives, reviews (~every 10 slices)
```

**Agent entrypoints:** [`CLAUDE.md`](CLAUDE.md) (commands + architecture), this file (progress),
[`README.md`](README.md) (human onboarding), [`docs/reviews/`](docs/reviews/) (periodic engineering audits).
