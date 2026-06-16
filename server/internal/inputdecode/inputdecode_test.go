package inputdecode

import (
	"encoding/json"
	"testing"
)

func TestDecodeTeleportAllowsTownLevel(t *testing.T) {
	in, ok := Decode(TypeTeleport, "msg_tp_town", "", json.RawMessage(`{"target_level":0}`))
	if !ok {
		t.Fatal("Decode teleport target_level 0 rejected")
	}
	if in.Teleport == nil || in.Teleport.TargetLevel != 0 {
		t.Fatalf("decoded teleport = %+v, want target level 0", in.Teleport)
	}
}

func TestDecodeTeleportRejectsMissingOrPositiveTargetLevel(t *testing.T) {
	tests := []struct {
		name    string
		payload json.RawMessage
	}{
		{name: "missing", payload: json.RawMessage(`{}`)},
		{name: "positive", payload: json.RawMessage(`{"target_level":1}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := Decode(TypeTeleport, "msg_tp", "", tt.payload); ok {
				t.Fatalf("Decode(%s) accepted invalid teleport payload", tt.payload)
			}
		})
	}
}

func TestDecodeAllocateStatIntent(t *testing.T) {
	in, ok := Decode(TypeAllocateStat, "msg_stat", "", json.RawMessage(`{"stat":"vit","points":1}`))
	if !ok {
		t.Fatal("Decode allocate_stat_intent rejected valid payload")
	}
	if in.AllocateStat == nil || in.AllocateStat.Stat != "vit" || in.AllocateStat.Points != 1 {
		t.Fatalf("decoded allocate stat = %+v", in.AllocateStat)
	}
}

func TestDecodeAllocateStatIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{"stat":"luck","points":1}`),
		json.RawMessage(`{"stat":"str","points":0}`),
		json.RawMessage(`{"stat":"str","points":-1}`),
		json.RawMessage(`{"points":1}`),
		json.RawMessage(`{"stat":"str"}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeAllocateStat, "msg_stat", "", payload); ok {
			t.Fatalf("Decode accepted invalid allocate stat payload %s", payload)
		}
	}
}

func TestDecodeAllocateSkillPointIntent(t *testing.T) {
	in, ok := Decode(TypeAllocateSkillPoint, "msg_skill", "", json.RawMessage(`{"skill_id":"magic_bolt"}`))
	if !ok {
		t.Fatal("Decode allocate_skill_point_intent rejected valid payload")
	}
	if in.AllocateSkillPoint == nil || in.AllocateSkillPoint.SkillID != "magic_bolt" {
		t.Fatalf("decoded allocate skill point = %+v", in.AllocateSkillPoint)
	}
	if !IsClientIntent(TypeAllocateSkillPoint) {
		t.Fatal("allocate_skill_point_intent not marked as client intent")
	}
}

func TestDecodeCastSkillIntent(t *testing.T) {
	targeted, ok := Decode(TypeCastSkill, "msg_cast_target", "", json.RawMessage(`{"skill_id":"magic_bolt","target_id":"1002"}`))
	if !ok {
		t.Fatal("Decode cast_skill_intent rejected target payload")
	}
	if targeted.CastSkill == nil || targeted.CastSkill.SkillID != "magic_bolt" || targeted.CastSkill.TargetID != "1002" || targeted.CastSkill.Direction != nil {
		t.Fatalf("decoded targeted cast skill = %+v", targeted.CastSkill)
	}

	directional, ok := Decode(TypeCastSkill, "msg_cast_dir", "", json.RawMessage(`{"skill_id":"magic_bolt","direction":{"x":1,"y":0}}`))
	if !ok {
		t.Fatal("Decode cast_skill_intent rejected direction payload")
	}
	if directional.CastSkill == nil || directional.CastSkill.Direction == nil || directional.CastSkill.Direction.X != 1 || directional.CastSkill.Direction.Y != 0 {
		t.Fatalf("decoded directional cast skill = %+v", directional.CastSkill)
	}
	if !IsClientIntent(TypeCastSkill) {
		t.Fatal("cast_skill_intent not marked as client intent")
	}
}

func TestDecodeCompanionCommandIntent(t *testing.T) {
	in, ok := Decode(TypeCompanionCommand, "msg_companion_stance", "corr_companion_stance", json.RawMessage(`{"stance":"passive"}`))
	if !ok {
		t.Fatal("Decode companion_command_intent rejected valid payload")
	}
	if in.CompanionCommand == nil || in.CompanionCommand.Stance != "passive" {
		t.Fatalf("decoded companion command = %+v", in.CompanionCommand)
	}
	if !IsClientIntent(TypeCompanionCommand) {
		t.Fatal("companion_command_intent not marked as client intent")
	}
}

func TestDecodeCompanionCommandIntentRejectsInvalidPayload(t *testing.T) {
	for _, payload := range []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"stance":""}`),
		json.RawMessage(`{"stance":"hold"}`),
		json.RawMessage(`{"stance":1}`),
	} {
		if _, ok := Decode(TypeCompanionCommand, "msg_companion_stance", "", payload); ok {
			t.Fatalf("Decode accepted invalid companion command payload %s", payload)
		}
	}
}

func TestDecodeSkillIntentsRejectInvalidPayload(t *testing.T) {
	for _, payload := range []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"skill_id":""}`),
	} {
		if _, ok := Decode(TypeAllocateSkillPoint, "msg_skill", "", payload); ok {
			t.Fatalf("Decode accepted invalid allocate skill payload %s", payload)
		}
	}
	for _, payload := range []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"skill_id":"magic_bolt"}`),
		json.RawMessage(`{"skill_id":"","target_id":"1002"}`),
		json.RawMessage(`{"skill_id":"magic_bolt","direction":null}`),
	} {
		if _, ok := Decode(TypeCastSkill, "msg_cast", "", payload); ok {
			t.Fatalf("Decode accepted invalid cast skill payload %s", payload)
		}
	}
}

func TestDecodeDirectionalAttackIntent(t *testing.T) {
	in, ok := Decode(TypeDirectional, "msg_dir", "corr_dir", json.RawMessage(`{"direction":{"x":1,"y":0}}`))
	if !ok {
		t.Fatal("Decode directional_attack_intent rejected valid payload")
	}
	if in.DirectionalAttack == nil || in.DirectionalAttack.Direction.X != 1 || in.DirectionalAttack.Direction.Y != 0 {
		t.Fatalf("decoded directional attack = %+v", in.DirectionalAttack)
	}
	if !IsClientIntent(TypeDirectional) {
		t.Fatal("directional_attack_intent not marked as client intent")
	}
}

func TestDecodeDirectionalAttackIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"direction":null}`),
		json.RawMessage(`{"direction":{"x":"bad","y":0}}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeDirectional, "msg_dir", "", payload); ok {
			t.Fatalf("Decode accepted invalid directional attack payload %s", payload)
		}
	}
}

func TestDecodeShopBuyIntent(t *testing.T) {
	in, ok := Decode(TypeShopBuy, "msg_buy", "corr_buy", json.RawMessage(`{"shop_entity_id":"1013","offer_id":"fixed:red_potion"}`))
	if !ok {
		t.Fatal("Decode shop_buy_intent rejected valid payload")
	}
	if in.ShopBuy == nil || in.ShopBuy.ShopEntityID != "1013" || in.ShopBuy.OfferID != "fixed:red_potion" {
		t.Fatalf("decoded shop buy = %+v", in.ShopBuy)
	}
	if !IsClientIntent(TypeShopBuy) {
		t.Fatal("shop_buy_intent not marked as client intent")
	}
}

func TestDecodeShopBuyIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"shop_entity_id":"1013"}`),
		json.RawMessage(`{"offer_id":"fixed:red_potion"}`),
		json.RawMessage(`{"shop_entity_id":1013,"offer_id":"fixed:red_potion"}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeShopBuy, "msg_buy", "", payload); ok {
			t.Fatalf("Decode accepted invalid shop buy payload %s", payload)
		}
	}
}

func TestDecodeShopSellIntent(t *testing.T) {
	in, ok := Decode(TypeShopSell, "msg_sell", "corr_sell", json.RawMessage(`{"shop_entity_id":"1013","item_instance_id":"1004"}`))
	if !ok {
		t.Fatal("Decode shop_sell_intent rejected valid payload")
	}
	if in.ShopSell == nil || in.ShopSell.ShopEntityID != "1013" || in.ShopSell.ItemInstanceID != "1004" {
		t.Fatalf("decoded shop sell = %+v", in.ShopSell)
	}
	if !IsClientIntent(TypeShopSell) {
		t.Fatal("shop_sell_intent not marked as client intent")
	}
}

func TestDecodeShopSellIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"shop_entity_id":"1013"}`),
		json.RawMessage(`{"item_instance_id":"1004"}`),
		json.RawMessage(`{"shop_entity_id":"1013","item_instance_id":1004}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeShopSell, "msg_sell", "", payload); ok {
			t.Fatalf("Decode accepted invalid shop sell payload %s", payload)
		}
	}
}

func TestDecodeShopRerollIntent(t *testing.T) {
	in, ok := Decode(TypeShopReroll, "msg_reroll", "corr_reroll", json.RawMessage(`{"shop_entity_id":"1013"}`))
	if !ok {
		t.Fatal("Decode shop_reroll_intent rejected valid payload")
	}
	if in.ShopReroll == nil || in.ShopReroll.ShopEntityID != "1013" {
		t.Fatalf("decoded shop reroll = %+v", in.ShopReroll)
	}
	if !IsClientIntent(TypeShopReroll) {
		t.Fatal("shop_reroll_intent not marked as client intent")
	}
}

func TestDecodeShopRerollIntentRejectsInvalidPayload(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"shop_entity_id":1013}`),
	}
	for _, payload := range tests {
		if _, ok := Decode(TypeShopReroll, "msg_reroll", "", payload); ok {
			t.Fatalf("Decode accepted invalid shop reroll payload %s", payload)
		}
	}
}

func TestDecodeBishopDebugIntents(t *testing.T) {
	tests := []struct {
		typ  string
		want string
	}{
		{typ: TypeBishopDebugLevel, want: "level"},
		{typ: TypeBishopDebugSkill, want: "skill"},
		{typ: TypeBishopDebugStat, want: "stat"},
	}
	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			in, ok := Decode(tt.typ, "msg_bishop_debug", "corr_bishop_debug", json.RawMessage(`{"bishop_entity_id":"1013"}`))
			if !ok {
				t.Fatalf("Decode %s rejected valid payload", tt.typ)
			}
			if !IsClientIntent(tt.typ) {
				t.Fatalf("%s not marked as client intent", tt.typ)
			}
			switch tt.want {
			case "level":
				if in.BishopDebugLevel == nil || in.BishopDebugLevel.BishopEntityID != "1013" {
					t.Fatalf("decoded bishop debug level = %+v", in.BishopDebugLevel)
				}
			case "skill":
				if in.BishopDebugSkill == nil || in.BishopDebugSkill.BishopEntityID != "1013" {
					t.Fatalf("decoded bishop debug skill = %+v", in.BishopDebugSkill)
				}
			case "stat":
				if in.BishopDebugStat == nil || in.BishopDebugStat.BishopEntityID != "1013" {
					t.Fatalf("decoded bishop debug stat = %+v", in.BishopDebugStat)
				}
			}
		})
	}
}

func TestDecodeBishopReviveAllIntent(t *testing.T) {
	in, ok := Decode(TypeBishopReviveAll, "msg_bishop_revive_all", "corr_bishop_revive_all", json.RawMessage(`{"bishop_entity_id":"1013"}`))
	if !ok {
		t.Fatal("Decode rejected valid bishop revive all payload")
	}
	if !IsClientIntent(TypeBishopReviveAll) {
		t.Fatal("bishop_revive_all_intent not marked as client intent")
	}
	if in.BishopReviveAll == nil || in.BishopReviveAll.BishopEntityID != "1013" {
		t.Fatalf("decoded bishop revive all = %+v", in.BishopReviveAll)
	}
}

func TestDecodeBishopDebugIntentsRejectInvalidPayload(t *testing.T) {
	for _, typ := range []string{TypeBishopDebugLevel, TypeBishopDebugSkill, TypeBishopDebugStat} {
		for _, payload := range []json.RawMessage{
			json.RawMessage(`{}`),
			json.RawMessage(`{"bishop_entity_id":""}`),
			json.RawMessage(`{"bishop_entity_id":1013}`),
		} {
			if _, ok := Decode(typ, "msg_bishop_debug", "", payload); ok {
				t.Fatalf("Decode %s accepted invalid payload %s", typ, payload)
			}
		}
	}
}

func TestDecodeStoredShopIntent(t *testing.T) {
	raw := []byte(`{
		"type":"shop_buy_intent",
		"message_id":"msg_buy_stored",
		"correlation_id":"corr_buy_stored",
		"payload":{"shop_entity_id":"1013","offer_id":"generated:depth3:000"}
	}`)
	in, ok := DecodeStored(raw)
	if !ok {
		t.Fatal("DecodeStored rejected shop_buy_intent")
	}
	if in.ShopBuy == nil || in.ShopBuy.ShopEntityID != "1013" || in.ShopBuy.OfferID != "generated:depth3:000" {
		t.Fatalf("decoded stored shop buy = %+v", in.ShopBuy)
	}
}
