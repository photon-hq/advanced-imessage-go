package imessage

// Poll is an iMessage poll with its options and current votes.
type Poll struct {
	PollMessageGUID string
	ChatGUID        string
	Title           string
	Options         []PollOption
	Votes           []PollParticipantVote
}

// PollOption is one choice in a [Poll].
type PollOption struct {
	// OptionID is the stable option identifier used when voting.
	OptionID string
	Text     string
	// CreatorHandle is the address that added the option, or "" if unknown.
	CreatorHandle string
}

// PollParticipantVote records one participant's vote.
type PollParticipantVote struct {
	Participant Handle
	OptionID    string
}

// PollChangeDelta is the specific change carried by a [PollChanged] event. Use
// a type switch over the variant structs below.
type PollChangeDelta interface{ isPollChangeDelta() }

// PollCreated reports a poll being created.
type PollCreated struct {
	Title   string
	Options []PollOption
}

// PollOptionAdded reports an option being added.
type PollOptionAdded struct {
	Title   string
	Options []PollOption
}

// PollVoted reports a vote being cast for an option.
type PollVoted struct{ OptionID string }

// PollUnvoted reports a vote being retracted from an option.
type PollUnvoted struct{ OptionID string }

func (PollCreated) isPollChangeDelta()     {}
func (PollOptionAdded) isPollChangeDelta() {}
func (PollVoted) isPollChangeDelta()       {}
func (PollUnvoted) isPollChangeDelta()     {}
