# Vector Database Schema Comparison

**Purpose**: Understand schema differences between Pinecone, Weaviate, Qdrant, and Milvus to design automated migration mapping.

---

## 1. Pinecone Schema Structure

### Record Format

```json
{
  "id": "document1#chunk1",           // Unique string identifier
  "values": [0.023, -0.032, ...],     // Dense vector (required)
  "sparse_values": {                   // Sparse vector (optional, for hybrid)
    "values": [1.79, 0.41, ...],
    "indices": [822745112, 1009084850, ...]
  },
  "metadata": {                        // Flat JSON (no nested objects)
    "document_id": "document1",        // String
    "chunk_number": 1,                 // Number (int or float → 64-bit float)
    "is_public": true,                 // Boolean
    "tags": ["tutorial", "intro"],    // List of strings
    "created_at": "2024-01-15"        // String (ISO date)
  }
}
```

### Key Characteristics

| Feature | Details |
|---------|---------|
| **ID Format** | String (up to 1KB), structured IDs recommended (`tenant_id#doc_id#chunk_id`) |
| **Vector Types** | Dense, Sparse, or Hybrid (both) |
| **Metadata Limit** | 40 KB per record |
| **Metadata Types** | String, Number (→float64), Boolean, List[string] |
| **Nested Objects** | ❌ NOT supported |
| **Namespaces** | ✅ Supported (logical partitioning) |
| **Index Types** | Serverless (default), Pod-based (legacy) |

### Index Configuration

```python
pc.create_index(
    name="my-index",
    dimension=1536,
    metric="cosine",      # cosine, dotproduct, euclidean
    spec=ServerlessSpec(
        cloud="aws",
        region="us-east-1"
    )
)
```

---

## 2. Weaviate Schema Structure

### Collection (Class) Definition

```json
{
  "class": "DocumentChunk",
  "description": "A chunk of a document for RAG",
  "vectorizer": "text2vec-openai",
  "vectorIndexType": "hnsw",  // hnsw, flat, or dynamic
  "vectorIndexConfig": {
    "distance": "cosine",     // cosine, dot, l2-squared, hamming
    "maxConnections": 32,
    "efConstruction": 128,
    "cleanupIntervalSeconds": 300
  },
  "properties": [
    {
      "name": "document_id",
      "dataType": ["string"],
      "description": "Parent document ID"
    },
    {
      "name": "chunk_number",
      "dataType": ["int"],
      "description": "Chunk sequence number"
    },
    {
      "name": "chunk_text",
      "dataType": ["text"],
      "description": "Original text content"
    },
    {
      "name": "created_at",
      "dataType": ["date"],
      "description": "Creation timestamp"
    },
    {
      "name": "tags",
      "dataType": ["string[]"],
      "description": "Array of tags"
    },
    {
      "name": "is_public",
      "dataType": ["boolean"],
      "description": "Public visibility flag"
    }
  ]
}
```

### Object (Record) Format

```json
{
  "id": "abc123-def456",        // UUID (auto-generated or custom)
  "class": "DocumentChunk",
  "vector": [0.023, -0.032, ...],
  "properties": {                // Typed properties (schema-enforced)
    "document_id": "document1",
    "chunk_number": 1,
    "chunk_text": "First chunk...",
    "created_at": "2024-01-15T00:00:00Z",
    "tags": ["tutorial", "intro"],
    "is_public": true
  },
  "creationTimeUnix": 1705276800000,
  "lastUpdateTimeUnix": 1705276800000
}
```

### Key Characteristics

| Feature | Details |
|---------|---------|
| **Schema** | ✅ Required (strongly typed) |
| **ID Format** | UUID (v4, auto-generated or custom) |
| **Vector Types** | Dense only |
| **Properties** | Typed: string, int, number, boolean, date, text, string[], etc. |
| **Nested Objects** | ✅ Supported via `object` and `object[]` types |
| **Multi-Tenancy** | ✅ Per-tenant isolation |
| **Index Types** | HNSW (default), Flat, Dynamic (auto-switch) |
| **Quantization** | BQ, PQ, RQ, SQ supported |

### Index Parameters (HNSW)

| Parameter | Type | Default | Mutable | Description |
|-----------|------|---------|---------|-------------|
| `distance` | string | cosine | ❌ | Distance metric |
| `maxConnections` | int | 32 | ❌ | Max connections per layer |
| `efConstruction` | int | 128 | ❌ | Build-time search depth |
| `cleanupIntervalSeconds` | int | 300 | ✅ | Tombstone cleanup frequency |
| `ef` | int | -1 (dynamic) | ✅ | Query-time search depth |
| `vectorCacheMaxObjects` | int | 1e12 | ✅ | Memory cache limit |

---

## 3. Qdrant Schema Structure

### Collection Definition

```json
{
  "collection_name": "document_chunks",
  "vectors": {
    "size": 1536,
    "distance": "Cosine",      // Cosine, Dot, Euclid, Manhattan
    "multivector_config": null // Optional for multi-vector
  },
  "shard_number": 1,
  "replication_factor": 1,
  "write_consistency_factor": 1,
  "on_disk_payload": false,    // Store payload on disk (saves RAM)
  "optimizers_config": {
    "deleted_threshold": 0.2,
    "vacuum_min_vector_number": 1000,
    "max_segment_size": 100000,
    "memmap_threshold": 10000
  },
  "quantization_config": {     // Optional compression
    "scalar": {
      "type": "int8",
      "quantile": 0.99,
      "always_ram": true
    }
  }
}
```

### Point (Record) Format

```json
{
  "id": "document1#chunk1",     // UUID, string, or integer
  "vector": [0.023, -0.032, ...],
  "payload": {                  // Flexible JSON (nested OK!)
    "document_id": "document1",
    "chunk_number": 1,
    "chunk_text": "First chunk...",
    "metadata": {               // ✅ Nested objects supported
      "author": "John Doe",
      "department": "Engineering"
    },
    "tags": ["tutorial", "intro"],
    "created_at": "2024-01-15T00:00:00Z",
    "is_public": true
  }
}
```

### Key Characteristics

| Feature | Details |
|---------|---------|
| **Schema** | ❌ Schema-less (flexible payload) |
| **ID Format** | UUID, string, or uint64 |
| **Vector Types** | Dense, Sparse, Multi-vector |
| **Payload** | Flexible JSON (nested objects ✅) |
| **Payload Indexing** | Manual (create indexes on fields) |
| **Namespaces** | ❌ Use payload filtering instead |
| **Sharding** | Auto-sharding by hash or custom key |
| **Quantization** | Scalar (int8), Product, Binary |

### Payload Indexing

```json
// Create index on payload field
PUT /collections/document_chunks/index
{
  "field_name": "document_id",
  "field_schema": "keyword"  // keyword, integer, float, geo, text
}
```

---

## 4. Milvus Schema Structure

### Collection Definition

```python
from pymilvus import CollectionSchema, FieldSchema, DataType

# Define fields
id_field = FieldSchema(
    name="id",
    dtype=DataType.VARCHAR,
    max_length=128,
    is_primary=True
)

vector_field = FieldSchema(
    name="embedding",
    dtype=DataType.FLOAT_VECTOR,
    dim=1536
)

doc_id_field = FieldSchema(
    name="document_id",
    dtype=DataType.VARCHAR,
    max_length=256
)

chunk_num_field = FieldSchema(
    name="chunk_number",
    dtype=DataType.INT64
)

metadata_field = FieldSchema(
    name="metadata_json",
    dtype=DataType.JSON  // ✅ Full JSON support
)

# Create schema
schema = CollectionSchema(
    fields=[id_field, vector_field, doc_id_field, chunk_num_field, metadata_field],
    description="Document chunks for RAG",
    enable_dynamic_field=True  // Allow dynamic fields
)
```

### Entity (Record) Format

```json
{
  "id": "uuid-here",
  "embedding": [0.023, -0.032, ...],
  "document_id": "document1",
  "chunk_number": 1,
  "metadata_json": {            // Full JSON support
    "author": "John Doe",
    "tags": ["tutorial", "intro"],
    "nested": {
      "department": "Engineering"
    }
  },
  "created_at": "2024-01-15"
}
```

### Key Characteristics

| Feature | Details |
|---------|---------|
| **Schema** | ✅ Required (typed fields) |
| **Dynamic Fields** | ✅ Enabled via `enable_dynamic_field=true` |
| **ID Format** | INT64 or VARCHAR (primary key) |
| **Vector Types** | FLOAT_VECTOR, BINARY_VECTOR, FLOAT16_VECTOR, BFLOAT16_VECTOR |
| **JSON Support** | ✅ Full JSON data type |
| **Partitioning** | ✅ Manual partitions (by field value) |
| **Index Types** | FLAT, IVF_FLAT, IVF_SQ8, HNSW, DISKANN, SCANN |
| **Compression** | SCANN, DiskANN for disk-based search |

### Index Configuration

```python
index_params = {
    "metric_type": "COSINE",    // L2, IP, COSINE
    "index_type": "HNSW",
    "params": {
        "M": 32,                // Max connections
        "efConstruction": 128   // Build-time search depth
    }
}
collection.create_index(field_name="embedding", index_params=index_params)
```

---

## 5. Schema Mapping Matrix

### Field Type Mappings

| Pinecone | Weaviate | Qdrant | Milvus | Migration Notes |
|----------|----------|--------|--------|-----------------|
| `string` | `string` | `keyword` (indexed) | `VARCHAR` | ✅ Direct mapping |
| `number` (int) | `int` | `integer` | `INT64` | ✅ Direct mapping |
| `number` (float) | `number` | `float` | `FLOAT` | ⚠️ Pinecone converts all to float64 |
| `boolean` | `boolean` | `bool` (in payload) | `BOOL` | ✅ Direct mapping |
| `string[]` | `string[]` | Array in payload | `JSON` array | ⚠️ Qdrant/Milvus need JSON |
| N/A | `date` | ISO string in payload | `VARCHAR` (ISO) | ⚠️ Pinecone/Weaviate store as string |
| N/A | `text` | Text in payload | `VARCHAR` | ℹ️ No semantic difference |
| Metadata (flat JSON) | Properties (typed) | Payload (flexible JSON) | JSON + typed fields | 🔄 Requires transformation |

### Index Parameter Mappings

| Concept | Pinecone | Weaviate | Qdrant | Milvus |
|---------|----------|----------|--------|--------|
| Distance Metric | `cosine`, `dotproduct`, `euclidean` | `cosine`, `dot`, `l2-squared`, `hamming` | `Cosine`, `Dot`, `Euclid`, `Manhattan` | `COSINE`, `IP`, `L2` |
| Max Connections | N/A (managed) | `maxConnections` (32) | N/A (auto) | `M` (32) |
| Search Depth | N/A (managed) | `ef` (-1 dynamic) | `hnsw_ef` (128) | `ef` (64) |
| Build Quality | N/A (managed) | `efConstruction` (128) | N/A | `efConstruction` (128) |

---

## 6. Migration Challenges & Solutions

### Challenge 1: Schema Enforcement
- **Pinecone**: Schema-less (flat metadata)
- **Weaviate**: Strong schema (typed properties)
- **Solution**: Infer schema from sample data, create Weaviate classes dynamically

### Challenge 2: Nested Objects
- **Pinecone**: ❌ No nested objects
- **Qdrant/Weaviate**: ✅ Nested objects supported
- **Solution**: Flatten nested objects during Pinecone → Qdrant migration using dot notation (`author.name`)

### Challenge 3: Array Types
- **Pinecone**: `List[string]` only
- **Milvus**: Requires JSON for arrays
- **Solution**: Serialize arrays as JSON strings during migration

### Challenge 4: ID Formats
- **Pinecone**: String (custom format)
- **Weaviate**: UUID required
- **Solution**: Generate UUIDs from Pinecone IDs using SHA256 hash or keep as custom UUIDs

### Challenge 5: Namespaces vs Partitions
- **Pinecone**: Namespaces (logical grouping)
- **Milvus**: Partitions (physical separation)
- **Solution**: Map Pinecone namespaces to Milvus partitions or use payload filtering

---

## 7. Recommended Migration Strategy

### Phase 1: Schema Inference
1. Sample 10K records from source
2. Analyze metadata/payload structure
3. Infer field types and detect nested objects
4. Generate target schema definition

### Phase 2: Schema Creation
1. Create target collection/class with inferred schema
2. Configure index parameters (match source settings)
3. Create payload/property indexes for filterable fields

### Phase 3: Data Transformation
1. Transform IDs (if needed)
2. Flatten/unflatten nested structures
3. Convert field types (arrays → JSON, etc.)
4. Validate transformed records against target schema

### Phase 4: Validation
1. Sample 1K transformed records
2. Verify cosine similarity >0.98
3. Check metadata completeness
4. Test filtering queries

---

**Next Steps**: Implement schema inference engine and field type mapper in Go.

---

*Last updated: February 22, 2026*  
*Sources: Pinecone Docs, Weaviate Documentation, Qdrant Documentation, Milvus Documentation*
