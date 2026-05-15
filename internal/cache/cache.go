package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func defaultCacheDir() string {
	if dir := os.Getenv("SEAL_CACHE_DIR"); dir != "" {
		return dir
	}
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "seal")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "seal-cache")
	}
	return filepath.Join(home, ".cache", "seal")
}

func IndexesDir() string {
	dir := filepath.Join(defaultCacheDir(), "indexes")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func IndexPath(branch string) string {
	return filepath.Join(IndexesDir(), fmt.Sprintf("%s-index.json", branch))
}

func ManifestPath() string {
	return filepath.Join(IndexesDir(), "manifest.json")
}

func ReadIndex(branch string) ([]byte, error) {
	p := IndexPath(branch)
	return os.ReadFile(p)
}

type ManifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type Manifest struct {
	SchemaVersion int             `json:"schema_version"`
	GeneratedAt   string          `json:"generated_at"`
	Files         []ManifestFile `json:"files"`
}

func LoadManifest() (*Manifest, error) {
	data, err := os.ReadFile(ManifestPath())
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func VerifyIndex(branch string) error {
	idx := IndexPath(branch)
	data, err := os.ReadFile(idx)
	if err != nil {
		return err
	}
	m, err := LoadManifest()
	if err != nil {
		return fmt.Errorf("no manifest available to verify: %w", err)
	}
	wanted := sha256.Sum256(data)
	wantedHex := hex.EncodeToString(wanted[:])
	for _, f := range m.Files {
		if f.Path == fmt.Sprintf("%s-index.json", branch) {
			if f.SHA256 != wantedHex {
				return fmt.Errorf("checksum mismatch for %s: got %s, want %s", branch, wantedHex, f.SHA256)
			}
			return nil
		}
	}
	return fmt.Errorf("branch not found in manifest: %s", branch)
}

func DownloadRaw(url string, outPath string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(outPath), 0755)
	return os.WriteFile(outPath, data, 0644)
}
