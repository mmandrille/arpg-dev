package game

// RarityDef controls how many bounded stat rolls a rolled item gets.
type RarityDef struct {
	Weight         int    `json:"weight"`
	StatRolls      int    `json:"stat_rolls"`
	StatRollsMin   int    `json:"stat_rolls_min"`
	StatRollsMax   int    `json:"stat_rolls_max"`
	RandomRollable *bool  `json:"random_rollable,omitempty"`
	NamePrefix     string `json:"name_prefix"`
}

// RollableStatDef is one weighted bounded stat increment.
type RollableStatDef struct {
	Stat      string `json:"stat"`
	MinRarity string `json:"min_rarity,omitempty"`
	Min       int    `json:"min"`
	Max       int    `json:"max"`
	Weight    int    `json:"weight"`
}

func (r *Rules) rarityRandomRollable(rarityID string) bool {
	rarity, ok := r.Rarities[rarityID]
	if !ok {
		return false
	}
	return rarity.RandomRollable == nil || *rarity.RandomRollable
}

func itemRarityRank(rarityID string) int {
	switch rarityID {
	case "common":
		return 0
	case "magic":
		return 1
	case "rare":
		return 2
	case "unique", "set":
		return 3
	default:
		return -1
	}
}
