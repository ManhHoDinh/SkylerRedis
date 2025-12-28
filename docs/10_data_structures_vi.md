# Tài liệu SkylerRedis

## 10. Các Cấu trúc Dữ liệu Cốt lõi: Lists và Sets

Ngoài kho lưu trữ Key-Value cơ bản, SkylerRedis còn triển khai hai cấu trúc dữ liệu nền tảng của Redis: Lists (Danh sách) và Sets (Tập hợp). Các cấu trúc này cung cấp những cách phức tạp hơn để tổ chức và tương tác với dữ liệu. Tất cả các cấu trúc dữ liệu đều được quản lý trên cơ sở từng shard, đảm bảo hệ thống có khả năng mở rộng hiệu quả.

### Lists (Danh sách)

Một List trong SkylerRedis là một chuỗi các string, được sắp xếp theo thứ tự chèn. Nó được triển khai như một Go slice (`[]string`) đơn giản được liên kết với một key. Điều này cung cấp độ phức tạp thời gian `O(1)` cho `RPUSH` và `O(n)` cho `LPUSH`.

-   **Lưu trữ**: Dữ liệu List được lưu trữ trong map `RPush` bên trong mỗi `Shard`:
    ```go
    // tương đương Shard.RPush trong mã nguồn
    map[string][]string
    ```
-   **Các lệnh**:
    -   `LPUSH key element1 [element2 ...]`: Chèn một hoặc nhiều phần tử vào đầu danh sách. Trả về độ dài mới của danh sách.
    -   `RPUSH key element1 [element2 ...]`: Chèn một hoặc nhiều phần tử vào cuối danh sách. Trả về độ dài mới của danh sách.
    -   `LPOP key [count]`: Xóa và trả về một hoặc nhiều phần tử từ đầu danh sách.
    -   `LRANGE key start stop`: Trả về một khoảng các phần tử được chỉ định từ danh sách.
    -   `LLEN key`: Trả về độ dài của danh sách.

### Sets (Tập hợp)

Một Set là một tập hợp không có thứ tự của các string duy nhất. Việc thêm cùng một phần tử nhiều lần sẽ chỉ dẫn đến một bản sao duy nhất được lưu trữ.

-   **Lưu trữ**: Sets được triển khai bằng cách sử dụng `map[string]struct{}` trong Go, đây là một cách hiệu quả và phổ biến để biểu diễn một tập hợp. Giá trị `struct{}` không tiêu tốn bộ nhớ. Dữ liệu Set được lưu trữ trong map `Sets` bên trong mỗi `Shard`:
    ```go
    // tương đương Shard.Sets trong mã nguồn
    map[string]map[string]struct{}
    ```
-   **Các lệnh**:
    -   `SADD key member1 [member2 ...]`: Thêm một hoặc nhiều thành viên vào tập hợp. Trả về số lượng thành viên được thêm mới (tức là chưa tồn tại trong tập hợp).
    -   `SREM key member1 [member2 ...]`: Xóa một hoặc nhiều thành viên khỏi tập hợp. Trả về số lượng thành viên đã được xóa thành công.
    -   `SISMEMBER key member`: Trả về `1` nếu thành viên tồn tại trong tập hợp, và `0` nếu không.
    -   `SCARD key`: Trả về số lượng thành viên trong tập hợp (lực lượng của nó).
    -   `SMEMBERS key`: Trả về tất cả các thành viên của tập hợp dưới dạng một mảng. Lưu ý rằng thứ tự không được đảm bảo.
