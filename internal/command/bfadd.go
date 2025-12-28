package command

import (
	"SkylerRedis/internal/memory"
	"SkylerRedis/internal/utils"
	"net"

	"github.com/willf/bloom"
)

// Bfadd implements the BF.ADD command.
type Bfadd struct{}

const (
	// Default Bloom Filter capacity
	defaultBloomN = 10000
	// Default Bloom Filter false positive rate
	defaultBloomFp = 0.01
)

func (cmd Bfadd) Handle(conn net.Conn, args []string, isMaster bool, masterReplID string, masterReplOffset int, connectedSlaves int, shard *memory.Shard) {
	if len(args) != 3 {
		utils.WriteError(conn, "ERR wrong number of arguments for 'bf.add' command")
		return
	}

	key := args[1]
	item := args[2]

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	filter, ok := shard.BloomFilters[key]
	if !ok {
		// Create a new filter with default parameters if it doesn't exist.
		filter = bloom.NewWithEstimates(defaultBloomN, defaultBloomFp)
		shard.BloomFilters[key] = filter
	}

	// TestAndAdd returns true if the item was likely already in the filter.
	// Redis's BF.ADD returns 1 if the item was added, 0 if it may have already existed.
	// So we need to invert the boolean result.
	if filter.TestAndAddString(item) {
		utils.WriteInteger(conn, 0) // Item may have already been in the filter
	} else {
		utils.WriteInteger(conn, 1) // Item was definitely not in the filter
	}
}
