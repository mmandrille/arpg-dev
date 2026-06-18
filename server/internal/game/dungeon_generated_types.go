package game

type generatedDungeonLevel struct {
	levelNum    int
	walls       []wallObstacle
	stairs      []generatedStair
	teleporters []generatedTeleporter
	chests      []generatedChest
	doors       []generatedDoor
	monsters    []generatedMonster
	loot        []generatedLoot
}

type generatedStair struct {
	defID string
	pos   Vec2
	state string
}

type generatedTeleporter struct {
	defID string
	pos   Vec2
	state string
}

type generatedChest struct {
	defID          string
	lootTable      string
	pos            Vec2
	questReward    bool
	eliteObjective bool
}

type generatedDoor struct {
	defID string
	pos   Vec2
	state string
}

type generatedLoot struct {
	itemDefID string
	pos       Vec2
}

type generatedMonster struct {
	defID        string
	packID       string
	packLeader   bool
	rarityID     string
	bossTemplate string
	isBoss       bool
	visualModel  string
	visualTint   string
	visualScale  float64
	lootTable    string
	pos          Vec2
	maxHP        int
	attackDamage *DamageRange
	xpReward     int
}
