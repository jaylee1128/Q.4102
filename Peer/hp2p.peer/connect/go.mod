module hp2p.go/connect

go 1.16

require (
    github.com/gorilla/websocket v1.4.2
    github.com/pion/webrtc/v3 v3.0.32 // indirect
    hp2p.util/util v0.0.0
    hp2p.pb/hp2p_pb v0.0.0
)

replace hp2p.util/util => ../../hp2p.util
replace hp2p.pb/hp2p_pb => ../../hp2p.pb