package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// CreateSessionStartSnapshot freezes a member's replay inputs when the member
// joins the session: progression, items, bindings, account storage, recoverable
// corpses, and the mercenary roster. Rows are insert-only.
func (s *Store) CreateSessionStartSnapshot(ctx context.Context, snap SessionStartSnapshot) error {
	sessionID, accountID, characterID := snap.SessionID, snap.AccountID, snap.CharacterID
	items, waypoints, hotbar, skillBinds := snap.Items, snap.Waypoints, snap.Hotbar, snap.SkillBinds
	shopStock, stashItems, stashGold := snap.ShopStock, snap.StashItems, snap.StashGold
	resources, resourceBagItems := snap.Resources, snap.ResourceBagItems
	var progression CharacterProgression
	if snap.Progression != nil {
		progression = *snap.Progression
	}
	characterClass := progression.CharacterClass
	if characterClass == "" {
		characterClass = "barbarian"
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_character_progression (
			   session_id, account_id, character_id, character_class, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth, hired_mercenary_character_id
			 )
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			 ON CONFLICT (session_id, account_id, character_id) DO NOTHING`,
			sessionID, accountID, characterID, characterClass, progression.Level, progression.Experience, progression.UnspentStatPoints, progression.UnspentSkillPoints,
			progression.Stats.Str, progression.Stats.Dex, progression.Stats.Vit, progression.Stats.Magic, progression.Gold, progression.DeepestDungeonDepth, progression.HiredMercenaryCharacterID,
		); err != nil {
			return fmt.Errorf("store: insert session start progression: %w", err)
		}
		for skillID, rank := range progression.SkillRanks {
			if rank < 0 {
				return fmt.Errorf("store: insert session start skill rank: negative rank for %s", skillID)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_character_skill_ranks (session_id, account_id, character_id, skill_id, rank)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, account_id, character_id, skill_id) DO NOTHING`,
				sessionID, accountID, characterID, skillID, rank,
			); err != nil {
				return fmt.Errorf("store: insert session start skill rank: %w", err)
			}
		}
		for _, item := range items {
			var slot any
			if item.Slot != "" {
				slot = item.Slot
			}
			location := item.Location
			if location == "" {
				location = ItemLocationInventory
			}
			rolledStats := item.RolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_item_instances (session_id, id, account_id, character_id, item_def_id, location, slot, equipped, weapon_set, rolled_stats)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
				 ON CONFLICT (session_id, account_id, character_id, id) DO NOTHING`,
				sessionID, item.ID, accountID, characterID, item.ItemDefID, location, slot, item.Equipped, normalizeWeaponSet(item.WeaponSet), []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: insert session start item: %w", err)
			}
		}
		for _, wp := range waypoints {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_waypoints (session_id, character_id, level)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (session_id, character_id, level) DO NOTHING`,
				sessionID, characterID, wp.Level,
			); err != nil {
				return fmt.Errorf("store: insert session start waypoint: %w", err)
			}
		}
		for _, slot := range hotbar {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_hotbar_slots (session_id, account_id, character_id, slot_index, item_instance_id)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, account_id, character_id, slot_index) DO NOTHING`,
				sessionID, accountID, characterID, slot.SlotIndex, slot.ItemInstanceID,
			); err != nil {
				return fmt.Errorf("store: insert session start hotbar: %w", err)
			}
		}
		keys := normalizeSkillFunctionKeys(skillBinds.FunctionKeys)
		for slot, skillID := range keys {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_skill_bindings (session_id, account_id, character_id, slot_index, skill_id)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, account_id, character_id, slot_index) DO NOTHING`,
				sessionID, accountID, characterID, slot, skillID,
			); err != nil {
				return fmt.Errorf("store: insert session start skill binding: %w", err)
			}
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_skill_preferences (session_id, account_id, character_id, right_click_skill_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (session_id, account_id, character_id) DO NOTHING`,
			sessionID, accountID, characterID, skillBinds.RightClickSkillID,
		); err != nil {
			return fmt.Errorf("store: insert session start skill preference: %w", err)
		}
		for _, stock := range shopStock {
			rolledPayload := stock.RolledPayload
			if len(rolledPayload) == 0 {
				rolledPayload = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_shop_stock (
				   session_id, account_id, character_id, shop_id, refresh_key, offer_index, offer_id, source_depth, item_template_id, rolled_payload, buy_price, available
				 )
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12)
				 ON CONFLICT (session_id, account_id, character_id, shop_id, offer_id) DO NOTHING`,
				sessionID, accountID, characterID, stock.ShopID, stock.RefreshKey, stock.OfferIndex, stock.OfferID,
				stock.SourceDepth, stock.ItemTemplateID, []byte(rolledPayload), stock.BuyPrice, stock.Available,
			); err != nil {
				return fmt.Errorf("store: insert session start shop stock: %w", err)
			}
		}
		for _, stashItem := range stashItems {
			rolledStats := stashItem.RolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			var sourceCharacterID any
			if stashItem.SourceCharacterID != "" {
				sourceCharacterID = stashItem.SourceCharacterID
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_account_stash_items (
				   session_id, account_id, stash_item_id, source_character_id, item_def_id, rolled_stats
				 )
				 VALUES ($1, $2, $3, $4, $5, $6::jsonb)
				 ON CONFLICT (session_id, account_id, stash_item_id) DO NOTHING`,
				sessionID, accountID, stashItem.StashItemID, sourceCharacterID, stashItem.ItemDefID, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: insert session start account stash item: %w", err)
			}
		}
		stashGoldAccountID := stashGold.AccountID
		if stashGoldAccountID == "" {
			stashGoldAccountID = accountID
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_account_stash_gold (session_id, account_id, gold)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (session_id, account_id) DO NOTHING`,
			sessionID, stashGoldAccountID, stashGold.Gold,
		); err != nil {
			return fmt.Errorf("store: insert session start account stash gold: %w", err)
		}
		for _, resource := range resources {
			if resource.ResourceID == "" || resource.Amount < 0 {
				return fmt.Errorf("store: insert session start account resource: invalid resource %q amount %d", resource.ResourceID, resource.Amount)
			}
			resourceAccountID := resource.AccountID
			if resourceAccountID == "" {
				resourceAccountID = accountID
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_account_resource_wallet (session_id, account_id, resource_id, amount)
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (session_id, account_id, resource_id) DO NOTHING`,
				sessionID, resourceAccountID, resource.ResourceID, resource.Amount,
			); err != nil {
				return fmt.Errorf("store: insert session start account resource: %w", err)
			}
		}
		for _, bagItem := range resourceBagItems {
			rolledStats := bagItem.RolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			var sourceCharacterID any
			if bagItem.SourceCharacterID != "" {
				sourceCharacterID = bagItem.SourceCharacterID
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_account_resource_bag_items (
				   session_id, account_id, bag_item_id, source_character_id, item_def_id, rolled_stats
				 )
				 VALUES ($1, $2, $3, $4, $5, $6::jsonb)
				 ON CONFLICT (session_id, account_id, bag_item_id) DO NOTHING`,
				sessionID, accountID, bagItem.BagItemID, sourceCharacterID, bagItem.ItemDefID, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: insert session start account resource bag item: %w", err)
			}
		}
		if err := insertSessionStartCorpses(ctx, tx, sessionID, accountID, characterID, snap.Corpses); err != nil {
			return err
		}
		return insertSessionStartMercenaryRoster(ctx, tx, sessionID, accountID, characterID, snap.MercenaryRoster)
	})
}

// sessionStartCorpseItem is the frozen JSON shape of one corpse item.
type sessionStartCorpseItem struct {
	ID          string          `json:"id"`
	ItemDefID   string          `json:"item_def_id"`
	Location    string          `json:"location"`
	Slot        string          `json:"slot,omitempty"`
	Equipped    bool            `json:"equipped,omitempty"`
	WeaponSet   int             `json:"weapon_set,omitempty"`
	RolledStats json.RawMessage `json:"rolled_stats,omitempty"`
}

type sessionStartMercenaryStats struct {
	Str   int `json:"str"`
	Dex   int `json:"dex"`
	Vit   int `json:"vit"`
	Magic int `json:"magic"`
}

type sessionStartMercenaryProgression struct {
	CharacterClass            string                     `json:"character_class"`
	Level                     int                        `json:"level"`
	Experience                int                        `json:"experience"`
	UnspentStatPoints         int                        `json:"unspent_stat_points"`
	UnspentSkillPoints        int                        `json:"unspent_skill_points"`
	SkillRanks                map[string]int             `json:"skill_ranks,omitempty"`
	Stats                     sessionStartMercenaryStats `json:"stats"`
	Gold                      int                        `json:"gold"`
	DeepestDungeonDepth       int                        `json:"deepest_dungeon_depth"`
	HiredMercenaryCharacterID string                     `json:"hired_mercenary_character_id,omitempty"`
}

func encodeSessionStartMercenaryProgression(progression CharacterProgression) ([]byte, error) {
	return json.Marshal(sessionStartMercenaryProgression{
		CharacterClass:            progression.CharacterClass,
		Level:                     progression.Level,
		Experience:                progression.Experience,
		UnspentStatPoints:         progression.UnspentStatPoints,
		UnspentSkillPoints:        progression.UnspentSkillPoints,
		SkillRanks:                progression.SkillRanks,
		Stats:                     sessionStartMercenaryStats{Str: progression.Stats.Str, Dex: progression.Stats.Dex, Vit: progression.Stats.Vit, Magic: progression.Stats.Magic},
		Gold:                      progression.Gold,
		DeepestDungeonDepth:       progression.DeepestDungeonDepth,
		HiredMercenaryCharacterID: progression.HiredMercenaryCharacterID,
	})
}

func decodeSessionStartMercenaryProgression(data []byte, accountID, characterID string) (CharacterProgression, error) {
	var frozen sessionStartMercenaryProgression
	if err := json.Unmarshal(data, &frozen); err != nil {
		return CharacterProgression{}, err
	}
	return CharacterProgression{
		AccountID:                 accountID,
		CharacterID:               characterID,
		CharacterClass:            frozen.CharacterClass,
		Level:                     frozen.Level,
		Experience:                frozen.Experience,
		UnspentStatPoints:         frozen.UnspentStatPoints,
		UnspentSkillPoints:        frozen.UnspentSkillPoints,
		SkillRanks:                frozen.SkillRanks,
		Stats:                     CharacterBaseStats{Str: frozen.Stats.Str, Dex: frozen.Stats.Dex, Vit: frozen.Stats.Vit, Magic: frozen.Stats.Magic},
		Gold:                      frozen.Gold,
		DeepestDungeonDepth:       frozen.DeepestDungeonDepth,
		HiredMercenaryCharacterID: frozen.HiredMercenaryCharacterID,
	}, nil
}

// insertSessionStartCorpses stores corpses in list order; ordinal preserves it
// because the sim spawns already-generated-level corpses in load order.
func insertSessionStartCorpses(ctx context.Context, tx pgx.Tx, sessionID, accountID, characterID string, corpses []CharacterCorpse) error {
	for ordinal, corpse := range corpses {
		items := make([]sessionStartCorpseItem, 0, len(corpse.Items))
		for _, item := range corpse.Items {
			items = append(items, sessionStartCorpseItem{
				ID:          item.ID,
				ItemDefID:   item.ItemDefID,
				Location:    item.Location,
				Slot:        item.Slot,
				Equipped:    item.Equipped,
				WeaponSet:   item.WeaponSet,
				RolledStats: item.RolledStats,
			})
		}
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("store: encode session start corpse items: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_character_corpses (
			   session_id, account_id, character_id, corpse_character_id, ordinal, name, level, death_level, items
			 )
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
			 ON CONFLICT (session_id, account_id, character_id, corpse_character_id) DO NOTHING`,
			sessionID, accountID, characterID, corpse.CharacterID, ordinal, corpse.Name, corpse.Level, corpse.DeathLevel, itemsJSON,
		); err != nil {
			return fmt.Errorf("store: insert session start corpse: %w", err)
		}
	}
	return nil
}

func insertSessionStartMercenaryRoster(ctx context.Context, tx pgx.Tx, sessionID, accountID, characterID string, roster []MercenaryCharacterSnapshot) error {
	for _, mercenary := range roster {
		if mercenary.CharacterID == "" || mercenary.CharacterID == characterID {
			return fmt.Errorf("store: insert session start mercenary: invalid character id")
		}
		progressionJSON, err := encodeSessionStartMercenaryProgression(mercenary.Progression)
		if err != nil {
			return fmt.Errorf("store: encode session start mercenary progression: %w", err)
		}
		items := make([]sessionStartCorpseItem, 0, len(mercenary.Items))
		for _, item := range mercenary.Items {
			if item.CharacterID != "" && item.CharacterID != mercenary.CharacterID {
				return fmt.Errorf("store: insert session start mercenary: item character does not match roster character")
			}
			items = append(items, sessionStartCorpseItem{
				ID:          item.ID,
				ItemDefID:   item.ItemDefID,
				Location:    item.Location,
				Slot:        item.Slot,
				Equipped:    item.Equipped,
				WeaponSet:   item.WeaponSet,
				RolledStats: item.RolledStats,
			})
		}
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("store: encode session start mercenary items: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_mercenary_roster (
			   session_id, account_id, character_id, mercenary_character_id, name, character_class, dead, progression, items
			 )
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb)
			 ON CONFLICT (session_id, account_id, character_id, mercenary_character_id) DO NOTHING`,
			sessionID, accountID, characterID, mercenary.CharacterID, mercenary.Name, mercenary.CharacterClass, mercenary.Dead, progressionJSON, itemsJSON,
		); err != nil {
			return fmt.Errorf("store: insert session start mercenary: %w", err)
		}
	}
	return nil
}

func (s *Store) loadSessionStartCorpses(ctx context.Context, sessionID, accountID, characterID string) ([]CharacterCorpse, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT corpse_character_id, name, level, death_level, items
		 FROM session_start_character_corpses
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY ordinal ASC`,
		sessionID, accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: load session start corpses: %w", err)
	}
	defer rows.Close()
	var corpses []CharacterCorpse
	for rows.Next() {
		var corpse CharacterCorpse
		var itemsJSON []byte
		if err := rows.Scan(&corpse.CharacterID, &corpse.Name, &corpse.Level, &corpse.DeathLevel, &itemsJSON); err != nil {
			return nil, fmt.Errorf("store: scan session start corpse: %w", err)
		}
		var items []sessionStartCorpseItem
		if err := json.Unmarshal(itemsJSON, &items); err != nil {
			return nil, fmt.Errorf("store: decode session start corpse items: %w", err)
		}
		for _, item := range items {
			corpse.Items = append(corpse.Items, CharacterItemInstance{
				ID:          item.ID,
				AccountID:   accountID,
				CharacterID: corpse.CharacterID,
				ItemDefID:   item.ItemDefID,
				Location:    item.Location,
				Slot:        item.Slot,
				Equipped:    item.Equipped,
				WeaponSet:   item.WeaponSet,
				RolledStats: item.RolledStats,
			})
		}
		corpses = append(corpses, corpse)
	}
	return corpses, rows.Err()
}

func (s *Store) loadSessionStartMercenaryRoster(ctx context.Context, sessionID, accountID, characterID string) ([]MercenaryCharacterSnapshot, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT mercenary_character_id, name, character_class, dead, progression, items
		 FROM session_start_mercenary_roster
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY mercenary_character_id ASC`,
		sessionID, accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: load session start mercenary roster: %w", err)
	}
	defer rows.Close()
	var roster []MercenaryCharacterSnapshot
	for rows.Next() {
		var mercenary MercenaryCharacterSnapshot
		var progressionJSON, itemsJSON []byte
		if err := rows.Scan(&mercenary.CharacterID, &mercenary.Name, &mercenary.CharacterClass, &mercenary.Dead, &progressionJSON, &itemsJSON); err != nil {
			return nil, fmt.Errorf("store: scan session start mercenary: %w", err)
		}
		mercenary.Progression, err = decodeSessionStartMercenaryProgression(progressionJSON, accountID, mercenary.CharacterID)
		if err != nil {
			return nil, fmt.Errorf("store: decode session start mercenary progression: %w", err)
		}
		var items []sessionStartCorpseItem
		if err := json.Unmarshal(itemsJSON, &items); err != nil {
			return nil, fmt.Errorf("store: decode session start mercenary items: %w", err)
		}
		for _, item := range items {
			mercenary.Items = append(mercenary.Items, CharacterItemInstance{
				ID:          item.ID,
				AccountID:   accountID,
				CharacterID: mercenary.CharacterID,
				ItemDefID:   item.ItemDefID,
				Location:    item.Location,
				Slot:        item.Slot,
				Equipped:    item.Equipped,
				WeaponSet:   item.WeaponSet,
				RolledStats: item.RolledStats,
			})
		}
		roster = append(roster, mercenary)
	}
	return roster, rows.Err()
}
