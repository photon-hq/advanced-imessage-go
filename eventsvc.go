package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// EventClient replays the durable event log. Obtain it from [Client.Events].
type EventClient struct {
	svc imessagev1connect.EventServiceClient
}

// CatchUp replays durable events after afterSequence (pass 0 to replay from the
// beginning). The stream delivers each historical event and ends with a
// [CatchupComplete] carrying the head sequence; open a live subscription after
// that sequence for gap-free history-then-live consumption. Always Close the
// returned stream.
func (e *EventClient) CatchUp(ctx context.Context, afterSequence uint64) *CatchUpStream {
	req := &imessagev1.CatchUpEventsRequest{}
	if afterSequence > 0 {
		req.SetAfterSequence(afterSequence)
	}
	sub := transport.Subscribe(ctx,
		func(ctx context.Context) (*connect.ServerStreamForClient[imessagev1.CatchUpEventsResponse], error) {
			return e.svc.CatchUpEvents(ctx, connect.NewRequest(req))
		},
		catchUpEventFromProto,
	)
	return &CatchUpStream{stream[CatchUpEvent]{src: sub}}
}
