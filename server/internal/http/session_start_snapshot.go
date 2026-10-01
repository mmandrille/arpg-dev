package httpapi

import (
	"context"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) createSessionStartSnapshot(w http.ResponseWriter, ctx context.Context, sessionID, accountID, characterID string) bool {
	items, err := s.store.ListCharacterItems(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character items")
		return false
	}
	character, err := s.store.GetCharacter(ctx, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character")
		return false
	}
	progression, err := s.store.GetOrCreateCharacterProgression(ctx, accountID, characterID, progressionDefaultsFromRules(s.rules, character.CharacterClass))
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character progression")
		return false
	}
	progression.CharacterClass = character.CharacterClass
	waypoints, err := s.store.ListAccountWaypoints(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load account waypoints")
		return false
	}
	hotbar, err := s.store.ListCharacterHotbar(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character hotbar")
		return false
	}
	skillBinds, err := s.store.GetOrCreateCharacterSkillBindings(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character skill bindings")
		return false
	}
	shopStock, err := s.store.ListCharacterShopStock(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character shop stock")
		return false
	}
	stashItems, err := s.store.ListAccountStashItems(ctx, accountID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load account stash")
		return false
	}
	stashGold, err := s.store.GetOrCreateAccountStashGold(ctx, accountID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load account stash gold")
		return false
	}
	resources, err := s.store.ListAccountResources(ctx, accountID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load account resources")
		return false
	}
	if err := s.store.MigrateCharacterResourceItemsToResourceBag(ctx, accountID, characterID); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not migrate resource items")
		return false
	}
	resourceBagItems, err := s.store.ListAccountResourceBagItems(ctx, accountID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load account resource bag")
		return false
	}
	corpses, err := s.store.ListRecoverableCharacterCorpses(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load character corpses")
		return false
	}
	mercenaryRoster, err := s.sessionMercenaryRosterSnapshot(ctx, accountID, characterID)
	if err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load mercenary roster")
		return false
	}
	if err := s.store.CreateSessionStartSnapshot(ctx, store.SessionStartSnapshot{
		SessionID:        sessionID,
		AccountID:        accountID,
		CharacterID:      characterID,
		Items:            items,
		Waypoints:        waypoints,
		Hotbar:           hotbar,
		SkillBinds:       skillBinds,
		ShopStock:        shopStock,
		StashItems:       stashItems,
		StashGold:        stashGold,
		Resources:        resources,
		ResourceBagItems: resourceBagItems,
		Corpses:          corpses,
		MercenaryRoster:  mercenaryRoster,
		Progression:      &progression,
	}); err != nil {
		s.metrics.PersistenceErrors.Inc()
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create session start snapshot")
		return false
	}
	return true
}
