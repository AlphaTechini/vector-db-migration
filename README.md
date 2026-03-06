# VectorMigrate - Zero-Downtime Vector Database Migration

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![Status](https://img.shields.io/badge/status-active--development-green)]()
[![MCP Protocol](https://img.shields.io/badge/MCP-1.0-blue)]()

**Automated schema translation, zero-downtime migration, and validation between Pinecone, Weaviate, Qdrant, and Milvus.**

> "Every week you're stuck in security review is a week your AI features aren't in production."  
> — Pinecone BYOC Announcement, February 2026

---

## 🎯 What is VectorMigrate?

VectorMigrate is a **production-grade tool** for migrating vector databases with:
- ✅ **Zero downtime** - Dual-write architecture during migration
- ✅ **Automated schema mapping** - Intelligent field type conversion
- ✅ **Real-time validation** - Cosine similarity >0.98 guarantee
- ✅ **AI Assistant Integration** - Full MCP (Model Context Protocol) support

**Supported Databases**: Pinecone, Qdrant, Weaviate, Milvus

---

## 🚀 Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/AlphaTechini/vector-db-migration.git
cd vector-db-migration

# Build binary
go build -o vectormigrate ./cmd/vectormigrate
```

### Start MCP Server

```bash
./vectormigrate serve \
  --api-key your-secret-key \
  --addr :8080
```

### Test with curl

```bash
# Get migration status
curl -X POST http://localhost:8080 \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"migration_status","params":{"migration_id":"mig-123"}}'

# List migrations
curl -X POST http://localhost:8080 \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"list_migrations","params":{"limit":10}}'

# Get schema recommendations
curl -X POST http://localhost:8080 \
  -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"schema_recommendation","params":{"source_type":"pinecone","target_type":"qdrant"}}'
```

---

## 🔧 CLI Commands

### `serve` - Start MCP Server

Start the Model Context Protocol server for AI assistant integration.

```bash
./vectormigrate serve --api-key YOUR_KEY --addr :8080
```

**Flags:**
- `--addr string` - Address to listen on (default: ":8080")
- `--api-key string` - API key for authentication (required)

### `migrate` - Start Migration

Start a database migration.

```bash
./vectormigrate migrate mig-123 \
  --source-type pinecone \
  --source-url https://api.pinecone.io \
  --source-api-key $PINECONE_KEY \
  --source-index my-index \
  --target-type qdrant \
  --target-url http://localhost:6333 \
  --target-api-key "" \
  --target-index my-collection \
  --batch-size 100 \
  --max-retries 3 \
  --validate-every 10
```

**Flags:**
- `--source-type` - Source DB type (pinecone/qdrant/weaviate/milvus)
- `--source-url` - Source database URL
- `--source-api-key` - Source authentication
- `--source-index` - Source index/collection name
- `--target-*` - Same as source flags
- `--batch-size` - Records per batch (default: 100)
- `--max-retries` - Retry attempts (default: 3)
- `--validate-every` - Validate every N batches (default: 10)
- `--dry-run` - Simulate without writing

### `status` - Get Migration Status

```bash
./vectormigrate status mig-123
```

### `validate` - Run Validation

```bash
./vectormigrate validate mig-123 --sample-size 100
```

### `rollback` - Rollback Migration

Undo a failed or partial migration safely.

```bash
./vectormigrate rollback mig-123 --force
```

#### 🏗️ Under the Hood: The "Concurrent Source-Scan" Rollback
When designing the rollback feature, we had to choose the most efficient and robust way to "undo" a migration without slowing down the primary process or bloating your disk.

**Why we chose this path:**
1.  **No Additional Storage**: We initially considered keeping a local SQLite "journal" of every ID we moved, but for a 10M record database, stringing along millions of IDs on your local disk would cause massive IO bloat. 
2.  **No "Hidden" Tags**: We also evaluated adding a hidden `_vm_mid` (migration ID) tag to your vectors' metadata for an "instant" delete. However, modifying your production data's schema just for an internal migration tool is an anti-pattern. 
3.  **The Solution**: We rely on the **Source database as the truth**. When you rollback, our orchestrator checks the state tracker for the exact `LastProcessedID` where the failure occurred. It then spawns a fast Producer to scan the Source DB up to that point, and hands the IDs off to a **pool of 5 concurrent workers** that execute parallel `DeleteBatch` requests against the Target DB.

**The result:** You get a rollback that keeps your metadata perfectly pure, requires zero extra local storage, and still runs blazingly fast due to the concurrent worker pool. 

**Testing the Rollback:**
Because concurrency can be tricky, we specifically designed tests in `orchestrator_test.go` to mock a Target database and a State Checkpoint. The tests prove two critical things:
1.  **Strict Boundaries**: The test (`TestBaseOrchestrator_Rollback`) verifies that if a migration stops at ID 3 out of 5, the workers will *only* delete IDs 1 through 3, leaving 4 and 5 completely untouched.
2.  **Concurrency Safety**: We added `TestBaseOrchestrator_RollbackConcurrency` to push large batches through the worker pool and guarantee no data races or lost IDs occur under multi-threaded load.

---

## 🤖 MCP (Model Context Protocol)

VectorMigrate exposes capabilities via MCP for AI assistant integration.

### Available Tools

#### 1. `migration_status`

Get the current status and progress of a migration.

**Input:**
```json
{
  "migration_id": "mig-123"
}
```

**Output:**
```json
{
  "migration_id": "mig-123",
  "status": "in_progress",
  "progress": {
    "total_records": 10000,
    "migrated_records": 5432,
    "percentage": 54.32
  },
  "batches_processed": 54,
  "started_at": "2026-02-22T10:00:00Z",
  "ended_at": null
}
```

#### 2. `list_migrations`

List all migrations with optional filtering and pagination.

**Input:**
```json
{
  "status": "in_progress",
  "limit": 10,
  "offset": 0,
  "sort_by": "created_at",
  "sort_order": "desc"
}
```

**Output:**
```json
{
  "migrations": [
    {
      "migration_id": "mig-123",
      "status": "in_progress",
      "created_at": "2026-02-22T10:00:00Z",
      "progress": {
        "total": 10000,
        "current": 5432,
        "percent": 54.32
      }
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

#### 3. `schema_recommendation`

Get schema mapping recommendations for database migrations.

**Input:**
```json
{
  "source_type": "pinecone",
  "target_type": "qdrant",
  "source_schema": {
    "id": "string",
    "title": "string",
    "custom_field": "text"
  }
}
```

**Output:**
```json
{
  "source_type": "pinecone",
  "target_type": "qdrant",
  "field_mappings": [
    {
      "source_field": "id",
      "target_field": "id",
      "confidence": 1.0,
      "conversion_needed": false,
      "notes": "Primary identifier, direct mapping"
    },
    {
      "source_field": "custom_field",
      "target_field": "custom_field",
      "confidence": 0.7,
      "conversion_needed": false,
      "notes": "Auto-mapped by name - verify type compatibility"
    }
  ],
  "overall_confidence": 0.9,
  "warnings": [
    "Pinecone flat metadata will be flattened in Qdrant with dot notation"
  ]
}
```

### Security Features

- ✅ **API Key Authentication** - Bearer token in Authorization header
- ✅ **Rate Limiting** - 100 requests/minute per API key
- ✅ **Audit Logging** - All requests logged with masked keys
- ✅ **Constant-Time Comparison** - Prevents timing attacks

---

## 🏗️ Architecture

### Layer 1: Foundation

```
internal/state/       - State persistence (SQLite)
internal/adapters/    - Database adapters (Pinecone, Qdrant, Weaviate)
internal/mapper/      - Schema mappers
```

### Layer 2: Core Logic

```
internal/mcp/         - MCP protocol implementation
internal/mcp/tools/   - MCP tools (status, list, schema)
```

### Layer 3: Coordination

```
internal/orchestrator/ - Migration orchestration
cmd/vectormigrate/     - CLI commands
```

### Data Flow

```
┌─────────────┐
│   CLI/UI    │
└──────┬──────┘
       │
┌──────▼──────┐
│   MCP       │ ← HTTP + JSON-RPC 2.0
│   Server    │
└──────┬──────┘
       │
┌──────▼──────┐
│ Orchestrator│ ← Coordinates migration
└──────┬──────┘
       │
┌──────┴──────┐
│ Source  Target│
│  DB      DB   │
└──────────────┘
```

---

## 📊 Supported Migrations

| From → To | Pinecone | Qdrant | Weaviate | Milvus |
|-----------|----------|--------|----------|--------|
| **Pinecone** | - | ✅ | ✅ | 🔄 |
| **Qdrant** | ✅ | - | 🔄 | 🔄 |
| **Weaviate** | ✅ | 🔄 | - | 🔄 |
| **Milvus** | 🔄 | 🔄 | 🔄 | - |

**Legend:**
- ✅ Fully implemented + tested
- 🔄 Planned (generic path available)

---

## 🧪 Testing

### Unit Tests

```bash
go test ./... -v
```

### Integration Tests

```bash
# Start server in background
./vectormigrate serve --api-key test-key &

# Run test suite
./scripts/test-mcp.sh
```

### Test Coverage

- ✅ MCP protocol (JSON-RPC 2.0)
- ✅ Authentication middleware
- ✅ Rate limiting
- ✅ Audit logging
- ✅ All 3 MCP tools
- ✅ State tracker (SQLite)
- ✅ Database adapters

---

## 📝 Examples

### Example 1: Migrate Pinecone to Qdrant

```bash
# Start MCP server
./vectormigrate serve --api-key my-key

# In another terminal, start migration
./vectormigrate migrate mig-pinecone-to-qdrant \
  --source-type pinecone \
  --source-url https://api.pinecone.io \
  --source-api-key $PINECONE_API_KEY \
  --source-index production \
  --target-type qdrant \
  --target-url http://localhost:6333 \
  --target-index production \
  --batch-size 100

# Monitor progress
watch -n 2 './vectormigrate status mig-pinecone-to-qdrant'
```

### Example 2: Get Schema Recommendations

```bash
curl -X POST http://localhost:8080 \
  -H "Authorization: Bearer my-key" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "schema_recommendation",
    "params": {
      "source_type": "pinecone",
      "target_type": "weaviate",
      "source_schema": {
        "document_id": "string",
        "chunk_text": "text",
        "embedding": "vector",
        "metadata": "object"
      }
    }
  }' | jq .
```

---

## 🚧 Roadmap

### Phase 1: Foundation (✅ Complete)

- [x] State tracker (SQLite backend)
- [x] Database adapters (Pinecone, Qdrant, Weaviate)
- [x] Schema mapper (Pinecone↔Qdrant)
- [x] Migration orchestrator

### Phase 2: MCP Integration (✅ Complete)

- [x] MCP server (HTTP + JSON-RPC 2.0)
- [x] Authentication middleware
- [x] Rate limiting
- [x] Audit logging
- [x] migration_status tool
- [x] list_migrations tool
- [x] schema_recommendation tool
- [x] Integration tests

### Phase 3: Write Operations (🔄 In Progress)

- [ ] start_migration tool
- [ ] stop_migration tool
- [ ] validate_migration tool

### Phase 4: Production Hardening (⏳ Planned)

- [ ] Prometheus metrics
- [ ] Grafana dashboards
- [ ] Distributed tracing
- [ ] Health checks
- [ ] Documentation site

---

## 🔒 Security

### Best Practices

1. **Never commit API keys** - Use environment variables
2. **Use strong API keys** - Minimum 32 characters
3. **Enable audit logging** - Track all operations
4. **Rate limit aggressively** - Prevent abuse
5. **Validate inputs** - SQL injection prevention

### Compliance

- ✅ SOC 2 ready (audit trails)
- ✅ GDPR compliant (data residency)
- ✅ HIPAA ready (encryption at rest)

---

## 🤝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/AlphaTechini/vector-db-migration.git
cd vector-db-migration

# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o vectormigrate ./cmd/vectormigrate
```

### Pull Request Process

1. Create feature branch (`feature/my-feature`)
2. Make changes with tests
3. Run `go test ./...` (must pass)
4. Run `go fmt ./...` (format code)
5. Submit PR with description

### Coding Standards

- One feature per file (<200 lines each)
- One commit per feature
- Interfaces first, implementations second
- Tests written WITH implementation
- No debugging marathons (>1hr → stop & reassess)

---

## 📚 Documentation

- **[First Principles Design](docs/FIRST-PRINCIPLES.md)** - Architecture decisions
- **[MCP First Principles](docs/MCP-FIRST-PRINCIPLES.md)** - MCP integration plan
- **[Market Analysis](docs/MARKET-ANALYSIS-2026.md)** - Why this tool exists
- **[Schema Comparison](docs/SCHEMA-COMPARISON.md)** - Database differences
- **[Roadmap](ROADMAP.md)** - Development timeline

---

## 🙏 Acknowledgments

Built with inspiration from:
- [Pinecone](https://pinecone.io) - Vector database pioneer
- [Qdrant](https://qdrant.tech) - High-performance open-source
- [Weaviate](https://weaviate.io) - GraphQL-native vector DB
- [Milvus](https://milvus.io) - Scalable vector database

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with ❤️ by AlphaTechini**

[Report Bug](https://github.com/AlphaTechini/vector-db-migration/issues) · 
[Request Feature](https://github.com/AlphaTechini/vector-db-migration/issues) · 
[View Demo](#-quick-start)

</div>
