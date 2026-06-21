package ical

import (
	"os"
	"strings"
	"testing"
)

const indexNowKey = "9243cdd67ce6db495ec7acddec1dec27"

// TestIndexNowKeyFileContents asserts that the key file contents equal its filename stem.
func TestIndexNowKeyFileContents(t *testing.T) {
	path := "website/static/" + indexNowKey + ".txt"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("key file not found at %s: %v", path, err)
	}
	got := strings.TrimSpace(string(data))
	if got != indexNowKey {
		t.Errorf("key file contents %q != filename stem %q", got, indexNowKey)
	}
}

// TestIndexNowScriptReferencesKey asserts that scripts/indexnow.sh references the same key.
func TestIndexNowScriptReferencesKey(t *testing.T) {
	data, err := os.ReadFile("scripts/indexnow.sh")
	if err != nil {
		t.Fatalf("script not found: %v", err)
	}
	if !strings.Contains(string(data), indexNowKey) {
		t.Errorf("scripts/indexnow.sh does not reference key %q", indexNowKey)
	}
}
