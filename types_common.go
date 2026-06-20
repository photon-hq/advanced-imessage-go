package imessage

// ChatServiceType identifies the transport a chat or address uses.
type ChatServiceType string

const (
	// ServiceIMessage is Apple iMessage.
	ServiceIMessage ChatServiceType = "iMessage"
	// ServiceSMS is carrier SMS/MMS.
	ServiceSMS ChatServiceType = "SMS"
	// ServiceRCS is RCS.
	ServiceRCS ChatServiceType = "RCS"
	// ServiceUnknown is an unrecognized or unspecified service.
	ServiceUnknown ChatServiceType = "unknown"
)

// Handle is an address (phone number or email) on a single service. It is the
// sender, participant, actor, or voter in the various responses and events.
type Handle struct {
	// Address is the phone number or email, for example "+15551234567" or
	// "person@example.com".
	Address string
	// Service is the service the address was seen on.
	Service ChatServiceType
	// Country is the ISO country code Apple associates with the address, or ""
	// if unknown.
	Country string
}

// AddressInfo is the result of looking up an address across every service it
// has been seen on.
type AddressInfo struct {
	// Address is the queried phone number or email.
	Address string
	// Services lists every service the address is reachable on.
	Services []ChatServiceType
	// Country is the ISO country code Apple associates with the address, or ""
	// if unknown.
	Country string
}
