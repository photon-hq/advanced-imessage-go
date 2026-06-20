package imessage

import (
	imessagev1 "buf.build/gen/go/photon-hq/imessage/protocolbuffers/go/photon/imessage/v1"
)

func sharedLocationFromProto(pb *imessagev1.SharedFriendLocation) SharedFriendLocation {
	if pb == nil {
		return SharedFriendLocation{}
	}
	loc := SharedFriendLocation{
		Address:              pb.GetAddress(),
		Name:                 pb.GetName(),
		LongAddress:          pb.GetLongAddress(),
		ShortAddress:         pb.GetShortAddress(),
		IsLocatingInProgress: pb.GetIsLocatingInProgress(),
		LocationType:         locationTypeFromProto(pb.GetLocationType()),
		LocationTimestamp:    timePtrFromProto(pb.GetLocationTimestamp()),
		ExpiresAt:            timePtrFromProto(pb.GetExpiresAt()),
	}
	if pb.HasLatitude() {
		loc.Latitude = ptrFloat(pb.GetLatitude())
	}
	if pb.HasLongitude() {
		loc.Longitude = ptrFloat(pb.GetLongitude())
	}
	if pb.HasAccuracy() {
		loc.Accuracy = ptrFloat(pb.GetAccuracy())
	}
	return loc
}

func locationUpdateFromProto(pb *imessagev1.WatchSharedFriendLocationsResponse) (LocationUpdate, bool, error) {
	if pb.WhichPayload() != imessagev1.WatchSharedFriendLocationsResponse_LocationUpdated_case {
		return LocationUpdate{}, false, nil
	}
	u := pb.GetLocationUpdated()
	return LocationUpdate{
		Location:       sharedLocationFromProto(u.GetLocation()),
		SourceSequence: u.GetSourceSequence(),
	}, true, nil
}
