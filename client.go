package imessage

import (
	imessagev1connect "buf.build/gen/go/photon-hq/imessage/connectrpc/go/photon/imessage/v1/imessagev1connect"
	"connectrpc.com/connect"

	"github.com/photon-hq/advanced-imessage-go/internal/transport"
)

// Client is a connection to the Advanced iMessage API. It is safe for
// concurrent use by multiple goroutines. Create one with [New] and release it
// with [Client.Close].
type Client struct {
	httpClient connect.HTTPClient

	messages    *MessageClient
	chats       *ChatClient
	groups      *GroupClient
	attachments *AttachmentClient
	addresses   *AddressClient
	locations   *LocationClient
	polls       *PollClient
	events      *EventClient
}

// New dials the Advanced iMessage server at address and authenticates every
// request with src. The address is a host:port; TLS is used unless
// [WithoutTLS] is supplied. Call [Client.Close] to release resources.
func New(address string, src TokenSource, opts ...Option) (*Client, error) {
	if address == "" {
		return nil, &Error{Code: CodeInvalidArgument, ConnectCode: connect.CodeInvalidArgument, Message: "address is required"}
	}
	if src == nil {
		return nil, &Error{Code: CodeInvalidArgument, ConnectCode: connect.CodeInvalidArgument, Message: "token source is required"}
	}

	cfg := defaultConfig()
	for _, o := range opts {
		o.apply(&cfg)
	}

	hc, baseURL, copts := transport.Dial(transport.Config{
		Address:         address,
		Insecure:        cfg.insecure,
		Timeout:         cfg.timeout,
		RetryEnabled:    cfg.retryEnabled,
		Retry:           transport.RetryConfig(cfg.retry),
		AutoIdempotency: cfg.autoIdempotency,
		HTTPClient:      cfg.httpClient,
		Token:           src.Token,
	})

	c := &Client{httpClient: hc}
	c.messages = &MessageClient{svc: imessagev1connect.NewMessageServiceClient(hc, baseURL, copts...)}
	c.chats = &ChatClient{svc: imessagev1connect.NewChatServiceClient(hc, baseURL, copts...)}
	c.groups = &GroupClient{svc: imessagev1connect.NewGroupServiceClient(hc, baseURL, copts...)}
	c.attachments = &AttachmentClient{svc: imessagev1connect.NewAttachmentServiceClient(hc, baseURL, copts...)}
	c.addresses = &AddressClient{svc: imessagev1connect.NewAddressServiceClient(hc, baseURL, copts...)}
	c.locations = &LocationClient{svc: imessagev1connect.NewLocationServiceClient(hc, baseURL, copts...)}
	c.polls = &PollClient{svc: imessagev1connect.NewPollServiceClient(hc, baseURL, copts...)}
	c.events = &EventClient{svc: imessagev1connect.NewEventServiceClient(hc, baseURL, copts...)}
	return c, nil
}

// Close releases idle connections held by the client. It is safe to call once
// the client is no longer needed; in-flight calls are not interrupted.
func (c *Client) Close() error {
	type idleCloser interface{ CloseIdleConnections() }
	if ic, ok := c.httpClient.(idleCloser); ok {
		ic.CloseIdleConnections()
	}
	return nil
}

// Messages returns the message resource client.
func (c *Client) Messages() *MessageClient { return c.messages }

// Chats returns the chat resource client.
func (c *Client) Chats() *ChatClient { return c.chats }

// Groups returns the group resource client.
func (c *Client) Groups() *GroupClient { return c.groups }

// Attachments returns the attachment resource client.
func (c *Client) Attachments() *AttachmentClient { return c.attachments }

// Addresses returns the address resource client.
func (c *Client) Addresses() *AddressClient { return c.addresses }

// Locations returns the location resource client.
func (c *Client) Locations() *LocationClient { return c.locations }

// Polls returns the poll resource client.
func (c *Client) Polls() *PollClient { return c.polls }

// Events returns the durable event-log resource client.
func (c *Client) Events() *EventClient { return c.events }
