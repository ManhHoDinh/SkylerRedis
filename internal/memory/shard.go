package memory

import (
	"sync"

	"github.com/willf/bloom"
)

// Shard holds all the data and state for a single shard, making it self-contained.
// This is crucial for implementing a thread-per-shard architecture.
type Shard struct {
	Store            map[string]*Entry
	Sets             map[string]map[string]struct{}
	BloomFilters     map[string]*bloom.BloomFilter
	CountMinSketches map[string]*Sketch
	RPush            map[string][]string // For lists
	Stream           map[string]StreamEntry // For streams
	StreamIDs        []string               // For stream IDs
	LruClock         uint64
	MaxMemory        int
	Mu               sync.Mutex
}

// NewShard creates and initializes a new Shard instance.
func NewShard(maxMemory int) *Shard {
	return &Shard{
		Store:            make(map[string]*Entry),
		Sets:             make(map[string]map[string]struct{}),
		BloomFilters:     make(map[string]*bloom.BloomFilter),
		CountMinSketches: make(map[string]*Sketch),
		RPush:            make(map[string][]string), // Initialize RPush
		Stream:           make(map[string]StreamEntry), // Initialize Stream
		StreamIDs:        make([]string, 0),        // Initialize StreamIDs
		LruClock:         0,
		MaxMemory:        maxMemory,
		Mu:               sync.Mutex{},
	}
}
