# 15. Design Decisions: Key Expiry

## 1. Objective

Implement a mechanism to automatically delete keys that have been assigned a "Time-To-Live" (TTL), for example, via commands like `SET key value EX seconds` or `EXPIRE key seconds`. This mechanism needs to balance several factors:
-   **Correctness:** A key must be considered non-existent as soon as it expires.
-   **Memory Efficiency:** Expired keys must be cleaned up from memory to prevent waste (memory leaks).
-   **Performance:** The cleanup process must not significantly impact the server's overall performance.

## 2. Possible Solutions

### Solution 1: Passive / Lazy Expiration

-   **Structure:** The server does nothing proactively. Instead, it only checks when it needs to.
-   **Workflow:** When a command (e.g., `GET`, `TTL`) accesses a key, the server first checks if that key has expired.
    -   If it has expired, the server deletes the key immediately and behaves as if the key never existed (e.g., returning `(nil)` for a `GET`).
    -   If it has not expired, the command is executed as usual.
-   **Pros:** Extremely simple to implement. It costs no CPU time for keys that are never accessed again.
-   **Cons:** Keys that expire but are never accessed again will remain in memory forever. This leads to significant memory waste and is unacceptable in a production environment.

### Solution 2: Active Expiration

-   **Structure:** A background process (goroutine) runs periodically to clean up expired keys.
-   **Workflow:** At regular intervals (e.g., 10 times per second), this process performs a cleanup cycle. In each cycle, it:
    1.  Takes a random sample of keys with TTLs from the keyspace.
    2.  Checks and deletes any keys in the sample that have expired.
    3.  The loop terminates early if it exceeds a certain time limit (to avoid monopolizing the CPU) or if the percentage of cleared keys is low (indicating fewer expired keys need cleaning).
-   **Pros:** Proactively frees memory, significantly mitigating memory waste from untouched keys.
-   **Cons:** It consumes some CPU resources for the background scanning. There is a certain latency—a key might be expired but not yet deleted until the next cleanup cycle.

### Solution 3: Hybrid Approach (The Redis Way)

-   **Structure:** Combines both of the above methods to leverage their pros and eliminate their cons.
-   **Workflow:**
    1.  **Lazy Expiration:** Whenever a key is accessed, the server *always* checks its expiration time first (as in Solution 1). This guarantees that no client can ever read stale/expired data.
    2.  **Active Expiration:** Concurrently, a background process runs periodically to scan and delete expired keys (as in Solution 2). This reclaims the memory used by millions of keys that are no longer accessed.
-   **Pros:** This is the most comprehensive and robust solution. It both ensures the data returned to the client is always correct and proactively cleans up wasted memory.
-   **Cons:** It is more complex to implement than choosing just one of the two methods.

## 3. Trade-off Analysis

| Solution | Correctness (Client View) | Memory Efficiency | CPU Impact | Implementation Complexity |
| :--- | :--- | :--- | :--- | :--- |
| 1. Lazy | **High** | **Very Poor** (Causes memory leaks) | **Very Low** | **Low** |
| 2. Active | **Low** (Can read stale keys) | **Good** | **Low-Medium** | **Medium** |
| 3. Hybrid | **High** | **Excellent** | **Low-Medium** | **High** |

## 4. SkylerRedis's Choice and Rationale

**Choice:** **Solution 3: Hybrid Approach.**

**Rationale:**

1.  **Redis Compatibility and Effectiveness:** This is precisely the strategy that Redis uses. To build a proper Redis clone, implementing both mechanisms is mandatory. The "Lazy Expiry" test cases confirm the requirement for Solution 1, and the existence of `ActiveExpiration.md` implies the need for Solution 2.

2.  **Preventing Memory Leaks:** Using only Lazy Expiration (Solution 1) is insufficient. It would lead to expired keys permanently occupying memory if they are never requested again. The Active Expiration mechanism is **required** to solve this problem, ensuring the server can run stably over long periods.

3.  **Ensuring Data Correctness:** Using only Active Expiration (Solution 2) is also insufficient. Due to the delay between when a key expires and when the background process deletes it, a client could potentially read stale data during that window. Lazy Expiration closes this loophole by checking the TTL at the exact moment of access.

**Conclusion:** The hybrid approach is the only solution that properly solves the two core problems of key expiration: **data correctness for the client** and **effective memory management for the server**.
