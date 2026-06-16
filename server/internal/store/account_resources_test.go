package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestAccountResourceWalletAddSpendAndScope(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	acct, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "wallet+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert account: %v", err)
	}
	other, err := s.UpsertAccountByEmail(ctx, ids.New("acct"), "wallet-other+"+ids.Token()[:12]+"@example.test")
	if err != nil {
		t.Fatalf("upsert other account: %v", err)
	}

	added, err := s.AddAccountResource(ctx, acct.ID, "upgrade_shard", 2)
	if err != nil {
		t.Fatalf("add resource: %v", err)
	}
	if added.Amount != 2 {
		t.Fatalf("first resource amount = %d, want 2", added.Amount)
	}
	added, err = s.AddAccountResource(ctx, acct.ID, "upgrade_shard", 1)
	if err != nil {
		t.Fatalf("add resource again: %v", err)
	}
	if added.Amount != 3 {
		t.Fatalf("second resource amount = %d, want 3", added.Amount)
	}
	if _, err := s.AddAccountResource(ctx, acct.ID, "upgrade_shard", 0); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("add zero resource = %v, want ErrConflict", err)
	}

	spent, err := s.SpendAccountResource(ctx, acct.ID, "upgrade_shard", 2)
	if err != nil {
		t.Fatalf("spend resource: %v", err)
	}
	if spent.Amount != 1 {
		t.Fatalf("spent resource amount = %d, want 1", spent.Amount)
	}
	if _, err := s.SpendAccountResource(ctx, acct.ID, "upgrade_shard", 2); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("overspend resource = %v, want ErrConflict", err)
	}
	if _, err := s.SpendAccountResource(ctx, other.ID, "upgrade_shard", 1); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("spend foreign empty resource = %v, want ErrConflict", err)
	}

	resources, err := s.ListAccountResources(ctx, acct.ID)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	if len(resources) != 1 || resources[0].ResourceID != "upgrade_shard" || resources[0].Amount != 1 {
		t.Fatalf("resources = %+v, want one upgrade_shard amount 1", resources)
	}
	otherResources, err := s.ListAccountResources(ctx, other.ID)
	if err != nil {
		t.Fatalf("list other resources: %v", err)
	}
	if len(otherResources) != 0 {
		t.Fatalf("other resources = %+v, want none", otherResources)
	}
}
