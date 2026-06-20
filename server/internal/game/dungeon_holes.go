package game

// Holes (chasms) are a hard-blocking floor hazard. They share the generation
// algorithm in dungeon_floor_features.go with water, differing only in their seed
// token and kind.

func holeFeatureSpec(rules DungeonGenerationRules) floorFeatureSpec {
	return floorFeatureSpec{
		rules:     rules.ObstacleGeneration.Holes,
		field:     "holes",
		seedToken: "|holes|",
		kind:      obstacleKindHole,
	}
}

func validateHoleGenerationRules(holes FloorFeatureGenerationRules, floor DungeonFloorSize) error {
	return validateFloorFeatureRules(floorFeatureSpec{rules: holes, field: "holes", kind: obstacleKindHole}, floor)
}

func placeDungeonHoles(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	return placeDungeonFloorFeature(seed, holeFeatureSpec(rules), rules, out)
}
