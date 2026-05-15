#!/usr/bin/env bash
set -euo pipefail

# Acceptance test for seal search quality
# Requires seal binary built at ./seal-test or path in SEAL env var

SEAL="${SEAL:-./seal-test}"
PASS=0
FAIL=0

function check_search() {
  local query="$1"
  local expected_id="$2"
  local branch="${3:-main}"
  local limit="${4:-5}"

  echo "TEST: search '$query' (branch=$branch) expecting top-$limit to contain $expected_id"
  ids=$($SEAL search "$query" --branch "$branch" --json --limit "$limit" 2>/dev/null | jq -r '.results[].id')
  if echo "$ids" | grep -qF "$expected_id"; then
    echo "  PASS"
    PASS=$((PASS+1))
  else
    echo "  FAIL"
    echo "  IDs found:"
    echo "$ids" | head -n 10 | sed 's/^/    /'
    FAIL=$((FAIL+1))
  fi
}

# Representative queries from issue acceptance criteria
check_search "ENS resolver risk"              "ens/interface-compliance#verify-resolver-interface-support"
check_search "SEAL 911 war room"              "incident-management/playbooks/seal-911-war-room-guidelines#seal-911"
check_search "signer onboarding"              "multisig-for-protocols/runbooks/signer-rotation#signer-rotation-runbook"
check_search "dependency risk"                "supply-chain/web3-supply-chain-threats#smart-contract-dependency-risks"
check_search "signed commits"                 "devsecops/code-signing#1-require-signed-commits-on-protected-branches"

echo ""
echo "Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
