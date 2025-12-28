# Bước 2: I/O Multiplexing với epoll/kqueue

Tài liệu này giải thích "tại sao" và "làm thế nào" để triển khai một vòng lặp sự kiện (event loop) dựa trên I/O multiplexing, đây là cốt lõi của server hiệu năng cao của chúng ta.

## 1. Vấn đề với vòng lặp `Accept()` đơn giản

Trong TCP server cơ bản của chúng ta, lệnh gọi `ln.Accept()` là **blocking (chặn)**. Server bị kẹt ở dòng này cho đến khi có một client mới kết nối. Tương tự, việc đọc dữ liệu từ một kết nối client (`conn.Read()`) cũng là một hoạt động blocking.

Một cách phổ biến nhưng không hiệu quả để xử lý nhiều client là mô hình **"thread-per-connection"** (hoặc trong Go là **"goroutine-per-connection"**).

```go
// Mô hình không hiệu quả
for {
    conn, _ := ln.Accept()
    go handleConnection(conn) // Một goroutine mới cho mỗi client
}
```

### Đánh đổi của mô hình Goroutine-per-Connection:

*   **Ưu điểm:**
    *   Rất đơn giản để viết và hiểu.
*   **Nhược điểm:**
    *   **Sử dụng bộ nhớ cao:** Mỗi goroutine tiêu tốn ít nhất 2KB cho stack. 10.000 client sẽ có nghĩa là ít nhất 20MB bộ nhớ stack, cộng với chi phí quản lý của Go runtime.
    *   **Chi phí của Scheduler:** Go scheduler phải quản lý một số lượng lớn goroutine. Mặc dù nó rất hiệu quả, nó không phải là miễn phí. Đối với một cơ sở dữ liệu hiệu năng cao, chúng ta muốn kiểm soát trực tiếp hơn.
    *   **Không có kiểm soát chi tiết:** Rất khó để triển khai các tính năng như backpressure hoặc ưu tiên các hoạt động nhất định khi scheduler hoàn toàn kiểm soát.

## 2. Giải pháp: I/O Multiplexing

I/O Multiplexing cho phép một luồng duy nhất giám sát nhiều hoạt động I/O (như chấp nhận kết nối mới hoặc đọc từ các kết nối hiện có) một cách đồng thời.

Thay vì hỏi kernel "Kết nối cụ thể này đã sẵn sàng để đọc chưa?" (blocking), chúng ta hỏi "Có *bất kỳ* kết nối nào trong số những kết nối tôi quan tâm đã sẵn sàng cho một hoạt động I/O chưa?" (có thể không blocking).

Điều này đạt được thông qua các system call như `epoll` (trên Linux) và `kqueue` (trên macOS/BSD).

### Tại sao điều này tốt hơn cho một cơ sở dữ liệu như Redis:

*   **Hiệu quả cao:** Một luồng duy nhất có thể xử lý hàng ngàn kết nối. Đây là mô hình mà chính Redis sử dụng. Nó giảm đáng kể việc sử dụng bộ nhớ và chi phí chuyển đổi ngữ cảnh (context-switching).
*   **Hiệu suất có thể dự đoán:** Với một event loop đơn luồng, không có cuộc đua dữ liệu (data race) hoặc cần khóa (lock) trong đường dẫn nóng (hot path), dẫn đến độ trễ nhất quán hơn.
*   **Toàn quyền kiểm soát:** Chúng ta quyết định chính xác phải làm gì khi một sự kiện xảy ra, cho chúng ta quyền kiểm soát cần thiết cho các tính năng nâng cao.

## 3. `epoll` và `kqueue`

*   **`epoll`**: Một API dành riêng cho Linux. Nó có khả năng mở rộng cao. Bạn tạo một instance `epoll`, cho nó biết file descriptor (FD) nào cần theo dõi và cho những sự kiện nào (ví dụ: "sẵn sàng để đọc"). Sau đó, bạn gọi `epoll_wait()` để chặn cho đến khi một hoặc nhiều sự kiện sẵn sàng.
*   **`kqueue`**: Tương đương trên các hệ thống macOS và BSD. Nó hoạt động với một khái niệm tương tự về "kevents" (sự kiện kernel).

Package `golang.org/x/sys/unix` của Go cung cấp quyền truy cập trực tiếp vào các system call này, cho phép chúng ta xây dựng event loop của riêng mình.

## 4. Kế hoạch triển khai cấp cao trong Go

Mục tiêu của chúng ta là tạo ra một `EventLoop` lắng nghe các sự kiện từ kernel và điều phối chúng.

1.  **Lấy File Descriptor (FD):** Các listener và connection mạng trong Go được đại diện bởi các file descriptor trong hệ điều hành bên dưới. Chúng ta cần lấy số nguyên FD này để đăng ký nó với `epoll` hoặc `kqueue`. Chúng ta có thể sử dụng `syscall.RawConn` cho việc này.

2.  **Tạo Poller:** Chúng ta sẽ tạo một struct bao bọc, gọi là `Poller`, để trừu tượng hóa các chi tiết cụ thể của nền tảng `epoll` và `kqueue`. Nó sẽ có các phương thức như `Add(fd)` và `Wait()`. Các build tag của Go (`//go:build linux`) sẽ được sử dụng để biên dịch đúng phần triển khai cho mỗi hệ điều hành.

3.  **Event Loop:**
    *   Vòng lặp server chính sẽ thay đổi. Thay vì gọi `ln.Accept()`, chúng ta sẽ thêm FD của listener vào `Poller` của chúng ta.
    *   Vòng lặp sẽ gọi `Poller.Wait()`, nó sẽ chặn cho đến khi một sự kiện sẵn sàng.
    *   `Poller.Wait()` sẽ trả về một danh sách các FD có sự kiện đang chờ xử lý.
    *   **Nếu sự kiện xảy ra trên FD của listener:** Điều đó có nghĩa là một client mới đang cố gắng kết nối. Chúng ta gọi `ln.Accept()` (lúc này sẽ trả về ngay lập tức mà không bị chặn) và thêm FD của kết nối mới vào `Poller`.
    *   **Nếu sự kiện xảy ra trên FD của kết nối client:** Điều đó có nghĩa là client đã gửi dữ liệu. Chúng ta có thể đọc từ nó.

### Đánh đổi của phương pháp này:

*   **Ưu điểm:**
    *   Hiệu suất cực cao và sử dụng tài nguyên thấp, có khả năng xử lý C10k và hơn thế nữa.
    *   Toàn quyền kiểm soát việc xử lý kết nối.
*   **Nhược điểm:**
    *   **Phức tạp hơn nhiều:** Về cơ bản, chúng ta đang xây dựng lại chức năng mà Go runtime cung cấp miễn phí. Điều này đòi hỏi phải xử lý cẩn thận các chi tiết cấp thấp và các trường hợp đặc biệt tiềm ẩn của từng hệ điều hành.
    *   **Không theo chuẩn Go (Non-idiomatic Go):** Mô hình "goroutine-per-connection" là cách viết máy chủ mạng thông thường trong Go. Chúng ta đang đi chệch hướng này để đạt được hiệu suất và kiến trúc giống như Redis.

Bây giờ, tôi đã tạo xong phiên bản tiếng Việt của tài liệu này. Sau đây, bạn có thể cho tôi biết nếu bạn đã sẵn sàng để bắt đầu triển khai điều này, và tôi có thể hướng dẫn bạn tạo các tệp đầu tiên cho event loop.
