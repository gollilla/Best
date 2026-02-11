package agent

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"
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

// Command executes a command and returns the output
// Note: Most Bedrock servers handle commands through chat rather than CommandRequest packets
func (a *Agent) Command(cmd string) (*types.CommandOutput, error) {
	if !a.isConnected.Load() {
		return nil, fmt.Errorf("not connected")
	}

	// Ensure command has leading slash
	if len(cmd) > 0 && cmd[0] != '/' {
		cmd = "/" + cmd
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(a.ctx, a.options.Timeout)
	defer cancel()

	// Collect text messages that arrive after sending the command
	textMessages := []string{}
	var mu sync.Mutex
	done := make(chan struct{})

	// Listen for text packets (command responses)
	handlerID := a.emitter.On(events.EventChat, func(data events.EventData) {
		msg, ok := data.(*types.ChatMessage)
		if !ok {
			return
		}

		if msg.Type != "chat" { // Only collect system/raw messages, not player chat
			mu.Lock()
			textMessages = append(textMessages, msg.Message)
			mu.Unlock()
		}
	})
	defer a.emitter.Off(events.EventChat, handlerID)

	// Send command as a chat message (this is how most Bedrock servers handle commands)
	if err := a.Chat(cmd); err != nil {
		return nil, err
	}

	// Wait a short time to collect all response messages
	// Most command responses arrive within 1 second
	go func() {
		time.Sleep(1000 * time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
		// Collection period ended, return what we got
		mu.Lock()
		msgs := textMessages
		mu.Unlock()

		if len(msgs) > 0 {
			output := strings.Join(msgs, "\n")
			// Check if the output contains error indicators
			success := !isCommandError(output)

			return &types.CommandOutput{
				Command:    cmd,
				Success:    success,
				Output:     output,
				StatusCode: boolToStatusCode(success),
			}, nil
		}

		// No messages received - command might have failed silently
		return &types.CommandOutput{
			Command:    cmd,
			Success:    false,
			Output:     "",
			StatusCode: 1,
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("command timeout: %s", cmd)
	}
}

// isCommandError checks if the command output contains error indicators
func isCommandError(output string) bool {
	lowerOutput := strings.ToLower(output)

	// Common Minecraft Bedrock error patterns
	errorPatterns := []string{
		"unknown command",
		"incorrect argument",
		"syntax error",
		"no targets matched",
		"permission denied",
		"not enough permissions",
		"you do not have permission",
		"unable to",
		"cannot",
		"failed to",
		"error:",
		"invalid",
		"usage:",
	}

	for _, pattern := range errorPatterns {
		if strings.Contains(lowerOutput, pattern) {
			return true
		}
	}

	return false
}

// boolToStatusCode converts a boolean success value to a status code
func boolToStatusCode(success bool) int32 {
	if success {
		return 0
	}
	return 1
}

// Goto teleports the player to the specified position
func (a *Agent) Goto(pos types.Position) error {
	cmd := fmt.Sprintf("/tp @s %.2f %.2f %.2f", pos.X, pos.Y, pos.Z)
	_, err := a.Command(cmd)
	return err
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

