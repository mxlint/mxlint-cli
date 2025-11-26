# Cache Architecture

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    EvalAllWithResults                       │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │             For each Rule in Rules                   │  │
│  │                                                       │  │
│  │  ┌───────────────────────────────────────────────┐   │  │
│  │  │       For each Input File matching Pattern    │   │  │
│  │  │                                                │   │  │
│  │  │  1. Generate Cache Key                        │   │  │
│  │  │     ├─ SHA256(rule content)                   │   │  │
│  │  │     └─ SHA256(input content)                  │   │  │
│  │  │                                                │   │  │
│  │  │  2. Check Cache                               │   │  │
│  │  │     ├─ Hit? → Use cached result               │   │  │
│  │  │     └─ Miss? → Evaluate + cache result        │   │  │
│  │  │                                                │   │  │
│  │  └───────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

Cache Storage:
~/.cache/mxlint/
├── {rule1-hash}-{input1-hash}.json
├── {rule1-hash}-{input2-hash}.json
├── {rule2-hash}-{input1-hash}.json
└── ...
```

## Cache Decision Flow

```
┌─────────────────────────────────────┐
│  Start: Evaluate Rule on Input     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Generate Cache Key                 │
│  - SHA256(rule content)             │
│  - SHA256(input content)            │
└──────────────┬──────────────────────┘
               │
               ▼
         ┌─────────────┐
         │ Cache Exists?│
         └─────┬───┬────┘
               │   │
         Yes   │   │   No
               │   │
       ┌───────┘   └───────┐
       │                   │
       ▼                   ▼
┌──────────────┐    ┌──────────────┐
│ Cache Hit    │    │ Cache Miss   │
│ Load Result  │    │ Evaluate Rule│
└──────┬───────┘    └──────┬───────┘
       │                   │
       │                   ▼
       │            ┌──────────────┐
       │            │ Save to Cache│
       │            └──────┬───────┘
       │                   │
       └───────┬───────────┘
               │
               ▼
    ┌──────────────────────┐
    │  Return Result       │
    └──────────────────────┘
```

## Performance Comparison

### Scenario 1: First Run (Cold Cache)
```
Time: 100% (baseline)
Cache: 0 hits, N misses
Result: All rules evaluated
```

### Scenario 2: Second Run (Warm Cache, No Changes)
```
Time: ~5-10% (90-95% faster)
Cache: N hits, 0 misses
Result: All results from cache
```

### Scenario 3: Incremental Changes (1 file modified)
```
Time: ~15-20% (80-85% faster)
Cache: N-1 hits, 1 miss
Result: 1 rule re-evaluated, rest from cache
```

### Scenario 4: Rule Modified (All inputs need re-evaluation)
```
Time: ~50-60% (if half the rules changed)
Cache: M hits, N misses (where M = unchanged rules × inputs)
Result: Changed rule re-evaluated against all inputs
```

## Cache Key Generation

```go
// Pseudo-code
func createCacheKey(rulePath, inputPath) CacheKey {
    ruleContent := readFile(rulePath)
    inputContent := readFile(inputPath)
    
    return CacheKey{
        RuleHash:  SHA256(ruleContent),
        InputHash: SHA256(inputContent),
    }
}
```

## Cache File Structure

Each cache file (`~/.cache/mxlint/{ruleHash}-{inputHash}.json`):

```json
{
  "version": "v1",
  "cache_key": {
    "rule_hash": "abc123def456...",
    "input_hash": "789ghi012jkl..."
  },
  "testcase": {
    "name": "path/to/input.yaml",
    "time": 0.123,
    "failure": null,
    "skipped": null
  }
}
```

## Cache Invalidation Strategy

### Automatic Invalidation
Cache is automatically invalidated when:
- Rule file content changes (different SHA256 hash)
- Input file content changes (different SHA256 hash)

### Manual Invalidation
Users can manually clear cache:
```bash
mxlint cache-clear
```

### Version-based Invalidation
Cache entries with different version numbers are ignored:
- Current version: `v1`
- Future versions will invalidate old cache entries

## Cache Management Commands

### View Statistics
```bash
mxlint cache-stats

Output:
  Cache Statistics:
    Entries: 150
    Total Size: 2.3 MB
```

### Clear Cache
```bash
mxlint cache-clear

Output:
  Cache cleared: ~/.cache/mxlint
```

## Error Handling

```
Cache Error → Log debug message → Continue with normal evaluation
```

Cache errors never fail the lint operation:
- Read error → Falls back to normal evaluation
- Write error → Logs warning, continues
- Invalid cache → Ignores entry, evaluates normally

## Concurrency Safety

The caching implementation is thread-safe:
- Each goroutine handles its own cache operations
- File system operations are atomic at the OS level
- No shared state between goroutines
- Cache misses may result in duplicate evaluations (acceptable)

## Cache Location

```
Default: ~/.cache/mxlint/

Platform-specific:
- Linux:   ~/.cache/mxlint/
- macOS:   ~/.cache/mxlint/
- Windows: %USERPROFILE%/.cache/mxlint/
```

## Scalability Considerations

### Current Implementation
- One file per cache entry
- No size limits
- No expiration policy
- Simple file-based storage

### Future Enhancements (if needed)
- Maximum cache size limit
- LRU eviction policy
- Time-based expiration
- Cache compression
- Database backend for large caches

