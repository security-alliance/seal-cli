package main

import (
	"fmt"
	"os"
	"strings"
)

type commandFunc func(args []string) error

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	var run commandFunc
	switch cmd {
	case "update":
		run = runUpdate
	case "list":
		run = runList
	case "search":
		run = runSearch
	case "fetch":
		run = runFetch
	case "compare":
		run = runCompare
	case "emergency":
		run = runEmergency
	case "tips":
		run = runTips
	case "-h", "--help", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`seal - SEAL Frameworks CLI

Usage:
  seal <command> [options]

Commands:
  update       Download latest index artifacts
  list         List available frameworks
  search       Search sections by query
  fetch        Fetch a section by ID
  compare      Compare a document path across branches
  emergency    Show SEAL 911 contact and template
  tips         Show SEAL tips contact and template

Global Flags:
  --branch main|develop  Target branch (default: main)
  --json                 Output JSON
  --limit <n>            Max results for search (default: 20)
  --open                 Open Telegram link (opt-in)
  --source <name>        Update source: release|raw (default: release)
  --ref <commit|branch>  Raw source ref for update
`)
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func getFlag(args []string, name string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1]
		}
	}
	return ""
}

func branchFromArgs(args []string) string {
	b := getFlag(args, "--branch")
	if b == "" {
		return "main"
	}
	if b != "main" && b != "develop" {
		fmt.Fprintf(os.Stderr, "Invalid branch: %s (must be main or develop)\n", b)
		os.Exit(1)
	}
	return b
}

func trimQuote(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s[1 : len(s)-1]
	}
	return s
}
