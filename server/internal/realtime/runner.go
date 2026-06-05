package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const sendQueueSize = 256

// runner owns a single authenticated WebSocket connection and its session.
type runner struct {
	conn    *websocket.Conn
	sim     *game.Sim
	sess    store.Session
	store   store.Repository
	log     *slog.Logger
	metrics *metrics.Metrics

	mu       sync.Mutex
	buffer   map[uint64][]game.Input
	seen     map[string]bool
	received map[string]time.Time
	seq      int64

	sendCh    chan outEnvelope
	done      chan struct{}
	closeOnce sync.Once
}

func newRunner(conn *websocket.Conn, sim *game.Sim, sess store.Session, st store.Repository, log *slog.Logger, m *metrics.Metrics, meta *replay.ResumeMetadata) *runner {
	seen := make(map[string]bool)
	seq := int64(0)
	if meta != nil {
		for id := range meta.SeenMessageIDs {
			seen[id] = true
		}
		seq = meta.NextSequence
	}
	return &runner{
		conn:     conn,
		sim:      sim,
		sess:     sess,
		store:    st,
		log:      logging.Component(log, "realtime").With("session_id", sess.ID),
		metrics:  m,
		buffer:   make(map[uint64][]game.Input),
		seen:     seen,
		received: make(map[string]time.Time),
		seq:      seq,
		sendCh:   make(chan outEnvelope, sendQueueSize),
		done:     make(chan struct{}),
	}
}

// run drives the connection until it closes or the context is cancelled.
func (r *runner) run(ctx context.Context) {
	r.metrics.WSConnections.Inc()
	defer r.metrics.WSConnections.Dec()
	defer r.close()

	go r.writeLoop()
	go r.readLoop()

	// Initial full snapshot so the client can render immediately.
	r.enqueue(r.snapshotEnvelope())

	r.tickLoop(ctx)
}

func (r *runner) close() {
	r.closeOnce.Do(func() {
		close(r.done)
		_ = r.conn.Close()
	})
}

// --- write side -------------------------------------------------------------

func (r *runner) writeLoop() {
	for {
		select {
		case <-r.done:
			return
		case env := <-r.sendCh:
			if err := r.conn.WriteJSON(env); err != nil {
				r.close()
				return
			}
		}
	}
}

// enqueue queues a message for the single writer. If the queue is full the
// client is too slow and the connection is closed.
func (r *runner) enqueue(env outEnvelope) {
	select {
	case <-r.done:
	case r.sendCh <- env:
	default:
		r.log.Warn("send queue full; closing connection")
		r.close()
	}
}

func (r *runner) snapshotEnvelope() outEnvelope {
	snap := r.sim.Snapshot()
	return outEnvelope{
		Type:      typeSnapshot,
		MessageID: ids.New("msg"),
		SessionID: r.sess.ID,
		Tick:      snap.ServerTick,
		Payload:   snap,
	}
}

func (r *runner) sendError(code, message string) {
	r.enqueue(outEnvelope{
		Type:      typeError,
		MessageID: ids.New("msg"),
		SessionID: r.sess.ID,
		Tick:      r.currentTick(),
		Payload:   errorPayload{Code: code, Message: message},
	})
}

func (r *runner) currentTick() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sim.CurrentTick()
}

// --- read side --------------------------------------------------------------

func (r *runner) readLoop() {
	for {
		_, data, err := r.conn.ReadMessage()
		if err != nil {
			r.close()
			return
		}
		r.handleMessage(data)
	}
}

func (r *runner) handleMessage(data []byte) {
	var env inEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		r.sendError("bad_message", "malformed JSON envelope")
		return
	}
	if env.Type == "" || env.MessageID == "" {
		r.sendError("bad_message", "envelope missing type or message_id")
		return
	}
	if env.SessionID != "" && env.SessionID != r.sess.ID {
		r.sendError("bad_session", "session_id does not match this connection")
		return
	}

	if env.Type == typeClientReady {
		// Re-send a full snapshot and acknowledge readiness.
		r.enqueue(r.snapshotEnvelope())
		r.enqueue(r.acceptedEnvelope(env.MessageID, r.currentTick(), env.CorrelationID))
		return
	}

	if !isClientIntent(env.Type) {
		r.sendError("bad_message", "unknown message type: "+env.Type)
		return
	}

	in, ok := decodeInput(env)
	if !ok {
		r.rejectIntent(env.MessageID, "invalid_payload", env.CorrelationID)
		return
	}

	// Buffer the input for its (clamped) tick under the lock, then persist
	// outside the lock.
	r.mu.Lock()
	if r.seen[env.MessageID] {
		r.mu.Unlock()
		r.rejectIntent(env.MessageID, "duplicate", env.CorrelationID)
		return
	}
	r.seen[env.MessageID] = true
	cur := r.sim.CurrentTick()
	t := env.Tick
	if t < cur {
		t = cur // late input: apply at the current tick (acknowledged, not dropped)
	}
	in.Sequence = r.seq
	r.seq++
	r.buffer[t] = append(r.buffer[t], in)
	r.received[env.MessageID] = time.Now()
	rec := store.SessionInput{
		ID:            ids.New("inp"),
		SessionID:     r.sess.ID,
		Tick:          int64(t),
		Sequence:      in.Sequence,
		MessageID:     env.MessageID,
		CorrelationID: env.CorrelationID,
		// Persist the full envelope so replay can recover the message type
		// (session_inputs has no type column, per spec 4.6).
		Payload: json.RawMessage(data),
	}
	r.mu.Unlock()

	if err := r.store.AppendInput(context.Background(), rec); err != nil {
		r.metrics.PersistenceErrors.Inc()
		r.log.Error("persist input", "error", err)
	}
}

func (r *runner) acceptedEnvelope(messageID string, tick uint64, corr string) outEnvelope {
	return outEnvelope{
		Type:          typeIntentAccepted,
		MessageID:     ids.New("msg"),
		SessionID:     r.sess.ID,
		Tick:          tick,
		CorrelationID: corr,
		Payload:       intentAcceptedPayload{AcceptedMessageID: messageID, ServerTick: tick},
	}
}

func (r *runner) rejectIntent(messageID, reason, corr string) {
	r.metrics.RejectedIntents.Inc()
	r.enqueue(outEnvelope{
		Type:          typeIntentRejected,
		MessageID:     ids.New("msg"),
		SessionID:     r.sess.ID,
		Tick:          r.currentTick(),
		CorrelationID: corr,
		Payload:       intentRejectedPayload{RejectedMessageID: messageID, Reason: reason},
	})
}

// --- tick loop --------------------------------------------------------------

func (r *runner) tickLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second / tickHz)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.done:
			return
		case <-ticker.C:
			r.doTick()
		}
	}
}

func (r *runner) doTick() {
	start := time.Now()

	r.mu.Lock()
	t := r.sim.CurrentTick()
	inputs := r.buffer[t]
	delete(r.buffer, t)
	sortInputs(inputs)
	res := r.sim.Tick(inputs)
	// Capture per-input receipt times for latency, then forget them.
	latencies := make([]time.Duration, 0, len(res.Acks)+len(res.Rejects))
	for _, a := range res.Acks {
		if recv, ok := r.received[a.MessageID]; ok {
			latencies = append(latencies, time.Since(recv))
			delete(r.received, a.MessageID)
		}
	}
	r.mu.Unlock()

	r.metrics.TickDuration.Observe(time.Since(start).Seconds())
	for _, l := range latencies {
		r.metrics.MessageLatency.Observe(l.Seconds())
	}

	// Acks / rejects.
	for _, a := range res.Acks {
		r.enqueue(r.acceptedEnvelope(a.MessageID, res.Tick, ""))
	}
	for _, rej := range res.Rejects {
		r.rejectIntent(rej.MessageID, rej.Reason, "")
	}

	// State delta (only when something changed). Slices are coerced to non-nil
	// so they marshal as [] not null, matching the state_delta schema.
	if len(res.Changes) > 0 || len(res.Events) > 0 {
		changes := res.Changes
		if changes == nil {
			changes = []game.Change{}
		}
		events := res.Events
		if events == nil {
			events = []game.Event{}
		}
		r.enqueue(outEnvelope{
			Type:      typeStateDelta,
			MessageID: ids.New("msg"),
			SessionID: r.sess.ID,
			Tick:      res.Tick,
			Payload:   stateDeltaPayload{ServerTick: res.Tick, Changes: changes, Events: events},
		})
	}

	r.persistTick(res)
}

// persistTick writes events and inventory mutations produced by the tick.
func (r *runner) persistTick(res game.TickResult) {
	ctx := context.Background()
	for i, ev := range res.Events {
		payload, _ := json.Marshal(ev)
		err := r.store.AppendEvent(ctx, store.SessionEvent{
			ID:            ids.New("evt"),
			SessionID:     r.sess.ID,
			Tick:          int64(res.Tick),
			Sequence:      int64(i),
			EventType:     ev.EventType,
			CorrelationID: ev.CorrelationID,
			Payload:       payload,
		})
		if err != nil {
			r.metrics.PersistenceErrors.Inc()
			r.log.Error("persist event", "error", err)
		}
	}

	for _, c := range res.Changes {
		switch c.Op {
		case game.OpInventoryAdd:
			if c.Item == nil {
				continue
			}
			err := r.store.AddInventoryItem(ctx, store.InventoryItem{
				ID:          c.Item.ItemInstanceID,
				SessionID:   r.sess.ID,
				AccountID:   r.sess.AccountID,
				CharacterID: r.sess.CharacterID,
				ItemDefID:   c.Item.ItemDefID,
				Slot:        c.Item.Slot,
				Equipped:    c.Item.Equipped,
			})
			if err != nil {
				r.metrics.PersistenceErrors.Inc()
				r.log.Error("persist inventory add", "error", err)
			}
		case game.OpInventoryUpdate:
			if c.Item == nil {
				continue
			}
			if err := r.store.SetEquipped(ctx, r.sess.ID, c.Item.ItemInstanceID, c.Item.Slot, c.Item.Equipped); err != nil {
				r.metrics.PersistenceErrors.Inc()
				r.log.Error("persist inventory update", "error", err)
			}
		}
	}

	if err := r.store.TouchSession(ctx, r.sess.ID); err != nil {
		r.metrics.PersistenceErrors.Inc()
	}
}

// sortInputs orders inputs deterministically by (sequence, message_id).
func sortInputs(inputs []game.Input) {
	for i := 1; i < len(inputs); i++ {
		for j := i; j > 0 && less(inputs[j], inputs[j-1]); j-- {
			inputs[j], inputs[j-1] = inputs[j-1], inputs[j]
		}
	}
}

func less(a, b game.Input) bool {
	if a.Sequence != b.Sequence {
		return a.Sequence < b.Sequence
	}
	return a.MessageID < b.MessageID
}
