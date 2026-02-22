# VectorMigrate - Zero-Downtime Vector Database Migration

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Status](https://img.shields.io/badge/status-active-development-green)]()

**Automated schema translation, zero-downtime migration, and validation between Pinecone, Weaviate, Qdrant, and Milvus.**

> "Every week you're stuck in security review is a week your AI features aren't in production."  
> — Pinecone BYOC Announcement, February 2026

## 🚀 The Problem

Companies need to migrate vector databases because of:
- **Security blockers** - Enterprise can't use cloud-hosted solutions
- **Cost explosion** - Bills grow 10x at scale, need self-hosted alternatives
- **Vendor lock-in** - Need portability between providers
- **Feature gaps** - Outgrowing current database capabilities

But migration is **painful**:
- Schema differences require manual mapping
- Embedding validation is error-prone
- Downtime is unacceptable for production AI
- One mistake = corrupted vectors = broken semantic search

## ✨ The Solution

VectorMigrate provides:

### 🔁 Automated Schema Translation
- Auto-map metadata, indexes, and configurations
- Support for Pinecone ↔ Weaviate ↔ Qdrant ↔ Milvus
- Custom field transformations via DSL

### ⚡ Zero-Downtime Sync
- Dual-write architecture during migration
- Traffic switching with single command
- Instant rollback if issues detected
- Real-time sync status monitoring

### ✅ Embedding Validation
- Cosine similarity checks (target: >0.98)
- Recall metrics before/after migration
- Statistical sampling (10K+ vectors)
- Full audit reports for compliance

### 📊 Performance Benchmarks
- Latency comparison (p50, p95, p99)
- Throughput analysis (queries/sec)
- Index size optimization recommendations
- Cost projection reports

## 🛠️ Tech Stack

- **Backend**: Go 1.21+ (performance, single binary)
- **CLI**: Cobra for command-line interface
- **API**: REST + gRPC for automation
- **Databases**: Native SDKs for all supported vector DBs
- **Validation**: Custom cosine similarity engine
- **Monitoring**: Prometheus metrics + Grafana dashboards

## 📦 Installation

### From Source (Development)
```bash
git clone https://github.com/AlphaTechini/vector-db-migration.git
cd vector-db-migration
go build ./...
```

### From Binary (Production)
```bash
# Download latest release
curl -LO https://github.com/AlphaTechini/vector-db-migration/releases/latest/download/vectormigrate-linux-amd64
chmod +x vectormigrate-linux-amd64
sudo mv vectormigrate-linux-amd64 /usr/local/bin/vectormigrate
```

### Docker
```bash
docker pull alphatechini/vectormigrate:latest
docker run --rm alphatechini/vectormigrate version
```

## 🚀 Quick Start

### 1. Connect Your Databases
```bash
# Configure source (Pinecone)
vectormigrate config set source \
  --type pinecone \
  --api-key $PINECONE_API_KEY \
  --environment us-west1-gcp \
  --index products

# Configure target (Weaviate)
vectormigrate config set target \
  --type weaviate \
  --url https://your-cluster.weaviate.network \
  --api-key $WEAVIATE_API_KEY \
  --class Products
```

### 2. Validate Connection
```bash
vectormigrate validate connection
```

### 3. Run Schema Mapping
```bash
vectormigrate schema map --output schema-mapping.json
```

### 4. Test Migration (Sample)
```bash
vectormigrate migrate test --limit 1000 --validate
```

### 5. Full Migration with Dual-Write
```bash
vectormigrate migrate full \
  --dual-write \
  --validation strict \
  --rollback-on-error
```

### 6. Switch Traffic
```bash
vectormigrate traffic switch --target
```

## 📖 Documentation

- **[Getting Started Guide](docs/getting-started.md)** - First-time setup
- **[Schema Mapping](docs/schema-mapping.md)** - Field transformations
- **[Migration Modes](docs/migration-modes.md)** - Test, full, incremental
- **[Validation](docs/validation.md)** - Embedding integrity checks
- **[Performance Tuning](docs/performance.md)** - Optimization tips
- **[API Reference](docs/api.md)** - REST API documentation
- **[CLI Reference](docs/cli.md)** - Command reference

## 🏗️ Architecture

```
┌─────────────────┐
│   Source DB     │
│   (Pinecone)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Schema Mapper  │ ←── schema-mapping.json
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  Dual-Write     │─────►│  Validation  │
│  Sync Engine    │      │   Engine     │
└────────┬────────┘      └──────────────┘
         │
         ▼
┌─────────────────┐
│   Target DB     │
│   (Weaviate)    │
└─────────────────┘
```

## 🔒 Security & Compliance

- **Encryption**: TLS 1.3 in transit, AES-256 at rest
- **Audit Logs**: Full migration history with timestamps
- **SOC 2**: Type II compliant (Enterprise plan)
- **HIPAA**: BAA available for healthcare migrations
- **GDPR**: Data processing agreements included

## 💰 Pricing

| Tier | Price | Vectors | Features |
|------|-------|---------|----------|
| **Starter** | $499/migration | Up to 1M | Schema mapping, basic validation, email support |
| **Pro** | $1,999/migration | Up to 50M | Zero-downtime sync, full validation, priority support |
| **Enterprise** | Custom | Unlimited | Dedicated engineer, SLA, compliance docs, on-prem |

👉 **Join the waitlist**: [vectormigrate.dev](https://vectormigrate.dev)

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup
```bash
git clone https://github.com/AlphaTechini/vector-db-migration.git
cd vector-db-migration
go mod download
go test ./...
```

### Running Tests
```bash
# Unit tests
go test ./internal/...

# Integration tests (requires test DB instances)
go test -tags=integration ./internal/...

# End-to-end tests
./scripts/e2e-test.sh
```

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with inspiration from:
- [Pinecone](https://pinecone.io) - Vector database pioneer
- [Weaviate](https://weaviate.io) - Open-source vector search
- [Qdrant](https://qdrant.tech) - High-performance vector engine
- [Milvus](https://milvus.io) - Scalable vector database

## 📬 Contact

- **Website**: [vectormigrate.dev](https://vectormigrate.dev)
- **Twitter**: [@VectorMigrate](https://twitter.com/VectorMigrate)
- **Discord**: [Join our community](https://discord.gg/vectormigrate)
- **Email**: hello@vectormigrate.dev

---

<div align="center">

[Report Bug](https://github.com/AlphaTechini/vector-db-migration/issues) · [Request Feature](https://github.com/AlphaTechini/vector-db-migration/issues) · [Join Waitlist](https://vectormigrate.dev)

</div>
