package game

import "fmt"

const (
	monsterAttackStyleMelee  = "melee"
	monsterAttackStyleDive   = "dive"
	monsterAttackStylePounce = "pounce"
)

func (d MonsterDef) effectiveAttackStyle() string {
	if d.AttackStyle == "" {
		return monsterAttackStyleMelee
	}
	return d.AttackStyle
}

func validateMonsterAttackStyle(id string, def MonsterDef, attackMode string, behavior string, unarmedReach float64) error {
	attackStyle := def.effectiveAttackStyle()
	switch attackStyle {
	case monsterAttackStyleMelee:
		return nil
	case monsterAttackStyleDive, monsterAttackStylePounce:
		if attackMode != attackModeMelee {
			return fmt.Errorf("game: invalid rules monsters.%s.attack_style: %s requires melee attack_mode", id, attackStyle)
		}
		if behavior != monsterBehaviorChase {
			return fmt.Errorf("game: invalid rules monsters.%s.attack_style: %s requires chase behavior", id, attackStyle)
		}
		if def.AttackDamage == nil || def.AttackCooldown <= 0 {
			return fmt.Errorf("game: invalid rules monsters.%s.attack_style: %s requires attack_damage and attack_cooldown_ticks", id, attackStyle)
		}
		if attackStyle == monsterAttackStylePounce && def.AttackRange <= unarmedReach {
			return fmt.Errorf("game: invalid rules monsters.%s.attack_range: pounce reach must exceed melee reach", id)
		}
		return nil
	default:
		return fmt.Errorf("game: invalid rules monsters.%s.attack_style: %s", id, def.AttackStyle)
	}
}
