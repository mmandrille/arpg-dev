package store_test

import (
	"context"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestReviveDeadCharactersIsAccountScoped(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	suffix := ids.Token()[:12]
	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "revive+"+suffix+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	otherAcct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "revive-other+"+suffix+"@example.test")
	if err != nil {
		t.Fatalf("upsert other account: %v", err)
	}
	alive, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Alive", "barbarian")
	if err != nil {
		t.Fatalf("create alive: %v", err)
	}
	deadA, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Dead A", "sorcerer")
	if err != nil {
		t.Fatalf("create dead a: %v", err)
	}
	deadB, err := s.CreateCharacter(ctx, ids.New("char"), acct.ID, "Dead B", "ranger")
	if err != nil {
		t.Fatalf("create dead b: %v", err)
	}
	otherDead, err := s.CreateCharacter(ctx, ids.New("char"), otherAcct.ID, "Other Dead", "paladin")
	if err != nil {
		t.Fatalf("create other dead: %v", err)
	}
	for _, c := range []store.Character{deadA, deadB} {
		if err := s.MarkCharacterDead(ctx, acct.ID, c.ID, -3); err != nil {
			t.Fatalf("mark %s dead: %v", c.ID, err)
		}
	}
	if err := s.MarkCharacterDead(ctx, otherAcct.ID, otherDead.ID, -4); err != nil {
		t.Fatalf("mark other dead: %v", err)
	}

	revived, err := s.ReviveDeadCharacters(ctx, acct.ID)
	if err != nil {
		t.Fatalf("revive dead characters: %v", err)
	}
	if revived != 2 {
		t.Fatalf("revived = %d, want 2", revived)
	}
	for _, c := range []store.Character{alive, deadA, deadB} {
		got, err := s.GetCharacter(ctx, c.ID)
		if err != nil {
			t.Fatalf("get character %s: %v", c.ID, err)
		}
		if got.Dead || got.DeathLevel != nil {
			t.Fatalf("character %s after revive = dead %v death_level %v", c.ID, got.Dead, got.DeathLevel)
		}
	}
	other, err := s.GetCharacter(ctx, otherDead.ID)
	if err != nil {
		t.Fatalf("get other dead: %v", err)
	}
	if !other.Dead || other.DeathLevel == nil || *other.DeathLevel != -4 {
		t.Fatalf("other account character after revive = %+v", other)
	}
}
