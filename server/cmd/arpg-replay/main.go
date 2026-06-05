// Command arpg-replay re-simulates a recorded session from its seed + input
// stream and verifies it reproduces the recorded authoritative events. Exit
// code 0 = match, 1 = mismatch, 2 = usage/runtime error (ADR-0001 D8.2).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/replay"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func main() {
	sessionID := flag.String("session-id", "", "session id to verify")
	jsonOut := flag.Bool("json", false, "print the full report as JSON")
	flag.Parse()

	if *sessionID == "" {
		fmt.Fprintln(os.Stderr, "usage: arpg-replay --session-id <id>")
		os.Exit(2)
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "rules: %v\n", err)
		os.Exit(2)
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rules: %v\n", err)
		os.Exit(2)
	}

	report, err := replay.Verify(ctx, db, rules, *sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(2)
	}

	if *jsonOut {
		b, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("session=%s seed=%s inputs=%d recorded_events=%d derived_events=%d\n",
			report.SessionID, report.Seed, report.InputCount, report.RecordedEventCount, report.DerivedEventCount)
		if report.Match {
			fmt.Println("REPLAY OK: re-simulation reproduced recorded authoritative events")
		} else {
			fmt.Printf("REPLAY MISMATCH: %s\n", report.Mismatch)
		}
	}

	if !report.Match {
		os.Exit(1)
	}
}
