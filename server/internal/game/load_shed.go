package game

// SystemLoadShedInputType is a server-authored stored input. It carries the
// live runner's wall-clock load-shedding decision into the deterministic sim so
// replay can reproduce it (v476). Clients can never send it: inputdecode only
// accepts it from the persisted input stream.
//
// A row recorded at tick T describes a decision made after tick T finished.
// Live runners apply it immediately with ApplyLoadShed; replay passes it with
// tick T's inputs and the sim applies it after finalizing tick T, so both paths
// mutate state at the same point.
const SystemLoadShedInputType = "system_load_shed"

// LoadShedDirective is the load-shedding decision the runner made after a
// tick. OverloadDegrade requests an overload degradation window;
// CombatMovementThrottle is the combat-phase pressure signal. Movement is
// throttled on the next tick when either the window started or the pressure
// signal is set.
type LoadShedDirective struct {
	OverloadDegrade        bool
	CombatMovementThrottle bool
}

// ApplyLoadShed is the only runtime entry point for wall-clock driven
// degradation. Callers must record the same directive as a
// SystemLoadShedInputType input for the tick that just finished. It reports
// whether an overload degradation window was started or extended.
func (s *Sim) ApplyLoadShed(d LoadShedDirective) bool {
	applied := false
	if d.OverloadDegrade {
		applied = s.applyOverloadDegradation()
	}
	s.setCombatMovementThrottle(applied || d.CombatMovementThrottle)
	return applied
}

// LoadShedChanges reports whether applying d would mutate sim state. Runners use
// it to record directives only when they matter.
func (s *Sim) LoadShedChanges(d LoadShedDirective) bool {
	return d.OverloadDegrade || d.CombatMovementThrottle != s.combatMovementThrottled
}

// splitLoadShedInputs removes recorded load-shed directives from a tick's
// inputs. Only actor-less inputs count: player inputs always carry an actor, so
// a directive with an actor is dropped rather than honored.
func splitLoadShedInputs(inputs []Input) ([]Input, []LoadShedDirective) {
	found := false
	for _, in := range inputs {
		if in.Type == SystemLoadShedInputType {
			found = true
			break
		}
	}
	if !found {
		return inputs, nil
	}
	players := make([]Input, 0, len(inputs))
	var directives []LoadShedDirective
	for _, in := range inputs {
		if in.Type != SystemLoadShedInputType {
			players = append(players, in)
			continue
		}
		if in.ActorPlayerID == 0 && in.LoadShed != nil {
			directives = append(directives, *in.LoadShed)
		}
	}
	return players, directives
}
