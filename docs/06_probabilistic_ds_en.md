# SkylerRedis Documentation

## 6. Probabilistic Data Structures

Probabilistic data structures are an essential part of modern high-performance systems, offering approximate answers to queries with significant memory and/or computational savings. SkylerRedis implements two such structures: the Bloom Filter and the Count-Min Sketch.

### Bloom Filter

A Bloom Filter is a space-efficient probabilistic data structure that is used to test whether an element is a member of a set. False positive matches are possible, but false negatives are not.

- **Use Case**: Often used to quickly check if an item *might* be in a large dataset without having to query the full dataset. Examples include checking for already-seen URLs (web crawlers), preventing duplicate recommendations, or blocking known bad passwords.
- **Implementation**: SkylerRedis implements Bloom Filters using the `github.com/willf/bloom` library. Each Bloom Filter instance is stored in `memory.BloomFilters = make(map[string]*bloom.BloomFilter)`.
- **Commands**:
    - `BFADD key item`: Adds an item to the Bloom Filter named `key`. Returns `1` if the item was newly added, `0` if it may have already existed. If the filter does not exist, it's created with default parameters (capacity: 10,000, false positive rate: 0.01).
    - `BFEXISTS key item`: Checks if an item *might* exist in the Bloom Filter named `key`. Returns `1` if the item might be in the filter, `0` if it is definitely not.

### Count-Min Sketch

A Count-Min Sketch is a probabilistic data structure used for frequency estimation of events in a data stream. It can approximate the frequency of any element in a stream with a small error rate and high confidence. False positives (overestimation) are possible, but false negatives (underestimation) are not.

- **Use Case**: Ideal for scenarios where you need to estimate how many times certain events or items have occurred in a large, dynamic dataset without storing all individual occurrences. Examples include counting website page views, network packet counts, or trending topics.
- **Implementation**: SkylerRedis implements its own Count-Min Sketch from scratch within the `internal/memory` package to ensure full control and avoid external dependency issues. Each Sketch instance is stored in `memory.CountMinSketches = make(map[string]*memory.Sketch)`.
    - **Hashing**: Uses a combination of `hash/fnv` to generate multiple hash functions.
    - **Table**: A 2D array (`[][]uint64`) stores the counts.
- **Commands**:
    - `CMSINCRBY key item increment`: Increments the estimated count for `item` in the Count-Min Sketch named `key` by `increment`. If the sketch does not exist, it's created with default parameters (depth: 5, width: 100,000). Returns `OK`.
    - `CMSQUERY key item`: Returns the estimated count for `item` in the Count-Min Sketch named `key`. Returns `0` if the sketch or item does not exist.
