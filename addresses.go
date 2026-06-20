package imessage

import (
	"context"

	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
	"connectrpc.com/connect"
)

// AddressClient looks up addresses. Obtain it from [Client.Addresses].
type AddressClient struct {
	svc imessagev1connect.AddressServiceClient
}

// Get returns the services an address has been seen on and its country.
func (a *AddressClient) Get(ctx context.Context, address string) (AddressInfo, error) {
	req := &imessagev1.GetAddressInfoRequest{}
	req.SetAddress(address)
	resp, err := a.svc.GetAddressInfo(ctx, connect.NewRequest(req))
	if err != nil {
		return AddressInfo{}, asError(err)
	}
	return addressInfoFromProto(resp.Msg.GetInfo()), nil
}

// IsFocusSilenced reports whether Messages currently treats the address as
// silenced by a Focus.
func (a *AddressClient) IsFocusSilenced(ctx context.Context, address string) (bool, error) {
	req := &imessagev1.GetFocusStatusRequest{}
	req.SetAddress(address)
	resp, err := a.svc.GetFocusStatus(ctx, connect.NewRequest(req))
	if err != nil {
		return false, asError(err)
	}
	return resp.Msg.GetIsSilencedByFocus(), nil
}

// IsIMessageAvailable reports whether the address is reachable via iMessage at
// query time.
func (a *AddressClient) IsIMessageAvailable(ctx context.Context, address string) (bool, error) {
	req := &imessagev1.GetIMessageAvailabilityRequest{}
	req.SetAddress(address)
	resp, err := a.svc.GetIMessageAvailability(ctx, connect.NewRequest(req))
	if err != nil {
		return false, asError(err)
	}
	return resp.Msg.GetIsAvailable(), nil
}
