package search

import (
	"strings"
	"testing"

	"github.com/security-alliance/seal-cli/internal/index"
)

func TestNormalizeQuery(t *testing.T) {
	terms := normalizeQuery("ENS resolver risk")
	if len(terms) != 3 {
		t.Fatalf("expected 3 terms, got %v", terms)
	}
	if terms[0] != "ens" || terms[1] != "resolver" || terms[2] != "risk" {
		t.Fatalf("unexpected terms: %v", terms)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	idx := &index.Index{
		Sections: []index.Section{
			{ID: "a#x", SectionTitle: "Something", Framework: "x"},
		},
	}
	results := Search(idx, "", 10)
	if len(results) != 0 {
		t.Fatal("expected no results for empty query")
	}
}

func TestSearchRanking(t *testing.T) {
	idx := &index.Index{
		Sections: []index.Section{
			{ID: "a#x", SectionTitle: "ENS resolver risk overview", Framework: "ens", Path: "ens/overview", Content: "ENS resolvers are important"},
			{ID: "b#y", SectionTitle: "Wallet security", Framework: "wallet-security", Path: "wallet-security/overview", Content: "hardware wallets"},
			{ID: "c#z", SectionTitle: "Incident response", Framework: "incident-management", Path: "incident-management/overview", Content: "SEAL 911 war room"},
		},
	}
	results := Search(idx, "ENS resolver", 10)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if !strings.Contains(results[0].Section.SectionTitle, "ENS") {
		t.Fatalf("expected ENS result first, got %s", results[0].Section.SectionTitle)
	}
	if len(results) >= 2 && results[0].Score < results[1].Score {
		t.Fatal("expected first result to have highest score")
	}
}

func TestSearchLimit(t *testing.T) {
	idx := &index.Index{
		Sections: []index.Section{
			{ID: "a#x", SectionTitle: "First", Framework: "x", Content: "alpha beta"},
			{ID: "b#y", SectionTitle: "Second", Framework: "x", Content: "alpha beta"},
			{ID: "c#z", SectionTitle: "Third", Framework: "x", Content: "alpha beta"},
		},
	}
	results := Search(idx, "alpha", 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearchMultiTermScoring(t *testing.T) {
	idx := &index.Index{
		Sections: []index.Section{
			{ID: "a#x", SectionTitle: "Multisig signer onboarding", Framework: "multisig", Content: "onboarding new signers"},
			{ID: "b#y", SectionTitle: "Signer rotation", Framework: "multisig", Content: "rotate signers"},
		},
	}
	results := Search(idx, "signer onboarding", 10)
	if len(results) < 2 {
		t.Fatal("expected at least 2 results")
	}
	if !strings.Contains(results[0].Section.SectionTitle, "onboarding") {
		t.Fatalf("expected onboarding result first, got %s", results[0].Section.SectionTitle)
	}
}
