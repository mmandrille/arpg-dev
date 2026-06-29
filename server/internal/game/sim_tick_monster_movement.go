package game

// advanceMonsterMovement runs full monster chase/movement for the active level.
func (s *Sim) advanceMonsterMovement(res *TickResult) {
	s.advanceMonsterMovementWithLOD(res, false)
}

// advanceMonsterMovementWithLOD optionally skips low-priority monsters when movement LOD applies.
func (s *Sim) advanceMonsterMovementWithLOD(res *TickResult, applyMovementLOD bool) {
	nav := s.activeNav()
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		monster := s.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		def, ok := s.rules.Monsters[monster.monsterDefID]
		if !ok || def.effectiveBehavior() != monsterBehaviorChase {
			continue
		}
		if monster.isBoss && monster.bossPhaseKind == "active" {
			continue
		}
		if applyMovementLOD && !s.monsterMovementHighPrecision(monster) && !s.monsterMovementLODAllowsTick(monster) {
			continue
		}
		if leader := s.eliteMinionLeader(monster); leader != nil {
			s.advanceEliteMinionMovement(monster, leader, def, res)
			continue
		}
		targetPlayer := s.nearestLivingPlayerForMonster(s.activeLevel(), monster)
		if targetPlayer == nil {
			continue
		}
		player := s.activeLevel().entities[targetPlayer.PlayerID]
		if player == nil {
			continue
		}
		s.usePlayer(targetPlayer)
		prevMode := monster.aiMode
		if monster.isBoss {
			monster.aiMode = monsterAIModeChase
		} else {
			s.updateMonsterAIMode(monster, player, def, prevMode, res)
		}
		if monster.aiMode == monsterAIModeIdle {
			continue
		}
		s.updateMonsterRangedMeleeEngagement(monster, player, def)
		goal, hasGoal := s.monsterMovementGoal(monster, player, def)
		if !hasGoal {
			continue
		}
		if distance(monster.pos, goal) <= nav.StopDistance && s.monsterInAttackRange(monster, player, def) {
			continue
		}
		s.moveMonsterToPoint(monster, def, goal, res)
	}
}
