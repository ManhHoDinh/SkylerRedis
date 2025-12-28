package memory

import (
	"math"
	"time"
)

const (
	evictionSampleSize  = 20
	lruEvictionSample   = 10 // Number of keys to sample for LRU
	evictionTargetRatio = 0.95 // Evict until current size is 95% of maxmemory
)

// EvictRandomKeys implements the active key expiration algorithm.
// It samples a small number of keys and removes any that have expired.
func (s *Shard) EvictRandomKeys() { // Added shard receiver
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if len(s.Store) == 0 {
		return
	}

	expiredKeys := make([]string, 0, evictionSampleSize)
	count := 0
	for key := range s.Store {
		if count >= evictionSampleSize {
			break
		}
		expiredKeys = append(expiredKeys, key)
		count++
	}

	for _, key := range expiredKeys {
		entry, ok := s.Store[key]
		if !ok {
			continue 
		}

		if !entry.ExpiryTime.IsZero() && time.Now().After(entry.ExpiryTime) {
			delete(s.Store, key)
		}
	}
}

// EvictKeysByLRU implements the approximate LRU eviction policy.
// It samples keys and removes the one with the lowest LRU value until
// the store size is below the configured MaxMemory.
func (s *Shard) EvictKeysByLRU() { // Added shard receiver
	// MaxMemory == 0 means no limit.
	// If current store length is already less than MaxMemory, no eviction needed.
	if s.MaxMemory == 0 || len(s.Store) < s.MaxMemory {
		return // No memory limit or already within limit
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	// Keep evicting until we are below the target ratio of MaxMemory
	targetSize := int(float64(s.MaxMemory) * evictionTargetRatio)
	if targetSize == 0 { // Avoid division by zero or infinite loop if MaxMemory is 1
		targetSize = s.MaxMemory
	}

	for len(s.Store) > targetSize {
		// Sample a few keys
		sample := make([]string, 0, lruEvictionSample)
		
		// Fill the sample
		filled := 0
		for key := range s.Store {
			sample = append(sample, key)
			filled++
			if filled >= lruEvictionSample {
				break
			}
		}

		if len(sample) == 0 { // No keys to evict, should not happen if len(s.Store) > 0 and len(s.Store) > targetSize
			break 
		}

		// Find the key with the lowest LRU value in the sample
		var lruKeyToEvict string
		minLRU := uint64(math.MaxUint64)

		for _, key := range sample {
			entry, ok := s.Store[key]
			if !ok {
				continue 
			}
			if entry.LRU < minLRU {
				minLRU = entry.LRU
				lruKeyToEvict = key
			}
		}

		if lruKeyToEvict == "" { // No suitable key found to evict in sample, should not happen if len(sample) > 0
			break
		}

		delete(s.Store, lruKeyToEvict)
	}
}
