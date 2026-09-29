package game

// SystemTickCheckpointInputType is a server-authored stored input (v481). A
// row at tick T says "live finished tick T". It changes nothing: replay drops it
// before TickResults, but it counts toward the replayed tick range, so quiet
// ticks after the last event or input still run in replay and resume. Clients
// can never send it: inputdecode only accepts it from the persisted stream.
const SystemTickCheckpointInputType = "system_tick_checkpoint"
