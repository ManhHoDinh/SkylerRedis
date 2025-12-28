# Tài liệu SkylerRedis

## 5. Mô hình đồng thời (Concurrency): Kiến trúc Thread-per-Shard

SkylerRedis được thiết kế cho thông lượng cao, có khả năng xử lý hàng chục nghìn kết nối đồng thời. Để đạt được điều này, nó từ bỏ mô hình đơn giản nhưng hạn chế "goroutine-per-connection" để chuyển sang kiến trúc **thread-per-shard** hiệu quả cao, được cung cấp sức mạnh bởi cơ chế thông báo sự kiện I/O `epoll` của Linux. Mô hình này lấy cảm hứng từ cách tiếp cận của Redis Cluster để mở rộng.

### Vấn đề của "Goroutine-per-Connection"

Cách đơn giản nhất để viết một server mạng trong Go là tạo một goroutine mới cho mỗi kết nối đến:

```go
// Mô hình cũ, đơn giản
for {
    conn, _ := listener.Accept()
    go handleConnection(conn) // Một goroutine mới cho mỗi client
}
```

Mặc dù hoạt động tốt với số lượng client nhỏ, mô hình này gặp phải chi phí vận hành (overhead) đáng kể khi mở rộng:
- **Sử dụng bộ nhớ cao:** Mỗi goroutine tiêu thụ ít nhất 2KB không gian stack. 10.000 kết nối sẽ tốn ít nhất 20MB bộ nhớ chỉ để chứa stack.
- **Chi phí của Scheduler:** Go runtime scheduler phải quản lý một số lượng lớn goroutine. Khi số lượng kết nối tăng lên, chi phí lập lịch và chuyển đổi ngữ cảnh (context-switching) giữa chúng trở thành một nút thắt cổ chai lớn về hiệu năng.

### Kiến trúc Thread-per-Shard

SkylerRedis triển khai kiến trúc "shared-nothing, thread-per-shard". Thay vì một kho dữ liệu toàn cục duy nhất, toàn bộ tập dữ liệu được phân vùng trên nhiều instance `Shard` độc lập. Mỗi `Shard` sau đó được quản lý bởi event loop chuyên dụng của riêng nó.

Các thành phần cốt lõi của kiến trúc này là:

#### 1. Đóng gói Shard

-   Tất cả các cấu trúc dữ liệu (`Store`, `Sets`, `BloomFilters`, `CountMinSketches`, `RPush`, `Stream`), trạng thái liên quan (`LruClock`, `MaxMemory`), và mutex (`Mu`) của chúng đều được đóng gói trong một struct `memory.Shard`. Điều này đảm bảo rằng mỗi shard hoàn toàn tự chứa và hoạt động độc lập.
-   Package `memory` giờ đây duy trì một map các instance `Shard` này (`memory.Shards`).

#### 2. Băm Key và định tuyến

-   Khi một lệnh client đến, key được trích xuất (thường là `args[1]`).
-   Một hàm băm (`fnv.New32a`) được áp dụng cho key.
-   Giá trị băm sau đó được sử dụng để xác định shard đích: `shardID := int(h.Sum32() % uint32(numShards))`.
-   Hàm `memory.GetShardForKey(key)` truy xuất instance `Shard` chính xác, đảm bảo rằng tất cả các thao tác trên một key cụ thể luôn nhắm mục tiêu đến cùng một shard.

#### 3. Nhiều Event Loops

-   Thay vì một event loop toàn cục duy nhất, `app/main.go` giờ đây khởi tạo nhiều instance `Shard` (được cấu hình qua `--numshards`).
-   Mỗi `Shard` chạy event loop chuyên dụng của riêng nó cho các tác vụ nền (xóa key hết hạn, xóa key theo LRU) trong một goroutine riêng biệt.
-   Đối với các thao tác hướng client, một `EventLoop` duy nhất từ `internal/eventloop` vẫn được sử dụng để chấp nhận kết nối và thăm dò các sự kiện I/O. Tuy nhiên, khi một sự kiện xảy ra, lệnh được định tuyến ngay lập tức đến `Shard` chính xác dựa trên key của nó và được thực thi đồng bộ trong ngữ cảnh của shard đó.

#### 4. Mô hình đồng thời

-   Server duy trì một `EventLoop` toàn cục (từ `internal/eventloop`) để giám sát tất cả các kết nối client bằng cách sử dụng `epoll`.
-   Khi client gửi một lệnh, `ReadCallback` của `EventLoop` trích xuất key, sử dụng `memory.GetShardForKey()` để tìm `Shard` liên quan.
-   Lệnh sau đó được gửi đến `command.HandleCommand`, hàm này thực thi nó trên instance `Shard` cụ thể.
-   Điều quan trọng là, các lệnh hoạt động trên dữ liệu cụ thể của shard chiếm giữ và giải phóng mutex (`shard.Mu`) thuộc về shard cụ thể đó. Điều này đảm bảo an toàn đa luồng *trong nội bộ* shard.

### Lợi ích của mô hình Thread-per-Shard

-   **Khả năng mở rộng đa lõi thực sự**: Bằng cách phân vùng dữ liệu và hoạt động trên các instance `Shard` độc lập với mutex của riêng chúng, SkylerRedis có thể tận dụng hiệu quả nhiều lõi CPU. Các thao tác trên các key thuộc các shard khác nhau có thể tiến hành song song mà không bị tranh chấp khóa toàn cục.
-   **Khả năng đồng thời nâng cao**: Mỗi shard quản lý dữ liệu của riêng nó một cách độc lập, giảm phạm vi khóa. Điều này làm tăng đáng kể số lượng hoạt động đồng thời mà server có thể xử lý so với một lock toàn cục duy nhất.
-   **Độ trễ thấp hơn**: Giảm tranh chấp khóa có nghĩa là các lệnh được xử lý nhanh hơn, dẫn đến độ trễ thấp hơn và nhất quán hơn.
-   **Khả năng mở rộng lớn**: Server có thể xử lý số lượng kết nối lớn hơn và tập dữ liệu lớn hơn, vì các tài nguyên được phân phối và quản lý trên nhiều đơn vị độc lập.
-   **Dễ bảo trì hơn**: Việc đóng gói dữ liệu và logic trong `Shard`s làm cho codebase trở nên modular hơn và dễ hiểu hơn.

Kiến trúc này cung cấp một nền tảng vững chắc để xây dựng một cơ sở dữ liệu tương thích Redis hiệu suất cao, có khả năng mở rộng.
