package game

import (
	"fmt"
	"math"
)

const (
	damageTypeForce     = "force"
	damageTypeCold      = "cold"
	damageTypeFire      = "fire"
	damageTypePoison    = "poison"
	damageTypeLightning = "lightning"
)

func validateDamageType(label string, damageType string) error {
	if damageType == "" || validDamageType(damageType) {
		return nil
	}
	return fmt.Errorf("game: invalid rules %s: unsupported damage type %s", label, damageType)
}

func canonicalDamageType(damageType string) string {
	switch damageType {
	case damageTypeCold, damageTypeFire, damageTypePoison, damageTypeLightning:
		return damageType
	default:
		return damageTypeForce
	}
}

func validDamageType(damageType string) bool {
	return damageType == damageTypeForce ||
		damageType == damageTypeCold ||
		damageType == damageTypeFire ||
		damageType == damageTypePoison ||
		damageType == damageTypeLightning
}

func (s *Sim) skillDamageType(def SkillDef) string {
	return canonicalDamageType(def.DamageType)
}

func (s *Sim) playerWeaponDamageTypeForSlot(slot string) string {
	if slot == "" {
		slot = mainHandSlot
	}
	item := s.findItemByID(s.equipped[slot])
	if item == nil {
		return damageTypeForce
	}
	if item.rollPayload != nil {
		return damageTypeForce
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok {
		return canonicalDamageType(def.DamageType)
	}
	return damageTypeForce
}

func (s *Sim) applyMonsterResistanceToOutcome(target *entity, damageType string, outcome *combatResolution) {
	if outcome == nil {
		return
	}
	damageType = canonicalDamageType(damageType)
	outcome.DamageType = damageType
	if target == nil || target.kind != monsterEntity || !outcome.Hit || outcome.Blocked {
		return
	}
	resistance := s.monsterResistance(target, damageType)
	if resistance == 0 {
		return
	}
	outcome.Damage = s.applyResistanceToDamage(outcome.MitigatedDamage, resistance)
}

func (s *Sim) applyResistanceToDamage(damage int, resistance float64) int {
	if resistance >= 1 {
		return 0
	}
	adjusted := int(math.Round(float64(damage) * (1.0 - resistance)))
	if adjusted < s.rules.Combat.MinimumDamage {
		return s.rules.Combat.MinimumDamage
	}
	return adjusted
}

func (s *Sim) monsterResistance(target *entity, damageType string) float64 {
	if target == nil || target.monsterDefID == "" {
		return 0
	}
	def, ok := s.rules.Monsters[target.monsterDefID]
	if !ok || len(def.Resistances) == 0 {
		return 0
	}
	return clampFloat(def.Resistances[canonicalDamageType(damageType)], -1, 1)
}
