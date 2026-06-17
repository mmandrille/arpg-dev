# Bot and smoke scenario catalog

Reference catalog for protocol bot and client bot scenarios. Read when adding or debugging
end-to-end proofs — not required for routine slice startup.

## Scripted vertical slice flow (bot + smoke)

Every slice keeps this loop working unless the spec explicitly changes it:

```text
dev-login → create session → move → attack training dummy → pick up loot → equip rusty_sword
```

After v4 the player **survives with reduced HP** (`hp < 10`). Monster dies; player may take retaliation
each successful hit. After v7 this flow lives in `tools/bot/scenarios/01_vertical_slice.json`; additional
scenario JSON files are automatically included in filename order in `make bot` and `make bot-visual`.
Every protocol bot scenario has a hard **10.0 second** full-run budget. When a proof grows past
that, shorten setup to the behavior under test with compact lab worlds or focused lower-level tests
instead of waiting through unrelated traversal, farming, or natural timing loops.

The scenario catalog also includes:

```text
gear_before_combat: walk to rusty_sword → pick up → equip → one-shot reward dummy → pick up training_badge
collision_lab: pass through middle wall gap → kill monster on far side
inventory_lab: pick up rusty_sword → equip → unequip → drop → re-pickup → re-equip
heal_lab: pick up red_potion x2 → take damage → use potion twice → full HP
chase_lab / chase_maze / leash_lab: wait while chase monster closes; kite beyond leash and return
dungeon_levels / teleporter_lab: start in town, descend/ascend generated floors; discover teleporters and fast-travel back
character_persistence: same-account fresh sessions retain gear/equipment and discovered waypoint access
rolled_drops: kill dungeon mob → pick up/equip rolled cave_blade → prove rolled metadata persists
main_menu_flow: Create Game root flow → settings → listed co-op session → pause input lock → return → existing-character fresh session
treasure_classes_and_guarded_chests: pinned chest floor → kill guarded mob → open chest once → pick up chest loot
character_stats_and_leveling: descend to dungeon → kill mobs for XP → level up → spend VIT → prove persistence
full_equipment: pick up/equip paper-doll gear → prove hand occupancy → assign belt-gated hotbar → prove persistence
dungeon_equipment_drops: compact equipment lab → pick up/equip rolled equipment → prove persistence; depth-banded generation stays in lower-level tests/goldens
monster_rarity_loot_scaling: descend to generated dungeon → assert champion rarity → kill → pick up rolled loot → prove persistence
combat_stat_effects: combat lab proofs for miss, crit, armor floor, block, monster crit/block, projectile impact, and stat breakdowns
client_combat_feedback: equip gear → assert stat breakdowns → prove normal/crit/miss/block floating text and settings toggle
true_coop_session: host creates co-op → guest joins → shared-level visibility → independent movement → disconnect/reconnect → replay proof
model_reaction_polish: attack training dummy → prove monster hit reaction → prove local player hit reaction → kill dummy → prove terminal corpse reaction
boss_floor_gate: start on compact boss floor → assert locked exits → observe boss phase telegraphs including `stone_lance` and `summon_wolves` → assert summoned wolf adds → kill boss → unlock exits → descend to -6
boss_enrage_phase: compact boss floor → damage Cave Warden below threshold → observe server-authored `boss_enraged`
boss_kill_reward_polish: compact boss floor → kill Cave Warden → observe `boss_killed` with `boss_template_id` and client reward status
mercenary_hiring_board: compact town lab → open mercenary board → spend gold to hire one guard → prove hired guard follows and damages a target
mercenary_roster_ui: headless Godot client opens the mercenary board → hires the fixed guard → asserts the roster panel and mercenary companion HUD icon
companion_stance_command: compact town lab → hire a guard → switch passive/assist stances → assert stance events/state and returned combat behavior
paladin_class_foundation / barbarian_class_foundation / sorcerer_class_foundation / rogue_class_foundation / ranger_class_foundation: class starter gear → movement → at least three basic attacks → all current class skills
ranger_piercing_and_pinning_shots: Ranger casts Pinning Shot to root a chase target, waits for expiry, then fires Piercing Shot through lined-up monsters
ranger_volley_and_visual_showcase: Ranger shows starter bow, Pinning Shot, Piercing Shot, and Volley in a compact visual lab
sorcerer_arcane_barrage: Sorcerer casts Arcane Barrage and proves authoritative projectile damage
paladin_sanctuary: Paladin casts Sanctuary and proves authoritative area defense effects
inventory_capacity_and_paper_doll: fill base 15-capacity bag → reject full pickup → equip capacity belt → fill expanded 20-capacity bag
combat_control_and_boss_ai_fixes: equip training bow → fire directional free shot → prove damage, group aggro, and monster movement
session_browser_uncapped_coop: host creates listed co-op → two peers join from active list → prove three-player visibility, disconnect/reconnect, and replay
ui_currency_and_mana_polish: pick up gold instead of reward badges, persist character wallet, and use/reject blue mana potions
reachable_dungeon_obstacles: descend through generated dungeon floors → assert generated interior wall layout → route to loot beyond obstacles → prove replay
dungeon_wall_rendering: headless Godot client descends to generated floors → assert authoritative non-perimeter wall rendering state
vendor_appraisal_quotes: open compact vendor lab with rolled loot → assert server-authored offer summaries, comparisons, sell appraisals, buy, sell, replay
vendor_item_comparison: headless Godot client opens vendor → assert visible offer/sell details, comparison rows, buy, and sell
shop_stock_lifecycle: compact vendor generated stock → sell-to-buyback → rebuy → fresh-session buyback cleared and generated stock retained
client_shop_stock_lifecycle: headless Godot client opens vendor → sell to buyback → fixed purchase refresh → sell/rebuy buyback → assert fixed/generated rows remain visible
equipment_requirements_and_preview: pick up requirement-gated gear → reject unmet equip → level and allocate STR → equip → prove persistence
client_equipment_requirements_and_preview: headless Godot client opens inventory → assert requirement-status and equip-preview rows
skill_points_and_magic_bolt: level to 5 → learn Magic Bolt at baseline Magic 5 → cast → reject rank 2/cooldown recast → recover → prove replay/fresh persistence
client_skill_points_and_magic_bolt: headless Godot client opens skill panel → proves baseline Magic 5 availability and rank 2 Magic 8 gating → observes skill bar cooldown and recovery
rage_and_heal_skills: level to the second skill-point grant → learn Rage and Heal → cast Rage → fresh heal_lab session casts Heal and proves skill-sourced healing
menu_create_join_flow: Join Game empty state → Settings Create Game Type Solo → solo Create Game → existing-character fresh session
join_game_listed_session: protocol host holds active listed co-op session → Godot guest joins via Join Game → remote host visible
coop_rewards_and_scaling: compact three-account co-op → nearby host/guest share full XP → out-of-range guest excluded → replay/fresh persistence; different-level exclusion stays in lower-level tests
gold_autopickup_shared_loot: compact co-op loot lab → shared floor gold race → lowest player id wins private wallet update → item loot still requires click
account_stash_storage: acquire dungeon loot/gold → open town stash → deposit/withdraw item and gold → replay/reconnect/state/fresh session persistence
market_stash_listing_foundation: HTTP/store proof creates active market listing from stash item → browse active listings → reject foreign cancel → cancel back to stash
client_account_stash_panel: headless Godot client opens stash → verifies bag/stash item sync → deposits/withdraws item and gold
blacksmith_upgrade_ui: headless Godot client funds stash gold, deposits a rolled stash item, opens town blacksmith, upgrades once, and asserts item level/gold changes
live_unique_drops_all_effects: compact protocol lab picks up a deterministic unique rolled item and asserts its live effect_ids payload
ranged_monster_ai: compact archer lab → assert dungeon_archer → observe archer-sourced ranged player damage; generated archer placement stays in lower-level/client coverage
client_ranged_monster_ai: headless Godot client descends to generated dungeon → asserts bow marker → observes ranged player damage
client_boss_health_bar_ui: headless Godot client descends to first boss floor → asserts Cave Warden boss health bar
client_boss_phase_readability: headless Godot client descends to first boss floor → asserts boss phase countdown and telegraph marker
character_select_summaries: headless Godot client opens Create Game → asserts character row level/gold/depth/status summaries
mystery_seller_paid_reroll: open mystery seller → spend 50 gold to reroll concealed stock → prove old offers are replaced, gold persists, and replay/fresh stock remain deterministic
unique_chest_client_proof: headless Godot client opens the purple unique chest and asserts named unique rows expose readable effect summaries
stash_search_and_sorting: headless Godot client opens stash → searches and sorts bag/stash rows → deposits/withdraws by stable server IDs
elite_objective_minimap_pin: headless Godot client descends to a deterministic elite-objective floor → asserts compact minimap pin visibility and active debug state
my_market_offers_panel: headless Godot client opens market board → loads My Offers → asserts an outgoing offer row for a foreign listing
buyer_offer_cancel_ui: headless Godot client opens My Offers → cancels an outgoing offer → verifies the offer is canceled and the offered item returns to stash
market_trade_receipts: headless Godot client cancels an outgoing market offer → opens Receipts → verifies the account-scoped `offer_canceled` receipt
mystery_seller_silhouettes: headless Godot client opens the mystery seller → verifies concealed rows include safe slot-derived silhouette clues
material_wallet_details: headless Godot client auto-picks an upgrade shard → verifies the compact wallet tooltip includes catalog-backed details
blacksmith_recipe_selector: headless Godot client stages a blacksmith item → verifies the active recipe appears in the preview
mercenary_stats_card: headless Godot client hires a mercenary → verifies the companion panel stats card shows HP, stance, state, and id
```

**Verify:**

```bash
make db-up && make server    # terminal 1
make bot                     # terminal 2 — all protocol bot scenarios
make client-unit             # headless Godot unit gates (no server required)
make client-smoke            # headless Godot gates + slice smoke
make bot-client              # Godot client bot scenarios; requires live server
make ci                      # full suite
make bot-visual              # optional — record all bot scenarios and watch replay playlist in Godot
make bot-visual scenario=07_inventory_lab.json  # optional — replay one scenario by file name or id
```

---
