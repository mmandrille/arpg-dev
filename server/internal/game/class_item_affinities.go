package game

import (
	"fmt"
	"math"
	"strings"
)

// ClassAffinityDef is a template-owned class affinity roll range.
type ClassAffinityDef struct {
	Class string `json:"class"`
	Stat  string `json:"stat"`
	Min   int    `json:"min"`
	Max   int    `json:"max"`
	Mode  string `json:"mode,omitempty"`
}

// ClassAffinityRoll is a rolled class affinity stored on an item payload.
type ClassAffinityRoll struct {
	Class string `json:"class"`
	Stat  string `json:"stat"`
	Value int    `json:"value"`
	Mode  string `json:"mode,omitempty"`
}

// ClassAffinityStatusView is server-authored affinity usability for the viewing character.
type ClassAffinityStatusView struct {
	Class   string `json:"class"`
	Stat    string `json:"stat"`
	Value   int    `json:"value"`
	Active  bool   `json:"active"`
	Display string `json:"display"`
}

func rollClassAffinities(defs []ClassAffinityDef, rng *RNG) []ClassAffinityRoll {
	if len(defs) == 0 || rng == nil {
		return nil
	}
	out := make([]ClassAffinityRoll, 0, len(defs))
	for _, def := range defs {
		if def.Class == "" || def.Stat == "" || def.Max < def.Min {
			continue
		}
		value := def.Min
		if def.Max > def.Min {
			value = def.Min + rng.IntN(def.Max-def.Min+1)
		}
		out = append(out, ClassAffinityRoll{
			Class: def.Class,
			Stat:  def.Stat,
			Value: value,
			Mode:  def.Mode,
		})
	}
	return out
}

func classAffinityActive(characterClass string, affinity ClassAffinityRoll) bool {
	switch affinity.Mode {
	case "penalty_if_not_class":
		return characterClass != affinity.Class
	default:
		return characterClass == affinity.Class
	}
}

func classAffinityDisplay(affinity ClassAffinityRoll) string {
	classLabel := titleClassName(affinity.Class)
	if affinity.Mode == "penalty_if_not_class" {
		classLabel = "non-" + classLabel
	}
	return fmt.Sprintf("%s (%s)", formatAffinityStatValue(affinity.Stat, affinity.Value), classLabel)
}

func titleClassName(class string) string {
	if class == "" {
		return ""
	}
	return strings.ToUpper(class[:1]) + class[1:]
}

func formatAffinityStatValue(stat string, value int) string {
	sign := "+"
	if value < 0 {
		sign = ""
	}
	switch stat {
	case "damage_percent", "attack_speed_percent", "reach_percent", "max_mana_percent", "block_percent":
		return fmt.Sprintf("%s%d%% %s", sign, value, humanAffinityStat(stat))
	default:
		return fmt.Sprintf("%s%d %s", sign, value, humanAffinityStat(stat))
	}
}

func humanAffinityStat(stat string) string {
	switch stat {
	case "attack_speed_percent":
		return "attack speed"
	case "damage_percent":
		return "damage"
	case "reach_percent":
		return "range"
	case "max_mana_percent":
		return "max mana"
	default:
		return strings.ReplaceAll(stat, "_", " ")
	}
}

func (s *Sim) annotateClassAffinityStatus(payload *ItemRollPayload, set func([]ClassAffinityStatusView)) {
	if payload == nil || len(payload.ClassAffinities) == 0 {
		return
	}
	set(s.classAffinityStatus(payload.ClassAffinities))
}

func (s *Sim) equippedClassAffinityTotals() map[string]int {
	totals := map[string]int{}
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil || item.rollPayload == nil {
			continue
		}
		for _, affinity := range item.rollPayload.ClassAffinities {
			if !classAffinityActive(s.progression.CharacterClass, affinity) {
				continue
			}
			totals[affinity.Stat] += affinity.Value
		}
	}
	return totals
}

func (s *Sim) classAffinityStatus(affinities []ClassAffinityRoll) []ClassAffinityStatusView {
	if len(affinities) == 0 {
		return nil
	}
	characterClass := s.progression.CharacterClass
	out := make([]ClassAffinityStatusView, 0, len(affinities))
	for _, affinity := range affinities {
		active := classAffinityActive(characterClass, affinity)
		out = append(out, ClassAffinityStatusView{
			Class:   affinity.Class,
			Stat:    affinity.Stat,
			Value:   affinity.Value,
			Active:  active,
			Display: classAffinityDisplay(affinity),
		})
	}
	return out
}

func (s *Sim) itemClassAffinityTotal(item *invItem, stat string) int {
	if item == nil || item.rollPayload == nil {
		return 0
	}
	total := 0
	for _, affinity := range item.rollPayload.ClassAffinities {
		if affinity.Stat != stat {
			continue
		}
		if !classAffinityActive(s.progression.CharacterClass, affinity) {
			continue
		}
		total += affinity.Value
	}
	return total
}

func (s *Sim) applyClassAffinityCombatStats(
	damageMin, damageMax, itemSpeedPercent, maxMana *float64,
	damageMinSources, damageMaxSources, attackSpeedSources, maxManaSources *[]StatBreakdownSourceView,
) {
	totals := s.equippedClassAffinityTotals()
	if pct := totals["damage_percent"]; pct != 0 {
		beforeMin, beforeMax := *damageMin, *damageMax
		*damageMin = applyPercentDelta(*damageMin, pct)
		*damageMax = applyPercentDelta(*damageMax, pct)
		*damageMinSources = append(*damageMinSources, StatBreakdownSourceView{Label: "Class affinity damage", Value: *damageMin - beforeMin, Kind: "class_affinity"})
		*damageMaxSources = append(*damageMaxSources, StatBreakdownSourceView{Label: "Class affinity damage", Value: *damageMax - beforeMax, Kind: "class_affinity"})
	}
	if pct := totals["attack_speed_percent"]; pct != 0 {
		*itemSpeedPercent += float64(pct)
		*attackSpeedSources = append(*attackSpeedSources, StatBreakdownSourceView{Label: "Class affinity attack speed", Value: float64(pct), Kind: "class_affinity"})
	}
	if pct := totals["max_mana_percent"]; pct != 0 {
		before := *maxMana
		*maxMana = applyPercentDelta(*maxMana, pct)
		*maxManaSources = append(*maxManaSources, StatBreakdownSourceView{Label: "Class affinity max mana", Value: *maxMana - before, Kind: "class_affinity"})
	}
}

func applyPercentDelta(base float64, percent int) float64 {
	if percent == 0 {
		return base
	}
	return base * (1.0 + float64(percent)/100.0)
}

func scaleIntRangeByPercent(minValue, maxValue int, percent int) (int, int) {
	if percent == 0 {
		return minValue, maxValue
	}
	scale := 1.0 + float64(percent)/100.0
	outMin := int(math.Round(float64(minValue) * scale))
	outMax := int(math.Round(float64(maxValue) * scale))
	if outMin < 0 {
		outMin = 0
	}
	if outMax < outMin {
		outMax = outMin
	}
	return outMin, outMax
}

func validateClassAffinityDefs(templateID string, affinities []ClassAffinityDef, classes map[string]CharacterClassDef) error {
	for i, affinity := range affinities {
		if _, ok := classes[affinity.Class]; !ok {
			return fmt.Errorf("game: invalid rules item_templates.%s.class_affinities[%d].class: unknown class %s", templateID, i, affinity.Class)
		}
		if !isSupportedClassAffinityStat(affinity.Stat) {
			return fmt.Errorf("game: invalid rules item_templates.%s.class_affinities[%d].stat: unsupported stat %s", templateID, i, affinity.Stat)
		}
		if affinity.Max < affinity.Min {
			return fmt.Errorf("game: invalid rules item_templates.%s.class_affinities[%d]: max must be >= min", templateID, i)
		}
		if affinity.Mode != "" && affinity.Mode != "penalty_if_not_class" {
			return fmt.Errorf("game: invalid rules item_templates.%s.class_affinities[%d].mode: %s", templateID, i, affinity.Mode)
		}
	}

	return nil
}

func isSupportedClassAffinityStat(stat string) bool {
	switch stat {
	case "damage_percent", "attack_speed_percent", "reach_percent", "max_mana_percent":
		return true
	default:
		return false
	}
}
