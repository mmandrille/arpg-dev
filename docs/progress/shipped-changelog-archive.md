# Shipped changelog archive

Historical "Recently closed" notes extracted from `PROGRESS.md` during v200 progress-doc
compaction. Canonical per-slice proof lives in [`docs/as-built/`](../as-built/) — prefer
as-built for slice details.

### Recently closed

**A mercenary companion archetype is in place.** v198 adds `mercenary_guard` as a shared-rules
companion archetype, gives it visual/text catalog entries, and proves owned follow/assist combat in
`mercenary_foundation_lab` with protocol bot scenario `86_mercenary_foundation`.

**Blacksmith upgrades now have data-driven success chance.** v197 adds
`item_upgrade_success_chance_percent`, returns per-attempt success, spends gold on failed attempts
without mutating item stats, and exposes the chance in the blacksmith panel/debug state.

**Rolled items now carry authoritative item level.** v196 adds `item_level` to the durable
rolled-item payload, surfaces it across loot/inventory/stash/shop protocol views, maps generated
template rolls from source depth, and adds protocol bot scenario `85_item_level_progression`.

**Boss and elite reward sources now have authored special drops.** v195 extends treasure-class
entries with `unique_item_id` and `set_item_id`, lets boss and elite objective rewards spawn fixed
unique/set payloads, and adds protocol bot scenario `84_boss_special_drops` proving Cave Warden
drops `Conduit Staff` and `Stormrunner Covenant Bow`.

**A second five-piece set package is live.** v194 adds `Stormrunner Covenant`, a bow/head/gloves/
boots/ring set with fixed stats and 2/3/4/full-set bonuses for dexterity, crit chance, attack
speed, all skills, skill damage, and Magic Find. Protocol bot scenario `83_second_set_package`
proves the new set package is available through the unique chest.

**Unique items can now modify a named skill.** v193 adds the `Arcane Conduit` effect and
`Conduit Staff` named unique, threads skill identity into server-owned skill damage, and proves
Magic Bolt receives the unique damage bonus while other skills remain baseline. Protocol bot
scenario `82_unique_skill_modifier` proves the named unique is available through the unique chest.

**Magic Find is now a visible loot stat.** v192 adds `magic_find_percent` as a rollable equipment
stat, exposes equipped Magic Find in derived stats and stat breakdowns, applies it only to monster
item-template rarity rolls, and adds protocol bot scenario `81_magic_find_stat` for deterministic
pickup/equip/stat proof.

**Rolled items now get readable affix names.** v191 adds deterministic affix-style
display names for magic-or-higher non-unique, non-set rolled items, such as
`Focused Rare Sorcerer Staff`, while keeping common, unique, set, and protocol payload ownership
unchanged. Protocol bot scenario `80_skill_affix_rolls` now asserts the generated display name.

**Skill utility affixes now roll and affect casts.** v189 adds rare-or-higher
`skill_cooldown_reduction_percent` and `skill_mana_cost_reduction` roll candidates, validates and
prices them, applies equipped reductions to authoritative skill cooldown and mana spend, updates
client stat labels, and adds protocol bot scenario `80_skill_affix_rolls`.

**Rare combat affixes now affect authoritative combat.** v188 adds rare `hit_chance`,
`crit_chance`, and `evade_chance` roll candidates, prices and validates them as percent-point
stats, aggregates equipped rolls into player derived stats, resolves defender evade before block
with the existing miss outcome, and adds bot scenario `79_live_rare_combat_affixes`.

**Item rarity now controls roll counts and inherited roll pools.** v187 replaces fixed item
`stat_rolls` with data-driven rarity roll-count ranges, adds `min_rarity` roll candidates,
keeps fixed set rarity out of random rolls, updates item/shop goldens, and adds protocol bot
scenario `78_rarity_roll_pools` proving high-rarity roll count and inherited stat-pool payloads.

**Elite minions now use leader-driven pack AI.** v186 routes non-leader elite pack members
through deterministic leader follow/assist behavior, suppresses standalone passive aggro while
their leader is idle, preserves elite objective/aura metadata, and adds protocol bot scenario
`77_elite_minion_pack_ai` proving leader-driven pack combat.

**Companion rank limits and Revive scaling are data-driven.** v185 replaces
hardcoded one-active companion rules with shared limit data, keeps Ranger wolf at one active
through tuning, gives Revive +1 active monster every 3 ranks, and adds protocol bot scenario
`76_companion_rank_scaling_and_limits` proving rank-4 two-companion Revive with scaled HP

**Paladin defensive skills now update stats and prevent damage correctly.** v190 exposes effective
block in derived stats, refreshes derived stat payloads when skill buffs start/end, keeps Holy
Shield's armor/block benefit visible, and changes Sanctuary into a 5-unit, 6-second yellow immunity
dome with a 60-second cooldown.

**Engineering review v190 is complete.** The review found the batch healthy and `make ci` green,
with follow-up pressure on companion AI tuning configurability, Sanctuary client radius drift,
monster-rarity validation pins, and continued large-file extraction discipline.

**Sorcerer can revive slain monsters as companions.** v184 adds the `revive`
Sorcerer skill with a data-driven `revive_companion` payload, dead non-boss targeting,
boss/living target rejection, original monster visual identity, and protocol bot scenario
`75_sorcerer_revive_companion` proving kill, revive, companion spawn, and companion damage.

**Ranger can now summon a black wolf companion.** v183 adds the `black_wolf_companion`
Ranger skill with a data-driven `summon_companion` payload, one-active-wolf replacement,
black quadruped presentation, and protocol bot scenario `74_ranger_wolf_companion`
proving cast, spawn, follow, and companion-sourced damage.

**Server-owned companion actors now have a foundation.** v182 adds a distinct `companion`
entity type to v8 snapshots/deltas, a compact `companion_ai_lab`, server-owned follow/target/melee
AI, and protocol bot scenario `73_companion_ai_foundation` proving a test companion follows its
owner and damages a lab monster.

**The first five-piece set item package is live.** v181 adds the green `Verdant Vanguard` set
catalog, exposes its five fixed pieces in the debug unique chest, applies server-authoritative
2/3/4/full-set bonuses from equipped pieces, and colors set items green in existing loot/tooltips.
The requested skill-cast visual troubleshooting command already exists as
`make skill-visual skill=<skill_id> rank=<rank>` and `make skill-visual-list`.

**Upgrade resources now drop as tangible loot.** v180 adds `upgrade_shard`, presentation metadata,
and a deterministic ranged-lab drop source, with protocol bot scenario
`72_upgrade_resource_drop.json` proving pickup into inventory.

**Passive mana regeneration is now an explicit gameplay event.** v179 emits
`player_mana_regenerated` when passive stat-driven regen restores mana, validates the event in v8
schemas, and proves the loop with protocol bot scenario `71_mana_regeneration.json`.

**Cave Warden now summons combat adds.** v178 adds the data-driven `summon_wolves` pattern,
spawns normal server-owned `dungeon_wolf` adds exactly once during the active phase, emits
`boss_summoned_adds`, and keeps the boss-floor protocol and client readability proofs green.

**Cave Warden now has a server-authored ranged line pattern.** v177 adds the data-driven
`stone_lance` pattern, validates matching line telegraph/active width metadata, locks aim at
telegraph start, and proves authoritative line hit/miss behavior plus the boss-floor protocol and
client readability scenarios.

**Elite objective chests now have a compact minimap pin.** v176 adds a top-right minimap-style
widget with a player dot and objective pin for closed `elite_objective` chests, derives its
position from existing client entity metadata, and proves the active state with client bot scenario
`45_elite_objective_minimap_pin`.

**Elite objective state is now visible in the HUD.** v175 adds a compact tracker for generated
elite-objective floors, showing remaining elite leaders, claim-ready state, and completion from
existing entity metadata. Client bot scenario `44_elite_objective_hud` proves the active state on
the pinned objective floor.

**Quest journal foundation is now visible in the Godot client.** v174 adds a `J`-toggle journal
panel that lists the current floor reward-chest objective from `quest_reward` entity metadata and
marks it complete after the marked chest opens. Bot debug/assertions cover the panel, and
`client/scripts/main.gd`'s file-size baseline was lowered after redundant blank-line compaction.

**Random quest reward chests now have a client-visible marker.** v173 threads server-authored
`quest_reward` metadata through v8 entity views, tracks reward chest ids on generated dungeon
levels, renders a distinct Godot chest marker, and proves the pinned random quest floor with
protocol and client bot scenarios.

**Loot filter mode now persists locally.** v172 saves the existing loot label rarity threshold in
`ClientSettings`, restores it when the Godot client starts, and keeps invalid or missing saved
values normalized to `All`. The feature remains display-only and local; server loot ownership,
protocol, and pickup authority are unchanged.

**Sorcerer and Paladin now have third-row active skills.** v171 adds Sorcerer `arcane_barrage`
and Paladin `sanctuary` to the shared skill catalog, presentations, and i18n text. Validation
guards their class ownership/prerequisite links, Godot skill UI tests cover the expanded class
lists, and protocol bot scenarios prove both new skills cast through the server-owned skill path.

**The v170 engineering review gate is complete.** The review set starts at
[`docs/reviews/20260614_v170-overview.md`](docs/reviews/20260614_v170-overview.md), with backend,
client, and shared/tooling companion reports. It records the v161-v170 feature-and-paydown batch at
`main` commit `05804d77`, notes `make maintainability` passing with 33 grandfathered files / 65347
lines and 37 legacy helper-global injections, and points the next batch toward `game_test.go`
domain drains, typed bot runtime assertion context, `main.gd`/bot runner ownership splits, and more
`validate_shared.py` catalog extraction.

**Main-config gameplay validation now has a focused helper.** v170 adds
`tools/validate_main_config.py` for `main_config` gameplay bounds and dungeon monster drop-source
checks, keeps `validate_shared.py` as the shared-validation entrypoint, adds a bad-drop-rate
regression, and lowers the `validate_shared.py` maintainability baseline from 3149 to 3140 lines.

**Gold auto-pickup tests now live in a focused Go test file.** v169 moves the gold auto-pickup
domain from `game_test.go` into `gold_auto_pickup_test.go`, keeps shared helpers in place, and
lowers the `game_test.go` maintainability baseline from 9116 to 8905 lines.

**Client bot action-step validation now has a focused helper.** v168 adds
`bot_action_step_validator.gd` for key, click, drag, hotbar, stash, multiplayer, market,
blacksmith, and menu action validations while preserving `BotStepCatalog.validate_step` as the
public API. The catalog drops from 322 to 250 lines.

**Protocol bot runtime economy assertions now have a focused helper.** v167 adds
`runtime_economy_assertions.py` for runtime shop/stash counts, details, appraisals, and events while
keeping `run_runtime_assertions` as the public entrypoint. The main runtime assertion module drops
from 520 to 480 lines.

**Client bot UI/menu assertions now have a focused helper.** v166 adds
`bot_ui_assertion_handlers.gd` for menu, character panel, session, multiplayer filter, settings,
pause, and character-info assertions while keeping `BotAssertionHandlers.evaluate` as the public
dispatcher. The main assertion dispatcher drops from 299 to 267 lines.

**Inventory panel routing now reuses the transfer router for slot-kind parsing.** v165 removes the
duplicate equipment slot-kind parser from `inventory_panel.gd`, relies on `InventoryTransferRouter`
for that routing primitive, extends the focused router unit test, and lowers the panel
maintainability baseline from 1559 to 1528 lines.

**Join Game sessions can now be searched and sorted client-side.** v164 adds display-only
search/sort controls to `MultiplayerSessionsPanel`, keeping active-session discovery and joins
server-authoritative while exposing filter state to client bot assertions and extending
`21_join_game_listed_session.json` to prove filtering before joining.

**Inventory transfer routing is now isolated.** v163 moves inventory double-click, shift-click, and
drag/drop routing decisions into `client/scripts/inventory_transfer_router.gd`, preserving existing
shop, stash, market, blacksmith, corpse, unique-chest, equip, unequip, use, and hotbar intent
payloads while lowering `inventory_panel.gd` from 1583 to 1534 lines.

**Elite objective chests now have client presentation.** v162 carries optional `elite_objective`
metadata in v8 entity views, renders marked reward chests with a display-only objective marker in
the Godot client, and adds client bot presentation proof through
`41_objective_chest_presentation.json` while preserving the existing open-lid/glow behavior.

**Elite objective chests now require a full elite leader clear.** v161 strengthens the v159
side-objective gate so objective chests stay locked while any generated pack leader on the floor is
alive, then open through the existing treasure chest path once all leaders are dead. The protocol
bot scenario now targets pack leaders directly with a survivable Barbarian debug setup, and focused
Go coverage proves partial clears still reject.

**Generated dungeon runtime population now has a focused server module.** v160 moves dungeon
population from `sim.go` into `server/internal/game/dungeon_population.go`, preserving generated
stairs, teleporters, chests, elite-objective chest IDs, loose loot, generated monsters, boss visual
selection, rarity scaling, party HP scaling, and corpse spawning behavior. The slice adds focused
population tests, keeps protocol/client behavior unchanged, and lowers the `sim.go` maintainability
baseline from 7022 to 6836 lines.

**The v160 engineering review gate is complete.** The new review set starts at
[`docs/reviews/20260614_v160-overview.md`](docs/reviews/20260614_v160-overview.md), with backend,
client, and shared/tooling companion reports. It records the v151-v159 feature-and-paydown batch at
`main` commit `4a46229e`, notes `make maintainability` passing with 33 grandfathered files / 65747
lines and 37 legacy helper-global injections, and points the next batch toward dungeon population
extraction, `game_test.go` draining, inventory transfer/staging routing, client bot dispatch splits,
protocol bot runtime assertion splits, and `validate_shared.py` paydown.

**Elite objective chests now require objective completion.** v159 preserves v158 generated
elite-objective chest placement, carries the objective identity into runtime `LevelState`, rejects
objective chest activation with `elite_objective_incomplete` until at least one generated pack
leader on that floor has been killed, then reuses the existing treasure chest loot path. The slice
also extracts interactable activation into `server/internal/game/interactables.go`, lowers the
`sim.go` ratchet baseline, trims pre-existing `inventory_panel.gd` line-count drift, and updates
`68_dungeon_elite_side_objective` as the protocol bot proof.

**Barbarian, Rogue, and Ranger have one new higher-row active skill each.** v154 adds
`earthbreaker`, `shadow_flurry`, and `split_arrow` as data-driven shared skill catalog entries with
presentation/i18n metadata, prerequisite validation, focused Go and Godot loader/panel coverage, and
three protocol bot proofs:
`62_barbarian_earthbreaker`, `63_rogue_shadow_flurry`, and `64_ranger_split_arrow`.
The slice deliberately reuses existing skill capability types and leaves existing skill definitions
unchanged.

**Random quest reward floors now seed a first generated quest hook.** v155 rolls roughly 10% of
generated non-boss dungeon floors on a separate deterministic RNG stream, adds one extra reachable
treasure chest on rolled floors, excludes boss floors, and proves the user-facing reward path with
`65_random_quest_reward_floor`. Full NPC offers, quest logs, and durable quest state remain deferred.

**Weapon set swap and hand tabs are playable.** v156 adds two authoritative hand sets, `R` swaps the
active set through `swap_weapon_set_intent`, and the inventory paper doll has set I/II hand tabs for
viewing and equipping either hand configuration. Existing `equipped.main_hand` / `off_hand` remain
the active compatibility fields, while v8 snapshots/deltas expose `active_weapon_set` and
`weapon_sets`. Durable two-set storage and richer tab art remain deferred.

**Skill bindings now have a secondary row.** v157 expands authoritative `function_keys` from 8 to
16 entries, preserving F1-F8 as primary slots and adding Shift+F1 through Shift+F8 as secondary
slots. The store accepts slots 0-15, snapshots/deltas expose the full fixed array, the client maps
Shift+F keys to the secondary row, and `67_skill_secondary_bindings` proves the runtime contract.

**Elite-pack floors now get a small side-objective reward hook.** v158 adds a data-backed
`elite_objective` rule block, places one extra reachable treasure chest only when generated dungeon
monsters include an elite pack leader, and proves the path with `68_dungeon_elite_side_objective`.
Full quest log text and special chest presentation remain deferred.

**Extraction independence is now a maintainability gate.** v151 adds
`scripts/check-extraction-coupling-ratchet.py`, wires it into `make maintainability`, and baselines
the current 43 legacy `helpers=globals()` call sites in `tools/bot/run.py`. New helper-global
namespace laundering now fails CI, reductions must lower the baseline, and `CLAUDE.md` states that
an extracted module only counts when it is importable and unit-testable without importing the source
file or receiving its whole namespace. The dedicated `run.py` split campaign is frozen unless a
future slice performs the real typed `BotContext` refactor.

**The v150 engineering review gate is complete.** The new review set starts at
[`docs/reviews/20260614_v150-overview.md`](docs/reviews/20260614_v150-overview.md), with backend,
client, and shared/tooling companion reports. It records the v141-v149 architecture paydown trend
(`make maintainability`: 33 grandfathered files / 65592 lines, down from 68778 at v140) and points
the next batch toward `game_test.go`, remaining `sim.go` domains, client/Python bot assertion
dispatch, `BotStepCatalog.validate_step`, and `tools/validate_shared.py` paydown.

**Python bot co-op runtime helpers are split out of `run.py`.** v149 moves reusable co-op peer
connect/close, peer pumping, wait/send/accept helpers, player entity selectors, and party role
assertion into `tools/bot/coop_runtime.py`, keeps `run.py` compatibility wrappers, and lowers the
`run.py` ratchet baseline from 4288 to 4269 lines. Scenario-specific co-op proof bodies remain in
`run.py` for future paydown.

**Python bot state ingestion is split out of `run.py`.** v148 moves snapshot/delta ingestion,
teleporter parsing, inventory/stash/hotbar mutation helpers, cooldown decay, active-level clearing,
initial-position tracking, and runtime distance tracking into `tools/bot/state_ingest.py`, keeps
`run.py` compatibility wrappers, and lowers the `run.py` ratchet baseline from 4546 to 4288 lines.

**Python bot wait/pump helpers are split out of `run.py`.** v147 moves accept/reject waits,
event/progression waits, level/teleporter waits, player-position waits, and message pumping into
`tools/bot/wait_runtime.py`, keeps `run.py` compatibility wrappers, and lowers the `run.py` ratchet
baseline from 4612 to 4546 lines.

**Python bot movement runtime helpers are split out of `run.py`.** v146 moves walking,
move-to-position, in-range movement, movement candidate calculation, and movement accept/reject
waiting into `tools/bot/movement_runtime.py`, keeps `run.py` compatibility wrappers, and lowers the
`run.py` ratchet baseline from 4768 to 4612 lines.

**Python bot runtime assertion dispatch is split out of `run.py`.** v145 moves snapshot/runtime
assertion dispatch into `tools/bot/runtime_assertions.py`, keeps `run.py` compatibility wrappers,
and lowers the `run.py` ratchet baseline from 5179 to 4768 lines.

**Client bot runner dispatch is split out of `bot_scenario_runner.gd`.** v144 moves step
catalog/validation, wait dispatch, assertion dispatch, and action dispatch into focused helper
scripts, keeps `BotScenarioRunner` as the public API, and lowers the runner ratchet baseline from
2376 to 1665 lines.

**Client bot facade helpers are split out of `main.gd`.** v143 moves shop, stash, bishop,
blacksmith, market, hotbar, stat, skill-bar, and directional skill-cast bot adapter bodies into
`client/scripts/bot_facade.gd`, keeps `main.gd` public `bot_*` wrappers for `BotController`, adds
headless fake-panel unit coverage, and lowers the `main.gd` ratchet baseline from 6769 to 6703
lines.

**Sim load and player lifecycle helpers are split out of `sim.go`.** v142 moves persisted load
methods and payload clone helpers into `server/internal/game/sim_load.go`, moves co-op player
creation, respawn, spawn selection, context switching, save/default, and sorted player helpers into
`server/internal/game/sim_players.go`, and lowers the `sim.go` ratchet baseline from 7801 to 7045
lines.

**Market store persistence is split out of `repos.go`.** v141 moves listing, offer, expiration,
audit, summary, and market-only helper code into `server/internal/store/market_repo.go` and
`server/internal/store/market_helpers.go`, keeps purchase in `market_purchase.go`, lowers the
`repos.go` ratchet baseline from 3052 to 2315 lines, and updates CODEMAP so future market work
loads focused store files first.

**The v140 engineering review gate is complete.** The new review set starts at
[`docs/reviews/20260613_v140-overview.md`](docs/reviews/20260613_v140-overview.md), with backend,
client, and shared/tooling companion reports. It clears the v140 cadence gate and points the next
batch toward market store extraction, client bot facade/runner split, Python bot assertion split,
and continued CODEMAP/ratchet discipline.

**Market summary and offer reads now expire stale listings first.** v139 makes
`GetMarketSummary` and `ListMarketOffersForSeller` run the existing market expiration sweep before
returning active counts or seller offers. Focused store tests prove those reads restore the seller's
listed item, refund reserved bidder items, clear active counts/offers, and append the
`listing_expired` audit row.

**CODEMAP and the maintainability ratchet are now reduction-oriented.** v138 adds
`docs/CODEMAP.md` as the domain-to-files index, validates its paths through `make validate-shared`,
adds lower-bound file-size ratcheting so shrinks lock in, prints the grandfathered-file trend, and
makes `make ci` run `make maintainability` before the full suite.

**Protocol bot stash assertions now have a focused helper module.** v137 moves stash filtering,
selection, id lookup, count/gold/capacity assertions, and stash-event matching into
`tools/bot/stash_assertions.py`, with direct pytest coverage and no gameplay/protocol behavior
changes. The closeout also repairs brittle existing bot/golden assumptions around debug-gated
scenarios, replay guest ids, and unique item display names, and refreshes the
`server/internal/game/sim.go` file-size baseline for pre-existing drift surfaced by the ratchet;
v137 did not touch the sim.

**The purple unique chest now has a Godot client proof.** v136 adds client bot scenario
`unique_chest_client_proof`, opens `town_unique_chest`, and asserts the `Embercall Blade` plus
`Stormstring Bow` rows expose readable effect summaries in the client stash panel state.

**A second named unique is live in the deterministic chest.** v135 adds `stormstring_bow`, an
enabled ready bow-based named unique with the live `stormbound_echo` effect, and extends rule tests
so both hand-authored named unique payloads and chest rows are covered.

**Unique effect tooltips are readable.** v134 loads `unique_effects.v0.json` in the Godot shared
item rule loader and appends readable unique-effect names plus summaries at the bottom of inventory,
stash/unique chest, and market item tooltips.

**Named unique validation is split and covered.** v133 moves named unique catalog validation out of
the large shared validator body into `tools/validate_unique_items.py`, with Python bad-catalog tests
and Go `LoadRules` parity tests for invalid named unique packages.

**Embercall Blade is now a fixed named unique package.** v132 turns the ready
`embercall_blade` catalog row into a deterministic named item with fixed rolled stats and the live
`everburning_wound` effect. The purple town unique chest now includes Embercall Blade alongside the
effect-coverage rows, and `61_purple_town_unique_chest` proves the named unique can be taken.

**Purple town unique chest is available for testing all current unique effects.** v131 adds a
purple `town_unique_chest` interactable in town. Activating it grants deterministic unique rolled
items covering every enabled ready unique effect exactly once, bypassing inventory capacity only for
this debug/test chest. `61_purple_town_unique_chest` proves the coverage through the protocol bot.

**Market ownership transitions now expire, withdraw, and audit cleanly.** v128-v130 add 24-hour
listing expiration with seller/bidder refunds, bidder-owned offer withdrawal, and
`market_audit_records` for publish, offer, accept/reject/cancel, purchase, listing cancel, and
expiration. The accept-offer regression now proves the bidder loses offered items and receives the
listed item.

**The v130 engineering review is complete.** The review set recommends the next unique-items batch
keep the purple town unique chest server-authored and deterministic, then follow with fixed named
unique packages, unique/effect validation split work, and player-facing unique inspection.
The review pre-task also refreshed the maintainability baseline to current v130 file sizes after
detecting pre-existing ratchet drift; future slices should avoid growing those large files further.

**Town service inventory bridge wiring is split out of `main.gd`.** v127 adds
`client/scripts/town_service_bridge.gd` for market/blacksmith context toggles and inventory staging
intent routing, with a focused headless GDScript test in client smoke.

**Skill validation is split out of the monolithic shared validator.** v126 moves skill class,
Magic Bolt tuning, skill presentation, prerequisite, and skill golden parity checks into
`tools/validate_skills.py` while keeping `make validate-shared` behavior intact.

**Bot skill scenarios can derive skill caps from shared rules.** v125 adds
`max_rank: "from_rules"` support to bot skill progression assertions and migrates the Magic Bolt,
Rage/Heal, and Ranger Volley proofs away from hardcoded skill-cap tuning locks.

**Inventory market and blacksmith actions can stage bag items directly.** v121 shipped
inventory-origin market listings, inventory-origin multi-item offers, and inventory-origin
blacksmith upgrades that reserve items server-side before using the existing authoritative stash
paths. The missing lifecycle closeout was backfilled after audit.

**Ranger is now a playable bow class.** v122 adds Ranger as the fifth class with dexterity-leaning
stats, a green bow icon, a deterministic tall hooded model, starter bow loadout, and protocol bot
scenario `58_ranger_class_foundation` proving creation and ranged basic combat.

**Ranger has its first two bow skills.** v123 adds schema-backed `Piercing Shot` and `Pinning Shot`,
with authoritative line-pierce damage, root movement control, green projectile/icon VFX, a visible
pinning-root marker, and protocol scenario `59_ranger_piercing_and_pinning_shots`.

**Ranger now has a complete three-skill starter kit.** v124 adds `Volley` as an authoritative
multi-arrow fan attack, green Volley icon/projectile presentation, and visual scenario
`60_ranger_volley_and_visual_showcase` covering starter bow plus all three Ranger skills.

**Tuning-friendly rule tests started with the skills panel.** v120 converts the Godot skills-panel
Magic Bolt test to derive requirements, mana cost, and max-rank expectations from `SkillRulesLoader`
instead of copying current shared-rule tuning values. The v120 review set is also complete and
points the next batch toward focused `main.gd`, validator, bot-runner, and test-bucket splits.

**Live unique drops now expose the full enabled effect catalog.** v119 marks named unique metadata
ready, keeps live behavior on rolled equipment `effect_ids`, proves every enabled unique effect can
be selected by at least one compatible template, and adds protocol scenario
`57_live_unique_drops_all_effects` for a deterministic unique drop.

**Item upgrades are now usable from a town blacksmith.** v118 adds a server-authored
`town_blacksmith` service in town and vendor lab, a focused Godot upgrade panel for account-stash
items, and client bot scenario `39_blacksmith_upgrade_ui`. The upgrade route remains authoritative,
and the store now supports both legacy flat rolled stats and current generated rolled-stat payloads.

**Elite command aura radius is now previewed in Godot.** v116 adds optional generated-pack metadata
to monster entity views, renders a display-only radius ring around visible pack leaders whose
followers are server-marked with `elite_command`, and proves the marker plus shared-radius debug
state with client bot scenario `37_elite_aura_radius_preview`.

**Market active offers are now inspectable and acceptable from Godot.** v117 adds seller-side offer
inspection to the market board, lets sellers accept an active item offer through the existing HTTP
contract, refreshes the listing list after acceptance, and proves the flow with client bot scenario
`38_market_active_offer_ui`.

**Market purchase is now usable from the Godot board.** v115 adds a buyer-only `Buy` action for
priced listings, calls the existing v111 purchase route, refreshes active listings, and proves the
flow with a seller-listing preflight plus client bot scenario `36_market_purchase_ui`.

**Market board priced listing UI is now proven in Godot.** v114 adds a deterministic publish price
control, sends `price_gold` through the existing listing-create HTTP route, renders listing prices in
browse rows, and proves stash-item publication through client bot scenario `35_market_board_ui`.

**Elite command aura is now client-readable.** v113 exposes server-owned `elite_command` aura state
through existing monster `effect_ids` when a generated pack follower is actively buffed, then renders
a compact Godot marker on those monsters with a focused client bot proof.

**Elite packs now have an authoritative aura foundation.** v112 adds a data-driven `elite_command`
aura under dungeon monster placement rules, preserves generated pack metadata on live monsters, and
applies a nearby living leader damage bonus to same-pack followers. Aura radius previews, additional
aura types, and richer production VFX remain deferred.

**Market listings can now be purchased for stash gold.** v111 adds optional `price_gold` on
market listings and a direct purchase route that atomically debits buyer stash gold, credits seller
stash gold, delivers the listed item to the buyer stash, marks the listing accepted, and refunds
active item offers. The first purchase proof stays store/HTTP-only; player-facing market UI,
notifications, expiration, fees, and listing edits remain deferred.

**Item upgrades can now repeat with scaling gold costs.** v110 extends the account-stash upgrade
route so equipment can upgrade through `item_upgrade_max_level`, charging
`item_upgrade_cost_gold + current_item_level * item_upgrade_cost_growth_per_level` from stash gold
and preserving deterministic stat mutation. The v110 review also catches the maintainability
baseline up to the post-v109 file sizes so the ratchet enforces from the current repo state.

**Barbarian and Sorcerer now have second combat skills.** v89 adds Cleave as a
server-owned cone weapon attack with pushback and Ice Shard as a cold projectile with stackable
slow plus deterministic shard fan-out; skill visuals now list and replay both skills from the
shared catalog.

**English text now has a shared catalog foundation.** v90 adds `shared/i18n/en.json`, schema and
validator coverage, skill/monster text keys, and a Godot text lookup service with fallback behavior.
Menu, pause, Settings, stat, class summary, skill, skill-bar, and status-effect helpers now resolve
through the catalog so v91 can add Spanish plus the Settings language selector.

**Spanish localization is selectable from Settings.** v91 adds `shared/i18n/es.json`, validates locale
catalogs against English, persists the selected language, and refreshes menu, pause, and Settings
labels immediately while falling back to English for missing keys.

**Town bishop respec service is live.** v92 adds a red `town_bishop` interactable that heals HP/mana
on activation and opens a compact service panel with a 250 gold Respec action. The server owns gold
deduction, stat reset/refund, skill rank refund, cooldown clearing, resource refill, and rejection
when the player cannot afford the service.

**Market listings now accept multi-item offers.** v93 adds active/accepted/rejected market offers
backed by 1-10 bidder stash items. Sellers can inspect offers, accept one offer to atomically swap
the listed item for offered items through account stashes, or cancel the listing and refund all
active offers.

**The first item upgrade action is server-owned.** v94 adds main-config tuning for starter upgrade
cost/max level plus an authenticated account-stash upgrade route. The store spends stash gold,
increments `item_level`, and increases one existing rolled stat deterministically while preserving
market eligibility for upgraded items.

**The unique item catalog has a disabled seed.** v95 adds schema-backed `unique_items.v0.json` with
`embercall_blade` as a non-player-facing unique concept. Validation cross-checks the base template
and keeps the seed disabled until a future behavior-changing unique effect path exists.

**Town now reads as a wider hub.** v96 distributes town services at least 5 tiles from the central
campfire, adds two procedural wood cabins, improves the town ground texture, and adds
`$showme --focus town` for focused visual feedback without changing server authority.

**New heroes now get class starter kits.** v97 seeds explicitly created paladins with sword/shield,
sorcerers with a two-handed staff, and barbarians with a slower harder-hitting axe, plus one health
and one mana potion. The starter staff also introduces item-backed max mana and skill damage scaling.
Follow-up minor improvements after v97 added dedicated starter staff/axe models, item-family
presentation assets, class-specific character models, magic scaling for existing skill effects, and
floor-loot presentation fixes. These are considered unversioned polish/consolidation commits, not
separate numbered slices; the next gameplay slice remains v98.

**Rogue class foundation is playable.** v98 adds Rogue as the fourth selectable class with a slimmer
deterministic character model, dagger class icon, dexterity-leaning starting stats, and a durable
starter kit of two common swords plus one health and one mana potion. Rogues can equip one-handed
melee weapons in `off_hand`; non-Rogue classes still cannot.

**Rogue starter skills are authoritative.** v99 makes Poison Stab deal weapon damage plus
rank-scaled poison ticks, makes Dash move through and damage crossed monsters from shared skill
data, and gives Rogues independent off-hand basic attacks at 1.5x main-hand cadence. The Rogue
foundation bot scenario now learns Dash and Poison Stab, dashes through a target, poisons it, and
observes two main-hand attacks plus one off-hand attack.

**Damage types and monster resistances are authoritative.** v100 adds canonical `force`, `cold`,
`poison`, and `lightning` damage types, skill/item fallback to `force`, monster resistance maps,
and `damage_type` on combat events. Lightning now deals half damage to flying lab/bat targets and
50% bonus damage to quadruped/wolf targets, proven through focused Go tests and the
`damage_types_and_resistances` protocol bot scenario.

**Undead poison immunity is playable and visible.** v101 adds a localized `dungeon_undead` monster
with full poison resistance, a generated skeleton GLB/scene wired through the monster visual
catalog, and a compact lab scenario. Poison Stab now applies poisoned status on a connected hit
even when resistance mitigates the hit to zero, and poison ticks against undead emit authoritative
zero-damage poison events instead of lowering HP.

**Every playable class has a foundation visual scenario.** v102 adds Paladin, Barbarian, and
Sorcerer class-foundation protocol/visual scenarios alongside the existing Rogue scenario. Each
scenario proves starter gear, movement, at least three basic attacks, and every current class skill;
Python coverage now fails if a playable class lacks a foundation scenario or a class skill is not
referenced by that class scenario.

**Skill visual replays now seed requested rank directly.** v88 lets `make skill-visual
skill=<id> rank=<n>` start from the requested class, minimum level/stats, and skill rank without
first killing an XP dummy or allocating a skill point during the replay.

**Combat/world state now persists on same-session resume.** v5 replays recorded
inputs before the WebSocket `session_snapshot`, so monster death, player HP,
inventory, equipped state, and ID continuity are restored authoritatively.

**World preset identity now persists on sessions.** v7 stores `world_id`, so fresh WebSocket attach,
resume, `/state`, replay verification, and replay timeline all reconstruct the same initial layout.

**Equipped weapon damage now changes authoritative combat.** v8 resolves `rusty_sword.damage`
from equipped server state at attack time and proves the equipped gear scenario kills the reward
dummy in one acknowledged attack.

**Solid collision now blocks movement through bodies and walls.** v9 resolves player movement
against live monsters and static world walls, while collision lab proves routed movement and
deterministic replay.

**Click action and melee reach are now authoritative.** v10 unifies combat/pickup/door activation
behind `action_intent`, enforces reach from shared rules, and proves a replayable opening door.

**Click-to-move and action auto-approach are now authoritative.** v11 adds shared navigation
rules, deterministic server A*, `move_to_intent`, and a path-maze bot proof.

**Ranged projectile combat is now authoritative.** v12 adds ranged weapon rules, projectile
entities, swept collision, impact-time hit/damage, and a ranged-lab bot/replay proof.

**Inventory UI, unequip, and player drop are now authoritative.** v13 adds protocol-backed
unequip/drop intents, deterministic adjacent loot placement, persisted inventory removal, and a
display-only Godot panel that mirrors server snapshots/deltas.

**Current item presentation is now shared-data-driven.** v15 adds presentation metadata for all
current item definitions and uses it for inventory icons and ground loot silhouettes without
server/protocol changes.

**Consumable healing is now authoritative.** v16 adds `use_intent`, red potion heal rules, HP cap
goldens, server-owned inventory removal, and a client-only hotbar that sends use intents.

**Monster chase movement is now authoritative.** v17 adds opt-in chase behavior, deterministic
monster pathing around solids, leash return, chase/lab bot scenarios, and client walk presentation
from position deltas.

**Dungeon levels, stairs, teleporters, town entry, and dungeon monster threat are now authoritative.**
v18 adds multi-level dungeon state and generated stairs; v19 adds generated teleporters, session
discovery, and server-owned fast travel with a client-only waypoint panel; v20 makes town level `0`
the fresh play-session entry and keeps dungeon floors lazy; v21 spawns deterministic hostile dungeon
mobs that chase and proactively damage the player.

**Character inventory/equipment and waypoint unlocks now persist across fresh sessions.** v22 moves
durable item instances and discovered waypoint levels to the default character, while preserving
session-start snapshots for deterministic replay and keeping HP, dungeon maps, monsters, corpses,
opened doors, and floor drops session-scoped.

**Dungeon mobs now drop rolled weapon gear.** v23 adds server-authoritative item templates,
deterministic rarity/stat rolls, rolled weapon damage, rolled payload persistence, and tooltip
presentation for the first rolled weapon template.

**The client now has a player-facing menu shell.** v24 adds named character list/create APIs,
fresh-session Continue/New Game flows, local window-size settings, ESC pause, Return to Main Menu,
and a Godot client bot proof for the complete menu path.

**Treasure classes and guarded chests are now authoritative.** v25 adds data-driven multi-attempt
monster/chest loot, deterministic rare chest generation with guarded monster bonus, open-once chest
loot, and bot/golden coverage for the complete path.

**Character stats and leveling are now authoritative.** v26 adds durable XP, levels, stat points,
base stats, derived substats, stat allocation, VIT max-HP effects, STR damage contribution, a
Godot character sheet, an XP bar, and protocol/client bot proofs.

**Sustained left-click controls are now client-side.** v27 adds hold-to-attack on monsters and
hold-to-move on floor by repeating existing `action_intent` / `move_to_intent` at `SEND_INTERVAL`,
with sticky targets, move epsilon, and headless unit coverage — no protocol or server changes.

**Full paper-doll equipment and belt-gated hotbar are now authoritative.** v28 replaces the single
weapon slot with full equipment slots, two-hand occupancy, droppable gear templates, persisted
character hotbar layout, replay-safe session hotbar snapshots, and protocol/client bot proofs for
server-synced paper-doll and belt capacity behavior.

**Generated dungeon drops now reach the expanded equipment catalog.** v29 adds temporary depth
bands, depth-specific monster/chest treasure classes, validation for full v28 template reachability
by depth `3+`, golden fixtures for varied equipment outcomes, and a real generated dungeon bot proof.

**Generated dungeon monster rarity now scales challenge and loot depth.** v30 adds deterministic
generated monster rarity tiers, scaled HP/damage/XP, effective monster loot depth offsets,
monster rarity in protocol/replay state, player/enemy tinting, and a real generated dungeon bot
proof for non-common rarity.

**Combat stats now affect authoritative outcomes.** v31 applies hit, crit, armor, block, minimum
damage, and effective stat breakdowns across player and monster combat, then renders normal, crit,
miss, and block feedback from protocol events in Godot.

**The test floor now separates contracts from tuning details.** v32 keeps exact locks for replay,
schema, formula parity, persistence boundaries, and named UI/protocol contracts, while converting
brittle dungeon size, generated population, movement timing, rarity tuning, and selector-index
assumptions to semantic, range, derived, or eventual checks.

**True two-player co-op sessions are now authoritative.** v33 adds server-owned co-op session
membership, hashed join codes, actor-tagged inputs, per-player sim state, recipient-scoped realtime
snapshots/deltas, remote-player Godot rendering, and a protocol bot proof for join, movement,
disconnect/reconnect, and replay.

**Character-like model reactions are now unified in the Godot client.** v34 adds client-only
hit/death transform and tint reactions for local players, remote co-op players, and monsters;
remote co-op players now reuse the humanoid character model with a distinct dark tint.

**The first boss floor gate is now authoritative.** v35 adds a compact level `-5` boss arena,
telegraphed boss phases, locked down-stair/teleporter exits until boss death, boss visual scale
metadata, and protocol/replay/bot proof for unlock and descent.

**Inventory capacity and the paper-doll bag grid are now authoritative.** v36 adds server-derived
`inventory_rows` / `inventory_capacity`, an item-granted row source, full-bag and overflow rejection
guards, a 5-column capacity grid, and protocol/client bot proofs.

**Combat control and boss AI fixes are now authoritative.** v37 adds server-owned directional attacks,
authoritative stop movement, aggro-on-hit with nearby contagious group aggro, boss chase/damage repair,
and protocol/client unit proofs.

**Session browser and uncapped co-op are now authoritative.** v38 adds persisted listed co-op sessions,
active session summaries, listed join without join code, three-plus-member realtime/replay proofs, a
Godot Multiplayer menu path, and local/remote multi-client menu launchers. Empty listed sessions are
hidden from discovery, and a listed session is ended when its last connected player disconnects.

**Character gold, mana, and related UI polish are now authoritative.** v39 adds durable character
gold, currency loot pickup, generated gold scaling, snapshot/delta/replay wallet coverage, player
mana, blue mana potions, DEX-sourced armor, and Godot HUD/inventory/menu polish.

**Reachable generated dungeon obstacles are now authoritative.** v40 adds deterministic generated
interior dungeon walls, obstacle reachability retries, authoritative protocol wall layouts,
Godot server-layout rendering, and protocol/client bot proofs that generated walls exist without
blocking generated targets.

**The town vendor and first gold sink are now authoritative.** v41 adds the `town_vendor`, protocol
v4 shop buy/sell contracts, fixed potion stock, deterministic generated offers based on deepest
dungeon depth, durable gold mutations, deepest-depth persistence, and protocol/client bot proofs for
shop open, buy, sell, reconnect, replay, and fresh-session persistence.

**Vendor appraisals and direct item comparison are now authoritative.** v42 extends `shop_opened`
with server-authored summary, appraisal, and comparison views, and proves the richer protocol plus
Godot panel through protocol/client bot scenarios.

**Equipment requirements and equip previews are now authoritative.** v43 expands item-template
requirements to level/base stats, rejects unmet equips before mutation, annotates loot/inventory/shop
views with server-authored requirement status and equip-preview deltas, and proves the path through
protocol and Godot client bot scenarios.

**Skill points and Magic Bolt are now authoritative.** v44 adds durable skill points/ranks,
protocol v5 skill state, attack-speed-derived cooldowns, a server-owned Magic Bolt cast/reject/recover
loop, and protocol/client bot proofs through replay, reconnect, and fresh-session persistence.

**Menu Create Game and Join Game flows now match the backend session model.** v45 replaces the
player-facing Continue/New Game/Multiplayer root menu with Create Game, Join Game, Settings, and
Exit; persists the Create Game Type setting; and proves co-op/solo create plus Join Game empty-state
behavior through client bot scenarios.

**The real Godot Join Game path now has a multi-account listed-session proof.** v46 adds a
client-bot preflight host that holds an active listed co-op backend session, then drives a separate
Godot guest through Join Game, character selection, listed join, WebSocket connect, and remote-player
presence assertions.

**Town vendor stock is now finite and refresh-gated.** v47 persists per-character generated stock,
consumes purchased generated offers, refreshes stock only on newly unlocked non-town waypoints,
limits shop rarity to `rare`, and keeps buyback rows session-local and cleared when the actor leaves
town.

**Co-op rewards and monster scaling are now authoritative.** v48 grants full monster XP to nearby
eligible party members, excludes dead/disconnected/far/different-level members, scales monster
HP/damage logarithmically with active same-level party count, and routes private progression by
recipient owner.

**Gold is now auto-pickable, but loot stays shared.** v49 keeps one shared floor entity per drop,
adds passive gold pickup for the first eligible player in deterministic order, and leaves non-gold
items click-required. Personal loot, reservations, hidden/duplicated drops, shared/split gold, and
item auto-pickup remain deferred/non-goals.

**Account stash storage is now authoritative.** v50 adds a town stash interactable, protocol v7
stash contracts, account-owned item/gold persistence, replay-safe session-start stash snapshots,
server-owned item/gold transfers, owner-private realtime fanout, and protocol/client bot proofs for
item and gold storage across fresh sessions.

**Mystery seller core is now authoritative.** v51 adds a town mystery seller, protocol v8 concealed
shop rows, deterministic per-character hidden stock, reveal-on-purchase events, and protocol/client
bot proofs for hidden offers, purchase reveal, replay, and fresh-session consumed stock.

**Ranged monster AI is now authoritative.** v52 adds generated dungeon archers, data-driven
melee/ranged monster composition, server-owned monster projectiles that respect walls and target
players, and a minimal Godot bow marker with protocol/client bot proofs.

**Boss health bar UI is now client-visible.** v53 adds a top-center Godot boss health bar driven by
existing authoritative boss entity hp/max hp and metadata, plus client unit and bot scenario proof
for the first `cave_warden` boss floor.

**Character select summaries are now server-authored.** v54 extends `GET /v0/characters` with
level, gold, and deepest-depth summary fields, renders them in the Godot character picker, and
proves the menu path with focused store, HTTP, client-unit, and client-bot coverage.

**Monolith decomposition and quality gates are now in place.** v55 proves that the god-file
tax from the v53 review can be paid down without behavior change: the sim.go handler registry
(handlers.go, −1,056 LOC from sim.go) means new intents never touch the dispatcher; ItemRulesLoader
eliminates ×5 GDScript item-loader duplication; ShopRNG and bot_types.py are now importable
independently. The determinism lint (`make lint-determinism`) converts the core sim invariant from
CLAUDE.md prose to a failing CI step; `make regen-golden` closes the manual-edit correctness
hazard on golden fixtures; and `test_delta_apply.gd` adds the first unit coverage to the
highest-risk zero-tested client code. All 265 Go tests, 59 Python tests, and 15 GDScript unit
tests pass; CI is now 9 phases.

**Generated monster attacks are a little faster.** v56 tunes regular generated dungeon monsters
without changing damage, movement, bosses, or lab fixtures: `dungeon_mob` cooldown is now 32 ticks
and `dungeon_archer` cooldown is now 75 ticks. The dungeon monster attack golden owns the melee
cooldown, Go/GDScript golden checks cross-check it against shared rules, protocol bot scenarios
prove archer damage, boss-floor traversal, and skill-progression combat, and a missing
`item_rolls.json` description field found by `make validate-shared` is restored.

**Boss phase readability is now client-visible.** v57 keeps server combat unchanged and adds
display-only phase state to the Godot boss health bar: phase kind, pattern id, phase index,
duration, remaining ticks, and phase ratio. Telegraph phases now attach a primitive
`BossTelegraphMarker` under the boss using server-authored radius/color, and the client bot runner
can assert both the bar countdown and the in-world marker.

**Cave Warden has a second boss pattern.** v58 adds `ground_slam`, a data-driven circle telegraph
and active hit shape, to the Cave Warden deck after `charged_melee`. The Go sim now cycles boss
pattern decks deterministically in declared order, resolves circle boss hits server-side, and the
protocol bot can assert phase events by payload fields such as `pattern_id`.

**Magic Bolt is now catalog-driven.** v59 moves Magic Bolt into a schema-backed skill catalog with
class/tree metadata, bounded requirement/cost/damage/projectile/cooldown helpers, and a separate
skill presentation catalog. The server now validates supported skills generically and enforces
`magic >= 15` for both learning and casting; Godot resolves skill panel and hotbar labels/tooltips
from shared skill data while server progression and cooldown state remain authoritative.

**The first content-library manifest is live for skills.** v60 adds a schema-backed
`shared/content/content_libraries.v0.json` index for skill rules and skill presentations. Go and
Godot loaders now resolve skill content through manifest paths while runtime state, protocol,
replay, goldens, and UI state keep stable skill IDs such as `magic_bolt`. Validation and focused
tests prove relative path resolution, duplicate-ID rejection, and unknown manifest group rejection.

**Rage and Heal are now authoritative active skills.** v61 expands the data-driven skill catalog
with closed declarative effect rows for a self stat-percent buff and an allied area percent heal.
The server owns mana, cooldowns, buff expiry, max-HP sync, visual scale metadata, and skill-sourced
healing events; the Godot skill tree and hotbar now select among multiple first-row skills.

**Generated monster stats now scale by dungeon depth.** v62 moves regular generated dungeon monster
HP, damage, XP, and related combat pressure onto depth-aware shared rules while keeping boss
templates bespoke. Go and GDScript golden checks prove the scaling rules, and protocol bot coverage
keeps generated dungeon combat replayable.

**Default sim construction now returns errors instead of panicking.** v63 changes the exported
`game.NewSim` default-world constructor to return `(*Sim, error)`, adds explicit `MustNewSim`
panic behavior for tests with known-valid fixtures, and covers invalid default-world construction
without crashing. This closes the v60 backend review's runtime sim construction finding.

**Mystery seller paid reroll is now authoritative.** v64 adds the `shop_reroll_intent`,
a server-owned 50 gold spend, deterministic `|reroll:N` stock refresh keys, complete replacement
of concealed mystery stock, and protocol/client bot proofs for reroll, replay, and fresh-session
persistence.

**Stash search and sorting are now client-visible.** v65 adds display-only Godot controls for
searching and sorting bag/stash rows by acquired order, name, rarity, or slot, while keeping all
deposit/withdraw mutation keyed by server-authored `stash_item_id` / inventory item IDs.

**Progress backlog hygiene is current through v66.** v66 corrects the canonical discovery metadata
after v64/v65 by marking shipped candidates complete, adding their scenario catalog entries, and
narrowing deferred backlog text to still-open adjacent work.

**Boss kill reward status is now explicit.** v67 emits a dedicated `boss_killed` event with
`boss_template_id` for boss deaths while preserving the existing `monster_killed`, loot, XP, and
exit-unlock flow. The Godot client now exposes a `Cave Warden defeated` reward status, and protocol
plus client bot coverage prove the boss-specific signal.

**Market listings now have a stash-backed foundation.** v68 adds active/canceled market listing
persistence and authenticated HTTP routes to create a listing from an account stash item, browse
active listings, and cancel an owned listing back to stash. Offers, purchases, pricing, expiration,
and Godot market UI remain deferred.

**Character class identity is now authoritative.** v69 persists `character_class`, exposes it in
character APIs, validates create requests against shared progression class rules, and uses class
rules to seed fresh progression. The default `barbarian` preserves the prior baseline stats while
`sorcerer` and `paladin` prove divergent starts.

**Class gates now affect gameplay.** v70 maps each skill to its class, rejects cross-class skill
spend/cast attempts, adds one fixed class-required weapon per class, and rejects wrong-class weapon
equips. Session-start snapshots and realtime reconstruction carry class identity so restrictions
survive the authoritative boundary.

**v70 engineering review steering.** The v70 review recommends a small maintenance follow-up for
realtime fanout level snapshots and defensive `equipped_update.slot` handling in Godot/Python bot,
plus keeping the v71 class picker contained in character-select UI with shared class presentation
metadata where practical.

**Class creation is now player-facing.** v71 adds class picker blocks with code-native class
sprites, hover tooltips for class stats/skills, selected-class create request plumbing, and class
icons at the start of character rows. Client unit and bot coverage prove Sorcerer selection without
changing server authority.

**Monster visuals are now catalog-driven.** v72 adds shared monster visual metadata, deterministic
quadruped/flyer placeholder assets, wolf/bat monster definitions with unchanged chase mechanics,
and deterministic boss model pools across dummy, quadruped, and tiny flyer visuals. Godot resolves
monster scenes through the catalog, and the showme monster lineup was approved before final CI.

**Stats and skills now use draggable window chrome.** v73 adds a reusable Godot titlebar shell with
close button, titlebar-only dragging, viewport clamping, and debug proof, then migrates the
character stats and skills panels without changing server authority or gameplay protocol.

**Gameplay item panels now share draggable chrome.** v74 migrates inventory, shop, and stash onto
the reusable titlebar shell while preserving item drag/drop, buy/sell, reroll, stash search/sort,
and existing gameplay-panel APIs.

**Custom gameplay window layout now persists locally.** v75 saves and restores clamped positions
for character stats, skills, inventory, shop, and stash through `user://window_layout.cfg`, while
disabling normal persistence during client unit tests.

**Main gameplay config foundation is now validated.** v76 adds `main_config.v0.json`, exposes it
through the Go rules loader, and adds drift guards against current combat, movement, and dungeon
monster drop defaults until follow-up slices consume those values directly.

**Main gameplay config now drives attack cadence and movement.** v77 makes
`base_attack_interval_ticks` and `base_movement_speed` operational server gameplay inputs, with
focused tests proving edits to `main_config.v0.json` take effect without touching older rule files.

**Main gameplay config now drives dungeon monster drop chance.** v78 applies
`base_drop_rate_percent` to dungeon monster treasure-class primary attempts during rules loading,
so the global drop chance can be tuned from `main_config.v0.json` without hand-editing each depth
class.

**Generated dungeon fights now form packs.** The pack-aggro slice adds data-driven pack sizing,
monster assist radius, deterministic close pack placement, and a protocol bot proof that damaging
one generated monster can emit multiple `monster_aggro` events. This landed after the already-used
v76-v78 main-config slice numbers, so the as-built/spec files retain the requested v76 label while
the canonical lifecycle continues through v79/v80.

**Generated packs now have role and leader foundations.** v79 adds internal pack roles, pack
composition constraints, and deterministic elite leader markers so future elite behavior can build
on structured encounters without exposing new protocol fields yet.

**Combat threat readability is now visible.** v80 maps existing authoritative `monster_aggro`
events to display-only `AGGRO` floating text in the Godot client, adds a `threat` damage-number
variant, and proves it with client unit, focused client-bot, and protocol pack-aggro coverage.

**v80 engineering review steering.** The v80 review keeps the repo at 8.4/10 overall and recommends
small follow-ups for combat event presenter extraction and splitting the largest Python validation/bot
files by domain. The realtime fanout level snapshot finding was closed in v82, defensive client
payload parsing was closed in v83, and client bot step registry duplication was closed in v84.

**v90 engineering review steering.** The v90 review keeps the repo at 8.5/10 overall, confirms
`make ci` green after blocker cleanup, and steers the next batch toward localization: central English
text keys, Spanish translations, Settings language selection, and English fallback. It also keeps
the standing rule that touched large files should be split or shrunk rather than extended.

**v100 engineering review steering.** The v100 review keeps the repo at 8.5/10 overall and steers
the next gameplay batch toward shared damage types, data-driven monster resistances, and focused
combat helpers/tests. Undead poison immunity should consume the same resistance contract rather
than introducing a one-off immunity flag. The review gate also records a maintenance exception for
the already-landed v99 growth in `main.gd`, `test_item_visuals.gd`, `sim.go`, and `tools/bot/run.py`;
future slices touching those files should split or shrink them before adding more behavior.

**Maintainability ratchet is now explicit.** New source/test/tool files target a 600-line maximum,
existing over-limit files are grandfathered in `.maintainability/file-size-baseline.tsv`, and
`make maintainability` enforces that new files do not exceed the target while grandfathered files
do not grow by more than 25 lines without a documented maintenance exception.

**Paladin Holy Shield is now authoritative and visible.** v81 adds a data-driven Paladin area
defensive buff, per-player active effect ids, armor/block stat application, server-owned expiry,
Rage-style status UI presentation, and a gold shield/shine around every affected hero.

**Realtime fanout now uses tick-time client level snapshots.** v82 captures connected client levels
under the session loop mutex and passes that snapshot into fanout, closing the v70/v80 review
finding without changing protocol, gameplay, or client presentation. The slice also removes stale
attack-interval-derived exact expectations from Magic Bolt, Rage, Heal, Holy Shield, and matching
client bot scenarios that CI surfaced; exact cooldown math remains owned by shared golden tests.
The model-reaction client scenario now uses a safe low-HP lab dummy for terminal reaction proof
instead of depending on a long basic-attack sequence against combat-stat targets.

**Client envelope payload parsing is now defensive.** v83 routes central Godot `_handle_message`
payload access through a dictionary guard, so missing/null/non-dictionary payloads on accepted,
rejected, error, and delta envelopes no longer crash the client message boundary.

**Client bot step registration now has one source of truth.** v84 derives `ALL_STEP_TYPES` from
the wait/assert/action category arrays in `bot_scenario_runner.gd`, preserving unknown-step
validation while removing a duplicated maintenance list.

