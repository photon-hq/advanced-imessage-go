package imessage

// Chat is the canonical projection of a chat joined with its participants and
// (optionally) its most recent message.
type Chat struct {
	// GUID is the stable chat identifier, the key for every chat-scoped call.
	GUID ChatGUID
	// ChatIdentifier is Apple's chat_identifier (peer address for 1:1 chats,
	// synthetic group ID for groups); "" if absent.
	ChatIdentifier string
	// GroupID is the stable iMessage group ID; set on group chats only.
	GroupID string
	// DisplayName is the chat's display name; "" if unset.
	DisplayName string
	IsGroup     bool
	Service     ChatServiceType
	IsArchived  bool
	// IsFiltered reports whether Apple routed the chat to Unknown Senders.
	IsFiltered bool
	// UnreadCount is the number of unread messages; nil when unavailable (0 is
	// a real "all read" value).
	UnreadCount *int
	// Participants are in the order they joined the chat.
	Participants []Handle
	// LastMessage is the most recent message, if the server included it.
	LastMessage *Message
}

// CreateChatOptions configures [ChatClient.Create]. A nil value creates a chat
// without sending an initial message.
type CreateChatOptions struct {
	// Message is the optional initial message text.
	Message string
	// AttributedBody is an optional pre-encoded attributed body for the initial
	// message.
	AttributedBody []byte
	// Effect is an optional expressive-send effect for the initial message.
	Effect MessageEffect
	// Subject is an optional subject for the initial message.
	Subject string
	// ClientMessageID de-duplicates retried creation.
	ClientMessageID string
}

// CreateChatResult is the outcome of creating a chat.
type CreateChatResult struct {
	Chat Chat
	// InitialMessage is the sent initial message, if one was requested.
	InitialMessage *Message
}
