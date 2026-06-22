module github.com/photon-hq/advanced-imessage-go

go 1.24.0

require (
	buf.build/gen/go/photon-hq/imessage/connectrpc/go v1.19.2-20260620033527-059c960d2c7a.1
	buf.build/gen/go/photon-hq/imessage/protocolbuffers/go v1.36.11-20260620033527-059c960d2c7a.1
	connectrpc.com/connect v1.19.2
	github.com/google/uuid v1.6.0
	go.uber.org/goleak v1.3.0
	golang.org/x/net v0.48.0
	google.golang.org/protobuf v1.36.11
)

require golang.org/x/text v0.32.0 // indirect
