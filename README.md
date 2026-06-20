# advanced-imessage-go

A Go client for the Advanced iMessage API (`photon.imessage.v1`) — an idiomatic
port of the TypeScript [`@photon-ai/advanced-imessage`](https://github.com/photon-hq/advanced-imessage-ts).
It speaks gRPC to the Photon backend and exposes a small, hand-written API over
the eight services: addresses, attachments, chats, events, groups, locations,
messages, and polls.

## Install

```sh
go get github.com/photon-hq/advanced-imessage-go
```

Requires **Go 1.25+** (the `connectrpc.com/connect` runtime sets the floor).

## Quick start

```go
package main

import (
	"context"
	"log"

	imessage "github.com/photon-hq/advanced-imessage-go"
)

func main() {
	client, err := imessage.New("imsg.example.com:443", imessage.StaticToken("api-token"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	chat, err := imessage.ParseChatGUID("any;-;+15551234567")
	if err != nil {
		log.Fatal(err)
	}

	msg, err := client.Messages().SendText(context.Background(), chat, "Hello from Go!", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sent", msg.GUID)
}
```

The client is safe for concurrent use. Every resource is reached through an
accessor (`client.Messages()`, `client.Chats()`, …); every method takes a
`context.Context` first and an optional, nil-able options struct last.

### Options

`New` takes functional options:

- `WithoutTLS()` — connect over cleartext HTTP/2 (local development).
- `WithTimeout(d)` — default deadline for unary RPCs that lack one.
- `WithRetry(...)` — retry server-marked-retryable unary calls with jittered backoff.
- `WithAutoIdempotency()` — attach a generated idempotency key to mutating calls.
- `WithHTTPClient(hc)` — supply your own HTTP/2 client.

Authentication is a `TokenSource`: use `StaticToken` for a fixed key or
`TokenFunc` for rotating credentials.

## Streaming

Subscribe, watch, catch-up, and download operations return a stream you consume
with a range-over-func iterator. Always `Close` it (a deferred `Close` is
idiomatic) and check `Err` after the loop:

```go
sub := client.Messages().Subscribe(ctx, nil)
defer sub.Close()

for ev := range sub.Events() {
	switch e := ev.(type) {
	case imessage.MessageReceived:
		log.Printf("new message %s in %s", e.Message.GUID, e.ChatGUID)
	case imessage.MessageReactionAdded:
		log.Printf("%s reacted with %s", e.MessageGUID, e.Reaction.Kind)
	}
}
if err := sub.Err(); err != nil {
	log.Printf("stream ended: %v", err)
}
```

Events are sealed sum types — handle them with a type switch.

### Resuming across disconnects

Live subscriptions do not reconnect on their own. To resume gap-free, persist
the last delivered sequence and replay with `client.Events().CatchUp(seq)`
before reopening a live subscription — or let `ResumableMessages` do it for you:

```go
// store implements imessage.SequenceStore (durable, your responsibility).
sub := client.ResumableMessages(ctx, store, nil)
defer sub.Close()
for ev := range sub.Events() {
	// delivery is at-least-once; make handling idempotent
}
```

## Errors

Every failure is an `*imessage.Error`. Match categories with `errors.Is` against
the package sentinels or the `Is*` helpers, and read structured detail with
`errors.As`:

```go
if imessage.IsRateLimited(err) {
	// back off
}

var ierr *imessage.Error
if errors.As(err, &ierr) {
	log.Printf("code=%s retryable=%v", ierr.Code, ierr.Retryable)
}
```

## Development

```sh
make check   # tidy + vet + lint + race tests
make test    # go test ./...
```

Run the cross-language message interop suite against the TypeScript SDK with:

```sh
make test-interop
```

That target installs `@photon-ai/advanced-imessage` into `testdata/interop-js`,
starts an in-memory Go gRPC server, runs both SDKs through the same message
operations and event streams, and compares the protobuf requests they emit.

The generated Connect SDK comes from the Buf Schema Registry
(`buf.build/photon-hq/imessage`); bump it with `make update-sdk`.

## License

[MIT](LICENSE).
