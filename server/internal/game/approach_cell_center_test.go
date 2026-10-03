package game

import "testing"

// Regression for v535: in a wider generated layout the only free cells beside a closed room-threshold
// door are far from the door by their corner but close by their center. The approach search must still
// find a goal, so a player click on the door never answers no_path.
func TestClosedRoomDoorApproachAcceptsCellCenter(t *testing.T) {
	sim, err := NewSimWithWorld("sess_door_cell_center", "v526_room_threshold_03", loadRules(t), "room_door_interaction_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}
	var door *entity
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == woodenDoorDefID {
			door = e
			break
		}
	}
	if door == nil {
		t.Fatal("generated floor has no room-threshold door for this seed")
	}
	goal, steps, ok := sim.findMeleeApproachGoal(door)
	if !ok || len(steps) == 0 {
		t.Fatalf("no approach goal for closed door at %+v (ok=%v steps=%d)", door.pos, ok, len(steps))
	}
	if !meleeInRange(distance(goal, door.pos), sim.playerMeleeReach(), sim.targetInteractionRadius(door)) {
		t.Fatalf("approach goal %+v is not within interaction range of door %+v", goal, door.pos)
	}
}
