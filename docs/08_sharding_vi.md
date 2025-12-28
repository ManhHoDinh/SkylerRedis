# Tài liệu SkylerRedis

## 8. Sharding và phân phối dữ liệu

SkylerRedis triển khai cơ chế sharding để phân phối tập dữ liệu của mình trên nhiều đơn vị lưu trữ độc lập, cho phép tận dụng tốt hơn các bộ xử lý đa lõi và tăng cường khả năng mở rộng tổng thể. Tài liệu này mô tả chi tiết cách dữ liệu được phân mảnh và quản lý trong hệ thống.

### Khái niệm Shard

Thay vì một kho dữ liệu toàn cục duy nhất, SkylerRedis phân vùng dữ liệu của mình thành nhiều instance `Shard` riêng biệt. Mỗi `Shard` là một đơn vị tự chứa, nắm giữ bộ cấu trúc dữ liệu riêng (kho Key-Value, Sets, Bloom Filters, Count-Min Sketches, Lists, Streams), cùng với đồng hồ LRU, giới hạn bộ nhớ và mutex để đồng bộ hóa nội bộ.

Struct `Shard` được định nghĩa trong `internal/memory/shard.go`:

```go
// internal/memory/shard.go
type Shard struct {
	Store            map[string]Entry
	Sets             map[string]map[string]struct{}
	BloomFilters     map[string]*bloom.BloomFilter
	CountMinSketches map[string]*Sketch
	RPush            map[string][]string
	Stream           map[string]StreamEntry
	StreamIDs        []string
	LruClock         uint64
	MaxMemory        int
	Mu               sync.Mutex // Mutex cho dữ liệu cụ thể của shard này
}
```

### Quản lý Shard toàn cục

Package `internal/memory/main.go` chịu trách nhiệm khởi tạo và quản lý các instance `Shard` này trên toàn cục:

-   **`numShards`**: Được cấu hình thông qua cờ dòng lệnh `--numshards`, tham số này xác định số lượng instance `Shard` mà server sẽ tạo.
-   **`memory.Shards`**: Một map toàn cục (`map[int]*Shard`) chứa các tham chiếu đến tất cả các instance `Shard` đã được tạo.
-   **`memory.InitShards(numShards, maxMemory)`**: Được gọi trong quá trình khởi động server (`app/main.go`), hàm này tạo `numShards` instance của `Shard` và điền vào map `memory.Shards`. Mỗi shard được khởi tạo với giới hạn `maxMemory` đã cấu hình.
-   **`shardsMu sync.RWMutex`**: Một `RWMutex` chuyên dụng (`shardsMu`) bảo vệ truy cập vào map `memory.Shards`, đảm bảo các thao tác đọc và ghi trên map của các shard an toàn trong môi trường đa luồng.

### Định tuyến dựa trên Key

Khi một lệnh client thao tác trên một key, SkylerRedis sẽ xác định instance `Shard` cụ thể nào chịu trách nhiệm cho key đó. Quá trình này được gọi là định tuyến dựa trên key.

#### `memory.GetShardForKey(key string) *Shard`

Hàm này, nằm trong `internal/memory/main.go`, là thành phần trung tâm cho việc định tuyến key:

1.  **Băm (Hashing)**: Nó sử dụng thuật toán băm FNV-1a (Fowler-Noll-Vo) (`hash/fnv`) để tính toán giá trị băm cho `key` đầu vào.
    ```go
    h := fnv.New32a()
    h.Write([]byte(key))
    hashValue := h.Sum32()
    ```
2.  **Tính toán Shard ID**: `shardID` được xác định bằng cách lấy modulo của giá trị băm với tổng số shard đã cấu hình:
    ```go
    shardID := int(hashValue % uint32(numShards))
    ```
    Điều này đảm bảo rằng key luôn được ánh xạ một cách nhất quán đến cùng một shard trong tất cả các thao tác.
3.  **Truy xuất Shard**: Hàm sau đó trả về instance `Shard` tương ứng với `shardID` đã tính toán từ map `memory.Shards`.

### Lợi ích của Sharding

-   **Khả năng mở rộng đa lõi**: Bằng cách phân phối dữ liệu trên nhiều instance `Shard`, SkylerRedis có thể xử lý các lệnh nhắm mục tiêu đến các key khác nhau (được băm vào các shard khác nhau) một cách song song. Mỗi shard hoạt động phần lớn độc lập, giảm tranh chấp tài nguyên và cho phép hệ thống tận dụng nhiều lõi CPU.
-   **Giảm tranh chấp khóa**: Mỗi `Shard` có `Mu` (mutex) riêng của nó. Các lệnh chỉ cần chiếm giữ khóa cho shard cụ thể mà chúng đang thao tác, thay vì một lock toàn cục cho toàn bộ tập dữ liệu. Điều này cải thiện đáng kể khả năng đồng thời, đặc biệt dưới tải ghi cao.
-   **Cô lập**: Các lỗi hoặc tắc nghẽn hiệu suất trong một shard ít có khả năng ảnh hưởng đến các shard khác, góp phần tạo nên một hệ thống bền vững hơn.
-   **Quản lý bộ nhớ**: Mỗi shard quản lý giới hạn bộ nhớ (`MaxMemory`) và chính sách xóa LRU của riêng mình một cách độc lập.

### Ví dụ sử dụng

Để khởi động SkylerRedis với sharding được bật, sử dụng cờ `--numshards`:

```sh
# Khởi động với 4 shard, mỗi shard có maxmemory là 1000 key
docker run -d -p 6379:6379 --name skyler-redis-server skyler-redis --numshards 4 --maxmemory 1000
```

Khi bạn thực hiện thao tác `SET` hoặc `GET`, `memory.GetShardForKey` sẽ xác định shard nào trong số 4 shard chịu trách nhiệm cho key đó.
