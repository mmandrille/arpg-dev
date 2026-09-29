package sessionsetup

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type setupRepo struct {
	store.Repository
	members []store.SessionMember
	starts  map[string]store.SessionStartSnapshot
	classes map[string]string
}

func (r *setupRepo) ListSessionMembers(context.Context, string) ([]store.SessionMember, error) {
	return r.members, nil
}

func (r *setupRepo) LoadSessionStartSnapshotForMember(_ context.Context, _, _, characterID string) (store.SessionStartSnapshot, error) {
	return r.starts[characterID], nil
}

func (r *setupRepo) GetCharacter(_ context.Context, characterID string) (store.Character, error) {
	return store.Character{ID: characterID, CharacterClass: r.classes[characterID]}, nil
}

func (r *setupRepo) ListCharacters(context.Context, string) ([]store.CharacterSummary, error) {
	return nil, nil
}

func loadRules(t *testing.T) *game.Rules {
	t.Helper()
	dir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("find rules: %v", err)
	}
	rules, err := game.LoadRules(dir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	return rules
}

func TestMembersSynthesizesLegacyHostAndSortsHostFirst(t *testing.T) {
	sess := store.Session{ID: "sess", AccountID: "acct_host", CharacterID: "char_host"}
	legacy, err := Members(context.Background(), &setupRepo{}, sess)
	if err != nil || len(legacy) != 1 || legacy[0].Role != store.SessionMemberHost || legacy[0].CharacterID != "char_host" {
		t.Fatalf("legacy members = %+v err=%v, want synthesized host", legacy, err)
	}
	repo := &setupRepo{members: []store.SessionMember{
		{AccountID: "b", CharacterID: "late", Role: store.SessionMemberGuest, JoinedTick: 9},
		{AccountID: "a", CharacterID: "early", Role: store.SessionMemberGuest, JoinedTick: 0},
		{AccountID: "acct_host", CharacterID: "char_host", Role: store.SessionMemberHost},
	}}
	members, _ := Members(context.Background(), repo, sess)
	if got := [3]string{members[0].CharacterID, members[1].CharacterID, members[2].CharacterID}; got != [3]string{"char_host", "early", "late"} {
		t.Fatalf("member order = %v", got)
	}
	if Host(members).CharacterID != "char_host" {
		t.Fatalf("host = %+v", Host(members))
	}
}

func TestResolveAppliesCharacterClassOverLegacySnapshotClass(t *testing.T) {
	rules := loadRules(t)
	repo := &setupRepo{
		starts:  map[string]store.SessionStartSnapshot{"char": {Progression: &store.CharacterProgression{CharacterClass: "barbarian", Level: 4}}},
		classes: map[string]string{"char": "sorcerer"},
	}
	member, err := Resolve(context.Background(), repo, rules, "sess", store.SessionMember{CharacterID: "char"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if member.Progression.CharacterClass != "sorcerer" || member.Progression.Level != 4 {
		t.Fatalf("progression = %+v, want sorcerer level 4", member.Progression)
	}
}

// AddGuest must load everything that moves the entity ID allocator, and the
// shared converter must keep WeaponSet (replay's old copy dropped it).
func TestAddGuestLoadsResourceBagCorpsesAndWeaponSets(t *testing.T) {
	rules := loadRules(t)
	sim, err := game.NewSimWithWorldProgression("sess", "seed", rules, "dungeon_levels", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	const bagID, corpseItemID = 770001, 770500
	guest := Member{
		Member:      store.SessionMember{AccountID: "acct_guest", CharacterID: "char_guest", Role: store.SessionMemberGuest},
		Progression: rules.DefaultCharacterProgressionState(),
		Start: store.SessionStartSnapshot{
			ResourceBagItems: []store.AccountResourceBagItem{{BagItemID: "770001", ItemDefID: "renew_stone", RolledStats: json.RawMessage(`{}`)}},
			Corpses: []store.CharacterCorpse{{CharacterID: "char_dead", Name: "Fallen", Level: 2, DeathLevel: -1, Items: []store.CharacterItemInstance{
				{ID: "770500", ItemDefID: "rusty_sword", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{}`)},
			}}},
		},
	}
	playerID, err := AddGuest(sim, guest)
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	if bag := sim.SnapshotForPlayer(playerID).ResourceBagItems; len(bag) != 1 {
		t.Fatalf("guest resource bag = %+v, want 1 item", bag)
	}
	next, err := sim.AddGuestPlayer("acct_probe", "char_probe", "Probe", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add probe: %v", err)
	}
	if next <= corpseItemID || next <= bagID {
		t.Fatalf("next entity id %d did not advance past persisted bag/corpse ids", next)
	}

	items := persistedItems([]store.CharacterItemInstance{{ID: "1", ItemDefID: "rusty_sword", Location: store.ItemLocationEquipped, WeaponSet: 1}})
	if len(items) != 1 || items[0].WeaponSet != 1 {
		t.Fatalf("persisted items = %+v, want WeaponSet kept", items)
	}
}
