package imessage

import "time"

// LocationType classifies the precision/source of a shared location.
type LocationType string

const (
	// LocationLegacy is a legacy Find My location.
	LocationLegacy LocationType = "legacy"
	// LocationLive is a continuously updating live location.
	LocationLive LocationType = "live"
	// LocationShallow is a coarse one-shot location.
	LocationShallow LocationType = "shallow"
	// LocationUnknown is an unrecognized location type.
	LocationUnknown LocationType = "unknown"
)

// SharedFriendLocation is a friend's shared Find My location.
type SharedFriendLocation struct {
	Address string
	// Name is the friend's display name, or "" if unknown.
	Name string
	// Latitude, Longitude, and Accuracy are nil when the coordinate is
	// unavailable (for example while locating is still in progress).
	Latitude  *float64
	Longitude *float64
	Accuracy  *float64
	// LongAddress and ShortAddress are reverse-geocoded address strings.
	LongAddress          string
	ShortAddress         string
	IsLocatingInProgress bool
	LocationType         LocationType
	LocationTimestamp    *time.Time
	ExpiresAt            *time.Time
}

// LocationUpdate is one frame of a [LocationClient.Watch] stream.
type LocationUpdate struct {
	Location SharedFriendLocation
	// SourceSequence is the watch stream's own sequence; it is independent of
	// the durable event log and may contain gaps.
	SourceSequence uint64
}

// LocationRequestReceipt is the acknowledgement of a location-sharing request.
type LocationRequestReceipt struct {
	Address string
	Status  string
	// Reason is an optional explanation when Status is not a success.
	Reason string
	// MessageGUID is the request card's message GUID, if one was created.
	MessageGUID string
}
