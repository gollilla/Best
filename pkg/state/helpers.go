package state

import (
	"math"

	"github.com/gollilla/best/pkg/types"
)

// DistanceTo calculates the Euclidean distance between two positions
func DistanceTo(from, to types.Position) float64 {
	dx := to.X - from.X
	dy := to.Y - from.Y
	dz := to.Z - from.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// DistanceToSquared calculates the squared distance (faster, no sqrt)
func DistanceToSquared(from, to types.Position) float64 {
	dx := to.X - from.X
	dy := to.Y - from.Y
	dz := to.Z - from.Z
	return dx*dx + dy*dy + dz*dz
}

// HorizontalDistance calculates distance ignoring Y coordinate
func HorizontalDistance(from, to types.Position) float64 {
	dx := to.X - from.X
	dz := to.Z - from.Z
	return math.Sqrt(dx*dx + dz*dz)
}
