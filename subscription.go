package imessage

import "iter"

// seqSource adapts a caller-supplied iterator (and optional terminal error)
// into an [eventSource]. It runs no goroutine and holds no resources, so it is
// trivially leak-free and Close is a no-op. It backs the New*Subscription
// constructors below.
type seqSource[T any] struct {
	events iter.Seq[T]
	err    error
}

func (s seqSource[T]) Events() iter.Seq[T] {
	if s.events == nil {
		return func(func(T) bool) {}
	}
	return s.events
}

func (s seqSource[T]) Close() error { return nil }

func (s seqSource[T]) Err() error { return s.err }

// The New*Subscription constructors build a subscription that yields a fixed
// sequence of events instead of talking to a server. They let callers unit-test
// event-handling code without standing up a transport: pass any
// [iter.Seq] (for example slices.Values(events)) and range over Events as
// usual. Err reports nil and Close is a no-op.
//
//	sub := imessage.NewMessageSubscription(slices.Values([]imessage.MessageEvent{
//		imessage.MessageReceived{ /* ... */ },
//	}))
//	defer sub.Close()
//	for ev := range sub.Events() {
//		handle(ev) // the code under test
//	}

// NewMessageSubscription returns a [MessageSubscription] that yields events.
func NewMessageSubscription(events iter.Seq[MessageEvent]) *MessageSubscription {
	return &MessageSubscription{stream[MessageEvent]{src: seqSource[MessageEvent]{events: events}}}
}

// NewChatSubscription returns a [ChatSubscription] that yields events.
func NewChatSubscription(events iter.Seq[ChatEvent]) *ChatSubscription {
	return &ChatSubscription{stream[ChatEvent]{src: seqSource[ChatEvent]{events: events}}}
}

// NewGroupSubscription returns a [GroupSubscription] that yields events.
func NewGroupSubscription(events iter.Seq[GroupEvent]) *GroupSubscription {
	return &GroupSubscription{stream[GroupEvent]{src: seqSource[GroupEvent]{events: events}}}
}

// NewPollSubscription returns a [PollSubscription] that yields events.
func NewPollSubscription(events iter.Seq[PollEvent]) *PollSubscription {
	return &PollSubscription{stream[PollEvent]{src: seqSource[PollEvent]{events: events}}}
}

// NewCatchUpStream returns a [CatchUpStream] that yields events.
func NewCatchUpStream(events iter.Seq[CatchUpEvent]) *CatchUpStream {
	return &CatchUpStream{stream[CatchUpEvent]{src: seqSource[CatchUpEvent]{events: events}}}
}

// NewLocationWatch returns a [LocationWatch] that yields updates.
func NewLocationWatch(events iter.Seq[LocationUpdate]) *LocationWatch {
	return &LocationWatch{stream[LocationUpdate]{src: seqSource[LocationUpdate]{events: events}}}
}

// NewDownloadStream returns a [DownloadStream] that yields chunks.
func NewDownloadStream(events iter.Seq[DownloadChunk]) *DownloadStream {
	return &DownloadStream{stream[DownloadChunk]{src: seqSource[DownloadChunk]{events: events}}}
}

// Compile-time guard that seqSource satisfies the eventSource contract.
var _ eventSource[MessageEvent] = seqSource[MessageEvent]{}
