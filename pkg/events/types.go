package events

// EventName represents the type of event
type EventName string

// Event type constants
const (
	EventJoin            EventName = "join"
	EventSpawn           EventName = "spawn"
	EventDisconnect      EventName = "disconnect"
	EventError           EventName = "error"
	EventChat            EventName = "chat"
	EventPositionUpdate  EventName = "position_update"
	EventHealthUpdate    EventName = "health_update"
	EventHungerUpdate    EventName = "hunger_update"
	EventGamemodeUpdate  EventName = "gamemode_update"
	EventForm            EventName = "form"
	EventCommandOutput   EventName = "command_output"
	EventChunkLoaded     EventName = "chunk_loaded"
	EventBlockUpdate     EventName = "block_update"
	EventBlockBreakStart EventName = "block_break_start"
	EventBlockBreakAbort EventName = "block_break_abort"
	EventBlockBreakComplete EventName = "block_break_complete"
	EventInventoryUpdate    EventName = "inventory_update"
	EventInventorySlotUpdate EventName = "inventory_slot_update"
	EventEffectAdd          EventName = "effect_add"
	EventEffectRemove       EventName = "effect_remove"
	EventEffectUpdate       EventName = "effect_update"
	EventEntityAdd          EventName = "entity_add"
	EventEntitySpawn        EventName = "entity_spawn"
	EventEntityRemove       EventName = "entity_remove"
	EventScoreUpdate        EventName = "score_update"
	EventPermissionUpdate   EventName = "permission_update"
	EventTagUpdate          EventName = "tag_update"
	EventTitle              EventName = "title"
	EventSound           EventName = "sound"
	EventParticle        EventName = "particle"
	EventDimensionChange EventName = "dimension_change"
	EventDeath           EventName = "death"
	EventRespawn         EventName = "respawn"
	EventTeleport        EventName = "teleport"
	EventPacket          EventName = "packet"
)

// EventData represents any event payload
type EventData interface{}

// FilterFunc is a function that filters event data
type FilterFunc func(EventData) bool
