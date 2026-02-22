# MCP Integration Roadmap

**Created**: February 22, 2026  
**Based on**: MCP-FIRST-PRINCIPLES.md

---

## 🎯 Implementation Decisions (Confirmed)

1. **Protocol**: HTTP+JSON-RPC 2.0 (Option B - Custom minimal implementation)
2. **Deployment**: Separate process (`vectormigrate-mcp-server`)
3. **Authentication**: API Key (Bearer token in Authorization header)
4. **Priority Tool**: `migration_status` (most common, read-only, safe)
5. **Testing**: Unit tests only for now (mock AI assistant deferred)

---

## 📅 Phase 4: MCP Integration (8 Weeks)

### Week 1: MCP Foundation (Days 1-7)

**Milestone**: HTTP server accepting JSON-RPC requests

#### Day 1-2: Server Skeleton
- [ ] `internal/mcp/server.go` - HTTP server with JSON-RPC parser
- [ ] `internal/mcp/types.go` - JSON-RPC types (Request, Response, Error)
- [ ] `internal/mcp/registry.go` - Tool registry (static registration)
- [ ] `internal/mcp/handler.go` - Route requests to registered tools
- [ ] Unit tests for JSON-RPC parsing

#### Day 3-4: First Tool (migration_status)
- [ ] `internal/mcp/tools/status.go` - migration_status tool implementation
- [ ] JSON Schema for input validation
- [ ] Example responses in schema
- [ ] Unit tests for tool logic

#### Day 5-7: Security Basics
- [ ] `internal/mcp/auth.go` - API key validation middleware
- [ ] `internal/mcp/ratelimit.go` - Rate limiting (100 req/min per key)
- [ ] `internal/mcp/audit.go` - Audit logging (all requests)
- [ ] Security review of implementation

**Deliverable**: Working MCP server with 1 tool + auth

---

### Week 2: More Read-Only Tools (Days 8-14)

**Milestone**: Discovery + recommendation tools

#### Day 8-9: List Migrations Tool
- [ ] `internal/mcp/tools/list.go` - list_migrations tool
- [ ] Support filtering by status, date range
- [ ] Pagination support (limit, offset)
- [ ] Unit tests

#### Day 10-11: Schema Recommendation Tool
- [ ] `internal/mcp/tools/schema.go` - schema_recommendation tool
- [ ] Analyze source schema
- [ ] Recommend field mappings for target type
- [ ] Include confidence scores
- [ ] Unit tests

#### Day 12-14: Polish + Documentation
- [ ] Error message improvements (include suggestions)
- [ ] Tool documentation (README.md for each tool)
- [ ] Example curl commands
- [ ] Integration tests (HTTP client → server)

**Deliverable**: 3 read-only tools working + documented

---

### Week 3-4: Write Operations (Days 15-28)

**Milestone**: Start/stop/validate migrations via MCP

#### Week 3: Start Migration Tool
- [ ] `internal/mcp/tools/start.go` - start_migration tool
- [ ] Require API key authorization
- [ ] Validate all inputs (source/target config)
- [ ] Return migration ID immediately
- [ ] Background execution (don't block response)
- [ ] Unit tests + integration tests

#### Week 4: Stop + Validate Tools
- [ ] `internal/mcp/tools/stop.go` - stop_migration tool
- [ ] `internal/mcp/tools/validate.go` - validate_migration tool
- [ ] Security audit (penetration testing)
- [ ] Rate limiting for expensive operations
- [ ] Confirmation for destructive operations

**Deliverable**: Full migration lifecycle via MCP

---

### Week 5-6: Production Hardening (Days 29-42)

**Milestone**: Production-ready MCP server

#### Observability
- [ ] Prometheus metrics (request count, latency, errors)
- [ ] Grafana dashboard
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Log aggregation (structured JSON logs)

#### Reliability
- [ ] Graceful shutdown (drain connections)
- [ ] Health check endpoint (`/healthz`)
- [ ] Readiness probe (`/readyz`)
- [ ] Circuit breaker for downstream calls

#### Security
- [ ] Automated secret scanning in CI
- [ ] Dependency vulnerability scanning
- [ ] Penetration testing report
- [ ] Security documentation

**Deliverable**: Production-hardened MCP server

---

### Week 7-8: Documentation + Beta (Days 43-56)

**Milestone**: Ready for beta testing

#### Documentation
- [ ] User guide (setup, configuration, tools reference)
- [ ] API reference (JSON-RPC schemas for all tools)
- [ ] Example conversations (AI assistant scenarios)
- [ ] Troubleshooting guide

#### Beta Testing
- [ ] Internal testing (use VectorMigrate MCP ourselves)
- [ ] External beta (5-10 friendly users)
- [ ] Feedback collection + iteration
- [ ] Bug fixes + performance improvements

**Deliverable**: Beta-ready MCP server with docs

---

## 🔬 Testing Strategy (Deferred)

### Mock AI Assistant (Phase 4.4 or later)
- [ ] Simulate AI assistant sending JSON-RPC requests
- [ ] Validate response structure
- [ ] Test error handling
- [ ] Record conversations for regression testing

### Recorded Conversations (Phase 4.4 or later)
- [ ] `testdata/conversations/*.yaml` files
- [ ] Replay conversations on every build
- [ ] Detect breaking changes automatically

**Decision**: Defer until after core tools are stable

---

## 📊 Success Metrics

### Technical
- [ ] All tools respond in <500ms (p95)
- [ ] Zero security vulnerabilities (pen test passed)
- [ ] 90%+ test coverage
- [ ] Works with Claude, GPT-4, Gemini

### User Experience
- [ ] AI assistants can complete migration without human intervention
- [ ] Clear error messages with suggestions
- [ ] Progress updates every 5 seconds
- [ ] Easy to troubleshoot issues

### Adoption
- [ ] 10+ AI assistants integrated in first month
- [ ] Documentation complete with examples
- [ ] Community contributions (tools, extensions)

---

## 🚀 Next Immediate Action

**Start Week 1, Day 1**: MCP Server Skeleton

**Files to Create**:
1. `internal/mcp/server.go` (~100 lines)
2. `internal/mcp/types.go` (~80 lines)
3. `internal/mcp/registry.go` (~60 lines)
4. `internal/mcp/handler.go` (~80 lines)
5. `internal/mcp/server_test.go` (~100 lines)

**Total**: ~420 lines of code  
**Time Estimate**: 1-2 days  
**Commit Strategy**: One commit per file

**Ready to start!** 🎯
