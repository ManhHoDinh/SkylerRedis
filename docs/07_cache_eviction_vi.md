# Tài liệu SkylerRedis

## 7. Xóa Cache (Approximate LRU)

SkylerRedis triển khai cơ chế xóa cache để quản lý việc sử dụng bộ nhớ khi số lượng key được lưu trữ vượt quá giới hạn cấu hình. Điều này rất quan trọng để ngăn chặn lỗi tràn bộ nhớ (out-of-memory) và duy trì hiệu suất trong các kịch bản có lượng dữ liệu vào lớn. Chính sách được triển khai là **Approximate Least Recently Used (LRU)**, tương tự như cách Redis tự xử lý việc xóa bộ nhớ.

### Cấu hình: `--maxmemory`

Giới hạn bộ nhớ được cấu hình thông qua một cờ dòng lệnh:

```
./redis_clone --maxmemory <số_lượng_key>
```

-   **`--maxmemory <N>`**: Đặt số lượng key tối đa (cụ thể là các cặp key-value trong `memory.Store`) mà SkylerRedis sẽ lưu trữ. Nếu giới hạn này bị vượt quá, chính sách xóa sẽ được kích hoạt. Giá trị `0` (mặc định) có nghĩa là không có giới hạn bộ nhớ nào được áp dụng.

### Cơ chế theo dõi LRU

Để triển khai chính sách LRU xấp xỉ, SkylerRedis theo dõi "thời gian truy cập cuối cùng" cho mỗi key bằng cách sử dụng một đồng hồ LRU toàn cục và một trường trong mỗi `Entry`.

#### 1. Trường `Entry.LRU`

Mỗi struct `Entry` trong `internal/memory/storage.go` hiện bao gồm một trường `LRU`:

```go
// internal/memory/storage.go
type Entry struct {
	Value      string
	ExpiryTime time.Time
	LRU        uint64 // Đại diện cho thời gian truy cập gần nhất được ước tính để xóa theo LRU
}
```

Trường `LRU` này lưu trữ một giá trị `uint64` đại diện cho thời gian truy cập gần nhất được ước tính cho key cụ thể đó.

#### 2. Bộ đếm toàn cục `memory.LruClock`

Một bộ đếm toàn cục, `memory.LruClock`, được duy trì trong `internal/memory/main.go`:

```go
// internal/memory/main.go
var (
    // ...
    LruClock         uint64 // Đồng hồ LRU toàn cục cho việc xóa theo LRU xấp xỉ
    // ...
)
```

-   **Khởi tạo & Cập nhật**:
    -   Bất cứ khi nào một key được truy cập (ví dụ: thông qua `GET`), trường `Entry.LRU` của nó được cập nhật bằng giá trị hiện tại của `memory.LruClock`.
    -   Bất cứ khi nào một key được tạo hoặc sửa đổi (ví dụ: thông qua `SET`), trường `Entry.LRU` của nó được khởi tạo bằng giá trị `memory.LruClock` hiện tại.
    -   Sau mỗi thao tác như vậy, `memory.LruClock` được tăng lên. Điều này đảm bảo rằng các key được truy cập/sửa đổi gần đây có giá trị `LRU` cao hơn (gần đây hơn).

### Thuật toán xóa theo LRU xấp xỉ

Logic xóa cốt lõi nằm trong hàm `EvictKeysByLRU()` trong `internal/memory/eviction.go`. Thuật toán này được kích hoạt khi số lượng key trong `memory.Store` vượt quá giới hạn `MaxMemory`.

#### Cách hoạt động (`EvictKeysByLRU`):

1.  **Điều kiện kích hoạt**:
    -   **Định kỳ**: Một goroutine nền (khởi động trong `app/main.go`) định kỳ gọi `EvictKeysByLRU()` (ví dụ: mỗi 100ms), cùng với việc xóa key hết hạn.
    -   **Khi ghi**: Sau các lệnh nhất định sửa đổi hoặc thêm key (ví dụ: `SET`), `EvictKeysByLRU()` được gọi ngay lập tức để ngăn bộ nhớ tăng đáng kể vượt quá giới hạn.

2.  **Kiểm tra giới hạn bộ nhớ**: Hàm trước tiên kiểm tra xem `memory.MaxMemory` đã được đặt chưa và liệu `len(memory.Store)` có thực sự vượt quá giới hạn này không. Nếu không, hàm sẽ trả về.

3.  **Kích thước mục tiêu**: Việc xóa không dừng lại chính xác ở `MaxMemory`. Thay vào đó, nó nhằm mục đích giảm kích thước store xuống một `targetSize`, là `MaxMemory` nhân với `evictionTargetRatio` (ví dụ: 95% của `MaxMemory`). Điều này ngăn thuật toán xóa liên tục bị kích hoạt cho các biến động nhỏ.

4.  **Lấy mẫu và vòng lặp xóa**:
    -   Trong khi `len(memory.Store)` lớn hơn `targetSize`:
        -   Một số lượng nhỏ các key (ví dụ: `lruEvictionSample = 10` key) được lấy mẫu ngẫu nhiên từ `memory.Store`. Do thứ tự lặp map ngẫu nhiên của Go, việc lặp đến kích thước mẫu cung cấp một mẫu đủ ngẫu nhiên.
        -   Trong số các key được lấy mẫu này, key có giá trị `Entry.LRU` *thấp nhất* được xác định. Key này được coi là "ít được sử dụng gần đây nhất" trong mẫu đó.
        -   Key được xác định sau đó bị xóa khỏi `memory.Store`.

Quá trình lấy mẫu và xóa lặp đi lặp lại này tiếp tục cho đến khi `len(memory.Store)` giảm xuống dưới hoặc bằng `targetSize`. Cách tiếp cận xác suất này đảm bảo việc thu hồi bộ nhớ hiệu quả mà không cần thực hiện quét toàn bộ không gian key, tuân thủ mục tiêu xóa `O(1)` như đã đề cập trong roadmap.
