package game

import (
	"regexp"
	"testing"
)

// protocolHexColor is the #RRGGBB shape both v8 protocol schemas require for entity visual_tint.
var protocolHexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func TestCompanionVisualTintsUseProtocolHexForm(t *testing.T) {
	rules := loadRules(t)
	checked := 0
	for _, skillID := range sortedStringKeys(rules.Skills) {
		tint := rules.Skills[skillID].Companion.VisualTint
		if tint == "" {
			continue
		}
		checked++
		if got := protocolVisualTint(tint); !protocolHexColor.MatchString(got) {
			t.Fatalf("skill %s companion tint %q -> %q, want #RRGGBB", skillID, tint, got)
		}
	}
	if checked == 0 {
		t.Fatal("no companion skill declares a visual_tint; the rules-form conversion is untested")
	}
	if !protocolHexColor.MatchString(revivedCompanionCorpseVisualTint) {
		t.Fatalf("revived corpse tint %q, want #RRGGBB", revivedCompanionCorpseVisualTint)
	}
}

func TestProtocolVisualTintKeepsEmptyAndPrefixedValues(t *testing.T) {
	for in, want := range map[string]string{"": "", "#b77cff": "#b77cff", "101014": "#101014"} {
		if got := protocolVisualTint(in); got != want {
			t.Fatalf("protocolVisualTint(%q) = %q, want %q", in, got, want)
		}
	}
}
