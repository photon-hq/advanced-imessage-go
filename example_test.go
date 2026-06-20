package imessage_test

import (
	"context"
	"errors"
	"fmt"
	"log"

	imessage "github.com/photon-hq/advanced-imessage-go"
)

// Example shows creating a client and sending a message.
func Example() {
	client, err := imessage.New("imsg.example.com:443", imessage.StaticToken("api-token"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	chat, err := imessage.ParseChatGUID("iMessage;-;+15551234567")
	if err != nil {
		log.Print(err)
		return
	}

	msg, err := client.Messages().SendText(context.Background(), chat, "Hello from Go!", nil)
	if err != nil {
		if imessage.IsRateLimited(err) {
			log.Println("slow down")
		}
		log.Print(err)
		return
	}
	fmt.Println(msg.GUID)
}

// Example_subscribe shows consuming a live message stream.
func Example_subscribe() {
	client, err := imessage.New("imsg.example.com:443", imessage.StaticToken("api-token"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sub := client.Messages().Subscribe(context.Background(), nil)
	defer sub.Close()

	for ev := range sub.Events() {
		switch e := ev.(type) {
		case imessage.MessageReceived:
			fmt.Printf("new message %s in %s\n", e.Message.GUID, e.ChatGUID)
		case imessage.MessageReactionAdded:
			fmt.Printf("%s reacted with %s\n", e.MessageGUID, e.Reaction.Kind)
		}
	}
	if err := sub.Err(); err != nil {
		log.Printf("stream ended: %v", err)
	}
}

// Example_resumable shows a gap-free subscription backed by a sequence store.
func Example_resumable() {
	client, err := imessage.New("imsg.example.com:443", imessage.StaticToken("api-token"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var store imessage.SequenceStore // supply your own durable implementation

	sub := client.ResumableMessages(context.Background(), store, nil)
	defer sub.Close()

	for ev := range sub.Events() {
		if recv, ok := ev.(imessage.MessageReceived); ok {
			fmt.Println(recv.Message.GUID)
		}
	}
	if err := sub.Err(); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("stream ended: %v", err)
	}
}
