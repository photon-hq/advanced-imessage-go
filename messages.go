package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// MessageClient sends, edits, lists, and subscribes to messages. Obtain it from
// [Client.Messages].
type MessageClient struct {
	svc imessagev1connect.MessageServiceClient
}

// SubscribeOptions filters a live event subscription to a single chat. A nil
// value (or zero Chat) subscribes to every chat.
type SubscribeOptions struct {
	// Chat, when non-empty, limits the stream to one chat.
	Chat ChatGUID
}

// SendText sends a text (or rich-text) message to chat.
func (m *MessageClient) SendText(ctx context.Context, chat ChatGUID, text string, opts *SendTextOptions) (Message, error) {
	req := &imessagev1.SendTextMessageRequest{}
	req.SetChatGuid(chat.String())
	req.SetText(text)
	if opts != nil {
		if opts.ReplyTo != nil {
			req.SetReplyTo(replyTargetToProto(opts.ReplyTo))
		}
		if opts.Subject != "" {
			req.SetSubject(opts.Subject)
		}
		if opts.Effect != "" {
			req.SetEffectId(string(opts.Effect))
		}
		if opts.EnableDataDetection {
			req.SetEnableDataDetection(true)
		}
		if opts.EnableLinkPreview {
			req.SetEnableLinkPreview(true)
		}
		if len(opts.Formatting) > 0 {
			req.SetFormatting(textFormatInputsToProto(opts.Formatting))
		}
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	resp, err := m.svc.SendTextMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// SendAttachment sends a previously uploaded attachment to chat.
func (m *MessageClient) SendAttachment(ctx context.Context, chat ChatGUID, attachmentGUID string, opts *SendAttachmentOptions) (Message, error) {
	ref := &imessagev1.AttachmentRef{}
	ref.SetAttachmentGuid(attachmentGUID)
	req := &imessagev1.SendAttachmentMessageRequest{}
	req.SetChatGuid(chat.String())
	req.SetAttachment(ref)
	if opts != nil {
		if opts.ReplyTo != nil {
			req.SetReplyTo(replyTargetToProto(opts.ReplyTo))
		}
		if opts.Effect != "" {
			req.SetEffectId(string(opts.Effect))
		}
		if opts.IsAudioMessage {
			req.SetIsAudioMessage(true)
		}
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	resp, err := m.svc.SendAttachmentMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// SendMultipart sends a message composed of several bubble parts to chat.
func (m *MessageClient) SendMultipart(ctx context.Context, chat ChatGUID, parts []MessagePart, opts *SendMultipartOptions) (Message, error) {
	req := &imessagev1.SendMultipartMessageRequest{}
	req.SetChatGuid(chat.String())
	protoParts := make([]*imessagev1.MessagePart, 0, len(parts))
	for _, p := range parts {
		protoParts = append(protoParts, messagePartToProto(p))
	}
	req.SetParts(protoParts)
	if opts != nil {
		if opts.ReplyTo != nil {
			req.SetReplyTo(replyTargetToProto(opts.ReplyTo))
		}
		if opts.Subject != "" {
			req.SetSubject(opts.Subject)
		}
		if opts.Effect != "" {
			req.SetEffectId(string(opts.Effect))
		}
		if opts.EnableDataDetection {
			req.SetEnableDataDetection(true)
		}
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	resp, err := m.svc.SendMultipartMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// SendCustomizedMiniApp sends an iMessage mini-app card backed by the caller's
// own extension.
func (m *MessageClient) SendCustomizedMiniApp(ctx context.Context, chat ChatGUID, message CustomizedMiniAppMessage, opts *IdempotencyOptions) (Message, error) {
	layout := &imessagev1.MiniAppLayout{}
	if message.Layout.Caption != "" {
		layout.SetCaption(message.Layout.Caption)
	}
	if message.Layout.Subcaption != "" {
		layout.SetSubcaption(message.Layout.Subcaption)
	}
	if message.Layout.TrailingCaption != "" {
		layout.SetTrailingCaption(message.Layout.TrailingCaption)
	}
	if message.Layout.TrailingSubcaption != "" {
		layout.SetTrailingSubcaption(message.Layout.TrailingSubcaption)
	}
	if message.Layout.ImageTitle != "" {
		layout.SetImageTitle(message.Layout.ImageTitle)
	}
	if message.Layout.ImageSubtitle != "" {
		layout.SetImageSubtitle(message.Layout.ImageSubtitle)
	}
	if message.Layout.Summary != "" {
		layout.SetSummary(message.Layout.Summary)
	}
	if len(message.Layout.Image) > 0 {
		layout.SetImage(message.Layout.Image)
	}

	req := &imessagev1.SendCustomizedMiniAppMessageRequest{}
	req.SetChatGuid(chat.String())
	req.SetTeamId(message.TeamID)
	req.SetExtensionBundleId(message.ExtensionBundleID)
	req.SetAppName(message.AppName)
	req.SetUrl(message.URL)
	req.SetLayout(layout)
	if message.AppStoreID != nil {
		req.SetAppStoreId(*message.AppStoreID)
	}
	if opts != nil && opts.ClientMessageID != "" {
		req.SetClientMessageId(opts.ClientMessageID)
	}
	resp, err := m.svc.SendCustomizedMiniAppMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// Edit replaces the text of a sent message.
func (m *MessageClient) Edit(ctx context.Context, chat ChatGUID, messageGUID, newText string, opts *EditOptions) (Message, error) {
	var partIndex *int
	req := &imessagev1.EditMessageRequest{}
	if opts != nil {
		partIndex = opts.PartIndex
		if opts.BackwardCompatText != "" {
			req.SetBackwardCompatText(opts.BackwardCompatText)
		}
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	req.SetTarget(messageTarget(chat, messageGUID, partIndex))
	req.SetNewText(newText)
	resp, err := m.svc.EditMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// Unsend retracts a sent message.
func (m *MessageClient) Unsend(ctx context.Context, chat ChatGUID, messageGUID string, opts *PartOptions) error {
	var partIndex *int
	req := &imessagev1.UnsendMessageRequest{}
	if opts != nil {
		partIndex = opts.PartIndex
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	req.SetTarget(messageTarget(chat, messageGUID, partIndex))
	_, err := m.svc.UnsendMessage(ctx, connect.NewRequest(req))
	return asError(err)
}

// SetReaction adds (isSet=true) or removes (isSet=false) a tapback on a message.
func (m *MessageClient) SetReaction(ctx context.Context, chat ChatGUID, messageGUID string, reaction Reaction, isSet bool, opts *PartOptions) (Message, error) {
	r := &imessagev1.MessageReaction{}
	r.SetKind(reactionKindToProto(reaction.Kind))
	if reaction.Emoji != "" {
		r.SetEmoji(reaction.Emoji)
	}
	var partIndex *int
	req := &imessagev1.SetReactionRequest{}
	if opts != nil {
		partIndex = opts.PartIndex
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	req.SetTarget(messageTarget(chat, messageGUID, partIndex))
	req.SetReaction(r)
	req.SetIsSet(isSet)
	resp, err := m.svc.SetReaction(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// PlaceSticker places a sticker attachment on a message.
func (m *MessageClient) PlaceSticker(ctx context.Context, chat ChatGUID, messageGUID, stickerGUID string, placement StickerPlacement, opts *PartOptions) (Message, error) {
	ref := &imessagev1.AttachmentRef{}
	ref.SetAttachmentGuid(stickerGUID)
	var partIndex *int
	req := &imessagev1.PlaceStickerRequest{}
	if opts != nil {
		partIndex = opts.PartIndex
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	req.SetTarget(messageTarget(chat, messageGUID, partIndex))
	req.SetSticker(ref)
	req.SetPlacement(stickerPlacementToProto(placement))
	resp, err := m.svc.PlaceSticker(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// NotifySilenced triggers Apple's per-message "Notify Anyway" action.
func (m *MessageClient) NotifySilenced(ctx context.Context, chat ChatGUID, messageGUID string, opts *IdempotencyOptions) error {
	req := &imessagev1.NotifySilencedMessageRequest{}
	req.SetChatGuid(chat.String())
	req.SetMessageGuid(messageGUID)
	if opts != nil && opts.ClientMessageID != "" {
		req.SetClientMessageId(opts.ClientMessageID)
	}
	_, err := m.svc.NotifySilencedMessage(ctx, connect.NewRequest(req))
	return asError(err)
}

// Get returns a single message by GUID.
func (m *MessageClient) Get(ctx context.Context, messageGUID string) (Message, error) {
	req := &imessagev1.GetMessageRequest{}
	req.SetMessageGuid(messageGUID)
	resp, err := m.svc.GetMessage(ctx, connect.NewRequest(req))
	if err != nil {
		return Message{}, asError(err)
	}
	return messageFromProto(resp.Msg.GetMessage()), nil
}

// ListRecent lists recent messages across all chats.
func (m *MessageClient) ListRecent(ctx context.Context, opts *MessageListFilter) (MessageListPage, error) {
	req := &imessagev1.ListRecentMessagesRequest{}
	applyListFilter(req, opts)
	resp, err := m.svc.ListRecentMessages(ctx, connect.NewRequest(req))
	if err != nil {
		return MessageListPage{}, asError(err)
	}
	return messageListPage(resp.Msg), nil
}

// ListInChat lists messages within a single chat.
func (m *MessageClient) ListInChat(ctx context.Context, chat ChatGUID, opts *MessageListFilter) (MessageListPage, error) {
	req := &imessagev1.ListChatMessagesRequest{}
	req.SetChatGuid(chat.String())
	applyListFilter(req, opts)
	resp, err := m.svc.ListChatMessages(ctx, connect.NewRequest(req))
	if err != nil {
		return MessageListPage{}, asError(err)
	}
	return messageListPage(resp.Msg), nil
}

// GetEmbeddedMedia returns decoded media embedded in a message.
func (m *MessageClient) GetEmbeddedMedia(ctx context.Context, chat ChatGUID, messageGUID string) (EmbeddedMedia, error) {
	req := &imessagev1.GetEmbeddedMediaRequest{}
	req.SetChatGuid(chat.String())
	req.SetMessageGuid(messageGUID)
	resp, err := m.svc.GetEmbeddedMedia(ctx, connect.NewRequest(req))
	if err != nil {
		return EmbeddedMedia{}, asError(err)
	}
	media := resp.Msg.GetMedia()
	return EmbeddedMedia{Data: media.GetData(), MIMEType: media.GetMimeType()}, nil
}

// Subscribe opens a live stream of message events. Always Close the returned
// subscription.
func (m *MessageClient) Subscribe(ctx context.Context, opts *SubscribeOptions) *MessageSubscription {
	req := &imessagev1.SubscribeMessageEventsRequest{}
	if opts != nil && opts.Chat != "" {
		req.SetChatGuid(opts.Chat.String())
	}
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.SubscribeMessageEventsResponse], error) {
			return m.svc.SubscribeMessageEvents(ctx, connect.NewRequest(req))
		},
		messageEventFromProto,
	)
	return &MessageSubscription{stream[MessageEvent]{src: sub}}
}

// listFilterSetter is satisfied by both ListRecentMessagesRequest and
// ListChatMessagesRequest, so one helper can apply a filter to either.
type listFilterSetter interface {
	SetPageSize(int32)
	SetPageToken(string)
	SetIsFromMe(bool)
	SetIsRead(bool)
	SetBefore(*timestamppb.Timestamp)
	SetAfter(*timestamppb.Timestamp)
}

func applyListFilter(req listFilterSetter, f *MessageListFilter) {
	if f == nil {
		return
	}
	if f.PageSize > 0 {
		req.SetPageSize(int32(f.PageSize))
	}
	if f.PageToken != "" {
		req.SetPageToken(f.PageToken)
	}
	if f.IsFromMe != nil {
		req.SetIsFromMe(*f.IsFromMe)
	}
	if f.IsRead != nil {
		req.SetIsRead(*f.IsRead)
	}
	if f.Before != nil {
		req.SetBefore(timestamppb.New(*f.Before))
	}
	if f.After != nil {
		req.SetAfter(timestamppb.New(*f.After))
	}
}

// listMessagesResp is satisfied by both list responses.
type listMessagesResp interface {
	GetMessages() []*imessagev1.Message
	GetNextPageToken() string
}

func messageListPage(resp listMessagesResp) MessageListPage {
	page := MessageListPage{NextPageToken: resp.GetNextPageToken()}
	for _, m := range resp.GetMessages() {
		page.Messages = append(page.Messages, messageFromProto(m))
	}
	return page
}
