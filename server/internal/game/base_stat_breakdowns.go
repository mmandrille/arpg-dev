package game

type baseStatBreakdownState struct {
	key     string
	current int
	sources []StatBreakdownSourceView
}

func (s *Sim) baseStatBreakdownViews() []StatBreakdownView {
	states := s.baseStatBreakdownStates()
	for _, slot := range equipmentSlots {
		item := s.findItemByID(s.equipped[slot])
		if item == nil {
			continue
		}
		label := s.itemDisplayName(item)
		itemID := idStr(item.instanceID)
		baseStats, rolledStats := s.itemBaseAndRollStats(item)
		for idx := range states {
			states[idx].addItemSource(baseStats[states[idx].key], label, itemID, "equipment_base")
			states[idx].addItemSource(rolledStats[states[idx].key], label, itemID, "equipment_roll")
		}
	}
	setStats := s.equippedSetBonusStats()
	for idx := range states {
		states[idx].addSource(setStats[states[idx].key], "Set bonus", "set_bonus")
		_, rows := s.passiveSkillStatSources(states[idx].key, 1)
		states[idx].addRows(rows)
	}
	for _, skillID := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[skillID]
		if effect.EndsTick <= s.tick || effect.Percent <= 0 {
			continue
		}
		for _, stat := range effect.Stats {
			for idx := range states {
				if states[idx].key != stat {
					continue
				}
				before := states[idx].current
				after := scaleStatPercent(before, effect.Percent)
				if after != before {
					states[idx].addSource(after-before, s.skillEffectLabel(effect), "skill_effect")
				}
			}
		}
	}
	out := make([]StatBreakdownView, 0, len(states))
	for _, state := range states {
		out = append(out, StatBreakdownView{
			Key:           state.key,
			Value:         float64(state.current),
			UncappedValue: float64(state.current),
			Cap:           nil,
			Sources:       state.sources,
		})
	}
	return out
}

func (s *Sim) baseStatBreakdownStates() []baseStatBreakdownState {
	base := s.progression.BaseStats
	return []baseStatBreakdownState{
		newBaseStatBreakdownState("str", "STR", base.Str),
		newBaseStatBreakdownState("dex", "DEX", base.Dex),
		newBaseStatBreakdownState("vit", "VIT", base.Vit),
		newBaseStatBreakdownState("magic", "Magic", base.Magic),
	}
}

func newBaseStatBreakdownState(key string, label string, base int) baseStatBreakdownState {
	return baseStatBreakdownState{
		key:     key,
		current: base,
		sources: []StatBreakdownSourceView{{
			Label: "Base " + label,
			Value: float64(base),
			Kind:  "base_stat",
		}},
	}
}

func (b *baseStatBreakdownState) addItemSource(value int, label string, itemID string, kind string) {
	if value == 0 {
		return
	}
	b.current += value
	b.sources = append(b.sources, StatBreakdownSourceView{
		Label:          label,
		Value:          float64(value),
		Kind:           kind,
		ItemInstanceID: itemID,
	})
}

func (b *baseStatBreakdownState) addSource(value int, label string, kind string) {
	if value == 0 {
		return
	}
	b.current += value
	b.sources = append(b.sources, StatBreakdownSourceView{
		Label: label,
		Value: float64(value),
		Kind:  kind,
	})
}

func (b *baseStatBreakdownState) addRows(rows []StatBreakdownSourceView) {
	for _, row := range rows {
		value := int(row.Value)
		if value == 0 {
			continue
		}
		b.current += value
		b.sources = append(b.sources, row)
	}
}
