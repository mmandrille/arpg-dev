package game

import (
	"fmt"
	"math"
)

const (
	gravePactEffectID          = "grave_pact"
	bloodPriceEffectID         = "blood_price"
	pilgrimsMomentumEffectID   = "pilgrims_momentum"
	lanternOfTheFallenEffectID = "lantern_of_the_fallen"
)

type uniquePilgrimMomentumState struct {
	MoveTicks int
	EndsTick  uint64
	Charged   bool
}

func cloneUniquePilgrimMomentum(in map[uint64]uniquePilgrimMomentumState) map[uint64]uniquePilgrimMomentumState {
	if len(in) == 0 {
		return make(map[uint64]uniquePilgrimMomentumState)
	}
	out := make(map[uint64]uniquePilgrimMomentumState, len(in))
	for key, state := range in { //nolint:determinism — pure map clone, output is a map
		out[key] = state
	}
	return out
}

func (s *Sim) usePlayerUniquePilgrimMomentum(ps *playerState) {
	s.uniquePilgrimMomentum = ps.UniquePilgrimMomentum
	if s.uniquePilgrimMomentum == nil {
		s.uniquePilgrimMomentum = make(map[uint64]uniquePilgrimMomentumState)
	}
}

func (s *Sim) tryBloodPriceForSkill(player *entity, manaCost int, corr string, res *TickResult) bool {
	if player == nil || manaCost <= 0 || player.mana >= manaCost {
		return true
	}
	for _, effectID := range s.equippedUniqueEffectIDs(player.id) {
		if effectID != bloodPriceEffectID {
			continue
		}
		def, ok := s.liveUniqueEffect(effectID, "on_skill_resource_shortfall")
		if !ok || uniqueEffectStringParam(def, "resource_shortfall", "mana") != "mana" || uniqueEffectStringParam(def, "fallback_resource", "hp") != "hp" {
			continue
		}
		missingMana := manaCost - player.mana
		hpCost := missingMana * uniqueEffectIntParam(def, "hp_cost_per_missing_mana", 1)
		minHP := uniqueEffectIntParam(def, "minimum_remaining_hp", 1)
		if hpCost <= 0 || player.hp-hpCost < minHP {
			continue
		}
		player.hp -= hpCost
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(player.id),
			SourceEntityID: idStr(player.id),
			TargetEntityID: idStr(player.id),
			CorrelationID:  corr,
			SkillID:        def.ID,
			Amount:         intPtr(hpCost),
		})
		return true
	}
	return false
}

func (s *Sim) updatePilgrimMomentumMovement(player *entity, moved bool, res *TickResult) {
	if player == nil || player.kind != playerEntity {
		return
	}
	if _, ok := s.liveUniqueEffect(pilgrimsMomentumEffectID, "on_continuous_movement_attack"); !ok || !containsStringValue(s.equippedUniqueEffectIDs(player.id), pilgrimsMomentumEffectID) {
		delete(s.uniquePilgrimMomentum, player.id)
		return
	}
	state := s.uniquePilgrimMomentum[player.id]
	if !moved {
		if !state.Charged {
			delete(s.uniquePilgrimMomentum, player.id)
		}
		return
	}
	def := s.rules.UniqueEffects[pilgrimsMomentumEffectID]
	requiredTicks := uniqueEffectIntParam(def, "required_movement_seconds", 0) * 10
	expireTicks := uniqueEffectIntParam(def, "charge_expire_seconds", 0) * 10
	if requiredTicks <= 0 || expireTicks <= 0 {
		return
	}
	state.MoveTicks++
	if !state.Charged && state.MoveTicks >= requiredTicks {
		state.Charged = true
		state.EndsTick = s.tick + uint64(expireTicks)
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(player.id),
			SourceEntityID: idStr(player.id),
			TargetEntityID: idStr(player.id),
			SkillID:        pilgrimsMomentumEffectID,
			RemainingTicks: intPtr(expireTicks),
			TotalTicks:     intPtr(expireTicks),
		})
	}
	s.uniquePilgrimMomentum[player.id] = state
}

func (s *Sim) applyPilgrimMomentumBeforeHeroHit(target *entity, playerID uint64, damageRange DamageRange) DamageRange {
	state, ok := s.uniquePilgrimMomentum[playerID]
	if !ok || !state.Charged || s.tick >= state.EndsTick {
		delete(s.uniquePilgrimMomentum, playerID)
		return damageRange
	}
	def, ok := s.liveUniqueEffect(pilgrimsMomentumEffectID, "on_continuous_movement_attack")
	if !ok || target == nil || target.kind != monsterEntity {
		return damageRange
	}
	bonusPercent := uniqueEffectIntParam(def, "bonus_damage_percent", 0)
	damageRange.Min = applyPercentBonus(damageRange.Min, bonusPercent)
	damageRange.Max = applyPercentBonus(damageRange.Max, bonusPercent)
	return damageRange
}

func (s *Sim) triggerPilgrimMomentumAfterHeroHit(target *entity, playerID uint64, corr string, res *TickResult, outcome combatResolution) {
	state, ok := s.uniquePilgrimMomentum[playerID]
	if !ok || !state.Charged || s.tick >= state.EndsTick || outcome.Damage <= 0 {
		return
	}
	def, ok := s.liveUniqueEffect(pilgrimsMomentumEffectID, "on_continuous_movement_attack")
	if !ok {
		return
	}
	delete(s.uniquePilgrimMomentum, playerID)
	tiles := uniqueEffectFloatParam(def, "knockback_tiles", 0)
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || tiles <= 0 {
		return
	}
	player := s.activeLevel().entities[playerID]
	if player == nil {
		return
	}
	dir := normalize(Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		return
	}
	before := target.pos
	target.pos = s.resolveMonsterMovement(target, Vec2{X: dir.X * tiles, Y: dir.Y * tiles})
	if target.pos != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	}
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(playerID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        def.ID,
		Amount:         intPtr(int(math.Round(tiles))),
	})
}

func (s *Sim) triggerResourceUniqueEffectsOnMonsterKilled(monster *entity, sourceID uint64, corr string, res *TickResult) {
	player := s.activeLevel().entities[sourceID]
	if monster == nil || monster.kind != monsterEntity || player == nil || player.kind != playerEntity {
		return
	}
	for _, effectID := range s.equippedUniqueEffectIDs(sourceID) {
		switch effectID {
		case gravePactEffectID:
			if def, ok := s.liveUniqueEffect(effectID, "on_enemy_killed"); ok {
				s.tryGravePact(player, def, corr, res)
			}
		case lanternOfTheFallenEffectID:
			if def, ok := s.liveUniqueEffect(effectID, "on_enemy_killed"); ok {
				s.tryLanternOfTheFallen(monster, player, def, corr, res)
			}
		}
	}
}

func (s *Sim) tryGravePact(player *entity, def UniqueEffectDef, corr string, res *TickResult) {
	threshold := uniqueEffectIntParam(def, "hero_hp_percent_below", 0)
	if threshold <= 0 || player.hp >= percentOf(player.maxHP, threshold) {
		return
	}
	heal := percentOf(player.maxHP, uniqueEffectIntParam(def, "heal_percent_max_hp", 0))
	if minHeal := uniqueEffectIntParam(def, "minimum_heal", 0); heal < minHeal {
		heal = minHeal
	}
	s.healUniquePlayer(player, def.ID, heal, corr, res)
}

func (s *Sim) tryLanternOfTheFallen(monster *entity, source *entity, def UniqueEffectDef, corr string, res *TickResult) {
	radius := uniqueEffectFloatParam(def, "trigger_radius_tiles", 0)
	if radius <= 0 {
		return
	}
	target := s.lowestHealthPlayerNear(monster.pos, radius)
	if target == nil {
		return
	}
	heal := percentOf(target.maxHP, uniqueEffectIntParam(def, "heal_percent_max_hp", 0))
	s.healUniquePlayer(target, def.ID, heal, corr, res)
	durationTicks := uniqueEffectIntParam(def, "wisp_duration_seconds", 0) * 10
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(source.id),
		TargetEntityID: idStr(target.id),
		CorrelationID:  corr,
		SkillID:        def.ID,
		Amount:         intPtr(heal),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
}

func (s *Sim) lowestHealthPlayerNear(pos Vec2, radius float64) *entity {
	var best *entity
	bestRatio := math.Inf(1)
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		player := s.activeLevel().entities[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != s.currentLevel || player == nil || player.kind != playerEntity || player.hp <= 0 || player.hp >= player.maxHP {
			continue
		}
		if distance(pos, player.pos) > radius+meleeRangeEpsilon {
			continue
		}
		ratio := float64(player.hp) / float64(player.maxHP)
		if best == nil || ratio < bestRatio || (math.Abs(ratio-bestRatio) <= 1e-9 && player.id < best.id) {
			best = player
			bestRatio = ratio
		}
	}
	return best
}

func (s *Sim) healUniquePlayer(player *entity, skillID string, heal int, corr string, res *TickResult) {
	if player == nil || player.hp <= 0 || heal <= 0 {
		return
	}
	if missing := player.maxHP - player.hp; heal > missing {
		heal = missing
	}
	if heal <= 0 {
		return
	}
	player.hp += heal
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	res.Events = append(res.Events, Event{
		EventType:      "player_healed",
		EntityID:       idStr(player.id),
		TargetEntityID: idStr(player.id),
		CorrelationID:  corr,
		SkillID:        skillID,
		Heal:           intPtr(heal),
	})
}

func uniqueResourceEffectKey(effectID string, entityID uint64) string {
	return fmt.Sprintf("%s:%d", effectID, entityID)
}
