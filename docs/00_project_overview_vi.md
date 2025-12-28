# SkylerRedis - Tổng quan dự án

Tài liệu này cung cấp cái nhìn tổng quan về dự án SkylerRedis.

## 1. Mục tiêu

SkylerRedis là một dự án học tập nhằm **xây dựng lại Redis từ góc độ low-level**. Trọng tâm chính là hiểu rõ kiến trúc và các nguyên tắc thiết kế nội bộ của Redis, không phải là tạo ra một bản sao đầy đủ tính năng.

## 2. Các chủ đề cốt lõi

Dự án sẽ được triển khai bằng **Go (Golang)** và sẽ bao gồm các lĩnh vực chính sau:

-   **Networking & I/O Multiplexing**: Xây dựng một TCP server non-blocking sử dụng các system call cấp thấp (`epoll`, `kqueue`).
-   **Giao thức tuần tự hóa Redis (RESP)**: Triển khai một bộ phân tích (parser) tùy chỉnh cho giao thức RESP2 để giao tiếp với `redis-cli`.
-   **Công cụ Key-Value cốt lõi**: Triển khai các lệnh cơ bản như `GET`, `SET`, `DEL` với hỗ trợ Time-To-Live (TTL).
-   **Các cấu trúc dữ liệu nâng cao**: Triển khai Set, Bloom Filter và Count-Min Sketch.
-   **Cơ chế loại bỏ cache (Cache Eviction)**: Triển khai chính sách loại bỏ gần đúng LRU (Least Recently Used).
-   **Mô hình đồng시 (Concurrency Model)**: Khám phá kiến trúc shared-nothing, thread-per-shard để đạt được khả năng mở rộng.

## 3. Kiến trúc

Kiến trúc tổng thể sẽ là:

```
Client
  ↓
TCP Listener (luồng chính)
  ↓
Shard Router (băm key)
  ↓
Event Loop (một trên mỗi luồng OS/shard)
  ↓
Go KV Engine (Go map + các cấu trúc dữ liệu tùy chỉnh)
```

## 4. Mục tiêu học tập

-   Hiểu sâu về hoạt động bên trong của Redis.
-   Nâng cao kỹ năng lập trình và thiết kế hệ thống cấp thấp.
-   Xây dựng một dự án portfolio đủ mạnh để thảo luận trong các cuộc phỏng vấn kỹ sư backend cấp mid/senior.
