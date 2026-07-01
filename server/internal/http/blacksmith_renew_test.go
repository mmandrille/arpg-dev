package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestRenewInventoryItemRoute(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	accountID, token := loginEmail(t, h, "renew-route+"+suffix+"@example.test")
	char := createCharacter(t, h, token, "Renew Route Hero")
	prog := store.CharacterProgression{AccountID: accountID, CharacterID: char.CharacterID, CharacterClass: "barbarian", Level: 1, Gold: 800, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := db.UpsertCharacterProgression(ctx, accountID, prog); err != nil {
		t.Fatal(err)
	}
	rolled := json.RawMessage(`{"item_template_id":"long_sword","display_name":"Long Sword","rarity":"common","stats":{"damage_min":1,"damage_max":2,"item_level":1}}`)
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "renew_blade_" + suffix, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: "long_sword", Location: store.ItemLocationInventory, RolledStats: rolled}); err != nil {
		t.Fatal(err)
	}
	stoneStats, err := game.MarshalRenewStoneRolledStats(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "renew_stone_" + suffix, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: game.RenewStoneItemDefID, Location: store.ItemLocationInventory, RolledStats: stoneStats}); err != nil {
		t.Fatal(err)
	}
	sellPrice := httpTestItemSellPrice(t, "long_sword", rolled)
	rec := postJSON(h, "/v0/account-stash/items/renew", token, map[string]string{
		"item_instance_id": "renew_blade_" + suffix,
		"character_id":     char.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("renew status = %d body=%s", rec.Code, rec.Body.String())
	}
	var renewed renewInventoryItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &renewed); err != nil {
		t.Fatal(err)
	}
	if renewed.RecipeID != "item_renew" || renewed.CostGold != sellPrice || !renewed.Success {
		t.Fatalf("renew response = %+v", renewed)
	}
	if renewed.ResourceItemDefID != game.RenewStoneItemDefID {
		t.Fatalf("resource = %s", renewed.ResourceItemDefID)
	}
}

func TestBlacksmithRecipeRouteRejectsHoneAndArmor(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	accountID, token := loginEmail(t, h, "renew-recipe+"+suffix+"@example.test")
	char := createCharacter(t, h, token, "Recipe Reject Hero")
	if err := db.UpsertCharacterProgression(ctx, accountID, store.CharacterProgression{AccountID: accountID, CharacterID: char.CharacterID, CharacterClass: "barbarian", Level: 1, Gold: 800, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "reject_blade_" + suffix, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: "long_sword", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountID, char.CharacterID, "reject_blade_"+suffix, "reject_blade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	for _, recipeID := range []string{"weapon_honing", "armor_reinforcement"} {
		rec := postJSON(h, "/v0/account-stash/items/reject_blade_stash_"+suffix+"/upgrade", token, map[string]string{"recipe_id": recipeID})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d body=%s", recipeID, rec.Code, rec.Body.String())
		}
	}
}
