# seal-cli Implementation Notes

## Decisions (locked)
- Language: Go
- Binary: seal
- Distribution: GitHub Releases only (initially); no npm/pnpm for runtime
- Index artifacts: released from /mcp
- IDs: stable path#anchor, backward incompatible (nobody is using current IDs)
- Emergency handle: @SEAL_911_bot
- Tips handle: @SEAL_tips_bot
- qmd: deferred until benchmarked
- Module: github.com/security-alliance/seal-cli

## Repo Roles
- /vocs: canonical Frameworks content
- /mcp: indexer, MCP server, index artifacts, manifest+checksums
- /seal-cli: Go CLI source + releases

## Implementation Order
1. Fix MCP index schema, IDs, URLs, branch handling, search caching, manifest
2. Build Go CLI skeleton
3. Implement update, list, search, fetch, compare
4. Implement emergency, tips
5. Tests, query-quality checks
6. Release workflow scaffolding

## Env
- Node 20+ for /mcp and /vocs
- pnpm pinned (vocs uses 10.15.0; mcp should pin too)
- Go 1.22+ for /seal-cli
