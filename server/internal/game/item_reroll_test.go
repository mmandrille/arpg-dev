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
	if payload.ItemTemplateID != "cave_blade" || payload.Rarity != "magic" || payload.ItemLevel != 1 {
		t.Fatalf("renewed payload = %+v", payload)
	}
	if payload.DisplayName == "Sharp Cave Blade" {
		t.Fatalf("expected display name to change after reroll: %q", payload.DisplayName)
	}
}
