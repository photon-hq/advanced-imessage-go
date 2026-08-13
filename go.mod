module github.com/photon-hq/advanced-imessage-go

go 1.25.0

require (
	buf.build/gen/go/photon-hq/imessage/connectrpc/go v1.20.0-20260812120326-36bf2dfbbba4.1
	buf.build/gen/go/photon-hq/imessage/protocolbuffers/go v1.36.11-20260812120326-36bf2dfbbba4.1
	connectrpc.com/connect v1.20.0
	github.com/google/uuid v1.6.0
	go.uber.org/goleak v1.3.0
	golang.org/x/net v0.55.0
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260810153831-ec0a7760b754 // indirect
)
