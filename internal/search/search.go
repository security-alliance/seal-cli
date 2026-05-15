package search

import (
	"strings"
	"unicode"

	"github.com/security-alliance/seal-cli/internal/index"
)

type Result struct {
	Score   float64
	Section index.Section
}

func normalizeQuery(q string) []string {
	fields := strings.FieldsFunc(q, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.ToLower(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func matchesTerms(text string, terms []string) float64 {
	text = strings.ToLower(text)
	score := 0.0
	for _, t := range terms {
		if strings.Contains(text, t) {
			score += 1.0
		}
	}
	return score / float64(len(terms))
}

func Search(idx *index.Index, query string, limit int) []Result {
	terms := normalizeQuery(query)
	if len(terms) == 0 {
		return nil
	}

	var results []Result
	for _, sec := range idx.Sections {
		score := 0.0
		// Title match weight
		score += matchesTerms(sec.SectionTitle, terms) * 4.0
		score += matchesTerms(sec.PageTitle, terms) * 2.5
		score += matchesTerms(sec.Framework, terms) * 1.5
		score += matchesTerms(sec.Path, terms) * 1.2
		score += matchesTerms(sec.Content, terms) * 1.0
		if sec.Description != "" {
			score += matchesTerms(sec.Description, terms) * 1.5
		}
		if score > 0 {
			results = append(results, Result{Score: score, Section: sec})
		}
	}

	// Simple bubble sort by score desc for small result sets
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	return results
}
