package imessage

import "iter"

// eventSource is the common contract behind every public subscription: a direct
// transport stream or the resumable engine.
type eventSource[T any] interface {
	Events() iter.Seq[T]
	Err() error
	Close() error
}

// stream is the common core of every public subscription type. It adapts an
// event source, translating its terminal error to an [*Error].
type stream[T any] struct {
	src eventSource[T]
}

// Events returns a single-use iterator over the stream's events. Range over it,
// then check [stream.Err]. Breaking out of the loop releases the RPC.
func (s *stream[T]) Events() iter.Seq[T] { return s.src.Events() }

// Err returns the terminal error that stopped the stream, or nil if it ended
// cleanly. Call it after ranging over Events.
func (s *stream[T]) Err() error { return asError(s.src.Err()) }

// Close cancels the stream and blocks until its goroutine has exited. It is
// safe to call more than once; deferring it is idiomatic.
func (s *stream[T]) Close() error { return s.src.Close() }

// MessageSubscription is a live stream of [MessageEvent] values.
type MessageSubscription struct{ stream[MessageEvent] }

// ChatSubscription is a live stream of [ChatEvent] values.
type ChatSubscription struct{ stream[ChatEvent] }

// GroupSubscription is a live stream of [GroupEvent] values.
type GroupSubscription struct{ stream[GroupEvent] }

// PollSubscription is a live stream of [PollEvent] values.
type PollSubscription struct{ stream[PollEvent] }

// CatchUpStream is a finite replay of [CatchUpEvent] values that ends with a
// [CatchupComplete].
type CatchUpStream struct{ stream[CatchUpEvent] }

// LocationWatch is a transient stream of [LocationUpdate] values.
type LocationWatch struct{ stream[LocationUpdate] }

// DownloadStream is a stream of [DownloadChunk] frames for an attachment
// download.
type DownloadStream struct{ stream[DownloadChunk] }
