package realtime

import "sync"

// loopClient overflow coalescing keeps the tick loop from blocking when sendCh
// is full: state_delta envelopes merge in memory and drain from writeLoop.
type clientSendOverflow struct {
	mu       sync.Mutex
	pending  *outEnvelope
}

func (o *clientSendOverflow) merge(env outEnvelope) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.pending == nil {
		copy := env
		o.pending = &copy
		return true
	}
	merged, ok := coalesceOutbound(*o.pending, env)
	if !ok {
		return false
	}
	*o.pending = merged
	return true
}

func (o *clientSendOverflow) take() (outEnvelope, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.pending == nil {
		return outEnvelope{}, false
	}
	env := *o.pending
	o.pending = nil
	return env, true
}

func coalesceOutbound(existing, incoming outEnvelope) (outEnvelope, bool) {
	if existing.Type != typeStateDelta || incoming.Type != typeStateDelta {
		return outEnvelope{}, false
	}
	existingPayload, ok := existing.Payload.(stateDeltaPayload)
	if !ok {
		return outEnvelope{}, false
	}
	incomingPayload, ok := incoming.Payload.(stateDeltaPayload)
	if !ok {
		return outEnvelope{}, false
	}
	if incoming.Tick > existing.Tick {
		existing.Tick = incoming.Tick
	}
	if incomingPayload.ServerTick > existingPayload.ServerTick {
		existingPayload.ServerTick = incomingPayload.ServerTick
	}
	existingPayload.Changes = append(existingPayload.Changes, incomingPayload.Changes...)
	existingPayload.Events = append(existingPayload.Events, incomingPayload.Events...)
	if incomingPayload.Performance != nil {
		existingPayload.Performance = incomingPayload.Performance
	}
	existing.Payload = existingPayload
	return existing, true
}
