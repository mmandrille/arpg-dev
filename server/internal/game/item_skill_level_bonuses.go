package game

import (
	"fmt"
	"sort"
	"strings"
)

const (
	statRandomSkillLevel      = "random_skill_level"
	statRandomClassSkillLevel = "random_class_skill_level"
)

// SkillLevelBonusRoll is a rolled per-skill level bonus stored on item payload.
type SkillLevelBonusRoll struct {
	SkillID string `json:"skill_id"`
	Value   int    `json:"value"`
}

// SkillBonusStatusView is server-authored per-skill bonus usability for tooltips.
type SkillBonusStatusView struct {
	SkillID    string `json:"skill_id"`
	SkillClass string `json:"skill_class,omitempty"`
	Value      int    `json:"value"`
	Active     bool   `json:"active"`
	Display    string `json:"display"`
}

func isSkillLevelRollStat(stat string) bool {
	switch stat {
	case statRandomSkillLevel, statRandomClassSkillLevel:
		return true
	default:
		return false
	}
}

func sortedSkillIDsForBonusRoll(rules *Rules, classFilter string) []string {
	if rules == nil || len(rules.Skills) == 0 {
		return nil
	}
	out := make([]string, 0, len(rules.Skills))
	for skillID, def := range rules.Skills {
		if classFilter != "" && def.Class != classFilter {
			continue
		}
		out = append(out, skillID)
	}
	sort.Strings(out)

	return out
}

func (r *Rules) rollSkillLevelBonus(classFilter string, itemLevel int, rng *RNG) (SkillLevelBonusRoll, bool) {
	if rng == nil || itemLevel < 1 {
		return SkillLevelBonusRoll{}, false
	}
	pool := sortedSkillIDsForBonusRoll(r, classFilter)
	if len(pool) == 0 {
		return SkillLevelBonusRoll{}, false
	}
	skillID := pool[rng.IntN(len(pool))]
	value := 1 + rng.IntN(itemLevel)

	return SkillLevelBonusRoll{SkillID: skillID, Value: value}, true
}

func applyRollableStat(
	stat RollableStatDef,
	stats map[string]int,
	bonuses *[]SkillLevelBonusRoll,
	rules *Rules,
	template ItemTemplateDef,
	itemLevel int,
	rng *RNG,
) {
	switch stat.Stat {
	case statRandomSkillLevel:
		if roll, ok := rules.rollSkillLevelBonus("", itemLevel, rng); ok {
			*bonuses = append(*bonuses, roll)
		}
	case statRandomClassSkillLevel:
		if roll, ok := rules.rollSkillLevelBonus(template.ClassRequired, itemLevel, rng); ok {
			*bonuses = append(*bonuses, roll)
		}
	default:
		stats[stat.Stat] += stat.Min + rng.IntN(stat.Max-stat.Min+1)
	}
}

func rollAffixesOntoPayload(
	stats map[string]int,
	bonuses *[]SkillLevelBonusRoll,
	rollableStats []RollableStatDef,
	rules *Rules,
	template ItemTemplateDef,
	itemLevel int,
	rng *RNG,
	rollCount int,
) []RollableStatDef {
	pool := rollableStats
	for i := 0; i < rollCount; i++ {
		stat, ok := weightedRollableStat(pool, rng)
		if !ok {
			continue
		}
		applyRollableStat(stat, stats, bonuses, rules, template, itemLevel, rng)
		if isElementalWeaponAffix(stat.Stat) {
			pool = filterOutElementalWeaponAffixes(pool)
		}
	}

	return pool
}

func skillLevelBonusDisplay(rules *Rules, bonus SkillLevelBonusRoll) string {
	label := bonus.SkillID
	if rules != nil {
		if def, ok := rules.Skills[bonus.SkillID]; ok && def.Name != "" {
			label = def.Name
		}
	}

	return fmt.Sprintf("+%d %s", bonus.Value, label)
}

func (s *Sim) skillBonusActive(skillID string) bool {
	if skillID == "" || s.progression.SkillRanks[skillID] <= 0 {
		return false
	}
	def, ok := s.rules.Skills[skillID]
	if !ok {
		return false
	}
	if def.Class != "" && def.Class != s.progression.CharacterClass {
		return false
	}

	return true
}

func (s *Sim) skillBonusStatus(bonuses []SkillLevelBonusRoll) []SkillBonusStatusView {
	if len(bonuses) == 0 {
		return nil
	}
	out := make([]SkillBonusStatusView, 0, len(bonuses))
	for _, bonus := range bonuses {
		if bonus.SkillID == "" || bonus.Value <= 0 {
			continue
		}
		skillClass := ""
		if def, ok := s.rules.Skills[bonus.SkillID]; ok {
			skillClass = def.Class
		}
		out = append(out, SkillBonusStatusView{
			SkillID:    bonus.SkillID,
			SkillClass: skillClass,
			Value:      bonus.Value,
			Active:     s.skillBonusActive(bonus.SkillID),
			Display:    skillLevelBonusDisplay(s.rules, bonus),
		})
	}

	return out
}

func (s *Sim) annotateSkillBonusStatus(payload *ItemRollPayload, set func([]SkillBonusStatusView)) {
	if payload == nil || len(payload.SkillLevelBonuses) == 0 {
		return
	}
	set(s.skillBonusStatus(payload.SkillLevelBonuses))
}

func (s *Sim) equippedPerSkillBonus(skillID string) int {
	if skillID == "" || !s.skillBonusActive(skillID) {
		return 0
	}
	total := 0
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil || item.rollPayload == nil {
			continue
		}
		for _, bonus := range item.rollPayload.SkillLevelBonuses {
			if bonus.SkillID == skillID {
				total += bonus.Value
			}
		}
	}

	return total
}

func cloneSkillLevelBonusRolls(in []SkillLevelBonusRoll) []SkillLevelBonusRoll {
	if len(in) == 0 {
		return nil
	}
	out := make([]SkillLevelBonusRoll, len(in))
	copy(out, in)

	return out
}

func validateSkillLevelRollStat(templateID string, stat RollableStatDef, template ItemTemplateDef) error {
	if !isSkillLevelRollStat(stat.Stat) {
		return nil
	}
	if stat.Min != 0 || stat.Max != 0 {
		return fmt.Errorf("game: invalid rules item_templates.%s rollable stat %s: min and max must be 0", templateID, stat.Stat)
	}
	if stat.Stat == statRandomClassSkillLevel && strings.TrimSpace(template.ClassRequired) == "" {
		return fmt.Errorf("game: invalid rules item_templates.%s rollable stat %s: requires class_required", templateID, stat.Stat)
	}

	return nil
}
