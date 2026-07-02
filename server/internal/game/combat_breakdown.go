package game

import (
	"fmt"
	"strconv"
)

// CombatBreakdownLineView is one labeled step in an authoritative damage formula log.
type CombatBreakdownLineView struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func buildCombatDamageBreakdown(
	attacker effectiveCombatStats,
	defender effectiveCombatStats,
	damageRange DamageRange,
	outcome combatResolution,
	resistance float64,
	damageType string,
	attackLabel string,
) []CombatBreakdownLineView {
	lines := []CombatBreakdownLineView{}
	if attackLabel != "" {
		lines = append(lines, CombatBreakdownLineView{Label: "Attack", Value: attackLabel})
	}

	lines = append(lines, CombatBreakdownLineView{
		Label: "Damage range",
		Value: fmt.Sprintf("%d-%d", damageRange.Min, damageRange.Max),
	})
	lines = append(lines, CombatBreakdownLineView{
		Label: "Hit chance",
		Value: fmt.Sprintf("%.0f%%", attacker.HitChance*100),
	})

	if outcome.Outcome == "miss" {
		lines = append(lines, CombatBreakdownLineView{Label: "Result", Value: "Miss"})
		return lines
	}

	if outcome.Blocked {
		lines = append(lines, CombatBreakdownLineView{
			Label: "Block chance",
			Value: fmt.Sprintf("%.0f%%", defender.BlockPercent),
		})
		lines = append(lines, CombatBreakdownLineView{Label: "Result", Value: "Blocked"})
		return lines
	}

	lines = append(lines, CombatBreakdownLineView{
		Label: "Rolled damage",
		Value: strconv.Itoa(outcome.RawDamage),
	})

	if outcome.Critical {
		lines = append(lines, CombatBreakdownLineView{
			Label: "Critical multiplier",
			Value: fmt.Sprintf("%.2fx", attacker.CritDamage),
		})
	}

	lines = append(lines, CombatBreakdownLineView{
		Label: "Target armor",
		Value: fmt.Sprintf("%.0f", defender.Armor),
	})
	lines = append(lines, CombatBreakdownLineView{
		Label: "After armor",
		Value: strconv.Itoa(outcome.MitigatedDamage),
	})

	if resistance != 0 && damageType != "" && damageType != damageTypeForce {
		lines = append(lines, CombatBreakdownLineView{
			Label: fmt.Sprintf("%s resistance", damageType),
			Value: fmt.Sprintf("%.0f%%", resistance*100),
		})
	}

	lines = append(lines, CombatBreakdownLineView{
		Label: "Final damage",
		Value: strconv.Itoa(outcome.Damage),
	})

	return lines
}

func buildSkillDamageBreakdown(
	defender effectiveCombatStats,
	damageRange DamageRange,
	outcome combatResolution,
	resistance float64,
	damageType string,
	attackLabel string,
) []CombatBreakdownLineView {
	lines := []CombatBreakdownLineView{}
	if attackLabel != "" {
		lines = append(lines, CombatBreakdownLineView{Label: "Attack", Value: attackLabel})
	}

	lines = append(lines, CombatBreakdownLineView{
		Label: "Damage range",
		Value: fmt.Sprintf("%d-%d", damageRange.Min, damageRange.Max),
	})
	lines = append(lines, CombatBreakdownLineView{
		Label: "Rolled damage",
		Value: strconv.Itoa(outcome.RawDamage),
	})
	lines = append(lines, CombatBreakdownLineView{
		Label: "Target armor",
		Value: fmt.Sprintf("%.0f", defender.Armor),
	})
	lines = append(lines, CombatBreakdownLineView{
		Label: "After armor",
		Value: strconv.Itoa(outcome.MitigatedDamage),
	})

	if resistance != 0 && damageType != "" && damageType != damageTypeForce {
		lines = append(lines, CombatBreakdownLineView{
			Label: fmt.Sprintf("%s resistance", damageType),
			Value: fmt.Sprintf("%.0f%%", resistance*100),
		})
	}

	lines = append(lines, CombatBreakdownLineView{
		Label: "Final damage",
		Value: strconv.Itoa(outcome.Damage),
	})

	return lines
}
