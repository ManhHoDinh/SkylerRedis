# 13. Phân tích Thiết kế: Persistence (RDB)

## 1. Mục tiêu

Triển khai cơ chế lưu trữ bền bỉ (persistence) cho SkylerRedis bằng cách tạo và đọc các file RDB (Redis Database). Mục tiêu chính là lưu lại toàn bộ trạng thái của database trong bộ nhớ (bao gồm key, value, và thời gian hết hạn) vào một file nhị phân nhỏ gọn khi máy chủ tắt, và khôi phục lại trạng thái đó từ file khi máy chủ khởi động lại.

## 2. Các Hướng Tiếp Cận cho Định dạng File

### Hướng 1: Dùng Định dạng Text (ví dụ: JSON, CSV)

-   **Cấu trúc:** Lưu mỗi cặp key-value như một object trong file JSON hoặc một dòng trong file CSV.
    ```json
    [
      {"key": "mykey", "value": "myvalue", "expires_at": 1672531199000},
      {"key": "anotherkey", "value": "anothervalue", "expires_at": null}
    ]
    ```
-   **Ưu điểm:** Con người có thể đọc và sửa file dễ dàng. Rất đơn giản để tạo và phân tích (parse) bằng các thư viện chuẩn có sẵn trong mọi ngôn ngữ.
-   **Nhược điểm:** Kích thước file rất lớn, gây tốn dung lượng đĩa và làm chậm tốc độ I/O. Quá trình phân tích text chậm hơn nhiều so với phân tích nhị phân. Hoàn toàn không tương thích với Redis.

### Hướng 2: Dùng Định dạng Nhị phân Tùy chỉnh, Đơn giản

-   **Cấu trúc:** Tự định nghĩa một cấu trúc nhị phân đơn giản. Ví dụ, mỗi record có thể có dạng: `[4 byte độ dài key][key][4 byte độ dài value][value][8 byte expiry timestamp]`.
-   **Ưu điểm:** Kích thước file nhỏ gọn hơn nhiều so với định dạng text. Tốc độ đọc và ghi nhanh hơn đáng kể vì không cần xử lý text.
-   **Nhược điểm:** Vẫn không tương thích với Redis. Phải tự viết code để chuyển đổi dữ liệu (serialize/deserialize), dễ phát sinh lỗi và khó bảo trì nếu cấu trúc thay đổi.

### Hướng 3: Tuân thủ Định dạng RDB của Redis

-   **Cấu trúc:** Triển khai chính xác định dạng file RDB mà Redis sử dụng. Đây là một định dạng nhị phân phức tạp và được tối ưu cao.
    -   **Header:** File luôn bắt đầu bằng một chuỗi magic `REDIS` và số phiên bản, ví dụ: `REDIS0009`.
    -   **Opcodes (Mã lệnh):** Sử dụng các byte đơn lẻ để đánh dấu các loại thông tin hoặc cấu trúc dữ liệu khác nhau. Ví dụ từ testcase:
        -   `0xFE`: Mã lệnh chọn Database (SELECTDB).
        -   `0xFD`: Mã lệnh chứa thời gian hết hạn theo giây (expiry in seconds).
        -   `0xFC`: Mã lệnh chứa thời gian hết hạn theo millisecond (expiry in milliseconds).
        -   `0x00`: Mã lệnh báo hiệu bắt đầu một cặp key-value kiểu String.
        -   `0xFF`: Mã lệnh kết thúc file (End Of File).
    -   **Mã hóa Độ dài (Length-Encoding):** Sử dụng một phương pháp mã hóa đặc biệt (length-prefixed encoding) để lưu độ dài của chuỗi một cách cực kỳ hiệu quả, giúp tiết kiệm không gian.
-   **Ưu điểm:** Tương thích 100% với Redis, cho phép SkylerRedis đọc file `.rdb` do Redis tạo ra và ngược lại. Kích thước file cực kỳ nhỏ gọn và tốc độ xử lý rất nhanh.
-   **Nhược điểm:** Phức tạp nhất để triển khai. Đòi hỏi phải đọc hiểu sâu đặc tả kỹ thuật của định dạng RDB và xử lý cẩn thận ở mức độ từng byte.

## 3. Phân tích Trade-offs

| Hướng Tiếp Cận | Kích thước File | Tốc độ Đọc/Ghi | Độ Phức Tạp Triển Khai | Mức độ Tương thích Redis |
| :--- | :--- | :--- | :--- | :--- |
| 1. JSON/Text | **Rất Lớn** | **Chậm** | **Rất Thấp** | **Không** |
| 2. Nhị phân Đơn giản | **Trung bình** | **Nhanh** | **Trung bình** | **Không** |
| 3. Định dạng RDB | **Rất Nhỏ** | **Rất Nhanh** | **Rất Cao** | **Hoàn toàn** |

## 4. Lựa chọn của SkylerRedis và Lý do

**Lựa chọn:** **Hướng 3: Tuân thủ Định dạng RDB của Redis.**

**Lý do:**

1.  **Yêu cầu về Tính Tương thích:** Mục tiêu cốt lõi của SkylerRedis là mô phỏng lại Redis. Khả năng đọc và ghi file RDB chuẩn là một yêu cầu không thể thiếu, được nhấn mạnh qua các testcase (kiểm tra header `REDIS0009`, các opcode...). Điều này cho phép SkylerRedis có thể thay thế hoặc hoạt động song song trong một hệ sinh thái Redis thực tế.
2.  **Hiệu quả Tối đa:** Định dạng RDB được thiết kế để nhỏ gọn và nhanh chóng một cách tối đa. Đối với một database trong bộ nhớ, việc giảm thiểu thời gian I/O khi khởi động (đọc file) và tắt máy (ghi file) là cực kỳ quan trọng để đảm bảo thời gian downtime thấp và trải nghiệm người dùng tốt.
3.  **Tính Hoàn thiện và Mở rộng:** Định dạng RDB đã định nghĩa sẵn cách lưu trữ tất cả các cấu trúc dữ liệu phức tạp của Redis (Lists, Hashes, Sets, ZSets...). Bằng cách tuân thủ định dạng này ngay từ đầu, SkylerRedis đã mở đường cho việc hỗ trợ lưu trữ bền bỉ các kiểu dữ liệu khác trong tương lai một cách nhất quán. Mặc dù độ phức tạp khi triển khai là rất cao, đây là một sự đầu tư xứng đáng cho một bản sao Redis đúng nghĩa.
