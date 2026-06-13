package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const (
	testDevToken   = "test-dev-token"
	testDebugToken = "test-debug-token"
)

// fullServer builds a server backed by real Postgres, or skips if unreachable.
func fullServer(t *testing.T) http.Handler {
	return fullServerWithConfig(t, config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, MetricsEnabled: true})
}

func fullServerWithConfig(t *testing.T, cfg config.Config) http.Handler {
	h, _ := fullServerWithConfigAndStore(t, cfg)
	return h
}

func fullServerWithStore(t *testing.T) (http.Handler, *store.Store) {
	return fullServerWithConfigAndStore(t, config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, MetricsEnabled: true})
}

func fullServerWithConfigAndStore(t *testing.T, cfg config.Config) (http.Handler, *store.Store) {
	t.Helper()
	url := "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := store.Connect(ctx, url)
	if err != nil {
		t.Skipf("skipping auth/session test: no Postgres: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(db.Close)
	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		t.Fatalf("rules dir: %v", err)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	h := New(Deps{
		Config:  cfg,
		Logger:  logging.NewTo(io.Discard, "local"),
		Metrics: metrics.New(),
		Store:   db,
		Auth:    auth.NewService(testDevToken, db, db),
		Rules:   rules,
		Ready:   db.Ping,
	}).Handler()
	return h, db
}

func postJSON(h http.Handler, path, bearer string, body any) *httptest.ResponseRecorder {
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func getJSON(h http.Handler, path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func deleteJSON(h http.Handler, path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func putDebugJSON(h http.Handler, path, bearer string, body any, debugToken string) *httptest.ResponseRecorder {
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if debugToken != "" {
		req.Header.Set("X-Debug-Token", debugToken)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func login(t *testing.T, h http.Handler) (accountID, token string) {
	return loginEmail(t, h, testEmail(t, "dev"))
}

func loginEmail(t *testing.T, h http.Handler, email string) (accountID, token string) {
	t.Helper()
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": email, "dev_token": testDevToken,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res devLoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if res.AccessToken == "" || res.AccountID == "" {
		t.Fatalf("login missing token/account: %+v", res)
	}
	t.Cleanup(func() {
		rec := getJSON(h, "/v0/characters", res.AccessToken)
		if rec.Code != http.StatusOK {
			t.Logf("cleanup list characters for %s status = %d, body = %s", email, rec.Code, rec.Body.String())
			return
		}
		var listed listCharactersResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
			t.Logf("cleanup decode characters for %s: %v", email, err)
			return
		}
		for _, char := range listed.Characters {
			rec := deleteJSON(h, "/v0/characters/"+char.CharacterID, res.AccessToken)
			if rec.Code != http.StatusNoContent && rec.Code != http.StatusNotFound {
				t.Logf("cleanup delete character %s for %s status = %d, body = %s", char.CharacterID, email, rec.Code, rec.Body.String())
			}
		}
	})
	return res.AccountID, res.AccessToken
}

func testEmail(t *testing.T, label string) string {
	t.Helper()
	return label + "+" + ids.Token()[:12] + "@example.test"
}

func TestDevLoginReturnsNormalizedAccountEmail(t *testing.T) {
	h := fullServer(t)
	email := "client-normalized+" + ids.Token()[:12] + "@mail.test"
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": "  " + strings.ToUpper(email) + "  ", "dev_token": testDevToken,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res devLoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if res.Email != email {
		t.Fatalf("login email = %q, want normalized account email", res.Email)
	}
	if res.AccountID == "" || res.AccessToken == "" || res.ExpiresAt == "" {
		t.Fatalf("login response missing identity fields: %+v", res)
	}
}

func createCharacter(t *testing.T, h http.Handler, token, name string) characterResponse {
	t.Helper()
	return createCharacterWithClass(t, h, token, name, "")
}

func createCharacterWithClass(t *testing.T, h http.Handler, token, name, characterClass string) characterResponse {
	t.Helper()
	body := map[string]string{"name": name}
	if characterClass != "" {
		body["character_class"] = characterClass
	}
	rec := postJSON(h, "/v0/characters", token, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create character status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res characterResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode character: %v", err)
	}
	if res.CharacterID == "" || res.Name == "" || res.CharacterClass == "" || res.CreatedAt == "" {
		t.Fatalf("incomplete character response: %+v", res)
	}
	return res
}

func TestDevLoginInvalidToken(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": "dev@example.test", "dev_token": "wrong",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestDevLoginInvalidEmail(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/auth/dev-login", "", map[string]string{
		"email": "not-an-email", "dev_token": testDevToken,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCreateSessionRequiresAuth(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/sessions", "", map[string]any{"mode": "solo"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestCreateSessionInvalidToken(t *testing.T) {
	h := fullServer(t)
	rec := postJSON(h, "/v0/sessions", "garbage-token", map[string]any{"mode": "solo"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestCharacterAPIRequiresAuth(t *testing.T) {
	h := fullServer(t)

	rec := getJSON(h, "/v0/characters", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("list status = %d, want 401", rec.Code)
	}
	rec = postJSON(h, "/v0/characters", "", map[string]string{"name": "Mara"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("create status = %d, want 401", rec.Code)
	}
	rec = deleteJSON(h, "/v0/characters/char_missing", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("delete status = %d, want 401", rec.Code)
	}
}

func TestCreateCharacterValidationAndList(t *testing.T) {
	h := fullServer(t)
	_, token := loginEmail(t, h, testEmail(t, "characters-validation"))

	rec := postJSON(h, "/v0/characters", token, map[string]string{"name": "   "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty name status = %d, want 400", rec.Code)
	}
	rec = postJSON(h, "/v0/characters", token, map[string]string{"name": strings.Repeat("x", 25)})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("long name status = %d, want 400", rec.Code)
	}
	rec = postJSON(h, "/v0/characters", token, map[string]string{"name": "Bad Class", "character_class": "necromancer"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid class status = %d, want 400", rec.Code)
	}

	first := createCharacter(t, h, token, "  Mara  ")
	if first.Name != "Mara" {
		t.Fatalf("trimmed name = %q, want Mara", first.Name)
	}
	if first.CharacterClass != "barbarian" {
		t.Fatalf("default character_class = %q, want barbarian", first.CharacterClass)
	}
	if first.Level != 1 || first.Gold != 0 || first.DeepestDungeonDepth != 0 {
		t.Fatalf("new character summary = %+v, want level 1 gold 0 depth 0", first)
	}
	second := createCharacterWithClass(t, h, token, "Mara", "sorcerer")
	if second.Name != "Mara" || second.CharacterID == first.CharacterID {
		t.Fatalf("duplicate character not created independently: first=%+v second=%+v", first, second)
	}
	if second.CharacterClass != "sorcerer" {
		t.Fatalf("explicit character_class = %q, want sorcerer", second.CharacterClass)
	}

	rec = getJSON(h, "/v0/characters", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	var found int
	for _, c := range listed.Characters {
		if c.CharacterID == first.CharacterID || c.CharacterID == second.CharacterID {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("listed characters missing created rows: %+v", listed.Characters)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == first.CharacterID && (c.Level != 1 || c.Gold != 0 || c.DeepestDungeonDepth != 0) {
			t.Fatalf("default listed summary = %+v, want level 1 gold 0 depth 0", c)
		}
		if c.CharacterID == second.CharacterID && c.CharacterClass != "sorcerer" {
			t.Fatalf("listed explicit class = %+v, want sorcerer", c)
		}
	}
}

func TestCharacterClassSeedsSessionStartProgression(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	accountID, token := loginEmail(t, h, "characters-class-stats+"+ids.Token()[:12]+"@example.test")
	barbarian := createCharacterWithClass(t, h, token, "Class Barbarian", "barbarian")
	sorcerer := createCharacterWithClass(t, h, token, "Class Sorcerer", "sorcerer")

	rec := postJSON(h, "/v0/sessions", token, map[string]string{"mode": "solo", "world_id": "vertical_slice", "character_id": barbarian.CharacterID})
	if rec.Code != http.StatusCreated {
		t.Fatalf("barbarian session status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions", token, map[string]string{"mode": "solo", "world_id": "vertical_slice", "character_id": sorcerer.CharacterID})
	if rec.Code != http.StatusCreated {
		t.Fatalf("sorcerer session status = %d, body = %s", rec.Code, rec.Body.String())
	}

	barbProgression, err := db.GetCharacterProgression(ctx, accountID, barbarian.CharacterID)
	if err != nil {
		t.Fatalf("barbarian progression: %v", err)
	}
	sorcProgression, err := db.GetCharacterProgression(ctx, accountID, sorcerer.CharacterID)
	if err != nil {
		t.Fatalf("sorcerer progression: %v", err)
	}
	if barbProgression.Stats.Str != 5 || barbProgression.Stats.Vit != 5 || barbProgression.Stats.Magic != 5 {
		t.Fatalf("barbarian stats = %+v, want class start", barbProgression.Stats)
	}
	if sorcProgression.Stats.Str != 3 || sorcProgression.Stats.Magic != 5 {
		t.Fatalf("sorcerer stats = %+v, want class start", sorcProgression.Stats)
	}
}

func TestDebugCharacterProgressionSeedsOwnedCharacter(t *testing.T) {
	h, db := fullServerWithConfigAndStore(t, config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, DebugToken: testDebugToken})
	ctx := context.Background()
	accountID, token := loginEmail(t, h, "debug-progression+"+ids.Token()[:12]+"@example.test")
	hero := createCharacterWithClass(t, h, token, "Debug Paladin", "paladin")

	rec := putDebugJSON(h, "/v0/debug/characters/"+hero.CharacterID+"/progression", token, map[string]any{
		"level":                5,
		"experience":           0,
		"unspent_stat_points":  0,
		"unspent_skill_points": 0,
		"stats":                map[string]int{"str": 6, "dex": 4, "vit": 13, "magic": 13},
		"skill_ranks":          map[string]int{"holy_shield": 5},
	}, testDebugToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("debug seed status = %d, body = %s", rec.Code, rec.Body.String())
	}

	progression, err := db.GetCharacterProgression(ctx, accountID, hero.CharacterID)
	if err != nil {
		t.Fatalf("load seeded progression: %v", err)
	}
	if progression.Level != 5 || progression.Experience != 0 || progression.UnspentSkillPoints != 0 {
		t.Fatalf("seeded progression = %+v, want level 5 with no XP/unspent skill points", progression)
	}
	if progression.Stats.Vit != 13 || progression.Stats.Magic != 13 || progression.SkillRanks["holy_shield"] != 5 {
		t.Fatalf("seeded stats/ranks = %+v ranks=%+v, want holy_shield rank 5 requirements", progression.Stats, progression.SkillRanks)
	}
}

func TestCharacterListIncludesProgressionSummariesAndDefaults(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	accountID, token := loginEmail(t, h, "characters-summary+"+ids.Token()[:12]+"@example.test")
	hero := createCharacter(t, h, token, "Summary Hero")
	fresh := createCharacter(t, h, token, "Fresh Hero")

	if err := db.UpsertCharacterProgression(ctx, accountID, store.CharacterProgression{
		CharacterID:         hero.CharacterID,
		Level:               7,
		Experience:          144,
		UnspentStatPoints:   2,
		UnspentSkillPoints:  1,
		Stats:               store.CharacterBaseStats{Str: 8, Dex: 7, Vit: 6, Magic: 5},
		Gold:                321,
		DeepestDungeonDepth: 4,
		SkillRanks:          map[string]int{"magic_bolt": 2},
	}); err != nil {
		t.Fatalf("upsert progression: %v", err)
	}

	otherAccountID, otherToken := loginEmail(t, h, "characters-summary-other+"+ids.Token()[:12]+"@example.test")
	otherHero := createCharacter(t, h, otherToken, "Other Hero")
	if err := db.UpsertCharacterProgression(ctx, otherAccountID, store.CharacterProgression{
		CharacterID:         otherHero.CharacterID,
		Level:               9,
		Stats:               store.CharacterBaseStats{Str: 9, Dex: 9, Vit: 9, Magic: 9},
		Gold:                999,
		DeepestDungeonDepth: 8,
		SkillRanks:          map[string]int{"magic_bolt": 1},
	}); err != nil {
		t.Fatalf("upsert other progression: %v", err)
	}

	rec := getJSON(h, "/v0/characters", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	var sawHero, sawFresh bool
	for _, c := range listed.Characters {
		switch c.CharacterID {
		case hero.CharacterID:
			sawHero = true
			if c.Level != 7 || c.Gold != 321 || c.DeepestDungeonDepth != 4 {
				t.Fatalf("progression summary = %+v, want level 7 gold 321 depth 4", c)
			}
		case fresh.CharacterID:
			sawFresh = true
			if c.Level != 1 || c.Gold != 0 || c.DeepestDungeonDepth != 0 {
				t.Fatalf("fresh summary = %+v, want defaults", c)
			}
		case otherHero.CharacterID:
			t.Fatalf("account-scoped list leaked other character summary: %+v", listed.Characters)
		}
	}
	if !sawHero || !sawFresh {
		t.Fatalf("listed summaries missing characters hero=%v fresh=%v rows=%+v", sawHero, sawFresh, listed.Characters)
	}
}

func TestMarketListingRoutesMoveStashItemAndRejectForeignCancel(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	accountID, token := loginEmail(t, h, "market-routes+"+ids.Token()[:12]+"@example.test")
	char := createCharacter(t, h, token, "Market Seller")
	otherAccountID, otherToken := loginEmail(t, h, "market-routes-other+"+ids.Token()[:12]+"@example.test")
	if otherAccountID == accountID {
		t.Fatal("expected distinct accounts")
	}
	item := store.CharacterItemInstance{
		ID:          "market_route_item",
		AccountID:   accountID,
		CharacterID: char.CharacterID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"damage_min":1}`),
	}
	if err := db.AddCharacterItem(ctx, item); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountID, char.CharacterID, item.ID, "stash_route_item"); err != nil {
		t.Fatal(err)
	}

	rec := postJSON(h, "/v0/market/listings", token, map[string]string{"stash_item_id": "stash_route_item"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ListingID == "" || created.Status != store.MarketListingActive || created.ItemDefID != "rusty_sword" {
		t.Fatalf("created listing = %+v", created)
	}
	stashItems, err := db.ListAccountStashItems(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stashItems) != 0 {
		t.Fatalf("stash after listing = %+v", stashItems)
	}

	rec = getJSON(h, "/v0/market/listings", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list listings status = %d body=%s", rec.Code, rec.Body.String())
	}
	var listed listMarketListingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Listings) == 0 || listed.Listings[0].ListingID != created.ListingID {
		t.Fatalf("listed market rows = %+v", listed.Listings)
	}

	rec = postJSON(h, "/v0/market/listings/"+created.ListingID+"/cancel", otherToken, map[string]string{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign cancel status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/market/listings/"+created.ListingID+"/cancel", token, map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var canceled marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &canceled); err != nil {
		t.Fatal(err)
	}
	if canceled.Status != store.MarketListingCanceled {
		t.Fatalf("canceled = %+v", canceled)
	}
	stashItems, err = db.ListAccountStashItems(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stashItems) != 1 || stashItems[0].StashItemID != "stash_route_item" {
		t.Fatalf("stash after cancel = %+v", stashItems)
	}
}

func TestMarketListingRouteAcceptsInventoryItem(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	accountID, token := loginEmail(t, h, "market-inventory-listing+"+suffix+"@example.test")
	char := createCharacter(t, h, token, "Inventory Market Seller")
	itemID := "market_inventory_listing_item_" + suffix
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          itemID,
		AccountID:   accountID,
		CharacterID: char.CharacterID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"damage_min":2}`),
	}); err != nil {
		t.Fatal(err)
	}

	rec := postJSON(h, "/v0/market/listings", token, map[string]any{
		"item_instance_id": itemID,
		"character_id":     char.CharacterID,
		"price_gold":       77,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create inventory listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.StashItemID == "" || created.ItemDefID != "rusty_sword" || created.PriceGold != 77 {
		t.Fatalf("inventory listing = %+v", created)
	}
	items, err := db.ListCharacterItems(ctx, accountID, char.CharacterID)
	if err != nil {
		t.Fatal(err)
	}
	if characterItemsContain(items, itemID) {
		t.Fatalf("listed item still on character after inventory listing = %+v", items)
	}
	stashItems, err := db.ListAccountStashItems(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stashItems) != 0 {
		t.Fatalf("stash should be reserved into listing, got %+v", stashItems)
	}
}

func TestAccountStashItemUpgradeRoute(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	accountID, token := loginEmail(t, h, "stash-upgrade+"+suffix+"@example.test")
	char := createCharacter(t, h, token, "Upgrade Route Hero")
	prog := store.CharacterProgression{AccountID: accountID, CharacterID: char.CharacterID, CharacterClass: "barbarian", Level: 1, Gold: 300, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := db.UpsertCharacterProgression(ctx, accountID, prog); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "route_upgrade_item_" + suffix, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountID, char.CharacterID, "route_upgrade_item_"+suffix, "route_upgrade_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.TransferCharacterGoldToAccountStash(ctx, accountID, char.CharacterID, 250); err != nil {
		t.Fatal(err)
	}
	rec := postJSON(h, "/v0/account-stash/items/route_upgrade_stash_"+suffix+"/upgrade", token, map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("upgrade status = %d body=%s", rec.Code, rec.Body.String())
	}
	var upgraded upgradeAccountStashItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &upgraded); err != nil {
		t.Fatal(err)
	}
	if upgraded.StashGold != 150 || upgraded.CostGold != 100 {
		t.Fatalf("upgrade balances = %+v", upgraded)
	}
	var stats map[string]int
	if err := json.Unmarshal(upgraded.Item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats["item_level"] != 1 || stats["damage_max"] != 5 {
		t.Fatalf("upgraded route stats = %+v", stats)
	}
	rec = postJSON(h, "/v0/account-stash/items/route_upgrade_stash_"+suffix+"/upgrade", token, map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("second upgrade status = %d body=%s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &upgraded); err != nil {
		t.Fatal(err)
	}
	if upgraded.StashGold != 0 || upgraded.CostGold != 150 {
		t.Fatalf("second upgrade balances = %+v", upgraded)
	}
	if err := json.Unmarshal(upgraded.Item.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats["item_level"] != 2 || stats["damage_max"] != 6 {
		t.Fatalf("second upgraded route stats = %+v", stats)
	}
	rec = postJSON(h, "/v0/account-stash/items/route_upgrade_stash_"+suffix+"/upgrade", token, map[string]string{})
	if rec.Code != http.StatusConflict {
		t.Fatalf("max-level upgrade status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountStashItemUpgradeRouteAcceptsInventoryItem(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	accountID, token := loginEmail(t, h, "inventory-upgrade+"+suffix+"@example.test")
	char := createCharacter(t, h, token, "Inventory Upgrade Hero")
	prog := store.CharacterProgression{AccountID: accountID, CharacterID: char.CharacterID, CharacterClass: "barbarian", Level: 1, Gold: 100, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}
	if err := db.UpsertCharacterProgression(ctx, accountID, prog); err != nil {
		t.Fatal(err)
	}
	itemID := "inventory_upgrade_item_" + suffix
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: accountID, CharacterID: char.CharacterID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":2,"damage_max":4}`)}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.TransferCharacterGoldToAccountStash(ctx, accountID, char.CharacterID, 40); err != nil {
		t.Fatal(err)
	}
	rec := postJSON(h, "/v0/account-stash/items/upgrade", token, map[string]string{
		"item_instance_id": itemID,
		"character_id":     char.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("inventory upgrade status = %d body=%s", rec.Code, rec.Body.String())
	}
	var upgraded upgradeInventoryItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &upgraded); err != nil {
		t.Fatal(err)
	}
	if upgraded.Item.ItemInstanceID != itemID || upgraded.Gold != 0 || upgraded.StashGold != 0 || upgraded.CostGold != 100 {
		t.Fatalf("inventory upgrade response = %+v", upgraded)
	}
	items, err := db.ListCharacterItems(ctx, accountID, char.CharacterID)
	if err != nil {
		t.Fatal(err)
	}
	if !characterItemsContain(items, itemID) {
		t.Fatalf("upgraded item missing from character after inventory upgrade = %+v", items)
	}
	var found store.CharacterItemInstance
	for _, item := range items {
		if item.ID == itemID {
			found = item
		}
	}
	var stats map[string]int
	if err := json.Unmarshal(found.RolledStats, &stats); err != nil {
		t.Fatal(err)
	}
	if stats["item_level"] != 1 || stats["damage_max"] != 5 {
		t.Fatalf("upgraded inventory stats = %+v raw=%s", stats, string(found.RolledStats))
	}
	stashItems, err := db.ListAccountStashItems(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range stashItems {
		if item.StashItemID == "upgrade_"+itemID {
			t.Fatalf("inventory upgrade left reserved stash item behind: %+v", stashItems)
		}
	}
}

func TestMarketOfferRoutesSubmitListAndAccept(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	sellerID, sellerToken := loginEmail(t, h, "market-offer-seller+"+suffix+"@example.test")
	sellerChar := createCharacter(t, h, sellerToken, "Offer Seller")
	bidderID, bidderToken := loginEmail(t, h, "market-offer-bidder+"+suffix+"@example.test")
	bidderChar := createCharacter(t, h, bidderToken, "Offer Bidder")
	foreignID, foreignToken := loginEmail(t, h, "market-offer-foreign+"+suffix+"@example.test")
	if sellerID == bidderID || bidderID == foreignID || sellerID == foreignID {
		t.Fatal("expected distinct market route accounts")
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          "market_offer_listing_item_" + suffix,
		AccountID:   sellerID,
		CharacterID: sellerChar.CharacterID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"damage_min":2}`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, sellerID, sellerChar.CharacterID, "market_offer_listing_item_"+suffix, "market_offer_listing_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	for _, itemID := range []string{"market_offer_bid_item_a_" + suffix, "market_offer_bid_item_b_" + suffix} {
		if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: bidderID, CharacterID: bidderChar.CharacterID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
			t.Fatal(err)
		}
		if _, err := db.TransferCharacterItemToAccountStash(ctx, bidderID, bidderChar.CharacterID, itemID, "stash_"+itemID); err != nil {
			t.Fatal(err)
		}
	}

	rec := postJSON(h, "/v0/market/listings", sellerToken, map[string]string{"stash_item_id": "market_offer_listing_stash_" + suffix})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var listing marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", sellerToken, createMarketOfferRequest{StashItemIDs: []string{"market_offer_listing_stash_" + suffix}})
	if rec.Code != http.StatusConflict {
		t.Fatalf("self offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", bidderToken, createMarketOfferRequest{StashItemIDs: []string{"stash_market_offer_bid_item_a_" + suffix, "stash_market_offer_bid_item_b_" + suffix}})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	var offer marketOfferResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &offer); err != nil {
		t.Fatal(err)
	}
	if offer.OfferID == "" || offer.Status != store.MarketOfferActive || len(offer.Items) != 2 {
		t.Fatalf("created offer = %+v", offer)
	}
	rec = getJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", foreignToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign list offers status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = getJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", sellerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("seller list offers status = %d body=%s", rec.Code, rec.Body.String())
	}
	var offers listMarketOffersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &offers); err != nil {
		t.Fatal(err)
	}
	if len(offers.Offers) != 1 || offers.Offers[0].OfferID != offer.OfferID {
		t.Fatalf("listed offers = %+v", offers.Offers)
	}
	rec = getJSON(h, "/v0/market/summary", sellerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("seller market summary status = %d body=%s", rec.Code, rec.Body.String())
	}
	var sellerSummary marketSummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &sellerSummary); err != nil {
		t.Fatal(err)
	}
	if sellerSummary.PublishedListings != 1 || sellerSummary.IncomingBids != 1 {
		t.Fatalf("seller market summary = %+v, want 1 listing and 1 bid", sellerSummary)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers/"+offer.OfferID+"/accept", foreignToken, map[string]string{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign accept status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers/"+offer.OfferID+"/accept", sellerToken, map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("accept offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	var accepted marketOfferResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted.Status != store.MarketOfferAccepted {
		t.Fatalf("accepted offer = %+v", accepted)
	}
	sellerStash, err := db.ListAccountStashItems(ctx, sellerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sellerStash) != 2 {
		t.Fatalf("seller stash after accept = %+v", sellerStash)
	}
	bidderStash, err := db.ListAccountStashItems(ctx, bidderID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "market_offer_listing_stash_"+suffix {
		t.Fatalf("bidder stash after accept = %+v", bidderStash)
	}
}

func TestMarketOfferRouteAcceptsInventoryItems(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	sellerID, sellerToken := loginEmail(t, h, "market-inventory-offer-seller+"+suffix+"@example.test")
	sellerChar := createCharacter(t, h, sellerToken, "Inventory Offer Seller")
	bidderID, bidderToken := loginEmail(t, h, "market-inventory-offer-bidder+"+suffix+"@example.test")
	bidderChar := createCharacter(t, h, bidderToken, "Inventory Offer Bidder")
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "inventory_offer_listing_item_" + suffix, AccountID: sellerID, CharacterID: sellerChar.CharacterID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, sellerID, sellerChar.CharacterID, "inventory_offer_listing_item_"+suffix, "inventory_offer_listing_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	var bidderItemIDs []string
	for _, itemID := range []string{"inventory_offer_bid_item_a_" + suffix, "inventory_offer_bid_item_b_" + suffix} {
		bidderItemIDs = append(bidderItemIDs, itemID)
		if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: bidderID, CharacterID: bidderChar.CharacterID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
			t.Fatal(err)
		}
	}
	rec := postJSON(h, "/v0/market/listings", sellerToken, map[string]string{"stash_item_id": "inventory_offer_listing_stash_" + suffix})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var listing marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", bidderToken, map[string]any{
		"item_instance_ids": bidderItemIDs,
		"character_id":      bidderChar.CharacterID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("inventory offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	var offer marketOfferResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &offer); err != nil {
		t.Fatal(err)
	}
	if offer.OfferID == "" || len(offer.Items) != 2 {
		t.Fatalf("inventory offer = %+v", offer)
	}
	items, err := db.ListCharacterItems(ctx, bidderID, bidderChar.CharacterID)
	if err != nil {
		t.Fatal(err)
	}
	for _, itemID := range bidderItemIDs {
		if characterItemsContain(items, itemID) {
			t.Fatalf("offered item %s still on bidder character after inventory offer = %+v", itemID, items)
		}
	}
}

func characterItemsContain(items []store.CharacterItemInstance, itemID string) bool {
	for _, item := range items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

func TestCharactersAreAccountScoped(t *testing.T) {
	h := fullServer(t)
	_, tokenA := loginEmail(t, h, testEmail(t, "characters-account-a"))
	_, tokenB := loginEmail(t, h, testEmail(t, "characters-account-b"))

	charA := createCharacter(t, h, tokenA, "Account A Hero")
	rec := getJSON(h, "/v0/characters", tokenB)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == charA.CharacterID {
			t.Fatalf("account B saw account A character: %+v", listed.Characters)
		}
	}
}

func TestStableDevAccountsKeepAccountBoundStateSeparate(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	accountA, tokenA := loginEmail(t, h, testEmail(t, "client-one"))
	accountB, tokenB := loginEmail(t, h, testEmail(t, "client-two"))
	if accountA == accountB {
		t.Fatal("expected stable client emails to resolve to distinct accounts")
	}
	beforeA := getMarketSummary(t, h, tokenA)
	beforeB := getMarketSummary(t, h, tokenB)

	charA := createCharacter(t, h, tokenA, "Client One Only")
	charB := createCharacter(t, h, tokenB, "Client Two Only")
	rec := getJSON(h, "/v0/characters", tokenB)
	if rec.Code != http.StatusOK {
		t.Fatalf("second account character list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listedB listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listedB); err != nil {
		t.Fatalf("decode second account characters: %v", err)
	}
	for _, c := range listedB.Characters {
		if c.CharacterID == charA.CharacterID || c.Name == charA.Name {
			t.Fatalf("second account saw first account character: %+v", listedB.Characters)
		}
	}

	itemID := "client1_bound_item_" + ids.Token()[:12]
	stashItemID := "client1_bound_stash_item_" + ids.Token()[:12]
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          itemID,
		AccountID:   accountA,
		CharacterID: charA.CharacterID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"damage_min":1}`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, accountA, charA.CharacterID, itemID, stashItemID); err != nil {
		t.Fatal(err)
	}
	stashB, err := db.ListAccountStashItems(ctx, accountB)
	if err != nil {
		t.Fatal(err)
	}
	if len(stashB) != 0 {
		t.Fatalf("second account stash leaked first account item: %+v", stashB)
	}

	rec = postJSON(h, "/v0/market/listings", tokenA, map[string]string{"stash_item_id": stashItemID})
	if rec.Code != http.StatusCreated {
		t.Fatalf("first account create listing status = %d, body = %s", rec.Code, rec.Body.String())
	}
	summaryA := getMarketSummary(t, h, tokenA)
	if summaryA.PublishedListings != beforeA.PublishedListings+1 {
		t.Fatalf("first account market summary = %+v, want one new published listing over %+v", summaryA, beforeA)
	}
	summaryB := getMarketSummary(t, h, tokenB)
	if summaryB.PublishedListings != beforeB.PublishedListings || summaryB.IncomingBids != beforeB.IncomingBids {
		t.Fatalf("second account market summary leaked first account state: before=%+v after=%+v", beforeB, summaryB)
	}
	if charB.CharacterID == "" {
		t.Fatal("second account character was not created")
	}
}

func getMarketSummary(t *testing.T, h http.Handler, token string) marketSummaryResponse {
	t.Helper()
	rec := getJSON(h, "/v0/market/summary", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("market summary status = %d body=%s", rec.Code, rec.Body.String())
	}
	var summary marketSummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	return summary
}

func TestDeleteCharacterRemovesOwnedCharacter(t *testing.T) {
	h := fullServer(t)
	_, token := loginEmail(t, h, "characters-delete+"+ids.Token()[:12]+"@example.test")
	keep := createCharacter(t, h, token, "Keep")
	remove := createCharacter(t, h, token, "Remove")

	rec := deleteJSON(h, "/v0/characters/"+remove.CharacterID, token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = getJSON(h, "/v0/characters", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Characters) != 2 {
		t.Fatalf("listed characters after delete = %+v, want 2 remaining", listed.Characters)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == remove.CharacterID {
			t.Fatalf("deleted character still listed: %+v", listed.Characters)
		}
	}
	var foundKeep bool
	for _, c := range listed.Characters {
		if c.CharacterID == keep.CharacterID {
			foundKeep = true
			break
		}
	}
	if !foundKeep {
		t.Fatalf("keep character missing after delete: %+v", listed.Characters)
	}

	rec = deleteJSON(h, "/v0/characters/"+remove.CharacterID, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", rec.Code)
	}
}

func TestDeleteCharacterIsAccountScoped(t *testing.T) {
	h := fullServer(t)
	suffix := ids.Token()[:12]
	_, tokenA := loginEmail(t, h, "characters-delete-a+"+suffix+"@example.test")
	_, tokenB := loginEmail(t, h, "characters-delete-b+"+suffix+"@example.test")
	charA := createCharacter(t, h, tokenA, "Account A Hero")

	rec := deleteJSON(h, "/v0/characters/"+charA.CharacterID, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete other account character status = %d, want 404", rec.Code)
	}

	rec = getJSON(h, "/v0/characters", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var listed listCharactersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Characters) != 2 {
		t.Fatalf("account A characters deleted by account B: %+v", listed.Characters)
	}
	for _, c := range listed.Characters {
		if c.CharacterID == charA.CharacterID {
			return
		}
	}
	t.Fatalf("account A character missing after cross-account delete attempt: %+v", listed.Characters)
}

func TestCreateAndResumeSession(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.SessionID == "" || created.Seed == "" || created.CharacterID == "" || created.WorldID != game.DefaultWorldID {
		t.Fatalf("incomplete session response: %+v", created)
	}
	if created.WSURL != "/v0/ws?session_id="+created.SessionID {
		t.Fatalf("ws_url = %q", created.WSURL)
	}

	// Resume the same session.
	resumeID := created.SessionID
	rec = postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "resume_session_id": resumeID})
	if rec.Code != http.StatusOK {
		t.Fatalf("resume status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resumed createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resumed)
	if resumed.SessionID != resumeID || resumed.Seed != created.Seed || resumed.WorldID != created.WorldID {
		t.Fatalf("resume mismatch: %+v vs %+v", resumed, created)
	}
}

func TestCreateCoopSessionAndJoinLifecycle(t *testing.T) {
	h := fullServer(t)
	_, hostToken := loginEmail(t, h, "coop-host+"+ids.Token()[:12]+"@example.test")
	_, guestToken := loginEmail(t, h, "coop-guest+"+ids.Token()[:12]+"@example.test")
	_, thirdToken := loginEmail(t, h, "coop-third+"+ids.Token()[:12]+"@example.test")
	_, fourthToken := loginEmail(t, h, "coop-fourth+"+ids.Token()[:12]+"@example.test")
	guestChar := createCharacter(t, h, guestToken, "Guest Hero")
	thirdChar := createCharacter(t, h, thirdToken, "Third Hero")
	fourthChar := createCharacter(t, h, fourthToken, "Fourth Hero")

	rec := postJSON(h, "/v0/sessions", hostToken, map[string]any{"mode": "coop", "world_id": "dungeon_levels"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create coop status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode coop create: %v", err)
	}
	if created.Mode != store.SessionModeCoop || created.JoinCode == "" || created.WorldID != "dungeon_levels" {
		t.Fatalf("coop create response mismatch: %+v", created)
	}
	if created.Listed {
		t.Fatalf("private coop unexpectedly listed: %+v", created)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", guestToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("join status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var joined createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &joined); err != nil {
		t.Fatalf("decode join: %v", err)
	}
	if joined.SessionID != created.SessionID || joined.CharacterID != guestChar.CharacterID || joined.Mode != store.SessionModeCoop || joined.JoinCode != "" {
		t.Fatalf("join response mismatch: %+v", joined)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", guestToken, map[string]any{
		"join_code": created.JoinCode, "character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate join status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "duplicate_member" {
		t.Fatalf("duplicate code = %q", body.Error.Code)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", thirdToken, map[string]any{
		"join_code": created.JoinCode, "character_id": thirdChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("third join status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", fourthToken, map[string]any{
		"join_code": created.JoinCode, "character_id": fourthChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("fourth join status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", thirdToken, map[string]any{
		"join_code": "join_wrong", "character_id": thirdChar.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("wrong code status = %d, want 404", rec.Code)
	}

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/join", thirdToken, map[string]any{
		"join_code": created.JoinCode, "character_id": created.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign character status = %d, want 404", rec.Code)
	}

	rec = postJSON(h, "/v0/sessions", hostToken, map[string]any{"mode": "coop"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create second coop status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var ended createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &ended)
	rec = postJSON(h, "/v0/sessions/"+ended.SessionID+"/end", hostToken, map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("end coop status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions/"+ended.SessionID+"/join", guestToken, map[string]any{
		"join_code": ended.JoinCode, "character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("join ended status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "session_ended" {
		t.Fatalf("ended join code = %q", body.Error.Code)
	}
}

func TestListedSessionListAndJoinWithoutCode(t *testing.T) {
	h := fullServer(t)
	_, hostToken := loginEmail(t, h, "listed-host+"+ids.Token()[:12]+"@example.test")
	_, guestToken := loginEmail(t, h, "listed-guest+"+ids.Token()[:12]+"@example.test")
	_, outsiderToken := loginEmail(t, h, "listed-outsider+"+ids.Token()[:12]+"@example.test")
	hostChar := createCharacter(t, h, hostToken, "Listed Host")
	guestChar := createCharacter(t, h, guestToken, "Listed Guest")
	outsiderChar := createCharacter(t, h, outsiderToken, "Listed Outsider")

	privateRec := postJSON(h, "/v0/sessions", hostToken, map[string]any{
		"mode": "coop", "world_id": "dungeon_levels", "character_id": hostChar.CharacterID,
	})
	if privateRec.Code != http.StatusCreated {
		t.Fatalf("private create status = %d, body = %s", privateRec.Code, privateRec.Body.String())
	}
	soloRec := postJSON(h, "/v0/sessions", hostToken, map[string]any{
		"mode": "solo", "world_id": "dungeon_levels", "character_id": hostChar.CharacterID,
	})
	if soloRec.Code != http.StatusCreated {
		t.Fatalf("solo create status = %d, body = %s", soloRec.Code, soloRec.Body.String())
	}
	listedRec := postJSON(h, "/v0/sessions", hostToken, map[string]any{
		"mode": "coop", "listed": true, "world_id": "dungeon_levels", "character_id": hostChar.CharacterID,
	})
	if listedRec.Code != http.StatusCreated {
		t.Fatalf("listed create status = %d, body = %s", listedRec.Code, listedRec.Body.String())
	}
	var listed createSessionResponse
	if err := json.Unmarshal(listedRec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode listed create: %v", err)
	}
	if !listed.Listed || listed.JoinCode == "" {
		t.Fatalf("listed create response mismatch: %+v", listed)
	}

	rec := getJSON(h, "/v0/sessions/active", guestToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("active list status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var active activeSessionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &active); err != nil {
		t.Fatalf("decode active sessions: %v", err)
	}
	var found activeSessionSummaryResponse
	for _, sess := range active.Sessions {
		if sess.SessionID == listed.SessionID {
			found = sess
		}
		if strings.Contains(rec.Body.String(), "join_") {
			t.Fatalf("active list exposed join code: %s", rec.Body.String())
		}
	}
	if found.SessionID != "" {
		t.Fatalf("unconnected listed session should not be active-browser visible: %+v", found)
	}

	rec = postJSON(h, "/v0/sessions/"+listed.SessionID+"/join", guestToken, map[string]any{
		"character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("listed join status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var joined createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &joined); err != nil {
		t.Fatalf("decode listed join: %v", err)
	}
	if joined.SessionID != listed.SessionID || joined.CharacterID != guestChar.CharacterID || !joined.Listed || joined.JoinCode != "" {
		t.Fatalf("listed join response mismatch: %+v", joined)
	}

	rec = postJSON(h, "/v0/sessions/"+listed.SessionID+"/join", outsiderToken, map[string]any{
		"character_id": guestChar.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-account listed join status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(h, "/v0/sessions/"+listed.SessionID+"/join", outsiderToken, map[string]any{
		"character_id": outsiderChar.CharacterID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("second listed join status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSessionCustomSeedLocalOnly(t *testing.T) {
	local := fullServer(t)
	_, localToken := login(t, local)
	rec := postJSON(local, "/v0/sessions", localToken, map[string]any{"mode": "solo", "seed": "pinned-test-seed"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("local custom seed status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.Seed != "pinned-test-seed" {
		t.Fatalf("local seed = %q, want pinned-test-seed", created.Seed)
	}

	remote := fullServerWithConfig(t, config.Config{Addr: ":0", Env: "remote", DevToken: testDevToken, MetricsEnabled: true})
	_, remoteToken := loginEmail(t, remote, testEmail(t, "remote-seed"))
	rec = postJSON(remote, "/v0/sessions", remoteToken, map[string]any{"mode": "solo", "seed": "pinned-test-seed"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("remote custom seed status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSessionWorldID(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "world_id": "gear_before_combat"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.WorldID != "gear_before_combat" {
		t.Fatalf("world_id = %q, want gear_before_combat", created.WorldID)
	}

	resumeID := created.SessionID
	rec = postJSON(h, "/v0/sessions", token, map[string]any{
		"mode": "solo", "resume_session_id": resumeID, "world_id": game.DefaultWorldID,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("resume status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resumed createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resumed)
	if resumed.WorldID != "gear_before_combat" {
		t.Fatalf("resume world_id = %q, want persisted gear_before_combat", resumed.WorldID)
	}
}

func TestCreateSessionWithSelectedCharacter(t *testing.T) {
	h := fullServer(t)
	_, tokenA := loginEmail(t, h, testEmail(t, "selected-character-a"))
	_, tokenB := loginEmail(t, h, testEmail(t, "selected-character-b"))
	charA := createCharacter(t, h, tokenA, "Selected Hero")

	rec := postJSON(h, "/v0/sessions", tokenA, map[string]any{
		"mode": "solo", "world_id": "dungeon_levels", "character_id": charA.CharacterID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create selected status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if created.CharacterID != charA.CharacterID || created.WorldID != "dungeon_levels" {
		t.Fatalf("selected session mismatch: %+v, character=%+v", created, charA)
	}

	rec = postJSON(h, "/v0/sessions", tokenB, map[string]any{
		"mode": "solo", "world_id": "dungeon_levels", "character_id": charA.CharacterID,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-account selected status = %d, want 404, body = %s", rec.Code, rec.Body.String())
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "character_not_found" {
		t.Fatalf("error code = %q, want character_not_found", body.Error.Code)
	}
}

func TestCreateSessionOmittedCharacterUsesDefault(t *testing.T) {
	h := fullServer(t)
	_, token := loginEmail(t, h, testEmail(t, "default-character-session"))

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "world_id": "dungeon_levels"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if created.CharacterID == "" || created.WorldID != "dungeon_levels" {
		t.Fatalf("default-character session incomplete: %+v", created)
	}
}

func TestEndSessionOwnerOnlyAndIdempotent(t *testing.T) {
	h := fullServer(t)
	_, tokenA := loginEmail(t, h, testEmail(t, "end-session-a"))
	_, tokenB := loginEmail(t, h, testEmail(t, "end-session-b"))

	rec := postJSON(h, "/v0/sessions", tokenA, map[string]any{"mode": "solo"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created createSessionResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/end", tokenB, map[string]any{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-account end status = %d, want 404", rec.Code)
	}
	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/end", tokenA, map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("end status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/sessions/"+created.SessionID+"/end", tokenA, map[string]any{})
	if rec.Code != http.StatusOK {
		t.Fatalf("second end status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSessionRejectsUnknownWorldID(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)

	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "world_id": "missing"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
	var body errorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "invalid_world_id" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestResumeUnknownSession(t *testing.T) {
	h := fullServer(t)
	_, token := login(t, h)
	missing := "sess_00000000000000000000000000"
	rec := postJSON(h, "/v0/sessions", token, map[string]any{"mode": "solo", "resume_session_id": missing})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
