# SkylerRedis - Project Development Summary

This document provides a detailed summary of the SkylerRedis project's development lifecycle, outlining the implementation process, challenges encountered, testing strategies, and instructions for setting up a replication topology.

---

## 1. Development Process Overview

The project was executed in five distinct phases, following the `skyler_redis_roadmap.md`.

#### Phase 1: Foundation & Planning
The project began by analyzing the roadmap to establish a clear project plan. A structured documentation directory was created with placeholders for future architectural documents, and a public-facing TODO list was set up to track progress.

#### Phase 2: Core Networking & Protocol
This crucial phase involved replacing the initial, basic "goroutine-per-connection" networking model with a high-performance, `epoll`-based event loop. This aligns with the architecture of modern servers like Redis. Platform-specific code was handled using Go build tags (`//go:build linux`), and a Docker-first workflow was adopted to ensure a consistent development and testing environment. The new layer was verified using `redis-cli` to confirm basic `PING`/`PONG` functionality.

#### Phase 3: Key-Value Engine & Data Structures
This phase focused on building the core data handling capabilities:
-   **Core KV & TTL**: Implemented `GET`/`SET` commands with support for time-to-live (TTL) expirations.
-   **Concurrency Safety**: A critical bug fix was introduced by adding mutex locks around all accesses to the central data store, preventing race conditions.
-   **Expiration Mechanisms**: A dual expiration strategy was implemented:
    1.  **Lazy Expiration**: Expired keys are deleted upon access (e.g., `GET`).
    2.  **Active Expiration**: A background task periodically samples and removes expired keys.
-   **Data Structures**: The server's capabilities were extended by implementing:
    1.  The `Set` data structure (`SADD`, `SREM`, `SCARD`, etc.).
    2.  Probabilistic Data Structures: **Bloom Filter** (`BFADD`, `BFEXISTS`) and a from-scratch implementation of a **Count-Min Sketch** (`CMSINCRBY`, `CMSQUERY`).
-   **Cache Eviction**: An approximate LRU (Least Recently Used) cache eviction policy was implemented, triggered when the number of keys exceeds a limit set by the `--maxmemory` flag.

#### Phase 4: Optimization & Concurrency (Sharding)
To prepare for multi-core scalability, the architecture was refactored from a single global data store to a sharded model:
-   **State Encapsulation**: All data and related state were encapsulated into a `Shard` struct.
-   **Sharded Command Handling**: All command handlers were modified to operate on a specific `Shard` instance.
-   **Key-based Routing**: A hash function (`fnv.New32a`) was implemented in `GetShardForKey` to route keys to their appropriate shard.
-   **Concurrency Fixes**: Concurrency bugs related to accessing the new `Shards` map were identified and fixed by introducing a dedicated `sync.RWMutex`.

#### Phase 5: Finalization & Evaluation
The final phase focused on documentation and performance analysis:
-   **Documentation**: All architectural documents were updated to reflect the new event loop, concurrency model, data structures, and eviction policies.
-   **Benchmarking**: Performance was measured using `redis-benchmark`. `SET` and `GET` workloads were tested against the server running with 1, 2, and 4 shards to evaluate scalability. The results were analyzed and documented, showing that the `SET` performance neared the 50,000 ops/sec target with 4 shards.

---

## 2. Challenges & Solutions

-   **Architectural Mismatch**: The initial networking layer was not scalable.
    -   **Solution**: A complete rewrite to an `epoll`-based event loop, establishing a high-performance foundation.
-   **Platform Specificity**: `epoll` is Linux-specific, causing build failures on other platforms.
    -   **Solution**: Used Go build tags (`//go:build linux`) and adopted a **Docker-first workflow** for consistent, cross-platform development.
-   **Dependency Issues (Count-Min Sketch)**: Several external libraries for Count-Min Sketch were either unavailable, flagged as a security risk, or had intractable build errors.
    -   **Solution**: Implemented the Count-Min Sketch algorithm from scratch in `internal/datastr`, which aligned perfectly with the project's low-level learning goals and eliminated external dependency problems.
-   **Concurrency Bugs**:
    1.  **Race Condition on Data Store**: The global data store was initially accessed without a mutex. **Solution**: A per-shard mutex (`shard.Mu`) was added to all data-modifying and data-accessing commands.
    2.  **Race Condition on Shards Map**: The `Shards` map itself was accessed concurrently. **Solution**: A dedicated `sync.RWMutex` (`shardsMu`) was introduced to protect the map.
-   **Nil Pointer Panic**: The server crashed on unknown commands (e.g., `CONFIG GET` from `redis-benchmark`).
    -   **Solution**: A `return` statement was added to the `default` case of the command dispatcher to prevent calling a method on a `nil` object.
-   **Incorrect LRU Eviction Logic**: Eviction was more aggressive than intended.
    -   **Solution**: Debugged by adding print statements to trace execution, analyzed the logic, and corrected the triggering condition in `EvictKeysByLRU`.

---

## 3. Testing & Verification Strategy

-   **Docker-based Workflow**: All development and testing were performed inside Docker containers to ensure a consistent and correct Linux environment.
-   **Unit & Functional Testing (`redis-cli`)**: New features were tested step-by-step using `redis-cli` (running in a separate Docker container) to verify return values and server state changes. This was crucial for debugging the LRU eviction logic.
-   **Performance Benchmarking (`redis-benchmark`)**:
    -   The standard `redis-benchmark` tool was used to measure performance (ops/sec).
    -   `SET` and `GET` workloads were tested with a high number of concurrent clients (`-c 200`).
    -   The server was run with 1, 2, and 4 shards (`--numshards`) to evaluate scalability.
-   **Debugging**: When tests failed or the server crashed, `docker logs` was used extensively to inspect server output, including custom debug prints added to trace program flow and state.

---

## 4. Setting Up Replication (Multiple Slaves)

You can run multiple SkylerRedis instances to form a master-slave replication topology. The following example uses Docker and maps different ports to the host machine.

#### Step 1: Start the Master

Run a SkylerRedis instance without the `--replicaof` flag. This instance defaults to a master.

```sh
# Run the master on port 6379
docker run -d -p 6379:6379 --name skyler-master skyler-redis
```

#### Step 2: Start the Slaves

Run additional SkylerRedis instances, using the `--replicaof` flag to point to the master. The key is to provide the master's address and map a different port for each slave.

**Note on `<master_ip>`**: When running Docker, containers can't simply use `localhost` or `127.0.0.1` to refer to the host machine. On Docker Desktop (Windows/Mac), you can often use the special DNS name `host.docker.internal`.

```sh
# Start the first slave, listening on port 6380, replicating from the master on 6379
docker run -d -p 6380:6380 --name skyler-slave1 skyler-redis --port 6380 --replicaof host.docker.internal 6379

# Start the second slave, listening on port 6381
docker run -d -p 6381:6381 --name skyler-slave2 skyler-redis --port 6381 --replicaof host.docker.internal 6379
```

#### Step 3: Verify Replication

1.  **Write to Master**: Connect to the master and set a key.
    ```sh
    redis-cli -p 6379 SET message "hello from master"
    ```
2.  **Read from Slaves**: Connect to each slave and get the key. The data should have been replicated.
    ```sh
    # Check slave 1
    redis-cli -p 6380 GET message
    # Expected output: "hello from master"

    # Check slave 2
    redis-cli -p 6381 GET message
    # Expected output: "hello from master"
    ```
This confirms that the replication link is active and commands are being propagated from the master to its slaves.
