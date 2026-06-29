package game

// tickCollisionCache stores per-tick sorted entity/player id lists for the
// active level so collision probes avoid rebuilding sorted slices each call.
type tickCollisionCache struct {
	levelNum  int
	entityIDs []uint64
	playerIDs []uint64
}

func (s *Sim) resetTickCollisionCache() {
	level := s.activeLevel()
	if level == nil {
		s.tickCollisionCache = tickCollisionCache{levelNum: s.currentLevel}
		return
	}
	s.tickCollisionCache = tickCollisionCache{
		levelNum:  s.currentLevel,
		entityIDs: sortedEntityIDs(level.entities),
		playerIDs: sortedPlayerIDs(s.players),
	}
}

func (s *Sim) cachedSortedEntityIDs() []uint64 {
	if s.tickCollisionCache.levelNum != s.currentLevel || s.tickCollisionCache.entityIDs == nil {
		s.resetTickCollisionCache()
	}
	return s.tickCollisionCache.entityIDs
}

func (s *Sim) cachedSortedPlayerIDs() []uint64 {
	if s.tickCollisionCache.levelNum != s.currentLevel || s.tickCollisionCache.playerIDs == nil {
		s.resetTickCollisionCache()
	}
	return s.tickCollisionCache.playerIDs
}
