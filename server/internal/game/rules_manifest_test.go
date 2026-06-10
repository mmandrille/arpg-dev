package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSkillRulesFromManifestUsesRepoManifest(t *testing.T) {
	rules := loadRules(t)
	skill, ok := rules.Skills["magic_bolt"]
	if !ok {
		t.Fatalf("magic_bolt missing from manifest-loaded skills")
	}
	if skill.Name != "Magic Bolt" || skill.Projectile.Visual != "magic_bolt_projectile" {
		t.Fatalf("magic_bolt loaded incorrectly: %+v", skill)
	}
}

func TestLoadSkillRulesFromManifestRejectsDuplicateIDs(t *testing.T) {
	rulesDir := writeSkillManifestFixture(t, map[string]string{
		"a.v0.json": `{"version":0,"skills":{"magic_bolt":{"name":"A"}}}`,
		"b.v0.json": `{"version":0,"skills":{"magic_bolt":{"name":"B"}}}`,
	}, []contentManifestEntry{
		{Group: "a", Path: "../rules/a.v0.json"},
		{Group: "b", Path: "../rules/b.v0.json"},
	})

	_, err := loadSkillRulesFromManifest(rulesDir)
	if err == nil || !strings.Contains(err.Error(), "duplicate skill id magic_bolt") {
		t.Fatalf("loadSkillRulesFromManifest error = %v, want duplicate skill id", err)
	}
}

func TestLoadSkillRulesFromManifestRejectsMissingListedFile(t *testing.T) {
	rulesDir := writeSkillManifestFixture(t, nil, []contentManifestEntry{
		{Group: "missing", Path: "../rules/missing.v0.json"},
	})

	_, err := loadSkillRulesFromManifest(rulesDir)
	if err == nil || !strings.Contains(err.Error(), "missing.v0.json") {
		t.Fatalf("loadSkillRulesFromManifest error = %v, want missing listed file", err)
	}
}

func writeSkillManifestFixture(t *testing.T, files map[string]string, entries []contentManifestEntry) string {
	t.Helper()
	sharedDir := t.TempDir()
	rulesDir := filepath.Join(sharedDir, "rules")
	contentDir := filepath.Join(sharedDir, "content")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("mkdir rules: %v", err)
	}
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		t.Fatalf("mkdir content: %v", err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(rulesDir, name), []byte(body+"\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	var builder strings.Builder
	builder.WriteString(`{"version":0,"rules":{"skills":[`)
	for idx, entry := range entries {
		if idx > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(`{"group":"`)
		builder.WriteString(entry.Group)
		builder.WriteString(`","path":"`)
		builder.WriteString(entry.Path)
		builder.WriteString(`"}`)
	}
	builder.WriteString(`]},"assets":{"skills":{"presentations":[{"group":"default","path":"../rules/a.v0.json"}]}}}`)
	if err := os.WriteFile(filepath.Join(contentDir, contentLibrariesFile), []byte(builder.String()+"\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return rulesDir
}
