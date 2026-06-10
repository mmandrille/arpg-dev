package game

import (
	"fmt"
	"os"
	"path/filepath"
)

const contentLibrariesFile = "content_libraries.v0.json"

type contentLibraryManifest struct {
	Version int                   `json:"version"`
	Rules   contentManifestRules  `json:"rules"`
	Assets  contentManifestAssets `json:"assets"`
}

type contentManifestRules struct {
	Skills []contentManifestEntry `json:"skills"`
}

type contentManifestAssets struct {
	Skills contentManifestSkillAssets `json:"skills"`
}

type contentManifestSkillAssets struct {
	Presentations []contentManifestEntry `json:"presentations"`
}

type contentManifestEntry struct {
	Group string `json:"group"`
	Path  string `json:"path"`
}

func loadSkillRulesFromManifest(rulesDir string) (map[string]SkillDef, error) {
	manifestPath := contentManifestPathForRulesDir(rulesDir)
	var manifest contentLibraryManifest
	if err := readJSON(manifestPath, &manifest); err != nil {
		return nil, err
	}
	if manifest.Version != 0 {
		return nil, fmt.Errorf("game: invalid content manifest version: %d", manifest.Version)
	}
	if len(manifest.Rules.Skills) == 0 {
		return nil, fmt.Errorf("game: invalid content manifest rules.skills: at least one file is required")
	}

	merged := make(map[string]SkillDef)
	for idx, entry := range manifest.Rules.Skills {
		path, err := resolveContentManifestEntry(manifestPath, fmt.Sprintf("rules.skills[%d]", idx), entry)
		if err != nil {
			return nil, err
		}
		var file struct {
			Skills map[string]SkillDef `json:"skills"`
		}
		if err := readJSON(path, &file); err != nil {
			return nil, err
		}
		if len(file.Skills) == 0 {
			return nil, fmt.Errorf("game: invalid content manifest rules.skills[%d].path %s: missing skills object", idx, entry.Path)
		}
		for _, id := range sortedStringKeys(file.Skills) {
			if _, exists := merged[id]; exists {
				return nil, fmt.Errorf("game: invalid content manifest rules.skills[%d].path %s: duplicate skill id %s", idx, entry.Path, id)
			}
			merged[id] = file.Skills[id]
		}
	}
	return merged, nil
}

func contentManifestPathForRulesDir(rulesDir string) string {
	return filepath.Join(filepath.Dir(rulesDir), "content", contentLibrariesFile)
}

func resolveContentManifestEntry(manifestPath string, label string, entry contentManifestEntry) (string, error) {
	if entry.Group == "" {
		return "", fmt.Errorf("game: invalid content manifest %s.group: required", label)
	}
	if entry.Path == "" {
		return "", fmt.Errorf("game: invalid content manifest %s.path: required", label)
	}
	if filepath.IsAbs(entry.Path) {
		return "", fmt.Errorf("game: invalid content manifest %s.path %s: must be relative", label, entry.Path)
	}
	path := filepath.Clean(filepath.Join(filepath.Dir(manifestPath), entry.Path))
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("game: invalid content manifest %s.path %s: %w", label, entry.Path, err)
	}
	return path, nil
}
