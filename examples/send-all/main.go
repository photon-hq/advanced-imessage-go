// Command send-all exercises every Advanced iMessage send and mutation API
// against one chat, printing a pass/fail/skip summary. It is a smoke test for
// the full surface: text, rich text, multipart, attachments, edits, reactions,
// stickers, unsend, notify-silenced, polls, and (optionally) mini-apps.
//
// Authentication uses only your Spectrum project credentials (see the shared
// spectrumauth helper).
//
// Usage:
//
//	export SPECTRUM_PROJECT_ID=your-project-id
//	export SPECTRUM_PROJECT_SECRET=your-project-secret
//	go run ./examples/send-all -chat "iMessage;-;+15551234567"
//
// Select a specific dedicated line with -line/-instance. Enable the mini-app
// step by passing -miniapp-team, -miniapp-bundle, and -miniapp-url.
//
// WARNING: this actually sends messages to the target chat.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/signal"
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
	chatGUID := flag.String("chat", "", "target chat GUID, e.g. \"iMessage;-;+15551234567\" (required)")
	lineSel := flag.String("line", "", "dedicated line to use, selected by phone number")
	instanceSel := flag.String("instance", "", "dedicated line to use, selected by instance id")
	spectrumURL := flag.String("spectrum-url", spectrumauth.ProductionURL, "Spectrum REST API base URL")
	miniAppTeam := flag.String("miniapp-team", "", "Apple Team ID for the mini-app step")
	miniAppBundle := flag.String("miniapp-bundle", "", "Messages extension bundle ID for the mini-app step")
	miniAppURL := flag.String("miniapp-url", "", "URL for the mini-app step")
	flag.Parse()

	if *projectID == "" {
		return errors.New("missing project ID: set -project or SPECTRUM_PROJECT_ID")
	}
	if *projectSecret == "" {
		return errors.New("missing project secret: set -secret or SPECTRUM_PROJECT_SECRET")
	}
	if *chatGUID == "" {
		return errors.New("missing target chat: set -chat")
	}
	chat, err := imessage.ParseChatGUID(*chatGUID)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	src, selected, err := spectrumauth.Connect(ctx, spectrumauth.Config{
		ProjectID:     *projectID,
		ProjectSecret: *projectSecret,
		BaseURL:       *spectrumURL,
		Line:          *lineSel,
		Instance:      *instanceSel,
	})
	if err != nil {
		return err
	}

	client, err := imessage.New(selected.Address, src,
		imessage.WithTimeout(30*time.Second),
		imessage.WithRetry(),
		imessage.WithAutoIdempotency(),
	)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	log.Printf("sending all message types on line %s to %s", selected.Number, chat)

	r := &recorder{}
	msgs := client.Messages()

	// 1. Plain text. Its GUID anchors the edit/reaction/sticker steps below.
	anchor, err := msgs.SendText(ctx, chat, "1/ plain text from send-all", nil)
	r.record("SendText (plain)", err)
	anchorGUID := anchor.GUID

	// 2. Rich text: threaded reply + subject + bubble effect + inline formatting
	//    + link preview + data detection.
	richOpts := &imessage.SendTextOptions{
		Subject:             "send-all subject",
		Effect:              imessage.EffectConfetti,
		EnableDataDetection: true,
		EnableLinkPreview:   true,
		Formatting: []imessage.TextFormatInput{
			{Kind: imessage.FormatBold, Start: 0, Length: 2},
			{Kind: imessage.FormatEffect, Start: 3, Length: 4, Effect: imessage.TextEffectBig},
		},
	}
	if anchorGUID != "" {
		richOpts.ReplyTo = &imessage.ReplyTarget{GUID: anchorGUID}
	}
	_, err = msgs.SendText(ctx, chat, "2/ rich reply https://photon.codes", richOpts)
	r.record("SendText (rich/reply/effect)", err)

	// 3. Upload an attachment — prerequisite for the attachment, multipart, and
	//    sticker steps.
	var attachmentGUID string
	pixel, perr := pngPixel()
	if perr != nil {
		r.record("Attachments.Upload", perr)
	} else {
		res, uerr := client.Attachments().Upload(ctx, imessage.AttachmentInput{
			FileName: "pixel.png",
			Data:     pixel,
		})
		r.record("Attachments.Upload", uerr)
		if uerr == nil {
			attachmentGUID = res.Attachment.GUID
		}
	}

	// 4. Send that attachment.
	if attachmentGUID != "" {
		_, err = msgs.SendAttachment(ctx, chat, attachmentGUID, nil)
		r.record("SendAttachment", err)
	} else {
		r.skip("SendAttachment", "attachment upload failed")
	}

	// 5. Multipart: two text bubbles, plus an attachment bubble when available.
	parts := []imessage.MessagePart{
		{Text: "5/ multipart bubble one", BubbleIndex: intp(0)},
		{Text: "multipart bubble two", BubbleIndex: intp(1)},
	}
	if attachmentGUID != "" {
		parts = append(parts, imessage.MessagePart{
			AttachmentGUID: attachmentGUID,
			AttachmentName: "pixel.png",
			BubbleIndex:    intp(2),
		})
	}
	_, err = msgs.SendMultipart(ctx, chat, parts, nil)
	r.record("SendMultipart", err)

	// 6. Edit the anchor.
	if anchorGUID != "" {
		_, err = msgs.Edit(ctx, chat, anchorGUID, "1/ plain text (edited)", &imessage.EditOptions{
			BackwardCompatText: "edited",
		})
		r.record("Edit", err)
	} else {
		r.skip("Edit", "no anchor message")
	}

	// 7. React to the anchor with a few tapback kinds (including a custom emoji),
	//    then remove one.
	if anchorGUID != "" {
		var aerr error
		for _, reaction := range []imessage.Reaction{
			{Kind: imessage.ReactionLove},
			{Kind: imessage.ReactionLike},
			{Kind: imessage.ReactionEmoji, Emoji: "🎉"},
		} {
			if _, e := msgs.AddReaction(ctx, chat, anchorGUID, reaction, nil); e != nil {
				aerr = e
				break
			}
		}
		r.record("AddReaction (love/like/emoji)", aerr)

		_, rerr := msgs.RemoveReaction(ctx, chat, anchorGUID, imessage.Reaction{Kind: imessage.ReactionLike}, nil)
		r.record("RemoveReaction (like)", rerr)
	} else {
		r.skip("AddReaction", "no anchor message")
		r.skip("RemoveReaction", "no anchor message")
	}

	// 8. Place the uploaded image as a sticker on the anchor.
	if anchorGUID != "" && attachmentGUID != "" {
		_, err = msgs.PlaceSticker(ctx, chat, anchorGUID, imessage.StickerInput{
			GUID:      attachmentGUID,
			Placement: imessage.StickerPlacement{X: 0.5, Y: 0.5},
		}, nil)
		r.record("PlaceSticker", err)
	} else {
		r.skip("PlaceSticker", "need anchor message + attachment")
	}

	// 9. Unsend a throwaway message.
	throwaway, terr := msgs.SendText(ctx, chat, "9/ throwaway to be unsent", nil)
	if terr != nil {
		r.record("Unsend", terr)
	} else {
		r.record("Unsend", msgs.Unsend(ctx, chat, throwaway.GUID, nil))
	}

	// 10. Notify-silenced. Often errors unless the chat is actually silenced, so
	//     a failure here is informational rather than a real defect.
	if anchorGUID != "" {
		r.record("NotifySilenced", msgs.NotifySilenced(ctx, chat, anchorGUID, nil))
	} else {
		r.skip("NotifySilenced", "no anchor message")
	}

	// 11. Poll lifecycle: create, vote, add an option, unvote.
	runPollSteps(ctx, client, chat, r)

	// 12. Customized mini-app, only when the extension config is supplied.
	if *miniAppTeam != "" && *miniAppBundle != "" && *miniAppURL != "" {
		_, err = msgs.SendCustomizedMiniApp(ctx, chat, imessage.CustomizedMiniAppMessage{
			AppName:           "send-all",
			ExtensionBundleID: *miniAppBundle,
			TeamID:            *miniAppTeam,
			URL:               *miniAppURL,
			Layout:            imessage.MiniAppLayout{Caption: "send-all", Subcaption: "mini app"},
		}, nil)
		r.record("SendCustomizedMiniApp", err)
	} else {
		r.skip("SendCustomizedMiniApp", "set -miniapp-team/-miniapp-bundle/-miniapp-url to enable")
	}

	return r.summary()
}

// runPollSteps creates a poll and walks it through vote, add-option, and unvote.
func runPollSteps(ctx context.Context, client *imessage.Client, chat imessage.ChatGUID, r *recorder) {
	polls := client.Polls()
	poll, err := polls.Create(ctx, chat, "send-all poll", []string{"Alpha", "Beta"}, nil)
	r.record("Polls.Create", err)
	if err != nil {
		r.skip("Polls.Vote", "poll create failed")
		r.skip("Polls.AddOption", "poll create failed")
		r.skip("Polls.Unvote", "poll create failed")
		return
	}

	if len(poll.Options) > 0 {
		_, verr := polls.Vote(ctx, poll.PollMessageGUID, poll.Options[0].OptionID, nil)
		r.record("Polls.Vote", verr)
	} else {
		r.skip("Polls.Vote", "poll has no options")
	}

	_, aerr := polls.AddOption(ctx, poll.PollMessageGUID, "Gamma", nil)
	r.record("Polls.AddOption", aerr)

	_, uerr := polls.Unvote(ctx, poll.PollMessageGUID, nil)
	r.record("Polls.Unvote", uerr)
}

// recorder accumulates per-step outcomes and prints a final summary.
type recorder struct {
	ok      int
	failed  int
	skipped int
}

func (r *recorder) record(name string, err error) {
	if err != nil {
		r.failed++
		log.Printf("  ✗ %s: %v", name, err)
		return
	}
	r.ok++
	log.Printf("  ✓ %s", name)
}

func (r *recorder) skip(name, why string) {
	r.skipped++
	log.Printf("  – %s (skipped: %s)", name, why)
}

func (r *recorder) summary() error {
	log.Printf("done: %d ok, %d failed, %d skipped", r.ok, r.failed, r.skipped)
	if r.failed > 0 {
		return fmt.Errorf("%d message type(s) failed", r.failed)
	}
	return nil
}

// pngPixel returns a 1x1 PNG to use as the uploaded attachment and sticker.
func pngPixel() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0x34, G: 0x98, B: 0xdb, A: 0xff})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func intp(i int) *int { return &i }
