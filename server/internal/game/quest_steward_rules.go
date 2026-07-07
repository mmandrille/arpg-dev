package game

import (
	"fmt"
	"sort"
)

type QuestStewardRules struct {
	HuntQuest      QuestStewardHuntQuestRules `json:"hunt_quest"`
	Trophies       []QuestStewardTrophyRule   `json:"trophies"`
	RewardFamilies []QuestStewardFamilyRule   `json:"reward_families"`

	trophyByMonster map[string]QuestStewardTrophyRule
	trophyByItem    map[string]QuestStewardTrophyRule
	familyByID      map[string]QuestStewardFamilyRule
}

type QuestStewardHuntQuestRules struct {
	Enabled            bool   `json:"enabled"`
	FloorChancePercent int    `json:"floor_chance_percent"`
	ChoiceCount        int    `json:"choice_count"`
	MinRarity          string `json:"min_rarity"`
}

type QuestStewardTrophyRule struct {
	MonsterDefID string `json:"monster_def_id"`
	ItemDefID    string `json:"item_def_id"`
	TrophyLabel  string `json:"trophy_label"`
}

type QuestStewardFamilyRule struct {
	FamilyID    string   `json:"family_id"`
	Label       string   `json:"label"`
	TemplateIDs []string `json:"template_ids"`
}

type generatedStewardHunt struct {
	MonsterIndex    int
	MonsterDefID    string
	TrophyItemDefID string
	TrophyLabel     string
}

type stewardHuntLevelState struct {
	Active          bool
	MonsterDefID    string
	MonsterName     string
	TrophyItemDefID string
	TrophyLabel     string
	SourceDepth     int
	TargetEntityID  uint64
	Complete        bool
}

type questStewardOffer struct {
	OfferID  string
	FamilyID string
	Label    string
}

type questStewardOffersState struct {
	GiverEntityID    uint64
	TrophyInstanceID uint64
	SourceDepth      int
	Offers           []questStewardOffer
}

func loadQuestStewardRules(dir string, r *Rules) error {
	var steward QuestStewardRules
	if err := readJSON(dir+"/quest_steward.v0.json", &steward); err != nil {
		return err
	}
	if err := validateQuestStewardRules(steward, r); err != nil {
		return err
	}
	steward.index()
	r.QuestSteward = steward
	return nil
}

func (q *QuestStewardRules) index() {
	q.trophyByMonster = make(map[string]QuestStewardTrophyRule, len(q.Trophies))
	q.trophyByItem = make(map[string]QuestStewardTrophyRule, len(q.Trophies))
	for _, trophy := range q.Trophies {
		q.trophyByMonster[trophy.MonsterDefID] = trophy
		q.trophyByItem[trophy.ItemDefID] = trophy
	}
	q.familyByID = make(map[string]QuestStewardFamilyRule, len(q.RewardFamilies))
	for _, family := range q.RewardFamilies {
		q.familyByID[family.FamilyID] = family
	}
}

func validateQuestStewardRules(steward QuestStewardRules, r *Rules) error {
	hunt := steward.HuntQuest
	if hunt.FloorChancePercent < 0 || hunt.FloorChancePercent > 100 {
		return fmt.Errorf("game: invalid rules quest_steward.hunt_quest.floor_chance_percent: must be 0..100")
	}
	if hunt.ChoiceCount <= 0 {
		return fmt.Errorf("game: invalid rules quest_steward.hunt_quest.choice_count: must be positive")
	}
	if hunt.MinRarity == "" {
		return fmt.Errorf("game: invalid rules quest_steward.hunt_quest.min_rarity: must be non-empty")
	}
	if _, ok := r.Rarities[hunt.MinRarity]; !ok {
		return fmt.Errorf("game: invalid rules quest_steward.hunt_quest.min_rarity: unknown rarity %s", hunt.MinRarity)
	}
	if len(steward.Trophies) == 0 {
		return fmt.Errorf("game: invalid rules quest_steward.trophies: must be non-empty")
	}
	seenMonsters := map[string]struct{}{}
	seenItems := map[string]struct{}{}
	for _, trophy := range steward.Trophies {
		if trophy.MonsterDefID == "" || trophy.ItemDefID == "" || trophy.TrophyLabel == "" {
			return fmt.Errorf("game: invalid rules quest_steward.trophies: monster_def_id, item_def_id, and trophy_label are required")
		}
		if _, ok := r.Monsters[trophy.MonsterDefID]; !ok {
			return fmt.Errorf("game: invalid rules quest_steward.trophies: unknown monster %s", trophy.MonsterDefID)
		}
		item, ok := r.Items[trophy.ItemDefID]
		if !ok {
			return fmt.Errorf("game: invalid rules quest_steward.trophies: unknown item %s", trophy.ItemDefID)
		}
		if item.Category != "quest" {
			return fmt.Errorf("game: invalid rules quest_steward.trophies: item %s must be category quest", trophy.ItemDefID)
		}
		if _, dup := seenMonsters[trophy.MonsterDefID]; dup {
			return fmt.Errorf("game: invalid rules quest_steward.trophies: duplicate monster %s", trophy.MonsterDefID)
		}
		seenMonsters[trophy.MonsterDefID] = struct{}{}
		if _, dup := seenItems[trophy.ItemDefID]; dup {
			return fmt.Errorf("game: invalid rules quest_steward.trophies: duplicate item %s", trophy.ItemDefID)
		}
		seenItems[trophy.ItemDefID] = struct{}{}
	}
	if len(steward.RewardFamilies) < hunt.ChoiceCount {
		return fmt.Errorf("game: invalid rules quest_steward.reward_families: need at least choice_count families")
	}
	seenFamilies := map[string]struct{}{}
	for _, family := range steward.RewardFamilies {
		if family.FamilyID == "" || family.Label == "" || len(family.TemplateIDs) == 0 {
			return fmt.Errorf("game: invalid rules quest_steward.reward_families: family_id, label, and template_ids are required")
		}
		if _, dup := seenFamilies[family.FamilyID]; dup {
			return fmt.Errorf("game: invalid rules quest_steward.reward_families: duplicate family %s", family.FamilyID)
		}
		seenFamilies[family.FamilyID] = struct{}{}
		for _, templateID := range family.TemplateIDs {
			if _, ok := r.ItemTemplates[templateID]; !ok {
				return fmt.Errorf("game: invalid rules quest_steward.reward_families.%s: unknown template %s", family.FamilyID, templateID)
			}
		}
	}

	return nil
}

func (r *Rules) questStewardTrophyForMonster(monsterDefID string) (QuestStewardTrophyRule, bool) {
	trophy, ok := r.QuestSteward.trophyByMonster[monsterDefID]
	return trophy, ok
}

func (r *Rules) questStewardTrophyForItem(itemDefID string) (QuestStewardTrophyRule, bool) {
	trophy, ok := r.QuestSteward.trophyByItem[itemDefID]
	return trophy, ok
}

func (r *Rules) questStewardFamilyByID(familyID string) (QuestStewardFamilyRule, bool) {
	family, ok := r.QuestSteward.familyByID[familyID]
	return family, ok
}

func sortedQuestStewardFamilies(families []QuestStewardFamilyRule) []QuestStewardFamilyRule {
	out := append([]QuestStewardFamilyRule(nil), families...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].FamilyID < out[j].FamilyID
	})

	return out
}
