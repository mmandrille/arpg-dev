package game

// Water is a hard-blocking floor hazard. It shares the generation algorithm in
// dungeon_floor_features.go with holes, differing only in its seed token and kind.

func waterFeatureSpec(rules DungeonGenerationRules) floorFeatureSpec {
	return floorFeatureSpec{
		rules:     rules.ObstacleGeneration.Water,
		field:     "water",
		seedToken: "|water|",
		kind:      obstacleKindWater,
	}
}

func validateWaterGenerationRules(water FloorFeatureGenerationRules, floor DungeonFloorSize) error {
	return validateFloorFeatureRules(floorFeatureSpec{rules: water, field: "water", kind: obstacleKindWater}, floor)
}

func placeDungeonWater(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	return placeDungeonFloorFeature(seed, waterFeatureSpec(rules), rules, out)
}
