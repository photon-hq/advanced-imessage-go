package imessage

import "time"

// MessageItemType classifies a message row.
type MessageItemType string

const (
	// ItemNormal is an ordinary message.
	ItemNormal MessageItemType = "normal"
	// ItemParticipantChange records a participant being added or removed.
	ItemParticipantChange MessageItemType = "participantChange"
	// ItemGroupNameChange records a group rename.
	ItemGroupNameChange MessageItemType = "groupNameChange"
	// ItemChatAction is an Apple-private chat action row.
	ItemChatAction MessageItemType = "chatAction"
	// ItemUnknown is an unrecognized item type.
	ItemUnknown MessageItemType = "unknown"
)

// MessageReactionKind is the kind of a tapback reaction.
type MessageReactionKind string

const (
	// ReactionLove is the heart tapback.
	ReactionLove MessageReactionKind = "love"
	// ReactionLike is the thumbs-up tapback.
	ReactionLike MessageReactionKind = "like"
	// ReactionDislike is the thumbs-down tapback.
	ReactionDislike MessageReactionKind = "dislike"
	// ReactionLaugh is the "ha ha" tapback.
	ReactionLaugh MessageReactionKind = "laugh"
	// ReactionEmphasize is the "!!" tapback.
	ReactionEmphasize MessageReactionKind = "emphasize"
	// ReactionQuestion is the "?" tapback.
	ReactionQuestion MessageReactionKind = "question"
	// ReactionEmoji is a custom emoji tapback; see [MessageReaction.Emoji].
	ReactionEmoji MessageReactionKind = "emoji"
	// ReactionSticker is a sticker tapback (received only; not settable).
	ReactionSticker MessageReactionKind = "sticker"
	// ReactionUnknown is an unrecognized reaction kind.
	ReactionUnknown MessageReactionKind = "unknown"
)

// MessageEffect is an iMessage expressive send effect (bubble or screen). Pass
// one in the send options to play it.
type MessageEffect string

// Bubble and screen effects.
const (
	EffectSlam        MessageEffect = "com.apple.MobileSMS.expressivesend.impact"
	EffectLoud        MessageEffect = "com.apple.MobileSMS.expressivesend.loud"
	EffectGentle      MessageEffect = "com.apple.MobileSMS.expressivesend.gentle"
	EffectInvisible   MessageEffect = "com.apple.MobileSMS.expressivesend.invisibleink"
	EffectConfetti    MessageEffect = "com.apple.messages.effect.CKConfettiEffect"
	EffectFireworks   MessageEffect = "com.apple.messages.effect.CKFireworksEffect"
	EffectBalloons    MessageEffect = "com.apple.messages.effect.CKBalloonEffect"
	EffectHeart       MessageEffect = "com.apple.messages.effect.CKHeartEffect"
	EffectLasers      MessageEffect = "com.apple.messages.effect.CKLasersEffect"
	EffectCelebration MessageEffect = "com.apple.messages.effect.CKHappyBirthdayEffect"
	EffectSparkles    MessageEffect = "com.apple.messages.effect.CKSparklesEffect"
	EffectSpotlight   MessageEffect = "com.apple.messages.effect.CKSpotlightEffect"
	EffectEcho        MessageEffect = "com.apple.messages.effect.CKEchoEffect"
)

// TextEffect is an inline animated text effect for rich-text sends.
type TextEffect string

// Inline text effects.
const (
	TextEffectBig     TextEffect = "big"
	TextEffectSmall   TextEffect = "small"
	TextEffectShake   TextEffect = "shake"
	TextEffectNod     TextEffect = "nod"
	TextEffectExplode TextEffect = "explode"
	TextEffectRipple  TextEffect = "ripple"
	TextEffectBloom   TextEffect = "bloom"
	TextEffectJitter  TextEffect = "jitter"
)

// Message is the canonical projection of an iMessage row joined with its
// attachments and any associated tapback.
type Message struct {
	GUID    string
	Content MessageContent
	Subject string

	DateCreated              time.Time
	DateRead                 *time.Time
	DateDelivered            *time.Time
	DateEdited               *time.Time
	DateRetracted            *time.Time
	DatePlayed               *time.Time
	DateExpressiveSendPlayed *time.Time

	Sender   *Handle
	IsFromMe bool

	IsSent             bool
	IsDelivered        bool
	IsDeliveredQuietly bool
	DidNotifyRecipient bool
	// SendErrorCode is Apple's internal send-error code (0 = no error).
	SendErrorCode int

	IsAudioMessage             bool
	IsAutoReply                bool
	IsSystemMessage            bool
	IsForward                  bool
	IsDelayed                  bool
	IsSpam                     bool
	DataDetectorResultsPresent bool
	IsArchived                 bool
	IsServiceMessage           bool
	IsCorrupt                  bool
	IsExpirable                bool

	// ShareStatus and ShareDirection mirror Apple's name-and-photo sharing
	// columns; nil when absent.
	ShareStatus    *int
	ShareDirection *int

	ItemType MessageItemType
	// GroupTitle is set on group-name-change rows.
	GroupTitle string
	// ChatActionType is the subtype of a chat-action row; nil when absent.
	ChatActionType *int

	ThreadOriginatorGUID string
	ThreadOriginatorPart string
	ReplyToGUID          string
	// ReplyTargetGUID is the canonical public reply target, or "" if this is
	// not a threaded reply.
	ReplyTargetGUID     string
	DestinationCallerID string
	// PartCount is the number of bubble parts in a multipart message; nil when
	// absent.
	PartCount       *int
	CachedRoomNames string

	// Tapback-row fields: set only when this Message is itself a tapback.
	ReactionTargetGUID      string
	ReactionTargetPartIndex *int
	Reaction                *MessageReaction
	ReactionSelected        *bool

	AppliedReactions []MessageAppliedReaction
	PlacedStickers   []MessagePlacedSticker
	ChatGUIDs        []string
}

// MessageContent is the rendered content of a message.
type MessageContent struct {
	Text                  string
	Attachments           []AttachmentInfo
	Formatting            []TextFormat
	Mentions              []MessageMention
	BalloonBundleID       string
	ExpressiveSendStyleID string
}

// MessageReaction is a tapback's kind and optional emoji.
type MessageReaction struct {
	Kind MessageReactionKind
	// Emoji is the custom emoji when Kind is [ReactionEmoji].
	Emoji string
}

// MessageAppliedReaction is a tapback currently applied to a message.
type MessageAppliedReaction struct {
	MessageGUID     string
	TargetPartIndex *int
	Reaction        MessageReaction
	Sender          *Handle
	IsFromMe        bool
	DateCreated     time.Time
}

// MessagePlacedSticker is a sticker placed on a message.
type MessagePlacedSticker struct {
	MessageGUID     string
	TargetPartIndex *int
	Sender          *Handle
	IsFromMe        bool
	DateCreated     time.Time
	Sticker         *AttachmentInfo
	Placement       *StickerPlacement
}

// StickerPlacement is where and how a sticker sits on a message. The optional
// fields are nil when the sticker uses Apple's default for that property.
type StickerPlacement struct {
	// X and Y are the sticker's anchor position over the target message,
	// expressed as fractions in [0, 1] of the bubble's width and height.
	X float64
	Y float64
	// Scale is the sticker's size multiplier.
	Scale *float64
	// Rotation is the sticker's rotation in radians.
	Rotation *float64
	// Width is the sticker's width as a fraction of the bubble width.
	Width *float64
}

// TextFormat is a formatting run reported on inbound content.
type TextFormat struct {
	Type       string
	Start      int
	Length     int
	EffectName string
}

// MessageMention is a mention span within a message body.
type MessageMention struct {
	Address string
	Start   int
	Length  int
}

// MessagePart is one bubble of a multipart send.
type MessagePart struct {
	Text             string
	AttachmentGUID   string
	AttachmentName   string
	MentionedAddress string
	BubbleIndex      *int
	Formatting       []TextFormatInput
}

// TextFormatKind selects the formatting applied by a [TextFormatInput].
type TextFormatKind string

// Formatting kinds.
const (
	FormatBold          TextFormatKind = "bold"
	FormatItalic        TextFormatKind = "italic"
	FormatUnderline     TextFormatKind = "underline"
	FormatStrikethrough TextFormatKind = "strikethrough"
	// FormatEffect applies the animated [TextFormatInput.Effect].
	FormatEffect TextFormatKind = "effect"
)

// TextFormatInput applies formatting to a span when sending rich text.
type TextFormatInput struct {
	Kind   TextFormatKind
	Start  int
	Length int
	// Effect is used only when Kind is [FormatEffect].
	Effect TextEffect
}

// ReplyTarget identifies the message (and optional part) a send replies to.
type ReplyTarget struct {
	GUID      string
	PartIndex *int
}

// Reaction is a settable tapback for [MessageClient.SetReaction]. Kind must be
// one of the user-settable kinds (not [ReactionSticker] or [ReactionUnknown]).
type Reaction struct {
	Kind  MessageReactionKind
	Emoji string
}

// EmbeddedMedia is decoded media embedded in a message (e.g. Digital Touch).
type EmbeddedMedia struct {
	Data     []byte
	MIMEType string
}

// IdempotencyOptions carries an optional client-supplied idempotency key. It is
// the options type for operations that take no other parameters.
type IdempotencyOptions struct {
	// ClientMessageID, when set, de-duplicates retried writes server-side.
	ClientMessageID string
}

// SendTextOptions configures [MessageClient.SendText]. A nil value sends a
// plain message.
type SendTextOptions struct {
	ReplyTo             *ReplyTarget
	Subject             string
	Effect              MessageEffect
	EnableDataDetection bool
	EnableLinkPreview   bool
	Formatting          []TextFormatInput
	ClientMessageID     string
}

// SendAttachmentOptions configures [MessageClient.SendAttachment].
type SendAttachmentOptions struct {
	ReplyTo         *ReplyTarget
	Effect          MessageEffect
	IsAudioMessage  bool
	ClientMessageID string
}

// SendMultipartOptions configures [MessageClient.SendMultipart].
type SendMultipartOptions struct {
	ReplyTo             *ReplyTarget
	Subject             string
	Effect              MessageEffect
	EnableDataDetection bool
	ClientMessageID     string
}

// EditOptions configures [MessageClient.Edit].
type EditOptions struct {
	// BackwardCompatText is shown to recipients on OS versions that cannot
	// render the edit.
	BackwardCompatText string
	PartIndex          *int
	ClientMessageID    string
}

// PartOptions configures part-targeted writes (unsend, set reaction, place
// sticker).
type PartOptions struct {
	PartIndex       *int
	ClientMessageID string
}

// CustomizedMiniAppMessage is an iMessage mini-app card backed by the caller's
// own extension. See [MessageClient.SendCustomizedMiniApp].
type CustomizedMiniAppMessage struct {
	AppName           string
	ExtensionBundleID string
	TeamID            string
	URL               string
	AppStoreID        *int64
	Layout            MiniAppLayout
}

// MiniAppLayout is the visual layout of a [CustomizedMiniAppMessage].
type MiniAppLayout struct {
	Caption            string
	Subcaption         string
	TrailingCaption    string
	TrailingSubcaption string
	ImageTitle         string
	ImageSubtitle      string
	Summary            string
	Image              []byte
}

// MessageListFilter filters and paginates message listings.
type MessageListFilter struct {
	Before    *time.Time
	After     *time.Time
	IsFromMe  *bool
	IsRead    *bool
	PageSize  int
	PageToken string
}

// MessageListPage is one page of a message listing.
type MessageListPage struct {
	Messages      []Message
	NextPageToken string
}
