package world

import (
	"sync"
)

// BlockRegistry maps block runtime IDs to block names
type BlockRegistry struct {
	idToName map[uint32]string
	nameToID map[string]uint32
	mu       sync.RWMutex
}

// NewBlockRegistry creates a new block registry
func NewBlockRegistry() *BlockRegistry {
	return &BlockRegistry{
		idToName: make(map[uint32]string),
		nameToID: make(map[string]uint32),
	}
}

// Register registers a block mapping
func (r *BlockRegistry) Register(runtimeID uint32, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idToName[runtimeID] = name
	r.nameToID[name] = runtimeID
}

// GetName returns the block name for a runtime ID
func (r *BlockRegistry) GetName(runtimeID uint32) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	name, ok := r.idToName[runtimeID]
	return name, ok
}

// GetID returns the runtime ID for a block name
func (r *BlockRegistry) GetID(name string) (uint32, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.nameToID[name]
	return id, ok
}

// Count returns the number of registered blocks
func (r *BlockRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.idToName)
}

// Clear clears all registered blocks
func (r *BlockRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idToName = make(map[uint32]string)
	r.nameToID = make(map[string]uint32)
}
