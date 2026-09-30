package game

// Input is a decoded client intent (or a server-authored stored input such as
// SystemLoadShedInputType or a SystemMember*InputType) applied to a specific tick.
type Input struct {
	MessageID                    string
	CorrelationID                string
	Sequence                     int64
	ActorPlayerID                uint64
	Type                         string
	Move                         *MoveIntent
	MoveTo                       *MoveToIntent
	DirectionalAttack            *DirectionalAttackIntent
	Action                       *ActionIntent
	Descend                      *DescendIntent
	Ascend                       *AscendIntent
	Teleport                     *TeleportIntent
	Equip                        *EquipIntent
	Unequip                      *UnequipIntent
	SwapWeaponSet                *SwapWeaponSetIntent
	Drop                         *DropIntent
	Use                          *UseIntent
	AssignHotbar                 *AssignHotbarIntent
	UseHotbar                    *UseHotbarIntent
	AllocateStat                 *AllocateStatIntent
	AllocateSkillPoint           *AllocateSkillPointIntent
	CastSkill                    *CastSkillIntent
	ChannelSkill                 *ChannelSkillIntent
	SetSkillBindings             *SetSkillBindingsIntent
	CompanionCommand             *CompanionCommandIntent
	ShopBuy                      *ShopBuyIntent
	ShopSell                     *ShopSellIntent
	ShopReroll                   *ShopRerollIntent
	QuestStewardPick             *QuestStewardPickIntent
	BishopRespec                 *BishopRespecIntent
	BishopReviveAll              *BishopReviveAllIntent
	BishopDebugLevel             *BishopDebugLevelIntent
	BishopDebugSkill             *BishopDebugSkillPointIntent
	BishopDebugStat              *BishopDebugStatPointIntent
	BishopDebugLootCatalog       *BishopDebugLootCatalogIntent
	BishopDebugLootSourceCatalog *BishopDebugLootSourceCatalogIntent
	BishopDebugForceLoot         *BishopDebugForceLootIntent
	StashDepositItem             *StashDepositItemIntent
	StashWithdrawItem            *StashWithdrawItemIntent
	StashDepositGold             *StashDepositGoldIntent
	StashWithdrawGold            *StashWithdrawGoldIntent
	ResourceBagDepositItem       *ResourceBagDepositItemIntent
	ResourceBagDepositStashItem  *ResourceBagDepositStashItemIntent
	ResourceBagWithdrawItem      *ResourceBagWithdrawItemIntent
	StashDepositResourceBagItem  *StashDepositResourceBagItemIntent
	CorpseWithdrawItem           *CorpseWithdrawItemIntent
	UniqueChestTakeItem          *UniqueChestTakeItemIntent
	DebugPlayerPos               *DebugPlayerPosIntent
	LoadShed                     *LoadShedDirective
	Member                       *MemberLifecycle
	StashSync                    *AccountStashSync
}
