package game

func (s *Sim) monsterMovementLODActive() bool {
	nav := s.activeNav()
	if nav.MonsterMovementLODUpdateIntervalTicks <= 1 {
		return false
	}
	if s.overloadDegraded() {
		return true
	}
	return s.activeLiveMonsterCount() >= nav.MonsterMovementLODMinLiveMonsters
}

func (s *Sim) activeLiveMonsterCount() int {
	level := s.activeLevel()
	if level == nil {
		return 0
	}
	count := 0
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == monsterEntity && e.hp > 0 {
			count++
		}
	}
	return count
}

func (s *Sim) monsterMovementLODAllowsTick(monster *entity) bool {
	if monster == nil || monster.kind != monsterEntity || !s.monsterMovementLODActive() {
		return true
	}
	if s.monsterMovementHighPrecision(monster) {
		return true
	}
	if s.overloadDegraded() {
		return false
	}
	interval := s.activeNav().MonsterMovementLODUpdateIntervalTicks
	return (int(s.tick)+int(monster.id%uint64(interval)))%interval == 0
}

func (s *Sim) monsterMovementHighPrecision(monster *entity) bool {
	if monster == nil || monster.isBoss || monster.monsterRarityID != "" || monster.monsterPackLeader {
		return true
	}
	nearDistance := s.activeNav().MonsterMovementLODNearDistance
	level := s.activeLevel()
	if level == nil {
		return false
	}
	for _, id := range sortedEntityIDs(level.entities) {
		player := level.entities[id]
		if player == nil || player.kind != playerEntity || player.hp <= 0 {
			continue
		}
		if distance(monster.pos, player.pos) <= nearDistance {
			return true
		}
	}
	return false
}
