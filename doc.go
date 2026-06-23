// Package imessage is a Go client for the Advanced iMessage API
// (photon.imessage.v1), a faithful and idiomatic port of the TypeScript
// library @photon-ai/advanced-imessage.
//
// # Getting started
//
// Create a [Client] with a server address and a [TokenSource], then reach the
// eight resource groups through their accessor methods:
//
//	client, err := imessage.New("imsg.example.com:443", imessage.StaticToken(token))
//	if err != nil {
//		return err
//	}
//	defer client.Close()
//
//	msg, err := client.Messages().SendText(ctx, chat, "hello", nil)
//
// # Resources
//
// The API surface is grouped exactly as the backend services are:
// [Client.Addresses], [Client.Attachments], [Client.Chats], [Client.Events],
// [Client.Groups], [Client.Locations], [Client.Messages], and [Client.Polls].
// Each accessor returns a concrete resource client whose methods take a
// [context.Context] first and an optional, nil-able options struct last.
//
// # Streaming
//
// The Subscribe, Watch, CatchUp, and Download operations return a stream whose
// events are consumed with a range-over-func iterator. Always call Close (a
// deferred Close is idiomatic), and check Err after the loop:
//
//	sub := client.Messages().Subscribe(ctx, nil)
//	defer sub.Close()
//	for ev := range sub.Events() {
//		switch e := ev.(type) {
//		case imessage.MessageReceived:
//			// ...
//		}
//	}
//	if err := sub.Err(); err != nil {
//		// terminal stream error
//	}
//
// Live subscriptions do not reconnect on their own. To resume gap-free after a
// disconnect, persist the last delivered sequence and replay with
// [Client.Events] CatchUp, or use the resumable helpers (see [SequenceStore]).
//
// # Errors
//
// Every RPC failure is reported as an [*Error]. Match categories with
// [errors.Is] against the package sentinels (for example [ErrNotFound]) or the
// Is* helper functions, and read the structured detail with [errors.As].
//
// # Testing
//
// To test code that calls the client, run an in-process Connect server and
// point [New] at it with [WithoutTLS] plus [WithHTTPClient] (an httptest server
// with unencrypted HTTP/2 enabled works well), mounting fakes built on the
// generated Unimplemented*ServiceHandler types.
//
// To unit-test event-handling code without any transport, build a subscription
// from a fixed iterator with the New*Subscription constructors (for example
// [NewMessageSubscription]) and range over it exactly as you would a live one.
package imessage
