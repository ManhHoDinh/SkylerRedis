# SkylerRedis - Project Overview

This document provides a high-level overview of the SkylerRedis project.

## 1. Goal

SkylerRedis is a learning project to **rebuild Redis from a low-level perspective**. The primary focus is on understanding the internal architecture and design principles of Redis, not on creating a feature-complete clone.

## 2. Core Topics

The project will be implemented in **Go (Golang)** and will cover the following key areas:

-   **Networking & I/O Multiplexing**: Building a non-blocking TCP server using low-level system calls (`epoll`, `kqueue`).
-   **Redis Serialization Protocol (RESP)**: Implementing a custom parser for the RESP2 protocol to communicate with `redis-cli`.
-   **Core Key-Value Engine**: Implementing basic commands like `GET`, `SET`, `DEL` with Time-To-Live (TTL) support.
-   **Advanced Data Structures**: Implementing Sets, Bloom Filters, and Count-Min Sketches.
-   **Cache Eviction**: Implementing an approximate LRU (Least Recently Used) eviction policy.
-   **Concurrency Model**: Exploring a shared-nothing, thread-per-shard architecture to achieve scalability.

## 3. Architecture

The general architecture will be:

```
Client
  ↓
TCP Listener (main thread)
  ↓
Shard Router (hashes the key)
  ↓
Event Loop (one per OS thread/shard)
  ↓
Go KV Engine (Go map + custom data structures)
```

## 4. Learning Objectives

-   Gain a deep understanding of Redis's internal workings.
-   Improve skills in low-level systems programming and design.
-   Build a portfolio project suitable for discussing in mid/senior backend engineering interviews.
