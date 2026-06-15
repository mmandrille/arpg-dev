package game

func countDropEntryRefs(refs ...string) int {
	count := 0
	for _, ref := range refs {
		if ref != "" {
			count++
		}
	}
	return count
}
