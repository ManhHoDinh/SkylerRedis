# SkylerRedis Documentation

## 5. Concurrency Model: The Thread-per-Shard Architecture

SkylerRedis is designed for high throughput, capable of handling tens of thousands of simultaneous connections. To achieve this, it abandons the simple but limited "goroutine-per-connection" model in favor of a highly efficient **thread-per-shard architecture**, powered by Linux's `epoll` I/O event notification facility. This model is inspired by Redis Cluster's approach to scalability.

### The Problem with "Goroutine-per-Connection"

The most straightforward way to write a network server in Go is to spawn a new goroutine for each incoming connection:

```go
// The old, naive model
for {
    conn, _ := listener.Accept()
    go handleConnection(conn) // New goroutine for every client
}
```

While this works for a small number of clients, it suffers from significant overhead at scale:
- **High Memory Usage:** Each goroutine consumes at least 2KB of stack space. 10,000 connections would mean at least 20MB of memory just for stacks.
- **Scheduler Overhead:** The Go runtime scheduler must manage a large number of goroutines. As the number of connections grows, the cost of scheduling and context-switching between them becomes a major performance bottleneck.

### The Thread-per-Shard Architecture

SkylerRedis implements a shared-nothing, thread-per-shard architecture. Instead of a single global data store, the entire dataset is partitioned across multiple independent `Shard` instances. Each `Shard` is then managed by its own dedicated event loop.

The core components of this architecture are:

#### 1. Shard Encapsulation

-   All data structures (`Store`, `Sets`, `BloomFilters`, `CountMinSketches`, `RPush`, `Stream`), their associated state (`LruClock`, `MaxMemory`), and their mutex (`Mu`) are encapsulated within a `memory.Shard` struct. This ensures that each shard is entirely self-contained and operates independently.
-   The `memory` package now maintains a map of these `Shard` instances (`memory.Shards`).

#### 2. Key Hashing and Routing

-   When a client command arrives, the key is extracted (typically `args[1]`).
-   A hashing function (`fnv.New32a`) is applied to the key.
-   The hash value is then used to determine the target shard: `shardID := int(h.Sum32() % uint32(numShards))`.
-   The `memory.GetShardForKey(key)` function retrieves the correct `Shard` instance, ensuring that all operations on a specific key consistently target the same shard.

#### 3. Multiple Event Loops

-   Instead of a single global event loop, `app/main.go` now initializes multiple `Shard` instances (configured via `--numshards`).
-   Each `Shard` runs its own dedicated event loop for background tasks (expired key eviction, LRU eviction) in a separate goroutine.
-   For client-facing operations, a single `EventLoop` from `internal/eventloop` is still used to accept connections and poll for I/O events. However, when an event occurs, the command is immediately routed to the correct `Shard` based on its key, and executed synchronously within that shard's context.

#### 4. Concurrency Model

-   The server maintains a global `EventLoop` (from `internal/eventloop`) which monitors all client connections using `epoll`.
-   When a client sends a command, `EventLoop`'s `ReadCallback` extracts the key, uses `memory.GetShardForKey()` to find the relevant `Shard`.
-   The command is then dispatched to `command.HandleCommand`, which executes it on the specific `Shard` instance.
-   Crucially, commands that operate on shard-specific data acquire and release the mutex (`shard.Mu`) belonging to that particular shard. This ensures thread safety *within* the shard.

### Benefits of the Thread-per-Shard Model

-   **True Multi-core Scalability**: By partitioning the data and operating on separate `Shard` instances with their own mutexes, SkylerRedis can effectively utilize multiple CPU cores. Operations on keys belonging to different shards can proceed in parallel without contention over global locks.
-   **Enhanced Concurrency**: Each shard manages its own data independently, reducing the scope of locking. This significantly increases the number of concurrent operations the server can handle compared to a single global lock.
-   **Lower Latency**: Reduced lock contention means commands are processed more quickly, leading to lower and more consistent latency.
-   **Massive Scalability**: The server can handle an even larger number of connections and a larger dataset, as resources are distributed and managed across multiple isolated units.
-   **Improved Maintainability**: Encapsulating data and logic within `Shard`s makes the codebase more modular and easier to reason about.

This architecture provides a robust foundation for building a high-performance, scalable Redis-compatible database.
