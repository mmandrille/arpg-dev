package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUniqueItemValidationRejectsInvalidRules(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(t *testing.T, rulesDir string)
		wantErr string
	}{
		{
			name: "unknown base template",
			mutate: func(t *testing.T, rulesDir string) {
				mutateUniqueItems(t, rulesDir, func(unique map[string]any) {
					unique["base_template_id"] = "missing_template"
				})
			},
			wantErr: "unknown template missing_template",
		},
		{
			name: "enabled status mismatch",
			mutate: func(t *testing.T, rulesDir string) {
				mutateUniqueItems(t, rulesDir, func(unique map[string]any) {
					unique["status"] = "disabled_seed"
				})
			},
			wantErr: "enabled entries must be ready",
		},
		{
			name: "duplicate fixed effect",
			mutate: func(t *testing.T, rulesDir string) {
				mutateUniqueItems(t, rulesDir, func(unique map[string]any) {
					unique["fixed_effect_ids"] = []any{"everburning_wound", "everburning_wound"}
				})
			},
			wantErr: "duplicate effect",
		},
		{
			name: "unknown fixed effect",
			mutate: func(t *testing.T, rulesDir string) {
				mutateUniqueItems(t, rulesDir, func(unique map[string]any) {
					unique["fixed_effect_ids"] = []any{"missing_effect"}
				})
			},
			wantErr: "unknown or inactive effect",
		},
		{
			name: "inactive fixed effect",
			mutate: func(t *testing.T, rulesDir string) {
				mutateUniqueEffects(t, rulesDir, "everburning_wound", func(effect map[string]any) {
					effect["enabled"] = false
					effect["status"] = "disabled_seed"
				})
			},
			wantErr: "must be enabled and ready",
		},
		{
			name: "incompatible fixed effect",
			mutate: func(t *testing.T, rulesDir string) {
				mutateUniqueItems(t, rulesDir, func(unique map[string]any) {
					unique["fixed_effect_ids"] = []any{"last_stand_glimmer"}
				})
			},
			wantErr: "incompatible with template type sword",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rulesDir := tempRulesDir(t)
			tc.mutate(t, rulesDir)

			_, err := LoadRules(rulesDir)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("LoadRules error = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func tempRulesDir(t *testing.T) string {
	t.Helper()
	sourceDir, err := FindSharedRulesDir()
	if err != nil {
		t.Fatalf("locate rules: %v", err)
	}
	targetRulesDir := filepath.Join(t.TempDir(), "rules")
	if err := copyTree(sourceDir, targetRulesDir); err != nil {
		t.Fatalf("copy rules: %v", err)
	}
	return targetRulesDir
}

func mutateUniqueItems(t *testing.T, rulesDir string, mutate func(unique map[string]any)) {
	t.Helper()
	path := filepath.Join(rulesDir, "unique_items.v0.json")
	doc := readJSONMap(t, path)
	uniques := doc["uniques"].(map[string]any)
	unique := uniques["embercall_blade"].(map[string]any)
	mutate(unique)
	writeJSONMap(t, path, doc)
}

func mutateUniqueEffects(t *testing.T, rulesDir string, effectID string, mutate func(effect map[string]any)) {
	t.Helper()
	path := filepath.Join(rulesDir, "unique_effects.v0.json")
	doc := readJSONMap(t, path)
	effects := doc["effects"].(map[string]any)
	effect := effects[effectID].(map[string]any)
	mutate(effect)
	writeJSONMap(t, path, doc)
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return doc
}

func writeJSONMap(t *testing.T, path string, doc map[string]any) {
	t.Helper()
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
