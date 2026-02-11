package protocol

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// handleText handles chat messages
func (c *Client) handleText(pk packet.Packet) {
	p := pk.(*packet.Text)

	message := p.Message
	sender := p.SourceName

	// For translation packets (like chat echoes from server),
	// the actual message is in Parameters
	if p.TextType == packet.TextTypeTranslation {
		// Parameters format: [sender, message]
		if len(p.Parameters) >= 2 {
			sender = p.Parameters[0]
			message = p.Parameters[1]
		} else if len(p.Parameters) == 1 {
			message = p.Parameters[0]
		}
	}

	msg := &types.ChatMessage{
		Type:      mapTextType(p.TextType),
		Sender:    sender,
		Message:   message,
		Timestamp: 0, // Will be set by caller if needed
		XUID:      p.XUID,
	}

	c.emitter.Emit(events.EventChat, msg)
}

// handleMovePlayer handles player movement
func (c *Client) handleMovePlayer(pk packet.Packet) {
	p := pk.(*packet.MovePlayer)

	// Update state if this is our player
	if p.EntityRuntimeID == uint64(c.state.RuntimeEntityID) {
		c.state.Position = types.Position{
			X: float64(p.Position.X()),
			Y: float64(p.Position.Y()),
			Z: float64(p.Position.Z()),
		}
		c.state.Rotation = types.Rotation{
			Yaw:   p.Yaw,
			Pitch: p.Pitch,
		}
		c.state.IsOnGround = p.OnGround

		c.emitter.Emit(events.EventPositionUpdate, c.state.Position)
	}
}

// handleStartGame handles the initial game start
func (c *Client) handleStartGame(pk packet.Packet) {
	p := pk.(*packet.StartGame)

	c.state.RuntimeEntityID = int64(p.EntityRuntimeID)
	c.state.Position = types.Position{
		X: float64(p.PlayerPosition.X()),
		Y: float64(p.PlayerPosition.Y()),
		Z: float64(p.PlayerPosition.Z()),
	}
	c.state.Gamemode = p.PlayerGameMode
}

// handleUpdateAttributes handles attribute updates (health, hunger, etc.)
func (c *Client) handleUpdateAttributes(pk packet.Packet) {
	p := pk.(*packet.UpdateAttributes)

	if p.EntityRuntimeID == uint64(c.state.RuntimeEntityID) {
		for _, attr := range p.Attributes {
			switch attr.Name {
			case "minecraft:health":
				c.state.Health = attr.Value
				c.emitter.Emit(events.EventHealthUpdate, attr.Value)
			case "minecraft:player.hunger":
				c.emitter.Emit(events.EventHungerUpdate, attr.Value)
			}
		}
	}
}

// handleSetPlayerGameType handles gamemode changes
func (c *Client) handleSetPlayerGameType(pk packet.Packet) {
	p := pk.(*packet.SetPlayerGameType)

	c.state.Gamemode = p.GameType
	c.emitter.Emit(events.EventGamemodeUpdate, p.GameType)
}

// handleDisconnect handles disconnection
func (c *Client) handleDisconnect(pk packet.Packet) {
	p := pk.(*packet.Disconnect)

	c.emitter.Emit(events.EventDisconnect, p.Message)
}

// mapTextType maps packet text types to our string types
func mapTextType(textType byte) string {
	switch textType {
	case packet.TextTypeRaw:
		return "raw"
	case packet.TextTypeChat:
		return "chat"
	case packet.TextTypeTranslation:
		return "translation"
	case packet.TextTypeSystem:
		return "system"
	case packet.TextTypeWhisper:
		return "whisper"
	case packet.TextTypeAnnouncement:
		return "announcement"
	case packet.TextTypeTip:
		return "tip"
	default:
		return "raw"
	}
}
