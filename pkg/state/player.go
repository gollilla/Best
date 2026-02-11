package state

import (
	"github.com/gollilla/best/pkg/types"
)

// CreateInitialState creates a new player state with default values
func CreateInitialState() *types.PlayerState {
	return &types.PlayerState{
		RuntimeEntityID: 0,
		Position: types.Position{
			X: 0,
			Y: 0,
			Z: 0,
		},
		Rotation: types.Rotation{
			Yaw:   0,
			Pitch: 0,
		},
		Health:     20,
		Gamemode:   0,
		Dimension:  "overworld",
		IsOnGround: false,
	}
}
