# MCP Integration: First Principles Design

**Date**: February 22, 2026  
**Purpose**: Break down MCP (Model Context Protocol) integration to fundamental truths before implementation

---

## 🔍 What Are We Actually Building?

**Core Problem**: AI agents need context about vector databases during migration, but currently operate blind.

**Solution**: Expose VectorMigrate capabilities via MCP (Model Context Protocol) so AI assistants can:
- Query migration status
- Start/stop migrations
- Get schema recommendations
- Validate migrations
- Access logs and metrics

**Constraints**:
1. **Security**: No direct database access via MCP
2. **Read-only by default**: Write operations require explicit authorization
3. **Audit trail**: All MCP operations logged
4. **Rate limiting**: Prevent abuse

---

## 🧱 Fundamental Truths (What Cannot Change)

### Truth 1: MCP is a Bridge, Not a Backend

MCP servers expose capabilities to AI assistants. They do NOT:
- Store data (use existing state tracker)
- Execute migrations directly (use orchestrator)
- Manage connections (use adapter factory)

**Implication**: MCP server is a thin wrapper around existing components.

### Truth 2: AI Assistants Need Structured Responses

AI assistants work best with:
- Clear schemas (JSON Schema)
- Enumerated options (not free text)
- Examples in responses
- Error messages with suggestions

**Implication**: Every MCP tool must have JSON Schema + examples.

### Truth 3: Operations Must Be Idempotent Where Possible

AI assistants may retry operations. We must handle:
- Duplicate "start migration" requests → Return existing migration ID
- Repeated "get status" calls → Cache recent results
- Cancelled then restarted migrations → Create new ID, don't reuse

**Implication**: Track operation history, use idempotency keys.

### Truth 4: Context Window is Limited

AI assistants have limited context (typically 8K-128K tokens). We must:
- Keep responses concise (<1K tokens typical)
- Paginate large result sets
- Summarize when possible
- Provide "get more details" tools

**Implication**: Default to summaries, offer drill-down.

### Truth 5: Security Boundaries are Critical

MCP servers run with AI assistant access. We must:
- Never expose API keys in responses
- Sanitize URLs (remove credentials)
- Validate all inputs (SQL injection, path traversal)
- Rate limit per client
- Log all operations

**Implication**: Security review before any MCP tool.

### Truth 6: Testing Requires Mock AI Assistants

We cannot test MCP with unit tests alone. We need:
- Mock AI assistant (simulates Claude/GPT)
- Recorded conversations (for regression testing)
- Integration tests with real MCP clients
- Security penetration testing

**Implication**: Build mock assistant early.

---

## 📋 What Must We Build First? (Dependency Order)

### Layer 1: Foundation (Must Exist Before Anything Else)

#### 1.1 MCP Server Skeleton
**Purpose**: Basic MCP server structure  
**Why first**: All other components depend on MCP protocol  
**Interface**:
```go
type MCPServer interface {
    Start(ctx context.Context, addr string) error
    Stop() error
    RegisterTool(name string, handler ToolHandler) error
}
```

**Implementation Decision**: Use official MCP SDK if available, otherwise minimal HTTP+JSON-RPC.

---

#### 1.2 Tool Registry
**Purpose**: Register and manage MCP tools  
**Why first**: Tools must be registered before exposure  
**Interface**:
```go
type ToolRegistry interface {
    Register(tool *Tool) error
    Get(name string) (*Tool, error)
    List() []*Tool
    Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
}
```

**Implementation**: In-memory registry with validation.

---

### Layer 2: Core Tools (Depends on Layer 1)

#### 2.1 Migration Status Tool
**Purpose**: Query migration status  
**Why second**: Most common operation, read-only (safe)  
**Schema**:
```json
{
  "name": "migration_status",
  "description": "Get current status of a migration",
  "inputSchema": {
    "type": "object",
    "properties": {
      "migration_id": {"type": "string", "description": "Migration ID"}
    },
    "required": ["migration_id"]
  }
}
```

**Response**:
```json
{
  "migration_id": "mig-123",
  "status": "in_progress",
  "progress": {
    "total_records": 10000,
    "migrated_records": 5000,
    "percentage": 50.0
  },
  "started_at": "2026-02-22T10:00:00Z"
}
```

---

#### 2.2 List Migrations Tool
**Purpose**: List all migrations  
**Why second**: Discovery operation, read-only  
**Schema**: Supports filtering by status, date range

---

#### 2.3 Schema Recommendation Tool
**Purpose**: Get schema mapping recommendations  
**Why second**: Helps AI assistants guide users  
**Input**: Source schema, target type  
**Output**: Recommended field mappings, type conversions

---

### Layer 3: Write Operations (Depends on Layer 2)

#### 3.1 Start Migration Tool
**Purpose**: Start a new migration  
**Why third**: Write operation, requires authorization  
**Security**: Require API key or OAuth token

---

#### 3.2 Stop Migration Tool
**Purpose**: Stop in-progress migration  
**Why third**: Write operation, irreversible

---

#### 3.3 Validate Migration Tool
**Purpose**: Run validation checks  
**Why third**: Can be expensive, rate limit

---

### Layer 4: Advanced Features (Optional)

#### 4.1 Real-time Progress Stream
**Purpose**: Stream progress updates  
**Why optional**: Requires WebSocket/SSE, complex

#### 4.2 Log Access Tool
**Purpose**: Query migration logs  
**Why optional**: Sensitive data, careful filtering needed

#### 4.3 Metrics Tool
**Purpose**: Get performance metrics  
**Why optional**: Requires metrics backend

---

## 🎯 Implementation Order (Week by Week)

### Week 1: MCP Foundation
- [ ] **Day 1-2**: Research MCP specification, choose SDK
- [ ] **Day 3-4**: Implement MCP server skeleton
- [ ] **Day 5**: Implement tool registry
- [ ] **Day 6-7**: Add logging + security middleware

### Week 2: Read-Only Tools
- [ ] **Day 1-2**: migration_status tool
- [ ] **Day 3**: list_migrations tool
- [ ] **Day 4-5**: schema_recommendation tool
- [ ] **Day 6-7**: Testing with mock AI assistant

### Week 3: Write Operations
- [ ] **Day 1-2**: start_migration tool (with auth)
- [ ] **Day 3**: stop_migration tool
- [ ] **Day 4**: validate_migration tool
- [ ] **Day 5-7**: Security audit + penetration testing

### Week 4: Polish + Documentation
- [ ] **Day 1-2**: Error message improvements
- [ ] **Day 3-4**: User documentation
- [ ] **Day 5**: Example conversations
- [ ] **Day 6-7**: Beta testing with real AI assistants

---

## 📏 Coding Standards (Non-Negotiable)

### Rule 1: One Tool Per File
- ❌ BAD: `tools.go` (500 lines, all tools)
- ✅ GOOD: `status_tool.go`, `list_tool.go`, `start_tool.go` (each <150 lines)

### Rule 2: JSON Schema for Every Tool
- Define input schema explicitly
- Include descriptions for all fields
- Provide examples
- Validate inputs before execution

### Rule 3: Security First
- No secrets in responses
- Input validation (whitelist, not blacklist)
- Rate limiting per client
- Audit logging for all operations

### Rule 4: Test with Mock Assistant
- Unit tests for tool logic
- Integration tests with mock AI assistant
- Recorded conversations for regression testing
- Security penetration testing

### Rule 5: No Debugging Marathons
- If stuck for >1 hour, stop and reassess
- Write failing test to isolate issue
- Ask for help with specific error message
- Never commit broken code "to debug later"

---

## ⚠️ Risks & Mitigations

### Risk 1: Security Vulnerabilities
**Impact**: Database credentials exposed  
**Mitigation**: 
- Security review before any tool
- Automated secret scanning in CI
- Penetration testing
- Bug bounty program

### Risk 2: Performance Issues
**Impact**: Slow responses, timeouts  
**Mitigation**:
- Caching for read operations
- Pagination for large result sets
- Rate limiting
- Performance benchmarks

### Risk 3: Breaking Changes
**Impact**: AI assistants break after update  
**Mitigation**:
- Semantic versioning
- Backward compatibility for 1 major version
- Deprecation warnings
- Migration guide

### Risk 4: AI Assistant Misuse
**Impact**: Unauthorized operations  
**Mitigation**:
- Explicit authorization for write operations
- Confirmation dialogs for dangerous operations
- Audit trail
- Rollback capability

---

## ✅ Success Criteria

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

**Start with Layer 1, Item 1.1**: MCP Server Skeleton

**File**: `internal/mcp/server.go`  
**Lines of Code**: ~150 (interface + minimal HTTP server)  
**Tests**: `internal/mcp/server_test.go` (~100 lines)  
**Commit Message**: "Add MCP server skeleton (Layer 1, Item 1.1)"

**Why this first**:
1. Smallest surface area (just protocol handling)
2. No external dependencies (pure Go HTTP+JSON)
3. All other components depend on it
4. Easy to test in isolation

---

**Ready to implement?** Let's start with the MCP server skeleton.
