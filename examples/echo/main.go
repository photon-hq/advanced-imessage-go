// Command echo is a minimal Advanced iMessage bot. It subscribes to the live
// message stream and replies to every inbound text message with the same text,
// sent as a threaded reply. It exercises the streaming and send APIs together.
//
// Authentication uses only your Spectrum project credentials: the shared
// spectrumauth helper exchanges the project ID and secret for a dedicated
// iMessage bearer token, derives the gRPC server address, and keeps the token
// fresh for the lifetime of the subscription.
//
// Usage:
//
//	export SPECTRUM_PROJECT_ID=your-project-id
//	export SPECTRUM_PROJECT_SECRET=your-project-secret
//	go run ./examples/echo                       # auto-selects the only line
//	go run ./examples/echo -line +15551234567    # pick a line by number
//	go run ./examples/echo -instance <id>        # or by instance id
//
// Use -prefix to prepend a marker to each echoed reply (for example
// -prefix="🔁 ").
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	imessage "github.com/photon-hq/advanced-imessage-go"
	"github.com/photon-hq/advanced-imessage-go/examples/internal/spectrumauth"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	projectID := flag.String("project", os.Getenv("SPECTRUM_PROJECT_ID"), "Spectrum project ID")
	projectSecret := flag.String("secret", os.Getenv("SPECTRUM_PROJECT_SECRET"), "Spectrum project secret")
	line := flag.String("line", "", "dedicated line to use, selected by phone number")
	instance := flag.String("instance", "", "dedicated line to use, selected by instance id")
	spectrumURL := flag.String("spectrum-url", spectrumauth.ProductionURL, "Spectrum REST API base URL")
	prefix := flag.String("prefix", "", "text prepended to each echoed reply")
	flag.Parse()

	if *projectID == "" {
		return errors.New("missing project ID: set -project or SPECTRUM_PROJECT_ID")
	}
	if *projectSecret == "" {
		return errors.New("missing project secret: set -secret or SPECTRUM_PROJECT_SECRET")
	}

	// A signal-cancelled context unblocks the stream for a clean shutdown and
	// also bounds the initial credential exchange.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	src, selected, err := spectrumauth.Connect(ctx, spectrumauth.Config{
		ProjectID:     *projectID,
		ProjectSecret: *projectSecret,
		BaseURL:       *spectrumURL,
		Line:          *line,
		Instance:      *instance,
	})
	if err != nil {
		return err
	}

	client, err := imessage.New(selected.Address, src,
		imessage.WithTimeout(30*time.Second), // bound each reply RPC
		imessage.WithRetry(),                 // retry server-retryable sends
		imessage.WithAutoIdempotency(),       // de-duplicate retried writes
	)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	sub := client.Messages().Subscribe(ctx, nil)
	defer func() { _ = sub.Close() }()

	log.Printf("echoing on line %s (instance %s) at %s — press Ctrl-C to stop",
		selected.Number, selected.InstanceID, selected.Address)

	for ev := range sub.Events() {
		// Only new messages are echoed; edits, reads, and reactions are ignored.
		recv, ok := ev.(imessage.MessageReceived)
		if !ok {
			continue
		}
		// Never reply to our own messages — that would loop forever.
		if recv.IsFromMe {
			continue
		}
		// Echo only ordinary text bubbles, skipping system rows and empty bodies.
		if recv.Message.ItemType != imessage.ItemNormal {
			continue
		}
		text := strings.TrimSpace(recv.Message.Content.Text)
		if text == "" {
			continue
		}

		if err := reply(ctx, client, recv, *prefix+text); err != nil {
			if imessage.IsRateLimited(err) {
				log.Printf("rate limited; backing off before the next reply")
				time.Sleep(2 * time.Second)
				continue
			}
			// One failed reply shouldn't take the bot down; log it and carry on.
			log.Printf("reply to %s failed: %v", recv.Message.GUID, err)
		}
	}

	// The loop ends when the stream stops. A cancelled context is the clean
	// Ctrl-C path; any other terminal error is worth surfacing.
	if err := sub.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	log.Print("shutting down")
	return nil
}

// reply echoes text into the originating chat as a threaded reply to the
// message that triggered it.
func reply(ctx context.Context, client *imessage.Client, recv imessage.MessageReceived, text string) error {
	chat, err := imessage.ParseChatGUID(recv.ChatGUID)
	if err != nil {
		return err
	}

	sent, err := client.Messages().SendText(ctx, chat, text, &imessage.SendTextOptions{
		ReplyTo: &imessage.ReplyTarget{GUID: recv.Message.GUID},
	})
	if err != nil {
		return err
	}

	from := "unknown sender"
	if recv.Message.Sender != nil {
		from = recv.Message.Sender.Address
	}
	log.Printf("echoed %q from %s in %s (reply %s)", text, from, chat, sent.GUID)
	return nil
}
