# Tài liệu SkylerRedis

## 9. Replication (Master-Slave)

SkylerRedis triển khai một mô hình sao chép (replication) master-slave cơ bản để cho phép dự phòng dữ liệu và mở rộng quy mô đọc. Trong mô hình này, một instance master xử lý tất cả các hoạt động ghi, và một hoặc nhiều instance slave nhận một luồng các lệnh ghi này để sao chép lại bộ dữ liệu của master.

### Vai trò

#### Master
-   **Chế độ mặc định**: Một instance SkylerRedis khởi động ở chế độ master theo mặc định.
-   **Trách nhiệm**:
    -   Chấp nhận cả lệnh đọc và ghi từ client.
    -   Theo dõi các slave đã kết nối.
    -   Lan truyền tất cả các lệnh ghi thành công đến các slave của nó.

#### Slave
-   **Chế độ chỉ đọc**: Một instance slave thực chất là chỉ đọc từ góc độ của client. Nó không chấp nhận các lệnh ghi.
-   **Trách nhiệm**:
    -   Kết nối đến một instance master khi khởi động.
    -   Thực hiện một quy trình "bắt tay" (handshake) để thiết lập một luồng sao chép.
    -   Nhận và áp dụng các lệnh ghi từ master vào bộ dữ liệu cục bộ của mình.

### Cấu hình

#### Khởi động với vai trò Master
Chỉ cần chạy server SkylerRedis mà không có bất kỳ cờ nào liên quan đến replication:
```sh
docker run -d -p 6379:6379 --name skyler-master skyler-redis
```

#### Khởi động với vai trò Slave
Sử dụng cờ `--replicaof` để chỉ định host và port của master. Bạn cũng phải chạy slave trên một port khác.
```sh
# <master_ip> có thể là 'host.docker.internal' trên Docker Desktop
docker run -d -p 6380:6380 --name skyler-slave1 skyler-redis --port 6380 --replicaof <master_ip> 6379
```

### Quy trình Handshake

Khi một slave kết nối đến master, nó khởi tạo một giao thức handshake để đồng bộ hóa:

1.  **`PING`**: Slave gửi một lệnh `PING` đến master để xác minh rằng kết nối đang hoạt động và master đang phản hồi.
2.  **`REPLCONF listening-port <port>`**: Slave thông báo cho master biết nó đang lắng nghe trên port nào. Master đăng ký slave vào danh sách nội bộ các replica của mình.
3.  **`REPLCONF capa psync2`**: Slave thông báo cho master rằng nó có khả năng cho "PSYNC2", một phiên bản hiện đại hơn của giao thức sao chép.
4.  **`PSYNC ? -1`**: Slave yêu cầu đồng bộ hóa. `? -1` chỉ ra rằng đây là một slave mới yêu cầu đồng bộ hóa toàn bộ.
5.  **Phản hồi của Master (`FULLRESYNC`)**: Master phản hồi bằng `+FULLRESYNC <replication_id> <offset>`. Sau đó, nó gửi một bản sao đầy đủ (snapshot) dữ liệu của mình dưới dạng một file RDB.
6.  **Truyền RDB**: Slave nhận file RDB và tải nó vào bộ nhớ, ghi đè lên bất kỳ dữ liệu hiện có nào. Điều này đưa trạng thái của nó ngang hàng với trạng thái của master tại một thời điểm cụ thể.
7.  **Lan truyền lệnh**: Sau khi truyền RDB, master bắt đầu truyền trực tiếp tất cả các lệnh ghi tiếp theo đến slave, slave sẽ áp dụng chúng vào bộ dữ liệu cục bộ của mình.

### Lan truyền lệnh (Command Propagation)

-   Sau handshake ban đầu, bất kỳ lệnh ghi nào được xử lý bởi master (ví dụ: `SET`, `SADD`, `DEL`, v.v.) cũng được gửi đến tất cả các slave đã kết nối.
-   Các slave nhận các lệnh này và thực thi chúng trên các instance shard cục bộ của riêng mình, đảm bảo chúng luôn đồng bộ với master.
-   Các slave định kỳ xác nhận replication offset của chúng cho master thông qua các lệnh `REPLCONF ACK <offset>`, cho phép master theo dõi tiến trình sao chép.
-   **(Tình trạng hiện tại)**: Triển khai hiện tại chuyển tiếp trực tiếp các lệnh ghi. Nó chưa hỗ trợ bộ đệm lệnh cho các slave bị ngắt kết nối hoặc xử lý đồng bộ hóa một phần (PSYNC2 với một offset).

Kiến trúc master-slave này cung cấp một nền tảng vững chắc để xây dựng các hệ thống phức tạp hơn, có tính sẵn sàng cao với SkylerRedis.
