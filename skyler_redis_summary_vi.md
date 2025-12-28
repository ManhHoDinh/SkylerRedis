# SkylerRedis - Tóm tắt Quá trình Phát triển Dự án

Tài liệu này cung cấp một bản tóm tắt chi tiết về vòng đời phát triển của dự án SkylerRedis, phác thảo quy trình triển khai, các thách thức gặp phải, chiến lược kiểm thử và hướng dẫn thiết lập một cấu trúc master-slave.

---

## 1. Tổng quan Quy trình Phát triển

Dự án được thực hiện theo năm giai đoạn riêng biệt, bám sát theo file `skyler_redis_roadmap.md`.

#### Giai đoạn 1: Nền tảng & Lập kế hoạch
Dự án bắt đầu bằng việc phân tích roadmap để thiết lập một kế hoạch rõ ràng. Một cấu trúc thư mục tài liệu đã được tạo ra với các file giữ chỗ cho các tài liệu kiến trúc trong tương lai, và một danh sách TODO công khai đã được thiết lập để theo dõi tiến độ.

#### Giai đoạn 2: Lõi Networking & Giao thức
Giai đoạn quan trọng này bao gồm việc thay thế mô hình mạng cơ bản ban đầu "goroutine-per-connection" bằng một event loop dựa trên `epoll` hiệu năng cao. Điều này phù hợp với kiến trúc của các server hiện đại như Redis. Mã nguồn đặc thù cho nền tảng (Linux cho `epoll`) đã được xử lý bằng cách sử dụng build tag của Go (`//go:build linux`), và một quy trình làm việc ưu tiên Docker đã được áp dụng để đảm bảo môi trường phát triển và kiểm thử nhất quán. Lớp mạng mới đã được xác minh bằng `redis-cli` để xác nhận chức năng `PING`/`PONG` cơ bản.

#### Giai đoạn 3: Bộ máy Key-Value & Cấu trúc dữ liệu
Giai đoạn này tập trung vào việc xây dựng các khả năng xử lý dữ liệu cốt lõi:
-   **Core KV & TTL**: Triển khai các lệnh `GET`/`SET` với hỗ trợ hết hạn theo thời gian (TTL).
-   **An toàn đồng thời (Concurrency Safety)**: Một bản sửa lỗi quan trọng đã được thực hiện bằng cách thêm mutex lock xung quanh tất cả các truy cập vào kho dữ liệu trung tâm, ngăn chặn race condition.
-   **Cơ chế hết hạn**: Một chiến lược hết hạn kép đã được triển khai:
    1.  **Hết hạn thụ động (Lazy Expiration)**: Các key hết hạn sẽ bị xóa khi được truy cập (ví dụ: `GET`).
    2.  **Hết hạn chủ động (Active Expiration)**: Một tác vụ nền định kỳ lấy mẫu và xóa các key đã hết hạn.
-   **Cấu trúc dữ liệu**: Khả năng của server được mở rộng bằng cách triển khai:
    1.  Cấu trúc dữ liệu `Set` (`SADD`, `SREM`, `SCARD`, v.v.).
    2.  Cấu trúc dữ liệu xác suất: **Bloom Filter** (`BFADD`, `BFEXISTS`) và một triển khai từ đầu của **Count-Min Sketch** (`CMSINCRBY`, `CMSQUERY`).
-   **Xóa Cache (Cache Eviction)**: Một chính sách xóa cache LRU (Least Recently Used) xấp xỉ đã được triển khai, được kích hoạt khi số lượng key vượt quá giới hạn được đặt bởi cờ `--maxmemory`.

#### Giai đoạn 4: Tối ưu hóa & Đồng thời (Sharding)
Để chuẩn bị cho khả năng mở rộng đa lõi, kiến trúc đã được tái cấu trúc từ một kho dữ liệu toàn cục duy nhất sang mô hình phân mảnh (sharded):
-   **Đóng gói Trạng thái**: Tất cả các cấu trúc dữ liệu và trạng thái liên quan được đóng gói trong một struct `Shard`.
-   **Xử lý Lệnh theo Shard**: Tất cả các trình xử lý lệnh đã được sửa đổi để hoạt động trên một instance `Shard` cụ thể.
-   **Định tuyến dựa trên Key**: Một hàm băm (`fnv.New32a`) đã được triển khai trong `GetShardForKey` để định tuyến các key đến shard thích hợp của chúng.
-   **Sửa lỗi đồng thời**: Các lỗi đồng thời liên quan đến việc truy cập map `Shards` mới đã được xác định và khắc phục bằng cách giới thiệu một `sync.RWMutex` chuyên dụng.

#### Giai đoạn 5: Hoàn thiện & Đánh giá
Giai đoạn cuối cùng tập trung vào tài liệu hóa và phân tích hiệu năng:
-   **Tài liệu**: Tất cả các tài liệu kiến trúc liên quan đã được cập nhật để phản ánh event loop mới, mô hình đồng thời, cấu trúc dữ liệu và chính sách xóa.
-   **Benchmarking**: Hiệu năng đã được đo lường bằng `redis-benchmark`. Các workload `SET` và `GET` đã được kiểm tra trên server chạy với 1, 2 và 4 shard để đánh giá khả năng mở rộng. Kết quả đã được phân tích và ghi lại, cho thấy hiệu năng `SET` đã gần đạt mục tiêu 50,000 ops/giây với 4 shard.

---

## 2. Các thách thức & Giải pháp

-   **Không tương thích về kiến trúc**: Lớp mạng ban đầu không có khả năng mở rộng.
    -   **Giải pháp**: Viết lại hoàn toàn lớp mạng thành một event loop dựa trên `epoll`, tạo nền tảng hiệu năng cao.
-   **Đặc thù nền tảng & Lỗi Build**: `epoll` chỉ dành riêng cho Linux, gây ra lỗi build trên các nền tảng khác.
    -   **Giải pháp**: Sử dụng build tag của Go (`//go:build linux`) và áp dụng **quy trình làm việc ưu tiên Docker** để phát triển nhất quán, đa nền tảng.
-   **Vấn đề Dependency (Count-Min Sketch)**: Một số thư viện bên ngoài cho Count-Min Sketch không khả dụng, bị gắn cờ là có rủi ro bảo mật, hoặc có lỗi build không thể giải quyết.
    -   **Giải pháp**: Tự triển khai thuật toán Count-Min Sketch từ đầu trong `internal/datastr`, điều này hoàn toàn phù hợp với mục tiêu học tập low-level của dự án và loại bỏ các vấn đề phụ thuộc bên ngoài.
-   **Lỗi đồng thời (Concurrency Bugs)**:
    1.  **Race Condition trên Kho dữ liệu**: Kho dữ liệu toàn cục ban đầu được truy cập mà không có mutex. **Giải pháp**: Một mutex cho mỗi shard (`shard.Mu`) đã được thêm vào tất cả các lệnh sửa đổi và truy cập dữ liệu.
    2.  **Race Condition trên Map `Shards`**: Map `Shards` bị truy cập đồng thời. **Giải pháp**: Một `sync.RWMutex` chuyên dụng (`shardsMu`) đã được giới thiệu để bảo vệ map.
-   **Panic do con trỏ nil**: Server bị crash khi gặp các lệnh không xác định (ví dụ: `CONFIG GET` từ `redis-benchmark`).
    -   **Giải pháp**: Thêm một câu lệnh `return` vào trường hợp `default` của bộ điều phối lệnh để ngăn việc gọi phương thức trên một đối tượng `nil`.
-   **Logic xóa LRU không chính xác**: Việc xóa diễn ra mạnh hơn so với dự định ban đầu.
    -   **Giải pháp**: Debug bằng cách thêm các câu lệnh print để theo dõi luồng thực thi, phân tích logic và sửa lại điều kiện kích hoạt trong `EvictKeysByLRU`.

---

## 3. Chiến lược Kiểm thử & Xác minh

-   **Quy trình làm việc dựa trên Docker**: Toàn bộ quá trình phát triển và kiểm thử được thực hiện bên trong các container Docker để đảm bảo một môi trường Linux nhất quán và chính xác.
-   **Kiểm thử đơn vị & Chức năng (`redis-cli`)**: Các tính năng mới được kiểm tra từng bước bằng `redis-cli` (chạy trong một container Docker riêng) để xác minh giá trị trả về và các thay đổi trạng thái của server. Điều này rất quan trọng để debug logic xóa LRU.
-   **Benchmarking hiệu năng (`redis-benchmark`)**:
    -   Công cụ tiêu chuẩn `redis-benchmark` đã được sử dụng để đo lường hiệu năng (ops/giây).
    -   Các workload `SET` và `GET` đã được kiểm tra với số lượng client đồng thời cao (`-c 200`).
    -   Server đã được chạy với 1, 2 và 4 shard (`--numshards`) để đánh giá khả năng mở rộng.
-   **Debugging**: Khi các bài kiểm tra thất bại hoặc server bị crash, `docker logs` đã được sử dụng rộng rãi để kiểm tra output của server, bao gồm cả các bản in debug tùy chỉnh được thêm vào để theo dõi luồng chương trình và trạng thái.

---

## 4. Thiết lập Replication (Nhiều Slave)

Bạn có thể chạy nhiều instance SkylerRedis để tạo thành một cấu trúc master-slave. Ví dụ sau sử dụng Docker và ánh xạ các port khác nhau ra máy chủ host.

#### Bước 1: Khởi động Master

Chạy một instance SkylerRedis mà không có cờ `--replicaof`. Instance này sẽ mặc định là master.

```sh
# Chạy master trên port 6379
docker run -d -p 6379:6379 --name skyler-master skyler-redis
```

#### Bước 2: Khởi động các Slave

Chạy các instance SkylerRedis bổ sung, sử dụng cờ `--replicaof` để trỏ đến master. Điểm mấu chốt là cung cấp địa chỉ của master và ánh xạ một port khác cho mỗi slave.

**Lưu ý về `<master_ip>`**: Khi chạy Docker, các container không thể chỉ sử dụng `localhost` hoặc `127.0.0.1` để tham chiếu đến máy chủ host. Trên Docker Desktop (Windows/Mac), bạn thường có thể sử dụng tên DNS đặc biệt là `host.docker.internal`.

```sh
# Khởi động slave đầu tiên, lắng nghe trên port 6380, sao chép từ master trên 6379
docker run -d -p 6380:6380 --name skyler-slave1 skyler-redis --port 6380 --replicaof host.docker.internal 6379

# Khởi động slave thứ hai, lắng nghe trên port 6381
docker run -d -p 6381:6381 --name skyler-slave2 skyler-redis --port 6381 --replicaof host.docker.internal 6379
```

#### Bước 3: Xác minh Replication

1.  **Ghi vào Master**: Kết nối đến master và đặt một key.
    ```sh
    redis-cli -p 6379 SET message "hello from master"
    ```
2.  **Đọc từ Slaves**: Kết nối đến từng slave và lấy key. Dữ liệu đáng lẽ đã được sao chép.
    ```sh
    # Kiểm tra slave 1
    redis-cli -p 6380 GET message
    # Output mong đợi: "hello from master"

    # Kiểm tra slave 2
    redis-cli -p 6381 GET message
    # Output mong đợi: "hello from master"
    ```
Điều này xác nhận rằng liên kết sao chép đang hoạt động và các lệnh đang được lan truyền từ master đến các slave của nó.
