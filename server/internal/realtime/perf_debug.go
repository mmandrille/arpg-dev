package realtime

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

const defaultPerfDebugInterval = time.Second

func perfDebugEnabled() bool {
	enabled, _ := strconv.ParseBool(os.Getenv("ARPG_PERF_DEBUG"))
	return enabled
}

func logBackendPerf(log *slog.Logger, tick uint64, started time.Time, simDuration time.Duration, inputs int, results []game.TickResult, clients int, snapshot game.PerfSnapshot) {
	changes := 0
	events := 0
	acks := 0
	rejects := 0
	for _, res := range results {
		changes += len(res.Changes)
		events += len(res.Events)
		acks += len(res.Acks)
		rejects += len(res.Rejects)
	}
	log.Info("backend_perf",
		"tick", tick,
		"total_ms", float64(time.Since(started).Microseconds())/1000.0,
		"sim_ms", float64(simDuration.Microseconds())/1000.0,
		"inputs", inputs,
		"results", len(results),
		"changes", changes,
		"events", events,
		"acks", acks,
		"rejects", rejects,
		"clients", clients,
		"game_level", snapshot.Level,
		"entities", snapshot.Entities,
		"players", snapshot.Players,
		"monsters", snapshot.Monsters,
		"live_monsters", snapshot.LiveMonsters,
		"companions", snapshot.Companions,
		"projectiles", snapshot.Projectiles,
		"loot", snapshot.Loot,
		"interactables", snapshot.Interactables,
		"walls", snapshot.Walls,
	)
}
