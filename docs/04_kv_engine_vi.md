# Tài liệu SkylerRedis

## 4. Lõi lưu trữ Key-Value

Trái tim của SkylerRedis là bộ máy lưu trữ Key-Value (KV) trong bộ nhớ. Tài liệu này mô tả chi tiết kiến trúc của nó, tập trung vào kho dữ liệu trung tâm, sự an toàn trong môi trường đa luồng (thread safety), và cơ chế hết hạn (expiration). Thiết kế này ưu tiên sự đơn giản và hiệu năng, lấy cảm hứng từ các nguyên tắc cốt lõi của Redis.

### Kho dữ liệu trung tâm

Tất cả dữ liệu key-value được lưu trong một map toàn cục duy nhất đặt tại package `internal/memory`:

```go
// internal/memory/main.go
var (
    // ...
    Store = make(map[string]Entry)
    // ...
)
```

`key` là một `string` đơn giản. `value` là một struct `Entry`, không chỉ chứa dữ liệu mà còn chứa siêu dữ liệu (metadata) về vòng đời của nó.

```go
// internal/memory/storage.go
type Entry struct {
	Value      string
	ExpiryTime time.Time
}
```

- **`Value`**: Giá trị `string` được liên kết với key.
- **`ExpiryTime`**: Một đối tượng `time.Time`. Nếu thời gian này là giá trị zero (`time.Time{}`), key sẽ không bao giờ hết hạn. Ngược lại, nó đại diện cho thời điểm tuyệt đối mà tại đó key trở nên không hợp lệ.

### An toàn đa luồng (Thread Safety)

Là một server mạng, SkylerRedis phải xử lý các yêu cầu đồng thời một cách an toàn. Kiến trúc hiện tại sử dụng một event loop duy nhất, xử lý các lệnh một cách tuần tự cho tất cả client. Tuy nhiên, chúng ta cũng có các tác vụ nền, chẳng hạn như cơ chế xóa key, chạy trong một goroutine riêng biệt.

Để ngăn chặn xung đột dữ liệu (race condition) giữa event loop chính và các tác vụ nền, tất cả truy cập vào các kho dữ liệu trung tâm (`Store`, `Sets`, `BloomFilters`, v.v.) đều được bảo vệ bởi một Mutex toàn cục duy nhất:

```go
// internal/memory/main.go
var (
    // ...
    Mu = sync.Mutex{}
    // ...
)
```

Bất kỳ trình xử lý lệnh nào đọc hoặc ghi vào bộ nhớ chia sẻ đều phải chiếm giữ lock này. Câu lệnh `defer` được sử dụng để đảm bảo lock luôn được giải phóng, ngay cả khi một hàm kết thúc sớm.

**Ví dụ từ lệnh `GET`:**
```go
// internal/command/get.go
func (Get) Handle(...) {
    // ...
	memory.Mu.Lock()
	defer memory.Mu.Unlock()

	entry, ok := memory.Store[key]
    // ... logic để kiểm tra hết hạn và trả về giá trị
}
```

Chiến lược khóa "thô" (coarse-grained) và đơn giản này hiện tại đang rất hiệu quả. Khi hệ thống phát triển, chúng ta có thể khám phá các kỹ thuật khóa "mịn" hơn (fine-grained locking) như sharding (phân mảnh) không gian key với một lock cho mỗi shard để cải thiện hơn nữa khả năng đồng thời, như đã nêu trong roadmap của dự án.

### Chiến lược hết hạn kép

Một tính năng quan trọng của một cơ sở dữ liệu giống Redis là khả năng tự động hết hạn của các key. SkylerRedis sử dụng một chiến lược kép để quản lý việc này, cân bằng giữa khả năng phản hồi nhanh và việc thu hồi bộ nhớ hiệu quả.

#### 1. Hết hạn thụ động (Passive / Lazy Expiration)

Các key được kiểm tra hết hạn mỗi khi chúng được truy cập. Lệnh `GET` (và các lệnh tương tự) chứa logic để kiểm tra xem `ExpiryTime` của một key đã trôi qua hay chưa.

- **Cách hoạt động:** Khi một client yêu cầu một key, server sẽ lấy ra `Entry`. Nó so sánh `ExpiryTime` của entry với thời gian hiện tại (`time.Now()`).
- **Hành động:** Nếu key được phát hiện đã hết hạn, nó sẽ bị xóa ngay lập tức khỏi `Store`, và lệnh sẽ hoạt động như thể key đó không bao giờ tồn tại (trả về `nil`).

Đây là một chiến lược cực kỳ hiệu quả vì nó không gây tốn chi phí cho các key không bao giờ được truy cập lại.

#### 2. Hết hạn chủ động (Active, Sampling-based Expiration)

Chỉ dựa vào hết hạn thụ động là không đủ, vì các key được đặt một lần và không bao giờ được truy cập lại sẽ tồn tại trong bộ nhớ mãi mãi, gây rò rỉ bộ nhớ.

Để giải quyết vấn đề này, SkylerRedis chạy một tác vụ nền để chủ động dọn dẹp các key đã hết hạn. Logic này được triển khai trong `internal/memory/eviction.go`.

- **Cách hoạt động:** Một goroutine nền, được khởi động trong `app/main.go`, chạy theo một khoảng thời gian đều đặn (ví dụ: mỗi 100ms).
- **Lấy mẫu (Sampling):** Tại mỗi tick, hàm `EvictRandomKeys` sẽ chọn một mẫu nhỏ, ngẫu nhiên các key (ví dụ: 20 key) từ `Store` chính.
- **Hành động:** Nó kiểm tra `ExpiryTime` cho mỗi key trong mẫu và xóa bất kỳ key nào đã hết hạn.

Cách tiếp cận xác suất này đảm bảo rằng bộ nhớ từ các key đã hết hạn cuối cùng sẽ được thu hồi, mà không cần phải thực hiện một cuộc quét tốn kém trên toàn bộ không gian key. Đây cũng là chiến lược cốt lõi mà chính Redis sử dụng để xóa key chủ động.