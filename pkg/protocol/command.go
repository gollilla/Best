package protocol

import (
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// handleCommandOutput handles command execution results
// Note: Some Bedrock servers (like PNX) send command output via CommandOutput packet
// while others (like PMMP) send via Text packets
func (c *Client) handleCommandOutput(pk packet.Packet) {
	p := pk.(*packet.CommandOutput)

	// Build output string from messages
	var outputLines []string
	for _, msg := range p.OutputMessages {
		// Include both message key and parameters for full context
		// Message may contain translation keys like "%commands.generic.unknown"
		var parts []string
		if msg.Message != "" {
			parts = append(parts, msg.Message)
		}
		if len(msg.Parameters) > 0 {
			parts = append(parts, msg.Parameters...)
		}
		text := strings.Join(parts, " ")
		if text != "" {
			outputLines = append(outputLines, text)
		}
	}

	output := &types.CommandOutput{
		Command:    "", // Will be set by the caller
		Success:    p.SuccessCount > 0,
		Output:     strings.Join(outputLines, "\n"),
		StatusCode: int32(p.OutputType),
	}

	c.emitter.Emit(events.EventCommandOutput, output)
}
