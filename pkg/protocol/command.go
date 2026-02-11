package protocol

import (
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// handleCommandOutput handles command execution results
// Note: Most Bedrock servers send command output via Text packets instead
func (c *Client) handleCommandOutput(pk packet.Packet) {
	p := pk.(*packet.CommandOutput)

	// Build output string from messages
	var outputLines []string
	for _, msg := range p.OutputMessages {
		// Combine message parameters
		text := strings.Join(msg.Parameters, " ")
		if text == "" && msg.Message != "" {
			text = msg.Message
		}
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
