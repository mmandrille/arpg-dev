package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestShouldDeferNonCriticalSimOnlyWithinBudget(t *testing.T) {
	budget := time.Second / tickHz
	if shouldDeferNonCritical(30*time.Millisecond, 35*time.Millisecond, 0) {
		t.Fatal("within-budget sim should not defer")
	}
	if !shouldDeferNonCritical(budget+10*time.Millisecond, budget+15*time.Millisecond, 0) {
		t.Fatal("sim over budget should defer")
	}
}

func TestShouldDeferNonCriticalElapsedTotalOverBudget(t *testing.T) {
	budget := time.Second / tickHz
	sim := 30 * time.Millisecond
	elapsed := budget + 5*time.Millisecond
	if !shouldDeferNonCritical(sim, elapsed, 0) {
		t.Fatal("total elapsed over budget should defer even when sim is within budget")
	}
}

func TestShouldDeferNonCriticalPersistPhaseOverBudget(t *testing.T) {
	budget := time.Second / tickHz
	if !shouldDeferNonCritical(20*time.Millisecond, 40*time.Millisecond, budget+5*time.Millisecond) {
		t.Fatal("persist phase over budget should defer")
	}
}

func TestAppendSessionEventsBatch(t *testing.T) {
	repo := &progressionPersistRepo{}
	loop := &sessionLoop{
		hub:  &Hub{store: repo},
		sess: store.Session{ID: "sess_batch_events"},
	}
	events := []game.Event{
		{EventType: "skill_cast", SkillID: "volley"},
		{EventType: "monster_damaged", SkillID: "volley"},
		{EventType: "monster_damaged", SkillID: "volley"},
	}
	seq := loop.appendSessionEvents(context.Background(), game.TickResult{Tick: 7}, events, 0)
	if seq != 3 {
		t.Fatalf("sequence = %d, want 3", seq)
	}
	if len(repo.events) != 3 {
		t.Fatalf("persisted events = %d, want 3", len(repo.events))
	}
	if repo.events[0].Sequence != 0 || repo.events[2].Sequence != 2 {
		t.Fatalf("event sequences = %d,%d,%d, want 0,1,2", repo.events[0].Sequence, repo.events[1].Sequence, repo.events[2].Sequence)
	}
}
