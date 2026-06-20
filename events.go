package imessage

import "time"

// EventContext is the common envelope on every live event: who caused it, in
// which chat, and when.
type EventContext struct {
	// Actor is the address that caused the change, or nil if unattributed.
	Actor *Handle
	// ChatGUID is the chat the event occurred in.
	ChatGUID string
	// IsFromMe reports whether the local user caused the change.
	IsFromMe bool
	// OccurredAt is when the change happened.
	OccurredAt time.Time
}

// MessageEvent is a change observed on a message subscription. Type-switch over
// the variant structs to handle it.
type MessageEvent interface{ isMessageEvent() }

// MessageReceived is a new inbound or outbound message.
type MessageReceived struct {
	EventContext
	Sequence uint64
	Message  Message
}

// MessageEdited is an edit to an existing message.
type MessageEdited struct {
	EventContext
	Sequence    uint64
	MessageGUID string
	Content     MessageContent
	EditedAt    time.Time
}

// MessageRead is a read receipt.
type MessageRead struct {
	EventContext
	Sequence    uint64
	MessageGUID string
	ReadAt      time.Time
}

// MessageUnsent is a retraction (unsend) of a message.
type MessageUnsent struct {
	EventContext
	Sequence    uint64
	MessageGUID string
	RetractedAt time.Time
}

// MessageReactionAdded is a tapback added to a message.
type MessageReactionAdded struct {
	EventContext
	Sequence        uint64
	MessageGUID     string
	Reaction        MessageReaction
	TargetPartIndex *int
}

// MessageReactionRemoved is a tapback removed from a message.
type MessageReactionRemoved struct {
	EventContext
	Sequence        uint64
	MessageGUID     string
	Reaction        MessageReaction
	TargetPartIndex *int
}

// MessageStickerPlaced is a sticker placed on a message.
type MessageStickerPlaced struct {
	EventContext
	Sequence        uint64
	MessageGUID     string
	Sticker         *AttachmentInfo
	Placement       *StickerPlacement
	TargetPartIndex *int
}

// ChatEvent is a change observed on a chat subscription.
type ChatEvent interface{ isChatEvent() }

// ChatBackgroundChanged reports the chat background being set or replaced.
type ChatBackgroundChanged struct {
	EventContext
	Sequence uint64
}

// ChatBackgroundRemoved reports the chat background being removed.
type ChatBackgroundRemoved struct {
	EventContext
	Sequence uint64
}

// ChatMarkedRead reports the chat being marked read.
type ChatMarkedRead struct {
	EventContext
	Sequence uint64
}

// ChatArchived reports the chat being archived.
type ChatArchived struct {
	EventContext
	Sequence uint64
}

// ChatUnarchived reports the chat being unarchived.
type ChatUnarchived struct {
	EventContext
	Sequence uint64
}

// GroupEvent is a change observed on a group subscription.
type GroupEvent interface{ isGroupEvent() }

// GroupChanged wraps a specific [GroupChange].
type GroupChanged struct {
	EventContext
	Sequence uint64
	Change   GroupChange
}

// PollEvent is a change observed on a poll subscription.
type PollEvent interface{ isPollEvent() }

// PollChanged wraps a specific [PollChangeDelta].
type PollChanged struct {
	EventContext
	Sequence        uint64
	PollMessageGUID string
	Delta           PollChangeDelta
}

// CatchUpEvent is an event delivered by [EventClient.CatchUp]. It is one of the
// live event variants (which also satisfy this interface) or [CatchupComplete].
type CatchUpEvent interface{ isCatchUpEvent() }

// CatchupComplete terminates a catch-up stream and reports the head sequence
// reached. Subsequent live subscriptions are guaranteed to start after it.
type CatchupComplete struct {
	HeadSequence uint64
}

// Message-event markers.
func (MessageReceived) isMessageEvent()        {}
func (MessageEdited) isMessageEvent()          {}
func (MessageRead) isMessageEvent()            {}
func (MessageUnsent) isMessageEvent()          {}
func (MessageReactionAdded) isMessageEvent()   {}
func (MessageReactionRemoved) isMessageEvent() {}
func (MessageStickerPlaced) isMessageEvent()   {}

// Chat-event markers.
func (ChatBackgroundChanged) isChatEvent() {}
func (ChatBackgroundRemoved) isChatEvent() {}
func (ChatMarkedRead) isChatEvent()        {}
func (ChatArchived) isChatEvent()          {}
func (ChatUnarchived) isChatEvent()        {}

// Group- and poll-event markers.
func (GroupChanged) isGroupEvent() {}
func (PollChanged) isPollEvent()   {}

// Catch-up markers: every live variant also flows through CatchUp.
func (MessageReceived) isCatchUpEvent()        {}
func (MessageEdited) isCatchUpEvent()          {}
func (MessageRead) isCatchUpEvent()            {}
func (MessageUnsent) isCatchUpEvent()          {}
func (MessageReactionAdded) isCatchUpEvent()   {}
func (MessageReactionRemoved) isCatchUpEvent() {}
func (MessageStickerPlaced) isCatchUpEvent()   {}
func (ChatBackgroundChanged) isCatchUpEvent()  {}
func (ChatBackgroundRemoved) isCatchUpEvent()  {}
func (ChatMarkedRead) isCatchUpEvent()         {}
func (ChatArchived) isCatchUpEvent()           {}
func (ChatUnarchived) isCatchUpEvent()         {}
func (GroupChanged) isCatchUpEvent()           {}
func (PollChanged) isCatchUpEvent()            {}
func (CatchupComplete) isCatchUpEvent()        {}
