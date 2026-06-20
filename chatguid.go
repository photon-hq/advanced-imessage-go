package imessage

import (
	"fmt"
	"strings"
)

// ChatGUID is a server-assigned chat identifier. It has the structured form
//
//	<service>;<kind>;<identifier>
//
// where kind is "-" for a direct (one-to-one) chat or "+" for a group chat, for
// example "iMessage;-;+15551234567" or "iMessage;+;chat0123...". Obtain one
// from [Chat.GUID] or a prior call's result; validate an externally sourced
// string with [ParseChatGUID].
type ChatGUID string

// ParseChatGUID validates s and returns it as a [ChatGUID]. It rejects bare
// addresses and any value that is not in <service>;<kind>;<identifier> form.
func ParseChatGUID(s string) (ChatGUID, error) {
	parts := strings.SplitN(s, ";", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("imessage: invalid chat guid %q: want <service>;<kind>;<identifier>", s)
	}
	if parts[0] == "" {
		return "", fmt.Errorf("imessage: invalid chat guid %q: empty service", s)
	}
	if parts[1] != "-" && parts[1] != "+" {
		return "", fmt.Errorf("imessage: invalid chat guid %q: kind must be %q or %q", s, "-", "+")
	}
	if parts[2] == "" {
		return "", fmt.Errorf("imessage: invalid chat guid %q: empty identifier", s)
	}
	return ChatGUID(s), nil
}

// String returns the GUID as a plain string.
func (c ChatGUID) String() string { return string(c) }

// IsGroup reports whether the GUID denotes a group chat (kind "+"). The result
// is only meaningful for a well-formed GUID; see [ParseChatGUID].
func (c ChatGUID) IsGroup() bool {
	parts := strings.SplitN(string(c), ";", 3)
	return len(parts) == 3 && parts[1] == "+"
}
