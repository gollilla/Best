package protocol

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// Client wraps gophertunnel's minecraft.Conn and manages packet handling
type Client struct {
	conn       *minecraft.Conn
	emitter    *events.Emitter
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	state      *types.PlayerState
	identifier string // Agent name or identifier for debugging

	// Packet handlers
	handlers map[uint32]PacketHandler

	mu sync.RWMutex
}

// PacketHandler is a function that handles a specific packet type
type PacketHandler func(pk packet.Packet)

// NewClient creates a new protocol client
func NewClient(emitter *events.Emitter, state *types.PlayerState, identifier string) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		emitter:    emitter,
		ctx:        ctx,
		cancel:     cancel,
		state:      state,
		identifier: identifier,
		handlers:   make(map[uint32]PacketHandler),
	}
}

// Connect establishes a connection to the Minecraft server
func (c *Client) Connect(opts types.ClientOptions) error {
	// Create dialer with minimal configuration
	// TokenSource is nil for offline/unauthenticated connections
	// KeepXBLIdentityData allows XUID to be sent even in offline mode (required for PNX)
	dialer := minecraft.Dialer{
		TokenSource:         nil,  // No authentication (offline mode)
		KeepXBLIdentityData: true, // Keep XUID for unique player UUIDs on PNX
	}

	// Set username in IdentityData with UUID and XUID
	if opts.Username != "" {
		// Generate unique XUID for each player to avoid UUID collision in PNX
		// PNX generates UUID from XUID: UUID.nameUUIDFromBytes(("pocket-auth-1-xuid:" + xuid).getBytes())
		// XUID should be 16 digits to match Xbox Live format and database constraints
		xuid := opts.XUID
		if xuid == "" {
			xuid = generateXUID()
		}
		dialer.IdentityData = login.IdentityData{
			DisplayName: opts.Username,
			Identity:    uuid.New().String(),
			XUID:        xuid,
		}
	}

	// Dial the server
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	conn, err := dialer.Dial("raknet", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn

	// Extract initial state from GameData (before handlers are registered)
	gameData := conn.GameData()
	c.state.Position = types.Position{
		X: float64(gameData.PlayerPosition.X()),
		Y: float64(gameData.PlayerPosition.Y()),
		Z: float64(gameData.PlayerPosition.Z()),
	}
	c.state.Gamemode = gameData.PlayerGameMode
	c.state.PermissionLevel = gameData.PlayerPermissions

	// Initialize scoreboard state
	c.state.Scoreboard = &types.ScoreboardState{
		Objectives: make(map[string]*types.ScoreboardObjective),
		Entries:    make(map[int64]*types.ScoreboardEntry),
	}

	// Register packet handlers
	c.registerHandlers()

	// Start packet reading goroutine
	c.wg.Add(1)
	go c.readPackets()

	// Emit join event
	c.emitter.Emit(events.EventJoin, nil)

	return nil
}

// DoSpawn performs the spawn sequence
func (c *Client) DoSpawn() error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Perform spawn sequence
	if err := c.conn.DoSpawn(); err != nil {
		return fmt.Errorf("failed to spawn: %w", err)
	}

	// Get game data and set RuntimeEntityID
	gameData := c.conn.GameData()
	c.state.RuntimeEntityID = int64(gameData.EntityRuntimeID)
	c.state.Position = types.Position{
		X: float64(gameData.PlayerPosition.X()),
		Y: float64(gameData.PlayerPosition.Y()),
		Z: float64(gameData.PlayerPosition.Z()),
	}
	c.state.Gamemode = gameData.PlayerGameMode

	// Emit spawn event
	c.emitter.Emit(events.EventSpawn, nil)

	return nil
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	c.cancel()

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	c.wg.Wait()
	return nil
}

// WritePacket sends a packet to the server
func (c *Client) WritePacket(pk packet.Packet) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WritePacket(pk)
}

// readPackets continuously reads packets from the connection
func (c *Client) readPackets() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			pk, err := c.conn.ReadPacket()
			if err != nil {
				c.emitter.Emit(events.EventError, err)
				c.emitter.Emit(events.EventDisconnect, "Connection error")
				return
			}

			// Handle the packet
			c.handlePacket(pk)

			// Emit generic packet event for debugging
			c.emitter.Emit(events.EventPacket, map[string]interface{}{
				"name":   fmt.Sprintf("%T", pk),
				"packet": pk,
			})
		}
	}
}

// handlePacket routes packets to registered handlers
func (c *Client) handlePacket(pk packet.Packet) {
	c.mu.RLock()
	handler, ok := c.handlers[pk.ID()]
	c.mu.RUnlock()

	if ok {
		handler(pk)
	}
}

// RegisterHandler registers a custom packet handler
func (c *Client) RegisterHandler(packetID uint32, handler PacketHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[packetID] = handler
}

// registerHandlers registers all built-in packet handlers
func (c *Client) registerHandlers() {
	// Phase 1: Core handlers
	c.RegisterHandler(packet.IDText, c.handleText)
	c.RegisterHandler(packet.IDMovePlayer, c.handleMovePlayer)
	c.RegisterHandler(packet.IDStartGame, c.handleStartGame)
	c.RegisterHandler(packet.IDUpdateAttributes, c.handleUpdateAttributes)
	c.RegisterHandler(packet.IDSetPlayerGameType, c.handleSetPlayerGameType)
	c.RegisterHandler(packet.IDUpdateAbilities, c.handleUpdateAbilities)
	c.RegisterHandler(packet.IDDisconnect, c.handleDisconnect)
	c.RegisterHandler(packet.IDCommandOutput, c.handleCommandOutput)

	// Phase 2: World and state handlers
	c.RegisterHandler(packet.IDUpdateBlock, c.handleUpdateBlock)
	c.RegisterHandler(packet.IDInventoryContent, c.handleInventoryContent)
	c.RegisterHandler(packet.IDInventorySlot, c.handleInventorySlot)
	c.RegisterHandler(packet.IDMobEffect, c.handleMobEffect)
	c.RegisterHandler(packet.IDAddActor, c.handleAddActor)
	c.RegisterHandler(packet.IDRemoveActor, c.handleRemoveActor)
	c.RegisterHandler(packet.IDLevelChunk, c.handleLevelChunk)

	// Phase 3: UI and display handlers
	c.RegisterHandler(packet.IDSetTitle, c.handleSetTitle)
	c.RegisterHandler(packet.IDSetScore, c.handleSetScore)
	c.RegisterHandler(packet.IDSetDisplayObjective, c.handleSetDisplayObjective)
	c.RegisterHandler(packet.IDRemoveObjective, c.handleRemoveObjective)
	c.RegisterHandler(packet.IDModalFormRequest, c.handleModalFormRequest)
}

// GetConn returns the underlying minecraft.Conn
func (c *Client) GetConn() *minecraft.Conn {
	return c.conn
}

// generateXUID generates a 16-digit XUID string similar to Xbox Live XUIDs
// Format: 16 digits (e.g., "2535405290845189")
// This avoids database length issues (some plugins expect max 20 characters)
func generateXUID() string {
	// Generate a random number between 1000000000000000 and 9999999999999999
	min := big.NewInt(1000000000000000)
	max := big.NewInt(9999999999999999)

	// Calculate range
	rangeBig := new(big.Int).Sub(max, min)
	rangeBig.Add(rangeBig, big.NewInt(1))

	// Generate random number in range
	n, err := rand.Int(rand.Reader, rangeBig)
	if err != nil {
		// Fallback to a deterministic value if random fails
		return "1000000000000000"
	}

	n.Add(n, min)
	return n.String()
}
