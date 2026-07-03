package game

func classRequirementStatus(classRequired, currentClass string) (RequirementStatusView, bool) {
	if classRequired == "" {
		return RequirementStatusView{}, true
	}

	met := classRequired == currentClass
	current := 0
	if met {
		current = 1
	}

	return RequirementStatusView{
		Stat:     "class",
		Required: 1,
		Current:  current,
		Met:      met,
		ClassID:  classRequired,
	}, met
}

func (s *Sim) itemClassRequired(item *invItem) string {
	if item == nil {
		return ""
	}

	if item.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if ok {
			return template.ClassRequired
		}
	}

	def, ok := s.rules.Items[item.itemDefID]
	if !ok {
		return ""
	}

	return def.ClassRequired
}

func (s *Sim) annotateItemRequirementStatus(requirements map[string]int, item *invItem, set func([]RequirementStatusView, *bool)) {
	status, met := s.requirementStatusForItem(requirements, item)
	if len(status) == 0 {
		return
	}

	metCopy := met
	set(status, &metCopy)
}

func (s *Sim) lootItemForRequirements(e *entity) *invItem {
	if e == nil || e.kind != lootEntity {
		return nil
	}

	if e.rollPayload != nil {
		return &invItem{
			itemDefID:   e.rollPayload.ItemTemplateID,
			rollPayload: e.rollPayload,
		}
	}

	if e.itemDefID == "" {
		return nil
	}

	return &invItem{itemDefID: e.itemDefID}
}

func appendClassRequirementStatus(status []RequirementStatusView, met bool, classRequired, currentClass string) ([]RequirementStatusView, bool) {
	classStatus, classMet := classRequirementStatus(classRequired, currentClass)
	if classRequired == "" {
		return status, met
	}

	return append(status, classStatus), met && classMet
}

func (s *Sim) requirementStatusForItem(requirements map[string]int, item *invItem) ([]RequirementStatusView, bool) {
	status, met := s.requirementStatus(requirements)

	return appendClassRequirementStatus(status, met, s.itemClassRequired(item), s.progression.CharacterClass)
}
