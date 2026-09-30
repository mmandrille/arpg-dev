package realtime

// maxClientTickLead bounds how far past the current sim tick a client input
// may be scheduled. Real clients send their last seen server tick (<= cur).
// The envelope tick is client-controlled and persisted with the input row, so
// an unbounded future tick would force every later resume, /state, /replay and
// `make replay` to simulate up to it, and would hold the input back from live.
const maxClientTickLead uint64 = 2

// scheduleClientTick returns the tick a client input runs on: past ticks run
// now, near-future ticks within maxClientTickLead are honoured, and anything
// further ahead also runs now. The returned tick is the one recorded for replay.
func scheduleClientTick(requested, cur uint64) uint64 {
	if requested < cur || requested > cur+maxClientTickLead {
		return cur
	}
	return requested
}
