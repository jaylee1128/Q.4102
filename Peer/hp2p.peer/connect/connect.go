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

package connect

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"hp2p.util/util"

	"github.com/pion/webrtc/v3"
)

type PeerPosition int

const (
	Init PeerPosition = iota
	InComing
	OutGoing
	SendHello
	Established
	InComingCandidate
	OutGoingCandidate
	InComingPrimary
	OutGoingPrimary
	Stable
)

func (p PeerPosition) String() string {
	switch p {
	case Init:
		return "Init"
	case InComing:
		return "InComing"
	case OutGoing:
		return "OutGoing"
	case SendHello:
		return "SendHello"
	case Established:
		return "Established"
	case InComingCandidate:
		return "InComingCandidate"
	case OutGoingCandidate:
		return "OutGoingCandidate"
	case InComingPrimary:
		return "InComingPrimary"
	case OutGoingPrimary:
		return "OutGoingPrimary"
	case Stable:
		return "Stable"
	default:
		return "Unknown"
	}
}

const (
	RecoveryByPush string = "push"
	RecoveryByPull string = "pull"
)

type Connect interface {
	Init(peerId string, configPath string)
	PeerId() string
	PeerOriginId() string
	GetOriginId(peerId string) string
	JoinAndConnect(recovery bool, accessKey string)
	CreateOverlay(hoc *util.HybridOverlayCreation) *util.OverlayInfo
	OverlayInfo() *util.OverlayInfo
	PeerInfo() *util.PeerInfo
	ServiceInfo() *util.ServiceInfo
	Join(recovery bool, accessKey string) *util.HybridOverlayJoinResponse
	ConnectAfterJoin(peerList *[]util.PeerInfo, accessKey string)
	OverlayModification(hom *util.HybridOverlayModification) *util.HybridOverlayModification
	OverlayRemove(hom *util.HybridOverlayRemoval) *util.HybridOverlayRemovalResponse
	OverlayRefresh(hor *util.HybridOverlayRefresh) *util.HybridOverlayRefreshResponse
	OverlayReport()
	OverlayReportBy(overlayId string) *util.HybridOverlayReportOverlay
	OverlayQuery(ovid *string, title *string, desc *string) bool
	QueryOverlay(ovid *string, title *string, desc *string) *util.HybridOverlayQueryResponse
	OverlayLeave() *util.HybridOverlayLeaveResponse
	OverlayLeaveBy(overlayId string) *util.HybridOverlayLeaveResponse
	UserQuery(hou *util.HybridOverlayUserQuery) *util.HybridOverlayUserQueryResponse
	SendTransmissionNoti(hom *util.HybridOverlayModification)
	SendOwnershipNoti(hom *util.HybridOverlayModification)
	SendSessionInfoNoti(hom *util.HybridOverlayModification)
	SendExpulsionNoti(peers *[]string)
	SendSessionTerminationNoti()
	OnTrack(toPeerId string, kind string, track *webrtc.TrackLocalStaticRTP)
	GetTrack(kind string) *webrtc.TrackLocalStaticRTP
	Release(done *chan *util.HybridOverlayLeaveResponse)
	Recovery()
	SendScanTree() int
	SendData(channelId string, dataType string, data []byte) int32
	CheckOwner(peerId string) bool
	HasConnection() bool
	ConnectionInfo() *util.NetworkResponse
	//IoTData(data *util.IoTData)
	//BlockChainData(data *util.BlockChainData)
	//MediaData(data *util.MediaAppData)
	GetClientConfig() util.ClientConfig
	//GetPeerInfo() *PeerInfo
	GetPeerConfig() *util.PeerConfig
	GetPeerStatus() *util.PeerStatus

	/*
		GetConnectedAppIds() *[]string
		AddConnectedAppIds(appid string)
		IsUDPConnection() bool
		GetIoTPeerList() *[]string
		GetIoTDataListByPeer(peerId string) *[]util.IoTDataResponse
		GetIoTLastDataByPeer(peerId string) *util.IoTDataResponse
		GetIoTDataListByType(keyword string) *[]*util.IoTTypeResponse
		GetIoTLastDataByType(keyword string) *[]*util.IoTTypeResponse

		SetScanTreeReportCallback(report func(path *[][]string, cseq int))
		SetRecvChatCallback(recvchat func(peerId string, msg string))
		SetRecvDataCallback(recvdata func(sender string, source string, data string))
		//SetLog2WebCallback(log2web func(log string))
		SetRecvIoTCallback(recviot func(msg string))
		SetRecvBlockChainCallback(recvblc func(msg string))
		SetRecvMediaCallback(recvmedia func(sender string, data *[]byte))
	*/
	SetHeartbeatInterval(interval int)
	SetHeartbeatTimeout(timeout int)
	GetHeartbeatInterval() int
	GetHeartbeatTimeout() int

	SetPeerIndex(index string)
	GetPeerIndex() string

	SetSessionTerminationCallback(callback func(overlayId string, peerId string) bool)
	SetPeerChangeCallback(callback func(overlayId string, peerId string, displayName string, leave bool) bool)
	SetSessionChangeCallback(callback func(change *util.SessionChangeInfo) bool)
	SetExpulsionCallback(callback func(overlayId string, peerId string) bool)
	SetDataCallback(callback func(overlayId string, channelId string, senderId string, dataType string, data *[]byte) bool)

	GetJoinPeers() *map[string]*util.JoinPeerInfo
	JoinPeersLock()
	JoinPeersUnlock()
	SetPrivateKey(privateKey *rsa.PrivateKey) bool
	SetPrivateKeyFromFile(privateKeyPath string) bool
	GetPublicKeyString() string
	GetJoinPeerPublicKey(peerOriginId string) *rsa.PublicKey
	ParsePublicKey(keystring string) *rsa.PublicKey
	SignStruct(origin interface{}) []byte
	SignData(origin []byte) []byte
	VerifyData(data *[]byte, sign *[]byte, publicKey *rsa.PublicKey) bool
}

type Common struct {
	util.HOMP
	ClientConfig util.ClientConfig
	PeerConfig   util.PeerConfig
	PeerInfo     util.PeerInfo
	peerOriginId string
	OverlayInfo  util.OverlayInfo
	//PeerAddress  string
	//PeerAuth     PeerAuth
	PeerStatus util.PeerStatus

	//peerInstanceId string

	EstabPeerCount      int
	HaveOutGoingPrimary bool

	letterRunes []rune
	joinTicker  *time.Ticker

	OutGoingPrimaryMux sync.Mutex
	PeerMapMux         sync.Mutex
	CommunicationMux   sync.Mutex

	VideoTrack       *webrtc.TrackLocalStaticRTP
	AudioTrack       *webrtc.TrackLocalStaticRTP
	ChangeVideoTrack bool
	ChangeAudioTrack bool

	LeaveOverlay bool

	CachingBufferMap      map[string]*util.CachingBuffer
	CachingBufferMapMutex sync.Mutex

	//RecvDataCallback func(sender string, source string, data string)
	//log2WebCallback          func(log string)
	//ReportScanTreeCallback func(path *[][]string, cseq int)
	//RecvChatCallback       func(peerId string, msg string)
	//RecvIoTCallback        func(msg string)
	//RecvBlockChainCallback func(msg string)
	//RecvMediaCallback      func(sender string, data *[]byte)

	UDPConnection bool

	//connectedAppIds map[string]bool

	GrandParentId string

	HeartbeatInterval int
	HeartbeatTimeout  int

	PeerIndex string

	PrivateKey   *rsa.PrivateKey
	publicKeyPEM []byte

	SessionTerminationCallback func(overlayId string, peerId string) bool
	PeerChangeCallback         func(overlayId string, peerId string, displayName string, leave bool) bool
	SessionChangeCallback      func(change *util.SessionChangeInfo) bool
	ExpulsionCallback          func(overlayId string, peerId string) bool
	DataCallback               func(overlayId string, channelId string, senderId string, dataType string, data *[]byte) bool

	//joinPeerMap map[string]map[string]*util.JoinPeerInfo
	joinPeerMap map[string]*util.JoinPeerInfo
	JoinPeerMux sync.Mutex

	ControlChannelId string
	VideoChannelId   string

	IsFirstPrimary bool

	//PeerJoinNotiMap map[string]string
}

func (conn *Common) CommonInit(peerId string, configPath string) {
	conn.peerOriginId = peerId
	var instanceId int64 = time.Now().UnixMicro()
	conn.PeerInfo.PeerId = peerId + ";" + strconv.FormatInt(instanceId, 10)
	//conn.PeerInfo.InstanceId = instanceId
	conn.ClientConfig = util.ReadClientConfig(configPath)
	util.SetLogPath(configPath)
	util.SetLevel(conn.ClientConfig.LogLevel)
	util.SetWriter(conn.ClientConfig.FileLog, conn.ClientConfig.StdoutLog)
	util.LogInit()
	util.PrintJson(util.INFO, "ClientConfig", conn.ClientConfig)
	conn.PeerConfig = util.ReadPeerConfig(configPath)
	conn.OverlayAddr = conn.ClientConfig.OverlayServerAddr
	conn.PeerInfo.TicketId = -1

	conn.PeerStatus.CostMap.Primary = []string{}
	conn.PeerStatus.CostMap.OutgoingCandidate = []string{}
	conn.PeerStatus.CostMap.IncomingCandidate = []string{}

	conn.letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	mrand.Seed(time.Now().UnixNano())
	if len(conn.PeerInfo.Auth.Password) <= 0 {
		conn.PeerInfo.Auth.Password = conn.RandStringRunes(16)
	}
	conn.EstabPeerCount = 0
	conn.HaveOutGoingPrimary = false
	conn.VideoTrack = nil
	conn.AudioTrack = nil
	conn.ChangeAudioTrack = false
	conn.ChangeVideoTrack = false

	conn.LeaveOverlay = false

	conn.UDPConnection = false

	conn.CachingBufferMap = make(map[string]*util.CachingBuffer)

	conn.GrandParentId = ""

	conn.HeartbeatInterval = 10
	conn.HeartbeatTimeout = 15

	conn.PrivateKey = nil
	conn.publicKeyPEM = []byte{}

	conn.joinPeerMap = make(map[string]*util.JoinPeerInfo)
	//conn.PeerJoinNotiMap = make(map[string]string)

	conn.IsFirstPrimary = true
}

/*func (conn *Common) GetPeerInfo() *PeerInfo {
	return &conn.PeerInfo
}*/

func (conn *Common) ServiceInfo() *util.ServiceInfo {
	return &conn.OverlayInfo.ServiceInfo
}

func (conn *Common) GetPeerIndex() string {
	return conn.PeerIndex
}

func (conn *Common) SetPeerIndex(index string) {
	conn.PeerIndex = index
}

func (conn *Common) SetHeartbeatInterval(interval int) {
	conn.HeartbeatInterval = interval
}

func (conn *Common) SetHeartbeatTimeout(timeout int) {
	conn.HeartbeatTimeout = timeout
}

func (conn *Common) GetHeartbeatInterval() int {
	return conn.HeartbeatInterval
}

func (conn *Common) GetHeartbeatTimeout() int {
	return conn.HeartbeatTimeout
}

func (conn *Common) GetPeerConfig() *util.PeerConfig {
	return &conn.PeerConfig
}

func (conn *Common) GetPeerStatus() *util.PeerStatus {
	return &conn.PeerStatus
}

func (conn *Common) IsUDPConnection() bool {
	return conn.UDPConnection
}

/*
func (conn *Common) GetConnectedAppIds() *[]string {
	appids := make([]string, 0)

	for key := range conn.connectedAppIds {
		appids = append(appids, key)
	}

	return &appids
}

func (conn *Common) AddConnectedAppIds(appid string) {
	conn.connectedAppIds[appid] = true
}
*/
/*
	func (conn *Common) GetIoTPeerList() *[]string {
		defer conn.CachingBufferMapMutex.Unlock()

		list := make([]string, 0)

		conn.CachingBufferMapMutex.Lock()

		for key, val := range conn.CachingBufferMap {
			for _, dp := range val.DataPackets {

				extHeader := util.BroadcastDataExtensionHeader{}
				err := json.Unmarshal((*dp.Payload.Payload)[:dp.Payload.Header.ReqParams.ExtHeaderLen], &extHeader)
				if err != nil {
					util.Println(util.ERROR, "extheader Unmarshal error:", err)
					break
				} else {
					if extHeader.AppId == consts.AppIdIoT {
						list = append(list, key)
						break
					}
				}
			}
		}

		return &list
	}

	func (conn *Common) GetIoTDataListByPeer(peerId string) *[]util.IoTDataResponse {
		defer conn.CachingBufferMapMutex.Unlock()

		list := make([]util.IoTDataResponse, 0)

		conn.CachingBufferMapMutex.Lock()

		val, ok := conn.CachingBufferMap[peerId]

		if ok {
			val.BufferMutax.Lock()

			for _, dp := range val.DataPackets {

				extHeader := util.BroadcastDataExtensionHeader{}
				err := json.Unmarshal((*dp.Payload.Payload)[:dp.Payload.Header.ReqParams.ExtHeaderLen], &extHeader)
				if err != nil {
					util.Println(util.ERROR, "extheader Unmarshal error:", err)
					continue
				} else {
					if extHeader.AppId != consts.AppIdIoT {
						continue
					}
				}

				iot := []util.IoTAppData{}
				json.Unmarshal(*dp.Payload.Payload, &iot)

				res := util.IoTDataResponse{}
				res.DateTime = dp.DateTime.Format("2006-01-02 15:04:05")
				res.Data = &iot

				list = append(list, res)
			}

			val.BufferMutax.Unlock()
		}

		return &list
	}

	func (conn *Common) GetIoTLastDataByPeer(peerId string) *util.IoTDataResponse {
		defer conn.CachingBufferMapMutex.Unlock()

		res := util.IoTDataResponse{}

		conn.CachingBufferMapMutex.Lock()

		val, ok := conn.CachingBufferMap[peerId]

		if ok {
			val.BufferMutax.Lock()

			for i := len(val.DataPackets) - 1; i >= 0; i-- {
				dp := val.DataPackets[i]

				extHeader := util.BroadcastDataExtensionHeader{}
				err := json.Unmarshal((*dp.Payload.Payload)[:dp.Payload.Header.ReqParams.ExtHeaderLen], &extHeader)
				if err != nil {
					util.Println(util.ERROR, "extheader Unmarshal error:", err)
					continue
				} else {
					if extHeader.AppId != consts.AppIdIoT {
						continue
					}
				}

				iot := []IoTAppData{}
				json.Unmarshal(*dp.Payload.Payload, &iot)

				res.DateTime = dp.DateTime.Format("2006-01-02 15:04:05")
				res.Data = &iot

				break
			}

			val.BufferMutax.Unlock()
		}

		return &res
	}

	func (conn *Common) GetIoTDataListByType(keyword string) *[]*IoTTypeResponse {
		defer conn.CachingBufferMapMutex.Unlock()

		list := make([]*IoTTypeResponse, 0)
		dic := make(map[string]*IoTTypeResponse)

		conn.CachingBufferMapMutex.Lock()

		for pid, val := range conn.CachingBufferMap {
			val.BufferMutax.Lock()

			for _, dp := range val.DataPackets {

				extHeader := BroadcastDataExtensionHeader{}
				err := json.Unmarshal((*dp.Payload.Payload)[:dp.Payload.Header.ReqParams.ExtHeaderLen], &extHeader)
				if err != nil {
					util.Println(util.ERROR, "extheader Unmarshal error:", err)
					continue
				} else {
					if extHeader.AppId != consts.AppIdIoT {
						continue
					}
				}

				iot := []IoTAppData{}
				json.Unmarshal(*dp.Payload.Payload, &iot)

				for _, appdata := range iot {
					if appdata.Keyword == keyword {

						data := IoTTypeResponseData{}
						data.DateTime = dp.DateTime.Format("2006-01-02 15:04:05")
						data.Data = appdata.Value

						res, ok := dic[pid]

						if ok {
							res.Data = append(res.Data, data)
						} else {
							res = new(IoTTypeResponse)
							res.PeerId = pid
							res.Data = append(res.Data, data)
							dic[pid] = res
						}
					}
				}
			}

			val.BufferMutax.Unlock()
		}

		for _, tr := range dic {
			list = append(list, tr)
		}

		return &list
	}

	func (conn *Common) GetIoTLastDataByType(keyword string) *[]*IoTTypeResponse {
		defer conn.CachingBufferMapMutex.Unlock()

		list := make([]*IoTTypeResponse, 0)

		conn.CachingBufferMapMutex.Lock()

		for pid, val := range conn.CachingBufferMap {
			val.BufferMutax.Lock()

		F1:
			for i := len(val.DataPackets) - 1; i >= 0; i-- {
				dp := val.DataPackets[i]

				extHeader := BroadcastDataExtensionHeader{}
				err := json.Unmarshal((*dp.Payload.Payload)[:dp.Payload.Header.ReqParams.ExtHeaderLen], &extHeader)
				if err != nil {
					util.Println(util.ERROR, "extheader Unmarshal error:", err)
					continue
				} else {
					if extHeader.AppId != consts.AppIdIoT {
						continue
					}
				}

				iot := []IoTAppData{}
				json.Unmarshal(*dp.Payload.Payload, &iot)

				for i := len(iot) - 1; i >= 0; i-- {
					if iot[i].Keyword == keyword {

						data := IoTTypeResponseData{}
						data.DateTime = dp.DateTime.Format("2006-01-02 15:04:05")
						data.Data = iot[i].Value

						res := new(IoTTypeResponse)
						res.PeerId = pid
						res.Data = append(res.Data, data)

						list = append(list, res)
						break F1
					}
				}
			}

			val.BufferMutax.Unlock()
		}

		return &list
	}
*/
func (conn *Common) GetClientConfig() util.ClientConfig {
	return conn.ClientConfig
}

func (conn *Common) GetTrack(kind string) *webrtc.TrackLocalStaticRTP {
	if kind == "video" {
		return conn.VideoTrack
	} else {
		return conn.AudioTrack
	}
}

func (conn *Common) RemoveCostMapPeer(costmap *[]string, toPeerId string) bool {
	for idx, peerid := range *costmap {
		if peerid == toPeerId {
			if len(*costmap) == 1 {
				*costmap = []string{}
			} else {
				(*costmap)[idx] = (*costmap)[len(*costmap)-1]
				*costmap = (*costmap)[:len(*costmap)-1]
			}

			return true
		}
	}

	return false
}

/*
func (conn *Common) SetRecvDataCallback(recvdata func(sender string, source string, data string)) {
	conn.RecvDataCallback = recvdata
}

func (conn *Common) SetLog2WebCallback(log2web func(log string)) {
	conn.log2WebCallback = log2web
}

func (conn *Common) SetScanTreeReportCallback(report func(path *[][]string, cseq int)) {
	conn.ReportScanTreeCallback = report
}

func (conn *Common) SetRecvChatCallback(recvchat func(peerId string, msg string)) {
	conn.RecvChatCallback = recvchat
}

func (conn *Common) SetRecvIoTCallback(recviot func(msg string)) {
	conn.RecvIoTCallback = recviot
}

func (conn *Common) SetRecvBlockChainCallback(recvblcn func(msg string)) {
	conn.RecvBlockChainCallback = recvblcn
}

func (conn *Common) SetRecvMediaCallback(recvmedia func(sender string, data *[]byte)) {
	conn.RecvMediaCallback = recvmedia
}*/

func (conn *Common) AddConnectionInfo(position PeerPosition, toPeerId string) {
	switch position {
	case InComingCandidate:
		conn.PeerStatus.NumInCandidate++
		conn.PeerStatus.CostMap.IncomingCandidate = append(conn.PeerStatus.CostMap.IncomingCandidate, toPeerId)
	case OutGoingCandidate:
		conn.PeerStatus.NumOutCandidate++
		conn.PeerStatus.CostMap.OutgoingCandidate = append(conn.PeerStatus.CostMap.OutgoingCandidate, toPeerId)
	case InComingPrimary:
		if conn.RemoveCostMapPeer(&conn.PeerStatus.CostMap.IncomingCandidate, toPeerId) {
			conn.PeerStatus.NumInCandidate--
		}

		conn.PeerStatus.NumPrimary++
		conn.PeerStatus.CostMap.Primary = append(conn.PeerStatus.CostMap.Primary, toPeerId)
	case OutGoingPrimary:
		if conn.RemoveCostMapPeer(&conn.PeerStatus.CostMap.OutgoingCandidate, toPeerId) {
			conn.PeerStatus.NumOutCandidate--
		}

		conn.PeerStatus.NumPrimary++
		conn.PeerStatus.CostMap.Primary = append(conn.PeerStatus.CostMap.Primary, toPeerId)
		conn.HaveOutGoingPrimary = true
	}

	conn.OverlayReport()
}

func (conn *Common) DelConnectionInfo(position PeerPosition, toPeerId string) {
	switch position {
	case InComingPrimary, OutGoingPrimary:
		if conn.RemoveCostMapPeer(&conn.PeerStatus.CostMap.Primary, toPeerId) {
			conn.PeerStatus.NumPrimary--

			if position == OutGoingPrimary {
				conn.HaveOutGoingPrimary = false
			}
		}

	case InComingCandidate:
		if conn.RemoveCostMapPeer(&conn.PeerStatus.CostMap.IncomingCandidate, toPeerId) {
			conn.PeerStatus.NumInCandidate--
		}

	case OutGoingCandidate:
		if conn.RemoveCostMapPeer(&conn.PeerStatus.CostMap.OutgoingCandidate, toPeerId) {
			conn.PeerStatus.NumOutCandidate--
		}
	}

	conn.OverlayReport()
}

func (conn *Common) RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = conn.letterRunes[mrand.Intn(len(conn.letterRunes))]
	}
	return string(b)
}

func (conn *Common) PeerId() string {
	return conn.PeerInfo.PeerId
}

func (conn *Common) PeerOriginId() string {
	return conn.peerOriginId
}

/*
	func (conn *Common) PeerInstanceId() string {
		if conn.peerInstanceId == "" {
			conn.peerInstanceId = conn.PeerInfo.PeerId + ";" + strconv.FormatInt(conn.PeerInfo.InstanceId, 10)
		}
		return conn.peerInstanceId
	}
*/

func (conn *Common) GetOriginId(peerId string) string {
	var originId string = peerId
	if strings.Contains(originId, ";") {
		originId = strings.Split(originId, ";")[0]
	}
	return originId
}

func (conn *Common) CheckOwner(peerId string) bool {
	owner := conn.GetOriginId(conn.OverlayInfo.OwnerId)
	peer := conn.GetOriginId(peerId)

	return owner == peer
}

func (conn *Common) CreateOverlay(hoc *util.HybridOverlayCreation) *util.OverlayInfo {
	conn.OverlayInfo.Copy(&hoc.Overlay)
	conn.OverlayInfo.Copy(conn.HOMP.CreateOverlay(hoc))

	return &conn.OverlayInfo
}

func (conn *Common) OverlayJoinBy(hoj *util.HybridOverlayJoin, recovery bool) *util.HybridOverlayJoinResponse {

	res := conn.HOMP.OverlayJoin(hoj, recovery)

	if res == nil {
		return nil
	}

	if len(res.Overlay.OverlayId) <= 0 {
		return nil
	}

	if !recovery {

		conn.OverlayInfo.Copy(&res.Overlay)
		conn.PeerInfo.TicketId = res.Peer.TicketId
		for _, channel := range res.Overlay.ServiceInfo.ChannelList {
			if channel.ChannelType == util.ChannelTypeControl {
				conn.ControlChannelId = channel.ChannelId
			}

			if channel.ChannelType == util.ChannelTypeVideo {
				conn.VideoChannelId = channel.ChannelId
			}
		}

		conn.JoinPeerMux.Lock()

		if joinpeer, ok := conn.joinPeerMap[conn.peerOriginId]; !ok {
			conn.joinPeerMap[conn.peerOriginId] = &util.JoinPeerInfo{}
			conn.joinPeerMap[conn.peerOriginId].PeerId = conn.PeerOriginId()
			conn.joinPeerMap[conn.peerOriginId].DisplayName = *conn.PeerInfo.DisplayName
			conn.joinPeerMap[conn.peerOriginId].PublicKeyPEM = conn.publicKeyPEM
			conn.joinPeerMap[conn.peerOriginId].PublicKey = nil
			conn.joinPeerMap[conn.peerOriginId].CachingMedia = make([]util.JoinPeerInfoCachingMedia, 0)
		} else {
			joinpeer.PeerId = conn.PeerOriginId()
			joinpeer.DisplayName = *conn.PeerInfo.DisplayName
			joinpeer.PublicKeyPEM = conn.publicKeyPEM
			joinpeer.PublicKey = nil
			if joinpeer.CachingMedia == nil {
				joinpeer.CachingMedia = make([]util.JoinPeerInfoCachingMedia, 0)
			}
		}

		conn.JoinPeerMux.Unlock()

		if res.Peer.Expires > 0 {
			conn.PeerConfig.Expires = res.Peer.Expires
			conn.joinTicker = time.NewTicker(time.Millisecond * time.Duration(float32(conn.PeerConfig.Expires)*0.8) * 1000)
			go func() {
				for range conn.joinTicker.C {

					if conn.LeaveOverlay {
						break
					}

					hor := util.HybridOverlayRefresh{}
					hor.Overlay.OverlayId = conn.OverlayInfo.OverlayId
					hor.Peer.PeerId = conn.PeerId()
					//hor.Peer.InstanceId = conn.PeerInfo.InstanceId
					hor.Peer.Address = conn.PeerInfo.Address
					hor.Peer.Auth = conn.PeerInfo.Auth
					conn.HOMP.OverlayRefresh(&hor)
				}
			}()
		}

		if conn.ServiceInfo().EndDatetime != nil && len(*conn.ServiceInfo().EndDatetime) > 0 {
			go func() {
				for {
					if conn.LeaveOverlay {
						break
					}

					now := time.Now()

					end, err := time.ParseInLocation("20060102150405", *conn.ServiceInfo().EndDatetime, time.Local)
					if err != nil {
						util.Println(util.ERROR, "Failed to parse end datetime:", err)
					} else {
						if end.Before(now) {
							conn.LeaveOverlay = true
							conn.OverlayLeave()

							if conn.SessionTerminationCallback != nil {
								conn.SessionTerminationCallback(conn.OverlayInfo.OverlayId, conn.PeerOriginId())
							}
							break
						}
					}

					time.Sleep(time.Second * 1)
				}
			}()

		}
	}

	return res
}

func (conn *Common) OverlayJoin(recovery bool) *util.HybridOverlayJoinResponse {
	hoj := new(util.HybridOverlayJoin)
	hoj.Overlay.OverlayId = conn.OverlayInfo.OverlayId
	if len(conn.OverlayInfo.Type) <= 0 {
		conn.OverlayInfo.Type = "core"
	}
	hoj.Overlay.Type = conn.OverlayInfo.Type
	if len(conn.OverlayInfo.SubType) <= 0 {
		conn.OverlayInfo.SubType = "tree"
	}
	hoj.Overlay.SubType = conn.OverlayInfo.SubType
	hoj.Overlay.Auth = &conn.OverlayInfo.Auth
	hoj.Peer.PeerId = conn.PeerId()
	hoj.Peer.DisplayName = conn.PeerInfo.DisplayName
	//hoj.Peer.InstanceId = conn.PeerInfo.InstanceId
	hoj.Peer.Address = conn.PeerInfo.Address
	hoj.Peer.Auth = conn.PeerInfo.Auth
	hoj.Peer.Expires = &conn.PeerConfig.Expires
	hoj.Peer.TicketId = &conn.PeerInfo.TicketId

	return conn.OverlayJoinBy(hoj, recovery)
}

func (conn *Common) OverlayModification(hom *util.HybridOverlayModification) *util.HybridOverlayModification {
	return conn.HOMP.OverlayModification(hom)
}

func (conn *Common) OverlayRemove(hom *util.HybridOverlayRemoval) *util.HybridOverlayRemovalResponse {
	return conn.HOMP.OverlayRemoval(hom)
}

func (conn *Common) OverlayRefresh(hor *util.HybridOverlayRefresh) *util.HybridOverlayRefreshResponse {
	return conn.HOMP.OverlayRefresh(hor)
}

func (conn *Common) OverlayReport() {
	conn.OverlayReportBy(conn.OverlayInfo.OverlayId)
}

func (conn *Common) OverlayReportBy(overlayId string) *util.HybridOverlayReportOverlay {
	hor := new(util.HybridOverlayReport)
	hor.Overlay.OverlayId = overlayId
	hor.Peer.PeerId = conn.PeerId()
	//hor.Peer.InstanceId = conn.PeerInfo.InstanceId
	hor.Status = conn.PeerStatus
	hor.Peer.Auth = conn.PeerInfo.Auth

	return conn.HOMP.OverlayReport(hor)
}

func (conn *Common) OverlayQuery(ovid *string, title *string, desc *string) bool {
	hoq := conn.QueryOverlay(ovid, title, desc)

	if hoq == nil || hoq.Overlay == nil || len(hoq.Overlay) <= 0 {

		//conn.log2WebCallback("Failed to query Overlay.")

		return false
	}

	//conn.log2WebCallback("Query Overlay.")

	conn.OverlayInfo.Copy(&(hoq.Overlay[0]))

	return true
}

func (conn *Common) QueryOverlay(ovid *string, title *string, desc *string) *util.HybridOverlayQueryResponse {
	hoq := conn.HOMP.QueryOverlay(ovid, title, desc)

	return hoq
}

func (conn *Common) OverlayLeave() *util.HybridOverlayLeaveResponse {
	return conn.OverlayLeaveBy(conn.OverlayInfo.OverlayId)
}

func (conn *Common) OverlayLeaveBy(overlayId string) *util.HybridOverlayLeaveResponse {

	if overlayId == "" {
		return nil
	}

	hol := new(util.HybridOverlayLeave)
	hol.Overlay.OverlayId = overlayId
	hol.Overlay.Auth = &conn.OverlayInfo.Auth
	hol.Peer.PeerId = conn.PeerId()
	//hol.Peer.InstanceId = conn.PeerInfo.InstanceId
	hol.Peer.Auth = &conn.PeerInfo.Auth

	res := conn.HOMP.OverlayLeave(hol)

	//conn.log2WebCallback("Leave Overlay.")

	util.Println(util.INFO, "Overlay leave response:", res)

	return res
}

func (conn *Common) UserQuery(hou *util.HybridOverlayUserQuery) *util.HybridOverlayUserQueryResponse {
	return conn.HOMP.UserQuery(hou)
}

func (conn *Common) SetSessionTerminationCallback(callback func(overlayId string, peerId string) bool) {
	conn.SessionTerminationCallback = callback
}

func (conn *Common) SetPeerChangeCallback(callback func(overlayId string, peerId string, displayName string, leave bool) bool) {
	conn.PeerChangeCallback = callback
}

func (conn *Common) SetSessionChangeCallback(callback func(change *util.SessionChangeInfo) bool) {
	conn.SessionChangeCallback = callback
}

func (conn *Common) SetExpulsionCallback(callback func(overlayId string, peerId string) bool) {
	conn.ExpulsionCallback = callback
}

func (conn *Common) SetDataCallback(callback func(overlayId string, channelId string, senderId string, dataType string, data *[]byte) bool) {
	conn.DataCallback = callback
}

func (conn *Common) GetJoinPeers() *map[string]*util.JoinPeerInfo {
	return &conn.joinPeerMap
}

func (conn *Common) GetJoinPeerPublicKey(peerOriginId string) *rsa.PublicKey {
	conn.JoinPeersLock()
	defer conn.JoinPeersUnlock()

	var peerId string = conn.GetOriginId(peerOriginId)

	var joinpeer *util.JoinPeerInfo = nil
	var ok bool = false

	if joinpeer, ok = conn.joinPeerMap[peerId]; !ok {
		joinpeer = &util.JoinPeerInfo{}
		joinpeer.PeerId = peerId
		joinpeer.DisplayName = ""
		joinpeer.PublicKeyPEM = []byte{}
		joinpeer.PublicKey = nil
		joinpeer.CachingMedia = make([]util.JoinPeerInfoCachingMedia, 0)
		conn.joinPeerMap[peerId] = joinpeer
	}

	if joinpeer.PublicKey == nil {
		if len(joinpeer.PublicKeyPEM) <= 0 {
			ureq := &util.HybridOverlayUserQuery{}
			ureq.OverlayId = conn.OverlayInfo.OverlayId
			ureq.PeerId = peerId

			ures := conn.HOMP.UserQuery(ureq)
			if ures != nil && ures.RspCode == 200 {
				if ures.Peer.DisplayName != nil && len(*ures.Peer.DisplayName) > 0 {
					joinpeer.DisplayName = *ures.Peer.DisplayName
				}

				if ures.Peer.Auth != nil && len(ures.Peer.Auth.PublicKey) > 0 {
					joinpeer.PublicKeyPEM = []byte(ures.Peer.Auth.PublicKey)
					joinpeer.PublicKey = conn.ParsePublicKey(ures.Peer.Auth.PublicKey)
				}
			}
		} else {
			joinpeer.PublicKey = conn.ParsePublicKey((string)(joinpeer.PublicKeyPEM))
		}
	}

	return joinpeer.PublicKey
}

func (conn *Common) ParsePublicKey(keystring string) (rslt *rsa.PublicKey) {
	defer func() {
		if err := recover(); err != nil {
			util.Println(util.ERROR, "ParsePublicKey error", keystring)
			rslt = nil
		}
	}()
	//var rslt *rsa.PublicKey = nil

	rslt = nil

	if len(keystring) <= 0 {
		util.Println(util.ERROR, "!!!! Failed to parse public key. keystring is 0 !!!!")
		return rslt
	}

	pubkeyblock, _ := pem.Decode([]byte(keystring))
	pubkeyinterface, err := x509.ParsePKIXPublicKey(pubkeyblock.Bytes)
	if err != nil {
		util.Println(util.ERROR, "!!!! Failed to parse public key. !!!!")
		return rslt
	}

	rslt = pubkeyinterface.(*rsa.PublicKey)

	return rslt
}

func (conn *Common) JoinPeersLock() {
	conn.JoinPeerMux.Lock()
}

func (conn *Common) JoinPeersUnlock() {
	conn.JoinPeerMux.Unlock()
}

func (conn *Common) SetPrivateKey(privateKey *rsa.PrivateKey) bool {
	conn.PrivateKey = privateKey

	publicKey := privateKey.PublicKey

	publicKeyData, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		util.Println(util.ERROR, "Failed to marshal public key.")
		return false
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyData,
	})

	conn.publicKeyPEM = publicKeyPEM
	util.Println(util.INFO, "publicKeyPEM : ", string(conn.publicKeyPEM))
	conn.PeerInfo.Auth.PublicKey = string(conn.publicKeyPEM)

	return true
}

func (conn *Common) SetPrivateKeyFromFile(privateKeyPath string) bool {
	keyfile, err := os.ReadFile(privateKeyPath)
	if err != nil {
		util.Println(util.ERROR, "No RSA private key found:", privateKeyPath)
		return false
	}

	privPem, _ := pem.Decode(keyfile)
	if privPem.Type != "RSA PRIVATE KEY" {
		util.Println(util.ERROR, "RSA private key is of the wrong type :", privPem.Type)
	}

	pemBytes := privPem.Bytes

	privateKey, err := x509.ParsePKCS1PrivateKey(pemBytes)

	if err != nil {
		util.Println(util.ERROR, "ParsePKCS1PrivateKey error:", err)
		return false
	}

	conn.PrivateKey = privateKey

	publicKey := privateKey.PublicKey

	publicKeyData, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		util.Println(util.ERROR, "Failed to marshal public key.")
		return false
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyData,
	})

	conn.publicKeyPEM = publicKeyPEM
	util.Println(util.INFO, "publicKeyPEM : ", string(conn.publicKeyPEM))
	conn.PeerInfo.Auth.PublicKey = string(conn.publicKeyPEM)

	return true
}

func (conn *Common) GetPublicKeyString() string {
	return string(conn.publicKeyPEM)
}

func (conn *Common) SignStruct(origin interface{}) []byte {
	if conn.PrivateKey == nil {
		util.Println(util.ERROR, "SignStruct private key is nil")
		return nil
	}

	buf, err := json.Marshal(origin)
	if err != nil {
		util.Println(util.ERROR, "SignStruct Marshal error:", err)
		return nil
	}

	hash := sha256.Sum256(buf)
	sign, err := rsa.SignPKCS1v15(rand.Reader, conn.PrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		util.Println(util.ERROR, "SignStruct sign error:", err)
		return nil
	}

	return sign
}

func (conn *Common) SignData(origin []byte) []byte {
	if conn.PrivateKey == nil {
		util.Println(util.ERROR, "SignData private key is nil")
		return nil
	}

	hash := sha256.Sum256(origin)
	sign, err := rsa.SignPKCS1v15(rand.Reader, conn.PrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		util.Println(util.ERROR, "SignData sign error:", err)
		return nil
	}

	return sign
}

func (conn *Common) VerifyData(data *[]byte, signed *[]byte, publicKey *rsa.PublicKey) bool {
	hash := sha256.Sum256(*data)
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], *signed)
	if err != nil {
		util.Println(util.ERROR, "VerifyData error:", err)
		return false
	}

	return true
}
