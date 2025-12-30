# 14. Design Decisions: Replication

## 1. Objective

Implement a master-slave replication mechanism to achieve two primary goals:
1.  **Read Scaling:** Distribute read queries across multiple slave instances to reduce the load on the master.
2.  **Data Redundancy:** Create copies of the data to serve as a backup in case the master instance fails.

The mechanism must handle two distinct phases: **initial synchronization** when a new slave connects, and **ongoing changes** to keep the slave up-to-date with the master.

## 2. Possible Solutions

### Solution 1: Statement-based Replication (The Redis Way)

-   **Structure:** This is the method employed by Redis. Instead of sending the *changed data*, the master sends the actual *write command* that caused the change. For example, when the master executes `INCR mykey`, it sends the literal command `INCR mykey` to all its slaves.
-   **Initial Synchronization:** The master creates a point-in-time snapshot of its data (as an RDB file) and sends it to the slave. The slave loads this snapshot to get the initial state. Afterward, the master begins streaming all write commands that occurred during and after the snapshot process.
-   **Pros:** Highly efficient in terms of network bandwidth, as commands are often much smaller than the data they modify (e.g., `SADD myset "a-very-long-member"`). The logic is simple, easy to understand, and to debug.
-   **Cons:** It can potentially lead to inconsistencies if a command is non-deterministic (e.g., `SADD myset (random-member)`). However, Redis is designed such that its commands are safe for replication, and non-deterministic commands often have alternative solutions.

### Solution 2: Row-based Replication

-   **Structure:** Common in relational database systems. Instead of sending the command, the master sends the *final effect* of the change on the data. For instance, after `INCR mykey` (assuming the value changes from 10 to 11), the master would send a message equivalent to "key `mykey` now has the value `11`".
-   **Pros:** Completely immune to issues from non-deterministic commands. It always ensures that the final data on the slaves is consistent in value.
-   **Cons:** It can generate significantly more network traffic. For example, an `LREM` command that removes 1000 elements from a list is a single short command, but with row-based replication, it might need to send information about all 1000 deleted elements, causing a massive spike in network cost.

### Solution 3: Write-Ahead Log (WAL) Shipping

-   **Structure:** The master writes all changes to a log file (the Write-Ahead Log) before applying them to memory. The slaves then receive a copy of this log file and "replay" the changes in the exact same order.
-   **Pros:** Extremely robust and durable, forming the foundation of large databases like PostgreSQL. It enables advanced techniques like point-in-time recovery.
-   **Cons:** Very complex to implement for an in-memory database like Redis. It increases the latency of every write operation because it must first be committed to a log on disk. This is not how Redis operates.

## 3. Trade-off Analysis

| Solution | Network Efficiency | Handling Non-determinism | Implementation Complexity | Redis Compatibility |
| :--- | :--- | :--- | :--- | :--- |
| 1. Statement-based | **Excellent** | **Weak** (but an acceptable issue in Redis) | **Medium** | **Full** |
| 2. Row-based | **Moderate** | **Strong** | **High** | **None** |
| 3. WAL Shipping | **Good** | **Strong** | **Very High** | **None** |

## 4. SkylerRedis's Choice and Rationale

**Choice:** **Solution 1: Statement-based Replication.**

**Rationale:**

1.  **Redis Compatibility:** This is precisely how Redis handles replication. The entire handshake process (`PSYNC ? -1`), the RDB transfer, and the subsequent command streaming are the foundation of Redis replication. The project's test cases confirm and require this exact workflow. The existing `09_replication_en.md` file details the steps of this method, and this choice aligns with that established design.
2.  **Optimization for In-Memory Databases:** In Redis, write operations are typically very fast, and the commands themselves are concise. Sending these short commands over the network is the most natural and efficient approach, avoiding the overhead of parsing and sending complex data changes.
3.  **Balance of Simplicity and Efficiency:** While it has a theoretical weakness with non-deterministic commands, in practice, Redis's write commands are designed to be replication-safe. This approach provides the best balance of performance, simplicity, and consistency guarantees within the context of Redis.
