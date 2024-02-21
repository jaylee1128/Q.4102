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

package webrtc

import (
	"connect"
	"crypto/rsa"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"hp2p.util/util"

	"github.com/pion/rtcp"
	pwebrtc "github.com/pion/webrtc/v3"
)

type Peer struct {
	ToPeerId   string
	ToOriginId string

	//ToInstanceId     string
	//ToPeerInstanceId string
	ToTicketId int
	Info       *connect.Common
	IsOutGoing bool
	Position   connect.PeerPosition

	signalSend       *chan interface{}
	ppChan           chan interface{}
	connectChan      *chan bool
	broadcastChan    *chan interface{}
	buffermapResChan *chan *util.BuffermapResponse

	candidatesMux sync.Mutex

	pendingCandidates []*pwebrtc.ICECandidate

	webRtcConfig   pwebrtc.Configuration
	peerConnection *pwebrtc.PeerConnection
	dataChannel    *pwebrtc.DataChannel
	ConnectObj     *WebrtcConnect

	heartbeatCount int
	releasePeer    bool

	MediaReceive bool

	probeTime *int64

	isVerticalCandidate bool
}

func NewPeer(toPeerId string, position connect.PeerPosition, connectChan *chan bool, conn *WebrtcConnect) *Peer {
	peer := new(Peer)

	peer.ToPeerId = toPeerId
	peer.ToOriginId = conn.GetOriginId(toPeerId)

	peer.Position = position
	peer.ConnectObj = conn

	peer.Info = &conn.Common
	peer.signalSend = &conn.sendChan
	peer.connectChan = connectChan
	peer.broadcastChan = &conn.broadcastChan

	peer.ppChan = make(chan interface{})
	peer.releasePeer = false
	peer.heartbeatCount = 0
	peer.MediaReceive = false

	peer.webRtcConfig = pwebrtc.Configuration{
		ICEServers: []pwebrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peer.pendingCandidates = make([]*pwebrtc.ICECandidate, 0)

	peer.peerConnection = nil

	peer.isVerticalCandidate = false

	return peer
}

func (self *Peer) Close() {

	if self.releasePeer {
		return
	}

	self.releasePeer = true

	if self.peerConnection.ConnectionState() <= pwebrtc.PeerConnectionStateConnected {
		if cErr := self.peerConnection.Close(); cErr != nil {
			log.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}

	close(*self.connectChan)
	close(self.ppChan)
}

func (self *Peer) InitConnection(position connect.PeerPosition) {
	if self.peerConnection.ConnectionState() <= pwebrtc.PeerConnectionStateConnected {
		if cErr := self.peerConnection.Close(); cErr != nil {
			log.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}
	self.peerConnection = nil

	self.Position = position

}

func (self *Peer) signalCandidate(c *pwebrtc.ICECandidate) error {
	candi := util.RTCIceCandidate{}
	candi.Candidate = c.ToJSON().Candidate
	candi.Toid = self.ToPeerId
	candi.Type = "candidate"

	util.Println(util.INFO, "send iceCandidate to", self.ToPeerId)
	*self.signalSend <- candi

	return nil
}

func (self *Peer) AddIceCandidate(ice util.RTCIceCandidate) {

	if self.releasePeer {
		return
	}

	err := self.peerConnection.AddICECandidate(pwebrtc.ICECandidateInit{Candidate: ice.Candidate})

	if err != nil {
		//panic(err)
		util.Println(util.ERROR, err)
	}
}

func (self *Peer) SetSdp(rsdp util.RTCSessionDescription) {
	err := self.peerConnection.SetRemoteDescription(rsdp.Sdp)
	if err != nil {
		panic(err)
	}

	self.candidatesMux.Lock()
	for _, c := range self.pendingCandidates {
		if self.releasePeer {
			break
		}

		onICECandidateErr := self.signalCandidate(c)
		if onICECandidateErr != nil {
			panic(onICECandidateErr)
		}
	}
	self.candidatesMux.Unlock()
}

func (self *Peer) newPeerConnection(createDataChannel bool) {
	peerConnection, err := pwebrtc.NewPeerConnection(self.webRtcConfig)
	if err != nil {
		panic(err)
	}

	self.peerConnection = peerConnection

	peerConnection.OnICECandidate(func(c *pwebrtc.ICECandidate) {
		if c == nil {
			return
		}

		if self.releasePeer {
			return
		}

		self.candidatesMux.Lock()
		defer self.candidatesMux.Unlock()

		desc := peerConnection.RemoteDescription()
		if desc == nil {
			self.pendingCandidates = append(self.pendingCandidates, c)
		} else if onICECandidateErr := self.signalCandidate(c); onICECandidateErr != nil {
			panic(onICECandidateErr)
		}
	})

	if createDataChannel {
		self.peerConnection.OnTrack(func(tr *pwebrtc.TrackRemote, r *pwebrtc.RTPReceiver) {
			util.Println(util.INFO, self.ToPeerId, "OnTrack!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! create", tr.Kind())
			self.setLocalTrack(tr, r)
		})

		dataChannel, err := self.peerConnection.CreateDataChannel("data", nil)
		if err != nil {
			panic(err)
		}
		self.dataChannel = dataChannel

		peerConnection.OnConnectionStateChange(func(s pwebrtc.PeerConnectionState) {
			util.Printf(util.WORK, self.ToPeerId, "Peer Connection State has changed: ", s.String())

			if !self.releasePeer && s >= pwebrtc.PeerConnectionStateFailed {
				self.Info.DelConnectionInfo(self.Position, self.ToPeerId)
				self.ConnectObj.DisconnectFrom(self)
			}
		})

		self.dataChannel.OnOpen(func() {
			self.OnDataChannelOpen()
		})

		self.dataChannel.OnMessage(func(msg pwebrtc.DataChannelMessage) {
			self.OnDataChannelMessage(msg)
		})
	} else {
		self.peerConnection.OnConnectionStateChange(func(s pwebrtc.PeerConnectionState) {
			util.Println(util.WORK, self.ToPeerId, "Connection State has changed:", s.String())

			if !self.releasePeer && s >= pwebrtc.PeerConnectionStateFailed {
				self.Info.DelConnectionInfo(self.Position, self.ToPeerId)
				self.ConnectObj.DisconnectFrom(self)
			}
		})

		self.peerConnection.OnTrack(func(tr *pwebrtc.TrackRemote, r *pwebrtc.RTPReceiver) {
			util.Println(util.WORK, self.ToPeerId, "OnTrack!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", tr.Kind())

			self.setLocalTrack(tr, r)
		})

		self.peerConnection.OnDataChannel(func(d *pwebrtc.DataChannel) {
			self.dataChannel = d

			self.dataChannel.OnOpen(func() {
				self.OnDataChannelOpen()
			})

			self.dataChannel.OnMessage(func(msg pwebrtc.DataChannelMessage) {
				self.OnDataChannelMessage(msg)
			})
		})
	}

}

func (self *Peer) setLocalTrack(tr *pwebrtc.TrackRemote, r *pwebrtc.RTPReceiver) {

	if self.Info.OverlayInfo.OwnerId == self.Info.PeerId() {
		return
	}

	self.MediaReceive = true

	go func() {
		if self.releasePeer {
			return
		}

		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			if self.releasePeer {
				return
			}

			if !self.MediaReceive {
				return
			}

			if rtcpSendErr := self.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(tr.SSRC())}}); rtcpSendErr != nil {
				util.Println(util.ERROR, self.ToPeerId, "WriteRTCP error:", rtcpSendErr)
			}
		}
	}()

	codecType := tr.Kind()

	var lTrack *pwebrtc.TrackLocalStaticRTP = nil

	if codecType == pwebrtc.RTPCodecTypeVideo {
		lTrack = self.Info.GetTrack("video")

		if lTrack == nil {
			localTrack, newTrackErr := pwebrtc.NewTrackLocalStaticRTP(tr.Codec().RTPCodecCapability, "video", "pion")
			if newTrackErr != nil {
				panic(newTrackErr)
			}
			lTrack = localTrack
		}

		self.ConnectObj.OnTrack(self.ToPeerId, "video", lTrack)
	} else {
		lTrack = self.Info.GetTrack("audio")

		if lTrack == nil {
			localTrack, newTrackErr := pwebrtc.NewTrackLocalStaticRTP(tr.Codec().RTPCodecCapability, "audio", "pion")
			if newTrackErr != nil {
				panic(newTrackErr)
			}
			lTrack = localTrack
		}

		self.ConnectObj.OnTrack(self.ToPeerId, "audio", lTrack)
	}

	rtpBuf := make([]byte, 1400)
	for {
		if self.releasePeer {
			break
		}

		if !self.MediaReceive {
			break
		}

		if lTrack == nil {
			continue
		}

		tr.SetReadDeadline(time.Now().Add(time.Second * 5))
		//util.Println(util.INFO, self.ToPeerId, "!!!!!!!! read start !!!!!!!!")
		i, _, readErr := tr.Read(rtpBuf)
		//util.Println(util.INFO, self.ToPeerId, "!!!!!!!! read end !!!!!!!!", i)
		if readErr != nil {
			util.Println(util.ERROR, self.ToPeerId, "localtrack read error:", readErr)
			continue
		}

		//util.Println(util.INFO, self.ToPeerId, "!!!!!!!! write !!!!!!!!", i)

		// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
		if _, err := lTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			//panic(err)
			util.Println(util.ERROR, self.ToPeerId, "localtrack write error:", err)
			continue
		}
	}
}

func (self *Peer) OnDataChannelOpen() {
	util.Printf(util.INFO, "Data channel '%s'-'%d' open.\n", self.dataChannel.Label(), self.dataChannel.ID())

	*self.connectChan <- true
}

func (self *Peer) OnDataChannelMessage(msg pwebrtc.DataChannelMessage) {
	length := binary.BigEndian.Uint16(msg.Data[2:4])
	var unknown interface{}
	err := json.Unmarshal(msg.Data[4:length+4], &unknown)

	if err != nil {
		util.Println(util.ERROR, "datachannel msg parsing error :", err)
		return
	}

	hstruct := unknown.(map[string]interface{})

	//util.Println(util.INFO, self.ToPeerId, "DataChannel recv :", hstruct)

	if c, ok := hstruct["req-code"]; ok {
		code := int(c.(float64))

		switch code {
		case util.ReqCode_Hello:
			header := util.HelloPeer{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvHello(&header)

		case util.ReqCode_Estab:
			header := util.EstabPeer{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvEstab(&header)

		case util.ReqCode_Probe:
			header := util.ProbePeerRequest{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvProbe(&header)

		case util.ReqCode_Primary:
			header := util.PrimaryPeer{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvPrimary(&header)

		case util.ReqCode_Candidate:
			header := util.CandidatePeer{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvCandidate(&header)

		case util.ReqCode_BroadcastData:
			header := util.BroadcastData{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvBroadcastData(&header, msg.Data[length+4:])

		case util.ReqCode_Release:
			header := util.ReleasePeer{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvRelease(&header)

		case util.ReqCode_HeartBeat:
			header := util.HeartBeat{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvHeartBeat(&header)

		case util.ReqCode_ScanTree:
			header := util.ScanTree{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvScanTree(&header)

		case util.ReqCode_Buffermap:
			header := util.Buffermap{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvBuffermap(&header)

		case util.ReqCode_GetData:
			header := util.GetDataRequset{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvGetData(&header, msg.Data[length+4:])
		}
	}

	if c, ok := hstruct["rsp-code"]; ok {
		code := int(c.(float64))

		switch code {
		case util.RspCode_Hello:
			header := util.HelloPeerResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvHelloResponse(&header)

		case util.RspCode_Estab_Yes, util.RspCode_Estab_No:
			header := util.EstabPeerResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvEstabResponse(&header)

		case util.RspCode_Probe:
			header := util.ProbePeerResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvProbeResponse(&header)

		case util.RspCode_Primary_Yes, util.RspCode_Primary_No:
			headerAll := util.PrimaryPeerResponseAll{}
			//headerAll.Header = util.PrimaryPeerResponse{}
			json.Unmarshal(msg.Data[4:length+4], &headerAll.Header)
			//headerAll.ExtensionHeader = util.PrimaryPeerResponseExtensionHeader{}
			var curlen int = (int)(length + 4)
			json.Unmarshal(msg.Data[curlen:curlen+headerAll.Header.RspParams.ExtHeaderLen], &headerAll.ExtensionHeader)
			curlen += headerAll.Header.RspParams.ExtHeaderLen
			headerAll.Payload = msg.Data[curlen:]

			self.recvPrimaryResponse(&headerAll)

		case util.RspCode_Candidate:
			header := util.CandidatePeerResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvCandidateResponse(&header)

		case util.RspCode_BroadcastData:
			header := util.BroadcastDataResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvBroadcastDataResponse(&header)

		case util.RspCode_Release:
			header := util.ReleasePeerResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvReleaseAck(&header)

		case util.RspCode_HeartBeat:
			header := util.HeartBeatResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvHeartBeatResponse(&header)

		case util.RspCode_ScanTreeLeaf, util.RspCode_ScanTreeNonLeaf:
			header := util.ScanTreeResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvScanTreeResponse(&header)

		case util.RspCode_Buffermap:
			header := util.BuffermapResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			self.recvBuffermapResponse(&header)

		case util.RspCode_GetData:
			header := util.GetDataResponse{}
			json.Unmarshal(msg.Data[4:length+4], &header)
			go self.recvGetDataResponse(&header, msg.Data[length+4:])
		}
	}
}

func (self *Peer) CreateOffer() {
	if self.peerConnection == nil {
		self.newPeerConnection(true)
	}

	offer, err := self.peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	if err = self.peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	rsdp := util.RTCSessionDescription{}
	rsdp.Toid = self.ToPeerId
	rsdp.Type = "offer"
	rsdp.Sdp = offer

	//log.Printf("send offer to %s", self.ToPeerId)
	*self.signalSend <- rsdp
}

func (self *Peer) ReceiveOffer(rsdp util.RTCSessionDescription) {
	if self.peerConnection == nil {
		self.newPeerConnection(false)
	} else {
		trs := self.peerConnection.GetTransceivers()

		if trs != nil && len(trs) > 0 {
			for _, tr := range trs {
				if tr != nil && tr.Sender() != nil {
					self.peerConnection.RemoveTrack(tr.Sender())
				}
			}
		}
	}

	if err := self.peerConnection.SetRemoteDescription(rsdp.Sdp); err != nil {
		panic(err)
	}

	answer, err := self.peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	ranswer := util.RTCSessionDescription{}
	ranswer.Sdp = answer
	ranswer.Toid = self.ToPeerId
	ranswer.Type = "answer"

	*self.signalSend <- ranswer

	err = self.peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	self.candidatesMux.Lock()
	for _, c := range self.pendingCandidates {
		if self.releasePeer {
			break
		}

		onICECandidateErr := self.signalCandidate(c)
		if onICECandidateErr != nil {
			panic(onICECandidateErr)
		}
	}
	self.candidatesMux.Unlock()
}

func (self *Peer) SendDataChannelString(msg string) {
	if self.dataChannel != nil {
		self.dataChannel.SendText(msg)
	}
}

func (self *Peer) SendDataChannelBytes(msg []byte) error {
	if self.dataChannel != nil {
		return self.dataChannel.Send(msg)
	}

	return fmt.Errorf("dataChannel is nil")
}

func (self *Peer) sendPPMessage(msg *util.PPMessage) error {
	if self.dataChannel != nil {

		if self.dataChannel.ReadyState() >= pwebrtc.DataChannelStateClosing {
			util.Println(util.ERROR, self.ToPeerId, "dataChannel closed")
			return fmt.Errorf("dataChannel closed")
		}

		buf := []byte{}
		buf = append(buf, msg.Ver)
		buf = append(buf, msg.Type)
		lenbytes := [2]byte{}
		binary.BigEndian.PutUint16(lenbytes[:], msg.Length)
		buf = append(buf, lenbytes[:]...)
		buf = append(buf, []byte(msg.Header)...)

		if len(msg.ExtensionHeader) > 0 {
			buf = append(buf, []byte(msg.ExtensionHeader)...)
		}

		if msg.Payload != nil && len(msg.Payload) > 0 {
			buf = append(buf, msg.Payload...)
		}

		return self.dataChannel.Send(buf)
	}

	util.Println(util.ERROR, self.ToPeerId, "dataChannel is nil")

	return fmt.Errorf("dataChannel is nil")
}

func (self *Peer) AddConnectionInfo(position connect.PeerPosition) {
	self.Position = position

	self.Info.AddConnectionInfo(position, self.ToPeerId)
}

func (self *Peer) sendHello() bool {
	hello := util.HelloPeer{}

	hello.ReqCode = util.ReqCode_Hello
	hello.ReqParams.Operation.OverlayId = self.Info.OverlayInfo.OverlayId
	hello.ReqParams.Operation.ConnNum = self.Info.PeerConfig.EstabPeerMaxCount //TODO check
	hello.ReqParams.Operation.Ttl = self.Info.PeerConfig.HelloPeerTTL

	hello.ReqParams.Peer.PeerId = self.Info.PeerId()
	hello.ReqParams.Peer.Address = self.Info.PeerInfo.Address
	hello.ReqParams.Peer.TicketId = self.Info.PeerInfo.TicketId

	msg := util.GetPPMessage(&hello, nil, nil)
	self.sendPPMessage(msg)
	util.Println(util.WORK, self.ToPeerId, "Send Hello :", msg)

	timeout := time.NewTimer(time.Second * 5)

	select {
	case res := <-self.ppChan:
		if !timeout.Stop() {
			<-timeout.C
		}

		if hres, ok := res.(*util.HelloPeerResponse); ok {
			if hres.RspCode != util.RspCode_Hello {
				util.Println(util.ERROR, self.ToPeerId, "Wrong response:", res)
				return false
			}

			util.Println(util.WORK, self.ToPeerId, "Recv Hello response:", hres)
		} else {
			util.Println(util.ERROR, self.ToPeerId, "Wrong response:", res)
			return false
		}

		return true

	case <-timeout.C:
		return false
	}
}

func (self *Peer) sendVerticalHello() bool {
	hello := util.HelloPeer{}

	hello.ReqCode = util.ReqCode_Hello
	hello.ReqParams.Operation.OverlayId = self.Info.OverlayInfo.OverlayId
	hello.ReqParams.Operation.ConnNum = self.Info.PeerConfig.EstabPeerMaxCount
	hello.ReqParams.Operation.Ttl = 1
	vc := true
	hello.ReqParams.Operation.VerticalCandidate = &vc

	hello.ReqParams.Peer.PeerId = self.Info.PeerId()
	hello.ReqParams.Peer.Address = self.Info.PeerInfo.Address
	hello.ReqParams.Peer.TicketId = self.Info.PeerInfo.TicketId

	msg := util.GetPPMessage(&hello, nil, nil)
	self.sendPPMessage(msg)
	util.Println(util.WORK, self.ToPeerId, "Send vertical-hello :", msg)

	timeout := time.NewTimer(time.Second * 5)

	select {
	case res := <-self.ppChan:
		if !timeout.Stop() {
			<-timeout.C
		}

		if hres, ok := res.(*util.HelloPeerResponse); ok {
			if hres.RspCode != util.RspCode_Hello {
				util.Println(util.ERROR, self.ToPeerId, "Wrong response:", res)
				return false
			}

			util.Println(util.WORK, self.ToPeerId, "Recv Hello response:", hres)
		} else {
			util.Println(util.ERROR, self.ToPeerId, "Wrong response:", res)
			return false
		}

		return true

	case <-timeout.C:
		return false
	}
}

func (self *Peer) RelayHello(hello *util.HelloPeer) bool {
	msg := util.GetPPMessage(&hello, nil, nil)
	self.sendPPMessage(msg)
	util.Println(util.WORK, self.ToPeerId, "Send Hello :", msg)

	select {
	case res := <-self.ppChan:
		if hres, ok := res.(*util.HelloPeerResponse); ok {
			util.Println(util.WORK, self.ToPeerId, "Recv Hello response:", hres)
			if hres.RspCode != util.RspCode_Hello {
				return false
			}
		} else {
			util.Println(util.ERROR, self.ToPeerId, "Wrong response:", res)
			return false
		}

		return true

	case <-time.After(time.Second * 5):
		return false
	}
}

func (self *Peer) recvHello(hello *util.HelloPeer) {
	util.Println(util.WORK, self.ToPeerId, "Recv Hello :", hello)

	if hello.ReqParams.Peer.TicketId < self.Info.PeerInfo.TicketId {
		util.Println(util.WORK, "TicketId less then mine. ignore.")

		if self.Position == connect.InComing {
			self.SendRelease()
			self.ConnectObj.DisconnectFrom(self)
		}

		return
	}

	for _, block := range *self.ConnectObj.ServiceInfo().BlockList {
		if block == hello.ReqParams.Peer.PeerId {
			util.Println(util.WORK, self.ToPeerId, "PeerId is blocked. send release.")
			self.SendRelease()
			self.ConnectObj.DisconnectFrom(self)
			return
		}
	}

	self.sendHelloResponse()
	<-time.After(time.Millisecond * 200)

	util.Println(util.INFO, self.ToPeerId, "position :", self.Position)

	if self.Position == connect.InComing {
		self.SendRelease()
		self.ConnectObj.DisconnectFrom(self)
	}

	*self.broadcastChan <- hello
}

func (self *Peer) sendHelloResponse() {
	hello := util.HelloPeerResponse{}
	hello.RspCode = util.RspCode_Hello

	util.Println(util.WORK, self.ToPeerId, "Send Hello resp :", hello)

	msg := util.GetPPMessage(&hello, nil, nil)
	self.sendPPMessage(msg)
}

func (self *Peer) recvHelloResponse(res *util.HelloPeerResponse) bool {
	util.Println(util.WORK, self.ToPeerId, "recvHelloResponse:", res)
	self.ppChan <- res
	return true
}

func (self *Peer) SendEstab(allowPrimary bool) bool {
	estab := util.EstabPeer{}
	estab.ReqCode = util.ReqCode_Estab
	estab.ReqParams.Operation.OverlayId = self.Info.OverlayInfo.OverlayId
	estab.ReqParams.Peer.PeerId = self.Info.PeerId()
	estab.ReqParams.Peer.TicketId = self.Info.PeerInfo.TicketId
	estab.ReqParams.AllowPrimaryReq = allowPrimary

	msg := util.GetPPMessage(&estab, nil, nil)
	self.sendPPMessage(msg)
	util.Println(util.WORK, self.ToPeerId, "Send Estab :", msg)

	select {
	case res := <-self.ppChan:
		if hres, ok := res.(*util.EstabPeerResponse); ok {
			util.Println(util.WORK, self.ToPeerId, "Recv Estab response:", hres)
			if hres.RspCode != util.RspCode_Estab_Yes {
				return false
			}
		} else {
			return false
		}

		return true

	case <-time.After(time.Second * 5):
		return false
	}
}

func (self *Peer) recvEstab(estab *util.EstabPeer) {
	util.Println(util.WORK, self.ToPeerId, "Recv Estab :", estab)

	if self.Info.OverlayInfo.OverlayId != estab.ReqParams.Operation.OverlayId {
		util.Println(util.WORK, self.ToPeerId, "Estab OverlayId not match")
		self.SendEstabResponse(false)
		self.ConnectObj.DisconnectFrom(self)
		return
	}

	self.ToTicketId = estab.ReqParams.Peer.TicketId
	self.isVerticalCandidate = estab.ReqParams.Peer.PeerId == self.Info.GrandParentId

	util.Println(util.INFO, self.ToPeerId, "----------------------------------------------------------------")
	util.Println(util.INFO, self.ToPeerId, "estab.ReqParams.Peer.PeerId:", estab.ReqParams.Peer.PeerId)
	util.Println(util.INFO, self.ToPeerId, "self.Info.GrandParentId:", self.Info.GrandParentId)
	util.Println(util.INFO, self.ToPeerId, "self.isVerticalCandidate:", self.isVerticalCandidate)
	util.Println(util.INFO, self.ToPeerId, "----------------------------------------------------------------")

	if self.Info.PeerConfig.ProbePeerTimeout > 0 {
		if self.Info.PeerConfig.EstabPeerMaxCount <= self.Info.EstabPeerCount && !self.isVerticalCandidate {
			util.Println(util.WORK, self.ToPeerId, "EstabPeerMaxCount <= EstabPeerCount")
			self.SendEstabResponse(false)
			self.ConnectObj.DisconnectFrom(self)
		} else {
			self.Info.EstabPeerCount++
			self.Position = connect.Established
			self.SendEstabResponse(true)
			go self.sendHeartBeat()
		}
	} else {
		self.Info.OutGoingPrimaryMux.Lock()
		if self.Info.HaveOutGoingPrimary {
			if self.candidateCheck(true) {
				go self.sendHeartBeat()
			}
		} else {
			self.SendEstabResponse(true)
			if estab.ReqParams.AllowPrimaryReq && self.setPrimary() {
				go self.sendHeartBeat()
			} else {
				if self.candidateCheck(false) {
					go self.sendHeartBeat()
				}
			}
		}
		self.Info.OutGoingPrimaryMux.Unlock()
	}
}

func (self *Peer) TryPrimary() {
	defer self.Info.OutGoingPrimaryMux.Unlock()

	self.Info.OutGoingPrimaryMux.Lock()
	if self.Info.HaveOutGoingPrimary {
		self.candidateCheck(false)
	} else {
		if !self.setPrimary() {
			self.candidateCheck(false)
		}
	}
}

func (self *Peer) closeLowestOutGoingCandidate() {
	util.Println(util.INFO, "closeLowestOutGoingCandidate!!!!!!!!!!!!!!!!!!!!!!!!!!")
	peer := self.ConnectObj.lowestOutGoingCandidate()
	if peer != nil {
		util.Println(util.INFO, "!!!!!!!!!!!!!!!!!!!LowestOutGoingCandidate", peer.ToPeerId)
		self.Info.DelConnectionInfo(peer.Position, peer.ToPeerId)
		self.ConnectObj.DisconnectFrom(peer)
	}
}

func (self *Peer) candidateCheck(sendres bool) bool {
	if self.Info.PeerStatus.NumOutCandidate >= self.Info.PeerConfig.MaxOutgoingCandidate {
		if self.isVerticalCandidate {
			self.closeLowestOutGoingCandidate()
			return self.candidateCheck(sendres)
		}
		if sendres {
			self.SendEstabResponse(false)
		}

		self.SendRelease()
		<-time.After(time.Second * 1)
		self.ConnectObj.DisconnectFrom(self)

		return false
	} else {
		if sendres {
			self.SendEstabResponse(true)
		}
		self.AddConnectionInfo(connect.OutGoingCandidate)

		if self.Info.PeerConfig.SendCandidate {
			self.setCandidate()
		}

		return true
	}
}

func (self *Peer) SendEstabResponse(ok bool) {
	res := util.EstabPeerResponse{}

	if ok {
		res.RspCode = util.RspCode_Estab_Yes
	} else {
		res.RspCode = util.RspCode_Estab_No
	}

	util.Println(util.WORK, self.ToPeerId, "Send Estab resp :", res)

	msg := util.GetPPMessage(&res, nil, nil)
	self.sendPPMessage(msg)
}

func (self *Peer) recvEstabResponse(res *util.EstabPeerResponse) {
	self.ppChan <- res
}

func (self *Peer) setPrimary() bool {
	primary := util.PrimaryPeer{}
	primary.ReqCode = util.ReqCode_Primary

	if self.Info.OverlayInfo.CrPolicy != nil && self.Info.OverlayInfo.CrPolicy.RecoveryBy == connect.RecoveryByPush {
		primary.ReqParams.Buffermap = self.ConnectObj.GetBuffermap()
	}

	msg := util.GetPPMessage(&primary, nil, nil)
	self.sendPPMessage(msg)
	util.Println(util.WORK, self.ToPeerId, "Set primary:", msg)

	timeout := time.NewTimer(time.Second * 5)

	select {
	case res := <-self.ppChan:

		if !timeout.Stop() {
			<-timeout.C
		}

		var hres *util.PrimaryPeerResponseAll = nil
		var ok bool

		if hres, ok = res.(*util.PrimaryPeerResponseAll); ok {
			util.Println(util.WORK, self.ToPeerId, "Recv Primary response:", hres)
		} else {
			return false
		}

		header := hres.Header

		if header.RspCode != util.RspCode_Primary_Yes {
			return false
		}

		self.AddConnectionInfo(connect.OutGoingPrimary)

		if self.Info.OverlayInfo.CrPolicy != nil {
			if self.Info.OverlayInfo.CrPolicy.RecoveryBy == connect.RecoveryByPush {
				self.ConnectObj.checkBuffermapForBroadcast(header.RspParams.Buffermap)
			} else if self.Info.OverlayInfo.CrPolicy.RecoveryBy == connect.RecoveryByPull {
				self.ConnectObj.recoveryDataForPull()
			}
		}

		extHeader := hres.ExtensionHeader

		self.checkJoinPeers(extHeader.Peers, hres.Payload)

		self.sendJoinNoti()

		self.checkMedia()

		if header.RspParams.ParentPeer != nil {
			util.Println(util.INFO, self.ToPeerId, "Parent peer !!!!!!!!!!!!!")
			util.Println(util.INFO, self.ToPeerId, "--->", header.RspParams.ParentPeer.PeerId, header.RspParams.ParentPeer.Address)

			self.Info.GrandParentId = header.RspParams.ParentPeer.PeerId

			go self.ConnectObj.sendVerticalHello(header.RspParams.ParentPeer.PeerId)
		}

		return true

	case <-timeout.C:
		util.Println(util.ERROR, self.ToPeerId, "Recv Primary timeout!!!")
		return false
	}
}

func (self *Peer) sendJoinNoti() {
	if !self.ConnectObj.IsFirstPrimary {
		return
	}

	self.ConnectObj.IsFirstPrimary = false

	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.Info.PeerId()
	params.Peer.Sequence = self.ConnectObj.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControlJoin{}
	extHeader.ChannelId = self.ConnectObj.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeJoin
	extHeader.JoinerInfo.PeerId = self.Info.PeerOriginId()
	extHeader.JoinerInfo.DisplayName = *self.Info.PeerInfo.DisplayName
	extHeader.JoinerInfo.PublicKey = self.Info.GetPublicKeyString()

	payload := self.ConnectObj.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, self.ToPeerId, "SignStruct error!! payload is nil")
		return
	}

	self.ConnectObj.BroadcastData(&params, &extHeader, &payload, self.Info.PeerId(), false, true)
}

func (self *Peer) checkJoinPeers(peers []util.PrimaryPeerResponseExtensionHeaderPeers, payload []byte) {
	if peers != nil && len(peers) > 0 {
		self.ConnectObj.JoinPeerMux.Lock()

		joinPeers := *self.ConnectObj.GetJoinPeers()
		var pos int = 0

		var joinPeersForNoti []*util.JoinPeerInfo = make([]*util.JoinPeerInfo, 0)

		for _, peer := range peers {
			if _, ok := joinPeers[peer.PeerId]; !ok {
				joinPeers[peer.PeerId] = &util.JoinPeerInfo{}
				joinPeers[peer.PeerId].PublicKey = nil
				joinPeersForNoti = append(joinPeersForNoti, joinPeers[peer.PeerId])
			}
			joinPeers[peer.PeerId].PeerId = peer.PeerId
			joinPeers[peer.PeerId].DisplayName = peer.DisplayName

			if peer.PublicKey && joinPeers[peer.PeerId].PublicKey == nil {
				joinPeers[peer.PeerId].PublicKeyPEM = payload[pos : pos+459]
				pos += 459

				joinPeers[peer.PeerId].PublicKey = self.ConnectObj.ParsePublicKey((string)(joinPeers[peer.PeerId].PublicKeyPEM))

				if joinPeers[peer.PeerId].PublicKey == nil {
					joinPeers[peer.PeerId].PublicKeyPEM = []byte{}
					joinPeers[peer.PeerId].PublicKey = nil
				}
			}

			if len(peer.CachedData) > 0 {
				for _, media := range peer.CachedData {
					var find bool = false
					for _, cache := range joinPeers[peer.PeerId].CachingMedia {
						if cache.ChannelId == media.ChannelId {
							find = true
							if cache.Sequence < media.Sequence {
								cache.Sequence = media.Sequence
								cache.Media = []byte{}
								cache.SignMedia = []byte{}
							}
							break
						}
					}

					if !find {
						joinPeers[peer.PeerId].CachingMedia = append(joinPeers[peer.PeerId].CachingMedia, util.JoinPeerInfoCachingMedia{
							ChannelId: media.ChannelId,
							Sequence:  media.Sequence,
							Media:     []byte{},
							SignMedia: []byte{},
						})
					}
				}
			}
		}

		self.ConnectObj.JoinPeerMux.Unlock()

		for _, peer := range joinPeersForNoti {
			if self.ConnectObj.PeerChangeCallback != nil {
				util.Println(util.INFO, self.ToPeerId, "PeerChangeCallback !!!!!!!!!!!!!")
				self.ConnectObj.PeerChangeCallback(self.Info.OverlayInfo.OverlayId, peer.PeerId, peer.DisplayName, false)
			}
		}
	}
}

func (self *Peer) checkMedia() {
	self.ConnectObj.JoinPeerMux.Lock()
	defer self.ConnectObj.JoinPeerMux.Unlock()

	joinPeers := *self.ConnectObj.GetJoinPeers()

	for _, peer := range joinPeers {
		for _, media := range peer.CachingMedia {
			if media.Sequence > 0 && len(media.SignMedia) == 0 {
				self.sendGetData(peer.PeerId, media.Sequence, media.ChannelId)
			}
		}
	}
}

func (self *Peer) recvPrimary(primary *util.PrimaryPeer) {
	self.Info.CommunicationMux.Lock()
	util.Println(util.WORK, self.ToPeerId, "Recv primary :", primary)

	if self.Info.PeerConfig.MaxPrimaryConnection <= self.Info.PeerStatus.NumPrimary {
		self.Info.CommunicationMux.Unlock()
		util.Println(util.WORK, self.ToPeerId, "MaxPrimaryConnection <= NumPrimary")
		self.sendPrimaryResponse(false)
	} else {
		util.Println(util.WORK, self.ToPeerId, "NumPrimary :", self.Info.PeerStatus.NumPrimary)
		self.sendPrimaryResponse(true)
		self.AddConnectionInfo(connect.InComingPrimary)

		self.Info.CommunicationMux.Unlock()

		if self.Info.OverlayInfo.CrPolicy != nil {
			if self.Info.OverlayInfo.CrPolicy.RecoveryBy == connect.RecoveryByPush {
				self.ConnectObj.checkBuffermapForBroadcast(primary.ReqParams.Buffermap)
			}
		}

		/*
			needReOffer := false

			if self.Info.VideoTrack != nil {
				_, err := self.peerConnection.AddTrack(self.Info.VideoTrack)

				if err != nil {
					panic(err)
				}

				needReOffer = true
			}

			if self.Info.AudioTrack != nil {
				_, err := self.peerConnection.AddTrack(self.Info.AudioTrack)

				if err != nil {
					panic(err)
				}

				needReOffer = true
			}

			if needReOffer {
				self.CreateOffer()
			}*/
	}

}

func (self *Peer) sendPrimaryResponse(ok bool) {
	pres := util.PrimaryPeerResponse{}

	if ok {
		pres.RspCode = util.RspCode_Primary_Yes

		if self.Info.OverlayInfo.CrPolicy.RecoveryBy == connect.RecoveryByPush {
			pres.RspParams.Buffermap = self.ConnectObj.GetBuffermap()
		}

		if self.Info.HaveOutGoingPrimary {
			parent := self.ConnectObj.outGoingPrimary()
			pres.RspParams.ParentPeer = new(util.ParentPeer)
			pres.RspParams.ParentPeer.PeerId = parent.ToPeerId
			pres.RspParams.ParentPeer.Address = parent.Info.PeerInfo.Address
		}

		extHeader := util.PrimaryPeerResponseExtensionHeader{}
		extHeader.Peers = make([]util.PrimaryPeerResponseExtensionHeaderPeers, 0)
		payload := []byte{}

		self.ConnectObj.JoinPeerMux.Lock()

		joinpeers := self.ConnectObj.GetJoinPeers()

		for _, joinpeer := range *joinpeers {
			extPeer := util.PrimaryPeerResponseExtensionHeaderPeers{
				PeerId:      joinpeer.PeerId,
				DisplayName: joinpeer.DisplayName,
			}

			if len(joinpeer.PublicKeyPEM) > 0 {
				extPeer.PublicKey = true
				payload = append(payload, joinpeer.PublicKeyPEM...)
			} else {
				extPeer.PublicKey = false
			}

			extPeer.CachedData = make([]util.PrimaryPeerResponseExtensionHeaderPeersCachingMedia, 0)
			if len(joinpeer.CachingMedia) > 0 {
				for _, cachingmedia := range joinpeer.CachingMedia {
					extCachingMedia := util.PrimaryPeerResponseExtensionHeaderPeersCachingMedia{
						ChannelId: cachingmedia.ChannelId,
						Sequence:  cachingmedia.Sequence,
					}

					extPeer.CachedData = append(extPeer.CachedData, extCachingMedia)
				}
			}

			extHeader.Peers = append(extHeader.Peers, extPeer)
		}

		self.ConnectObj.JoinPeerMux.Unlock()

		util.Println(util.WORK, self.ToPeerId, "Send Primary resp :", pres)
		msg := util.GetPPMessage(&pres, &extHeader, payload)
		self.sendPPMessage(msg)
	} else {
		pres.RspCode = util.RspCode_Primary_No
		util.Println(util.WORK, self.ToPeerId, "Send Primary resp :", pres)
		msg := util.GetPPMessage(&pres, nil, nil)
		self.sendPPMessage(msg)
	}
}

func (self *Peer) recvPrimaryResponse(res *util.PrimaryPeerResponseAll) {
	self.ppChan <- res
}

func (self *Peer) setCandidate() {
	candi := util.CandidatePeer{}
	candi.ReqCode = util.ReqCode_Candidate

	msg := util.GetPPMessage(&candi, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Set candidate:", msg)
}

func (self *Peer) recvCandidate(res *util.CandidatePeer) {
	util.Println(util.WORK, self.ToPeerId, "Recv candidate:", res)

	self.sendCandidateResponse()
}

func (self *Peer) sendCandidateResponse() {
	candi := util.CandidatePeerResponse{}
	candi.RspCode = util.RspCode_Candidate

	msg := util.GetPPMessage(&candi, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send candidate resp:", msg)
}

func (self *Peer) recvCandidateResponse(res *util.CandidatePeerResponse) {
	util.Println(util.WORK, self.ToPeerId, "Recv candidate resp:", res)
}

func (self *Peer) SendProbe() {
	pro := util.ProbePeerRequest{}
	pro.ReqCode = util.ReqCode_Probe
	pro.ReqParams.Operation.NtpTime = time.Now().Format("2006-01-02 15:04:05.000")

	msg := util.GetPPMessage(&pro, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send probe:", msg)
}

func (self *Peer) recvProbe(req *util.ProbePeerRequest) {
	res := util.ProbePeerResponse{}
	res.RspCode = util.RspCode_Probe
	res.RspParams.Operation.NtpTime = req.ReqParams.Operation.NtpTime

	//<-time.After(time.Second * time.Duration(1))

	msg := util.GetPPMessage(&res, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send probe response:", msg)
}

func (self *Peer) recvProbeResponse(res *util.ProbePeerResponse) {
	util.Println(util.WORK, self.ToPeerId, "Recv probe response:", res)

	loc, _ := time.LoadLocation("Asia/Seoul")
	prbtime, err := time.ParseInLocation("2006-01-02 15:04:05.000", res.RspParams.Operation.NtpTime, loc)
	if err != nil {
		util.Println(util.ERROR, self.ToPeerId, "probetime parse error:", err)
		self.probeTime = nil
	} else {
		now := time.Now().In(loc)
		duration := int64(now.Sub(prbtime) / time.Millisecond)
		self.probeTime = &duration
		util.Println(util.WORK, self.ToPeerId, "probe time:", duration, "ms")
	}
}

func (self *Peer) SendRelease() {
	rel := util.ReleasePeer{}
	rel.ReqCode = util.ReqCode_Release
	rel.ReqParams.Operation.Ack = self.Info.PeerConfig.ReleaseOperationAck

	msg := util.GetPPMessage(&rel, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send release peer:", msg)
}

func (self *Peer) recvRelease(req *util.ReleasePeer) {
	util.Println(util.WORK, self.ToPeerId, "Recv release peer:", req)

	if req.ReqParams.Operation.Ack {
		self.sendReleaseAck()
	}

	self.Info.DelConnectionInfo(self.Position, self.ToPeerId)
	self.ConnectObj.DisconnectFrom(self)
}

func (self *Peer) sendReleaseAck() {
	res := util.ReleasePeerResponse{}
	res.RspCode = util.RspCode_Release

	msg := util.GetPPMessage(&res, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send release ack:", msg)
}

func (self *Peer) recvReleaseAck(res *util.ReleasePeerResponse) {
	util.Println(util.WORK, self.ToPeerId, "Recv release ack:", res)

	self.releasePeer = true

}

func (self *Peer) sendHeartBeat() {
	interval := float32(self.Info.OverlayInfo.HeartbeatInterval) * 0.9

	for range time.Tick(time.Second * time.Duration(interval)) {
		if self.releasePeer {
			return
		}

		req := util.HeartBeat{}
		req.ReqCode = util.ReqCode_HeartBeat
		msg := util.GetPPMessage(&req, nil, nil)
		err := self.sendPPMessage(msg)

		if err != nil {
			util.Println(util.WORK, self.ToPeerId, "Send heartbeat failed!", err)
			self.Info.DelConnectionInfo(self.Position, self.ToPeerId)
			self.ConnectObj.DisconnectFrom(self)
		} else {
			//util.Println(util.WORK, self.ToPeerId, "Send heartbeat:", msg)
		}
	}
}

func (self *Peer) CheckHeartBeatRecved() {
	go func() {
		for range time.Tick(time.Second * 1) {
			if self.releasePeer {
				return
			}

			self.heartbeatCount++

			if self.heartbeatCount > self.Info.OverlayInfo.HeartbeatTimeout {
				util.Println(util.WORK, self.ToPeerId, "Heartbeat timeout!")
				self.Info.DelConnectionInfo(self.Position, self.ToPeerId)
				self.ConnectObj.DisconnectFrom(self)
			}

		}
	}()
}

func (self *Peer) recvHeartBeat(req *util.HeartBeat) {
	//util.Println(util.WORK, self.ToPeerId, "recv heartbeat:", req)
	self.heartbeatCount = 0
	self.sendHeartBeatResponse()
}

func (self *Peer) sendHeartBeatResponse() {
	res := util.HeartBeatResponse{}
	res.RspCode = util.RspCode_HeartBeat
	msg := util.GetPPMessage(&res, nil, nil)
	self.sendPPMessage(msg)

	//util.Println(util.WORK, self.ToPeerId, "Send heartbeat response:", msg)
}

func (self *Peer) recvHeartBeatResponse(res *util.HeartBeatResponse) {
	//util.Println(util.WORK, self.ToPeerId, "recv heartbeat response:", res)
}

func (self *Peer) sendScanTree(cseq int) {
	req := util.ScanTree{}
	req.ReqCode = util.ReqCode_ScanTree
	req.ReqParams.CSeq = cseq
	req.ReqParams.Overlay.OverlayId = self.Info.OverlayInfo.OverlayId
	req.ReqParams.Overlay.Via = append(req.ReqParams.Overlay.Via, []string{self.Info.PeerId(), self.Info.PeerInfo.Address})
	//req.ReqParams.Overlay.Path = append(req.ReqParams.Overlay.Path, []string{self.Info.PeerId(), strconv.Itoa(*self.Info.OverlayInfo.TicketId), self.Info.PeerAddress})
	req.ReqParams.Peer.PeerId = self.Info.PeerId()
	req.ReqParams.Peer.TicketId = self.Info.PeerInfo.TicketId
	req.ReqParams.Peer.Address = self.Info.PeerInfo.Address

	msg := util.GetPPMessage(&req, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send ScanTree:", msg)
}

func (self *Peer) broadcastScanTree(req *util.ScanTree) {
	msg := util.GetPPMessage(req, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Broadcast ScanTree:", msg)
}

func (self *Peer) recvScanTree(req *util.ScanTree) {
	util.Println(util.WORK, self.ToPeerId, "Recv ScanTree:", req)

	self.ConnectObj.RecvScanTree(req, self)
}

func (self *Peer) sendScanTreeResponse(res *util.ScanTreeResponse) {
	msg := util.GetPPMessage(res, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send ScanTree response:", msg)
}

func (self *Peer) recvScanTreeResponse(res *util.ScanTreeResponse) {
	util.Println(util.WORK, self.ToPeerId, "Recv ScanTree response:", res)

	if res.RspCode == util.RspCode_ScanTreeNonLeaf {
		return
	}

	self.ConnectObj.RecvScanTreeResponse(res)
}

func (self *Peer) recvBroadcastData(req *util.BroadcastData, data []byte) {
	util.Println(util.WORK, self.ToPeerId, "Recv broadcast data:", req)
	//util.Println(util.WORK, self.ToPeerId, "Recv broadcast data content:", data)

	params := req.ReqParams

	if params.Operation.Ack {
		self.sendBroadcastDataResponse()
	}

	if !self.ConnectObj.checkBroadcastDataSeq(params.Peer.PeerId, params.Peer.Sequence) {
		util.Println(util.WORK, self.ToPeerId, "Already recv broadcast data seq:", params.Peer.PeerId, " - ", params.Peer.Sequence)
		return
	}

	if data == nil || len(data) <= 0 {
		return
	}

	if params.Peer.PeerId == self.Info.PeerId() {
		return
	}

	if params.ExtHeaderLen > 0 {
		extbuf := data[:params.ExtHeaderLen]
		payload := data[params.ExtHeaderLen:]
		var extHeader interface{} = nil

		baseExtHeader := util.BroadcastDataExtensionHeader{}
		json.Unmarshal(extbuf, &baseExtHeader)

		if baseExtHeader.MediaType == util.MediaTypeControl {
			controlExtHeader := util.BroadcastDataExtensionHeaderControl{}
			json.Unmarshal(extbuf, &controlExtHeader)

			alreadyBroadcast := false

			if controlExtHeader.ControlType == util.ControlTypeJoin {
				joinExtHeader := util.BroadcastDataExtensionHeaderControlJoin{}
				json.Unmarshal(extbuf, &joinExtHeader)

				if joinExtHeader.JoinerInfo.PeerId == self.Info.PeerOriginId() {
					return
				}

				publicKey := self.ConnectObj.ParsePublicKey((string)(joinExtHeader.JoinerInfo.PublicKey))

				if !self.verifyData(&extbuf, &payload, publicKey) {
					return
				}

				self.checkJoinPeer(&joinExtHeader.JoinerInfo)

				extHeader = &joinExtHeader
			} else if controlExtHeader.ControlType == util.ControlTypeLeave {
				leaveExtHeader := util.BroadcastDataExtensionHeaderControlLeave{}
				json.Unmarshal(extbuf, &leaveExtHeader)

				if leaveExtHeader.LeaverInfo.PeerId == self.Info.PeerOriginId() {
					return
				}

				publicKey := self.ConnectObj.GetJoinPeerPublicKey(leaveExtHeader.LeaverInfo.PeerId)

				if !self.verifyData(&extbuf, &payload, publicKey) {
					return
				}

				self.leaveJoinPeer(&leaveExtHeader.LeaverInfo)

				extHeader = &leaveExtHeader
			} else {
				if !self.ConnectObj.CheckOwner(params.Peer.PeerId) {
					util.Println(util.WORK, self.ToPeerId, "Control data recv but not owner:", params.Peer.PeerId)
					return
				}

				publicKey := self.ConnectObj.GetJoinPeerPublicKey(params.Peer.PeerId)

				if controlExtHeader.ControlType == util.ControlTypeTransmission {
					transExtHeader := util.BroadcastDataExtensionHeaderControlTransmission{}
					json.Unmarshal(extbuf, &transExtHeader)

					if !self.verifyData(&extbuf, &payload, publicKey) {
						return
					}

					self.ConnectObj.ApplyTransmissionNoti(&transExtHeader, true)
				} else if controlExtHeader.ControlType == util.ControlTypeOwnership {
					ownerExtHeader := util.BroadcastDataExtensionHeaderControlOwnership{}
					json.Unmarshal(extbuf, &ownerExtHeader)

					if !self.verifyData(&extbuf, &payload, publicKey) {
						return
					}

					self.ConnectObj.ApplyOwnershipNoti(&ownerExtHeader, true)
				} else if controlExtHeader.ControlType == util.ControlTypeSessionInfo {
					sessionExtHeader := util.BroadcastDataExtensionHeaderControlSessionInfo{}
					json.Unmarshal(extbuf, &sessionExtHeader)

					if !self.verifyData(&extbuf, &payload, publicKey) {
						return
					}

					self.ConnectObj.ApplySessionInfoNoti(&sessionExtHeader, true)
				} else if controlExtHeader.ControlType == util.ControlTypeExpulsion {
					expulsionExtHeader := util.BroadcastDataExtensionHeaderControlExpulsion{}
					json.Unmarshal(extbuf, &expulsionExtHeader)

					if !self.verifyData(&extbuf, &payload, publicKey) {
						return
					}

					expList := make([]string, 0)
					for _, exp := range expulsionExtHeader.ExpulsionList {
						expList = append(expList, exp.PeerId)
					}

					self.ConnectObj.BroadcastData(&params, extHeader, &payload, params.Peer.PeerId, false, true)
					alreadyBroadcast = true
					self.ConnectObj.ApplyExpulsionNoti(&expList, true)
				} else if controlExtHeader.ControlType == util.ControlTypeTermination {
					if !self.verifyData(&extbuf, &payload, publicKey) {
						return
					}

					self.ConnectObj.BroadcastData(&params, extHeader, &payload, params.Peer.PeerId, false, true)
					alreadyBroadcast = true
					self.ConnectObj.ApplyTerminationNoti(true)
				}
			}

			if !alreadyBroadcast {
				go self.ConnectObj.BroadcastData(&params, extHeader, &payload, params.Peer.PeerId, false, true)
			}
		} else if baseExtHeader.MediaType == util.MediaTypeData {
			dataExtHeader := util.BroadcastDataExtensionHeader{}
			json.Unmarshal(extbuf, &dataExtHeader)

			if !self.ConnectObj.CheckSource(params.Peer.PeerId, dataExtHeader.ChannelId) {
				util.Println(util.WORK, self.ToPeerId, "Data recv but not source:", params.ExtHeaderLen)
				return
			}

			publicKey := self.ConnectObj.GetJoinPeerPublicKey(params.Peer.PeerId)

			signed := payload[len(payload)-256:]
			data := payload[:len(payload)-256]

			if !self.verifyData(&data, &signed, publicKey) {
				return
			}

			self.ConnectObj.NotiData(self.Info.OverlayInfo.OverlayId, dataExtHeader.ChannelId, self.ConnectObj.GetOriginId(params.Peer.PeerId), dataExtHeader.MediaType, &data)

			go self.ConnectObj.BroadcastData(&params, &dataExtHeader, &payload, params.Peer.PeerId, false, true)
		} else if baseExtHeader.MediaType == util.MediaTypeDataCache {
			dataExtHeader := util.BroadcastDataExtensionHeaderDataCache{}
			json.Unmarshal(extbuf, &dataExtHeader)

			if !self.ConnectObj.CheckSource(params.Peer.PeerId, dataExtHeader.ChannelId) {
				util.Println(util.WORK, self.ToPeerId, "Data recv but not source:", params)
				return
			}

			publicKey := self.ConnectObj.GetJoinPeerPublicKey(dataExtHeader.SourceId)

			signed := payload[len(payload)-256:]
			data := payload[:len(payload)-256]

			if !self.verifyData(&data, &signed, publicKey) {
				return
			}

			originId := self.ConnectObj.GetOriginId(dataExtHeader.SourceId)

			needNoti := self.ConnectObj.CachingMedia(&dataExtHeader, &data, &signed, originId)

			if needNoti {
				self.ConnectObj.NotiData(self.Info.OverlayInfo.OverlayId, dataExtHeader.ChannelId, originId, dataExtHeader.MediaType, &data)
			}

			go self.ConnectObj.BroadcastData(&params, &dataExtHeader, &payload, params.Peer.PeerId, false, true)
		} else {
			util.Println(util.ERROR, self.ToPeerId, "Recv data media-type error:", baseExtHeader.MediaType)

		}
	} else {
		util.Println(util.ERROR, self.ToPeerId, "Recv data ext-header-len error:", params.ExtHeaderLen)
	}
}

func (self *Peer) verifyData(data *[]byte, signed *[]byte, publicKey *rsa.PublicKey) bool {
	if publicKey == nil {
		util.Println(util.ERROR, self.ToPeerId, "PublicKey is nil")
		return false
	}

	if !self.ConnectObj.VerifyData(data, signed, publicKey) {
		util.Println(util.ERROR, self.ToPeerId, "!!!! Failed to VerifyData !!!!")
		return false
	}

	return true
}

func (self *Peer) checkJoinPeer(joinPeer *util.BroadcastDataExtensionHeaderControlJoinJoinerInfo) {
	self.ConnectObj.JoinPeerMux.Lock()
	defer self.ConnectObj.JoinPeerMux.Unlock()

	joinPeers := *self.ConnectObj.GetJoinPeers()
	if oldPeer, ok := joinPeers[joinPeer.PeerId]; !ok {
		joinPeers[joinPeer.PeerId] = &util.JoinPeerInfo{}
		joinPeers[joinPeer.PeerId].PeerId = joinPeer.PeerId
		joinPeers[joinPeer.PeerId].DisplayName = joinPeer.DisplayName
		joinPeers[joinPeer.PeerId].PublicKey = self.ConnectObj.ParsePublicKey(joinPeer.PublicKey)
		if joinPeers[joinPeer.PeerId].PublicKey == nil {
			joinPeers[joinPeer.PeerId].PublicKeyPEM = []byte{}
		} else {
			joinPeers[joinPeer.PeerId].PublicKeyPEM = []byte(joinPeer.PublicKey)
		}

		if self.ConnectObj.PeerChangeCallback != nil {
			util.Println(util.INFO, self.ToPeerId, "PeerChangeCallback - Join !!!!!!!!!!!!!", joinPeers[joinPeer.PeerId])
			self.ConnectObj.PeerChangeCallback(self.Info.OverlayInfo.OverlayId, joinPeer.PeerId, joinPeer.DisplayName, false)
		}
	} else {
		if oldPeer.PublicKeyPEM == nil || len(oldPeer.PublicKeyPEM) <= 0 {
			oldPeer.DisplayName = joinPeer.DisplayName
			oldPeer.PublicKey = self.ConnectObj.ParsePublicKey(joinPeer.PublicKey)
			if oldPeer.PublicKey == nil {
				oldPeer.PublicKeyPEM = []byte{}
			} else {
				oldPeer.PublicKeyPEM = []byte(joinPeer.PublicKey)
			}
		}
	}
}

func (self *Peer) leaveJoinPeer(leavePeer *util.BroadcastDataExtensionHeaderControlLeaveLeaverInfo) {
	self.ConnectObj.JoinPeerMux.Lock()
	defer self.ConnectObj.JoinPeerMux.Unlock()

	joinPeers := *self.ConnectObj.GetJoinPeers()
	if _, ok := joinPeers[leavePeer.PeerId]; ok {
		delete(joinPeers, leavePeer.PeerId)
	}

	if self.ConnectObj.PeerChangeCallback != nil {
		util.Println(util.INFO, self.ToPeerId, "PeerChangeCallback - Leave !!!!!!!!!!!!!", leavePeer)
		self.ConnectObj.PeerChangeCallback(self.Info.OverlayInfo.OverlayId, leavePeer.PeerId, leavePeer.DisplayName, true)
	}
}

func (self *Peer) recvBroadcastDataResponse(res *util.BroadcastDataResponse) {
	util.Println(util.WORK, self.ToPeerId, "Recv broadcast data response:", res)
}

func (self *Peer) sendBroadcastDataResponse() {
	res := util.BroadcastDataResponse{}
	res.RspCode = util.RspCode_BroadcastData

	msg := util.GetPPMessage(res, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send broadcastData response:", msg)
}

func (self *Peer) sendBuffermap(resChan *chan *util.BuffermapResponse) {
	req := util.Buffermap{}
	req.ReqCode = util.ReqCode_Buffermap
	req.ReqParams.OverlayId = self.Info.OverlayInfo.OverlayId

	self.buffermapResChan = resChan

	msg := util.GetPPMessage(req, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send buffermap:", msg)
}

func (self *Peer) recvBuffermap(req *util.Buffermap) {
	util.Println(util.WORK, self.ToPeerId, "recv buffermap:", req)

	if req.ReqParams.OverlayId != self.Info.OverlayInfo.OverlayId {
		util.Println(util.ERROR, self.ToPeerId, "recv buffermap but ovid wrong:", req)
		return
	}

	res := util.BuffermapResponse{}
	res.RspCode = util.RspCode_Buffermap
	res.RspParams.OverlayId = self.Info.OverlayInfo.OverlayId
	res.RspParams.Buffermap = self.ConnectObj.GetBuffermap()

	msg := util.GetPPMessage(res, nil, nil)
	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send buffermap response:", msg)
}

func (self *Peer) recvBuffermapResponse(res *util.BuffermapResponse) {
	defer recover()

	util.Println(util.WORK, self.ToPeerId, "recv buffermap response:", res)

	if res.RspParams.OverlayId != self.Info.OverlayInfo.OverlayId {
		util.Println(util.ERROR, self.ToPeerId, "recv buffermap response but ovid wrong:", res)
		return
	}

	if self.buffermapResChan != nil {
		*self.buffermapResChan <- res
	}
}

func (self *Peer) sendGetData(sourceId string, sequence int, channelId string) {
	req := util.GetDataRequset{}
	req.ReqCode = util.ReqCode_GetData
	req.ReqParams = util.GetDataRequsetParams{}
	req.ReqParams.OverlayId = self.Info.OverlayInfo.OverlayId
	req.ReqParams.SourceId = sourceId
	req.ReqParams.Sequence = sequence

	var msg *util.PPMessage = nil

	if channelId != "" {
		ext := util.GetDataRequsetExtensionHeader{}
		ext.ChannelId = channelId
		//ext.PeerId = sourceId

		msg = util.GetPPMessage(&req, &ext, nil)
	} else {
		msg = util.GetPPMessage(req, nil, nil)
	}

	self.sendPPMessage(msg)

	util.Println(util.WORK, self.ToPeerId, "Send GetData:", msg)
}

func (self *Peer) recvGetData(req *util.GetDataRequset, ext []byte) {
	util.Println(util.WORK, self.ToPeerId, "recv GetData:", req)

	if req.ReqParams.OverlayId != self.Info.OverlayInfo.OverlayId {
		util.Println(util.ERROR, self.ToPeerId, "recv GetData but ovid wrong:", req)
		return
	}

	if ext == nil { // old version
		payload := self.ConnectObj.GetCachingData(req.ReqParams.SourceId, req.ReqParams.Sequence)

		if payload != nil {
			res := util.GetDataResponse{}
			res.RspCode = util.RspCode_GetData
			res.RspParams.Peer.PeerId = payload.Header.ReqParams.Peer.PeerId
			res.RspParams.Sequence = payload.Header.ReqParams.Peer.Sequence
			res.RspParams.Payload = payload.Header.ReqParams.Payload
			res.RspParams.ExtHeaderLen = payload.Header.ReqParams.ExtHeaderLen

			/*extHeader := connect.BroadcastDataExtensionHeader{}
			json.Unmarshal((*payload.Payload)[:payload.Header.ReqParams.ExtHeaderLen], &extHeader)
			*/
			msg := util.GetPPMessage(&res, nil, *payload.Payload)
			self.sendPPMessage(msg)

			util.Println(util.WORK, self.ToPeerId, "Send GetData response:", msg)
		} else {
			util.Println(util.ERROR, self.ToPeerId, "recv GetData but don't have data")
		}
	} else {
		if len(ext) <= 0 {
			util.Println(util.ERROR, self.ToPeerId, "recv GetData but ext header nil")
			return
		}

		extHeader := util.GetDataRequsetExtensionHeader{}
		json.Unmarshal(ext, &extHeader)

		if req.ReqParams.SourceId == "" {
			util.PrintJson(util.ERROR, "recv GetData but peerid wrong:", extHeader)
			return
		}

		if extHeader.ChannelId == "" {
			util.PrintJson(util.ERROR, "recv GetData but channelid wrong:", extHeader)
			return
		}

		media := self.ConnectObj.GetCachingMedia(req.ReqParams.SourceId, extHeader.ChannelId)

		if media != nil {
			res := util.GetDataResponse{}
			res.RspCode = util.RspCode_GetData
			res.RspParams = util.GetDataRspParams{}
			res.RspParams.Peer.PeerId = req.ReqParams.SourceId
			res.RspParams.Sequence = media.Sequence
			res.RspParams.Payload = util.ResponsePayload{}
			res.RspParams.Payload.PayloadType = util.PayloadTypeOctetStream

			ext := util.GetDataResponseExtensionHeader{}
			ext.ChannelId = media.ChannelId
			ext.MediaType = util.MediaTypeDataCache
			//ext.Sequence = media.Sequence
			//ext.PeerId = extHeader.PeerId

			//signBuf := self.ConnectObj.SignData(media.Media)
			payload := make([]byte, 0)
			payload = append(payload, media.Media...)
			payload = append(payload, media.SignMedia...)

			res.RspParams.Payload.Length = len(payload)

			msg := util.GetPPMessage(&res, &ext, payload)
			self.sendPPMessage(msg)

			util.Println(util.WORK, self.ToPeerId, "Send GetData response:", msg)
		} else {
			util.Println(util.ERROR, self.ToPeerId, "recv GetData but don't have data", extHeader)
		}
	}
}

func (self *Peer) recvGetDataResponse(res *util.GetDataResponse, content []byte) {
	util.Println(util.WORK, self.ToPeerId, "recv GetData response:", res)

	params := res.RspParams

	if len(params.Peer.PeerId) <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "recv GetData response but peerid nil")
		return
	}

	if content == nil || len(content) <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "recv GetData response but data nil")
		return
	}

	if params.ExtHeaderLen <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "recv GetData response but ext-header-len nil")
		return
	}

	extbuf := content[:params.ExtHeaderLen]
	payload := content[params.ExtHeaderLen:]

	if extbuf == nil || len(extbuf) <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "recv GetData response but extbuf nil")
		return
	}

	if payload == nil || len(payload) <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "recv GetData response but payload nil")
		return
	}

	extHeader := util.GetDataResponseExtensionHeader{}
	json.Unmarshal(content[:params.ExtHeaderLen], &extHeader)

	util.Println(util.WORK, self.ToPeerId, "Recv data media-type :", extHeader.MediaType)

	if len(extHeader.ChannelId) <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "Recv data channel-id error:", extHeader)
		return
	}

	if params.Sequence <= 0 {
		util.Println(util.ERROR, self.ToPeerId, "Recv data sequence error:", params)
		return
	}

	sign := payload[len(payload)-256:]
	media := payload[:len(payload)-256]

	originId := self.ConnectObj.GetOriginId(params.Peer.PeerId)

	publicKey := self.ConnectObj.GetJoinPeerPublicKey(originId)

	if !self.verifyData(&media, &sign, publicKey) {
		return
	}

	broadHeader := util.BroadcastDataExtensionHeaderDataCache{}
	broadHeader.ChannelId = extHeader.ChannelId
	broadHeader.MediaType = extHeader.MediaType
	broadHeader.Sequence = params.Sequence
	broadHeader.SourceId = params.Peer.PeerId

	needNoti := true

	if extHeader.MediaType == util.MediaTypeDataCache {
		needNoti = self.ConnectObj.CachingMedia(&broadHeader, &media, &sign, originId)
	}

	if needNoti {
		self.ConnectObj.NotiData(self.Info.OverlayInfo.OverlayId, extHeader.ChannelId, originId, extHeader.MediaType, &media)
	}

	broadParams := util.BroadcastDataParams{}
	broadParams.Operation = &util.BroadcastDataParamsOperation{}
	broadParams.Operation.Ack = false
	broadParams.Operation.CandidatePath = false
	broadParams.Peer.PeerId = self.Info.PeerId()
	broadParams.Peer.Sequence = self.ConnectObj.getBroadcastDataSeq()

	go self.ConnectObj.BroadcastData(&broadParams, &broadHeader, &payload, self.Info.PeerId(), false, true)
}
