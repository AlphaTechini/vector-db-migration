# VectorMigrate: First Principles Design

**Date**: February 22, 2026  
**Purpose**: Break down the dual-write sync engine to fundamental truths before implementation

---

## 🔍 What Are We Actually Building?

**Core Problem**: Move vectors from Database A to Database B without downtime, data loss, or corruption.

**Constraints**:
1. **Zero Downtime**: Source DB must remain available during migration
2. **Data Integrity**: No vector corruption, no metadata loss
3. **Rollback Capability**: Must be able to revert instantly if issues detected
4. **Validation**: Must prove migration succeeded (cosine similarity >0.98)

---

## 🧱 Fundamental Truths (What Cannot Change)

### Truth 1: Migration is a State Machine

A migration has exactly **4 mutually exclusive states**:

```
NotStarted → InProgress → [Completed | RolledBack]
```

**Implications**:
- Must track state persistently (survives process restart)
- State transitions are one-way (except InProgress → RolledBack)
- Every operation must check current state before executing

### Truth 2: Data Flows in One Direction

```
Source DB → [Read] → [Transform] → [Write] → Target DB
                ↑                        ↑
            Checkpoint               Validation
```

**Implications**:
- Source is read-only during migration (never modified)
- Target is write-only during sync (never read from for migration)
- Transformation is pure function (same input → same output)

### Truth 3: Batches Are the Atomic Unit

We never migrate "one record at a time" because:
- Too many round-trips (performance)
- No atomicity guarantee (partial failures)
- Checkpointing overhead (too frequent)

**Batch Size Trade-off**:
- Too small (<100): High overhead, slow migration
- Too large (>10K): Long retry on failure, memory pressure
- **Sweet spot**: 1000 records per batch (empirically proven)

**Implications**:
- Batch is the unit of checkpointing
- Batch is the unit of retry
- Batch is the unit of validation sampling

### Truth 4: Checkpoints Must Survive Crashes

If the migration process dies at any point, we must resume exactly where we left off.

**What to checkpoint**:
1. Last successfully processed batch ID
2. Current migration state
3. Schema mapping used
4. Validation statistics (sampled so far)
5. Error counts and details

**Checkpoint Frequency**:
- After every batch (safest, but slower)
- After every N batches (faster, but more rework on crash)
- **Decision**: After every batch (correctness > speed for V1)

### Truth 5: Validation Happens in Parallel

We cannot wait until the end to validate—that's too late to fix issues.

**Validation Strategy**:
- Sample 10% of batches in real-time
- For each sampled batch:
  - Read source record
  - Read migrated target record
  - Compute cosine similarity
  - Verify metadata completeness
- Track running statistics (avg, min, max similarity)
- Alert if min similarity < threshold (0.98)

**Implications**:
- Validation runs async (doesn't block migration)
- Validation can trigger rollback (if failure rate too high)
- Validation results must be checkpointed too

### Truth 6: Rollback is Just Stopping + Switching Traffic

Rollback does NOT mean "undo all writes to target". It means:
1. Stop the migration immediately
2. Ensure traffic is pointing to source (already is, since we haven't switched yet)
3. Mark migration as failed

**Key Insight**: Until we switch traffic, source is still the active DB. Rollback is trivial.

**Post-Cutover Rollback** (Phase 2 problem):
- After traffic switch, target is active
- Rollback requires bidirectional sync or manual intervention
- **Out of scope for V1** (we'll design this in Phase 2)

---

## 📋 What Must We Build First? (Dependency Order)

### Layer 1: Foundation (Must Exist Before Anything Else)

#### 1.1 Migration State Tracker
**Purpose**: Persist and retrieve migration state  
**Why first**: Every other component depends on knowing current state  
**Interface**:
```go
type MigrationState string
const (
    StateNotStarted   MigrationState = "not_started"
    StateInProgress   MigrationState = "in_progress"
    StateCompleted    MigrationState = "completed"
    StateRolledBack   MigrationState = "rolled_back"
)

type StateTracker interface {
    GetState(migrationID string) (MigrationState, error)
    SetState(migrationID string, state MigrationState) error
    GetCheckpoint(migrationID string) (*Checkpoint, error)
    SaveCheckpoint(migrationID string, cp *Checkpoint) error
}
```

**Implementation Decision**: SQLite (embedded, ACID, no external deps)

---

#### 1.2 Database Adapters (Source + Target)
**Purpose**: Abstract database-specific operations  
**Why first**: Cannot read/write without adapters  
**Interface**:
```go
type Database interface {
    Connect(config DBConfig) error
    Close() error
    GetRecord(id string) (*Record, error)
    GetBatch(afterID string, limit int) ([]Record, error)
    UpsertRecord(record *Record) error
    UpsertBatch(records []Record) error
    ValidateConnection() error
    GetStats() (*DBStats, error)
}
```

**V1 Adapters to Implement**:
- `PineconeAdapter` (implements `Database`)
- `QdrantAdapter` (implements `Database`)
- `WeaviateAdapter` (implements `Database`)

**Design Rule**: One file per adapter (`adapter_pinecone.go`, `adapter_qdrant.go`, etc.)

---

### Layer 2: Core Logic (Depends on Layer 1)

#### 2.1 Schema Mapper
**Purpose**: Transform records from source schema to target schema  
**Why second**: Cannot transform without adapters to read source/write target  
**Interface**:
```go
type SchemaMapper interface {
    InferSchema(sampleRecords []Record) (*Schema, error)
    MapRecord(source *Record, targetSchema *Schema) (*Record, error)
    ValidateMapping(source *Record, target *Record) error
}
```

**V1 Mappers**:
- `PineconeToQdrantMapper`
- `PineconeToWeaviateMapper`
- `GenericToSochDBMapper`

**Design Rule**: One file per mapper (`mapper_pinecone_qdrant.go`, etc.)

---

#### 2.2 Batch Processor
**Purpose**: Read-transform-write one batch  
**Why second**: Core unit of work, depends on adapters + mapper  
**Interface**:
```go
type BatchProcessor struct {
    sourceDB  Database
    targetDB  Database
    mapper    SchemaMapper
    batchSize int
}

func (p *BatchProcessor) ProcessBatch(afterID string) (*BatchResult, error)
```

**Responsibilities**:
- Read batch from source (using `sourceDB.GetBatch()`)
- Transform each record (using `mapper.MapRecord()`)
- Write batch to target (using `targetDB.UpsertBatch()`)
- Return result (count, errors, duration)

**Design Rule**: Single file (`batch_processor.go`)

---

### Layer 3: Coordination (Depends on Layer 2)

#### 3.1 Migration Orchestrator
**Purpose**: Coordinate batches, checkpointing, validation  
**Why third**: Needs batch processor + state tracker  
**Interface**:
```go
type Orchestrator struct {
    stateTracker StateTracker
    batchProc    *BatchProcessor
    validator    *Validator
    config       MigrationConfig
}

func (o *Orchestrator) StartMigration() error
func (o *Orchestrator) ResumeMigration() error
func (o *Orchestrator) StopMigration() error
func (o *Orchestrator) Rollback() error
```

**Responsibilities**:
- Check current state (can only start if `NotStarted` or `InProgress`)
- Loop: Process batch → Checkpoint → Validate sample
- Handle errors (retry logic, stop on threshold)
- Update state on completion/rollback

**Design Rule**: Single file (`orchestrator.go`)

---

#### 3.2 Validator
**Purpose**: Sample and validate migrated records  
**Why third**: Runs parallel to orchestration, needs adapters  
**Interface**:
```go
type Validator struct {
    sourceDB  Database
    targetDB  Database
    sampleRate float64  // 0.10 = 10%
    minSimilarity float64  // 0.98
}

func (v *Validator) ValidateBatch(batch []Record) (*ValidationResult, error)
func (v *Validator) GetStats() *ValidationStats
```

**Responsibilities**:
- Randomly select batches for validation (based on sample rate)
- For each selected batch:
  - Read source records
  - Read target records
  - Compute cosine similarity
  - Check metadata completeness
- Track running stats (avg, min, max similarity)
- Alert if min < threshold

**Design Rule**: Single file (`validator.go`)

---

### Layer 4: User Interface (Depends on Layer 3)

#### 4.1 CLI Commands
**Purpose**: User-facing commands to control migrations  
**Why fourth**: Wraps orchestrator with UX  
**Commands**:
- `vectormigrate migrate start --source pinecone --target qdrant`
- `vectormigrate migrate resume --id <migration-id>`
- `vectormigrate migrate stop --id <migration-id>`
- `vectormigrate migrate rollback --id <migration-id>`
- `vectormigrate status --id <migration-id>`

**Design Rule**: One file per command (`cmd_migrate_start.go`, `cmd_status.go`, etc.)

---

## 🎯 Implementation Order (Week by Week)

### Week 1: Foundation
- [ ] **Day 1-2**: Define interfaces (`database.go`, `state_tracker.go`, `mapper.go`)
- [ ] **Day 3-4**: Implement Pinecone adapter (`adapter_pinecone.go`)
- [ ] **Day 5**: Implement Qdrant adapter (`adapter_qdrant.go`)
- [ ] **Day 6-7**: Implement SQLite state tracker (`state_tracker_sqlite.go`)

### Week 2: Core Logic
- [ ] **Day 1-2**: Implement Pinecone→Qdrant mapper (`mapper_pinecone_qdrant.go`)
- [ ] **Day 3-4**: Implement batch processor (`batch_processor.go`)
- [ ] **Day 5-6**: Implement validator (`validator.go`)
- [ ] **Day 7**: Integration test (read from Pinecone, write to Qdrant, validate)

### Week 3: Coordination
- [ ] **Day 1-3**: Implement orchestrator (`orchestrator.go`)
- [ ] **Day 4-5**: Add checkpointing logic
- [ ] **Day 6-7**: Add rollback mechanism

### Week 4: CLI + Testing
- [ ] **Day 1-3**: Implement CLI commands
- [ ] **Day 4-5**: End-to-end testing (1M vectors)
- [ ] **Day 6-7**: Bug fixes, documentation

---

## 📏 Coding Standards (Non-Negotiable)

### Rule 1: One Feature Per File
- ❌ BAD: `migration.go` (500 lines, does everything)
- ✅ GOOD: `orchestrator.go`, `batch_processor.go`, `validator.go` (each <200 lines)

### Rule 2: One Commit Per Feature
- ❌ BAD: "Implement entire migration engine" (20 files changed)
- ✅ GOOD: 
  - "Add Pinecone adapter" (1 file)
  - "Add Qdrant adapter" (1 file)
  - "Add SQLite state tracker" (1 file)

### Rule 3: Interfaces First
- Define interface before implementation
- Program to interface, not concrete type
- Makes testing easier (mock implementations)

### Rule 4: Test Each Component Independently
- Unit tests for each adapter
- Unit tests for each mapper
- Integration tests for orchestrator
- End-to-end tests for full migration

### Rule 5: No Debugging Marathons
- If stuck for >1 hour, stop and reassess
- Write a failing test to isolate issue
- Ask for help with specific error message
- Never commit broken code "to debug later"

---

## ✅ Next Immediate Action

**Start with Layer 1, Item 1.1**: Define the `StateTracker` interface and implement SQLite backend.

**File**: `internal/state/tracker.go`  
**Lines of Code**: ~150 (interface + structs + SQLite impl)  
**Tests**: `internal/state/tracker_test.go` (~100 lines)  
**Commit Message**: "Add migration state tracker with SQLite backend"

**Why this first**:
1. Smallest surface area (simple CRUD operations)
2. No external dependencies (SQLite via `modernc.org/sqlite`)
3. Every other component depends on it
4. Easy to test in isolation

---

**Ready to implement?** Let's start with the state tracker.
