package game

import (
	"fmt"
	"path/filepath"
)

type MercenaryRules struct {
	Offers []MercenaryOfferDef
}

type MercenaryOfferDef struct {
	OfferID      string `json:"offer_id"`
	MonsterDefID string `json:"monster_def_id"`
}

func (r *Rules) loadMercenaryRules(dir string) error {
	var file struct {
		Version int                 `json:"version"`
		Offers  []MercenaryOfferDef `json:"offers"`
	}
	if err := readJSON(filepath.Join(dir, "mercenaries.v0.json"), &file); err != nil {
		return err
	}
	if file.Version != 0 {
		return fmt.Errorf("game: invalid rules mercenaries.version: %d", file.Version)
	}
	if err := validateMercenaryOffers(file.Offers, r.Monsters); err != nil {
		return err
	}
	r.Mercenaries = MercenaryRules{Offers: append([]MercenaryOfferDef(nil), file.Offers...)}
	return nil
}

func validateMercenaryOffers(offers []MercenaryOfferDef, monsters map[string]MonsterDef) error {
	if len(offers) == 0 {
		return fmt.Errorf("game: invalid rules mercenaries.offers: at least one offer is required")
	}
	seen := map[string]bool{}
	for idx, offer := range offers {
		if offer.OfferID == "" {
			return fmt.Errorf("game: invalid rules mercenaries.offers[%d].offer_id: required", idx)
		}
		if seen[offer.OfferID] {
			return fmt.Errorf("game: invalid rules mercenaries.offers[%d].offer_id: duplicate %s", idx, offer.OfferID)
		}
		seen[offer.OfferID] = true
		if offer.MonsterDefID == "" {
			return fmt.Errorf("game: invalid rules mercenaries.offers[%d].monster_def_id: required", idx)
		}
		if _, ok := monsters[offer.MonsterDefID]; !ok {
			return fmt.Errorf("game: invalid rules mercenaries.offers[%d].monster_def_id: unknown monster %s", idx, offer.MonsterDefID)
		}
	}
	return nil
}

func (r MercenaryRules) SelectOffer(seed string, boardEntityID string) (MercenaryOfferDef, bool) {
	if len(r.Offers) == 0 {
		return MercenaryOfferDef{}, false
	}
	if len(r.Offers) == 1 {
		return r.Offers[0], true
	}
	rng := NewRNG(SeedToUint64(seed + "|mercenary_offer_selection|" + boardEntityID))
	return r.Offers[rng.IntN(len(r.Offers))], true
}

func (r MercenaryRules) OfferByMonsterDefID(monsterDefID string) (MercenaryOfferDef, bool) {
	for _, offer := range r.Offers {
		if offer.MonsterDefID == monsterDefID {
			return offer, true
		}
	}
	return MercenaryOfferDef{}, false
}
