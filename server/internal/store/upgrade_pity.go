package store

import (
	"encoding/json"
	"fmt"
)

func upgradePityFailures(raw json.RawMessage) (int, error) {
	payload := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return 0, fmt.Errorf("store: decode rolled stats pity: %w", err)
		}
	}
	rec, ok := payload["upgrade_pity"].(map[string]any)
	if !ok {
		return 0, nil
	}
	failures := numericStatValue(rec["failures"])
	if failures < 0 {
		return 0, nil
	}
	return failures, nil
}

func rolledStatsWithUpgradePityFailures(raw []byte, failures int) ([]byte, error) {
	payload := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("store: decode rolled stats pity: %w", err)
		}
	}
	if failures < 0 {
		failures = 0
	}
	payload["upgrade_pity"] = map[string]any{"failures": failures}
	out, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("store: encode rolled stats pity: %w", err)
	}
	return out, nil
}
