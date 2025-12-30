# 11. Design Decisions: Transactions (MULTI / EXEC)

## 1. Objective

Implement atomic transactions in SkylerRedis, allowing a client to group multiple commands to be executed as a single, indivisible operation. The key commands are `MULTI`, `EXEC`, and `DISCARD`, with the goal of maintaining full compatibility with Redis.

## 2. Possible Solutions

When implementing a transaction system, there are three primary approaches, each with distinct trade-offs:

### Solution 1: Client-side Command Queueing

-   **Workflow:** The client itself manages a queue of commands after the user calls `MULTI`. Upon receiving an `EXEC` command, the client bundles all queued commands and sends them to the server in a single network request (often called "batching" or "pipelining").
-   **Server-side:** The server receives and executes a continuous stream of commands without any intrinsic knowledge that they belong to a transaction.
-   **Implementation:** The transaction logic resides entirely within the client library (e.g., `redis-py`, `node-redis`). The server requires minimal to no changes.

### Solution 2: Server-side State & Command Queueing

-   **Workflow:**
    1.  When a client sends `MULTI`, the server flags that specific connection as being in a "transaction state."
    2.  Subsequent commands from this client are not executed immediately. Instead, they are appended to a private queue associated with that connection. For each command, the server replies with `+QUEUED`.
    3.  When the client sends `EXEC`, the server iterates through the command queue and executes the commands sequentially and atomically. During this execution, no commands from other clients are allowed to interleave.
    4.  If the client sends `DISCARD`, the server simply clears the queue and exits the transaction state.
-   **Implementation:** This is the standard Redis approach. The server must manage state and queues on a per-connection basis.

### Solution 3: Global Server Lock

-   **Workflow:** When a client sends `MULTI`, the server activates a global lock, preventing all other clients from executing any commands. The client in the transaction sends its commands one by one, and the server executes them immediately. Upon receiving `EXEC`, the server releases the lock.
-   **Implementation:** This is conceptually simple but has severe performance implications as it completely blocks all other clients.

## 3. Trade-off Analysis

| Solution | Performance | Atomicity Guarantee | Complexity | Memory Usage (Server) |
| :--- | :--- | :--- | :--- | :--- |
| **1. Client Queue** | **High** (Fewer network round-trips) | **Weak** (Not guaranteed by the server; another client could modify keys between commands) | **Medium** (Shifts complexity to all client libraries) | **Low** (Server remains stateless) |
| **2. Server Queue** | **Good** (Fast and controlled execution) | **Strong** (The server 100% guarantees atomicity during `EXEC`) | **Medium** (Server must manage per-connection state/queues) | **Medium** (Stores a command queue for each active transaction) |
| **3. Global Lock** | **Very Poor** (Kills concurrency) | **Strong** | **Low** (Simple lock mechanism) | **Low** |

## 4. SkylerRedis's Choice and Rationale

**Choice:** **Solution 2: Server-side State & Command Queueing.**

**Rationale:**

1.  **Redis Compatibility:** This is how Redis operates. The primary goal of SkylerRedis is to be a compatible clone of Redis, so adhering to its exact semantics and behavior is a core requirement. The project's test cases, which expect a `+QUEUED` reply, explicitly confirm that this behavior is mandatory.

2.  **Guaranteed Atomicity:** This approach provides a true "all-or-nothing" guarantee at the server level, which is the fundamental promise of a transaction. Solution 1 cannot provide this guarantee, as another client could interleave commands and alter data in the middle of the intended transaction.

3.  **Balanced Performance:** While a Global Lock (Solution 3) is easier to implement, its performance is unacceptable for a server designed to handle concurrent connections. The server-side queueing model provides atomicity for the *execution phase* (`EXEC`) without blocking unrelated clients during the *queuing phase*. This strikes the necessary and correct balance between data safety and concurrency, keeping the uninterruptible critical section as short as possible.
