# SkylerRedis Documentation

## 8. Sharding and Data Distribution

SkylerRedis implements a sharding mechanism to distribute its dataset across multiple independent storage units, enabling better utilization of multi-core processors and enhancing overall scalability. This document details how data is sharded and managed within the system.

### The Shard Concept

Instead of a single, monolithic global data store, SkylerRedis partitions its data into several `Shard` instances. Each `Shard` is a self-contained unit that holds its own set of data structures (Key-Value store, Sets, Bloom Filters, Count-Min Sketches, Lists, Streams), along with its own LRU clock, memory limit, and mutex for internal synchronization.

The `Shard` struct is defined in `internal/memory/shard.go`:

```go
// internal/memory/shard.go
type Shard struct {
	Store            map[string]Entry
	Sets             map[string]map[string]struct{}
	BloomFilters     map[string]*bloom.BloomFilter
	CountMinSketches map[string]*Sketch
	RPush            map[string][]string
	Stream           map[string]StreamEntry
	StreamIDs        []string
	LruClock         uint64
	MaxMemory        int
	Mu               sync.Mutex // Mutex for this specific shard's data
}
```

### Global Shards Management

The `internal/memory/main.go` package is responsible for initializing and managing these `Shard` instances globally:

-   **`numShards`**: Configured via the `--numshards` command-line flag, this determines how many `Shard` instances the server will create.
-   **`memory.Shards`**: A global map (`map[int]*Shard`) that holds references to all created `Shard` instances.
-   **`memory.InitShards(numShards, maxMemory)`**: Called during server startup (`app/main.go`), this function creates `numShards` instances of `Shard` and populates the `memory.Shards` map. Each shard is initialized with the configured `maxMemory` limit.
-   **`shardsMu sync.RWMutex`**: A dedicated `RWMutex` (`shardsMu`) protects access to the `memory.Shards` map itself, ensuring thread-safe read and write operations on the map of shards.

### Key-based Routing

When a client command operates on a key, SkylerRedis determines which specific `Shard` instance is responsible for that key. This process is called key-based routing.

#### `memory.GetShardForKey(key string) *Shard`

This function, located in `internal/memory/main.go`, is the central component for key routing:

1.  **Hashing**: It uses the FNV-1a (Fowler-Noll-Vo) hash algorithm (`hash/fnv`) to compute a hash value for the input `key`.
    ```go
    h := fnv.New32a()
    h.Write([]byte(key))
    hashValue := h.Sum32()
    ```
2.  **Shard ID Calculation**: The `shardID` is determined by taking the modulo of the hash value by the total number of configured shards:
    ```go
    shardID := int(hashValue % uint32(numShards))
    ```
    This ensures that the key is consistently mapped to the same shard across all operations.
3.  **Shard Retrieval**: The function then returns the `Shard` instance corresponding to the calculated `shardID` from the `memory.Shards` map.

### Benefits of Sharding

-   **Multi-core Scalability**: By distributing the data across multiple `Shard` instances, SkylerRedis can process commands targeting different keys (that hash to different shards) in parallel. Each shard operates largely independently, reducing contention for resources and allowing the system to leverage multiple CPU cores.
-   **Reduced Lock Contention**: Each `Shard` has its own `Mu` (mutex). Commands only need to acquire the lock for the specific shard they are operating on, rather than a global lock for the entire dataset. This significantly improves concurrency, especially under high write loads.
-   **Isolation**: Failures or performance bottlenecks in one shard are less likely to affect other shards, contributing to a more resilient system.
-   **Memory Management**: Each shard manages its own memory limit (`MaxMemory`) and LRU eviction policy independently.

### Usage Example

To start SkylerRedis with sharding enabled, use the `--numshards` flag:

```sh
# Start with 4 shards, each with a maxmemory of 1000 keys
docker run -d -p 6379:6379 --name skyler-redis-server skyler-redis --numshards 4 --maxmemory 1000
```

When you perform a `SET` or `GET` operation, `memory.GetShardForKey` will determine which of the 4 shards is responsible for that key.
