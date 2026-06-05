// Package replay re-simulates a recorded session from its seed and input
// stream and verifies that it reproduces the recorded authoritative events
// (ADR-0001 D8.2). It powers both the inspection /state endpoint and the
// replay verification CLI.
package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/realtime"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// derivedEvent mirrors a recorded session_event for comparison.
type derivedEvent struct {
	Tick      int64
	Sequence  int64
	EventType string
	Payload   json.RawMessage
}

// Report summarizes a replay verification.
type Report struct {
	SessionID          string        `json:"session_id"`
	Seed               string        `json:"seed"`
	InputCount         int           `json:"input_count"`
	RecordedEventCount int           `json:"recorded_event_count"`
	DerivedEventCount  int           `json:"derived_event_count"`
	Match              bool          `json:"match"`
	Mismatch           string        `json:"mismatch,omitempty"`
	Snapshot           game.Snapshot `json:"snapshot"`
}

// Reconstruct re-simulates the session from seed + recorded inputs, returning
// the resulting authoritative snapshot and the derived event stream.
func Reconstruct(ctx context.Context, repo store.Repository, rules *game.Rules, sessionID string) (game.Snapshot, []derivedEvent, store.Session, error) {
	sess, err := repo.GetSession(ctx, sessionID)
	if err != nil {
		return game.Snapshot{}, nil, store.Session{}, err
	}
	inputs, err := repo.ListInputs(ctx, sessionID)
	if err != nil {
		return game.Snapshot{}, nil, sess, err
	}
	recorded, err := repo.ListEvents(ctx, sessionID)
	if err != nil {
		return game.Snapshot{}, nil, sess, err
	}

	byTick := make(map[int64][]game.Input)
	maxTick := int64(0)
	for _, row := range inputs {
		in, ok := realtime.DecodeStored(row.Payload)
		if !ok {
			continue
		}
		in.Sequence = row.Sequence
		in.CorrelationID = row.CorrelationID
		byTick[row.Tick] = append(byTick[row.Tick], in)
		if row.Tick > maxTick {
			maxTick = row.Tick
		}
	}
	for _, ev := range recorded {
		if ev.Tick > maxTick {
			maxTick = ev.Tick
		}
	}

	sim := game.NewSim(sessionID, sess.Seed, rules)
	var derived []derivedEvent
	for t := int64(0); t <= maxTick; t++ {
		ins := byTick[t]
		sort.SliceStable(ins, func(i, j int) bool {
			if ins[i].Sequence != ins[j].Sequence {
				return ins[i].Sequence < ins[j].Sequence
			}
			return ins[i].MessageID < ins[j].MessageID
		})
		res := sim.Tick(ins)
		for i, ev := range res.Events {
			payload, _ := json.Marshal(ev)
			derived = append(derived, derivedEvent{
				Tick:      int64(res.Tick),
				Sequence:  int64(i),
				EventType: ev.EventType,
				Payload:   payload,
			})
		}
	}

	return sim.Snapshot(), derived, sess, nil
}

// Verify reconstructs the session and compares derived events against the
// recorded events. The returned Report has Match=false (and a Mismatch reason)
// if the same seed + inputs did not reproduce the recorded authoritative
// output.
func Verify(ctx context.Context, repo store.Repository, rules *game.Rules, sessionID string) (Report, error) {
	snap, derived, sess, err := Reconstruct(ctx, repo, rules, sessionID)
	if err != nil {
		return Report{}, err
	}
	recorded, err := repo.ListEvents(ctx, sessionID)
	if err != nil {
		return Report{}, err
	}
	inputs, err := repo.ListInputs(ctx, sessionID)
	if err != nil {
		return Report{}, err
	}

	rep := Report{
		SessionID:          sessionID,
		Seed:               sess.Seed,
		InputCount:         len(inputs),
		RecordedEventCount: len(recorded),
		DerivedEventCount:  len(derived),
		Snapshot:           snap,
		Match:              true,
	}

	if len(derived) != len(recorded) {
		rep.Match = false
		rep.Mismatch = fmt.Sprintf("event count: derived %d, recorded %d", len(derived), len(recorded))
		return rep, nil
	}
	for i := range derived {
		d, r := derived[i], recorded[i]
		if d.EventType != r.EventType || d.Tick != r.Tick || d.Sequence != r.Sequence {
			rep.Match = false
			rep.Mismatch = fmt.Sprintf("event %d: derived (%s,t%d,s%d) != recorded (%s,t%d,s%d)",
				i, d.EventType, d.Tick, d.Sequence, r.EventType, r.Tick, r.Sequence)
			return rep, nil
		}
		if !jsonEqual(d.Payload, r.Payload) {
			rep.Match = false
			rep.Mismatch = fmt.Sprintf("event %d payload differs: derived %s != recorded %s", i, d.Payload, r.Payload)
			return rep, nil
		}
	}
	return rep, nil
}

func jsonEqual(a, b []byte) bool {
	var ma, mb any
	if err := json.Unmarshal(a, &ma); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &mb); err != nil {
		return false
	}
	return reflect.DeepEqual(ma, mb)
}
