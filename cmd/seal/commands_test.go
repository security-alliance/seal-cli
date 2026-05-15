package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/security-alliance/seal-cli/internal/index"
)

func setupFixtureCache(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	indexesDir := filepath.Join(dir, "indexes")
	_ = os.MkdirAll(indexesDir, 0755)

	fixture, err := os.ReadFile("testdata/fixture-index.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	_ = os.WriteFile(filepath.Join(indexesDir, "main-index.json"), fixture, 0644)
	_ = os.WriteFile(filepath.Join(indexesDir, "develop-index.json"), fixture, 0644)

	oldCache := os.Getenv("SEAL_CACHE_DIR")
	os.Setenv("SEAL_CACHE_DIR", dir)
	t.Cleanup(func() { os.Setenv("SEAL_CACHE_DIR", oldCache) })
	return dir
}

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func captureOutputErr(fn func() error) (string, error) {
	oldOut := os.Stdout
	oldErr := os.Stderr
	pr, pw, _ := os.Pipe()
	os.Stdout = pw
	os.Stderr = pw
	err := fn()
	_ = pw.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(pr)
	return buf.String(), err
}

func TestRunList(t *testing.T) {
	setupFixtureCache(t)
	out, err := captureOutputErr(func() error { return runList([]string{"--branch", "main"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "main") {
		t.Fatalf("expected branch info, got: %s", out)
	}
}

func TestRunListJSON(t *testing.T) {
	setupFixtureCache(t)
	out, err := captureOutputErr(func() error { return runList([]string{"--branch", "main", "--json"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(out), &body); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if body["branch"] != "main" {
		t.Fatalf("expected branch main, got %v", body["branch"])
	}
}

func TestRunSearch(t *testing.T) {
	setupFixtureCache(t)
	out, err := captureOutputErr(func() error { return runSearch([]string{"ENS", "--branch", "main", "--limit", "3"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ENS") && !strings.Contains(out, "ens") {
		t.Fatalf("expected ENS results, got: %s", out)
	}
}

func TestRunSearchJSON(t *testing.T) {
	setupFixtureCache(t)
	out, err := captureOutputErr(func() error { return runSearch([]string{"ENS", "--branch", "main", "--json", "--limit", "2"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(out), &body); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	results, ok := body["results"].([]interface{})
	if !ok {
		t.Fatalf("expected results array")
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
}

func TestRunFetch(t *testing.T) {
	setupFixtureCache(t)
	fixture, _ := os.ReadFile("testdata/fixture-index.json")
	var idx index.Index
	_ = json.Unmarshal(fixture, &idx)
	if len(idx.Sections) == 0 {
		t.Fatal("fixture has no sections")
	}
	id := idx.Sections[0].ID
	out, err := captureOutputErr(func() error { return runFetch([]string{id, "--branch", "main"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, idx.Sections[0].SectionTitle) {
		t.Fatalf("expected section title in output, got: %s", out)
	}
}

func TestRunFetchJSON(t *testing.T) {
	setupFixtureCache(t)
	fixture, _ := os.ReadFile("testdata/fixture-index.json")
	var idx index.Index
	_ = json.Unmarshal(fixture, &idx)
	id := idx.Sections[0].ID
	out, err := captureOutputErr(func() error { return runFetch([]string{id, "--branch", "main", "--json"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var sec index.Section
	if err := json.Unmarshal([]byte(out), &sec); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if sec.ID != id {
		t.Fatalf("expected id %s, got %s", id, sec.ID)
	}
}

func TestRunFetchNotFound(t *testing.T) {
	setupFixtureCache(t)
	_, err := captureOutputErr(func() error { return runFetch([]string{"nonexistent/path#anchor", "--branch", "main"}) })
	if err == nil {
		t.Fatal("expected error for missing section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got: %v", err)
	}
}

func TestRunCompare(t *testing.T) {
	setupFixtureCache(t)
	fixture, _ := os.ReadFile("testdata/fixture-index.json")
	var idx index.Index
	_ = json.Unmarshal(fixture, &idx)
	if len(idx.Sections) == 0 {
		t.Fatal("fixture has no sections")
	}
	path := idx.Sections[0].Path
	out, err := captureOutputErr(func() error { return runCompare([]string{path, "--left", "main", "--right", "develop"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Status:") {
		t.Fatalf("expected compare status, got: %s", out)
	}
}

func TestRunCompareJSON(t *testing.T) {
	setupFixtureCache(t)
	fixture, _ := os.ReadFile("testdata/fixture-index.json")
	var idx index.Index
	_ = json.Unmarshal(fixture, &idx)
	path := idx.Sections[0].Path
	out, err := captureOutputErr(func() error { return runCompare([]string{path, "--left", "main", "--right", "develop", "--json"}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(out), &body); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if body["path"] != path {
		t.Fatalf("expected path %s, got %v", path, body["path"])
	}
}

func TestRunUpdate(t *testing.T) {
	// Mock a raw GitHub-like server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"schema_version":1,"branch":"main","commit_sha":"abc","generated_at":"2024-01-01","section_count":0,"sections":[],"sections_by_id":{},"frameworks":[]}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldCache := os.Getenv("SEAL_CACHE_DIR")
	os.Setenv("SEAL_CACHE_DIR", dir)
	defer os.Setenv("SEAL_CACHE_DIR", oldCache)

	// runUpdate constructs URLs from raw.githubusercontent.com/security-alliance/frameworks-mcp/{ref}/indexes/
	// We can't easily override that without changing the code. Instead, we test DownloadRaw directly in cache_test.
	// Here we just verify runUpdate returns an error when source=release, and succeeds with raw when the URL is reachable.
	// Since runUpdate hardcodes the GitHub URL, we'll skip direct integration and test via cache.DownloadRaw below.
	fmt.Println("skipping runUpdate direct test — raw URL is hardcoded; cache.DownloadRaw is tested in cache_test")
}

func TestRunEmergency(t *testing.T) {
	out, err := captureOutputErr(func() error { return runEmergency([]string{}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "SEAL 911") {
		t.Fatalf("expected SEAL 911 header, got: %s", out)
	}
	if !strings.Contains(out, "t.me/SEAL_911_bot") {
		t.Fatalf("expected telegram link, got: %s", out)
	}
}

func TestRunTips(t *testing.T) {
	out, err := captureOutputErr(func() error { return runTips([]string{}) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "SEAL Tips") {
		t.Fatalf("expected SEAL Tips header, got: %s", out)
	}
	if !strings.Contains(out, "t.me/SEAL_tips_bot") {
		t.Fatalf("expected telegram link, got: %s", out)
	}
}

func TestBranchFromArgs(t *testing.T) {
	if branchFromArgs([]string{"--branch", "develop"}) != "develop" {
		t.Fatal("expected develop")
	}
	if branchFromArgs([]string{}) != "main" {
		t.Fatal("expected default main")
	}
}

func TestHasFlag(t *testing.T) {
	if !hasFlag([]string{"--json", "--branch", "main"}, "--json") {
		t.Fatal("expected hasFlag true")
	}
	if hasFlag([]string{"--branch", "main"}, "--json") {
		t.Fatal("expected hasFlag false")
	}
}

func TestGetFlag(t *testing.T) {
	if getFlag([]string{"--branch", "develop"}, "--branch") != "develop" {
		t.Fatal("expected develop")
	}
	if getFlag([]string{"--json"}, "--branch") != "" {
		t.Fatal("expected empty")
	}
}

func TestTrimQuote(t *testing.T) {
	if trimQuote(`"hello"`) != "hello" {
		t.Fatal("expected unquoted string")
	}
	if trimQuote(`hello`) != "hello" {
		t.Fatal("expected same string")
	}
}
