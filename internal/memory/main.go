package memory

import (
	"SkylerRedis/internal/entity"
	"net"
	"sync"
	"hash/fnv" // Added for sharding hash function
)

var (
	// Global state not specific to a single shard.
	Blockings = make(map[string][]BlockingRequest)
	Queue     = make(map[net.Conn][][]string, 0)
	IsMulti   = make(map[net.Conn]bool)
	Stream    = map[string]StreamEntry{}
	Slaves    = make([]entity.BaseServer, 0)
	// OffSet    = 0 // Removed as offset is now part of Master/Slave struct
	StreamIDs = make([]string, 0)
	Mu        = sync.Mutex{} // Global mutex for global state (Blockings, Queue, etc.)

	// Shard-specific data will now be managed via Shards map.
	Shards map[int]*Shard // Map to hold multiple Shard instances.
	ShardsMu sync.RWMutex // Mutex to protect access to the Shards map itself.

	numShards int // Store the number of shards for GetShardForKey
)

// Global helper for initializing all shards.
func InitShards(n int, maxMemory int) {
	ShardsMu.Lock()
	defer ShardsMu.Unlock()

	numShards = n // Store the number of shards
	Shards = make(map[int]*Shard, numShards) // Initialize the map with capacity
	for i := 0; i < numShards; i++ {
		Shards[i] = NewShard(maxMemory)
	}
}

// GetShardForKey returns the appropriate Shard instance for a given key.
// It uses FNV-1a hash to determine the shard ID.
func GetShardForKey(key string) *Shard {
	ShardsMu.RLock()
	defer ShardsMu.RUnlock()

	if numShards == 0 { // Should not happen if InitShards is called
		return nil 
	}
	h := fnv.New32a()
	h.Write([]byte(key))
	shardID := int(h.Sum32() % uint32(numShards))
	return Shards[shardID] 
}
