# Step 2: I/O Multiplexing with epoll/kqueue

This document explains the "why" and "how" of implementing an I/O multiplexing-based event loop, which is the core of our high-performance server.

## 1. The Problem with the Simple `Accept()` Loop

In our basic TCP server, the `ln.Accept()` call is **blocking**. The server is stuck at this line until a new client connects. Similarly, reading from a client connection (`conn.Read()`) is also a blocking operation.

A common but inefficient way to handle multiple clients is the **"thread-per-connection"** model (or in Go, **"goroutine-per-connection"**).

```go
// Inefficient model
for {
    conn, _ := ln.Accept()
    go handleConnection(conn) // New goroutine for every client
}
```

### Tradeoffs of Goroutine-per-Connection:

*   **Pros:**
    *   Very simple to write and understand.
*   **Cons:**
    *   **High Memory Usage:** Each goroutine consumes at least 2KB of stack space. 10,000 clients would mean at least 20MB of stack memory, plus the overhead of the Go runtime managing them.
    *   **Scheduler Overhead:** The Go scheduler has to manage a large number of goroutines. While it's highly efficient, it's not free. For a high-performance database, we want more direct control.
    *   **No Fine-grained Control:** It's difficult to implement features like backpressure or prioritize certain operations when the scheduler is in full control.

## 2. The Solution: I/O Multiplexing

I/O Multiplexing allows a single thread to monitor multiple I/O operations (like accepting new connections or reading from existing ones) simultaneously.

Instead of asking the kernel "Is this specific connection ready for reading?" (blocking), we ask "Are *any* of these connections I'm interested in ready for an I/O operation?" (can be non-blocking).

This is achieved through system calls like `epoll` (on Linux) and `kqueue` (on macOS/BSD).

### Why this is better for a database like Redis:

*   **High Efficiency:** A single thread can handle thousands of connections. This is the model Redis itself uses. It dramatically reduces memory usage and context-switching overhead.
*   **Predictable Performance:** With a single-threaded event loop, there are no data races or need for locks in the hot path, leading to more consistent latency.
*   **Total Control:** We decide exactly what to do when an event occurs, giving us the control needed for advanced features.

## 3. `epoll` and `kqueue`

*   **`epoll`**: A Linux-specific API. It's highly scalable. You create an `epoll` instance, tell it which file descriptors (FDs) to monitor and for what events (e.g., "ready to read"). Then you call `epoll_wait()` to block until one or more events are ready.
*   **`kqueue`**: The equivalent on macOS and BSD systems. It works with a similar concept of "kevents" (kernel events).

Go's `golang.org/x/sys/unix` package provides direct access to these system calls, allowing us to build our own event loop.

## 4. High-Level Implementation Plan in Go

Our goal is to create an `EventLoop` that listens for events from the kernel and dispatches them.

1.  **Get the File Descriptor (FD):** Network listeners and connections in Go are represented by file descriptors in the underlying OS. We need to get this integer FD to register it with `epoll` or `kqueue`. We can use `syscall.RawConn` for this.

2.  **Create the Poller:** We will create a wrapper struct, let's call it `Poller`, that abstracts away the platform-specific details of `epoll` and `kqueue`. It will have methods like `Add(fd)` and `Wait()`. Go's build tags (`//go:build linux`) will be used to compile the correct implementation for each OS.

3.  **The Event Loop:**
    *   The main server loop will change. Instead of calling `ln.Accept()`, we will add the listener's FD to our `Poller`.
    *   The loop will call `Poller.Wait()`, which will block until an event is ready.
    *   `Poller.Wait()` will return a list of FDs that have pending events.
    *   **If the event is on the listener FD:** It means a new client is trying to connect. We call `ln.Accept()` (which will now return immediately without blocking) and add the new connection's FD to our `Poller`.
    *   **If the event is on a client connection FD:** It means the client has sent data. We can read from it.

### Tradeoffs of this Approach:

*   **Pros:**
    *   Extremely high performance and low resource usage, capable of handling C10k and beyond.
    *   Full control over connection handling.
*   **Cons:**
    *   **Much more complex:** We are essentially rebuilding functionality that the Go runtime provides for free. This requires careful handling of low-level details and potential OS-specific edge cases.
    *   **Non-idiomatic Go:** The "goroutine-per-connection" model is the idiomatic way to write network servers in Go. We are deviating from this to achieve Redis-like performance and architecture.

Now, I will create the Vietnamese version of this document. After that, you can tell me if you are ready to start implementing this, and I can guide you through creating the first files for the event loop.
