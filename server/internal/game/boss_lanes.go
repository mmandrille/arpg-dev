package game

import (
	"fmt"
	"math"
)

const (
	maxBossLaneWidth  = 10.0
	maxBossLaneLength = 30.0
)

// BossLaneView is both the locked strike frame and its current presentation stage.
// Its origin and axes are world-space and never follow later boss movement.
type BossLaneView struct {
	Origin        Vec2    `json:"origin"`
	Forward       Vec2    `json:"forward"`
	Right         Vec2    `json:"right"`
	Count         int     `json:"count"`
	Width         float64 `json:"width"`
	Length        float64 `json:"length"`
	SafeColor     string  `json:"safe_color"`
	DangerColor   string  `json:"danger_color"`
	SafeIntensity float64 `json:"safe_intensity"`
	SafeIndex     int     `json:"safe_index"`
	StageIndex    int     `json:"stage_index"`
	Intensity     float64 `json:"intensity"`
	DangerLanes   []int   `json:"danger_lanes"`
}

func validateBossLanes(patternID string, pattern BossPatternDef) error {
	lanes := pattern.Lanes
	if lanes == nil {
		for i, phase := range pattern.Phases {
			if phase.HitShape == "lanes" || phase.Shape == "lanes" || phase.TelegraphType == "lanes" || phase.LaneIntensity != 0 {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: lane phase requires lanes", patternID, i)
			}
		}
		return nil
	}
	if lanes.Count < 3 || lanes.Count > 9 || lanes.Width <= 2*playerRadius || lanes.Width > maxBossLaneWidth ||
		lanes.Length <= 2*playerRadius || lanes.Length > maxBossLaneLength ||
		lanes.SafeColor == "" || lanes.DangerColor == "" || lanes.SafeIntensity <= 0 || lanes.SafeIntensity > 1 || math.IsNaN(lanes.SafeIntensity) || math.IsInf(lanes.SafeIntensity, 0) ||
		math.IsNaN(lanes.Width) || math.IsInf(lanes.Width, 0) || math.IsNaN(lanes.Length) || math.IsInf(lanes.Length, 0) {
		return fmt.Errorf("game: invalid rules boss_patterns.%s.lanes: count, width, or length cannot fit a safe corridor", patternID)
	}
	if len(lanes.SafeSequence) == 0 || len(lanes.SafeSequence) > 16 {
		return fmt.Errorf("game: invalid rules boss_patterns.%s.lanes.safe_sequence: required and bounded", patternID)
	}
	if len(lanes.AimOffsetsDegrees) == 0 || len(lanes.AimOffsetsDegrees) > 8 {
		return fmt.Errorf("game: invalid rules boss_patterns.%s.lanes.aim_offsets_degrees: required and bounded", patternID)
	}
	for _, angle := range lanes.AimOffsetsDegrees {
		if angle < -180 || angle > 180 || math.IsNaN(angle) || math.IsInf(angle, 0) {
			return fmt.Errorf("game: invalid rules boss_patterns.%s.lanes.aim_offsets_degrees: invalid angle", patternID)
		}
	}
	for _, index := range lanes.SafeSequence {
		if index < 0 || index >= lanes.Count {
			return fmt.Errorf("game: invalid rules boss_patterns.%s.lanes.safe_sequence: invalid index", patternID)
		}
	}
	telegraphs, actives, recoveries := 0, 0, 0
	for i, phase := range pattern.Phases {
		switch phase.Kind {
		case "telegraph":
			if actives != 0 {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: warning after strike", patternID, i)
			}
			telegraphs++
			if phase.TelegraphType != "lanes" || phase.HitShape != "lanes" || phase.LaneIntensity <= 0 || phase.LaneIntensity > 1 || math.IsNaN(phase.LaneIntensity) || math.IsInf(phase.LaneIntensity, 0) {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: inconsistent lane warning", patternID, i)
			}
		case "active":
			if telegraphs < 2 || actives != 0 {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: lane strike before full warning", patternID, i)
			}
			actives++
			if phase.Shape != "lanes" || phase.Damage == nil {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: lane strike must match warning", patternID, i)
			}
		case "recovery":
			recoveries++
			if actives != 1 || recoveries != 1 {
				return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: recovery before strike", patternID, i)
			}
		default:
			return fmt.Errorf("game: invalid rules boss_patterns.%s.phases[%d]: invalid lane phase", patternID, i)
		}
	}
	if telegraphs < 2 || actives != 1 || recoveries != 1 {
		return fmt.Errorf("game: invalid rules boss_patterns.%s.lanes: staged warning and one strike required", patternID)
	}
	return nil
}

func bossLaneFrame(origin, forward Vec2, config BossLanesDef, safeIndex int) *BossLaneView {
	if distance(Vec2{}, forward) <= meleeRangeEpsilon {
		forward = Vec2{X: 1}
	}
	right := Vec2{X: forward.Y, Y: -forward.X}
	danger := make([]int, 0, config.Count-1)
	for i := 0; i < config.Count; i++ {
		if i != safeIndex {
			danger = append(danger, i)
		}
	}
	return &BossLaneView{Origin: origin, Forward: forward, Right: right, Count: config.Count,
		Width: config.Width, Length: config.Length, SafeColor: config.SafeColor,
		DangerColor: config.DangerColor, SafeIntensity: config.SafeIntensity,
		SafeIndex: safeIndex, DangerLanes: danger}
}

func (s *Sim) prepareBossLane(boss *entity, runtime bossPhaseRuntime) bool {
	if s == nil || s.rules == nil || boss == nil {
		return false
	}
	pattern, ok := s.rules.BossPatterns[runtime.patternID]
	if !ok || pattern.Lanes == nil || len(pattern.Lanes.SafeSequence) == 0 {
		return false
	}
	s.captureBossPhaseAim(boss, runtime.phase)
	safe := pattern.Lanes.SafeSequence[boss.bossLaneAttackCount%len(pattern.Lanes.SafeSequence)]
	for _, offset := range pattern.Lanes.AimOffsetsDegrees {
		angle := offset * math.Pi / 180
		forward := Vec2{
			X: boss.bossPhaseAim.X*math.Cos(angle) - boss.bossPhaseAim.Y*math.Sin(angle),
			Y: boss.bossPhaseAim.X*math.Sin(angle) + boss.bossPhaseAim.Y*math.Cos(angle),
		}
		lane := bossLaneFrame(boss.pos, forward, *pattern.Lanes, safe)
		if !s.bossLaneCorridorWalkable(lane) {
			continue
		}
		boss.bossLane = lane
		boss.bossLaneAttackCount++
		return true
	}
	boss.bossLane = nil
	return false
}

func (lane *BossLaneView) withPhase(phaseIndex int, phase BossPatternPhase) *BossLaneView {
	if lane == nil {
		return nil
	}
	view := *lane
	view.StageIndex = phaseIndex
	if phase.LaneIntensity > 0 {
		view.Intensity = phase.LaneIntensity
	}
	if phase.Kind == "active" {
		view.Intensity = 1
	}
	view.DangerLanes = append([]int(nil), lane.DangerLanes...)
	return &view
}

func (s *Sim) bossLaneCorridorWalkable(lane *BossLaneView) bool {
	if s == nil || lane == nil {
		return false
	}
	nav := s.activeNav()
	start := playerRadius + monsterRadius
	end := lane.Length - playerRadius
	if end <= start {
		return false
	}
	for step := 0; step <= int(math.Ceil((end-start)/playerRadius)); step++ {
		along := math.Min(end, start+float64(step)*playerRadius)
		center := Vec2{X: lane.Origin.X + lane.Forward.X*along + lane.Right.X*laneCrossCenter(lane, lane.SafeIndex),
			Y: lane.Origin.Y + lane.Forward.Y*along + lane.Right.Y*laneCrossCenter(lane, lane.SafeIndex)}
		cell := worldToGrid(nav, center)
		if cell.x < nav.GridBounds.MinX || cell.x > nav.GridBounds.MaxX || cell.y < nav.GridBounds.MinY || cell.y > nav.GridBounds.MaxY {
			return false
		}
		for _, wall := range s.activeWalls() {
			if obstacleBlocksMovement(wall) && circleIntersectsAABB(center, playerRadius, wall.pos, wall.size) {
				return false
			}
		}
	}
	return true
}

func laneCrossCenter(lane *BossLaneView, index int) float64 {
	return (float64(index) - float64(lane.Count-1)/2) * lane.Width
}

func bossLaneHitsPlayer(lane *BossLaneView, playerPos Vec2) bool {
	if lane == nil {
		return false
	}
	delta := Vec2{X: playerPos.X - lane.Origin.X, Y: playerPos.Y - lane.Origin.Y}
	along := delta.X*lane.Forward.X + delta.Y*lane.Forward.Y
	cross := delta.X*lane.Right.X + delta.Y*lane.Right.Y
	if along < -playerRadius || along > lane.Length+playerRadius {
		return false
	}
	for _, index := range lane.DangerLanes {
		if math.Abs(cross-laneCrossCenter(lane, index)) < lane.Width/2+playerRadius-meleeRangeEpsilon {
			return true
		}
	}
	return false
}
