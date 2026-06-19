package game

type bishopResourceCost struct {
	ResourceID string
	Count      int
}

func (s *Sim) bishopRespecResourceCost() bishopResourceCost {
	if s.rules == nil {
		return bishopResourceCost{}
	}
	gameplay := s.rules.MainConfig.Gameplay
	return bishopResourceCost{ResourceID: gameplay.BishopRespecResourceID, Count: gameplay.BishopRespecResourceCost}
}

func (s *Sim) bishopReviveResourceCost() bishopResourceCost {
	if s.rules == nil {
		return bishopResourceCost{}
	}
	gameplay := s.rules.MainConfig.Gameplay
	return bishopResourceCost{ResourceID: gameplay.BishopReviveResourceID, Count: gameplay.BishopReviveResourceCost}
}

func (s *Sim) bishopResourceBalance(cost bishopResourceCost) int {
	if cost.ResourceID == "" || s.resourceWallet == nil {
		return 0
	}
	return maxInt(0, s.resourceWallet[cost.ResourceID])
}

func (s *Sim) canPayBishopResourceCost(cost bishopResourceCost) bool {
	return cost.Count <= 0 || s.bishopResourceBalance(cost) >= cost.Count
}

func (s *Sim) consumeBishopResourceCost(cost bishopResourceCost, res *TickResult) (int, bool) {
	if cost.Count <= 0 {
		return s.bishopResourceBalance(cost), true
	}
	if !s.canPayBishopResourceCost(cost) {
		return s.bishopResourceBalance(cost), false
	}
	s.resourceWallet[cost.ResourceID] -= cost.Count
	balance := maxInt(0, s.resourceWallet[cost.ResourceID])
	if balance == 0 {
		delete(s.resourceWallet, cost.ResourceID)
	}
	res.Changes = append(res.Changes, Change{
		Op:             OpResourceWalletUpdate,
		OwnerPlayerID:  s.playerID,
		ResourceID:     cost.ResourceID,
		ResourceAmount: intPtr(balance),
	})
	s.savePlayer(s.defaultPlayer())
	return balance, true
}
