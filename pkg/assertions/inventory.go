package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// InventoryAssertion provides inventory-related assertions
type InventoryAssertion struct {
	agent AgentInterface
}

// ToHaveItem checks if the inventory contains a specific item
// itemID can be a full ID (e.g., "minecraft:diamond") or a partial match (e.g., "diamond")
func (i *InventoryAssertion) ToHaveItem(itemID string) {
	items := i.agent.GetInventory()

	for _, item := range items {
		if matchesItemID(item.ID, itemID) {
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected inventory to have item %q", itemID),
		itemID,
		getInventoryItemIDs(items),
	))
}

// ToHaveItemCount checks if the inventory contains a specific count of an item
func (i *InventoryAssertion) ToHaveItemCount(itemID string, expectedCount int32) {
	items := i.agent.GetInventory()

	var totalCount int32
	for _, item := range items {
		if matchesItemID(item.ID, itemID) {
			totalCount += item.Count
		}
	}

	if totalCount == expectedCount {
		return
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected inventory to have %d of item %q, but found %d", expectedCount, itemID, totalCount),
		expectedCount,
		totalCount,
	))
}

// ToHaveAtLeast checks if the inventory contains at least a certain count of an item
func (i *InventoryAssertion) ToHaveAtLeast(itemID string, minCount int32) {
	items := i.agent.GetInventory()

	var totalCount int32
	for _, item := range items {
		if matchesItemID(item.ID, itemID) {
			totalCount += item.Count
		}
	}

	if totalCount >= minCount {
		return
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected inventory to have at least %d of item %q, but found %d", minCount, itemID, totalCount),
		minCount,
		totalCount,
	))
}

// ToBeEmpty checks if the inventory is empty
func (i *InventoryAssertion) ToBeEmpty() {
	items := i.agent.GetInventory()

	if len(items) == 0 {
		return
	}

	panic(NewAssertionError(
		"expected inventory to be empty",
		0,
		len(items),
	))
}

// ToReceiveItem waits for an inventory update containing a specific item
func (i *InventoryAssertion) ToReceiveItem(ctx context.Context, itemID string) *types.InventoryItem {
	data, err := i.agent.Emitter().WaitFor(ctx, events.EventInventoryUpdate, func(d events.EventData) bool {
		items, ok := d.([]types.InventoryItem)
		if !ok {
			return false
		}

		for _, item := range items {
			if matchesItemID(item.ID, itemID) {
				return true
			}
		}
		return false
	})

	if err != nil {
		panic(err)
	}

	items := data.([]types.InventoryItem)
	for _, item := range items {
		if matchesItemID(item.ID, itemID) {
			return &item
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("received inventory update but item %q not found", itemID),
		itemID,
		nil,
	))
}

// ToReceiveItemInSlot waits for an inventory slot update for a specific slot
func (i *InventoryAssertion) ToReceiveItemInSlot(ctx context.Context, slot int32) *types.InventoryItem {
	data, err := i.agent.Emitter().WaitFor(ctx, events.EventInventorySlotUpdate, func(d events.EventData) bool {
		item, ok := d.(types.InventoryItem)
		if !ok {
			return false
		}
		return item.Slot == slot
	})

	if err != nil {
		panic(err)
	}

	item := data.(types.InventoryItem)
	return &item
}

// WaitForInventoryChange waits for any inventory change within the timeout
func (i *InventoryAssertion) WaitForInventoryChange(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := i.agent.Emitter().WaitFor(ctx, events.EventInventoryUpdate, nil)
	if err != nil {
		panic(err)
	}
}

// Helper functions

// matchesItemID checks if an item ID matches the expected pattern
// Supports:
// - Full IDs (minecraft:diamond, item:335)
// - Partial matches (diamond matches minecraft:diamond)
// - Network IDs (335 matches item:335)
func matchesItemID(actualID, expectedID string) bool {
	// Exact match
	if actualID == expectedID {
		return true
	}

	// Check if expectedID is a numeric ID (e.g., "335" matches "item:335")
	if strings.HasPrefix(actualID, "item:") && strings.TrimPrefix(actualID, "item:") == expectedID {
		return true
	}

	// Partial match (e.g., "diamond" matches "minecraft:diamond")
	if strings.Contains(actualID, expectedID) {
		return true
	}

	// Check if expected has namespace but actual doesn't
	if strings.Contains(expectedID, ":") {
		parts := strings.Split(expectedID, ":")
		if len(parts) == 2 && strings.Contains(actualID, parts[1]) {
			return true
		}
	}

	return false
}

// getInventoryItemIDs returns a list of item IDs in the inventory
func getInventoryItemIDs(items []types.InventoryItem) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
