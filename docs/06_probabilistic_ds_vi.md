# Tài liệu SkylerRedis

## 6. Cấu trúc dữ liệu xác suất (Probabilistic Data Structures)

Cấu trúc dữ liệu xác suất là một phần thiết yếu của các hệ thống hiệu năng cao hiện đại, cung cấp các câu trả lời gần đúng cho các truy vấn với bộ nhớ và/hoặc chi phí tính toán được tiết kiệm đáng kể. SkylerRedis triển khai hai cấu trúc như vậy: Bloom Filter và Count-Min Sketch.

### Bloom Filter

Bloom Filter là một cấu trúc dữ liệu xác suất hiệu quả về không gian, được sử dụng để kiểm tra xem một phần tử có phải là thành viên của một tập hợp hay không. Các trường hợp dương tính giả (false positive) có thể xảy ra, nhưng âm tính giả (false negative) thì không.

- **Trường hợp sử dụng**: Thường được dùng để nhanh chóng kiểm tra xem một mục *có thể* có trong một tập dữ liệu lớn hay không mà không cần truy vấn toàn bộ tập dữ liệu. Ví dụ bao gồm kiểm tra các URL đã xem (web crawlers), ngăn chặn đề xuất trùng lặp hoặc chặn các mật khẩu xấu đã biết.
- **Triển khai**: SkylerRedis triển khai Bloom Filters bằng cách sử dụng thư viện `github.com/willf/bloom`. Mỗi instance Bloom Filter được lưu trữ trong `memory.BloomFilters = make(map[string]*bloom.BloomFilter)`.
- **Lệnh**:
    - `BFADD key item`: Thêm một mục vào Bloom Filter có tên `key`. Trả về `1` nếu mục được thêm mới, `0` nếu nó có thể đã tồn tại. Nếu filter không tồn tại, nó được tạo với các tham số mặc định (dung lượng: 10.000, tỉ lệ dương tính giả: 0.01).
    - `BFEXISTS key item`: Kiểm tra xem một mục *có thể* tồn tại trong Bloom Filter có tên `key` hay không. Trả về `1` nếu mục có thể có trong filter, `0` nếu chắc chắn không có.

### Count-Min Sketch

Count-Min Sketch là một cấu trúc dữ liệu xác suất được sử dụng để ước tính tần suất của các sự kiện trong một luồng dữ liệu. Nó có thể ước tính gần đúng tần suất của bất kỳ phần tử nào trong một luồng với tỉ lệ lỗi nhỏ và độ tin cậy cao. Dương tính giả (ước tính quá cao) có thể xảy ra, nhưng âm tính giả (ước tính quá thấp) thì không.

- **Trường hợp sử dụng**: Lý tưởng cho các tình huống cần ước tính số lần xảy ra của các sự kiện hoặc mục cụ thể trong một tập dữ liệu lớn, động mà không cần lưu trữ tất cả các lần xuất hiện riêng lẻ. Ví dụ bao gồm đếm lượt xem trang web, số lượng gói mạng hoặc các chủ đề thịnh hành.
- **Triển khai**: SkylerRedis tự triển khai Count-Min Sketch từ đầu trong package `internal/memory` để đảm bảo kiểm soát hoàn toàn và tránh các vấn đề phụ thuộc bên ngoài. Mỗi instance Sketch được lưu trữ trong `memory.CountMinSketches = make(map[string]*memory.Sketch)`.
    - **Hashing**: Sử dụng kết hợp `hash/fnv` để tạo nhiều hàm hash.
    - **Bảng**: Một mảng 2 chiều (`[][]uint64`) lưu trữ các số đếm.
- **Lệnh**:
    - `CMSINCRBY key item increment`: Tăng số đếm ước tính cho `item` trong Count-Min Sketch có tên `key` bằng `increment`. Nếu sketch không tồn tại, nó được tạo với các tham số mặc định (depth: 5, width: 100,000). Trả về `OK`.
    - `CMSQUERY key item`: Trả về số đếm ước tính cho `item` trong Count-Min Sketch có tên `key`. Trả về `0` nếu sketch hoặc item không tồn tại.
