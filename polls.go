package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// PollClient creates and votes on iMessage polls. Obtain it from [Client.Polls].
type PollClient struct {
	svc imessagev1connect.PollServiceClient
}

// PollSubscribeOptions filters a poll subscription to a single poll.
type PollSubscribeOptions struct {
	// PollMessage, when non-empty, limits the stream to one poll.
	PollMessage string
}

// Create starts a poll in chat with the given title and choices.
func (p *PollClient) Create(ctx context.Context, chat ChatGUID, title string, choices []string, opts *IdempotencyOptions) (Poll, error) {
	req := &imessagev1.CreatePollRequest{}
	req.SetChatGuid(chat.String())
	req.SetTitle(title)
	req.SetOptions(choices)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := p.svc.CreatePoll(ctx, connect.NewRequest(req))
	if err != nil {
		return Poll{}, asError(err)
	}
	return pollFromProto(resp.Msg.GetPoll()), nil
}

// Get returns a poll by its poll-message GUID.
func (p *PollClient) Get(ctx context.Context, pollMessageGUID string) (Poll, error) {
	req := &imessagev1.GetPollRequest{}
	req.SetPollMessageGuid(pollMessageGUID)
	resp, err := p.svc.GetPoll(ctx, connect.NewRequest(req))
	if err != nil {
		return Poll{}, asError(err)
	}
	return pollFromProto(resp.Msg.GetPoll()), nil
}

// Vote casts the local user's vote for an option.
func (p *PollClient) Vote(ctx context.Context, pollMessageGUID, optionID string, opts *IdempotencyOptions) (Poll, error) {
	req := &imessagev1.VotePollRequest{}
	req.SetPollMessageGuid(pollMessageGUID)
	req.SetOptionIdentifier(optionID)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := p.svc.VotePoll(ctx, connect.NewRequest(req))
	if err != nil {
		return Poll{}, asError(err)
	}
	return pollFromProto(resp.Msg.GetPoll()), nil
}

// Unvote retracts the local user's vote.
func (p *PollClient) Unvote(ctx context.Context, pollMessageGUID string, opts *IdempotencyOptions) (Poll, error) {
	req := &imessagev1.UnvotePollRequest{}
	req.SetPollMessageGuid(pollMessageGUID)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := p.svc.UnvotePoll(ctx, connect.NewRequest(req))
	if err != nil {
		return Poll{}, asError(err)
	}
	return pollFromProto(resp.Msg.GetPoll()), nil
}

// AddOption adds a new option to an existing poll.
func (p *PollClient) AddOption(ctx context.Context, pollMessageGUID, text string, opts *IdempotencyOptions) (Poll, error) {
	req := &imessagev1.AddPollOptionRequest{}
	req.SetPollMessageGuid(pollMessageGUID)
	req.SetOptionText(text)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := p.svc.AddPollOption(ctx, connect.NewRequest(req))
	if err != nil {
		return Poll{}, asError(err)
	}
	return pollFromProto(resp.Msg.GetPoll()), nil
}

// Subscribe opens a live stream of poll events. Always Close the returned
// subscription.
func (p *PollClient) Subscribe(ctx context.Context, opts *PollSubscribeOptions) *PollSubscription {
	req := &imessagev1.SubscribePollEventsRequest{}
	if opts != nil && opts.PollMessage != "" {
		req.SetPollMessageGuid(opts.PollMessage)
	}
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.SubscribePollEventsResponse], error) {
			return p.svc.SubscribePollEvents(ctx, connect.NewRequest(req))
		},
		pollEventFromProto,
	)
	return &PollSubscription{stream[PollEvent]{src: sub}}
}
