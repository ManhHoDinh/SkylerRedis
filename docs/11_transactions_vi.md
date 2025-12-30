# 11. Phân tích Thiết kế: Transactions (MULTI / EXEC)

## 1. Mục tiêu

Triển khai các giao dịch nguyên tử (atomic transactions) trong SkylerRedis, cho phép client nhóm nhiều lệnh lại để thực thi như một thao tác đơn lẻ, không thể bị chia cắt. Các lệnh chính bao gồm `MULTI`, `EXEC`, và `DISCARD`, với mục tiêu đảm bảo tính tương thích hoàn toàn với Redis.

## 2. Các Hướng Tiếp Cận (Possible Solutions)

Khi triển khai hệ thống giao dịch, có ba hướng tiếp cận chính với các ưu và nhược điểm riêng:

### Hướng 1: Hàng đợi lệnh ở Client (Client-side Queueing)

-   **Luồng hoạt động:** Client sẽ tự quản lý một hàng đợi lệnh sau khi người dùng gọi `MULTI`. Khi nhận lệnh `EXEC`, client sẽ đóng gói toàn bộ các lệnh này và gửi chúng lên server trong một yêu cầu mạng duy nhất (thường gọi là "batching" hoặc "pipelining").
-   **Phía server:** Server chỉ nhận và thực thi một chuỗi lệnh nối tiếp nhau mà không hề biết chúng thuộc về một giao dịch.
-   **Triển khai:** Logic xử lý giao dịch nằm hoàn toàn ở thư viện client (vd: `redis-py`, `node-redis`). Server gần như không cần thay đổi.

### Hướng 2: Hàng đợi lệnh ở Server & Quản lý Trạng thái (Server-side State & Queueing)

-   **Luồng hoạt động:**
    1.  Khi client gửi `MULTI`, server sẽ đánh dấu kết nối (connection) đó đang ở trạng thái "trong giao dịch" (`in-transaction`).
    2.  Các lệnh tiếp theo từ client này sẽ không được thực thi ngay lập tức. Thay vào đó, chúng được thêm vào một hàng đợi riêng thuộc về kết nối đó. Với mỗi lệnh, server phản hồi `+QUEUED`.
    3.  Khi client gửi `EXEC`, server sẽ duyệt qua hàng đợi lệnh và thực thi chúng một cách tuần tự và nguyên tử. Trong suốt quá trình này, không có bất kỳ lệnh nào từ các client khác được phép xen vào.
    4.  Nếu client gửi `DISCARD`, server sẽ xóa hàng đợi và thoát khỏi trạng thái giao dịch.
-   **Triển khai:** Đây là cách tiếp cận chuẩn của Redis. Server phải quản lý trạng thái và hàng đợi cho từng kết nối.

### Hướng 3: Khóa Toàn cục trên Server (Global Server Lock)

-   **Luồng hoạt động:** Khi client gửi `MULTI`, server sẽ kích hoạt một "khóa" toàn cục, ngăn tất cả các client khác thực thi lệnh. Client trong giao dịch gửi lệnh nào, server thực thi ngay lệnh đó. Khi nhận được `EXEC`, server sẽ mở khóa.
-   **Triển khai:** Ý tưởng đơn giản, nhưng ảnh hưởng nghiêm trọng đến hiệu năng vì nó chặn hoàn toàn các client khác.

## 3. Phân tích Trade-offs

| Hướng Tiếp Cận | Hiệu Năng | Đảm Bảo Tính Nguyên Tử | Độ Phức Tạp | Sử Dụng Bộ Nhớ (Server) |
| :--- | :--- | :--- | :--- | :--- |
| **1. Client Queue** | **Cao** (Giảm round-trip mạng) | **Yếu** (Server không đảm bảo, client khác có thể sửa key giữa các lệnh) | **Trung bình** (Đẩy sự phức tạp cho các thư viện client) | **Thấp** (Server không cần trạng thái) |
| **2. Server Queue** | **Tốt** (Nhanh và có kiểm soát) | **Mạnh** (Server đảm bảo 100% tính nguyên tử khi `EXEC`) | **Trung bình** (Server cần quản lý trạng thái/hàng đợi mỗi kết nối) | **Trung bình** (Lưu hàng đợi lệnh cho mỗi giao dịch) |
| **3. Global Lock** | **Rất Tệ** (Giết chết hiệu năng đồng thời) | **Mạnh** | **Thấp** (Cơ chế lock đơn giản) | **Thấp** |

## 4. Lựa chọn của SkylerRedis và Lý do

**Lựa chọn:** **Hướng 2: Hàng đợi lệnh ở Server & Quản lý Trạng thái.**

**Lý do:**

1.  **Tương thích Redis (Redis Compatibility):** Đây là cách Redis hoạt động. Mục tiêu của SkylerRedis là xây dựng một bản sao tương thích của Redis, vì vậy việc tuân thủ đúng ngữ nghĩa và hành vi của nó là yêu cầu tiên quyết. Các testcase trong dự án (với phản hồi `+QUEUED`) đã khẳng định hành vi này là bắt buộc.

2.  **Đảm bảo Tính Nguyên tử:** Hướng tiếp cận này cung cấp một lời hứa "tất cả hoặc không có gì" (all-or-nothing) ở cấp độ server. Đây là giá trị cốt lõi của một giao dịch. Hướng 1 không thể đảm bảo điều này vì một client khác có thể xen vào và thay đổi dữ liệu giữa các lệnh được gửi đi.

3.  **Hiệu năng Cân bằng (Balanced Performance):** Mặc dù Khóa Toàn cục (Hướng 3) dễ cài đặt hơn, hiệu năng của nó là không thể chấp nhận được cho một máy chủ phục vụ nhiều kết nối đồng thời. Mô hình hàng đợi phía server cung cấp tính nguyên tử cho *giai đoạn thực thi* (`EXEC`) mà không chặn các client khác trong *giai đoạn xếp lệnh* (queuing phase). Điều này tạo ra sự cân bằng hợp lý và cần thiết giữa an toàn dữ liệu và tính đồng thời (concurrency), giữ cho "vùng tranh chấp" (critical section) ngắn nhất có thể.
