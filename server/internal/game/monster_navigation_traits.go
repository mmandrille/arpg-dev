package game

const (
	monsterNavigationTraitGrounded = "grounded"
	monsterNavigationTraitFlying   = "flying"
)

func (d MonsterDef) effectiveNavigationTrait() string {
	if d.NavigationTrait == "" {
		return monsterNavigationTraitGrounded
	}

	return d.NavigationTrait
}

func (d MonsterDef) ignoresObstacleKind(kind string) bool {
	if d.effectiveNavigationTrait() != monsterNavigationTraitFlying {
		return false
	}

	switch kind {
	case obstacleKindWater, obstacleKindHole:
		return true
	default:
		return false
	}
}

func monsterObstacleBlocksMovement(w wallObstacle, def MonsterDef) bool {
	if !obstacleBlocksMovement(w) {
		return false
	}

	return !def.ignoresObstacleKind(w.obstacleKind())
}

func (s *Sim) monsterNavigationDef(excludeMonsterID uint64) MonsterDef {
	if excludeMonsterID == 0 {
		return MonsterDef{}
	}
	level := s.activeLevel()
	if level == nil {
		return MonsterDef{}
	}
	e := level.entities[excludeMonsterID]
	if e == nil || e.kind != monsterEntity {
		return MonsterDef{}
	}
	return s.rules.Monsters[e.monsterDefID]
}

func validMonsterNavigationTrait(trait string) bool {
	switch trait {
	case "", monsterNavigationTraitGrounded, monsterNavigationTraitFlying:
		return true
	default:
		return false
	}
}

func validObstacleKind(kind string) bool {
	switch kind {
	case "", obstacleKindWall, obstacleKindWood, obstacleKindWater, obstacleKindHole, obstacleKindRock, obstacleKindColumn, obstacleKindRubble:
		return true
	default:
		return false
	}
}
