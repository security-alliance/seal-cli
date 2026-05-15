package index

import (
	"testing"
)

func TestLoadAcceptsV1(t *testing.T) {
	data := []byte(`{"schema_version":1,"branch":"main","commit_sha":"abc","generated_at":"2024-01-01","section_count":0,"sections":[],"sections_by_id":{},"frameworks":[]}`)
	idx, err := Load(data)
	if err != nil {
		t.Fatalf("expected v1 to load: %v", err)
	}
	if idx.SchemaVersion != 1 {
		t.Fatalf("expected schema 1, got %d", idx.SchemaVersion)
	}
}

func TestLoadAcceptsZeroForTransition(t *testing.T) {
	data := []byte(`{"branch":"main","commit_sha":"abc","generated_at":"2024-01-01","section_count":0,"sections":[],"sections_by_id":{},"frameworks":[]}`)
	_, err := Load(data)
	if err != nil {
		t.Fatalf("expected zero to load during transition: %v", err)
	}
}

func TestLoadRejectsUnknownVersion(t *testing.T) {
	data := []byte(`{"schema_version":99,"branch":"main","commit_sha":"abc","generated_at":"2024-01-01","section_count":0,"sections":[],"sections_by_id":{},"frameworks":[]}`)
	_, err := Load(data)
	if err == nil {
		t.Fatal("expected unknown version to be rejected")
	}
}

func TestIsDraft(t *testing.T) {
	if !(&Index{Branch: "develop"}).IsDraft() {
		t.Fatal("develop should be draft")
	}
	if (&Index{Branch: "main"}).IsDraft() {
		t.Fatal("main should not be draft")
	}
}
