# structure.md

Project folder structure and high-level mapping of where related logic resides.

This is the recommended reading order for a complete bottom-up understanding of the codebase. Each phase assumes the previous one is already loaded, so by the time you reach a file, every type it imports has already been seen.

---

## Root layout

```
vector-db-migration/
├── cmd/
│   └── vectormigrate/      # CLI entry point + commands
├── internal/
│   ├── adapters/           # Database adapters (Pinecone, Qdrant, Weaviate)
│   ├── mapper/             # Schema mappers between DB pairs
│   ├── mcp/                # MCP protocol + HTTP server
│   │   └── tools/          # MCP tool implementations
│   ├── orchestrator/       # Migration orchestration
│   └── state/              # State persistence (SQLite)
├── docs/                   # Design + analysis docs
├── scripts/                # Integration test scripts
├── web/                    # Landing page assets
├── landing/                # Landing page source
├── README.md
├── ROADMAP.md
├── ROADMAP-MCP.md
└── SETUP.md
```

---

## Reading order

### Phase 0 — Entry point
1. [cmd/vectormigrate/main.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/main.go) — cobra wiring + SIGINT/SIGTERM graceful shutdown.
2. [cmd/vectormigrate/factory.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/factory.go) — the bridge: instantiates adapters, mapper, state tracker, orchestrator. Read this before the layer files; it is the map of which types plug into which.

### Phase 1 — Foundation (Layer 1)
3. [internal/adapters/database.go](file:///C:/PROJECTS/vector-db-migration/internal/adapters/database.go) — the `Database` interface. Everything else implements this.
4. [internal/adapters/pinecone.go](file:///C:/PROJECTS/vector-db-migration/internal/adapters/pinecone.go) → [qdrant.go](file:///C:/PROJECTS/vector-db-migration/internal/adapters/qdrant.go) → [weaviate.go](file:///C:/PROJECTS/vector-db-migration/internal/adapters/weaviate.go) — concrete adapters. Read the interface first, then the impls.
5. [internal/state/tracker.go](file:///C:/PROJECTS/vector-db-migration/internal/state/tracker.go) — SQLite state tracker. Used by the orchestrator for checkpoints and rollback boundaries.
6. [internal/mapper/base.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/base.go) → [schema.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/schema.go) — `SchemaMapper` interface + core mapping logic.
7. [internal/mapper/pinecone_qdrant.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/pinecone_qdrant.go) (most documented pair) → [qdrant_pinecone.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/qdrant_pinecone.go) → [weaviate_pinecone.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/weaviate_pinecone.go) → [weaviate_qdrant.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/weaviate_qdrant.go) → [qdrant_weaviate.go](file:///C:/PROJECTS/vector-db-migration/internal/mapper/qdrant_weaviate.go) — per-pair mappers.

### Phase 2 — MCP server (Layer 2)
8. [internal/mcp/types.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/types.go) — JSON-RPC 2.0 type definitions. Everything in this layer imports these.
9. [internal/mcp/server.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/server.go) — HTTP server wiring.
10. [internal/mcp/handler.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/handler.go) — request dispatch.
11. [internal/mcp/auth.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/auth.go) → [ratelimit.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/ratelimit.go) → [audit.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/audit.go) — middleware stack.
12. [internal/mcp/registry.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/registry.go) — tool registry.
13. [internal/mcp/tools/utils.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/tools/utils.go) → [status.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/tools/status.go) → [list.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/tools/list.go) → [schema.go](file:///C:/PROJECTS/vector-db-migration/internal/mcp/tools/schema.go) — the 3 MCP tools.

### Phase 3 — Coordination (Layer 3)
14. [internal/orchestrator/base.go](file:///C:/PROJECTS/vector-db-migration/internal/orchestrator/base.go) — `MigrationOrchestrator` interface.
15. [internal/orchestrator/orchestrator.go](file:///C:/PROJECTS/vector-db-migration/internal/orchestrator/orchestrator.go) — implementation. Ties adapters + mapper + state together.

### Phase 4 — CLI commands (back to cmd/, everything they use is now loaded)
16. [cmd/vectormigrate/serve.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/serve.go) — MCP server command (also has its own SIGINT handling).
17. [cmd/vectormigrate/migrate.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/migrate.go) — migration command (uses orchestrator).
18. [cmd/vectormigrate/status.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/status.go) → [validate.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/validate.go) → [rollback.go](file:///C:/PROJECTS/vector-db-migration/cmd/vectormigrate/rollback.go).

### Tests
Tests (`*_test.go`) can be skipped on a first pass unless you want to understand intent. The most informative one is [internal/orchestrator/orchestrator_test.go](file:///C:/PROJECTS/vector-db-migration/internal/orchestrator/orchestrator_test.go), which documents the rollback concurrency guarantees described in the README.

---

## Where related logic resides

| Concern | Location |
|---|---|
| CLI entry + signal handling | `cmd/vectormigrate/main.go` |
| Dependency wiring | `cmd/vectormigrate/factory.go` |
| Database interface | `internal/adapters/database.go` |
| Pinecone / Qdrant / Weaviate impls | `internal/adapters/{pinecone,qdrant,weaviate}.go` |
| State persistence + checkpoints | `internal/state/tracker.go` |
| Schema mapper interface + core | `internal/mapper/{base,schema}.go` |
| Per-pair schema mappers | `internal/mapper/{db}_{db}.go` |
| JSON-RPC types | `internal/mcp/types.go` |
| MCP HTTP server | `internal/mcp/server.go` |
| Request dispatch | `internal/mcp/handler.go` |
| Auth / rate limit / audit | `internal/mcp/{auth,ratelimit,audit}.go` |
| Tool registry | `internal/mcp/registry.go` |
| MCP tool implementations | `internal/mcp/tools/*.go` |
| Orchestrator interface | `internal/orchestrator/base.go` |
| Orchestrator implementation | `internal/orchestrator/orchestrator.go` |
| CLI subcommands | `cmd/vectormigrate/{serve,migrate,status,validate,rollback}.go` |

---

## Folder READMEs

Each `internal/` subfolder is expected to ship its own `README.md` documenting architectural decisions and linking to the source files it contains. These are not yet present and should be added for full documentation coverage.

- `internal/adapters/README.md` — TODO
- `internal/mapper/README.md` — TODO
- `internal/mcp/README.md` — TODO
- `internal/mcp/tools/README.md` — TODO
- `internal/orchestrator/README.md` — TODO
- `internal/state/README.md` — TODO