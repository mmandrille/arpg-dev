package game

// simTickCtx bundles per-tick scratch state for TickResultsProfiled so later
// extractions can move phase helpers without threading many locals.
type simTickCtx struct {
	sim                *Sim
	resultByKey        map[tickResultKey]*TickResult
	ordered            []*TickResult
	transitionThisTick bool
}

type tickResultKey struct {
	level int
	actor uint64
}

func newSimTickCtx(sim *Sim) *simTickCtx {
	return &simTickCtx{
		sim:         sim,
		resultByKey: map[tickResultKey]*TickResult{},
		ordered:     []*TickResult{},
	}
}

func (ctx *simTickCtx) resultFor(level int, actor uint64) *TickResult {
	key := tickResultKey{level: level, actor: actor}
	if res := ctx.resultByKey[key]; res != nil {
		return res
	}
	res := &TickResult{
		Tick:  ctx.sim.tick,
		Level: level,
		ActorPlayerID: actor,
		Changes:       []Change{},
		Events:        []Event{},
	}
	ctx.resultByKey[key] = res
	ctx.ordered = append(ctx.ordered, res)
	return res
}

func (ctx *simTickCtx) markTransition() {
	ctx.transitionThisTick = true
}

func (ctx *simTickCtx) finalizeResults() []TickResult {
	ctx.sim.tick++
	ctx.sim.usePlayer(ctx.sim.defaultPlayer())

	results := make([]TickResult, 0, len(ctx.ordered))
	for _, res := range ctx.ordered {
		if len(res.Changes) == 0 && len(res.Events) == 0 && len(res.Acks) == 0 && len(res.Rejects) == 0 {
			continue
		}
		results = append(results, *res)
	}
	if len(results) == 0 {
		return []TickResult{{Tick: ctx.sim.tick - 1, Level: ctx.sim.currentLevel, Changes: []Change{}, Events: []Event{}}}
	}
	return results
}
