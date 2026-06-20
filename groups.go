package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// GroupClient administers group chats. Obtain it from [Client.Groups].
type GroupClient struct {
	svc imessagev1connect.GroupServiceClient
}

// SetDisplayName renames a group chat and returns the updated chat.
func (g *GroupClient) SetDisplayName(ctx context.Context, chat ChatGUID, displayName string, opts *IdempotencyOptions) (Chat, error) {
	req := &imessagev1.SetDisplayNameRequest{}
	req.SetChatGuid(chat.String())
	req.SetDisplayName(displayName)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := g.svc.SetDisplayName(ctx, connect.NewRequest(req))
	if err != nil {
		return Chat{}, asError(err)
	}
	return chatFromProto(resp.Msg.GetChat()), nil
}

// AddParticipants adds addresses to a group chat and returns the updated chat.
func (g *GroupClient) AddParticipants(ctx context.Context, chat ChatGUID, addresses []string, opts *IdempotencyOptions) (Chat, error) {
	req := &imessagev1.AddParticipantsRequest{}
	req.SetChatGuid(chat.String())
	req.SetAddresses(addresses)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := g.svc.AddParticipants(ctx, connect.NewRequest(req))
	if err != nil {
		return Chat{}, asError(err)
	}
	return chatFromProto(resp.Msg.GetChat()), nil
}

// RemoveParticipants removes addresses from a group chat and returns the
// updated chat.
func (g *GroupClient) RemoveParticipants(ctx context.Context, chat ChatGUID, addresses []string, opts *IdempotencyOptions) (Chat, error) {
	req := &imessagev1.RemoveParticipantsRequest{}
	req.SetChatGuid(chat.String())
	req.SetAddresses(addresses)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := g.svc.RemoveParticipants(ctx, connect.NewRequest(req))
	if err != nil {
		return Chat{}, asError(err)
	}
	return chatFromProto(resp.Msg.GetChat()), nil
}

// Leave makes the local user leave a group chat.
func (g *GroupClient) Leave(ctx context.Context, chat ChatGUID, opts *IdempotencyOptions) error {
	req := &imessagev1.LeaveGroupRequest{}
	req.SetChatGuid(chat.String())
	applyClientMessageID(opts, req.SetClientMessageId)
	_, err := g.svc.LeaveGroup(ctx, connect.NewRequest(req))
	return asError(err)
}

// SetIcon sets a group chat's icon image.
func (g *GroupClient) SetIcon(ctx context.Context, chat ChatGUID, data []byte, opts *IdempotencyOptions) error {
	req := &imessagev1.SetIconRequest{}
	req.SetChatGuid(chat.String())
	req.SetData(data)
	applyClientMessageID(opts, req.SetClientMessageId)
	_, err := g.svc.SetIcon(ctx, connect.NewRequest(req))
	return asError(err)
}

// GetIcon returns a group chat's icon image.
func (g *GroupClient) GetIcon(ctx context.Context, chat ChatGUID) (GroupIcon, error) {
	req := &imessagev1.GetIconRequest{}
	req.SetChatGuid(chat.String())
	resp, err := g.svc.GetIcon(ctx, connect.NewRequest(req))
	if err != nil {
		return GroupIcon{}, asError(err)
	}
	return GroupIcon{Data: resp.Msg.GetData(), MIMEType: resp.Msg.GetMimeType()}, nil
}

// RemoveIcon removes a group chat's icon.
func (g *GroupClient) RemoveIcon(ctx context.Context, chat ChatGUID, opts *IdempotencyOptions) error {
	req := &imessagev1.RemoveIconRequest{}
	req.SetChatGuid(chat.String())
	applyClientMessageID(opts, req.SetClientMessageId)
	_, err := g.svc.RemoveIcon(ctx, connect.NewRequest(req))
	return asError(err)
}

// Subscribe opens a live stream of group events. Always Close the returned
// subscription.
func (g *GroupClient) Subscribe(ctx context.Context, opts *SubscribeOptions) *GroupSubscription {
	req := &imessagev1.SubscribeGroupEventsRequest{}
	if opts != nil && opts.Chat != "" {
		req.SetChatGuid(opts.Chat.String())
	}
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.SubscribeGroupEventsResponse], error) {
			return g.svc.SubscribeGroupEvents(ctx, connect.NewRequest(req))
		},
		groupEventFromProto,
	)
	return &GroupSubscription{stream[GroupEvent]{src: sub}}
}

// applyClientMessageID sets the idempotency key from opts, if present.
func applyClientMessageID(opts *IdempotencyOptions, set func(string)) {
	if opts != nil && opts.ClientMessageID != "" {
		set(opts.ClientMessageID)
	}
}
