package world

import (
	"fmt"
)

// DecodeChunk decodes a chunk from raw data
// This is a simplified implementation - full chunk decoding is complex
// and would require palette handling, biome data, etc.
func DecodeChunk(data []byte, chunkX, chunkZ int32) (*Chunk, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty chunk data")
	}

	chunk := &Chunk{
		Position: ChunkPos{X: chunkX, Z: chunkZ},
		SubChunks: make([]*SubChunk, 0),
	}

	// TODO: Implement full chunk decoding with palette support
	// For now, we just create an empty chunk structure
	// Full implementation would:
	// 1. Read the number of sub-chunks
	// 2. For each sub-chunk, read the palette and block data
	// 3. Decode the block runtime IDs using the palette
	// 4. Store the decoded blocks

	return chunk, nil
}

// GetBlockAt returns the block runtime ID at the given position within the chunk
func (c *Chunk) GetBlockAt(x, y, z int) (uint32, error) {
	if x < 0 || x >= 16 || z < 0 || z >= 16 {
		return 0, fmt.Errorf("position out of chunk bounds")
	}

	subChunkY := y / 16
	if subChunkY < 0 || subChunkY >= len(c.SubChunks) {
		return 0, fmt.Errorf("y position out of range")
	}

	subChunk := c.SubChunks[subChunkY]
	if subChunk == nil {
		return 0, nil // Air
	}

	localY := y % 16
	index := (localY * 256) + (z * 16) + x

	if index < 0 || index >= len(subChunk.Blocks) {
		return 0, fmt.Errorf("block index out of range")
	}

	return subChunk.Blocks[index], nil
}

// SetBlockAt sets the block runtime ID at the given position within the chunk
func (c *Chunk) SetBlockAt(x, y, z int, runtimeID uint32) error {
	if x < 0 || x >= 16 || z < 0 || z >= 16 {
		return fmt.Errorf("position out of chunk bounds")
	}

	subChunkY := y / 16

	// Ensure we have enough sub-chunks
	for len(c.SubChunks) <= subChunkY {
		c.SubChunks = append(c.SubChunks, &SubChunk{
			Y:      int8(len(c.SubChunks)),
			Blocks: make([]uint32, 4096), // 16x16x16
		})
	}

	subChunk := c.SubChunks[subChunkY]
	localY := y % 16
	index := (localY * 256) + (z * 16) + x

	if index < 0 || index >= len(subChunk.Blocks) {
		return fmt.Errorf("block index out of range")
	}

	subChunk.Blocks[index] = runtimeID
	return nil
}
