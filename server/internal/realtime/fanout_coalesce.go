package realtime

import (
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
)

type clientFanoutBatch struct {
	acks    []intentAcceptedPayload
	rejects []intentRejectedPayload
	deltas  map[int]*stateDeltaPayload
	tick    uint64
}

func newClientFanoutBatch() *clientFanoutBatch {
	return &clientFanoutBatch{deltas: make(map[int]*stateDeltaPayload)}
}

func mergeStateDeltaPayload(dst, src *stateDeltaPayload) {
	if dst == nil || src == nil {
		return
	}
	dst.Changes = append(dst.Changes, src.Changes...)
	dst.Events = append(dst.Events, src.Events...)
	if src.Performance != nil {
		dst.Performance = src.Performance
	}
}

func (l *sessionLoop) fanoutTickResults(results []game.TickResult, clients []*loopClient, inputTypes map[string]string, levelsByPlayerID map[uint64]int) {
	batches := make(map[*loopClient]*clientFanoutBatch, len(clients))
	for _, client := range clients {
		batches[client] = newClientFanoutBatch()
	}
	for _, res := range results {
		l.accumulateFanoutResults(res, clients, inputTypes, levelsByPlayerID, batches)
	}
	for client, batch := range batches {
		l.flushClientFanoutBatch(client, batch)
	}
}

func (l *sessionLoop) accumulateFanoutResults(res game.TickResult, clients []*loopClient, inputTypes map[string]string, levelsByPlayerID map[uint64]int, batches map[*loopClient]*clientFanoutBatch) {
	for _, client := range clients {
		level, ok := levelsByPlayerID[client.playerID]
		if !ok {
			continue
		}
		batch := batches[client]
		if batch.tick < res.Tick {
			batch.tick = res.Tick
		}
		for _, ack := range res.Acks {
			if res.ActorPlayerID == client.playerID {
				batch.acks = append(batch.acks, intentAcceptedPayload{AcceptedMessageID: ack.MessageID, ServerTick: res.Tick})
				_ = inputTypes
			}
		}
		for _, rej := range res.Rejects {
			if res.ActorPlayerID == client.playerID {
				batch.rejects = append(batch.rejects, intentRejectedPayload{RejectedMessageID: rej.MessageID, Reason: rej.Reason})
			}
		}
		events := filterEventsForClient(res.Events, res.ActorPlayerID, client.playerID)
		events = l.sim.FilterEventsForPlayer(client.playerID, res.Level, events)
		if level != res.Level {
			if len(events) == 0 {
				continue
			}
			payload := &stateDeltaPayload{
				ServerTick: res.Tick,
				Level:      level,
				Changes:    []game.Change{},
				Events:     events,
			}
			if existing, ok := batch.deltas[level]; ok {
				mergeStateDeltaPayload(existing, payload)
			} else {
				batch.deltas[level] = payload
			}
			continue
		}
		changes := filterChangesForClient(res.Changes, res.ActorPlayerID, client.playerID)
		changes = l.sim.FilterChangesForPlayer(client.playerID, res.Level, changes)
		if len(changes) == 0 && len(events) == 0 {
			continue
		}
		payload := &stateDeltaPayload{
			ServerTick: res.Tick,
			Level:      res.Level,
			Changes:    changes,
			Events:     events,
		}
		if existing, ok := batch.deltas[res.Level]; ok {
			mergeStateDeltaPayload(existing, payload)
		} else {
			batch.deltas[res.Level] = payload
		}
	}
}

func (l *sessionLoop) flushClientFanoutBatch(client *loopClient, batch *clientFanoutBatch) {
	for _, ack := range batch.acks {
		client.enqueue(outEnvelope{
			Type:      typeIntentAccepted,
			MessageID: ids.New("msg"),
			SessionID: l.sess.ID,
			Tick:      ack.ServerTick,
			Payload:   ack,
		})
	}
	for _, rej := range batch.rejects {
		l.hub.metrics.RejectedIntents.Inc()
		client.enqueue(outEnvelope{
			Type:      typeIntentRejected,
			MessageID: ids.New("msg"),
			SessionID: l.sess.ID,
			Tick:      batch.tick,
			Payload:   rej,
		})
	}
	for _, payload := range batch.deltas {
		client.enqueue(outEnvelope{
			Type:      typeStateDelta,
			MessageID: ids.New("msg"),
			SessionID: l.sess.ID,
			Tick:      payload.ServerTick,
			Payload:   *payload,
		})
	}
}
