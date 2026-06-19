package httpapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
)

func TestDebugCharacterProgressionSeedsDeepestDungeonDepth(t *testing.T) {
	h, db := fullServerWithConfigAndStore(t, config.Config{Addr: ":0", Env: "local", DevToken: testDevToken, DebugToken: testDebugToken})
	ctx := context.Background()
	accountID, token := loginEmail(t, h, "debug-depth+"+ids.Token()[:12]+"@example.test")
	hero := createCharacter(t, h, token, "Depth Debugger")

	rec := putDebugJSON(h, "/v0/debug/characters/"+hero.CharacterID+"/progression", token, map[string]any{
		"level":                 1,
		"experience":            0,
		"unspent_stat_points":   0,
		"unspent_skill_points":  0,
		"deepest_dungeon_depth": 42,
	}, testDebugToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("debug seed status = %d, body = %s", rec.Code, rec.Body.String())
	}

	progression, err := db.GetCharacterProgression(ctx, accountID, hero.CharacterID)
	if err != nil {
		t.Fatalf("load seeded progression: %v", err)
	}
	if progression.DeepestDungeonDepth != 42 {
		t.Fatalf("seeded deepest depth = %d, want 42", progression.DeepestDungeonDepth)
	}
}
