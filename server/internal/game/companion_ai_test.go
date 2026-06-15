package game

import "testing"

func TestCompanionAppearsInSnapshotWithOwner(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_identity", "v182_companion_identity", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	companion := firstEntityByKind(sim, companionEntity)
	if companion == nil {
		t.Fatalf("missing companion in entities: %+v", sim.entities)
	}
	view := sim.entityView(companion)
	if view.Type != companionEntity || view.OwnerID != idStr(sim.playerID) || view.MonsterDefID != "combat_lab_crit_attacker" {
		t.Fatalf("companion view = %+v", view)
	}
	if view.HP == nil || view.MaxHP == nil || *view.HP <= 0 || *view.MaxHP <= 0 {
		t.Fatalf("companion hp view = %+v", view)
	}
}

func TestCompanionFollowsOwnerAndDamagesMonster(t *testing.T) {
	sim, err := NewSimWithWorld("sess_companion_ai", "v182_companion_ai", loadRules(t), "companion_ai_lab")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.entities[sim.playerID]
	companion := firstEntityByKind(sim, companionEntity)
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if player == nil || companion == nil || target == nil {
		t.Fatalf("setup player=%+v companion=%+v target=%+v", player, companion, target)
	}
	startCompanionPos := companion.pos
	startTargetHP := target.hp
	player.pos = Vec2{X: 6, Y: 5}

	var sawDamage bool
	for i := 0; i < 140; i++ {
		res := sim.Tick(nil)
		for _, ev := range res.Events {
			if ev.EventType == "monster_damaged" && ev.SourceEntityID == idStr(companion.id) && ev.TargetEntityID == idStr(target.id) {
				sawDamage = true
			}
		}
		if sawDamage {
			break
		}
	}
	if distance(companion.pos, startCompanionPos) < 0.5 {
		t.Fatalf("companion did not follow owner: start=%+v now=%+v", startCompanionPos, companion.pos)
	}
	if !sawDamage || target.hp >= startTargetHP {
		t.Fatalf("companion did not damage target: sawDamage=%v hp=%d start=%d companion=%+v target=%+v dist=%.3f target_id=%d", sawDamage, target.hp, startTargetHP, companion.pos, target.pos, distance(companion.pos, target.pos), companion.targetID)
	}
}
