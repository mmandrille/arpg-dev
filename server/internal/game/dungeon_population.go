package game

import (
	"fmt"
	"math"
)

func (s *Sim) populateDungeonLevel(level *LevelState) error {
	gen, err := GenerateDungeonLevel(s.seed, level.levelNum, s.rules.DungeonGeneration)
	if err != nil {
		return err
	}
	level.walls = gen.walls
	for _, stair := range gen.stairs {
		def := s.rules.Interactables[stair.defID]
		state := def.InitialState
		if stair.state != "" {
			state = stair.state
		}
		e := &entity{
			kind:              interactableEntity,
			pos:               stair.pos,
			interactableDefID: stair.defID,
			state:             state,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, teleporter := range gen.teleporters {
		def := s.rules.Interactables[teleporter.defID]
		state := def.InitialState
		if teleporter.state != "" {
			state = teleporter.state
		}
		e := &entity{
			kind:              interactableEntity,
			pos:               teleporter.pos,
			interactableDefID: teleporter.defID,
			state:             state,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, chest := range gen.chests {
		def := s.rules.Interactables[chest.defID]
		e := &entity{
			kind:              interactableEntity,
			pos:               chest.pos,
			interactableDefID: chest.defID,
			state:             def.InitialState,
			lootTable:         chest.lootTable,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
		if chest.eliteObjective {
			level.eliteObjectiveChestIDs[e.id] = true
		}
		if chest.questReward {
			level.questRewardChestIDs[e.id] = true
		}
	}
	for _, door := range gen.doors {
		def := s.rules.Interactables[door.defID]
		state := def.InitialState
		if door.state != "" {
			state = door.state
		}
		e := &entity{
			kind:              interactableEntity,
			pos:               door.pos,
			interactableDefID: door.defID,
			state:             state,
		}
		e.id = s.alloc()
		level.entities[e.id] = e
	}
	for _, generated := range gen.loot {
		if _, ok := s.rules.Items[generated.itemDefID]; !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown loot item %s", level.levelNum, generated.itemDefID)
		}
		loot := s.newLootEntity(generated.itemDefID, generated.pos, nil, goldRollContext{levelNum: level.levelNum})
		loot.id = s.alloc()
		level.entities[loot.id] = loot
	}
	for _, generated := range gen.monsters {
		def, ok := s.rules.Monsters[generated.defID]
		if !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown monster %s", level.levelNum, generated.defID)
		}
		lootTable := generated.lootTable
		if generated.isBoss {
			template, ok := s.rules.BossTemplates[generated.bossTemplate]
			if !ok {
				return fmt.Errorf("game: generate dungeon level %d: unknown boss template %s", level.levelNum, generated.bossTemplate)
			}
			var baseOK bool
			def, baseOK = s.rules.Monsters[template.BaseMonsterDefID]
			if !baseOK {
				return fmt.Errorf("game: generate dungeon level %d: unknown boss base monster %s", level.levelNum, template.BaseMonsterDefID)
			}
			generated.defID = template.BaseMonsterDefID
			lootTable = template.LootTable
			if generated.visualModel == "" {
				generated.visualModel = template.Visual.Model
				if len(template.Visual.ModelPool) > 0 {
					visualRNG := NewRNG(SeedToUint64(fmt.Sprintf("%s|boss_visual|%d|%s", s.seed, level.levelNum, generated.bossTemplate)))
					generated.visualModel = template.Visual.ModelPool[visualRNG.IntN(len(template.Visual.ModelPool))]
				}
			}
			generated.visualTint = template.Visual.Color
			generated.visualScale = template.Visual.Scale
		}
		if _, ok := s.rules.LootTables[lootTable]; !ok {
			return fmt.Errorf("game: generate dungeon level %d: unknown monster loot table %s", level.levelNum, lootTable)
		}
		monster := &entity{
			kind:                 monsterEntity,
			pos:                  generated.pos,
			spawnPos:             generated.pos,
			hp:                   def.MaxHP,
			maxHP:                def.MaxHP,
			monsterDefID:         generated.defID,
			monsterRarityID:      generated.rarityID,
			monsterPackID:        generated.packID,
			monsterPackLeader:    generated.packLeader,
			lootTable:            lootTable,
			aiMode:               monsterAIModeIdle,
			isBoss:               generated.isBoss,
			bossTemplateID:       generated.bossTemplate,
			visualModel:          generated.visualModel,
			visualTint:           generated.visualTint,
			visualScale:          generated.visualScale,
			bossPhaseIndex:       -1,
			bossPatternDeckIndex: -1,
		}
		if generated.isBoss {
			template := s.rules.BossTemplates[generated.bossTemplate]
			if len(template.PatternDeck) > 0 {
				monster.bossPatternDeckIndex = 0
				monster.bossPatternID = template.PatternDeck[0]
			}
			if template.Enrage != nil {
				monster.bossEnrageThreshold = template.Enrage.HealthRatioThreshold
			}
			monster.maxHP = roundPositive(float64(def.MaxHP) * template.HPMultiplier)
			monster.hp = monster.maxHP
			if def.AttackDamage != nil {
				scaledAttack := scaleDamageRange(*def.AttackDamage, template.DamageMultiplier)
				monster.monsterAttackDamage = &scaledAttack
			}
			monster.monsterXPReward = roundPositive(float64(def.XPReward) * template.HPMultiplier)
		} else if rarity, ok := s.rules.DungeonGeneration.MonsterRarity(generated.rarityID); ok {
			stats := s.generatedMonsterStats(def, level.levelNum, rarity)
			monster.maxHP = stats.maxHP
			monster.hp = monster.maxHP
			monster.visualScale = rarity.VisualScale
			monster.visualTint = rarity.Color
			monster.monsterAttackDamage = stats.attackDamage
			monster.monsterAttackCooldown = stats.attackCooldown
			monster.monsterArmor = stats.armor
			monster.monsterHitChance = stats.hitChance
			monster.monsterCritChance = stats.critChance
			monster.monsterBlockPercent = stats.blockPercent
			monster.monsterXPReward = stats.xpReward
		}
		s.applyPartyHPScale(level, monster)
		monster.id = s.alloc()
		level.entities[monster.id] = monster
	}
	s.spawnCorpsesOnLevel(level)
	return nil
}

type generatedMonsterEffectiveStats struct {
	maxHP          int
	attackDamage   *DamageRange
	attackCooldown int
	armor          float64
	hitChance      float64
	critChance     float64
	blockPercent   float64
	xpReward       int
}

func (s *Sim) generatedMonsterStats(def MonsterDef, levelNum int, rarity MonsterRarityDef) generatedMonsterEffectiveStats {
	depth := absInt(levelNum)
	depthIndex := DepthIndex(depth)
	scaling := s.rules.DungeonGeneration.MonsterDepthScaling
	stats := generatedMonsterEffectiveStats{
		maxHP:        roundPositive(float64(def.MaxHP) * DepthFactor(scaling.HPPerDepth, depthIndex) * rarity.HPMultiplier),
		armor:        math.Round((float64(def.Armor)+scaling.ArmorPerDepth*float64(depthIndex))*rarity.ArmorMultiplier + rarity.ArmorBonus),
		hitChance:    clampFloat(def.effectiveHitChance(s.rules.Combat)+scaling.HitChancePerDepth*float64(depthIndex)+rarity.HitChanceBonus, 0, scaling.MaxHitChance),
		critChance:   clampFloat(def.effectiveCritChance(s.rules.Combat)+scaling.CritChancePerDepth*float64(depthIndex)+rarity.CritChanceBonus, 0, scaling.MaxCritChance),
		blockPercent: clampFloat(float64(def.BlockPercent)+scaling.BlockPercentPerDepth*float64(depthIndex)+rarity.BlockPercentBonus, 0, scaling.MaxBlockPercent),
		xpReward:     roundPositive(float64(def.XPReward) * rarity.XPMultiplier),
	}
	if def.AttackDamage != nil {
		scaledAttack := scaleDamageRange(*def.AttackDamage, DepthFactor(scaling.DamagePerDepth, depthIndex)*rarity.DamageMultiplier)
		stats.attackDamage = &scaledAttack
	}
	if def.AttackCooldown > 0 {
		cooldownMultiplier := math.Pow(scaling.AttackCooldownMultiplierPerDepth, float64(depthIndex)) * rarity.AttackCooldownMultiplier
		cooldown := int(math.Round(float64(def.AttackCooldown) * cooldownMultiplier))
		if cooldown < scaling.MinAttackCooldownTicks {
			cooldown = scaling.MinAttackCooldownTicks
		}
		stats.attackCooldown = cooldown
	}
	return stats
}
