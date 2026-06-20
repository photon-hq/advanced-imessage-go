package imessage

// GroupIcon is a group chat's icon image.
type GroupIcon struct {
	Data     []byte
	MIMEType string
}

// GroupChange is the specific change carried by a [GroupChanged] event. Use a
// type switch over the variant structs below.
type GroupChange interface{ isGroupChange() }

// GroupDisplayNameChanged reports a group rename.
type GroupDisplayNameChanged struct{ DisplayName string }

// GroupParticipantAdded reports a participant being added.
type GroupParticipantAdded struct{ Participant Handle }

// GroupParticipantRemoved reports a participant being removed.
type GroupParticipantRemoved struct{ Participant Handle }

// GroupParticipantLeft reports a participant voluntarily leaving.
type GroupParticipantLeft struct{ Participant Handle }

// GroupIconChanged reports the group icon being set or replaced.
type GroupIconChanged struct{}

// GroupIconRemoved reports the group icon being removed.
type GroupIconRemoved struct{}

func (GroupDisplayNameChanged) isGroupChange() {}
func (GroupParticipantAdded) isGroupChange()   {}
func (GroupParticipantRemoved) isGroupChange() {}
func (GroupParticipantLeft) isGroupChange()    {}
func (GroupIconChanged) isGroupChange()        {}
func (GroupIconRemoved) isGroupChange()        {}
