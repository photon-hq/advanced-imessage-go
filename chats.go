package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// ChatClient creates, inspects, and subscribes to chats. Obtain it from
// [Client.Chats].
type ChatClient struct {
	svc imessagev1connect.ChatServiceClient
}

// ChatCountOptions configures [ChatClient.Count].
type ChatCountOptions struct {
	// IncludeArchived counts archived chats as well.
	IncludeArchived bool
}

// Count returns the number of chats.
func (c *ChatClient) Count(ctx context.Context, opts *ChatCountOptions) (int, error) {
	req := &imessagev1.GetChatCountRequest{}
	if opts != nil && opts.IncludeArchived {
		req.SetIncludeArchived(true)
	}
	resp, err := c.svc.GetChatCount(ctx, connect.NewRequest(req))
	if err != nil {
		return 0, asError(err)
	}
	return int(resp.Msg.GetCount()), nil
}

// Create starts a chat with the given addresses, optionally sending an initial
// message via [CreateChatOptions]. The chat is created on iMessage.
func (c *ChatClient) Create(ctx context.Context, addresses []string, opts *CreateChatOptions) (CreateChatResult, error) {
	req := &imessagev1.CreateChatRequest{}
	req.SetAddresses(addresses)
	req.SetService(imessagev1.ChatServiceType_CHAT_SERVICE_TYPE_IMESSAGE)
	if opts != nil {
		if opts.Message != "" || len(opts.AttributedBody) > 0 || opts.Subject != "" || opts.Effect != "" {
			im := &imessagev1.CreateChatInitialMessage{}
			im.SetText(opts.Message)
			if len(opts.AttributedBody) > 0 {
				im.SetAttributedBody(opts.AttributedBody)
			}
			if opts.Effect != "" {
				im.SetEffectId(string(opts.Effect))
			}
			if opts.Subject != "" {
				im.SetSubject(opts.Subject)
			}
			req.SetInitialMessage(im)
		}
		if opts.ClientMessageID != "" {
			req.SetClientMessageId(opts.ClientMessageID)
		}
	}
	resp, err := c.svc.CreateChat(ctx, connect.NewRequest(req))
	if err != nil {
		return CreateChatResult{}, asError(err)
	}
	res := CreateChatResult{Chat: chatFromProto(resp.Msg.GetChat())}
	if resp.Msg.HasInitialMessage() {
		res.InitialMessage = messagePtrFromProto(resp.Msg.GetInitialMessage())
	}
	return res, nil
}

// Get returns a chat by GUID.
func (c *ChatClient) Get(ctx context.Context, chat ChatGUID) (Chat, error) {
	req := &imessagev1.GetChatRequest{}
	req.SetChatGuid(chat.String())
	resp, err := c.svc.GetChat(ctx, connect.NewRequest(req))
	if err != nil {
		return Chat{}, asError(err)
	}
	return chatFromProto(resp.Msg.GetChat()), nil
}

// HasBackground reports whether the chat has a custom background.
func (c *ChatClient) HasBackground(ctx context.Context, chat ChatGUID) (bool, error) {
	req := &imessagev1.HasBackgroundRequest{}
	req.SetChatGuid(chat.String())
	resp, err := c.svc.HasBackground(ctx, connect.NewRequest(req))
	if err != nil {
		return false, asError(err)
	}
	return resp.Msg.GetBackgroundPresent(), nil
}

// SetBackground sets the chat's custom background image.
func (c *ChatClient) SetBackground(ctx context.Context, chat ChatGUID, data []byte) error {
	req := &imessagev1.SetBackgroundRequest{}
	req.SetChatGuid(chat.String())
	req.SetData(data)
	_, err := c.svc.SetBackground(ctx, connect.NewRequest(req))
	return asError(err)
}

// RemoveBackground clears the chat's custom background.
func (c *ChatClient) RemoveBackground(ctx context.Context, chat ChatGUID) error {
	req := &imessagev1.RemoveBackgroundRequest{}
	req.SetChatGuid(chat.String())
	_, err := c.svc.RemoveBackground(ctx, connect.NewRequest(req))
	return asError(err)
}

// MarkRead marks every unread message in the chat as read.
func (c *ChatClient) MarkRead(ctx context.Context, chat ChatGUID) error {
	req := &imessagev1.MarkChatReadRequest{}
	req.SetChatGuid(chat.String())
	_, err := c.svc.MarkChatRead(ctx, connect.NewRequest(req))
	return asError(err)
}

// ShareContactInfo pushes the local user's name-and-photo card into the chat.
func (c *ChatClient) ShareContactInfo(ctx context.Context, chat ChatGUID) error {
	req := &imessagev1.ShareContactInfoRequest{}
	req.SetChatGuid(chat.String())
	_, err := c.svc.ShareContactInfo(ctx, connect.NewRequest(req))
	return asError(err)
}

// SetTyping sets or clears the local typing indicator in the chat.
func (c *ChatClient) SetTyping(ctx context.Context, chat ChatGUID, isTyping bool) error {
	req := &imessagev1.SetTypingRequest{}
	req.SetChatGuid(chat.String())
	req.SetIsTyping(isTyping)
	_, err := c.svc.SetTyping(ctx, connect.NewRequest(req))
	return asError(err)
}

// Subscribe opens a live stream of chat events. Always Close the returned
// subscription.
func (c *ChatClient) Subscribe(ctx context.Context, opts *SubscribeOptions) *ChatSubscription {
	req := &imessagev1.SubscribeChatEventsRequest{}
	if opts != nil && opts.Chat != "" {
		req.SetChatGuid(opts.Chat.String())
	}
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.SubscribeChatEventsResponse], error) {
			return c.svc.SubscribeChatEvents(ctx, connect.NewRequest(req))
		},
		chatEventFromProto,
	)
	return &ChatSubscription{stream[ChatEvent]{src: sub}}
}
