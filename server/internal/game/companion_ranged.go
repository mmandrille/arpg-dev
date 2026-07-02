package game

func (s *Sim) companionEffectiveReach(companion *entity, def MonsterDef) float64 {
	if companion == nil {
		return s.monsterAttackReach(def)
	}
	if companion.companionAttackMode == attackModeRanged && companion.companionAttackRange > 0 {
		return companion.companionAttackRange
	}
	if companion.companionAttackReach > 0 {
		return companion.companionAttackReach
	}

	return s.monsterAttackReach(def)
}

func (s *Sim) fireCompanionProjectile(companion *entity, target *entity, damageRange DamageRange, res *TickResult) {
	if companion == nil || target == nil {
		return
	}
	dir := normalize(Vec2{X: target.pos.X - companion.pos.X, Y: target.pos.Y - companion.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		dir = Vec2{X: 1}
	}
	projectileDefID := companion.companionProjectileDefID
	if projectileDefID == "" {
		projectileDefID = "player_arrow"
	}
	projectile := &entity{
		kind:            projectileEntity,
		pos:             companion.pos,
		ownerID:         companion.id,
		targetID:        target.id,
		projectileDefID: projectileDefID,
		dir:             dir,
		speed:           companion.companionProjectileSpeed,
		maxDistance:     companion.companionAttackRange,
		damageRange:     damageRange,
		spawnTick:       s.tick,
	}
	projectile.id = s.alloc()
	s.activeLevel().entities[projectile.id] = projectile
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(projectile))})
}
