package game

const (
	CompanionStanceAssist  = "assist"
	CompanionStanceDefend  = "defend"
	CompanionStancePassive = "passive"
)

type CompanionCommandIntent struct {
	Stance string
}

func normalizeCompanionStance(stance string) string {
	if stance == "" {
		return CompanionStanceAssist
	}
	return stance
}

func validCompanionStance(stance string) bool {
	switch stance {
	case CompanionStanceAssist, CompanionStanceDefend, CompanionStancePassive:
		return true
	default:
		return false
	}
}

func (e *entity) companionStanceOrDefault() string {
	if e == nil || e.kind != companionEntity {
		return ""
	}
	return normalizeCompanionStance(e.companionStance)
}

func (s *Sim) handleCompanionCommand(in Input, res *TickResult) {
	if in.CompanionCommand == nil {
		res.reject(in.MessageID, "invalid_payload")
		return
	}
	stance := normalizeCompanionStance(in.CompanionCommand.Stance)
	if !validCompanionStance(stance) {
		res.reject(in.MessageID, "invalid_stance")
		return
	}
	level := s.activeLevel()
	if level == nil {
		res.reject(in.MessageID, "invalid_level")
		return
	}
	owner := level.entities[s.playerID]
	if owner == nil || owner.kind != playerEntity || owner.hp <= 0 {
		res.reject(in.MessageID, "player_dead")
		return
	}

	affected := 0
	for _, id := range sortedEntityIDs(level.entities) {
		companion := level.entities[id]
		if companion == nil || companion.kind != companionEntity || companion.ownerID != owner.id || companion.hp <= 0 {
			continue
		}
		companion.companionStance = stance
		if stance == CompanionStancePassive {
			companion.targetID = 0
		} else if companion.targetID != 0 && !s.validCompanionTarget(companion, level.entities[companion.targetID]) {
			companion.targetID = 0
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(companion))})
		affected++
	}
	if affected == 0 {
		res.reject(in.MessageID, "no_companion")
		return
	}
	res.Events = append(res.Events, Event{
		EventType:     "companion_stance_changed",
		EntityID:      idStr(owner.id),
		CorrelationID: in.CorrelationID,
		Stance:        stance,
		Amount:        intPtr(affected),
	})
	res.ack(in.MessageID)
}
