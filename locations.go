package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// LocationClient reads and watches shared Find My locations. Obtain it from
// [Client.Locations].
type LocationClient struct {
	svc imessagev1connect.LocationServiceClient
}

// LocationWatchOptions filters a location watch to a single address.
type LocationWatchOptions struct {
	// Address, when non-empty, limits the watch to one friend.
	Address string
}

// List returns every friend location currently shared with the local user.
func (l *LocationClient) List(ctx context.Context) ([]SharedFriendLocation, error) {
	resp, err := l.svc.ListSharedFriendLocations(ctx, connect.NewRequest(&imessagev1.ListSharedFriendLocationsRequest{}))
	if err != nil {
		return nil, asError(err)
	}
	locs := resp.Msg.GetLocations()
	out := make([]SharedFriendLocation, 0, len(locs))
	for _, loc := range locs {
		out = append(out, sharedLocationFromProto(loc))
	}
	return out, nil
}

// Get returns the shared location for one address.
func (l *LocationClient) Get(ctx context.Context, address string) (SharedFriendLocation, error) {
	req := &imessagev1.GetSharedFriendLocationRequest{}
	req.SetAddress(address)
	resp, err := l.svc.GetSharedFriendLocation(ctx, connect.NewRequest(req))
	if err != nil {
		return SharedFriendLocation{}, asError(err)
	}
	return sharedLocationFromProto(resp.Msg.GetLocation()), nil
}

// Request asks address (within chat) to share their location. It returns the
// request acknowledgement.
func (l *LocationClient) Request(ctx context.Context, chat ChatGUID, address string, opts *IdempotencyOptions) (LocationRequestReceipt, error) {
	req := &imessagev1.RequestFriendLocationSharingRequest{}
	req.SetChatGuid(chat.String())
	req.SetAddress(address)
	applyClientMessageID(opts, req.SetClientMessageId)
	resp, err := l.svc.RequestFriendLocationSharing(ctx, connect.NewRequest(req))
	if err != nil {
		return LocationRequestReceipt{}, asError(err)
	}
	return LocationRequestReceipt{
		Address:     resp.Msg.GetAddress(),
		Status:      resp.Msg.GetStatus(),
		Reason:      resp.Msg.GetReason(),
		MessageGUID: resp.Msg.GetMessageGuid(),
	}, nil
}

// Watch opens a transient stream of location updates. These updates are not
// part of the durable event log. Always Close the returned watch.
func (l *LocationClient) Watch(ctx context.Context, opts *LocationWatchOptions) *LocationWatch {
	req := &imessagev1.WatchSharedFriendLocationsRequest{}
	if opts != nil && opts.Address != "" {
		req.SetAddress(opts.Address)
	}
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.WatchSharedFriendLocationsResponse], error) {
			return l.svc.WatchSharedFriendLocations(ctx, connect.NewRequest(req))
		},
		locationUpdateFromProto,
	)
	return &LocationWatch{stream[LocationUpdate]{src: sub}}
}
