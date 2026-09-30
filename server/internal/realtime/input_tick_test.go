package realtime

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestScheduleClientTickBoundsFutureTicks(t *testing.T) {
	const cur = 100
	cases := []struct {
		name      string
		requested uint64
		want      uint64
	}{
		{"past tick runs now", 3, cur},
		{"current tick", cur, cur},
		{"near future within lead", cur + maxClientTickLead, cur + maxClientTickLead},
		{"beyond lead runs now", cur + maxClientTickLead + 1, cur},
		{"huge tick runs now", math.MaxUint64, cur},
	}
	for _, tc := range cases {
		if got := scheduleClientTick(tc.requested, cur); got != tc.want {
			t.Errorf("%s: scheduleClientTick(%d, %d) = %d, want %d", tc.name, tc.requested, cur, got, tc.want)
		}
	}
}

// A far-future envelope tick must be persisted at the current tick, so replay
// never has to simulate up to a client-chosen tick and the input runs live.
func TestFarFutureClientTickIsRecordedAtCurrentTick(t *testing.T) {
	ctx := context.Background()
	repo, loop := newLoadShedTestLoop(t)
	host := repo.memberByRole(store.SessionMemberHost)
	client := &loopClient{loop: loop, key: memberKey(host), member: host, sendCh: make(chan outEnvelope, 64), done: make(chan struct{})}
	client.playerID = admitLoadShedTestMember(ctx, loop, host)
	loop.clients[client.key] = client
	loop.doTick()

	cur := loop.sim.CurrentTick()
	env, err := json.Marshal(map[string]any{
		"type":       "move_intent",
		"message_id": "msg_far_future",
		"session_id": repo.session.ID,
		"tick":       uint64(1_000_000_000),
		"payload":    map[string]any{"direction": map[string]any{"x": 1, "y": 0}, "duration_ticks": 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	loop.handleClientMessage(client, env)

	var recorded *store.SessionInput
	for i := range repo.inputs {
		if repo.inputs[i].MessageID == "msg_far_future" {
			recorded = &repo.inputs[i]
		}
	}
	if recorded == nil {
		t.Fatal("far-future input was not persisted")
	}
	if recorded.Tick != int64(cur) {
		t.Fatalf("persisted tick = %d, want current tick %d", recorded.Tick, cur)
	}
	loop.doTick()
	drainLoadShedTestClient(client)
	report, err := replay.Verify(ctx, repo, repo.rules, repo.session.ID)
	if err != nil || !report.Match {
		t.Fatalf("verify err=%v match=%v mismatch=%s", err, report.Match, report.Mismatch)
	}
}
