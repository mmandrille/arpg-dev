package game

import "math"

func playerSpawnRoom(rules DungeonGenerationRules, margin float64) (dungeonRoom, bool) {
	spawn := rules.PlayerSpawn
	r := rules.RoomCorridorPCG
	width := r.RoomSizeMin.X
	height := r.RoomSizeMin.Y
	x0 := margin + rules.WallThickness
	y0 := spawn.Y - height*0.5
	if y0 < margin+rules.WallThickness {
		y0 = margin + rules.WallThickness
	}
	if y0+height > rules.FloorSize.Height-margin-rules.WallThickness {
		y0 = rules.FloorSize.Height - margin - rules.WallThickness - height
	}
	room := makeDungeonRoom(
		Vec2{X: x0, Y: y0},
		Vec2{X: x0 + width, Y: y0 + height},
		false,
		"rectangle",
		[roomShapeCellCount]bool{true, true, true, true, true, true, true, true, true},
	)
	if !pointInsideRoomInner(spawn, room, playerRadius+0.1) {
		return dungeonRoom{}, false
	}
	return room, true
}

func maxRoomsForFloor(rules DungeonGenerationRules) int {
	area := rules.FloorSize.Width * rules.FloorSize.Height
	avgRoom := rules.RoomCorridorPCG.RoomSizeMin.X * rules.RoomCorridorPCG.RoomSizeMin.Y
	if avgRoom <= 0 {
		return rules.RoomCorridorPCG.RoomCount.Max
	}
	cap := int(math.Floor(area / (avgRoom * 2.5)))
	if cap < rules.RoomCorridorPCG.RoomCount.Min {
		cap = rules.RoomCorridorPCG.RoomCount.Min
	}
	return minInt(cap, rules.RoomCorridorPCG.RoomCount.Max)
}

func pointInsideAnyRoom(p Vec2, rooms []dungeonRoom, margin float64) bool {
	for _, room := range rooms {
		if pointInsideRoomInner(p, room, margin) {
			return true
		}
	}
	return false
}
