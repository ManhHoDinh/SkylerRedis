# 12. Phân tích Thiết kế: Redis Streams

## 1. Mục tiêu

Triển khai cấu trúc dữ liệu Redis Stream, một dạng "append-only log" (log chỉ ghi tiếp vào cuối). Các tính năng cốt lõi cần có bao gồm:
-   `XADD`: Thêm một entry mới vào stream.
-   `XRANGE`: Truy vấn một khoảng entry dựa trên ID.
-   `XREAD`: Đọc dữ liệu từ một hoặc nhiều stream, bắt đầu từ một ID cụ thể, với khả năng `BLOCK` (chặn) để chờ dữ liệu mới.

Định dạng ID `ms-seq` (thời gian millisecond - số thứ tự) là một phần cực kỳ quan trọng và ảnh hưởng trực tiếp đến thiết kế.

## 2. Các Hướng Tiếp Cận (Possible Solutions)

### Hướng 1: Dùng List (Mảng động)

-   **Cấu trúc:** Coi một Stream như một List (mảng) chứa các entry. Mỗi entry là một struct/object chứa ID và dữ liệu.
-   **`XADD`:** Thêm entry mới vào cuối mảng (`append`). Việc tạo ID tự động (`*`) yêu cầu lấy thời gian hiện tại và duyệt ngược để tìm sequence number cuối cùng của cùng millisecond, hoặc mặc định là 0 nếu millisecond đã thay đổi.
-   **`XRANGE`/`XREAD`:** Yêu cầu duyệt toàn bộ mảng (linear scan) và lọc ra các entry có ID nằm trong khoảng cho trước. Thao tác này có độ phức tạp **O(N)**, với N là tổng số entry trong stream, nên rất chậm.
-   **Đánh giá:** Rất đơn giản để triển khai ban đầu nhưng hoàn toàn không hiệu quả khi stream có lượng dữ liệu lớn.

### Hướng 2: Dùng Cây nhị phân tự cân bằng (B-Tree/Red-Black Tree) hoặc Skip List

-   **Cấu trúc:** Lưu các entry của stream trong một cấu trúc dữ liệu được sắp xếp như B-Tree hoặc Skip List, với ID của entry được dùng làm khóa (key) để so sánh và sắp xếp.
-   **`XADD`:** Thêm một entry mới vào cây/skip list. Thao tác này có độ phức tạp **O(log N)**. Việc tạo ID tự động cũng hiệu quả vì có thể tìm thấy entry cuối cùng (lớn nhất) trong O(log N) hoặc O(1).
-   **`XRANGE`/`XREAD`:** Tìm kiếm theo khoảng (range query) trên các cấu trúc này rất hiệu quả, thường có độ phức tạp **O(log N + M)**, với M là số lượng entry trả về.
-   **Đánh giá:** Phức tạp hơn đáng kể so với dùng List, nhưng mang lại hiệu năng vượt trội và có khả năng mở rộng (scalable). Đây là một lựa chọn cân bằng.

### Hướng 3: Dùng Cây cơ số (Radix Tree) - Cách của Redis

-   **Cấu trúc:** Redis sử dụng một cấu trúc dữ liệu được tối ưu hóa cao gọi là Radix Tree để lưu các entry. Mỗi node trong cây đại diện cho một phần của ID (ví dụ, 8 byte của timestamp, 8 byte của sequence). Các entry được lưu ở các node lá.
-   **Ưu điểm:** Cấu trúc này cực kỳ hiệu quả về bộ nhớ vì các tiền tố (prefix) chung của các ID được chia sẻ giữa các node. Các thao tác thêm và truy vấn theo khoảng cũng rất nhanh.
-   **`XREAD BLOCK`:** Việc chặn và chờ dữ liệu mới có thể được triển khai bằng cách duy trì một danh sách các client đang chờ trên mỗi stream. Khi có lệnh `XADD` mới, server sẽ duyệt qua danh sách này, kiểm tra xem entry mới có thỏa mãn điều kiện chờ của client nào không và "đánh thức" client đó.
-   **Đánh giá:** Đây là hướng tiếp cận phức tạp nhất, đòi hỏi kiến thức sâu về cấu trúc dữ liệu. Tuy nhiên, nó mang lại hiệu năng và hiệu quả bộ nhớ tốt nhất, đúng như Redis gốc.

## 3. Phân tích Trade-offs

| Hướng Tiếp Cận | Hiệu Năng (XADD/XRANGE) | Hiệu Quả Bộ Nhớ | Độ Phức Tạp Triển Khai | Mức độ Tương thích Redis |
| :--- | :--- | :--- | :--- | :--- |
| 1. List/Array | **Rất Tệ** (O(N) cho tìm kiếm) | **Trung bình** | **Rất Thấp** | **Thấp** (Không thể scale) |
| 2. B-Tree/Skip List | **Tốt** (O(log N)) | **Tốt** | **Cao** | **Trung bình** (Hành vi đúng, cấu trúc bên trong khác) |
| 3. Radix Tree | **Tuyệt vời** (Tối ưu cho ID) | **Tuyệt vời** (Chia sẻ prefix) | **Rất Cao** | **Cao** (Đây là cách Redis làm) |

## 4. Lựa chọn của SkylerRedis và Lý do

**Lựa chọn (Giả định):** **Hướng 2: Dùng B-Tree hoặc Skip List.**

**Lý do:**

1.  **Cân bằng giữa Hiệu năng và Độ phức tạp:** Mặc dù Radix Tree (Hướng 3) là cấu trúc dữ liệu thực tế của Redis, việc tự triển khai nó từ đầu là một nhiệm vụ cực kỳ phức tạp và tốn thời gian. Nó có thể không phù hợp với mục tiêu học tập và xây dựng nhanh của một dự án như SkylerRedis.
2.  **Đáp ứng Yêu cầu về Hiệu năng:** Một B-Tree hoặc Skip List đã cung cấp hiệu năng **O(log N)** cho các hoạt động chính. Tốc độ này đủ nhanh để vượt qua tất cả các testcase về hiệu năng và quan trọng hơn là có khả năng mở rộng tốt hơn rất nhiều so với giải pháp dùng List (Hướng 1).
3.  **Tập trung vào Logic Nghiệp vụ:** Lựa chọn này cho phép người phát triển tập trung vào việc triển khai đúng *logic* của Streams (cơ chế tạo ID, blocking, consumer groups) mà không bị sa lầy vào việc tối ưu hóa cấu trúc dữ liệu ở mức thấp nhất. Có thể tận dụng các thư viện B-Tree có sẵn.
4.  **Triển khai Blocking:** Cơ chế blocking (`XREAD BLOCK`) có thể được xây dựng tương đối độc lập với cấu trúc dữ liệu nền. Một map dạng `map[streamName][]blockingClient` để theo dõi các client đang chờ là một khởi đầu đủ tốt và hiệu quả.
