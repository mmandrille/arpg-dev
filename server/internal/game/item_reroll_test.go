package game

import (
	"encoding/json"
	"testing"
)

func TestRerollItemRollPayloadPreservesUniqueFixedEffects(t *testing.T) {
	rules := loadRules(t)
	templateID := "cave_blade"
	template := rules.ItemTemplates[templateID]
	payload := ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    "Embercall Blade",
		Rarity:         "unique",
		ItemLevel:      2,
		Stats:          cloneIntMap(template.BaseStats),
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      []string{"everburning_wound"},
	}
	if named, ok := rules.namedUniqueForEffectIDs(templateID, payload.EffectIDs); ok {
		for stat, value := range named.FixedStats {
			payload.Stats[stat] = value
		}
		payload.DisplayName = named.DisplayName
	}
	rng := NewRNG(SeedToUint64("reroll_unique_fixed"))
	first, err := RerollItemRollPayload(rules, payload, rng)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RerollItemRollPayload(rules, payload, NewRNG(SeedToUint64("reroll_unique_other")))
	if err != nil {
		t.Fatal(err)
	}
	if first.ItemTemplateID != templateID || second.ItemTemplateID != templateID {
		t.Fatalf("template changed: %+v %+v", first, second)
	}
	if first.Rarity != "unique" || second.Rarity != "unique" {
		t.Fatalf("rarity changed: %+v %+v", first, second)
	}
	if first.ItemLevel != 2 || second.ItemLevel != 2 {
		t.Fatalf("item level changed: %+v %+v", first, second)
	}
	if len(first.EffectIDs) != 1 || first.EffectIDs[0] != "everburning_wound" {
		t.Fatalf("effects = %+v", first.EffectIDs)
	}
}

func TestRenewRolledStatsJSONPreservesNestedItemLevel(t *testing.T) {
	rules := loadRules(t)
	raw := json.RawMessage(`{"item_template_id":"cave_shield","display_name":"Stalwart Rare Cave Shield","rarity":"rare","stats":{"armor":43,"block_percent":25,"dex":15,"item_level":15},"requirements":{"str":77,"level":15}}`)
	rng := NewRNG(SeedToUint64("renew_nested_ilvl"))
	renewed, err := RenewRolledStatsJSON(rules, raw, rng)
	if err != nil {
		t.Fatal(err)
	}
	var payload ItemRollPayload
	if err := json.Unmarshal(renewed, &payload); err != nil {
		t.Fatal(err)
	}
	var outMap map[string]any
	if err := json.Unmarshal(renewed, &outMap); err != nil {
		t.Fatal(err)
	}
	stats := intStatMapFromAny(rollPayloadStatsMap(outMap))
	if intStatValue(stats["item_level"]) != 15 {
		t.Fatalf("nested renewed item_level = %+v", outMap)
	}
	if stats["dex"] > 400 {
		t.Fatalf("dex %d looks inflated after renew", stats["dex"])
	}
}

func TestRenewRolledStatsJSONFlatItemLevelFifteenStaysBounded(t *testing.T) {
	rules := loadRules(t)
	raw := json.RawMessage(`{"item_template_id":"cave_shield","display_name":"Stalwart Rare Cave Shield","rarity":"rare","item_level":15,"armor":43,"block_percent":25,"dex":15}`)
	rng := NewRNG(SeedToUint64("renew_flat_ilvl"))
	renewed, err := RenewRolledStatsJSON(rules, raw, rng)
	if err != nil {
		t.Fatal(err)
	}
	var outMap map[string]any
	if err := json.Unmarshal(renewed, &outMap); err != nil {
		t.Fatal(err)
	}
	if intStatValue(rollPayloadStatsMap(outMap)["item_level"]) != 15 {
		t.Fatalf("flat renewed item_level = %+v", outMap)
	}
	stats := intStatMapFromAny(rollPayloadStatsMap(outMap))
	if stats["dex"] > 400 || stats["block_percent"] > 150 {
		t.Fatalf("renew at ilvl 15 produced inflated stats: %+v", stats)
	}
}

func TestRenewRolledStatsJSONChangesAffixStats(t *testing.T) {
	rules := loadRules(t)
	raw := json.RawMessage(`{"item_template_id":"cave_blade","display_name":"Sharp Cave Blade","rarity":"magic","stats":{"damage_min":2,"damage_max":5,"item_level":1}}`)
	rng := NewRNG(SeedToUint64("renew_affix_roll"))
	renewed, err := RenewRolledStatsJSON(rules, raw, rng)
	if err != nil {
		t.Fatal(err)
	}
	var payload ItemRollPayload
	if err := json.Unmarshal(renewed, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ItemTemplateID != "cave_blade" || payload.Rarity != "magic" {
		t.Fatalf("renewed payload = %+v", payload)
	}
	var outMap map[string]any
	if err := json.Unmarshal(renewed, &outMap); err != nil {
		t.Fatal(err)
	}
	if intStatValue(rollPayloadStatsMap(outMap)["item_level"]) != 1 {
		t.Fatalf("renewed item_level = %+v", outMap)
	}
	if payload.DisplayName == "Sharp Cave Blade" {
		t.Fatalf("expected display name to change after reroll: %q", payload.DisplayName)
	}
}
