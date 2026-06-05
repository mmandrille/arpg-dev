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
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
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

// ResumeMetadata carries the runner state needed to continue after replaying
// historical inputs.
type ResumeMetadata struct {
	SeenMessageIDs map[string]bool
	NextSequence   int64
}

// RecordedInput is a store-independent input stamped with its authoritative
// simulation tick.
type RecordedInput struct {
	Tick  int64
	Input game.Input
}

// Reconstruction is the authoritative state rebuilt from seed + inputs.
type Reconstruction struct {
	Sim           *game.Sim
	Snapshot      game.Snapshot
	DerivedEvents []derivedEvent
	Session       store.Session
	Metadata      ResumeMetadata
}

// Reconstruct re-simulates the session from seed + recorded inputs, returning
// the restored sim, snapshot, derived event stream, and resume metadata.
func Reconstruct(ctx context.Context, repo store.Repository, rules *game.Rules, sessionID string) (Reconstruction, error) {
	sess, err := repo.GetSession(ctx, sessionID)
	if err != nil {
		return Reconstruction{}, err
	}
	inputs, err := repo.ListInputs(ctx, sessionID)
	if err != nil {
		return Reconstruction{Session: sess}, err
	}
	recorded, err := repo.ListEvents(ctx, sessionID)
	if err != nil {
		return Reconstruction{Session: sess}, err
	}

	recordedInputs, maxTick, err := StoredInputs(inputs)
	if err != nil {
		return Reconstruction{Session: sess}, err
	}
	for _, ev := range recorded {
		if ev.Tick > maxTick {
			maxTick = ev.Tick
		}
	}

	recon := ReconstructFromInputs(sessionID, sess.Seed, rules, recordedInputs, maxTick)
	recon.Session = sess
	return recon, nil
}

// StoredInputs converts durable input rows into replay inputs. The stored row
// owns sequencing and dedupe metadata; the JSON payload only supplies type and
// intent-specific fields.
func StoredInputs(rows []store.SessionInput) ([]RecordedInput, int64, error) {
	recorded := make([]RecordedInput, 0, len(rows))
	maxTick := int64(-1)
	for _, row := range rows {
		in, ok := inputdecode.DecodeStored(row.Payload)
		if !ok {
			return nil, maxTick, fmt.Errorf("decode stored input: session_id=%s input_id=%s tick=%d message_id=%s",
				row.SessionID, row.ID, row.Tick, row.MessageID)
		}
		in.MessageID = row.MessageID
		in.Sequence = row.Sequence
		in.CorrelationID = row.CorrelationID
		recorded = append(recorded, RecordedInput{Tick: row.Tick, Input: in})
		if row.Tick > maxTick {
			maxTick = row.Tick
		}
	}
	return recorded, maxTick, nil
}

// ReconstructFromInputs rebuilds a session from an already-decoded input stream.
// throughTick is inclusive; pass -1 for a fresh untouched session.
func ReconstructFromInputs(sessionID, seed string, rules *game.Rules, inputs []RecordedInput, throughTick int64) Reconstruction {
	byTick := make(map[int64][]game.Input)
	meta := ResumeMetadata{
		SeenMessageIDs: make(map[string]bool, len(inputs)),
	}
	for _, rec := range inputs {
		byTick[rec.Tick] = append(byTick[rec.Tick], rec.Input)
		if rec.Input.MessageID != "" {
			meta.SeenMessageIDs[rec.Input.MessageID] = true
		}
		if rec.Input.Sequence >= meta.NextSequence {
			meta.NextSequence = rec.Input.Sequence + 1
		}
	}

	sim := game.NewSim(sessionID, seed, rules)
	var derived []derivedEvent
	for t := int64(0); t <= throughTick; t++ {
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

	return Reconstruction{
		Sim:           sim,
		Snapshot:      sim.Snapshot(),
		DerivedEvents: derived,
		Metadata:      meta,
	}
}

// Verify reconstructs the session and compares derived events against the
// recorded events. The returned Report has Match=false (and a Mismatch reason)
// if the same seed + inputs did not reproduce the recorded authoritative
// output.
func Verify(ctx context.Context, repo store.Repository, rules *game.Rules, sessionID string) (Report, error) {
	recon, err := Reconstruct(ctx, repo, rules, sessionID)
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
		Seed:               recon.Session.Seed,
		InputCount:         len(inputs),
		RecordedEventCount: len(recorded),
		DerivedEventCount:  len(recon.DerivedEvents),
		Snapshot:           recon.Snapshot,
		Match:              true,
	}

	if len(recon.DerivedEvents) != len(recorded) {
		rep.Match = false
		rep.Mismatch = fmt.Sprintf("event count: derived %d, recorded %d", len(recon.DerivedEvents), len(recorded))
		return rep, nil
	}
	for i := range recon.DerivedEvents {
		d, r := recon.DerivedEvents[i], recorded[i]
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
