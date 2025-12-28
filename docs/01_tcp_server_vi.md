# Bước 1: Xây dựng một TCP Server cơ bản

Tài liệu này hướng dẫn bạn thực hiện bước đầu tiên của dự án SkylerRedis: tạo một TCP server cơ bản bằng Go.

## 1. Mục tiêu

Mục tiêu là tạo ra một TCP server đơn giản có thể chấp nhận kết nối từ các client. Server này sẽ tạo nền tảng cho cơ sở dữ liệu tương thích với Redis của chúng ta.

## 2. Các bước triển khai

### 2.1. Cấu trúc dự án

Chúng ta sẽ sử dụng cấu trúc thư mục sau:

```
skyler-redis/
├── cmd/
│   └── skyler-redis/
│       └── main.go   # Điểm khởi đầu của ứng dụng chính
├── server/
│   └── server.go     # Logic của server
└── go.mod            # Định nghĩa Go module
```

### 2.2. Khởi tạo Go Module

Chúng ta sẽ bắt đầu bằng việc khởi tạo một Go module mới. Điều này được thực hiện bằng lệnh `go mod init`. Một module là một tập hợp các package Go được lưu trữ trong một cây tệp với tệp `go.mod` ở gốc của nó.

```sh
go mod init github.com/ManhHoDinh/SkylerRedis
```

### 2.3. Logic của Server (`server/server.go`)

Chúng ta sẽ tạo một `Server` struct chứa cấu hình và listener. Hàm `NewServer` sẽ khởi tạo một server mới, và phương thức `Start` sẽ bắt đầu lắng nghe các kết nối đến.

Hiện tại, server sẽ chỉ đơn giản là chấp nhận một kết nối và sau đó đóng nó ngay lập tức. Điều này xác minh rằng nền tảng mạng đang hoạt động.

### 2.4. Ứng dụng chính (`cmd/skyler-redis/main.go`)

Hàm `main` sẽ:
1.  Tạo một instance server mới.
2.  Gọi phương thức `Start` để chạy server.
3.  Xử lý bất kỳ lỗi tiềm ẩn nào trong quá trình khởi động.

## 3. Cách chạy

Sau khi tạo các tệp, bạn có thể chạy server từ thư mục gốc của dự án:

```sh
go run ./cmd/skyler-redis
```

Bạn có thể kiểm tra kết nối bằng một công cụ như `telnet` hoặc `netcat`:

```sh
telnet localhost 6379
```

Server sẽ chấp nhận kết nối và ngay lập tức đóng nó.
