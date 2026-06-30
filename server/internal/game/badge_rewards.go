package game

const badgeRewardEventType = "badge_rewarded"

// BadgeRewardRule is one depth-scaled account-wallet resource reward.
type BadgeRewardRule struct {
	ResourceItemDefID     string `json:"resource_item_def_id"`
	UnlockDepth           int    `json:"unlock_depth"`
	BaseChancePercent     int    `json:"base_chance_percent"`
	ChancePerDepthPercent int    `json:"chance_per_depth_percent"`
}

type badgeRewardSource struct {
	EntityID       string
	SourceEntityID string
	BossTemplateID string
	Service        string
}

func badgeRewardChancePercent(rule BadgeRewardRule, depth int) int {
	if depth < rule.UnlockDepth {
		return 0
	}
	chance := rule.BaseChancePercent + (depth-rule.UnlockDepth)*rule.ChancePerDepthPercent
	if chance > 100 {
		return 100
	}
	if chance < 0 {
		return 0
	}
	return chance
}

func (r *Rules) isBadgeRewardResourceItem(itemDefID string) bool {
	if r == nil || itemDefID == "" {
		return false
	}
	if itemDefID == r.MainConfig.Gameplay.ItemUpgradeResourceID {
		return false
	}
	for _, rule := range r.MainConfig.Gameplay.BadgeRewardRules {
		if rule.ResourceItemDefID == itemDefID {
			return true
		}
	}
	return false
}

func (s *Sim) grantQuestTurnInBadgeRewards(giver *entity, corr string, res *TickResult) {
	source := badgeRewardSource{EntityID: idStr(giver.id), Service: questTurnInService}
	s.grantBadgeRewardsForDepth(s.progression.DeepestDungeonDepth, s.playerID, source, corr, res, false)
}

func (s *Sim) grantBossBadgeRewards(monster *entity, sourceID uint64, corr string, res *TickResult) {
	if monster == nil || !monster.isBoss {
		return
	}
	depth := absInt(s.activeLevel().levelNum)
	source := badgeRewardSource{
		EntityID:       idStr(monster.id),
		SourceEntityID: idStr(sourceID),
		BossTemplateID: monster.bossTemplateID,
	}
	s.grantBadgeRewardsForDepth(depth, s.badgeRewardPlayerIDForSource(sourceID), source, corr, res, true)
}

func (s *Sim) badgeRewardPlayerIDForSource(sourceID uint64) uint64 {
	if _, ok := s.players[sourceID]; ok {
		return sourceID
	}
	if source := s.activeLevel().entities[sourceID]; source != nil && source.kind == companionEntity && source.ownerID != 0 {
		if _, ok := s.players[source.ownerID]; ok {
			return source.ownerID
		}
	}
	return s.playerID
}

func (s *Sim) grantBadgeRewardsForDepth(depth int, playerID uint64, source badgeRewardSource, corr string, res *TickResult, bossSource bool) int {
	if depth <= 0 || s.rules == nil {
		return 0
	}
	granted := 0
	for _, rule := range s.rules.MainConfig.Gameplay.BadgeRewardRules {
		if bossSource && rule.ResourceItemDefID == s.rules.MainConfig.Gameplay.ItemUpgradeResourceID {
			continue
		}
		chance := badgeRewardChancePercent(rule, depth)
		if chance <= 0 || s.rng.IntN(100) >= chance {
			continue
		}
		if rule.ResourceItemDefID == UpgradeShardItemDefID {
			if !s.grantUpgradeShardItemForPlayer(playerID, depth, source, corr, res) {
				continue
			}
		} else if _, ok := s.grantWalletResourceForPlayer(playerID, rule.ResourceItemDefID, 1, res); !ok {
			continue
		}
		res.Events = append(res.Events, Event{
			EventType:      badgeRewardEventType,
			EntityID:       source.EntityID,
			SourceEntityID: source.SourceEntityID,
			BossTemplateID: source.BossTemplateID,
			CorrelationID:  corr,
			Service:        source.Service,
			Level:          intPtr(depth),
			ResourceID:     rule.ResourceItemDefID,
			Amount:         intPtr(1),
		})
		granted++
	}
	return granted
}

func (s *Sim) grantUpgradeShardItemForPlayer(playerID uint64, depth int, source badgeRewardSource, corr string, res *TickResult) bool {
	ps := s.players[playerID]
	if ps == nil {
		return false
	}
	current := s.players[s.playerID]
	if current != nil && current.PlayerID != playerID {
		s.savePlayer(current)
	}
	s.usePlayer(ps)

	level := RollItemLevel(s.rng, depth, s.rules.DungeonGeneration.ItemLevelTiers)
	if s.bagOccupancyCount()+1 > s.inventoryCapacity() {
		if current != nil {
			s.usePlayer(current)
		}
		return false
	}

	item := &invItem{
		instanceID:  s.alloc(),
		itemDefID:   UpgradeShardItemDefID,
		rollPayload: NewUpgradeShardRollPayload(level),
	}
	s.inventory = append(s.inventory, item)
	res.Changes = append(res.Changes, Change{Op: OpInventoryAdd, Item: ptrItemView(s.itemView(item))})
	res.Events = append(res.Events, Event{
		EventType:      "item_picked_up",
		EntityID:       idStr(playerID),
		CorrelationID:  corr,
		ItemInstanceID: idStr(item.instanceID),
	})
	s.savePlayer(ps)
	if current != nil && current.PlayerID != playerID {
		s.usePlayer(current)
	}

	return true
}

func (s *Sim) grantWalletResourceForPlayer(playerID uint64, resourceID string, amount int, res *TickResult) (int, bool) {
	if resourceID == "" || amount <= 0 {
		return 0, false
	}
	ps := s.players[playerID]
	if ps == nil {
		return 0, false
	}
	current := s.players[s.playerID]
	if current != nil && current.PlayerID == playerID {
		if s.resourceWallet == nil {
			s.resourceWallet = make(map[string]int)
		}
		s.resourceWallet[resourceID] += amount
		balance := s.resourceWallet[resourceID]
		res.Changes = append(res.Changes, Change{
			Op:             OpResourceWalletUpdate,
			OwnerPlayerID:  playerID,
			ResourceID:     resourceID,
			ResourceAmount: intPtr(balance),
		})
		s.savePlayer(current)
		return balance, true
	}
	if current != nil {
		s.savePlayer(current)
	}
	s.usePlayer(ps)
	if s.resourceWallet == nil {
		s.resourceWallet = make(map[string]int)
	}
	s.resourceWallet[resourceID] += amount
	balance := s.resourceWallet[resourceID]
	res.Changes = append(res.Changes, Change{
		Op:             OpResourceWalletUpdate,
		OwnerPlayerID:  playerID,
		ResourceID:     resourceID,
		ResourceAmount: intPtr(balance),
	})
	s.savePlayer(ps)
	if current != nil {
		s.usePlayer(current)
	}
	return balance, true
}
