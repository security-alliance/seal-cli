# seal-cli

SEAL Frameworks CLI — read-only retrieval for humans and agents.

## Install

Download the binary from [GitHub Releases](https://github.com/security-alliance/seal-cli/releases) and verify the checksum.

```bash
# Example (Linux amd64)
gh release download -R security-alliance/seal-cli --pattern 'seal_*_linux_amd64.tar.gz'
tar -xzf seal_*_linux_amd64.tar.gz
./seal --help
```

## Quick Start

```bash
# Download index artifacts
seal update

# List frameworks
seal list --branch main

# Search
seal search "ENS resolver risk"
seal search "SEAL 911 war room" --json --limit 5

# Fetch a section
seal fetch incident-management/playbooks/seal-911-war-room-guidelines#seal-911

# Compare branches
seal compare incident-management/playbooks/seal-911-war-room-guidelines --left main --right develop

# Emergency / Tips
seal emergency
seal tips
```

## Commands

| Command | Description |
|---|---|
| `update` | Download latest index artifacts from source (raw GitHub or release) |
| `list` | List available frameworks for a branch |
| `search` | Search sections by query |
| `fetch` | Fetch full section content by stable ID |
| `compare` | Compare document paths across branches |
| `emergency` | Print SEAL 911 contact and template |
| `tips` | Print SEAL tips contact and template |

## Flags

- `--branch main\|develop` — target branch (default: main)
- `--json` — output JSON
- `--limit <n>` — max search results
- `--open` — open Telegram link (opt-in)
- `--source release\|raw` — update source
- `--ref <commit\|branch>` — raw source ref

## Security

- Zero runtime dependencies (standard library only)
- Read-only by default
- No postinstall scripts or install-time network access
- Checksum verification via manifest.json
- `seal update` is the only command that touches the network

## License

CC BY-SA 4.0
