# SkylerRedis Documentation

## 10. Core Data Structures: Lists and Sets

Beyond the basic Key-Value store, SkylerRedis implements two fundamental Redis data structures: Lists and Sets. These structures provide more complex ways to organize and interact with data. All data structures are managed on a per-shard basis, ensuring that the system scales effectively.

### Lists

A List in SkylerRedis is a sequence of strings, sorted by insertion order. It is implemented as a simple Go slice (`[]string`) associated with a key. This provides `O(1)` time complexity for `RPUSH` and `O(n)` for `LPUSH`.

-   **Storage**: List data is stored in the `RPush` map within each `Shard`:
    ```go
    // a.k.a. Shard.RPush in the code
    map[string][]string
    ```
-   **Commands**:
    -   `LPUSH key element1 [element2 ...]`: Inserts one or more elements at the beginning (head) of the list. Returns the new length of the list.
    -   `RPUSH key element1 [element2 ...]`: Inserts one or more elements at the end (tail) of the list. Returns the new length of the list.
    -   `LPOP key [count]`: Removes and returns one or more elements from the beginning of the list.
    -   `LRANGE key start stop`: Returns the specified range of elements from the list.
    -   `LLEN key`: Returns the length of the list.

### Sets

A Set is an unordered collection of unique strings. Adding the same element multiple times will result in only one copy being stored.

-   **Storage**: Sets are implemented using a Go `map[string]struct{}`, which is a highly efficient and idiomatic way to represent a set. The `struct{}` value consumes zero memory. Set data is stored in the `Sets` map within each `Shard`:
    ```go
    // a.k.a. Shard.Sets in the code
    map[string]map[string]struct{}
    ```
-   **Commands**:
    -   `SADD key member1 [member2 ...]`: Adds one or more members to the set. Returns the number of members that were newly added (i.e., did not already exist in the set).
    -   `SREM key member1 [member2 ...]`: Removes one or more members from the set. Returns the number of members that were successfully removed.
    -   `SISMEMBER key member`: Returns `1` if the member exists in the set, and `0` otherwise.
    -   `SCARD key`: Returns the number of members in the set (its cardinality).
    -   `SMEMBERS key`: Returns all members of the set as an array. Note that the order is not guaranteed.
