package agent

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// Chat sends a chat message
func (a *Agent) Chat(message string) error {
	if !a.isConnected.Load() {
		return fmt.Errorf("not connected")
	}

	// Use Text packet for chat messages (original implementation)
	pk := &packet.Text{
		TextType:         packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName:       a.username,
		Message:          message,
		XUID:             "",
		PlatformChatID:   "",
	}

	return a.client.WritePacket(pk)
}

// Command sends a command to the server
// Send method is determined by agent configuration (commandSendMethod)
// Use Chat() or CommandOutput() assertions to wait for the response
func (a *Agent) Command(cmd string) error {
	// Ensure command has leading slash
	if len(cmd) > 0 && cmd[0] != '/' {
		cmd = "/" + cmd
	}

	// Send command based on configured method
	if a.commandSendMethod == "request" {
		return a.sendCommandViaRequest(cmd)
	}
	return a.sendCommandViaText(cmd)
}

// sendCommandViaText sends a command as a chat message (Text packet)
func (a *Agent) sendCommandViaText(cmd string) error {
	if !a.isConnected.Load() {
		return fmt.Errorf("not connected")
	}
	return a.Chat(cmd)
}

// sendCommandViaRequest sends a command via CommandRequest packet
func (a *Agent) sendCommandViaRequest(cmd string) error {
	if !a.isConnected.Load() {
		return fmt.Errorf("not connected")
	}

	// Remove leading slash for CommandRequest
	cmdLine := cmd
	if strings.HasPrefix(cmdLine, "/") {
		cmdLine = cmdLine[1:]
	}

	pk := &packet.CommandRequest{
		CommandLine: cmdLine,
		CommandOrigin: protocol.CommandOrigin{
			Origin:         protocol.CommandOriginPlayer,
			UUID:           uuid.New(),
			RequestID:      "",
			PlayerUniqueID: a.state.RuntimeEntityID,
		},
		Internal: false,
	}

	return a.client.WritePacket(pk)
}

// Goto teleports the player to the specified position
func (a *Agent) Goto(pos types.Position) error {
	cmd := fmt.Sprintf("/tp @s %.2f %.2f %.2f", pos.X, pos.Y, pos.Z)
	return a.Command(cmd)
}

// LookAt makes the player look at a specific position
func (a *Agent) LookAt(pos types.Position) error {
	current := a.Position()
	dx := pos.X - current.X
	dy := pos.Y - current.Y
	dz := pos.Z - current.Z

	// Calculate yaw and pitch
	yaw := float32(-math.Atan2(dx, dz) * (180 / math.Pi))
	distance := math.Sqrt(dx*dx + dz*dz)
	pitch := float32(-math.Atan2(dy, distance) * (180 / math.Pi))

	// Send move packet with new rotation
	pk := &packet.MovePlayer{
		EntityRuntimeID: uint64(a.state.RuntimeEntityID),
		Position:        mgl32.Vec3{float32(current.X), float32(current.Y), float32(current.Z)},
		Pitch:           pitch,
		Yaw:             yaw,
		HeadYaw:         yaw,
		Mode:            packet.MoveModeNormal,
		OnGround:        a.state.IsOnGround,
		Tick:            0,
	}

	return a.client.WritePacket(pk)
}

// SendPacket sends a raw packet to the server
func (a *Agent) SendPacket(pk packet.Packet) error {
	if !a.isConnected.Load() {
		return fmt.Errorf("not connected")
	}
	return a.client.WritePacket(pk)
}

// OnPacket registers a handler for a specific packet type
func (a *Agent) OnPacket(packetID uint32, handler func(pk packet.Packet)) {
	a.client.RegisterHandler(packetID, handler)
}

// WaitForChat waits for a chat message matching the filter
func (a *Agent) WaitForChat(ctx context.Context, filter func(*types.ChatMessage) bool) (*types.ChatMessage, error) {
	data, err := a.emitter.WaitFor(ctx, events.EventChat, func(d events.EventData) bool {
		msg, ok := d.(*types.ChatMessage)
		if !ok {
			return false
		}
		if filter == nil {
			return true
		}
		return filter(msg)
	})

	if err != nil {
		return nil, err
	}

	return data.(*types.ChatMessage), nil
}
