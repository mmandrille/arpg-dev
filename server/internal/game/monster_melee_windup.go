package game

import "strconv"

func (s *Sim) advanceMonsterMeleeWindups(res *TickResult) {
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		monster := s.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 || monster.attackWindupRemaining <= 0 {
			continue
		}
		monster.attackWindupRemaining--
		if monster.attackWindupRemaining > 0 {
			continue
		}
		target := s.activeLevel().entities[monster.attackWindupTargetID]
		if target == nil || target.hp <= 0 {
			monster.attackWindupTargetID = 0
			monster.attackWindupDamage = DamageRange{}
			continue
		}
		damage := monster.attackWindupDamage
		monster.attackWindupTargetID = 0
		monster.attackWindupDamage = DamageRange{}
		if target.kind == companionEntity {
			s.damageCompanionByMonster(monster, target, damage, "", res)
			continue
		}
		s.damagePlayerByMonster(monster, target, damage, "", res)
	}
}

func (s *Sim) tryStartMonsterMeleeWindup(monster *entity, target *entity, def MonsterDef, attackDamage DamageRange, res *TickResult) bool {
	if monster == nil || target == nil || def.AttackWindupTicks <= 0 {
		return false
	}
	if def.effectiveAttackStyle() != monsterAttackStyleMelee || def.effectiveAttackMode() != attackModeMelee {
		if def.effectiveAttackStyle() != monsterAttackStylePounce || def.effectiveAttackMode() != attackModeMelee {
			return false
		}
	}
	if monster.attackWindupRemaining > 0 {
		return true
	}
	monster.attackWindupRemaining = def.AttackWindupTicks
	monster.attackWindupTargetID = target.id
	monster.attackWindupDamage = attackDamage
	remaining := def.AttackWindupTicks
	total := def.AttackWindupTicks
	res.Events = append(res.Events, Event{
		EventType:      "monster_attack_windup",
		SourceEntityID: strconv.FormatUint(monster.id, 10),
		TargetEntityID: strconv.FormatUint(target.id, 10),
		RemainingTicks: &remaining,
		TotalTicks:     &total,
		AttackStyle:    def.effectiveAttackStyle(),
	})

	return true
}
