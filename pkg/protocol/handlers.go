package protocol

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// handleUpdateBlock handles block update packets
func (c *Client) handleUpdateBlock(pk packet.Packet) {
	p := pk.(*packet.UpdateBlock)

	update := &types.BlockUpdate{
		Position: types.Position{
			X: float64(p.Position.X()),
			Y: float64(p.Position.Y()),
			Z: float64(p.Position.Z()),
		},
		RuntimeID: int32(p.NewBlockRuntimeID),
	}

	c.emitter.Emit(events.EventBlockUpdate, update)
}

// handleInventoryContent handles full inventory updates
func (c *Client) handleInventoryContent(pk packet.Packet) {
	p := pk.(*packet.InventoryContent)

	// Convert items
	items := make([]types.InventoryItem, 0, len(p.Content))
	for i, item := range p.Content {
		if item.Stack.ItemType.NetworkID == 0 {
			continue // Skip air/empty slots
		}

		inventoryItem := types.InventoryItem{
			ID:    fmt.Sprintf("item:%d", item.Stack.ItemType.NetworkID), // Use NetworkID for now
			Count: int32(item.Stack.Count),
			Slot:  int32(i),
		}
		items = append(items, inventoryItem)
	}

	c.emitter.Emit(events.EventInventoryUpdate, items)
}

// handleInventorySlot handles single slot updates
func (c *Client) handleInventorySlot(pk packet.Packet) {
	p := pk.(*packet.InventorySlot)

	if p.NewItem.Stack.ItemType.NetworkID == 0 {
		// Empty slot
		c.emitter.Emit(events.EventInventorySlotUpdate, types.InventoryItem{
			Slot: int32(p.Slot),
		})
		return
	}

	item := types.InventoryItem{
		ID:    fmt.Sprintf("item:%d", p.NewItem.Stack.ItemType.NetworkID), // Use NetworkID for now
		Count: int32(p.NewItem.Stack.Count),
		Slot:  int32(p.Slot),
	}

	c.emitter.Emit(events.EventInventorySlotUpdate, item)
}

// handleMobEffect handles effect application/removal
func (c *Client) handleMobEffect(pk packet.Packet) {
	p := pk.(*packet.MobEffect)

	// Only handle effects for the player
	if p.EntityRuntimeID != uint64(c.state.RuntimeEntityID) {
		return
	}

	effect := &types.Effect{
		ID:        "", // Would need effect ID mapping
		Amplifier: int32(p.Amplifier),
		Duration:  int32(p.Duration),
		Visible:   p.Particles,
	}

	switch p.Operation {
	case packet.MobEffectAdd, packet.MobEffectModify:
		c.emitter.Emit(events.EventEffectAdd, effect)
	case packet.MobEffectRemove:
		c.emitter.Emit(events.EventEffectRemove, effect)
	}
}

// handleAddActor handles entity spawning
func (c *Client) handleAddActor(pk packet.Packet) {
	p := pk.(*packet.AddActor)

	entity := &types.Entity{
		RuntimeID: int64(p.EntityRuntimeID),
		Type:      p.EntityType,
		Position: types.Position{
			X: float64(p.Position.X()),
			Y: float64(p.Position.Y()),
			Z: float64(p.Position.Z()),
		},
	}

	c.emitter.Emit(events.EventEntityAdd, entity)
}

// handleRemoveActor handles entity removal
func (c *Client) handleRemoveActor(pk packet.Packet) {
	p := pk.(*packet.RemoveActor)

	c.emitter.Emit(events.EventEntityRemove, int64(p.EntityUniqueID))
}

// handleLevelChunk handles chunk data
func (c *Client) handleLevelChunk(pk packet.Packet) {
	// p := pk.(*packet.LevelChunk)

	// TODO: Implement chunk decoding
	// This is complex and requires:
	// 1. Parsing sub-chunk count
	// 2. Decoding palettes for each sub-chunk
	// 3. Decompressing and reading block data
	// 4. Updating the world state

	// For now, we just acknowledge receipt
	// Full implementation would decode and store chunk data
}

// handleSetTitle handles title/subtitle/actionbar display
func (c *Client) handleSetTitle(pk packet.Packet) {
	p := pk.(*packet.SetTitle)

	var titleType string
	switch p.ActionType {
	case packet.TitleActionSetTitle:
		titleType = "title"
	case packet.TitleActionSetSubtitle:
		titleType = "subtitle"
	case packet.TitleActionSetActionBar:
		titleType = "actionbar"
	case packet.TitleActionClear, packet.TitleActionReset:
		// Emit clear event
		c.emitter.Emit(events.EventTitle, &types.TitleDisplay{
			Type: "clear",
		})
		return
	default:
		return
	}

	titleDisplay := &types.TitleDisplay{
		Type:    titleType,
		Text:    p.Text,
		FadeIn:  int32(p.FadeInDuration),
		Stay:    int32(p.RemainDuration),
		FadeOut: int32(p.FadeOutDuration),
	}

	c.emitter.Emit(events.EventTitle, titleDisplay)
}

// handleSetScore handles scoreboard score updates
func (c *Client) handleSetScore(pk packet.Packet) {
	p := pk.(*packet.SetScore)
	fmt.Printf("[DEBUG] SetScore packet received with %d entries\n", len(p.Entries))

	for _, entry := range p.Entries {
		fmt.Printf("[DEBUG] Score entry - Objective: %s, Score: %d, DisplayName: %s\n",
			entry.ObjectiveName, entry.Score, entry.DisplayName)

		scoreEntry := &types.ScoreboardEntry{
			Objective: entry.ObjectiveName,
			Score:     entry.Score,
		}

		// Set display name if available
		if entry.DisplayName != "" {
			displayName := entry.DisplayName
			scoreEntry.DisplayName = &displayName
		}

		c.emitter.Emit(events.EventScoreUpdate, scoreEntry)
		fmt.Printf("[DEBUG] Emitted EventScoreUpdate for %s\n", entry.ObjectiveName)
	}
}

// handleSetDisplayObjective handles scoreboard display changes
func (c *Client) handleSetDisplayObjective(pk packet.Packet) {
	p := pk.(*packet.SetDisplayObjective)
	fmt.Printf("[DEBUG] SetDisplayObjective packet - Slot: %s, Objective: %s\n",
		p.DisplaySlot, p.ObjectiveName)

	// Emit display objective change
	displayInfo := map[string]interface{}{
		"displaySlot":   p.DisplaySlot,
		"objectiveName": p.ObjectiveName,
		"displayName":   p.DisplayName,
		"sortOrder":     p.SortOrder,
	}

	c.emitter.Emit(events.EventScoreUpdate, displayInfo)
	fmt.Printf("[DEBUG] Emitted EventScoreUpdate (display objective)\n")
}

// handleRemoveObjective handles scoreboard objective removal
func (c *Client) handleRemoveObjective(pk packet.Packet) {
	p := pk.(*packet.RemoveObjective)

	// Emit objective removal
	removeInfo := map[string]interface{}{
		"objectiveName": p.ObjectiveName,
		"removed":       true,
	}

	c.emitter.Emit(events.EventScoreUpdate, removeInfo)
}
