package game

// PerfSnapshot is a cheap, coarse simulation shape summary for local perf logs.
type PerfSnapshot struct {
	Level         int
	Entities      int
	Players       int
	Monsters      int
	Companions    int
	Projectiles   int
	Loot          int
	Interactables int
	LiveMonsters  int
	Walls         int
}

// PerfSnapshot returns counts for the current active level.
func (s *Sim) PerfSnapshot() PerfSnapshot {
	level := s.activeLevel()
	out := PerfSnapshot{Level: s.currentLevel, Walls: len(s.walls)}
	if level != nil {
		out.Level = level.levelNum
		out.Walls = len(level.walls)
		for _, e := range level.entities {
			if e == nil {
				continue
			}
			out.Entities++
			switch e.kind {
			case playerEntity:
				out.Players++
			case monsterEntity:
				out.Monsters++
				if e.hp > 0 {
					out.LiveMonsters++
				}
			case companionEntity:
				out.Companions++
			case projectileEntity:
				out.Projectiles++
			case lootEntity:
				out.Loot++
			case interactableEntity:
				out.Interactables++
			}
		}
		return out
	}
	for _, e := range s.entities {
		if e == nil {
			continue
		}
		out.Entities++
	}
	return out
}
