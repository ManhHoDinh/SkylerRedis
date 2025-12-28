# SkylerRedis — Memory Management

This document provides a summary of SkylerRedis's memory management strategies. The primary goal is to maintain a predictable and efficient memory footprint while handling a large number of keys and connections.

---

### 1. Approximate LRU Eviction

SkylerRedis implements an approximate Least Recently Used (LRU) eviction policy to enforce a memory limit.

-   **Configuration**: The memory limit is set via the `--maxmemory <N>` flag, where `N` is the maximum number of keys per shard.
-   **Mechanism**:
    -   Each key-value `Entry` has an `LRU` field, which stores a "last access time" based on a global, incrementing `LruClock`.
    -   When the number of keys in a shard exceeds `MaxMemory`, the eviction algorithm is triggered.
    -   It randomly samples a small number of keys, identifies the key with the lowest `LRU` value (the least recently used), and deletes it.
    -   This process repeats until the key count is below a target threshold (95% of `MaxMemory`).
-   **Reference**: For more details, see [`07_cache_eviction_en.md`](./07_cache_eviction_en.md).

---

### 2. Active Key Expiration (TTL)

To prevent expired keys from consuming memory indefinitely, SkylerRedis uses a dual expiration strategy.

-   **Passive Expiration**: When an expired key is accessed, it is deleted on the spot.
-   **Active Expiration**: A background task runs periodically (every 100ms per shard) to actively find and delete expired keys.
    -   This task randomly samples a small subset of keys, checks their `ExpiryTime`, and removes them if they are expired.
-   **Reference**: For more details, see [`04_kv_engine_en.md`](./04_kv_engine_en.md).

---

### 3. Data Structure Efficiency

-   **Sets**: The `Set` data structure is implemented using `map[string]struct{}`, where the `struct{}` value consumes zero memory. This is a highly memory-efficient way to store sets of unique strings in Go.
-   **Reference**: For more details, see [`10_data_structures_en.md`](./10_data_structures_en.md).

---

### 4. Low-Overhead Concurrency Model

-   **Event Loop**: By using a single event loop to manage thousands of connections, SkylerRedis avoids the high memory overhead of a "goroutine-per-connection" model. A connection is simply a file descriptor, not a 2KB+ goroutine stack.
-   **Sharding**: The `thread-per-shard` architecture partitions the dataset, meaning each shard manages a smaller, more predictable amount of memory. This also reduces lock contention, which can indirectly affect memory performance.
-   **Reference**: For more details, see [`05_concurrency_model_en.md`](./05_concurrency_model_en.md) and [`08_sharding_en.md`](./08_sharding_en.md).