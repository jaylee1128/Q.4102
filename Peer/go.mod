module hp2p.go.peer

go 1.19

require (
	connect v0.0.0
	hp2p.util/util v0.0.0
)

require (
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pion/datachannel v1.4.21 // indirect
	github.com/pion/dtls/v2 v2.2.6 // indirect
	github.com/pion/ice/v2 v2.1.10 // indirect
	github.com/pion/interceptor v0.0.13 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.6 // indirect
	github.com/pion/rtp v1.6.5 // indirect
	github.com/pion/sctp v1.7.12 // indirect
	github.com/pion/sdp/v3 v3.0.4 // indirect
	github.com/pion/srtp/v2 v2.0.2 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.12.3 // indirect
	github.com/pion/transport/v2 v2.0.2 // indirect
	github.com/pion/turn/v2 v2.0.5 // indirect
	github.com/pion/udp/v2 v2.0.1 // indirect
	github.com/pion/webrtc/v3 v3.0.32 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	hp2p.pb/hp2p_pb v0.0.0 // indirect
)

replace connect => ./hp2p.peer/connect

replace hp2p.util/util => ./hp2p.util

replace hp2p.pb/hp2p_pb => ./hp2p.pb
