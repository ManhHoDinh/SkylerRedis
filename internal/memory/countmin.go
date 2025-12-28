package memory

import (
	"hash/fnv"
	"math"
	"strconv"
)

// Sketch is a Count-Min Sketch data structure.
type Sketch struct {
	table [][]uint64
	depth uint64 // Number of hash functions
	width uint64 // Size of each hash array
	seed  uint64 // Base seed for hash functions
}

// NewCountMinSketch creates a new Count-Min Sketch.
// depth: number of hash functions (rows)
// width: size of each hash table (columns)
func NewCountMinSketch(depth, width uint64) *Sketch {
	table := make([][]uint64, depth)
	for i := uint64(0); i < depth; i++ {
		table[i] = make([]uint64, width)
	}
	return &Sketch{
		table: table,
		depth: depth,
		width: width,
		seed:  0x9747B28C, // A fixed seed
	}
}

// hash returns the i-th hash value for a given item.
// It uses a technique to generate k independent hash functions from two universal hash functions.
func (s *Sketch) hash(item []byte, i uint64) uint64 {
	h := fnv.New64a()
	h.Write(item)
	v := h.Sum64()

	// Generate a second hash based on the first hash and the current row index.
	// This helps create multiple "independent" hash functions.
	h2 := fnv.New64a()
	h2.Write([]byte(strconv.FormatUint(v, 10) + strconv.FormatUint(s.seed+i, 10)))
	v2 := h2.Sum64()

	// Combine two hashes
	// This generates k hash values for the item.
	return (v + i*v2) % s.width
}

// Add increments the count for an item by the given value.
func (s *Sketch) Add(item []byte, count uint64) {
	for i := uint64(0); i < s.depth; i++ {
		hashVal := s.hash(item, i)
		s.table[i][hashVal] += count
	}
}

// Count returns the estimated count for an item.
func (s *Sketch) Count(item []byte) uint64 {
	minCount := uint64(math.MaxUint64) // Initialize with max possible uint64
	for i := uint64(0); i < s.depth; i++ {
		hashVal := s.hash(item, i)
		currentCount := s.table[i][hashVal]
		if currentCount < minCount {
			minCount = currentCount
		}
	}
	return minCount
}
