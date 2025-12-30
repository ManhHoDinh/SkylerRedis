# 14. Phân tích Thiết kế: Sao chép (Replication)

## 1. Mục tiêu

Triển khai một cơ chế sao chép (replication) từ master tới slave nhằm hai mục đích chính:
1.  **Tăng khả năng đọc (Read Scaling):** Phân tán các truy vấn đọc qua nhiều slave để giảm tải cho master.
2.  **Dự phòng dữ liệu (Data Redundancy):** Tạo các bản sao của dữ liệu để phòng trường hợp master gặp sự cố.

Cơ chế này phải xử lý được hai giai đoạn: **đồng bộ hóa ban đầu** (initial synchronization) khi một slave mới kết nối, và **cập nhật thay đổi liên tục** (ongoing changes) để giữ cho slave luôn đồng bộ với master.

## 2. Các Hướng Tiếp Cận (Possible Solutions)

### Hướng 1: Sao chép dựa trên Câu lệnh (Statement-based Replication) - Cách của Redis

-   **Cấu trúc:** Đây là phương pháp được Redis sử dụng. Master không gửi *dữ liệu đã thay đổi* mà gửi chính *câu lệnh ghi* (write command) đã gây ra sự thay đổi đó cho các slave. Ví dụ, khi master thực thi `INCR mykey`, nó sẽ gửi nguyên văn câu lệnh `INCR mykey` tới tất cả các slave.
-   **Đồng bộ hóa ban đầu:** Master tạo một bản snapshot dữ liệu tại một thời điểm (dưới dạng file RDB), gửi cho slave. Slave tải bản snapshot này để có được trạng thái ban đầu. Sau đó, master bắt đầu gửi các lệnh ghi phát sinh trong và sau khi quá trình snapshot diễn ra.
-   **Ưu điểm:** Rất hiệu quả về mặt lưu lượng mạng vì các câu lệnh thường nhỏ hơn nhiều so với dữ liệu mà chúng thay đổi (ví dụ, `SADD myset "a-very-long-member"`). Logic đơn giản, dễ hiểu và debug.
-   **Nhược điểm:** Có thể gây ra sự không nhất quán nếu gặp các câu lệnh mang tính không xác định (non-deterministic). Ví dụ: `SADD myset (random-member)`. Tuy nhiên, Redis đã thiết kế các lệnh của mình để có thể sao chép an toàn, và các lệnh có yếu tố ngẫu nhiên thường có giải pháp thay thế.

### Hướng 2: Sao chép dựa trên Dữ liệu (Row-based Replication)

-   **Cấu trúc:** Phổ biến trong các hệ quản trị CSDL quan hệ. Thay vì gửi câu lệnh, master sẽ gửi đi *sự thay đổi cuối cùng* trên dữ liệu. Ví dụ, sau khi chạy `INCR mykey` (giả sử giá trị từ 10 thành 11), master sẽ gửi một thông điệp kiểu "key `mykey` bây giờ có giá trị là `11`".
-   **Ưu điểm:** Hoàn toàn miễn nhiễm với các lệnh không xác định. Luôn đảm bảo dữ liệu cuối cùng trên các slave là nhất quán về mặt giá trị.
-   **Nhược điểm:** Có thể tạo ra nhiều lưu lượng mạng hơn. Ví dụ, lệnh `LREM` xóa 1000 phần tử khỏi list chỉ là một câu lệnh ngắn, nhưng với row-based replication, nó có thể cần gửi thông tin về 1000 phần tử đã bị xóa, làm tăng vọt chi phí mạng.

### Hướng 3: Vận chuyển Write-Ahead Log (WAL Shipping)

-   **Cấu trúc:** Master ghi tất cả các thay đổi vào một file log (Write-Ahead Log) trước khi áp dụng chúng vào bộ nhớ. Các slave sẽ nhận được một bản sao của file log này và "phát lại" (replay) các thay đổi theo đúng thứ tự đã ghi.
-   **Ưu điểm:** Cực kỳ bền bỉ và mạnh mẽ, là nền tảng của các CSDL lớn như PostgreSQL. Cho phép thực hiện các kỹ thuật khôi phục tại một thời điểm (point-in-time recovery).
-   **Nhược điểm:** Rất phức tạp để triển khai cho một CSDL trong bộ nhớ như Redis. Nó làm tăng độ trễ của mọi thao tác ghi vì phải chờ ghi vào log trên đĩa trước. Đây không phải là cách Redis hoạt động.

## 3. Phân tích Trade-offs

| Hướng Tiếp Cận | Hiệu quả Mạng | Xử lý Lệnh không xác định | Độ Phức Tạp Triển Khai | Mức độ Tương thích Redis |
| :--- | :--- | :--- | :--- | :--- |
| 1. Statement-based | **Tuyệt vời** | **Yếu** (nhưng là vấn đề có thể chấp nhận trong Redis) | **Trung bình** | **Hoàn toàn** |
| 2. Row-based | **Trung bình** | **Mạnh** | **Cao** | **Không** |
| 3. WAL Shipping | **Tốt** | **Mạnh** | **Rất Cao** | **Không** |

## 4. Lựa chọn của SkylerRedis và Lý do

**Lựa chọn:** **Hướng 1: Sao chép dựa trên Câu lệnh (Statement-based Replication).**

**Lý do:**

1.  **Tương thích Redis:** Đây chính xác là cách Redis thực hiện sao chép. Toàn bộ quy trình handshake (`PSYNC ? -1`), truyền RDB, và sau đó là truyền dòng lệnh (command stream) là nền tảng của replication trong Redis. Các testcase của dự án đã xác nhận và yêu cầu quy trình này. File `09_replication_en.md` đã mô tả chi tiết các bước của phương pháp này, và lựa chọn này là để tuân thủ thiết kế đó.
2.  **Tối ưu cho CSDL In-Memory:** Với Redis, các thao tác ghi thường rất nhanh và các câu lệnh rất ngắn gọn. Việc gửi đi các câu lệnh này qua mạng là cách tiếp cận tự nhiên và hiệu quả nhất, tránh được chi phí (overhead) của việc phân tích và gửi đi các thay đổi dữ liệu phức tạp.
3.  **Cân bằng giữa Đơn giản và Hiệu quả:** Mặc dù có nhược điểm lý thuyết với các lệnh không xác định, trong thực tế, các lệnh ghi của Redis được thiết kế để có thể sao chép an toàn. Cách tiếp cận này cung cấp sự cân bằng tốt nhất giữa hiệu năng, sự đơn giản và đảm bảo tính nhất quán trong bối cảnh của Redis.
