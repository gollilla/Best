package protocol

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

// ItemData represents the item data from BedrockData
type ItemData struct {
	RuntimeID      int32 `json:"runtime_id"`
	ComponentBased bool  `json:"component_based"`
	Version        int   `json:"version"`
}

//go:embed required_item_list.json
var itemListJSON []byte

var (
	// NetworkIDToName maps NetworkID to item name (e.g., 335 -> "minecraft:diamond")
	NetworkIDToName map[int32]string

	// NameToNetworkID maps item name to NetworkID (e.g., "minecraft:diamond" -> 335)
	NameToNetworkID map[string]int32

	initOnce sync.Once
)

// InitItemRegistry initializes the item registry from embedded JSON
func InitItemRegistry() {
	initOnce.Do(func() {
		var itemList map[string]ItemData
		if err := json.Unmarshal(itemListJSON, &itemList); err != nil {
			panic(fmt.Sprintf("failed to parse item list: %v", err))
		}

		NetworkIDToName = make(map[int32]string, len(itemList))
		NameToNetworkID = make(map[string]int32, len(itemList))

		for name, data := range itemList {
			NetworkIDToName[data.RuntimeID] = name
			NameToNetworkID[name] = data.RuntimeID
		}
	})
}

// GetItemName returns the item name for a given NetworkID
// Returns empty string if not found
func GetItemName(networkID int32) string {
	InitItemRegistry()
	return NetworkIDToName[networkID]
}

// GetNetworkID returns the NetworkID for a given item name
// Returns 0 if not found
func GetNetworkID(name string) int32 {
	InitItemRegistry()
	return NameToNetworkID[name]
}

// GetItemID returns a user-friendly item ID
// Prefers the item name (minecraft:diamond) but falls back to "item:335" format
func GetItemID(networkID int32) string {
	if name := GetItemName(networkID); name != "" {
		return name
	}
	return fmt.Sprintf("item:%d", networkID)
}
