//
// The MIT License
//
// Copyright (c) 2022 ETRI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package util

import (
	"crypto/rsa"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	pwebrtc "github.com/pion/webrtc/v3"
)

type ClientConfig struct {
	OverlayServerAddr   string `json:"OVERLAY_SERVER_ADDR"`
	SignalingServerAddr string `json:"SIGNALING_SERVER_ADDR"`
	//UdpPort4IoT         int    `json:"IOT_UDP_PORT"`
	LogLevel  int  `json:"LOG_LEVEL"`
	FileLog   bool `json:"FILE_LOG"`
	StdoutLog bool `json:"STDOUT_LOG"`
}

func ReadClientConfig(path string) ClientConfig {
	config := ClientConfig{}
	config.OverlayServerAddr = "http://127.0.0.1:8081"
	config.SignalingServerAddr = "ws://127.0.0.1:8082"
	config.LogLevel = 2
	config.FileLog = false
	config.StdoutLog = true

	if len(path) > 0 && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	file, err := os.ReadFile(path + "clientconfig.json")
	if err == nil {
		_ = json.Unmarshal([]byte(file), &config)
	}

	return config
}

type PeerConfig struct {
	PeerEstabMaxCount            int  `json:"PEER_ESTAB_MAX_COUNT"`
	PeerEstabTimeout             int  `json:"PEER_ESTAB_TIMEOUT"`
	MaxPrimaryConnection         int  `json:"MAX_PRIMARY_CONNECTION"`
	MaxIncomingCandidate         int  `json:"MAX_INCOMING_CANDIDATE"`
	MaxOutgoingCandidate         int  `json:"MAX_OUTGOING_CANDIDATE"`
	HelloPeerTTL                 int  `json:"HELLO_PEER_TTL"`
	EstabPeerTimeout             int  `json:"ESTAB_PEER_TIMEOUT"`
	EstabPeerMaxCount            int  `json:"ESTAB_PEER_MAX_COUNT"`
	ProbePeerTimeout             int  `json:"PROBE_PEER_TIMEOUT"`
	Expires                      int  `json:"PEER_EXPIRES"`
	SendCandidate                bool `json:"SEND_CANDIDATE"`
	ReleaseOperationAck          bool `json:"RELEASE_OPERATION_ACK"`
	BroadcastOperationAck        bool `json:"BROADCAST_OPERATION_ACK"`
	RetryOverlayJoin             bool `json:"RETRY_OVERLAY_JOIN"`
	RetryOverlayJoinCount        int  `json:"RETRY_OVERLAY_JOIN_COUNT"`
	RetryOverlayJoinInterval     int  `json:"RETRY_OVERLAY_JOIN_INTERVAL"`
	RetryOverlayRecoveryInterval int  `json:"RETRY_OVERLAY_RECOVERY_INTERVAL"`
}

func ReadPeerConfig(path string) PeerConfig {
	config := PeerConfig{}
	config.PeerEstabMaxCount = 5
	config.PeerEstabTimeout = 5
	config.MaxPrimaryConnection = 2
	config.MaxIncomingCandidate = 30
	config.MaxOutgoingCandidate = 3
	config.HelloPeerTTL = 3
	config.EstabPeerTimeout = 5
	config.EstabPeerMaxCount = 3
	config.ProbePeerTimeout = 0
	config.Expires = 30
	config.SendCandidate = false
	config.ReleaseOperationAck = false
	config.BroadcastOperationAck = false
	config.RetryOverlayJoin = true
	config.RetryOverlayJoinCount = 5
	config.RetryOverlayJoinInterval = 5
	config.RetryOverlayRecoveryInterval = 10

	if len(path) > 0 && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	file, err := os.ReadFile(path + "peerconfig.json")
	if err == nil {
		_ = json.Unmarshal([]byte(file), &config)
	}

	PrintJson(INFO, "PeerConfig", config)

	return config
}

type CrPolicy struct {
	MNCache    int    `json:"mN_Cache"`
	MDCache    int    `json:"mD_Cache"`
	RecoveryBy string `json:"recovery-by"`
}

type ChannelAttributeVideoFeature struct {
	Target        string `json:"target"`
	Mode          string `json:"mode"`
	Resolution    string `json:"resolution"`
	FrameRate     string `json:"frame-rate"`
	KeypointsType string `json:"keypoints-type"`
	ModelUri      string `json:"model-uri"`
	Hash          string `json:"hash"`
	Dimension     string `json:"dimension"`
}

type ChannelAttributeAudio struct {
	Codec       string `json:"codec"`
	SampleRate  string `json:"sample-rate"`
	Bitrate     string `json:"bitrate"`
	ChannelType string `json:"channel-type"`
}

type ChannelAttributeText struct {
	Format   string `json:"format"`
	Encoding string `json:"encoding"`
}

type ChannelInfo struct {
	ChannelId        string       `json:"channel-id"`
	ChannelType      string       `json:"channel-type"`
	ChannelAttribute *interface{} `json:"channel-attribute,omitempty"`
	SourceList       *[]string    `json:"source-list,omitempty"`
}

type ServiceInfo struct {
	StartDatetime *string       `json:"start-datetime,omitempty"`
	EndDatetime   *string       `json:"end-datetime,omitempty"`
	Title         *string       `json:"title,omitempty"`
	Description   *string       `json:"description,omitempty"`
	SourceList    *[]string     `json:"source-list,omitempty"`
	BlockList     *[]string     `json:"block-list,omitempty"`
	ChannelList   []ChannelInfo `json:"channel-list"`
}

type HybridOverlayCreationOverlay struct {
	Title             string      `json:"title"`
	Type              string      `json:"type"`
	SubType           string      `json:"sub-type"`
	OwnerId           string      `json:"owner-id"`
	Description       string      `json:"description,omitempty"`
	HeartbeatInterval int         `json:"heartbeat-interval"`
	HeartbeatTimeout  int         `json:"heartbeat-timeout"`
	Auth              OverlayAuth `json:"auth"`
	CrPolicy          *CrPolicy   `json:"cr-policy,omitempty"`
	ServiceInfo       ServiceInfo `json:"service-info"`
}

type HybridOverlayCreation struct {
	Overlay HybridOverlayCreationOverlay `json:"overlay"`
}

type HybridOverlayCreationResponse struct {
	OverlayInfo HybridOverlayCreationResponseOverlayInfo `json:"overlay"`
}

type HybridOverlayCreationResponseOverlayInfo struct {
	OverlayId         string       `json:"overlay-id"`
	Type              string       `json:"type"`
	SubType           string       `json:"sub-type"`
	OwnerId           string       `json:"owner-id"`
	HeartbeatInterval int          `json:"heartbeat-interval"`
	HeartbeatTimeout  int          `json:"heartbeat-timeout"`
	Auth              *OverlayAuth `json:"auth,omitempty"`
	CrPolicy          *CrPolicy    `json:"cr-policy,omitempty"`
	ServiceInfo       ServiceInfo  `json:"service-info"`
}

type HybridOverlayJoinOverlay struct {
	OverlayId string       `json:"overlay-id"`
	Type      string       `json:"type"`
	SubType   string       `json:"sub-type"`
	Auth      *OverlayAuth `json:"auth,omitempty"`
}

type HybridOverlayJoinPeer struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64    `json:"instance-id"`
	Address     string   `json:"address"`
	Auth        PeerAuth `json:"auth"`
	Expires     *int     `json:"expires,omitempty"`
	TicketId    *int     `json:"ticket-id,omitempty"`
	DisplayName *string  `json:"display-name,omitempty"`
}

type HybridOverlayJoin struct {
	Overlay HybridOverlayJoinOverlay `json:"overlay"`
	Peer    HybridOverlayJoinPeer    `json:"peer"`
}

type HybridOverlayJoinResponseOverlay struct {
	OverlayId         string        `json:"overlay-id"`
	Type              string        `json:"type"`
	SubType           string        `json:"sub-type"`
	Status            OverlayStatus `json:"status"`
	HeartbeatInterval int           `json:"heartbeat-interval"`
	HeartbeatTimeout  int           `json:"heartbeat-timeout"`
	CrPolicy          *CrPolicy     `json:"cr-policy,omitempty"`
	ServiceInfo       ServiceInfo   `json:"service-info"`
}

type HybridOverlayJoinResponsePeer struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64  `json:"instance-id"`
	TicketId    int     `json:"ticket-id"`
	Expires     int     `json:"expires"`
	DisplayName *string `json:"display-name,omitempty"`
}

type HybridOverlayJoinResponse struct {
	RspCode int                              `json:"rsp-code"`
	Overlay HybridOverlayJoinResponseOverlay `json:"overlay"`
	Peer    HybridOverlayJoinResponsePeer    `json:"peer"`
}

type HybridOverlayModificationOverlay struct {
	OverlayId   string       `json:"overlay-id"`
	Title       *string      `json:"title,omitempty"`
	OwnerId     string       `json:"owner-id"`
	Description *string      `json:"description,omitempty"`
	Auth        OverlayAuth  `json:"auth"`
	ServiceInfo *ServiceInfo `json:"service-info,omitempty"`
}

type HybridOverlayModificationOwnership struct {
	OwnerId  *string `json:"owner-id,omitempty"`
	AdminKey *string `json:"admin-key,omitempty"`
}

type HybridOverlayModification struct {
	RspCode   int                                 `json:"rsp-code"`
	Overlay   HybridOverlayModificationOverlay    `json:"overlay"`
	Ownership *HybridOverlayModificationOwnership `json:"ownership,omitempty"`
}

type HybridOverlayRemovalOverlay struct {
	OverlayId string      `json:"overlay-id"`
	OwnerId   string      `json:"owner-id"`
	Auth      OverlayAuth `json:"auth"`
}

type HybridOverlayRemoval struct {
	Overlay HybridOverlayRemovalOverlay `json:"overlay"`
}

type HybridOverlayRemovalResponseOverlay struct {
	OverlayId string `json:"overlay-id"`
}

type HybridOverlayRemovalResponse struct {
	RspCode int                                 `json:"rsp-code"`
	Overlay HybridOverlayRemovalResponseOverlay `json:"overlay"`
}

type OverlayInfo struct {
	OverlayId         string      `json:"overlay-id"`
	Title             string      `json:"title"`
	Type              string      `json:"type"`
	SubType           string      `json:"sub-type"`
	OwnerId           string      `json:"owner-id"`
	Description       string      `json:"description"`
	HeartbeatInterval int         `json:"heartbeat-interval"`
	HeartbeatTimeout  int         `json:"heartbeat-timeout"`
	Auth              OverlayAuth `json:"auth"`
	CrPolicy          *CrPolicy   `json:"cr-policy"`
	//TicketId          int           `json:"ticket-id"`
	Status      OverlayStatus `json:"status"`
	ServiceInfo ServiceInfo   `json:"service-info"`
}

func (self *OverlayInfo) Copy(v interface{}) {
	switch v.(type) {
	case *HybridOverlayCreationOverlay:
		val := v.(*HybridOverlayCreationOverlay)

		if val == nil {
			return
		}

		self.Title = val.Title
		self.Type = val.Type
		self.SubType = val.SubType
		self.OwnerId = val.OwnerId
		if strings.Contains(self.OwnerId, ";") {
			self.OwnerId = strings.Split(self.OwnerId, ";")[0]
		}
		self.Description = val.Description
		self.HeartbeatInterval = val.HeartbeatInterval
		self.HeartbeatTimeout = val.HeartbeatTimeout
		self.Auth = val.Auth
		self.CrPolicy = val.CrPolicy
		self.ServiceInfo = val.ServiceInfo
	case *HybridOverlayCreationResponseOverlayInfo:
		val := v.(*HybridOverlayCreationResponseOverlayInfo)

		if val == nil {
			return
		}

		self.ServiceInfo = val.ServiceInfo
		self.OverlayId = val.OverlayId
	case *HybridOverlayJoinResponseOverlay:
		val := v.(*HybridOverlayJoinResponseOverlay)

		if val == nil {
			return
		}

		self.HeartbeatInterval = val.HeartbeatInterval
		self.HeartbeatTimeout = val.HeartbeatTimeout
		self.Status = val.Status
		self.ServiceInfo = val.ServiceInfo

	case *HybridOverlayQueryResponseOverlay:
		val := v.(*HybridOverlayQueryResponseOverlay)

		if val == nil {
			return
		}

		self.OverlayId = val.OverlayId
		if val.ServiceInfo.Title != nil && len(*val.ServiceInfo.Title) > 0 {
			self.Title = *val.ServiceInfo.Title
		} else {
			self.Title = val.Title
		}
		if val.ServiceInfo.Description != nil && len(*val.ServiceInfo.Description) > 0 {
			self.Description = *val.ServiceInfo.Description
		} else if val.Description != nil {
			self.Description = *val.Description
		}

		self.Type = val.Type
		self.SubType = val.SubType
		self.OwnerId = val.OwnerId
		//remove after ";" in owner-id
		if strings.Contains(self.OwnerId, ";") {
			self.OwnerId = strings.Split(self.OwnerId, ";")[0]
		}
		self.Status = val.Status
		self.Auth = val.Auth
		self.CrPolicy = val.CrPolicy
		self.ServiceInfo = val.ServiceInfo
	}
}

type PeerInfo struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64    `json:"instance-id,omitempty"`
	Address     string   `json:"address"`
	Auth        PeerAuth `json:"auth"`
	TicketId    int      `json:"ticket-id"`
	DisplayName *string  `json:"display-name,omitempty"`
}

type JoinPeerInfo struct {
	PeerId       string `json:"peer-id"`
	DisplayName  string `json:"display-name"`
	PublicKeyPEM []byte `json:"public-key"`
	PublicKey    *rsa.PublicKey
	CachingMedia []JoinPeerInfoCachingMedia `json:"caching-media"`
}

type JoinPeerInfoCachingMedia struct {
	ChannelId string `json:"channel-id"`
	Sequence  int    `json:"sequence"`
	Media     []byte `json:"media"`
	SignMedia []byte `json:"sign-media"`
}

type OverlayAuth struct {
	Type      string    `json:"type"`
	AdminKey  string    `json:"admin_key,omitempty"`
	AccessKey *string   `json:"access_key,omitempty"`
	PeerList  *[]string `json:"peerlist,omitempty"`
}

type OverlayStatus struct {
	NumPeers     int        `json:"num_peers"`
	PeerInfoList []PeerInfo `json:"peer_info_list,omitempty"`
	Status       string     `json:"status"`
}

type HybridOverlayQueryResponseOverlay struct {
	OverlayId   string        `json:"overlay-id"`
	Title       string        `json:"title"`
	Type        string        `json:"type"`
	SubType     string        `json:"sub-type"`
	OwnerId     string        `json:"owner-id"`
	Status      OverlayStatus `json:"status"`
	Description *string       `json:"description,omitempty"`
	Auth        OverlayAuth   `json:"auth"`
	CrPolicy    *CrPolicy     `json:"cr-policy,omitempty"`
	ServiceInfo ServiceInfo   `json:"service-info"`
}

type HybridOverlayQueryResponse struct {
	Overlay []HybridOverlayQueryResponseOverlay `json:"overlay"`
}

type HybridOverlayReportOverlay struct {
	OverlayId string `json:"overlay-id"`
}

type HybridOverlayReportPeer struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64    `json:"instance-id"`
	Auth PeerAuth `json:"auth"`
}

type HybridOverlayReport struct {
	Overlay HybridOverlayReportOverlay `json:"overlay"`
	Peer    HybridOverlayReportPeer    `json:"peer"`
	Status  PeerStatus                 `json:"status"`
}

type HybridOverlayReportResponse struct {
	Overlay HybridOverlayReportOverlay `json:"overlay"`
}

type HybridOverlayRefreshOverlay struct {
	OverlayId string       `json:"overlay-id"`
	Auth      *OverlayAuth `json:"auth,omitempty"`
}

type HybridOverlayRefreshPeer struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64    `json:"instance-id"`
	Address string   `json:"address"`
	Auth    PeerAuth `json:"auth"`
	Expires *int     `json:"expires,omitempty"`
}

type HybridOverlayRefresh struct {
	Overlay HybridOverlayRefreshOverlay `json:"overlay"`
	Peer    HybridOverlayRefreshPeer    `json:"peer"`
}

type HybridOverlayRefreshResponsePeer struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64  `json:"instance-id"`
	Expires int `json:"expires"`
}

type HybridOverlayRefreshResponse struct {
	Overlay HybridOverlayRefreshOverlay      `json:"overlay"`
	Peer    HybridOverlayRefreshResponsePeer `json:"peer"`
}

type PeerStatus struct {
	NumPrimary      int     `json:"num_primary"`
	NumOutCandidate int     `json:"num_out_candidate"`
	NumInCandidate  int     `json:"num_in_candidate"`
	CostMap         CostMap `json:"costmap"`
}

type CostMap struct {
	Primary           []string `json:"primary"`
	OutgoingCandidate []string `json:"outgoing_candidate"`
	IncomingCandidate []string `json:"incoming_candidate"`
}

type PeerAuth struct {
	Password  string `json:"password"`
	PublicKey string `json:"public-key"`
}

type HybridOverlayLeaveOverlay struct {
	OverlayId string       `json:"overlay-id"`
	Auth      *OverlayAuth `json:"auth,omitempty"`
}

type HybridOverlayLeavePeer struct {
	PeerId string `json:"peer-id"`
	//InstanceId int64     `json:"instance-id"`
	Auth *PeerAuth `json:"auth,omitempty"`
}

type HybridOverlayLeave struct {
	Overlay HybridOverlayLeaveOverlay `json:"overlay"`
	Peer    HybridOverlayLeavePeer    `json:"peer"`
}

type HybridOverlayLeaveResponseOverlay struct {
	OverlayId string `json:"overlay-id"`
}

type HybridOverlayLeaveResponse struct {
	RspCode int                               `json:"rsp-code"`
	Overlay HybridOverlayLeaveResponseOverlay `json:"overlay"`
}

type HybridOverlayUserQuery struct {
	OverlayId string `json:"overlay-id"`
	PeerId    string `json:"peer-id"`
}

type HybridOverlayUserQueryResponsePeer struct {
	PeerId      string    `json:"peer-id"`
	Auth        *PeerAuth `json:"auth,omitempty"`
	DisplayName *string   `json:"display-name,omitempty"`
}
type HybridOverlayUserQueryResponse struct {
	RspCode int                                `json:"rsp-code"`
	Peer    HybridOverlayUserQueryResponsePeer `json:"peer"`
}

type RTCSessionDescription struct {
	Fromid       string                     `json:"fromid"`
	Toid         string                     `json:"toid"`
	Fromticketid string                     `json:"fromticketid"`
	Sdp          pwebrtc.SessionDescription `json:"sdp"`
	Type         string                     `json:"type"`
}

type RTCIceCandidate struct {
	Fromid    string `json:"fromid"`
	Toid      string `json:"toid"`
	Candidate string `json:"candidate"`
	Type      string `json:"type"`
}

type TypeGetter struct {
	Type string `json:"type"`
}

type PPMessage struct {
	Ver             byte
	Type            byte
	Length          uint16
	Header          string
	ExtensionHeader string
	Payload         []byte
}

func GetPPMessage(header interface{}, extHeader interface{}, payload []byte) *PPMessage {
	pp := new(PPMessage)
	pp.Ver = 0x01
	pp.Type = 0x01

	if extHeader != nil {
		buf, err := json.Marshal(extHeader)
		if err != nil {
			Println(ERROR, "extHeader marshal error: ", err)
			return nil
		}

		switch header.(type) {
		case *BroadcastData:
			pp.ExtensionHeader = string(buf)
			header.(*BroadcastData).ReqParams.ExtHeaderLen = len(pp.ExtensionHeader)

		case *GetDataRequset:
			pp.ExtensionHeader = string(buf)
			header.(*GetDataRequset).ExtHeaderLen = len(pp.ExtensionHeader)

		case *GetDataResponse:
			pp.ExtensionHeader = string(buf)
			header.(*GetDataResponse).RspParams.ExtHeaderLen = len(pp.ExtensionHeader)

		case *PrimaryPeerResponse:
			pp.ExtensionHeader = string(buf)
			header.(*PrimaryPeerResponse).RspParams.ExtHeaderLen = len(pp.ExtensionHeader)
		}
	}

	buf, err := json.Marshal(header)

	if err != nil {
		Println(ERROR, "header marshal error: ", err)
		return nil
	}
	pp.Header = string(buf)
	pp.Length = uint16(len(pp.Header))
	pp.Payload = payload

	return pp
}

const (
	ReqCode_Hello           = 1
	RspCode_Hello           = 1202
	ReqCode_Estab           = 2
	RspCode_Estab_Yes       = 2200
	RspCode_Estab_No        = 2603
	ReqCode_Probe           = 3
	RspCode_Probe           = 3200
	ReqCode_Primary         = 4
	RspCode_Primary_Yes     = 4200
	RspCode_Primary_No      = 4603
	ReqCode_Candidate       = 5
	RspCode_Candidate       = 5200
	ReqCode_BroadcastData   = 6
	RspCode_BroadcastData   = 6200
	ReqCode_Release         = 7
	RspCode_Release         = 7200
	ReqCode_HeartBeat       = 8
	RspCode_HeartBeat       = 8200
	ReqCode_ScanTree        = 9
	RspCode_ScanTreeNonLeaf = 9100
	RspCode_ScanTreeLeaf    = 9200
	ReqCode_Buffermap       = 10
	RspCode_Buffermap       = 10200
	ReqCode_GetData         = 11
	RspCode_GetData         = 11200
)

type HelloPeer struct {
	ReqCode   int                `json:"req-code"`
	ReqParams HelloPeerReqParams `json:"req-params"`
}

type HelloPeerReqParams struct {
	Operation HelloPeerReqParamsOperation `json:"operation"`
	Peer      HelloPeerReqParamsPeer      `json:"peer"`
}

type HelloPeerReqParamsOperation struct {
	OverlayId         string `json:"overlay-id"`
	ConnNum           int    `json:"conn_num"`
	Ttl               int    `json:"ttl"`
	Recovery          *bool  `json:"recovery,omitempty"`
	VerticalCandidate *bool  `json:"vertical-candidate,omitempty"`
}

type HelloPeerReqParamsPeer struct {
	PeerId   string `json:"peer-id"`
	Address  string `json:"address"`
	TicketId int    `json:"ticket-id"`
}

type HelloPeerResponse struct {
	RspCode int `json:"rsp-code"`
}

type EstabPeer struct {
	ReqCode   int                `json:"req-code"`
	ReqParams EstabPeerReqParams `json:"req-params"`
}

type EstabPeerReqParams struct {
	Operation       EstabPeerReqParamsOperation `json:"operation"`
	AllowPrimaryReq bool                        `json:"allow-primary-req"`
	Peer            EstabPeerReqParamsPeer      `json:"peer"`
}

type EstabPeerReqParamsOperation struct {
	OverlayId string `json:"overlay-id"`
}

type EstabPeerReqParamsPeer struct {
	PeerId   string `json:"peer-id"`
	TicketId int    `json:"ticket-id"`
}

type EstabPeerResponse struct {
	RspCode int `json:"rsp-code"`
}

type ProbePeerReqParamsOperation struct {
	NtpTime string `json:"ntp-time"`
}

type ProbePeerReqParams struct {
	Operation ProbePeerReqParamsOperation `json:"operation"`
}

type ProbePeerRequest struct {
	ReqCode   int                `json:"req-code"`
	ReqParams ProbePeerReqParams `json:"req-params"`
}

type ProbePeerResponse struct {
	RspCode   int                `json:"rsp-code"`
	RspParams ProbePeerReqParams `json:"rsp-params"`
}

type PrimaryPeer struct {
	ReqCode   int               `json:"req-code"`
	ReqParams PrimaryPeerParams `json:"req-params"`
}

type PrimaryPeerParams struct {
	Buffermap *[]*PeerBuffermap `json:"buffermap,omitempty"`
}

type PeerBuffermap struct {
	SourcePeerId string
	Sequence     []int
}

type DataPacketPayload struct {
	Header  *BroadcastData
	Payload *[]byte
}

type DataPacket struct {
	Sequence int
	DateTime time.Time
	Payload  DataPacketPayload
}

type CachingBuffer struct {
	SourcePeerId string
	DataPackets  []DataPacket
	BufferMutax  sync.Mutex
}

type PrimaryPeerResponseAll struct {
	Header          PrimaryPeerResponse
	ExtensionHeader PrimaryPeerResponseExtensionHeader
	Payload         []byte
}

type PrimaryPeerResponse struct {
	RspCode   int                       `json:"rsp-code"`
	RspParams PrimaryPeerResponseParams `json:"rsp-params"`
}

type PrimaryPeerResponseParams struct {
	Buffermap    *[]*PeerBuffermap `json:"buffermap,omitempty"`
	ParentPeer   *ParentPeer       `json:"parent-peer,omitempty"`
	ExtHeaderLen int               `json:"ext-header-len"`
	Payload      ResponsePayload   `json:"payload"`
}

type ParentPeer struct {
	PeerId  string `json:"peer-id"`
	Address string `json:"address"`
}

type PrimaryPeerResponseExtensionHeader struct {
	Peers []PrimaryPeerResponseExtensionHeaderPeers `json:"peers"`
}

type PrimaryPeerResponseExtensionHeaderPeers struct {
	PeerId      string                                                `json:"peer-id"`
	DisplayName string                                                `json:"display-name"`
	PublicKey   bool                                                  `json:"public-key"`
	CachedData  []PrimaryPeerResponseExtensionHeaderPeersCachingMedia `json:"cached-data"`
}

type PrimaryPeerResponseExtensionHeaderPeersCachingMedia struct {
	ChannelId string `json:"channel-id"`
	Sequence  int    `json:"sequence"`
}

type CandidatePeer struct {
	ReqCode int `json:"req-code"`
}

type CandidatePeerResponse struct {
	RspCode int `json:"rsp-code"`
}

type ReleasePeerParamsOperation struct {
	Ack bool `json:"ack"`
}

type ReleasePeerParams struct {
	Operation ReleasePeerParamsOperation `json:"operation"`
}

type ReleasePeer struct {
	ReqCode   int               `json:"req-code"`
	ReqParams ReleasePeerParams `json:"req-params"`
}

type ReleasePeerResponse struct {
	RspCode int `json:"rsp-code"`
}

type HeartBeat struct {
	ReqCode int `json:"req-code"`
}

type HeartBeatResponse struct {
	RspCode int `json:"rsp-code"`
}

type ScanTree struct {
	ReqCode   int            `json:"req-code"`
	ReqParams ScanTreeParams `json:"req-params"`
}

type ScanTreeParams struct {
	CSeq    int             `json:"cseq"`
	Overlay ScanTreeOverlay `json:"overlay"`
	Peer    ScanTreePeer    `json:"peer"`
}

type ScanTreeOverlay struct {
	OverlayId string     `json:"overlay-id"`
	Via       [][]string `json:"via"`
	Path      [][]string `json:"path"`
}

type ScanTreePeer struct {
	PeerId   string `json:"peer-id"`
	Address  string `json:"address"`
	TicketId int    `json:"ticket-id"`
}

type ScanTreeResponse struct {
	RspCode   int            `json:"rsp-code"`
	RspParams ScanTreeParams `json:"rsp-params"`
}

type Buffermap struct {
	ReqCode   int             `json:"req-code"`
	ReqParams BuffermapParams `json:"req-params"`
}

type BuffermapParams struct {
	OverlayId string `json:"overlay-id"`
}

type BuffermapResponse struct {
	RspCode   int                `json:"rsp-code"`
	RspParams BuffermapRspParams `json:"rsp-params"`
}

type BuffermapRspParams struct {
	OverlayId string            `json:"overlay-id"`
	Buffermap *[]*PeerBuffermap `json:"buffermap,omitempty"`
}

type GetDataRequset struct {
	ReqCode      int                  `json:"req-code"`
	ReqParams    GetDataRequsetParams `json:"req-params"`
	ExtHeaderLen int                  `json:"ext-header-len"`
}

type GetDataRequsetParams struct {
	OverlayId string `json:"overlay-id"`
	SourceId  string `json:"source-id"`
	Sequence  int    `json:"sequence"`
}

type GetDataRequsetExtensionHeader struct {
	ChannelId string `json:"channel-id"`
	//PeerId    string `json:"peer-id"`
}

type GetDataResponse struct {
	RspCode   int              `json:"rsp-code"`
	RspParams GetDataRspParams `json:"rsp-params"`
}

type GetDataRspParams struct {
	Peer         GetDataRspParamsPeer `json:"peer"`
	Sequence     int                  `json:"sequence"`
	ExtHeaderLen int                  `json:"ext-header-len"`
	Payload      ResponsePayload      `json:"payload"`
}

type GetDataResponseExtensionHeader struct {
	ChannelId string `json:"channel-id"`
	MediaType string `json:"media-type"`
	//PeerId    string `json:"peer-id"`
	//Sequence  int    `json:"sequence"`
}

type GetDataRspParamsPeer struct {
	PeerId string `json:"peer-id"`
}

type BroadcastData struct {
	ReqCode   int                 `json:"req-code"`
	ReqParams BroadcastDataParams `json:"req-params"`
}

type BroadcastDataParams struct {
	Operation    *BroadcastDataParamsOperation `json:"operation"`
	Peer         BroadcastDataParamsPeer       `json:"peer"`
	Payload      ResponsePayload               `json:"payload"`
	ExtHeaderLen int                           `json:"ext-header-len"`
}

type BroadcastDataParamsOperation struct {
	Ack           bool `json:"ack"`
	CandidatePath bool `json:"candidate-path"`
}

type BroadcastDataParamsPeer struct {
	PeerId   string `json:"peer-id"`
	Sequence int    `json:"sequence"`
}

type ResponsePayload struct {
	Length      int    `json:"length"`
	PayloadType string `json:"payload-type"`
}

type BroadcastDataResponse struct {
	RspCode int `json:"rsp-code"`
}

type BroadcastDataExtensionHeader struct {
	ChannelId string `json:"channel-id"`
	MediaType string `json:"media-type"`
}

type BroadcastDataExtensionHeaderDataCache struct {
	BroadcastDataExtensionHeader
	SourceId string `json:"source-id"`
	Sequence int    `json:"sequence"`
}

type BroadcastDataExtensionHeaderControl struct {
	BroadcastDataExtensionHeader
	ControlType string `json:"control-type"`
}

type BroadcastDataExtensionHeaderControlJoin struct {
	BroadcastDataExtensionHeaderControl
	JoinerInfo BroadcastDataExtensionHeaderControlJoinJoinerInfo `json:"joiner-info"`
}

type BroadcastDataExtensionHeaderControlJoinJoinerInfo struct {
	PeerId      string `json:"peer-id"`
	DisplayName string `json:"display-name"`
	PublicKey   string `json:"public-key"`
}

type BroadcastDataExtensionHeaderControlLeave struct {
	BroadcastDataExtensionHeaderControl
	LeaverInfo BroadcastDataExtensionHeaderControlLeaveLeaverInfo `json:"leaver-info"`
}

type BroadcastDataExtensionHeaderControlLeaveLeaverInfo struct {
	PeerId      string `json:"peer-id"`
	DisplayName string `json:"display-name"`
}

type BroadcastDataExtensionHeaderControlTransmission struct {
	BroadcastDataExtensionHeaderControl
	SourceList BroadcastDataExtensionHeaderControlTransmissionSourceList `json:"source-list"`
}

type BroadcastDataExtensionHeaderControlTransmissionSourceList struct {
	ServiceSourceList *[]BroadcastDataExtensionHeaderControlPeerList                                `json:"service-src-list,omitempty"`
	ServiceBlockList  *[]BroadcastDataExtensionHeaderControlPeerList                                `json:"service-block-list,omitempty"`
	ChannelSourceList *[]BroadcastDataExtensionHeaderControlTransmissionSourceListChannelSourceList `json:"channel-src-list,omitempty"`
}

type BroadcastDataExtensionHeaderControlPeerList struct {
	PeerId string `json:"peer-id"`
}

type BroadcastDataExtensionHeaderControlTransmissionSourceListChannelSourceList struct {
	ChannelId  string                                        `json:"channel-id"`
	SourceList []BroadcastDataExtensionHeaderControlPeerList `json:"source-list"`
}

type BroadcastDataExtensionHeaderControlOwnership struct {
	BroadcastDataExtensionHeaderControl
	OwnerInfo BroadcastDataExtensionHeaderControlPeerList `json:"owner-info"`
}

type BroadcastDataExtensionHeaderControlSessionInfo struct {
	BroadcastDataExtensionHeaderControl
	SessionInfo BroadcastDataExtensionHeaderControlSessionInfoSession `json:"session-info"`
}

type BroadcastDataExtensionHeaderControlSessionInfoSession struct {
	Title         *string `json:"title,omitempty"`
	Description   *string `json:"description,omitempty"`
	StartDateTime *string `json:"start-datetime,omitempty"`
	EndDateTime   *string `json:"end-datetime,omitempty"`
}

type BroadcastDataExtensionHeaderControlExpulsion struct {
	BroadcastDataExtensionHeaderControl
	ExpulsionList []BroadcastDataExtensionHeaderControlPeerList `json:"expulsion-list"`
}

type SessionChangeInfo struct {
	OverlayId     string                      `json:"overlayId"`
	Title         *string                     `json:"title,omitempty"`
	Description   *string                     `json:"description,omitempty"`
	OwnerId       *string                     `json:"owner-Id,omitempty"`
	AccessKey     *string                     `json:"access-key,omitempty"`
	StartDateTime *string                     `json:"start-datetime,omitempty"`
	EndDateTime   *string                     `json:"end-datetime,omitempty"`
	SourceList    *[]string                   `json:"source-list,omitempty"`
	ChannelList   *[]SessionChangeInfoChannel `json:"channel-list,omitempty"`
}

type SessionChangeInfoChannel struct {
	ChannelId  string    `json:"channel-Id"`
	SourceList *[]string `json:"source-list,omitempty"`
}

type NetworkPeerInfo struct {
	PeerId   string `json:"peer_id"`
	TicketId int    `json:"ticket_id"`
}

type NetworkResponse struct {
	Peer              NetworkPeerInfo   `json:"peer"`
	Primary           []NetworkPeerInfo `json:"primary"`
	InComingCandidate []NetworkPeerInfo `json:"in_coming_candidate"`
	OutGoingCandidate []NetworkPeerInfo `json:"out_going_candidate"`
}

type IoTAppData struct {
	Keyword string      `json:"keyword"`
	Value   interface{} `json:"value"`
}

type IoTData struct {
	AppId   string
	AppData []IoTAppData
}

type BlockChainAppData struct {
	Sender   string
	Receiver string
	AppData  string
}

type BlockChainData struct {
	AppId   string
	AppData BlockChainAppData
}

type BlockChainRegister struct {
	AppId    string
	WalletId string
	IpAddr   string
	Port     string
}

type MediaData struct {
	AppId  string
	Length int
}

type MediaAppData struct {
	AppId   string
	AppData *[]byte
}

type IoTDataResponse struct {
	DateTime string        `json:"datetime"`
	Data     *[]IoTAppData `json:"data"`
}

type IoTTypeResponse struct {
	PeerId string                `json:"peer_id"`
	Data   []IoTTypeResponseData `json:"data"`
}

type IoTTypeResponseData struct {
	DateTime string      `json:"datetime"`
	Data     interface{} `json:"data"`
}

type ApiPeerInfoOverlay struct {
	OverlayTitle string `json:"overlay-title"`
	OverlayId    string `json:"overlay-id"`
	TicketId     int    `json:"ticket-id"`
}

type ApiPeerInfo struct {
	PeerId         string             `json:"peer-id"`
	NetworkAddress string             `json:"network-address"`
	OverlayNetwork ApiPeerInfoOverlay `json:"overlay-network"`
	DebugLevel     int                `json:"debug-level"`
	PeerAuth       string             `json:"peer-auth"`
}

type ApiPeerStatus struct {
	OverlayId              string   `json:"overlay-id"`
	Title                  string   `json:"title"`
	PeerId                 string   `json:"peer-id"`
	OwnerId                string   `json:"owner-id"`
	AppId                  []string `json:"app-id"`
	PrimaryCount           int      `json:"primary-count"`
	IncomingCandidateCount int      `json:"incoming-candidate-count"`
	OutgoingCandidateCount int      `json:"outgoing-candidate-count"`
	OverlayServerAddr      string   `json:"overlay-server-addr"`
	SignalingServerAddr    string   `json:"signaling-server-addr"`
	UdpPort                int      `json:"udp-port"`
}

type ApiOverlayInfo struct {
	OverlayId string `json:"overlay-id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	SubType   string `json:"sub-type"`
	OwnerId   string `json:"owner-id"`
	//Expires           int         `json:"expires"`
	Description       *string     `json:"description,omitempty"`
	Auth              OverlayAuth `json:"auth"`
	TicketId          int         `json:"ticket-id"`
	HeartbeatInterval int         `json:"heartbeat-interval"`
	HeartbeatTimeout  int         `json:"heartbeat-timeout"`
}

type ApiOverlayRefreshResponse struct {
	OverlayId string `json:"overlay-id"`
	Expires   int    `json:"expires"`
	PeerId    string `json:"peer-id"`
}

type ApiScanTreeData struct {
	CSeq int         `json:"cseq"`
	Data *[][]string `json:"data"`
}

type ApiScanTreeResult struct {
	PeerId   string
	TicketId int
	Address  string
}
