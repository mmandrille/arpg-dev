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

// Envelope is a protocol-shaped server-to-client replay message.
type Envelope struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
	Tick      uint64 `json:"tick"`
	Payload   any    `json:"payload"`
}

// Timeline is a visual/debug replay stream reconstructed from seed + inputs.
type Timeline struct {
	SessionID string     `json:"session_id"`
	Seed      string     `json:"seed"`
	Envelopes []Envelope `json:"envelopes"`
}

// StateDeltaPayload mirrors the live WebSocket state_delta payload without
// importing the realtime package into replay.
type StateDeltaPayload struct {
	ServerTick uint64        `json:"server_tick"`
	Level      int           `json:"level"`
	Changes    []game.Change `json:"changes"`
	Events     []game.Event  `json:"events"`
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

	start, err := repo.LoadSessionStartSnapshot(ctx, sessionID)
	if err != nil {
		return Reconstruction{Session: sess}, err
	}
	recon, err := ReconstructFromInputsWithProgression(sessionID, sess.Seed, rules, normalizeWorldID(sess.WorldID), recordedInputs, maxTick, persistedItems(start.Items), waypointLevels(start.Waypoints))
	if err != nil {
		return Reconstruction{Session: sess}, err
	}
	recon.Session = sess
	return recon, nil
}

// BuildTimeline reconstructs a protocol-shaped replay stream for local visual
// tooling. It intentionally emits only snapshot/delta messages, not acks, so
// consumers render authoritative state without re-sending inputs.
// throughTick extends simulation when the live session advanced without
// durable inputs (for example bot wait_ticks); values below maxTick are ignored.
func BuildTimeline(ctx context.Context, repo store.Repository, rules *game.Rules, sessionID string, throughTick int64) (Timeline, error) {
	sess, err := repo.GetSession(ctx, sessionID)
	if err != nil {
		return Timeline{}, err
	}
	inputs, err := repo.ListInputs(ctx, sessionID)
	if err != nil {
		return Timeline{}, err
	}
	recorded, err := repo.ListEvents(ctx, sessionID)
	if err != nil {
		return Timeline{}, err
	}
	recordedInputs, maxTick, err := StoredInputs(inputs)
	if err != nil {
		return Timeline{}, err
	}
	for _, ev := range recorded {
		if ev.Tick > maxTick {
			maxTick = ev.Tick
		}
	}
	if throughTick > maxTick {
		maxTick = throughTick
	}

	byTick := inputsByTick(recordedInputs)
	sim, err := game.NewSimWithWorld(sessionID, sess.Seed, rules, normalizeWorldID(sess.WorldID))
	if err != nil {
		return Timeline{}, err
	}
	start, err := repo.LoadSessionStartSnapshot(ctx, sessionID)
	if err != nil {
		return Timeline{}, err
	}
	sim.LoadInventory(persistedItems(start.Items))
	sim.LoadDiscoveredTeleporters(waypointLevels(start.Waypoints))
	out := Timeline{
		SessionID: sessionID,
		Seed:      sess.Seed,
		Envelopes: []Envelope{{
			Type:      "session_snapshot",
			MessageID: "replay-snapshot",
			SessionID: sessionID,
			Tick:      sim.CurrentTick(),
			Payload:   sim.Snapshot(),
		}},
	}

	for t := int64(0); t <= maxTick; t++ {
		ins := byTick[t]
		sortInputs(ins)
		results := sim.TickResults(ins)
		for i, res := range results {
			if len(res.Changes) == 0 && len(res.Events) == 0 {
				continue
			}
			out.Envelopes = append(out.Envelopes, Envelope{
				Type:      "state_delta",
				MessageID: fmt.Sprintf("replay-tick-%d-%d", res.Tick, i),
				SessionID: sessionID,
				Tick:      res.Tick,
				Payload: StateDeltaPayload{
					ServerTick: res.Tick,
					Level:      res.Level,
					Changes:    res.Changes,
					Events:     res.Events,
				},
			})
		}
	}
	return out, nil
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
func ReconstructFromInputs(sessionID, seed string, rules *game.Rules, worldID string, inputs []RecordedInput, throughTick int64) (Reconstruction, error) {
	return ReconstructFromInputsWithProgression(sessionID, seed, rules, worldID, inputs, throughTick, nil, nil)
}

func ReconstructFromInputsWithProgression(sessionID, seed string, rules *game.Rules, worldID string, inputs []RecordedInput, throughTick int64, items []game.PersistedItem, waypointLevels []int) (Reconstruction, error) {
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

	sim, err := game.NewSimWithWorld(sessionID, seed, rules, worldID)
	if err != nil {
		return Reconstruction{}, err
	}
	sim.LoadInventory(items)
	sim.LoadDiscoveredTeleporters(waypointLevels)
	var derived []derivedEvent
	for t := int64(0); t <= throughTick; t++ {
		ins := byTick[t]
		sortInputs(ins)
		results := sim.TickResults(ins)
		sequence := int64(0)
		for _, res := range results {
			for _, ev := range res.Events {
				payload, _ := json.Marshal(ev)
				derived = append(derived, derivedEvent{
					Tick:      int64(res.Tick),
					Sequence:  sequence,
					EventType: ev.EventType,
					Payload:   payload,
				})
				sequence++
			}
		}
	}

	return Reconstruction{
		Sim:           sim,
		Snapshot:      sim.Snapshot(),
		DerivedEvents: derived,
		Metadata:      meta,
	}, nil
}

func persistedItems(items []store.CharacterItemInstance) []game.PersistedItem {
	out := make([]game.PersistedItem, 0, len(items))
	for _, item := range items {
		if item.Location != store.ItemLocationInventory && item.Location != store.ItemLocationEquipped {
			continue
		}
		out = append(out, game.PersistedItem{
			InstanceID: item.ID,
			ItemDefID:  item.ItemDefID,
			Slot:       item.Slot,
			Equipped:   item.Equipped,
		})
	}
	return out
}

func waypointLevels(waypoints []store.CharacterWaypoint) []int {
	out := make([]int, 0, len(waypoints))
	for _, wp := range waypoints {
		out = append(out, wp.Level)
	}
	return out
}

func inputsByTick(inputs []RecordedInput) map[int64][]game.Input {
	byTick := make(map[int64][]game.Input)
	for _, rec := range inputs {
		byTick[rec.Tick] = append(byTick[rec.Tick], rec.Input)
	}
	return byTick
}

func sortInputs(ins []game.Input) {
	sort.SliceStable(ins, func(i, j int) bool {
		if ins[i].Sequence != ins[j].Sequence {
			return ins[i].Sequence < ins[j].Sequence
		}
		return ins[i].MessageID < ins[j].MessageID
	})
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

func normalizeWorldID(worldID string) string {
	if worldID == "" {
		return game.DefaultWorldID
	}

	return worldID
}
