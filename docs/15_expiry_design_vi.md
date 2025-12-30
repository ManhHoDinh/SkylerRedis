# 15. Phân tích Thiết kế: Cơ chế Hết hạn Key (Key Expiry)

## 1. Mục tiêu

Triển khai một cơ chế để tự động xóa các key đã được gán "Thời gian sống" (Time-To-Live - TTL), ví dụ thông qua các lệnh `SET key value EX seconds` hoặc `EXPIRE key seconds`. Cơ chế này cần đảm bảo sự cân bằng giữa các yếu tố:
-   **Tính chính xác:** Key phải được coi là không tồn tại ngay khi hết hạn.
-   **Hiệu quả bộ nhớ:** Key hết hạn phải được dọn dẹp khỏi bộ nhớ để tránh lãng phí (memory leak).
-   **Hiệu năng:** Quá trình dọn dẹp không được làm ảnh hưởng đáng kể đến hiệu năng chung của server.

## 2. Các Hướng Tiếp Cận (Solutions)

### Hướng 1: Hết hạn Bị động / Lười biếng (Passive / Lazy Expiration)

-   **Cấu trúc:** Server không chủ động làm gì cả. Thay vào đó, nó chỉ kiểm tra khi cần.
-   **Luồng hoạt động:** Khi một lệnh (ví dụ: `GET`, `TTL`) truy cập vào một key, server sẽ kiểm tra xem key đó đã hết hạn hay chưa.
    -   Nếu đã hết hạn, server sẽ xóa key ngay lập tức và hành xử như thể key không tồn tại (ví dụ, trả về `(nil)` cho lệnh `GET`).
    -   Nếu chưa hết hạn, server thực hiện lệnh như bình thường.
-   **Ưu điểm:** Cực kỳ đơn giản để triển khai. Không tốn CPU cho việc quét các key không bao giờ được truy cập lại.
-   **Nhược điểm:** Các key đã hết hạn nhưng không bao giờ được truy cập lại sẽ tồn tại vĩnh viễn trong bộ nhớ. Điều này gây ra lãng phí bộ nhớ nghiêm trọng và không thể chấp nhận được trong môi trường production.

### Hướng 2: Hết hạn Chủ động (Active Expiration)

-   **Cấu trúc:** Một tiến trình nền (background process/goroutine) chạy định kỳ để dọn dẹp.
-   **Luồng hoạt động:** Cứ mỗi khoảng thời gian (ví dụ, 10 lần mỗi giây), tiến trình này sẽ thực hiện một vòng lặp dọn dẹp. Trong mỗi vòng lặp, nó sẽ:
    1.  Lấy một mẫu ngẫu nhiên các key có gán TTL từ không gian key (keyspace).
    2.  Kiểm tra và xóa những key trong mẫu đó đã hết hạn.
    3.  Vòng lặp sẽ dừng sớm nếu thời gian xử lý vượt quá một ngưỡng nhất định (để không chiếm hết CPU) hoặc nếu tỷ lệ key bị xóa thấp (cho thấy không còn nhiều key hết hạn cần dọn).
-   **Ưu điểm:** Chủ động giải phóng bộ nhớ, giảm thiểu đáng kể tình trạng lãng phí bộ nhớ từ các key không được truy cập.
-   **Nhược điểm:** Tốn một phần tài nguyên CPU cho việc quét nền. Có một độ trễ nhất định — một key có thể đã hết hạn nhưng chưa bị xóa cho đến kỳ quét tiếp theo.

### Hướng 3: Phương pháp Kết hợp (Hybrid Approach) - Cách của Redis

-   **Cấu trúc:** Kết hợp cả hai phương pháp trên để tận dụng ưu điểm và loại bỏ nhược điểm của từng phương pháp.
-   **Luồng hoạt động:**
    1.  **Lazy Expiration (Bị động):** Mỗi khi một key được truy cập, server *luôn* kiểm tra thời gian hết hạn của nó trước tiên (như Hướng 1). Điều này đảm bảo rằng không một client nào có thể đọc được dữ liệu đã cũ/hết hạn.
    2.  **Active Expiration (Chủ động):** Đồng thời, một tiến trình nền chạy định kỳ để quét và xóa các key đã hết hạn (như Hướng 2). Điều này giúp thu hồi bộ nhớ từ hàng triệu key không còn được ai truy cập.
-   **Ưu điểm:** Là giải pháp toàn diện và mạnh mẽ nhất. Nó vừa đảm bảo dữ liệu trả về cho client luôn chính xác, vừa chủ động dọn dẹp bộ nhớ bị lãng phí.
-   **Nhược điểm:** Phức tạp hơn trong việc triển khai so với việc chỉ chọn một trong hai cách.

## 3. Phân tích Trade-offs

| Hướng Tiếp Cận | Độ chính xác (Client View) | Hiệu quả Bộ nhớ | Tác động CPU | Độ Phức Tạp Triển Khai |
| :--- | :--- | :--- | :--- | :--- |
| 1. Lazy (Bị động) | **Cao** | **Rất Tệ** (Gây memory leak) | **Rất Thấp** | **Thấp** |
| 2. Active (Chủ động) | **Thấp** (Có thể đọc key cũ) | **Tốt** | **Thấp-Trung bình** | **Trung bình** |
| 3. Hybrid (Kết hợp) | **Cao** | **Tuyệt vời** | **Thấp-Trung bình** | **Cao** |

## 4. Lựa chọn của SkylerRedis và Lý do

**Lựa chọn:** **Hướng 3: Phương pháp Kết hợp (Hybrid Approach).**

**Lý do:**

1.  **Tương thích và Hiệu quả như Redis:** Đây chính là chiến lược mà Redis sử dụng. Để xây dựng một bản sao Redis đúng nghĩa, việc áp dụng cả hai cơ chế là bắt buộc. Testcase về "Lazy Expiry" xác nhận yêu cầu của Hướng 1, và sự tồn tại của file `ActiveExpiration.md` cho thấy sự cần thiết của Hướng 2.

2.  **Ngăn chặn Lãng phí Bộ nhớ (Memory Leak):** Chỉ sử dụng Lazy Expiration (Hướng 1) là không đủ. Nó sẽ dẫn đến việc các key hết hạn chiếm dụng bộ nhớ vĩnh viễn nếu chúng không được gọi lại. Cơ chế Active Expiration là **bắt buộc** để giải quyết vấn đề này, đảm bảo server có thể hoạt động ổn định trong thời gian dài.

3.  **Đảm bảo Tính Đúng đắn của Dữ liệu:** Chỉ sử dụng Active Expiration (Hướng 2) cũng không đủ. Do có một khoảng trễ giữa thời điểm key hết hạn và thời điểm nó bị xóa, một client có thể đọc được dữ liệu đã hết hạn trong khoảng trễ đó. Cơ chế Lazy Expiration bịt lại lỗ hổng này bằng cách kiểm tra TTL ngay tại thời điểm truy cập.

**Kết luận:** Sự kết hợp này là giải pháp duy nhất giải quyết được cả hai vấne đề cốt lõi của bài toán hết hạn key: **tính đúng đắn của dữ liệu trả về cho client** và **quản lý bộ nhớ hiệu quả**.
