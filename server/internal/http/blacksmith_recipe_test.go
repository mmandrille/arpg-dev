package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestBlacksmithRecipeRouteHonorsRecipeID(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	accountID, token := loginEmail(t, h, "stash-recipe+"+suffix+"@example.test")
	char := createCharacter(t, h, token, "Recipe Route Hero")
	prog := store.CharacterProgression{AccountID: accountID, CharacterID: char.CharacterID, CharacterClass: "barbarian", Level: 1, Gold: 300, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := db.UpsertCharacterProgression(ctx, accountID, prog); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "recipe_blade_" + suffix, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountID, char.CharacterID, "recipe_blade_"+suffix, "recipe_blade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "recipe_shield_" + suffix, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: "cave_shield", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"armor":2,"block_percent":5}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountID, char.CharacterID, "recipe_shield_"+suffix, "recipe_shield_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.TransferCharacterGoldToAccountStash(ctx, accountID, char.CharacterID, 250); err != nil {
		t.Fatal(err)
	}
	rec := postJSON(h, "/v0/account-stash/items/recipe_blade_stash_"+suffix+"/upgrade", token, map[string]string{"recipe_id": "weapon_honing"})
	if rec.Code != http.StatusOK {
		t.Fatalf("weapon recipe status = %d body=%s", rec.Code, rec.Body.String())
	}
	var upgraded upgradeAccountStashItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &upgraded); err != nil {
		t.Fatal(err)
	}
	if upgraded.RecipeID != "weapon_honing" || upgraded.CostGold != 100 || !upgraded.Success {
		t.Fatalf("weapon recipe response = %+v", upgraded)
	}
	var stats struct {
		ItemLevel int `json:"item_level"`
		DamageMax int `json:"damage_max"`
	}
	if err := json.Unmarshal(upgraded.Item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats.ItemLevel != 1 || stats.DamageMax != 5 {
		t.Fatalf("weapon recipe stats = %+v", stats)
	}
	rec = postJSON(h, "/v0/account-stash/items/recipe_shield_stash_"+suffix+"/upgrade", token, map[string]string{"recipe_id": "weapon_honing"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("shield weapon recipe status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/account-stash/items/recipe_blade_stash_"+suffix+"/upgrade", token, map[string]string{"recipe_id": "armor_reinforcement"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("blade armor recipe status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/account-stash/items/recipe_shield_stash_"+suffix+"/upgrade", token, map[string]string{"recipe_id": "armor_reinforcement"})
	if rec.Code != http.StatusOK {
		t.Fatalf("armor recipe status = %d body=%s", rec.Code, rec.Body.String())
	}
	var reinforced upgradeAccountStashItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &reinforced); err != nil {
		t.Fatal(err)
	}
	if reinforced.RecipeID != "armor_reinforcement" || reinforced.CostGold != 100 || !reinforced.Success {
		t.Fatalf("armor recipe response = %+v", reinforced)
	}
	var armorStats struct {
		ItemLevel int `json:"item_level"`
		Armor     int `json:"armor"`
	}
	if err := json.Unmarshal(reinforced.Item.RolledStats, &armorStats); err != nil {
		t.Fatal(err)
	}
	if armorStats.ItemLevel != 1 || armorStats.Armor != 3 {
		t.Fatalf("armor recipe stats = %+v", armorStats)
	}
	rec = postJSON(h, "/v0/account-stash/items/recipe_shield_stash_"+suffix+"/upgrade", token, map[string]string{"recipe_id": "unknown_recipe"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown recipe status = %d body=%s", rec.Code, rec.Body.String())
	}
}
