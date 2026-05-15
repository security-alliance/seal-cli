package index

import (
	"encoding/json"
	"fmt"
)

const CurrentSchemaVersion = 1

type Section struct {
	ID            string   `json:"id"`
	Branch        string   `json:"branch"`
	Path          string   `json:"path"`
	PageTitle     string   `json:"page_title"`
	SectionTitle  string   `json:"section_title"`
	Framework     string   `json:"framework"`
	Tags          []string `json:"tags"`
	Content       string   `json:"content"`
	Snippet       string   `json:"snippet"`
	CanonicalURL  string   `json:"canonical_url"`
	RepoURL       string   `json:"repo_url"`
	SourceURL     string   `json:"source_url"`
	GithubURL     string   `json:"github_url"`
	CommitSHA     string   `json:"commit_sha"`
	HeadingAnchor string   `json:"heading_anchor"`
	SourceFile    string   `json:"source_file"`
	Description   string   `json:"description"`
	Status        string   `json:"status"`
	ContentHash   string   `json:"content_hash"`
}

type Framework struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Framework    string   `json:"framework"`
	SectionCount int      `json:"section_count"`
	Tags         []string `json:"tags"`
}

type Index struct {
	SchemaVersion int                `json:"schema_version"`
	Branch        string             `json:"branch"`
	CommitSHA     string             `json:"commit_sha"`
	GeneratedAt   string             `json:"generated_at"`
	SectionCount  int                `json:"section_count"`
	Sections      []Section          `json:"sections"`
	SectionsByID  map[string]Section `json:"sections_by_id"`
	Frameworks    []Framework        `json:"frameworks"`
}

func Load(data []byte) (*Index, error) {
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	if idx.SchemaVersion != CurrentSchemaVersion && idx.SchemaVersion != 0 {
		return nil, fmt.Errorf("unsupported index schema version: %d", idx.SchemaVersion)
	}
	return &idx, nil
}

func (idx *Index) IsDraft() bool {
	return idx.Branch == "develop"
}
