# SkylerRedis vs. Production Redis

This document provides a high-level comparison between **SkylerRedis** (this project) and a **production-grade Redis** instance. The goal is to clarify the scope and purpose of SkylerRedis as a learning project.

---

| Feature / Aspect          | Production Redis                                    | SkylerRedis                                                                                                | Notes                                                                                                    |
| ------------------------- | --------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| **Project Goal**          | High-performance, stable, feature-rich data store   | A learning project to understand low-level systems design and Redis's internal architecture.             | **SkylerRedis is NOT for production use.**                                                               |
| **Core Architecture**     | Single-threaded Event Loop (per shard in Cluster)   | Single-threaded Event Loop (per shard, with a global dispatcher).                                          | SkylerRedis successfully implements the core architectural pattern.                                      |
| **Sharding**              | Redis Cluster (complex, with hash slots)            | Simple key-based sharding using FNV-1a hash (`hash(key) % numShards`).                                     | SkylerRedis demonstrates the sharding concept but lacks the robustness and tooling of Redis Cluster.   |
| **Data Structures**       | - Strings<br>- Lists<br>- Sets<br>- Hashes<br>- Sorted Sets<br>- Streams<br>- HyperLogLog<br>- Bitmaps<br>- Geospatial<br>- Modules (Bloom, etc.) | - Strings<br>- Lists<br>- Sets<br>- Streams (partial)<br>- **Bloom Filter** (native)<br>- **Count-Min Sketch** (native) | SkylerRedis implements a subset of the most important data structures. Hashes and Sorted Sets are missing. |
| **Persistence**           | RDB (snapshots) & AOF (append-only file)            | **None.** Data is entirely in-memory and lost on restart.                                                  | Persistence is a major feature that was out of scope for this project.                                   |
| **Replication**           | Robust Master-Slave with full/partial resync        | Basic Master-Slave with command propagation. Lacks robust error handling and partial resynchronization.  | SkylerRedis demonstrates the core concepts of replication but is not fault-tolerant.                   |
| **Concurrency**           | Highly optimized, non-blocking I/O                  | Non-blocking I/O via `epoll`. Per-shard mutexes for data access.                                           | The concurrency model is a key strength of SkylerRedis, showing good scalability for `SET` operations.   |
| **Eviction & Expiration** | Configurable policies (LRU, LFU, etc.), TTL         | Approximate LRU and TTL with active/passive expiration.                                                    | SkylerRedis successfully implements the core Redis eviction and expiration strategies.                   |
| **Transactions**          | `MULTI`/`EXEC`                                      | Basic `MULTI`/`EXEC` queueing (without atomicity guarantees across shards).                              | The current transaction implementation is not truly atomic in a multi-shard environment.               |
| **High Availability**     | Redis Sentinel, Redis Cluster                       | **None.**                                                                                                  | High availability features are complex and out of scope.                                                 |

---

### Conclusion

SkylerRedis successfully achieves its goal as a **deep-dive learning project**. It rebuilds many of Redis's most important and complex internal features from scratch, including:
- A high-performance event loop.
- Key-based sharding for multi-core scalability.
- Probabilistic data structures.
- Approximate LRU eviction.

While it lacks the completeness, robustness, and persistence of a production Redis instance, it serves as an excellent case study in high-performance systems design and provides a solid foundation for understanding why Redis is so fast and efficient.