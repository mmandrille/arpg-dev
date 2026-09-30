// Package sessionsetup puts session members into a Sim from their frozen
// session-start snapshots. The live session build, live late co-op joins, and
// replay reconstruction all go through it, so member setup (and therefore
// entity ID allocation) cannot drift between live and replay (v478).
package sessionsetup

import (
	"context"
	"fmt"
	"sort"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/mercenaryroster"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// Member is a session member resolved against durable storage: its frozen
// start snapshot and the progression the sim is built with. Resolving does all
// I/O up front, so AddGuest can run under the live loop's sim lock.
type Member struct {
	Member      store.SessionMember
	Start       store.SessionStartSnapshot
	Progression game.CharacterProgressionState
}

// Members lists the session's members host-first, then by join tick. Legacy
// sessions without member rows get a synthesized host.
func Members(ctx context.Context, repo store.Repository, sess store.Session) ([]store.SessionMember, error) {
	members, err := repo.ListSessionMembers(ctx, sess.ID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	if len(members) == 0 {
		members = []store.SessionMember{{
			SessionID:   sess.ID,
			AccountID:   sess.AccountID,
			CharacterID: sess.CharacterID,
			Role:        store.SessionMemberHost,
			Status:      store.SessionMemberActive,
		}}
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].Role != members[j].Role {
			return members[i].Role == store.SessionMemberHost
		}
		if members[i].JoinedTick != members[j].JoinedTick {
			return members[i].JoinedTick < members[j].JoinedTick
		}
		if members[i].AccountID != members[j].AccountID {
			return members[i].AccountID < members[j].AccountID
		}
		return members[i].CharacterID < members[j].CharacterID
	})
	return members, nil
}

// Host returns the host member of a Members-sorted list.
func Host(members []store.SessionMember) store.SessionMember {
	for _, member := range members {
		if member.Role == store.SessionMemberHost {
			return member
		}
	}
	return members[0]
}

// JoinedSim reports whether the member's player entity was ever put into the
// sim. A negative joined_tick means it never was (for example, a guest that
// joined over HTTP while the loop was running and never attached), so replay
// must not allocate an entity for it.
func JoinedSim(member store.SessionMember) bool {
	return member.JoinedTick >= 0
}

// IsHost reports whether member is the same account/character as host.
func IsHost(member, host store.SessionMember) bool {
	return member.AccountID == host.AccountID && member.CharacterID == host.CharacterID
}

// Resolve loads a member's session-start snapshot and progression.
func Resolve(ctx context.Context, repo store.Repository, rules *game.Rules, sessionID string, member store.SessionMember) (Member, error) {
	start, err := repo.LoadSessionStartSnapshotForMember(ctx, sessionID, member.AccountID, member.CharacterID)
	if err != nil {
		return Member{}, fmt.Errorf("load start snapshot for %s/%s: %w", member.AccountID, member.CharacterID, err)
	}
	progression, err := progressionForMember(ctx, repo, rules, member, start.Progression)
	if err != nil {
		return Member{}, err
	}
	return Member{Member: member, Start: start, Progression: progression}, nil
}

// progressionForMember prefers the character row's class over the snapshot's.
// Snapshots taken before v70 stored a default class; the class of a character
// never changes afterwards, so reading it at replay time is reproducible.
func progressionForMember(ctx context.Context, repo store.Repository, rules *game.Rules, member store.SessionMember, progression *store.CharacterProgression) (game.CharacterProgressionState, error) {
	if progression == nil {
		return progressionState(rules, nil), nil
	}
	character, err := repo.GetCharacter(ctx, member.CharacterID)
	if err != nil {
		return game.CharacterProgressionState{}, fmt.Errorf("load character class: %w", err)
	}
	if character.CharacterClass == "" || character.CharacterClass == progression.CharacterClass {
		return progressionState(rules, progression), nil
	}
	updated := *progression
	updated.CharacterClass = character.CharacterClass
	return progressionState(rules, &updated), nil
}

// NewHostSim builds the session sim around the host: start state, mercenary
// roster, the restored hired companion, then the host's corpses. The order is
// part of the replay contract because the companion and corpses allocate IDs.
//
// configure, when non-nil, runs right after construction and before any member
// state loads. Live uses it for SetGameplayDebug: enabling debug seeds unique
// test chests, and doing that after the per-player loads re-seeds them and
// allocates IDs that replay never does.
func NewHostSim(ctx context.Context, repo store.Repository, rules *game.Rules, sess store.Session, host Member, configure func(*game.Sim)) (*game.Sim, error) {
	sim, err := game.NewSimWithWorldProgression(sess.ID, sess.Seed, rules, WorldID(sess.WorldID), host.Progression)
	if err != nil {
		return nil, err
	}
	if configure != nil {
		configure(sim)
	}
	hostID := sim.DefaultPlayerID()
	sim.SetPlayerMetadata(hostID, host.Member.AccountID, host.Member.CharacterID, DisplayName(host.Member), store.SessionMemberHost)
	loadMemberState(sim, hostID, host.Start)
	if err := mercenaryroster.LoadIntoSim(ctx, repo, rules, sim, host.Member.AccountID, host.Member.CharacterID); err != nil {
		return nil, err
	}
	sim.RestoreHiredMercenaryCompanion(hostID)
	loadMemberCorpses(sim, host)
	return sim, nil
}

// AddGuest adds a co-op guest with its start state and corpses. It does no
// I/O, so the live loop may call it while holding the sim lock.
func AddGuest(sim *game.Sim, guest Member) (uint64, error) {
	playerID, err := sim.AddGuestPlayer(guest.Member.AccountID, guest.Member.CharacterID, DisplayName(guest.Member), guest.Progression)
	if err != nil {
		return 0, err
	}
	loadMemberState(sim, playerID, guest.Start)
	loadMemberCorpses(sim, guest)
	return playerID, nil
}

func loadMemberState(sim *game.Sim, playerID uint64, start store.SessionStartSnapshot) {
	sim.LoadInventoryForPlayer(playerID, persistedItems(start.Items))
	sim.LoadHotbarForPlayer(playerID, persistedHotbar(start.Hotbar))
	sim.LoadSkillBindingsForPlayer(playerID, persistedSkillBindings(start.SkillBinds))
	sim.LoadDiscoveredTeleportersForPlayer(playerID, waypointLevels(start.Waypoints))
	sim.LoadShopStockForPlayer(playerID, persistedShopStock(start.ShopStock))
	sim.LoadAccountStashForPlayer(playerID, PersistedStashItems(start.StashItems), start.StashGold.Gold, 0)
	sim.LoadResourceWalletForPlayer(playerID, persistedResources(start.Resources))
	sim.LoadAccountResourceBagForPlayer(playerID, persistedResourceBagItems(start.ResourceBagItems))
}

func loadMemberCorpses(sim *game.Sim, member Member) {
	sim.LoadCharacterCorpses(persistedCorpses(member.Member.AccountID, member.Start.Corpses))
}

// DisplayName is the player name shown for a member.
func DisplayName(member store.SessionMember) string {
	if member.Role == store.SessionMemberHost {
		return "Hero"
	}
	if member.CharacterID == "" {
		return "Guest"
	}
	suffix := member.CharacterID
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	return "Guest " + suffix
}

// WorldID defaults an unset session world to the dungeon world.
func WorldID(worldID string) string {
	if worldID == "" {
		return game.DefaultWorldID
	}
	return worldID
}
