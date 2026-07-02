package game

// migrateLegacySkillRanks rewrites deprecated skill ids in persisted rank maps.
func migrateLegacySkillRanks(in map[string]int) {
	if rank, ok := in["ligthing"]; ok {
		if _, has := in["lightning"]; !has && rank > 0 {
			in["lightning"] = rank
		}
		delete(in, "ligthing")
	}
}
