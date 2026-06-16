package game

import (
	"fmt"
	"sort"
	"strconv"
)

// AddGuestPlayer creates another connected player in level 0 town. It is the
// deterministic co-op join path; player-vs-player collision remains disabled.
func (s *Sim) AddGuestPlayer(accountID, characterID, displayName string, progression CharacterProgressionState) (uint64, error) {
	if displayName == "" {
		displayName = "Guest"
	}
	level, err := s.ensureTravelLevel(townLevel)
	if err != nil {
		return 0, err
	}
	spawn := s.findTownSpawnPosition(level)
	progression = s.rules.normalizeProgressionState(progression)
	equipped := newEquippedMap()
	weaponSets := newWeaponSetMaps()
	hotbar := make([]uint64, maxHotbarCapacity)
	discovered := map[int]bool{townLevel: true}
	cooldowns := make(map[string]skillCooldownState)
	effects := make(map[string]skillEffectState)
	shopStock := make(map[string]*shopStockState)
	stashItems := []*stashItem{}
	stashCapacity := defaultStashCapacity
	character := progression
	gold := progression.Gold
	s.equipped = equipped
	s.weaponSets = weaponSets
	s.activeWeaponSet = defaultWeaponSet
	s.hotbar = hotbar
	s.discoveredTeleporters = discovered
	s.progression = character
	s.skillCooldowns = cooldowns
	s.skillEffects = effects
	s.shopStock = shopStock
	s.gold = gold
	s.stashItems = stashItems
	s.stashGold = 0
	s.stashCapacity = stashCapacity
	maxHP := s.currentMaxHP()
	maxMana := s.currentMaxMana()
	player := &entity{
		kind:        playerEntity,
		pos:         spawn,
		hp:          maxHP,
		maxHP:       maxHP,
		mana:        maxMana,
		maxMana:     maxMana,
		characterID: characterID,
		displayName: displayName,
	}
	player.id = s.alloc()
	level.entities[player.id] = player
	s.players[player.id] = &playerState{
		PlayerID:              player.id,
		AccountID:             accountID,
		CharacterID:           characterID,
		DisplayName:           displayName,
		Role:                  "guest",
		Connected:             true,
		CurrentLevel:          townLevel,
		Equipped:              equipped,
		WeaponSets:            cloneWeaponSetMaps(weaponSets),
		ActiveWeaponSet:       defaultWeaponSet,
		Hotbar:                hotbar,
		DiscoveredTeleporters: discovered,
		Progression:           character,
		SkillCooldowns:        cooldowns,
		SkillEffects:          effects,
		PoisonDots:            make(map[uint64]poisonDotState),
		UniqueBurnDots:        make(map[string]uniqueBurnDotState),
		UniqueExecutionMarks:  make(map[uint64]uniqueExecutionMarkState),
		UniqueHungerStacks:    make(map[uint64]uniqueHungerStackState),
		UniqueAshenReprisals:  make(map[uint64]uniqueAshenReprisalState),
		ShopStock:             shopStock,
		Gold:                  gold,
		StashItems:            stashItems,
		StashGold:             0,
		StashCapacity:         stashCapacity,
	}
	s.usePlayer(s.defaultPlayer())
	return player.id, nil
}

// SetPlayerMetadata fills party/player metadata for an existing player, usually
// the host player created with the Sim.
func (s *Sim) SetPlayerMetadata(playerID uint64, accountID, characterID, displayName, role string) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	if displayName == "" {
		displayName = ps.DisplayName
	}
	if role == "" {
		role = ps.Role
	}
	ps.AccountID = accountID
	ps.CharacterID = characterID
	ps.DisplayName = displayName
	ps.Role = role
	if e := s.levels[ps.CurrentLevel].entities[playerID]; e != nil {
		e.characterID = characterID
		e.displayName = displayName
	}
}

func (s *Sim) SetPlayerConnected(playerID uint64, connected bool) {
	if ps := s.players[playerID]; ps != nil {
		ps.Connected = connected
	}
}

func (s *Sim) DefaultPlayerID() uint64 {
	if ps := s.defaultPlayer(); ps != nil {
		return ps.PlayerID
	}
	return 0
}

func (s *Sim) PlayerCurrentLevel(playerID uint64) (int, bool) {
	ps := s.players[playerID]
	if ps == nil {
		return 0, false
	}
	return ps.CurrentLevel, true
}

func (s *Sim) PlayerConnected(playerID uint64) bool {
	ps := s.players[playerID]
	return ps != nil && ps.Connected
}

func (s *Sim) PlayerIDs() []uint64 {
	return sortedPlayerIDs(s.players)
}

func (s *Sim) PlayerIDForCharacter(characterID string) (uint64, bool) {
	if characterID == "" {
		return 0, false
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps != nil && ps.CharacterID == characterID {
			return playerID, true
		}
	}
	return 0, false
}

func ParseEntityID(id string) (uint64, bool) {
	n, err := strconv.ParseUint(id, 10, 64)
	return n, err == nil
}

func (s *Sim) RemovePlayerEntity(playerID uint64) {
	ps := s.players[playerID]
	if ps == nil {
		return
	}
	if level := s.levels[ps.CurrentLevel]; level != nil {
		delete(level.entities, playerID)
	}
	ps.Connected = false
	if s.playerID == playerID {
		s.usePlayer(s.defaultPlayer())
	}
}

func (s *Sim) RespawnPlayerInTown(playerID uint64) error {
	ps := s.players[playerID]
	if ps == nil {
		return fmt.Errorf("game: unknown player %d", playerID)
	}
	level, err := s.ensureTravelLevel(townLevel)
	if err != nil {
		return err
	}
	for _, lvl := range s.levels {
		delete(lvl.entities, playerID)
	}
	s.usePlayer(ps)
	maxHP := s.currentMaxHP()
	player := &entity{
		id:          playerID,
		kind:        playerEntity,
		pos:         s.findTownSpawnPosition(level),
		hp:          maxHP,
		maxHP:       maxHP,
		mana:        s.currentMaxMana(),
		maxMana:     s.currentMaxMana(),
		characterID: ps.CharacterID,
		displayName: ps.DisplayName,
	}
	level.entities[playerID] = player
	s.currentLevel = townLevel
	ps.CurrentLevel = townLevel
	ps.Connected = true
	s.savePlayer(ps)
	s.usePlayer(s.defaultPlayer())
	return nil
}

func (s *Sim) findTownSpawnPosition(level *LevelState) Vec2 {
	preferred := s.preferredTownSpawnPosition(level)
	if !s.spawnPositionBlocked(level, preferred) {
		return preferred
	}
	nav := s.navigationForLevel(level)
	step := nav.CellSize
	if step <= 0 {
		step = 1.0
	}
	for ring := 1; ring <= 8; ring++ {
		for dy := -ring; dy <= ring; dy++ {
			for dx := -ring; dx <= ring; dx++ {
				if absInt(dx) != ring && absInt(dy) != ring {
					continue
				}
				candidate := Vec2{
					X: preferred.X + float64(dx)*step,
					Y: preferred.Y + float64(dy)*step,
				}
				if !s.positionInNavigationBounds(nav, candidate) {
					continue
				}
				if !s.spawnPositionBlocked(level, candidate) {
					return candidate
				}
			}
		}
	}
	return preferred
}

func (s *Sim) preferredTownSpawnPosition(level *LevelState) Vec2 {
	if host := s.players[s.playerID]; host != nil {
		if lvl := s.levels[host.CurrentLevel]; lvl != nil && host.CurrentLevel == level.levelNum {
			if e := lvl.entities[host.PlayerID]; e != nil {
				return e.pos
			}
		}
	}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == playerEntity {
			return e.pos
		}
	}
	return Vec2{X: 4, Y: 10}
}

func (s *Sim) spawnPositionBlocked(level *LevelState, pos Vec2) bool {
	if level == nil {
		return true
	}
	for _, wall := range level.walls {
		if circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e == nil {
			continue
		}
		switch e.kind {
		case playerEntity:
			if e.hp > 0 && circlesOverlap(pos, playerRadius, e.pos, playerRadius) {
				return true
			}
		case monsterEntity:
			if e.hp > 0 && circlesOverlap(pos, playerRadius, e.pos, monsterRadius) {
				return true
			}
		case interactableEntity:
			if e.state == interactableClosed {
				if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
					if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (s *Sim) navigationForLevel(level *LevelState) NavigationRules {
	if level != nil && level.nav != nil {
		return *level.nav
	}
	return s.rules.Navigation
}

func (s *Sim) positionInNavigationBounds(nav NavigationRules, pos Vec2) bool {
	cell := worldToGrid(nav, pos)
	return cellInBounds(nav, cell)
}

func (s *Sim) defaultPlayer() *playerState {
	if ps := s.players[s.playerID]; ps != nil {
		return ps
	}
	for _, id := range sortedPlayerIDs(s.players) {
		return s.players[id]
	}
	return nil
}

func (s *Sim) playerForInput(in Input) *playerState {
	if in.ActorPlayerID != 0 {
		return s.players[in.ActorPlayerID]
	}
	return s.defaultPlayer()
}

func (s *Sim) usePlayer(ps *playerState) {
	if ps == nil {
		return
	}
	s.playerID = ps.PlayerID
	s.currentLevel = ps.CurrentLevel
	s.inventory = ps.Inventory
	s.equipped = ps.Equipped
	s.weaponSets = cloneWeaponSetMaps(ps.WeaponSets)
	s.activeWeaponSet = normalizeWeaponSetIndex(ps.ActiveWeaponSet)
	s.hotbar = ps.Hotbar
	s.discoveredTeleporters = ps.DiscoveredTeleporters
	s.progression = ps.Progression
	s.skillCooldowns = ps.SkillCooldowns
	if s.skillCooldowns == nil {
		s.skillCooldowns = make(map[string]skillCooldownState)
	}
	s.skillEffects = ps.SkillEffects
	if s.skillEffects == nil {
		s.skillEffects = make(map[string]skillEffectState)
	}
	s.poisonDots = ps.PoisonDots
	if s.poisonDots == nil {
		s.poisonDots = make(map[uint64]poisonDotState)
	}
	s.rogueMarks = ps.RogueMarks
	if s.rogueMarks == nil {
		s.rogueMarks = make(map[uint64]rogueMarkState)
	}
	s.uniqueBurnDots = ps.UniqueBurnDots
	if s.uniqueBurnDots == nil {
		s.uniqueBurnDots = make(map[string]uniqueBurnDotState)
	}
	s.uniqueExecutionMarks = ps.UniqueExecutionMarks
	if s.uniqueExecutionMarks == nil {
		s.uniqueExecutionMarks = make(map[uint64]uniqueExecutionMarkState)
	}
	s.uniqueHungerStacks = ps.UniqueHungerStacks
	if s.uniqueHungerStacks == nil {
		s.uniqueHungerStacks = make(map[uint64]uniqueHungerStackState)
	}
	s.uniqueAshenReprisals = ps.UniqueAshenReprisals
	if s.uniqueAshenReprisals == nil {
		s.uniqueAshenReprisals = make(map[uint64]uniqueAshenReprisalState)
	}
	s.usePlayerUniquePilgrimMomentum(ps)
	s.uniqueChests = restoreUniqueChestItems(ps.UniqueChestItems)
	s.skillFunctionKeys = normalizeSkillFunctionKeys(ps.SkillFunctionKeys)
	s.rightClickSkillID = ps.RightClickSkillID
	s.shopStock = ps.ShopStock
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	s.gold = ps.Gold
	s.stashItems = ps.StashItems
	s.stashGold = ps.StashGold
	s.stashCapacity = ps.StashCapacity
	if s.stashCapacity <= 0 {
		s.stashCapacity = defaultStashCapacity
	}
	s.hpRegenCarry = ps.HPRegenCarry
	s.manaRegenCarry = ps.ManaRegenCarry
	s.nextBasicAttackTick = ps.NextBasicAttackTick
	s.nextOffHandAttackTick = ps.NextOffHandAttackTick
	level := s.activeLevel()
	level.move = ps.Move
	level.autoNav = ps.AutoNav
	level.activeChannel = ps.ActiveChannel
	s.syncCompatibilityFields()
}

func (s *Sim) savePlayer(ps *playerState) {
	if ps == nil {
		return
	}
	s.syncEquippedHandsToActiveWeaponSet()
	ps.CurrentLevel = s.currentLevel
	ps.Inventory = s.inventory
	ps.Equipped = s.equipped
	ps.WeaponSets = cloneWeaponSetMaps(s.weaponSets)
	ps.ActiveWeaponSet = s.activeWeaponSet
	ps.Hotbar = s.hotbar
	ps.DiscoveredTeleporters = s.discoveredTeleporters
	ps.Progression = s.progression
	ps.SkillCooldowns = s.skillCooldowns
	ps.SkillEffects = s.skillEffects
	ps.PoisonDots = s.poisonDots
	ps.RogueMarks = s.rogueMarks
	ps.UniqueBurnDots = s.uniqueBurnDots
	ps.UniqueExecutionMarks = s.uniqueExecutionMarks
	ps.UniqueHungerStacks = s.uniqueHungerStacks
	ps.UniqueAshenReprisals = s.uniqueAshenReprisals
	ps.UniquePilgrimMomentum = s.uniquePilgrimMomentum
	ps.UniqueChestItems = cloneUniqueChestItems(s.uniqueChests)
	ps.SkillFunctionKeys = normalizeSkillFunctionKeys(s.skillFunctionKeys)
	ps.RightClickSkillID = s.rightClickSkillID
	ps.ShopStock = s.shopStock
	ps.Gold = s.gold
	ps.StashItems = s.stashItems
	ps.StashGold = s.stashGold
	ps.StashCapacity = s.stashCapacity
	ps.HPRegenCarry = s.hpRegenCarry
	ps.ManaRegenCarry = s.manaRegenCarry
	ps.NextBasicAttackTick = s.nextBasicAttackTick
	ps.NextOffHandAttackTick = s.nextOffHandAttackTick
	if level := s.levels[ps.CurrentLevel]; level != nil {
		ps.Move = level.move
		ps.AutoNav = level.autoNav
		ps.ActiveChannel = level.activeChannel
	}
}

func sortedPlayerIDs(players map[uint64]*playerState) []uint64 {
	ids := make([]uint64, 0, len(players))
	for id := range players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
