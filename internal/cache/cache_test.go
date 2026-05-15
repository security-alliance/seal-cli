package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCacheDir(t *testing.T) {
	// Ensure it returns a non-empty path
	dir := defaultCacheDir()
	if dir == "" {
		t.Fatal("expected non-empty cache dir")
	}
}

func TestIndexesDir(t *testing.T) {
	dir := IndexesDir()
	if dir == "" {
		t.Fatal("expected non-empty indexes dir")
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("indexes dir should exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("indexes dir should be a directory")
	}
}

func TestReadIndex(t *testing.T) {
	// Write a temporary index file
	dir := t.TempDir()
	oldCache := os.Getenv("SEAL_CACHE_DIR")
	os.Setenv("SEAL_CACHE_DIR", dir)
	defer os.Setenv("SEAL_CACHE_DIR", oldCache)

	idxPath := filepath.Join(dir, "indexes", "main-index.json")
	_ = os.MkdirAll(filepath.Dir(idxPath), 0755)
	data := []byte(`{"schema_version":1,"branch":"main","commit_sha":"abc","generated_at":"2024-01-01","section_count":0,"sections":[],"sections_by_id":{},"frameworks":[]}`)
	if err := os.WriteFile(idxPath, data, 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	read, err := ReadIndex("main")
	if err != nil {
		t.Fatalf("ReadIndex error: %v", err)
	}
	if string(read) != string(data) {
		t.Fatal("ReadIndex returned wrong content")
	}

	_, err = ReadIndex("develop")
	if err == nil {
		t.Fatal("expected error for missing develop index")
	}
}

func TestLoadManifest(t *testing.T) {
	dir := t.TempDir()
	oldCache := os.Getenv("SEAL_CACHE_DIR")
	os.Setenv("SEAL_CACHE_DIR", dir)
	defer os.Setenv("SEAL_CACHE_DIR", oldCache)

	manifestPath := filepath.Join(dir, "indexes", "manifest.json")
	_ = os.MkdirAll(filepath.Dir(manifestPath), 0755)
	m := Manifest{SchemaVersion: 1, GeneratedAt: "2024-01-01", Files: []ManifestFile{{Path: "main-index.json", SHA256: "aaa", Size: 1}}}
	b, _ := json.Marshal(m)
	_ = os.WriteFile(manifestPath, b, 0644)

	loaded, err := LoadManifest()
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	if loaded.Files[0].Path != "main-index.json" {
		t.Fatal("manifest file path mismatch")
	}

	// missing manifest
	os.Remove(manifestPath)
	_, err = LoadManifest()
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestVerifyIndex(t *testing.T) {
	dir := t.TempDir()
	oldCache := os.Getenv("SEAL_CACHE_DIR")
	os.Setenv("SEAL_CACHE_DIR", dir)
	defer os.Setenv("SEAL_CACHE_DIR", oldCache)

	idxPath := filepath.Join(dir, "indexes", "main-index.json")
	_ = os.MkdirAll(filepath.Dir(idxPath), 0755)
	data := []byte(`{"schema_version":1}`)
	_ = os.WriteFile(idxPath, data, 0644)

	h := sha256.Sum256(data)
	manifestPath := filepath.Join(dir, "indexes", "manifest.json")
	m := Manifest{SchemaVersion: 1, GeneratedAt: "2024-01-01", Files: []ManifestFile{{Path: "main-index.json", SHA256: hex.EncodeToString(h[:]), Size: int64(len(data))}}}
	b, _ := json.Marshal(m)
	_ = os.WriteFile(manifestPath, b, 0644)

	if err := VerifyIndex("main"); err != nil {
		t.Fatalf("VerifyIndex match failed: %v", err)
	}

	// mismatch checksum
	m.Files[0].SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
	b, _ = json.Marshal(m)
	_ = os.WriteFile(manifestPath, b, 0644)
	if err := VerifyIndex("main"); err == nil {
		t.Fatal("expected checksum mismatch")
	}
}

func TestDownloadRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello index"))
	}))
	defer server.Close()

	outPath := filepath.Join(t.TempDir(), "main-index.json")
	if err := DownloadRaw(server.URL, outPath); err != nil {
		t.Fatalf("DownloadRaw error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != "hello index" {
		t.Fatalf("unexpected content: %s", string(data))
	}

	// 404 should error
	server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server404.Close()
	if err := DownloadRaw(server404.URL, outPath+".404"); err == nil {
		t.Fatal("expected error for 404")
	}
}
