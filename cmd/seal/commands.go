package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/security-alliance/seal-cli/internal/cache"
	"github.com/security-alliance/seal-cli/internal/index"
	"github.com/security-alliance/seal-cli/internal/search"
)

func loadIndexForBranch(branch string) (*index.Index, error) {
	data, err := cache.ReadIndex(branch)
	if err != nil {
		return nil, fmt.Errorf("no index found for branch %s. Run 'seal update' first", branch)
	}
	return index.Load(data)
}

func runUpdate(args []string) error {
	source := getFlag(args, "--source")
	if source == "" {
		source = "release"
	}
	ref := getFlag(args, "--ref")
	if ref == "" {
		ref = "main"
	}
	branch := getFlag(args, "--branch")
	if branch == "" {
		branch = "both"
	}
	branches := []string{"main", "develop"}
	if branch != "both" {
		branches = []string{branch}
	}

	if source == "release" {
		repo := "security-alliance/frameworks-mcp"
		for _, b := range branches {
			tag, err := cache.FindLatestReleaseTag(repo, b)
			if err != nil {
				return fmt.Errorf("finding latest %s release: %w", b, err)
			}
			asset := fmt.Sprintf("%s-index.json", b)
			out := cache.IndexPath(b)
			fmt.Printf("Downloading %s from release %s ...\n", b, tag)
			if err := cache.DownloadRelease(repo, tag, asset, out); err != nil {
				return fmt.Errorf("error downloading %s: %w", b, err)
			}
			fmt.Printf("Saved to %s\n", out)
		}
		// Try manifest
		// Use the tag from the first branch to download manifest
		firstTag, err := cache.FindLatestReleaseTag(repo, branches[0])
		if err == nil {
			manifestOut := cache.ManifestPath()
			if err := cache.DownloadRelease(repo, firstTag, "manifest.json", manifestOut); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: manifest not available: %v\n", err)
			} else {
				fmt.Printf("Saved manifest to %s\n", manifestOut)
			}
		}
		return nil
	}

	baseURL := fmt.Sprintf("https://raw.githubusercontent.com/security-alliance/frameworks-mcp/%s/indexes/", ref)
	for _, b := range branches {
		url := baseURL + fmt.Sprintf("%s-index.json", b)
		out := cache.IndexPath(b)
		fmt.Printf("Downloading %s from %s ...\n", b, url)
		if err := cache.DownloadRaw(url, out); err != nil {
			return fmt.Errorf("error downloading %s: %w", b, err)
		}
		fmt.Printf("Saved to %s\n", out)
	}
	// Try manifest
	manifestURL := baseURL + "manifest.json"
	manifestOut := cache.ManifestPath()
	if err := cache.DownloadRaw(manifestURL, manifestOut); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: manifest not available: %v\n", err)
	} else {
		fmt.Printf("Saved manifest to %s\n", manifestOut)
	}
	return nil
}

func runList(args []string) error {
	branch := branchFromArgs(args)
	jsonOut := hasFlag(args, "--json")
	idx, err := loadIndexForBranch(branch)
	if err != nil {
		return err
	}
	if jsonOut {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"branch":     idx.Branch,
			"is_draft":   idx.IsDraft(),
			"total":      len(idx.Frameworks),
			"frameworks": idx.Frameworks,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Branch: %s (%d frameworks)\n", idx.Branch, len(idx.Frameworks))
	for _, f := range idx.Frameworks {
		fmt.Printf("  %s  (%d sections)\n", f.Name, f.SectionCount)
	}
	return nil
}

func runSearch(args []string) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "--") {
		return fmt.Errorf("usage: seal search <query> [--branch main|develop|both] [--limit N] [--json]")
	}
	query := trimQuote(args[0])
	branch := branchFromArgs(args)
	jsonOut := hasFlag(args, "--json")
	limitStr := getFlag(args, "--limit")
	limit := 20
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	branches := []string{branch}
	if branch == "both" {
		branches = []string{"main", "develop"}
	}

	var allResults []search.Result
	for _, b := range branches {
		idx, err := loadIndexForBranch(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		r := search.Search(idx, query, limit*2)
		allResults = append(allResults, r...)
	}

	// sort merged results
	for i := 0; i < len(allResults); i++ {
		for j := i + 1; j < len(allResults); j++ {
			if allResults[j].Score > allResults[i].Score {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	if jsonOut {
		var out []map[string]interface{}
		for _, r := range allResults {
			out = append(out, map[string]interface{}{
				"id":            r.Section.ID,
				"branch":        r.Section.Branch,
				"path":          r.Section.Path,
				"page_title":    r.Section.PageTitle,
				"section_title": r.Section.SectionTitle,
				"framework":     r.Section.Framework,
				"score":         r.Score,
				"excerpt":       r.Section.Snippet,
				"source": map[string]string{
					"branch":     r.Section.Branch,
					"site":       strings.TrimSuffix(r.Section.CanonicalURL, "#"+r.Section.HeadingAnchor),
					"repo":       r.Section.GithubURL,
					"commit_sha": r.Section.CommitSHA,
				},
			})
		}
		data, _ := json.MarshalIndent(map[string]interface{}{
			"query":   query,
			"branch":  branch,
			"results": out,
		}, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Query: %q | Branch: %s | Results: %d\n", query, branch, len(allResults))
	for _, r := range allResults {
		fmt.Printf("\n[%s] %s\n  Framework: %s | Score: %.2f\n  %s\n  URL: %s\n  Source: %s\n",
			r.Section.ID,
			r.Section.SectionTitle,
			r.Section.Framework,
			r.Score,
			r.Section.Snippet,
			r.Section.CanonicalURL,
			r.Section.GithubURL,
		)
	}
	return nil
}

func runFetch(args []string) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "--") {
		return fmt.Errorf("usage: seal fetch <section-id> [--branch main|develop] [--json]")
	}
	id := args[0]
	branch := branchFromArgs(args)
	jsonOut := hasFlag(args, "--json")
	idx, err := loadIndexForBranch(branch)
	if err != nil {
		return err
	}
	sec, ok := idx.SectionsByID[id]
	if !ok {
		// fallback: try by path#anchor direct match in Sections slice
		for _, s := range idx.Sections {
			if s.ID == id {
				sec = s
				ok = true
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("section not found: %s", id)
	}
	if jsonOut {
		data, _ := json.MarshalIndent(sec, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("ID: %s\nBranch: %s\nPath: %s\nTitle: %s\nFramework: %s\nSource: %s\nURL: %s\n\n%s\n",
		sec.ID, sec.Branch, sec.Path, sec.SectionTitle, sec.Framework, sec.GithubURL, sec.CanonicalURL, sec.Content)
	return nil
}

func runCompare(args []string) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "--") {
		return fmt.Errorf("usage: seal compare <path> [--left main|develop] [--right main|develop] [--json]")
	}
	pathStr := args[0]
	left := getFlag(args, "--left")
	if left == "" {
		left = "main"
	}
	right := getFlag(args, "--right")
	if right == "" {
		right = "develop"
	}
	jsonOut := hasFlag(args, "--json")
	leftIdx, err := loadIndexForBranch(left)
	if err != nil {
		return fmt.Errorf("error loading left index: %w", err)
	}
	rightIdx, err := loadIndexForBranch(right)
	if err != nil {
		return fmt.Errorf("error loading right index: %w", err)
	}

	var leftSecs, rightSecs []index.Section
	for _, s := range leftIdx.Sections {
		if s.Path == pathStr {
			leftSecs = append(leftSecs, s)
		}
	}
	for _, s := range rightIdx.Sections {
		if s.Path == pathStr {
			rightSecs = append(rightSecs, s)
		}
	}

	leftExists := len(leftSecs) > 0
	rightExists := len(rightSecs) > 0
	var leftPageTitle, rightPageTitle string
	if leftExists {
		leftPageTitle = leftSecs[0].PageTitle
	}
	if rightExists {
		rightPageTitle = rightSecs[0].PageTitle
	}

	leftHash := contentHash(leftSecs)
	rightHash := contentHash(rightSecs)

	status := "unchanged"
	if !leftExists && rightExists {
		status = "added"
	} else if leftExists && !rightExists {
		status = "removed"
	} else if leftHash != rightHash {
		status = "modified"
	}

	added := 0
	removed := 0
	if status == "modified" {
		leftMap := make(map[string]bool)
		for _, s := range leftSecs {
			leftMap[s.SectionTitle] = true
		}
		rightMap := make(map[string]bool)
		for _, s := range rightSecs {
			rightMap[s.SectionTitle] = true
		}
		for t := range rightMap {
			if !leftMap[t] {
				added++
			}
		}
		for t := range leftMap {
			if !rightMap[t] {
				removed++
			}
		}
	}

	result := map[string]interface{}{
		"path": pathStr,
		"left": map[string]interface{}{
			"exists":       leftExists,
			"page_title":   leftPageTitle,
			"sections":     sectionTitles(leftSecs),
			"content_hash": leftHash,
		},
		"right": map[string]interface{}{
			"exists":       rightExists,
			"page_title":   rightPageTitle,
			"sections":     sectionTitles(rightSecs),
			"content_hash": rightHash,
		},
		"changes": map[string]interface{}{
			"status":              status,
			"section_count_delta": len(rightSecs) - len(leftSecs),
			"summary":             fmt.Sprintf("Document '%s' %s: %d added, %d removed", pathStr, status, added, removed),
		},
		"canonical_urls": map[string]string{
			"left":  fmt.Sprintf("https://frameworks.securityalliance.org/%s", pathStr),
			"right": fmt.Sprintf("https://frameworks.securityalliance.dev/%s", pathStr),
		},
		"repo_urls": map[string]string{
			"left":  fmt.Sprintf("https://github.com/security-alliance/frameworks/blob/%s/docs/pages/%s.mdx", leftIdx.CommitSHA, pathStr),
			"right": fmt.Sprintf("https://github.com/security-alliance/frameworks/blob/%s/docs/pages/%s.mdx", rightIdx.CommitSHA, pathStr),
		},
		"commit_shas": map[string]string{
			"left":  leftIdx.CommitSHA,
			"right": rightIdx.CommitSHA,
		},
	}

	if jsonOut {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Path: %s\nLeft (%s): exists=%v title=%q sections=%d hash=%s\nRight (%s): exists=%v title=%q sections=%d hash=%s\nStatus: %s | delta=%d\n", pathStr, left, leftExists, leftPageTitle, len(leftSecs), leftHash, right, rightExists, rightPageTitle, len(rightSecs), rightHash, status, result["changes"].(map[string]interface{})["section_count_delta"])
	if status == "modified" {
		fmt.Printf("Summary: %d added, %d removed\n", added, removed)
	}
	return nil
}

func contentHash(secs []index.Section) string {
	var sb strings.Builder
	for _, s := range secs {
		sb.WriteString(s.Content)
		sb.WriteByte('\n')
	}
	h := sha256.Sum256([]byte(sb.String()))
	return fmt.Sprintf("%x", h)[:16]
}

func sectionTitles(secs []index.Section) []string {
	out := make([]string, len(secs))
	for i, s := range secs {
		out[i] = s.SectionTitle
	}
	return out
}

func runEmergency(args []string) error {
	jsonOut := hasFlag(args, "--json")
	openOpt := hasFlag(args, "--open")
	data := map[string]interface{}{
		"type":     "emergency",
		"telegram": "https://t.me/SEAL_911_bot",
		"handle":   "@SEAL_911_bot",
		"template": map[string][]string{
			"send": {
				"What happened, in one paragraph",
				"Whether funds, users, infra, domains, keys, or governance are at risk",
				"Relevant chains, contracts, addresses, tx hashes, domains, repos, or accounts",
				"Timeline and current status",
				"Actions already taken",
				"Best contact and timezone",
				"Whether public disclosure has happened",
			},
			"do_not_send": {
				"Private keys, seed phrases, passwords, API keys, or unreleased exploit details in public channels",
			},
		},
	}
	if jsonOut {
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Println("SEAL 911 - Emergency War Room")
	fmt.Println()
	fmt.Println("Telegram: https://t.me/SEAL_911_bot")
	fmt.Println()
	fmt.Println("Send:")
	for _, s := range data["template"].(map[string][]string)["send"] {
		fmt.Printf("  - %s\n", s)
	}
	fmt.Println()
	fmt.Println("Do not send:")
	for _, s := range data["template"].(map[string][]string)["do_not_send"] {
		fmt.Printf("  - %s\n", s)
	}
	if openOpt {
		fmt.Println("\nOpening https://t.me/SEAL_911_bot ...")
	}
	return nil
}

func runTips(args []string) error {
	jsonOut := hasFlag(args, "--json")
	openOpt := hasFlag(args, "--open")
	data := map[string]interface{}{
		"type":     "tips",
		"telegram": "https://t.me/SEAL_tips_bot",
		"handle":   "@SEAL_tips_bot",
		"template": map[string][]string{
			"send": {
				"What you observed",
				"Links, usernames, addresses, domains, screenshots, tx hashes, or repo links",
				"Why it appears suspicious",
				"Whether anyone is currently at risk",
				"How SEAL can contact you for follow-up",
			},
			"do_not_send": {
				"Threats, harassment, or legal demands",
			},
		},
	}
	if jsonOut {
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Println("SEAL Tips - Suspicious Activity Reporting")
	fmt.Println()
	fmt.Println("Telegram: https://t.me/SEAL_tips_bot")
	fmt.Println()
	fmt.Println("Send:")
	for _, s := range data["template"].(map[string][]string)["send"] {
		fmt.Printf("  - %s\n", s)
	}
	fmt.Println()
	fmt.Println("Do not send:")
	for _, s := range data["template"].(map[string][]string)["do_not_send"] {
		fmt.Printf("  - %s\n", s)
	}
	if openOpt {
		fmt.Println("\nOpening https://t.me/SEAL_tips_bot ...")
	}
	return nil
}
