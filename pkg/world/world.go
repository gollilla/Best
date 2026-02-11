package world

import (
	"sync"

	"github.com/gollilla/best/pkg/types"
)

// World manages the world state including blocks and chunks
type World struct {
	blocks   map[types.Position]*types.Block
	chunks   map[ChunkPos]*Chunk
	registry *BlockRegistry
	mu       sync.RWMutex
}

// ChunkPos represents a chunk position
type ChunkPos struct {
	X int32
	Z int32
}

// Chunk represents a chunk of blocks (16x256x16 or 16x384x16)
type Chunk struct {
	Position ChunkPos
	SubChunks []*SubChunk
}

// SubChunk represents a 16x16x16 section of a chunk
type SubChunk struct {
	Y      int8
	Blocks []uint32 // Block runtime IDs
}

// NewWorld creates a new world instance
func NewWorld() *World {
	return &World{
		blocks:   make(map[types.Position]*types.Block),
		chunks:   make(map[ChunkPos]*Chunk),
		registry: NewBlockRegistry(),
	}
}

// SetBlock sets a block at the given position
func (w *World) SetBlock(pos types.Position, block *types.Block) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.blocks[pos] = block
}

// GetBlock returns the block at the given position
func (w *World) GetBlock(pos types.Position) (*types.Block, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	block, ok := w.blocks[pos]
	return block, ok
}

// RemoveBlock removes a block at the given position
func (w *World) RemoveBlock(pos types.Position) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.blocks, pos)
}

// SetChunk sets a chunk
func (w *World) SetChunk(chunkPos ChunkPos, chunk *Chunk) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.chunks[chunkPos] = chunk
}

// GetChunk returns a chunk
func (w *World) GetChunk(chunkPos ChunkPos) (*Chunk, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	chunk, ok := w.chunks[chunkPos]
	return chunk, ok
}

// Registry returns the block registry
func (w *World) Registry() *BlockRegistry {
	return w.registry
}

// BlockCount returns the number of tracked blocks
func (w *World) BlockCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.blocks)
}

// ChunkCount returns the number of loaded chunks
func (w *World) ChunkCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.chunks)
}

// Clear clears all world data
func (w *World) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.blocks = make(map[types.Position]*types.Block)
	w.chunks = make(map[ChunkPos]*Chunk)
}
