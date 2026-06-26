package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonsterDiveAttackStyleValidation(t *testing.T) {
	expectMonsterRulesError(t, func(defs map[string]any) {
		bat := defs["dungeon_bat"].(map[string]any)
		bat["behavior"] = monsterBehaviorStatic
	}, "attack_style")
}

func TestMonsterAttackRangeValidationKeepsOrdinaryMeleeShort(t *testing.T) {
	expectMonsterRulesError(t, func(defs map[string]any) {
		mob := defs["dungeon_mob"].(map[string]any)
		mob["attack_range"] = 2.35
	}, "attack_range")
}

func TestBatDiveAttackStyleIsEmittedForDirectPlayerDamage(t *testing.T) {
	rules := loadRules(t)
	batDef := rules.Monsters["dungeon_bat"]
	batDef.HitChance = floatPtr(1)
	batDef.AttackDamage = &DamageRange{Min: 1, Max: 1}
	batDef.AttackCooldown = 1
	rules.Monsters["dungeon_bat"] = batDef
	mobDef := rules.Monsters["dungeon_mob"]
	mobDef.HitChance = floatPtr(1)
	mobDef.AttackDamage = &DamageRange{Min: 1, Max: 1}
	mobDef.AttackCooldown = 1
	mobDef.AttackWindupTicks = 0
	rules.Monsters["dungeon_mob"] = mobDef

	sim, err := NewSimWithWorld("sess_bat_dive_attack_style", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range sim.activeLevel().entities {
		if candidate.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 5, Y: 5}
	player.hp = playerStartHP

	bat := addTestMonster(sim, "dungeon_bat", Vec2{X: 6, Y: 5}, rules.Monsters["dungeon_bat"].MaxHP)
	bat.aiMode = monsterAIModeChase
	batResult := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&batResult)
	batEvent := firstEventBySource(batResult, "player_damaged", bat.id)
	if batEvent == nil {
		t.Fatalf("bat events = %+v, want player_damaged from bat", batResult.Events)
	}
	if batEvent.AttackStyle != monsterAttackStyleDive {
		t.Fatalf("bat attack style = %q, want %q in %+v", batEvent.AttackStyle, monsterAttackStyleDive, batResult.Events)
	}

	delete(sim.activeLevel().entities, bat.id)
	player.hp = playerStartHP
	mob := addTestMonster(sim, "dungeon_mob", Vec2{X: 6, Y: 5}, rules.Monsters["dungeon_mob"].MaxHP)
	mob.aiMode = monsterAIModeChase
	mobResult := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&mobResult)
	mobEvent := firstEventBySource(mobResult, "player_damaged", mob.id)
	if mobEvent == nil {
		t.Fatalf("mob events = %+v, want player_damaged from dungeon_mob", mobResult.Events)
	}
	if mobEvent.AttackStyle != "" {
		t.Fatalf("normal mob attack style = %q, want omitted", mobEvent.AttackStyle)
	}
}

func TestWolfPounceAttackStyleExtendsMeleeReach(t *testing.T) {
	rules := loadRules(t)
	wolfDef := rules.Monsters["dungeon_wolf"]
	wolfDef.HitChance = floatPtr(1)
	wolfDef.AttackDamage = &DamageRange{Min: 1, Max: 1}
	wolfDef.AttackCooldown = 1
	rules.Monsters["dungeon_wolf"] = wolfDef

	sim, err := NewSimWithWorld("sess_wolf_pounce_attack_style", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range sim.activeLevel().entities {
		if candidate.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 5, Y: 5}
	player.hp = playerStartHP
	wolfDistance := rules.Combat.UnarmedReach + playerRadius + 0.35
	if wolfDistance > rules.Monsters["dungeon_wolf"].AttackRange+playerRadius {
		t.Fatalf("test wolf distance %.2f exceeds pounce range %.2f", wolfDistance, rules.Monsters["dungeon_wolf"].AttackRange)
	}

	wolf := addTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + wolfDistance, Y: player.pos.Y}, rules.Monsters["dungeon_wolf"].MaxHP)
	wolf.aiMode = monsterAIModeChase
	result := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&result)
	event := firstEventBySource(result, "player_damaged", wolf.id)
	if event == nil {
		t.Fatalf("wolf events = %+v, want player_damaged from pounce range", result.Events)
	}
	if event.AttackStyle != monsterAttackStylePounce {
		t.Fatalf("wolf attack style = %q, want %q in %+v", event.AttackStyle, monsterAttackStylePounce, result.Events)
	}
}

func firstEventBySource(r TickResult, eventType string, sourceID uint64) *Event {
	for idx := range r.Events {
		event := &r.Events[idx]
		if event.EventType == eventType && event.SourceEntityID == idStr(sourceID) {
			return event
		}
	}
	return nil
}

func expectMonsterRulesError(t *testing.T, mutate func(defs map[string]any), want string) {
	t.Helper()
	sourceRulesDir, err := FindSharedRulesDir()
	if err != nil {
		t.Fatalf("locate rules: %v", err)
	}
	sourceSharedDir := filepath.Dir(sourceRulesDir)
	targetSharedDir := t.TempDir()
	targetRulesDir := filepath.Join(targetSharedDir, "rules")
	if err := copyTree(sourceRulesDir, targetRulesDir); err != nil {
		t.Fatalf("copy rules: %v", err)
	}
	if err := copyTree(filepath.Join(sourceSharedDir, "content"), filepath.Join(targetSharedDir, "content")); err != nil {
		t.Fatalf("copy content: %v", err)
	}
	monstersPath := filepath.Join(targetRulesDir, "monsters.v0.json")
	var monsters map[string]any
	b, err := os.ReadFile(monstersPath)
	if err != nil {
		t.Fatalf("read monsters: %v", err)
	}
	if err := json.Unmarshal(b, &monsters); err != nil {
		t.Fatalf("parse monsters: %v", err)
	}
	mutate(monsters["monsters"].(map[string]any))
	b, err = json.MarshalIndent(monsters, "", "  ")
	if err != nil {
		t.Fatalf("marshal monsters: %v", err)
	}
	if err := os.WriteFile(monstersPath, append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write monsters: %v", err)
	}
	if _, err := LoadRules(targetRulesDir); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("LoadRules error = %v, want containing %q", err, want)
	}
}
