# Step 1: Building a Basic TCP Server

This document guides you through the first step of the SkylerRedis project: creating a basic TCP server in Go.

## 1. Objective

The goal is to create a simple TCP server that can accept connections from clients. This server will form the foundation for our Redis-compatible database.

## 2. Implementation Steps

### 2.1. Project Structure

We will use the following directory structure:

```
skyler-redis/
├── cmd/
│   └── skyler-redis/
│       └── main.go   # Main application entry point
├── server/
│   └── server.go     # Server logic
└── go.mod            # Go module definition
```

### 2.2. Go Module Initialization

We'll start by initializing a new Go module. This is done using the `go mod init` command. A module is a collection of Go packages stored in a file tree with a `go.mod` file at its root.

```sh
go mod init github.com/ManhHoDinh/SkylerRedis
```

### 2.3. Server Logic (`server/server.go`)

We will create a `Server` struct that holds the configuration and listener. The `NewServer` function will initialize a new server, and the `Start` method will begin listening for incoming connections.

For now, the server will simply accept a connection and then close it immediately. This verifies that the networking foundation is working.

### 2.4. Main Application (`cmd/skyler-redis/main.go`)

The `main` function will:
1.  Create a new server instance.
2.  Call the `Start` method to run the server.
3.  Handle any potential errors during startup.

## 3. How to Run

After creating the files, you can run the server from the root of the project directory:

```sh
go run ./cmd/skyler-redis
```

You can test the connection using a tool like `telnet` or `netcat`:

```sh
telnet localhost 6379
```

The server should accept the connection and immediately close it.
