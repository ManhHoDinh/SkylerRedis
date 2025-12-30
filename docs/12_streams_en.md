# 12. Design Decisions: Redis Streams

## 1. Objective

Implement the Redis Stream data structure, an append-only log. This requires several core features:
-   `XADD`: To add a new entry to a stream.
-   `XRANGE`: To query a range of entries based on their IDs.
-   `XREAD`: To read entries from one or more streams, starting from a specific ID, with the optional capability to `BLOCK` and wait for new data.

The `ms-seq` (millisecond time - sequence number) ID format is a critical design component that heavily influences the implementation.

## 2. Possible Solutions

### Solution 1: List-based (Dynamic Array)

-   **Structure:** A Stream is treated as a simple list or dynamic array of entries. Each entry is a struct/object containing its ID and data.
-   **`XADD`:** Appends a new entry to the end of the array. Generating an automatic ID (`*`) requires getting the current time and iterating backward to find the last sequence number for the same millisecond, or defaulting to 0 if the millisecond has changed.
-   **`XRANGE`/`XREAD`:** Requires a full linear scan of the array to filter for entries whose IDs fall within the requested range. This is an **O(N)** operation, where N is the total number of entries, making it very slow.
-   **Assessment:** Very simple to implement initially but does not scale for streams with a large volume of data.

### Solution 2: Self-Balancing Binary Search Tree (B-Tree/Red-Black Tree) or Skip List

-   **Structure:** Stores stream entries in a sorted data structure like a B-Tree or Skip List, using the entry ID as the key for comparison and ordering.
-   **`XADD`:** Adds a new entry to the tree/skip list, which is an **O(log N)** operation. Automatic ID generation is also efficient because the last (largest) entry can be found in O(log N) or O(1) time.
-   **`XRANGE`/`XREAD`:** Range queries on these structures are highly efficient, typically with a complexity of **O(log N + M)**, where M is the number of entries returned.
-   **Assessment:** Significantly more complex to implement than a list, but it provides superior, scalable performance. This is a well-balanced choice.

### Solution 3: Radix Tree (The Redis Way)

-   **Structure:** Redis itself uses a highly optimized data structure called a Radix Tree. Each node in the tree represents a portion of the ID (e.g., the 8-byte timestamp, the 8-byte sequence number). Entries are stored at the leaf nodes.
-   **Advantages:** This structure is extremely memory-efficient because common prefixes of entry IDs are shared among nodes. Insertion and range scan operations are also exceptionally fast.
-   **`XREAD BLOCK`:** The blocking mechanism can be implemented by maintaining a list of waiting clients for each stream. When a new `XADD` command arrives, the server iterates through this list, checks if the new entry satisfies the conditions of any waiting client, and "wakes up" the relevant clients.
-   **Assessment:** This is the most complex approach, requiring deep knowledge of data structures. However, it delivers the best performance and memory efficiency, perfectly matching native Redis.

## 3. Trade-off Analysis

| Solution | Performance (XADD/XRANGE) | Memory Efficiency | Implementation Complexity | Redis Compatibility |
| :--- | :--- | :--- | :--- | :--- |
| 1. List/Array | **Very Poor** (O(N) for searches) | **Moderate** | **Very Low** | **Low** (Fails to scale) |
| 2. B-Tree/Skip List | **Good** (O(log N)) | **Good** | **High** | **Medium** (Correct behavior, but different internal structure) |
| 3. Radix Tree | **Excellent** (Optimized for IDs) | **Excellent** (Shared prefixes) | **Very High** | **High** (This is how Redis does it) |

## 4. SkylerRedis's Choice and Rationale

**Assumed Choice:** **Solution 2: B-Tree or Skip List.**

**Rationale:**

1.  **Balance of Performance and Complexity:** While the Radix Tree (Solution 3) is the authentic data structure used by Redis, implementing it from scratch is an extremely complex and time-consuming task. It may be outside the scope of a learning-oriented project like SkylerRedis.
2.  **Meets Performance Requirements:** A B-Tree or Skip List already provides **O(log N)** performance for core operations. This is fast enough to pass all performance-related test cases and, more importantly, is far more scalable than the naive List-based solution (Solution 1).
3.  **Focus on Business Logic:** This choice allows the developer to concentrate on correctly implementing the high-level *logic* of Streams (ID generation, blocking, consumer groups) without getting bogged down in low-level data structure optimization. Existing B-Tree libraries can be leveraged.
4.  **Blocking Implementation:** The blocking mechanism for `XREAD BLOCK` can be built relatively independently of the underlying data structure. A map like `map[streamName][]blockingClient` to track waiting clients is a sufficient and effective starting point.
