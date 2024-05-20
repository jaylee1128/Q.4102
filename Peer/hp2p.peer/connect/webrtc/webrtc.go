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
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"hp2p.util/util"

	pwebrtc "github.com/pion/webrtc/v3"
)

type ProbePeer struct {
	Peer           *Peer
	ProbeTimeMilli int64
}

type WebrtcConnect struct {
	connect.Common
	recvChan      chan interface{}
	sendChan      chan interface{}
	broadcastChan chan interface{}
	signalHandler SignalHandler

	position connect.PeerPosition

	peerMap                map[string]*Peer
	blockPeerMap           map[string]string
	peerDataSeqMap         map[string]int
	outGoingCandidateSlice []ProbePeer

	outGoingCandidateSliceMux sync.Mutex

	currentScanTreeCSeq int
	broadcastDataSeq    int
	//dataCacheSeq        int
}

func (self *WebrtcConnect) OverlayInfo() *util.OverlayInfo {
	return &self.Common.OverlayInfo
}

func (self *WebrtcConnect) PeerInfo() *util.PeerInfo {
	return &self.Common.PeerInfo
}

func (self *WebrtcConnect) Init(peerId string, configPath string) {
	self.CommonInit(peerId, configPath)

	self.recvChan = make(chan interface{})
	self.sendChan = make(chan interface{})
	self.broadcastChan = make(chan interface{})

	self.peerMap = make(map[string]*Peer)
	self.blockPeerMap = make(map[string]string)
	self.peerDataSeqMap = make(map[string]int)
	self.outGoingCandidateSlice = make([]ProbePeer, 0)

	//self.Common.PeerAddress = self.ClientConfig.SignalingServerAddr
	self.Common.PeerInfo.Address = self.ClientConfig.SignalingServerAddr

	go self.signalSend()
	go self.signalReceive()
	go self.broadcast()

	self.signalHandler.Start(self.Common.PeerInfo.Address, &self.recvChan)
	self.websocketHello()

	self.broadcastDataSeq = 0
	//self.dataCacheSeq = 0

	self.position = connect.Init
}

func (self *WebrtcConnect) HasConnection() bool {
	return len(*self.allPrimary()) > 0
}

func (self *WebrtcConnect) ConnectionInfo() *util.NetworkResponse {
	info := new(util.NetworkResponse)

	info.Peer.PeerId = self.PeerId()
	info.Peer.TicketId = self.PeerInfo().TicketId

	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.InComingPrimary || peer.Position == connect.OutGoingPrimary {
			info.Primary = append(info.Primary, util.NetworkPeerInfo{PeerId: peer.ToPeerId, TicketId: peer.ToTicketId})
		} else if peer.Position == connect.InComingCandidate {
			info.InComingCandidate = append(info.InComingCandidate, util.NetworkPeerInfo{PeerId: peer.ToPeerId, TicketId: peer.ToTicketId})
		} else if peer.Position == connect.OutGoingCandidate {
			info.OutGoingCandidate = append(info.OutGoingCandidate, util.NetworkPeerInfo{PeerId: peer.ToPeerId, TicketId: peer.ToTicketId})
		}
	}
	self.PeerMapMux.Unlock()

	return info
}

func (self *WebrtcConnect) OnTrack(toPeerId string, kind string, track *pwebrtc.TrackLocalStaticRTP) {
	if track != nil {
		util.Println(util.INFO, "OnTrack!!!!!!!!!!!!!!!!!! webrtc.go", kind)
	} else {
		util.Println(util.INFO, "Release Track!!!!!!!!!!!!!!!!!!", kind)
	}

	if kind == "video" {
		if self.VideoTrack == nil {
			self.VideoTrack = track
		} else {
			self.ChangeVideoTrack = true
			util.Println(util.INFO, toPeerId, "Change Video Track!!!!!!!!")
		}
	} else {
		if self.AudioTrack == nil {
			self.AudioTrack = track
		} else {
			self.ChangeAudioTrack = true
			util.Println(util.INFO, toPeerId, "Change Audio Track!!!!!!!!")
		}
	}

	if self.VideoTrack != nil && self.AudioTrack != nil && !(self.ChangeAudioTrack || self.ChangeVideoTrack) {
		peerlist := self.allPrimary()

		for _, peer := range *peerlist {

			if peer.ToPeerId == toPeerId {
				continue
			}

			peer.peerConnection.AddTrack(self.VideoTrack)
			peer.peerConnection.AddTrack(self.AudioTrack)
			peer.CreateOffer()
		}
	}

	if self.ChangeAudioTrack && self.ChangeVideoTrack {
		self.ChangeVideoTrack = false
		self.ChangeAudioTrack = false

		peerlist := self.allPrimary()

		var OGPrimaryPeer *Peer = nil
		var successReOffer bool = false

		util.Println(util.INFO, "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		for _, peer := range *peerlist {
			util.Println(util.INFO, peer.ToPeerId, "search peer!!!!!!!!")

			if peer.ToPeerId == toPeerId {
				continue
			}

			if peer.Position == connect.OutGoingPrimary {
				OGPrimaryPeer = peer
			}

			if peer.MediaReceive {
				util.Println(util.INFO, peer.ToPeerId, "Reoffer!!!!!!!!")
				peer.MediaReceive = false

				peer.peerConnection.AddTrack(self.VideoTrack)
				peer.peerConnection.AddTrack(self.AudioTrack)
				peer.CreateOffer()
				successReOffer = true
			}
		}

		if !successReOffer && OGPrimaryPeer != nil {
			util.Println(util.INFO, OGPrimaryPeer.ToPeerId, "Reoffer OGP!!!!!!!!")
			OGPrimaryPeer.MediaReceive = false

			OGPrimaryPeer.peerConnection.AddTrack(self.VideoTrack)
			OGPrimaryPeer.peerConnection.AddTrack(self.AudioTrack)
			OGPrimaryPeer.CreateOffer()
		}
	}
}

func (self *WebrtcConnect) broadcastHello(hello *util.HelloPeer) {
	util.Println(util.INFO, hello.ReqParams.Peer.PeerId, "broadcasthello:")

	hello.ReqParams.Operation.Ttl -= 1

	estab := false

	//util.Println(util.INFO, hello.ReqParams.Peer.PeerId, "NumPrimary:", self.PeerStatus.NumPrimary, ", MaxPrimaryConnection:", self.PeerConfig.MaxPrimaryConnection)
	//util.Println(util.INFO, hello.ReqParams.Peer.PeerId, "NumOutCandidate:", self.PeerStatus.NumOutCandidate, ", MaxOutgoingCandidate:", self.PeerConfig.MaxOutgoingCandidate)

	/*
		if hello.ReqParams.Operation.VerticalCandidate != nil && *hello.ReqParams.Operation.VerticalCandidate {
			estab = true
		}*/

	if self.PeerStatus.NumPrimary < self.PeerConfig.MaxPrimaryConnection ||
		self.PeerStatus.NumInCandidate < self.PeerConfig.MaxIncomingCandidate {
		hello.ReqParams.Operation.ConnNum -= 1
		self.PeerStatus.NumInCandidate++ // count up before send estab
		estab = true
		util.Println(util.INFO, hello.ReqParams.Peer.PeerId, "estab true, ConnNum:", hello.ReqParams.Operation.ConnNum)
	}

	if hello.ReqParams.Operation.Ttl > 0 {
		peerlist := self.inComingPrimary()

		if len(*peerlist) > 0 {
			util.Println(util.INFO, "before connNum:", hello.ReqParams.Operation.ConnNum)
			connNum := float64(hello.ReqParams.Operation.ConnNum) / float64(len(*peerlist))
			connNum = math.Ceil(connNum)
			hello.ReqParams.Operation.ConnNum = int(connNum)
			util.Println(util.INFO, "after connNum:", hello.ReqParams.Operation.ConnNum)

			for _, peer := range *peerlist {
				if peer.ToPeerId == hello.ReqParams.Peer.PeerId {
					continue
				}

				peer.RelayHello(hello)
			}
		}
	}

	if estab {
		//<-time.After(time.Second * 3)
		util.Println(util.INFO, "estab:")
		conPeer, ok := self.ConnectTo(hello.ReqParams.Peer.PeerId, connect.Established)

		if conPeer.releasePeer {
			return
		}

		if ok {

			if conPeer.Position != connect.Established {
				return
			}

			self.CommunicationMux.Lock()
			if conPeer.SendEstab(hello.ReqParams.Operation.VerticalCandidate == nil || !*hello.ReqParams.Operation.VerticalCandidate) {
				self.PeerStatus.NumInCandidate-- // double count in addConnectionInfo
				conPeer.AddConnectionInfo(connect.InComingCandidate)
				conPeer.CheckHeartBeatRecved()
				conPeer.ToTicketId = hello.ReqParams.Peer.TicketId
			} else {
				self.DisconnectFrom(conPeer)
				self.PeerStatus.NumInCandidate--
			}
			self.CommunicationMux.Unlock()
		}
	}
}

func (self *WebrtcConnect) broadcast() {
	for {
		msg := <-self.broadcastChan

		util.Println(util.INFO, "broadcast:", msg)

		switch msg.(type) {
		case *util.HelloPeer:
			hello := msg.(*util.HelloPeer)

			self.broadcastHello(hello)
		}
	}
}

func (self *WebrtcConnect) websocketHello() {
	hello := struct {
		Action string `json:"action"`
		PeerId string `json:"peer-id"`
	}{
		"hello",
		self.PeerId(),
	}

	self.sendChan <- hello
}

func (self *WebrtcConnect) websocketBye() {
	hello := struct {
		Action string `json:"action"`
		PeerId string `json:"peer-id"`
	}{
		"bye",
		self.PeerId(),
	}

	self.sendChan <- hello
}

func (self *WebrtcConnect) inComingPrimary() *[]*Peer {
	plist := new([]*Peer)
	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.InComingPrimary {
			*plist = append(*plist, peer)
		}
	}
	self.PeerMapMux.Unlock()
	return plist
}

func (self *WebrtcConnect) outGoingPrimary() *Peer {
	defer self.PeerMapMux.Unlock()

	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.OutGoingPrimary {
			return peer
		}
	}

	return nil
}

func (self *WebrtcConnect) allPrimary() *[]*Peer {
	plist := new([]*Peer)
	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.OutGoingPrimary || peer.Position == connect.InComingPrimary {
			*plist = append(*plist, peer)
		}
	}
	self.PeerMapMux.Unlock()
	return plist
}

func (self *WebrtcConnect) outGoingCandidate() *[]*Peer {
	plist := new([]*Peer)
	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.OutGoingCandidate {
			*plist = append(*plist, peer)
		}
	}
	self.PeerMapMux.Unlock()
	return plist
}

func (self *WebrtcConnect) allPeer() *[]*Peer {
	plist := new([]*Peer)
	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		*plist = append(*plist, peer)
	}
	self.PeerMapMux.Unlock()
	return plist
}

func (self *WebrtcConnect) lowestOutGoingCandidate() *Peer {
	var lowestPeer *Peer = nil

	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.OutGoingCandidate {
			if lowestPeer == nil || peer.ToTicketId > lowestPeer.ToTicketId {
				lowestPeer = peer
			}
		}
	}
	self.PeerMapMux.Unlock()

	return lowestPeer
}

func (self *WebrtcConnect) signalSend() {
	for {
		select {
		case send := <-self.sendChan:
			var obj interface{}

			switch send.(type) {
			case util.RTCSessionDescription:
				sdp := send.(util.RTCSessionDescription)
				sdp.Fromid = self.PeerId()
				obj = sdp
				util.Println(util.WORK, "send", sdp.Type, "to", sdp.Toid)
			case util.RTCIceCandidate:
				ice := send.(util.RTCIceCandidate)
				ice.Fromid = self.PeerId()
				obj = ice
			default:
				obj = send
			}

			buf, err := json.Marshal(obj)
			if err != nil {
				util.Println(util.ERROR, "signal json marshal:", err)
			} else {
				self.signalHandler.Send(buf)
			}
		}
	}
}

func (self *WebrtcConnect) signalReceive() {
	for {
		select {
		case recv := <-self.recvChan:
			switch recv.(type) {
			case util.RTCSessionDescription:
				sdp := recv.(util.RTCSessionDescription)
				//fmt.Println(sdp)

				if sdp.Toid == self.PeerId() {

					self.PeerMapMux.Lock()
					peer, ok := self.peerMap[sdp.Fromid]

					if ok {
						util.Println(util.WORK, "receive", sdp.Type, "from", sdp.Fromid, "(connected)")
						if sdp.Type == "answer" {
							peer.SetSdp(sdp)
						} else if sdp.Type == "offer" {
							if peer.Position != connect.InComingPrimary && peer.Position != connect.OutGoingPrimary {

								if peer.Position == connect.Established {

									util.Println(util.ERROR, "!!!!!!!!!!!!! cancel estab !!!!!!!!!!!!")
									self.PeerMapMux.Unlock()
									self.DisconnectFrom(peer)
									self.PeerMapMux.Lock()

									self.newPeerReceiveOffer(sdp)

								} else {
									util.Println(util.WORK, sdp.Fromid, "Already connect. ignore. position:", peer.Position)
								}
							} else {
								peer.ReceiveOffer(sdp)
							}
						}
					} else {
						if sdp.Type == "offer" {
							util.Printf(util.WORK, "receive offer from %s", sdp.Fromid)

							self.newPeerReceiveOffer(sdp)
						}
					}

					self.PeerMapMux.Unlock()
				}

			case util.RTCIceCandidate:
				ice := recv.(util.RTCIceCandidate)
				//fmt.Println(ice)

				if ice.Toid == self.PeerId() {
					peer, ok := self.peerMap[ice.Fromid]

					if ok {
						util.Printf(util.WORK, "receive iceCandidate from %s", ice.Fromid)
						peer.AddIceCandidate(ice)
					}
				}
			}
		}
	}
}

func (self *WebrtcConnect) newPeerReceiveOffer(sdp util.RTCSessionDescription) {
	connectChan := make(chan bool, 1)

	peer := NewPeer(sdp.Fromid, connect.InComing, &connectChan, self)
	self.peerMap[sdp.Fromid] = peer

	peer.ReceiveOffer(sdp)

	go func() {
		conn, ok := <-connectChan

		if !conn && ok {
			self.DisconnectFrom(peer)
		}
	}()
}

func (self *WebrtcConnect) probePeers() {
	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.Established {
			peer.SendProbe()
		} /* else { recovery 일때는 estab 아닌애들도 있다...
			self.DisconnectFrom(peer)
		}*/
	}
	self.PeerMapMux.Unlock()

	<-time.After(time.Second * time.Duration(self.PeerConfig.ProbePeerTimeout))

	self.outGoingCandidateSliceMux.Lock()
	self.outGoingCandidateSlice = make([]ProbePeer, 0)

	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {
		if peer.Position == connect.Established {
			if peer.probeTime != nil {
				pbPeer := ProbePeer{}
				pbPeer.Peer = peer
				pbPeer.ProbeTimeMilli = *peer.probeTime
				self.outGoingCandidateSlice = append(self.outGoingCandidateSlice, pbPeer)
			}
		}
	}
	self.PeerMapMux.Unlock()

	sort.Slice(self.outGoingCandidateSlice, func(i, j int) bool {
		return self.outGoingCandidateSlice[i].ProbeTimeMilli < self.outGoingCandidateSlice[j].ProbeTimeMilli
	})

	for _, pbPeer := range self.outGoingCandidateSlice {
		pbPeer.Peer.TryPrimary()
	}

	self.outGoingCandidateSliceMux.Unlock()
}

func (self *WebrtcConnect) Join(recovery bool, accessKey string) *util.HybridOverlayJoinResponse {

	if accessKey != "" {
		self.OverlayInfo().Auth.AccessKey = &accessKey
	}

	ovinfo := self.OverlayJoin(recovery)

	if ovinfo == nil {
		util.Println(util.ERROR, "Failed to join overlay.")
		return nil
	}

	return ovinfo

}

func (self *WebrtcConnect) ConnectToPeers(peers *[]util.PeerInfo, recovery bool) bool {
	ovPeerList := *peers
	for _, peer := range ovPeerList {

		if peer.PeerId == self.PeerId() {
			continue
		}

		self.position = connect.SendHello

		conPeer, conn := self.ConnectTo(peer.PeerId, connect.SendHello)

		if !conn {
			if conPeer != nil && conPeer.Position == connect.OutGoingCandidate {
				if conPeer.setPrimary() {
					util.Println(util.INFO, "Success primary with", conPeer.ToPeerId)
					return true
				}
			}
			util.Println(util.INFO, "Failed primary with", conPeer.ToPeerId)
			continue
		}

		if !conPeer.sendHello() {
			util.Println(util.ERROR, conPeer.ToPeerId, "Failed to send hello")
			self.DisconnectFrom(conPeer)
			continue
		}

		self.DisconnectFrom(conPeer)
		break
	}

	<-time.After(time.Second * time.Duration(self.PeerConfig.PeerEstabTimeout))

	if self.PeerConfig.ProbePeerTimeout > 0 {
		if util.LEVEL >= util.INFO {
			for pid, peer := range self.peerMap {
				util.Println(util.INFO, pid, " positon: ", peer.Position)
			}
		}

		self.probePeers()

		//TODO probe after retry next
	}

	if self.HaveOutGoingPrimary {
		if !self.LeaveOverlay {
			util.Println(util.INFO, "Success to make primary")
		}
		return true
	} else {
		if self.LeaveOverlay {
			return false
		}
		if recovery {
			util.Println(util.INFO, "Recovery after", self.PeerConfig.RetryOverlayRecoveryInterval, "sec.")
			<-time.After(time.Second * time.Duration(self.PeerConfig.RetryOverlayRecoveryInterval))
		} else if self.PeerConfig.RetryOverlayJoin {
			util.Println(util.INFO, "Retry after", self.PeerConfig.RetryOverlayJoinInterval, "sec.")
			<-time.After(time.Second * time.Duration(self.PeerConfig.RetryOverlayJoinInterval))
		}
	}

	return false
}

func (self *WebrtcConnect) JoinAndConnect(recovery bool, accessKey string) {

	if self.GetPublicKeyString() == "" {
		util.Println(util.ERROR, "Failed to get public key !!!")
		return
	}

	var retry int = 1

	if self.PeerConfig.RetryOverlayJoin {
		retry = self.PeerConfig.RetryOverlayJoinCount

		if retry <= 0 {
			retry = 1
		}
	}

	var cnt int = 0

	for recovery || cnt < retry {

		if self.LeaveOverlay {
			return
		}

		cnt++

		util.Println(util.INFO, "Overlay join count", cnt)

		ovinfo := self.Join(recovery || cnt > 1, accessKey)

		if ovinfo == nil {
			return
		}

		if ovinfo.RspCode == 407 {
			util.Println(util.ERROR, "Access key is wrong or dosen't have authority.")
			return
		}

		if ovinfo.RspCode != 200 && ovinfo.RspCode != 202 {
			util.Println(util.ERROR, "Failed to join overlay. RspCode:", ovinfo.RspCode)
			return
		}

		ovPeerList := ovinfo.Overlay.Status.PeerInfoList

		if ovPeerList == nil || len(ovPeerList) <= 0 {
			util.Println(util.ERROR, "Overlay peer list is empty.")
			return
		}

		util.Println(util.INFO, "Overlay Peer list: ", ovPeerList)

		if self.ConnectToPeers(&ovPeerList, recovery) {
			return
		}
	}
}

func (self *WebrtcConnect) ConnectAfterJoin(peerList *[]util.PeerInfo, accessKey string) {

	var retry int = 1

	if self.PeerConfig.RetryOverlayJoin {
		retry = self.PeerConfig.RetryOverlayJoinCount

		if retry <= 0 {
			retry = 1
		}
	}

	var cnt int = 0

	povPeerList := peerList

	for cnt < retry {

		cnt++

		util.Println(util.INFO, "Overlay join count", cnt)

		util.Println(util.INFO, "Overlay Peer list: ", *povPeerList)

		if self.ConnectToPeers(povPeerList, false) {
			return
		}

		ovinfo := self.Join(cnt > 0, accessKey)
		povPeerList = &ovinfo.Overlay.Status.PeerInfoList

		if povPeerList == nil {
			return
		}
	}
}

func (self *WebrtcConnect) ConnectPeersXXX(recovery bool) {

	/*
		var retry int = 1

		if self.PeerConfig.RetryOverlayJoin {
			retry = self.PeerConfig.RetryOverlayJoinCount

			if retry <= 0 {
				retry = 1
			}
		}

		var cnt int = 0

		if accessKey != "" {
			self.OverlayInfo().Auth.AccessKey = &accessKey
		}

		for recovery || cnt < retry {

			cnt++

			util.Println(util.INFO, "Overlay join count", cnt)

			ovinfo := self.OverlayJoin(recovery || cnt > 1)

			if ovinfo == nil {
				util.Println(util.ERROR, "Failed to join overlay.")
				return
			}

			ovPeerList := ovinfo.Status.PeerInfoList

			if ovPeerList == nil || len(ovPeerList) <= 0 {
				util.Println(util.ERROR, "Overlay peer list is empty.")
				return
			}

			util.Println(util.INFO, "Overlay Peer list: ", ovPeerList)

			for _, peer := range ovPeerList {

				if peer.PeerId == self.PeerId() {
					continue
				}

				self.position = connect.SendHello

				conPeer, conn := self.ConnectTo(peer.PeerId, connect.SendHello)

				if !conn {
					if conPeer != nil && conPeer.Position == connect.OutGoingCandidate {
						if conPeer.setPrimary() {
							util.Println(util.INFO, "Success primary with", conPeer.ToPeerId)
							break
						}
					}
					util.Println(util.INFO, "Failed primary with", conPeer.ToPeerId)
					continue
				}

				if !conPeer.sendHello() {
					util.Println(util.ERROR, conPeer.ToPeerId, "Failed to send hello")
					self.DisconnectFrom(conPeer)
					continue
				}

				self.DisconnectFrom(conPeer)
				break
			}

			<-time.After(time.Second * time.Duration(self.PeerConfig.PeerEstabTimeout))

			if self.PeerConfig.ProbePeerTimeout > 0 {
				if util.LEVEL >= util.INFO {
					for pid, peer := range self.peerMap {
						util.Println(util.INFO, pid, " positon: ", peer.Position)
					}
				}

				self.probePeers()

				//TODO probe after retry next
			}

			if self.HaveOutGoingPrimary {
				util.Println(util.INFO, "Success to make primary")
				break
			} else {
				if recovery {
					util.Println(util.INFO, "Recovery after", self.PeerConfig.RetryOverlayRecoveryInterval, "sec.")
					<-time.After(time.Second * time.Duration(self.PeerConfig.RetryOverlayRecoveryInterval))
				} else if self.PeerConfig.RetryOverlayJoin {
					util.Println(util.INFO, "Retry after", self.PeerConfig.RetryOverlayJoinInterval, "sec.")
					<-time.After(time.Second * time.Duration(self.PeerConfig.RetryOverlayJoinInterval))
				}
			}
		}*/
}

func (self *WebrtcConnect) sendVerticalHello(toPeerId string) {
	<-time.After(time.Second * 4)

	util.Println(util.WORK, "start sendVerticalHello to", toPeerId)

	peer, ok := self.ConnectTo(toPeerId, connect.SendHello)

	if ok {
		if peer.sendVerticalHello() {
			util.Println(util.WORK, "Success to send vertical-hello")
		} else {
			util.Println(util.WORK, "Failed to send vertical-hello")
		}

		self.DisconnectFrom(peer)
	} else if peer != nil {
		util.Println(util.INFO, "connect position:", peer.Position)

		if peer.Position != connect.OutGoingCandidate && peer.Position != connect.OutGoingPrimary {
			util.Println(util.WORK, "retry after 5sec - sendVerticalHello to", toPeerId)
			<-time.After(time.Second * 5)
			self.sendVerticalHello(toPeerId)
		} else if peer.Position == connect.OutGoingCandidate {
			peer.isVerticalCandidate = true
		}
	}
}

func (self *WebrtcConnect) ConnectTo(toPeerId string, positon connect.PeerPosition) (*Peer, bool) {
	util.Println(util.WORK, "Try connect to", toPeerId)
	self.PeerMapMux.Lock()

	old, ok := self.peerMap[toPeerId]

	var rslt bool = false
	var peer *Peer = nil

	if ok {
		util.Println(util.WORK, "Already connect to", toPeerId, "position:", old.Position)
		peer = old
		self.PeerMapMux.Unlock()
	} else {
		connectChan := make(chan bool)
		/*peer = NewPeer(toPeerId, &self.Common, positon,
		&self.sendChan, &connectChan, &self.broadcastChan, self.DisconnectFrom, self.OnTrack,
		self.RecvScanTree, self.RecvScanTreeResponse, self.recvChat, self.BroadcastData, self.GetBuffermap)*/

		peer = NewPeer(toPeerId, positon, &connectChan, self)

		self.peerMap[toPeerId] = peer
		self.PeerMapMux.Unlock()

		peer.CreateOffer()

		rslt = <-connectChan
	}

	return peer, rslt
}

func (self *WebrtcConnect) DisconnectFrom(peer *Peer) {
	if peer == nil {
		return
	}

	self.PeerMapMux.Lock()

	util.Println(util.WORK, "Disconnect from", peer.ToPeerId)

	delete(self.peerMap, peer.ToPeerId)

	self.JoinPeerMux.Lock()
	joinPeers := *self.GetJoinPeers()
	if _, ok := joinPeers[peer.ToPeerId]; ok {
		delete(joinPeers, peer.ToPeerId)
	}
	self.JoinPeerMux.Unlock()

	/*self.CachingBufferMapMutex.Lock()
	delete(self.CachingBufferMap, peer.ToPeerId)
	self.CachingBufferMapMutex.Unlock()*/

	peer.Close()

	self.PeerMapMux.Unlock()

	if self.LeaveOverlay {
		return
	}

	/*
		self.PeerChangeCallback(self.OverlayInfo().OverlayId, peer.ToPeerId, *peer.Info.PeerInfo.DisplayName, true)

		if _, ok := self.PeerJoinNotiMap[peer.ToPeerId]; ok {
			delete(self.PeerJoinNotiMap, peer.ToPeerId)
		}*/

	if peer.Position == connect.OutGoingPrimary {
		util.Println(util.WORK, "Disconnected from OutGoingPrimary. Try recovery.")
		go self.Recovery()
	}
}

func (self *WebrtcConnect) Recovery() {

	<-time.After(time.Second * 1)

	util.Println(util.WORK, "Start recovery.")

	plist := self.outGoingCandidate()

	if len(*plist) > 0 {
		for _, peer := range *plist {
			if peer.isVerticalCandidate {
				if peer.setPrimary() {
					return
				}
				break
			}
		}

		for _, peer := range *plist {
			if peer.isVerticalCandidate {
				continue
			}

			if peer.setPrimary() {
				break
			}
		}

		if !self.HaveOutGoingPrimary {
			self.JoinAndConnect(true, *self.Common.OverlayInfo.Auth.AccessKey)
		} /*else {
			if self.OverlayInfo().CrPolicy.RecoveryBy == connect.RecoveryByPull {
				go self.recoveryDataForPull()
			}
		}*/
	} else {
		self.JoinAndConnect(true, *self.Common.OverlayInfo.Auth.AccessKey)
	}
}

/*
func (self *WebrtcConnect) DisconnectById(toPeerId string) {
	peer := self.peerMap[toPeerId]

	self.DisconnectFrom(peer)
}
*/

func (self *WebrtcConnect) SendScanTree() int {
	self.currentScanTreeCSeq = time.Now().Nanosecond()
	for _, val := range *self.allPrimary() {
		val.sendScanTree(self.currentScanTreeCSeq)
	}

	return self.currentScanTreeCSeq
}

func (self *WebrtcConnect) RecvScanTree(req *util.ScanTree, peer *Peer) {

	if req.ReqParams.CSeq == self.currentScanTreeCSeq {
		return
	}

	peers := self.allPrimary()

	res := util.ScanTreeResponse{}
	res.RspParams = req.ReqParams

	if len(*peers) > 1 {
		res.RspCode = util.RspCode_ScanTreeNonLeaf
		peer.sendScanTreeResponse(&res)

		via := append([][]string{{self.PeerId(), self.Common.PeerInfo.Address}}, req.ReqParams.Overlay.Via...)
		req.ReqParams.Overlay.Via = via
		//req.ReqParams.Overlay.Path = path

		//peers = self.allPrimary()
		for _, primary := range *peers {
			if peer.ToPeerId != primary.ToPeerId {
				primary.broadcastScanTree(req)
			}
		}
	} else {
		res.RspCode = util.RspCode_ScanTreeLeaf
		path := append([][]string{{self.PeerId(), strconv.Itoa(self.PeerInfo().TicketId), self.Common.PeerInfo.Address}}, req.ReqParams.Overlay.Path...)
		res.RspParams.Overlay.Path = path
		peer.sendScanTreeResponse(&res)
	}
}

func (self *WebrtcConnect) RecvScanTreeResponse(res *util.ScanTreeResponse) {
	via := res.RspParams.Overlay.Via[0]

	if via[0] != self.PeerId() {
		util.Println(util.ERROR, "ScanTree response error: via[0] != me")
		return
	}

	if len(res.RspParams.Overlay.Via) == 1 {
		/*if self.ReportScanTreeCallback != nil {
			self.ReportScanTreeCallback(&res.RspParams.Overlay.Path, self.currentScanTreeCSeq)
		}*/

	} else {
		res.RspParams.Overlay.Via = res.RspParams.Overlay.Via[1:]

		util.Println(util.INFO, "ScanTree response relay:", res.RspParams.Overlay.Via)

		path := append([][]string{{via[0], strconv.Itoa(self.PeerInfo().TicketId), via[1]}}, res.RspParams.Overlay.Path...)
		res.RspParams.Overlay.Path = path

		for _, peer := range self.peerMap {
			util.Println(util.INFO, "~~~~~~~~~~~ topeer:", peer.ToPeerId)
			//if strings.Compare(peer.ToPeerId, res.RspParams.Overlay.Via[0][0]) == 0 {
			if peer.ToPeerId == res.RspParams.Overlay.Via[0][0] {
				util.Println(util.INFO, "~!!!!!!!!!!!!! send scan rel peer:", peer.ToPeerId)
				peer.sendScanTreeResponse(res)
				break
			}
		}
	}
}

func (self *WebrtcConnect) getBroadcastDataSeq() int {
	self.broadcastDataSeq += 1
	return self.broadcastDataSeq
}

// func (self *WebrtcConnect) getDataCacheSeq() int {
// 	self.dataCacheSeq += 1
// 	return self.dataCacheSeq
// }

func (self *WebrtcConnect) checkBroadcastDataSeq(peerId string, recvSeq int) bool {
	seq, ok := self.peerDataSeqMap[peerId]

	self.peerDataSeqMap[peerId] = recvSeq

	if ok {
		if recvSeq > seq {
			return true
		} else {
			self.peerDataSeqMap[peerId] = seq
			return false
		}
	} else {
		return true
	}
}

/*
func (self *WebrtcConnect) SendData(data *[]byte, candidatePath bool) {

	req := util.BroadcastData{}
	req.ReqCode = util.ReqCode_BroadcastData
	req.ReqParams.Operation = &util.BroadcastDataParamsOperation{}
	req.ReqParams.Operation.Ack = self.PeerConfig.BroadcastOperationAck
	req.ReqParams.Operation.CandidatePath = candidatePath
	req.ReqParams.Peer.PeerId = self.PeerId()
	req.ReqParams.Peer.Sequence = self.getBroadcastDataSeq()
	req.ReqParams.Payload.Length = len(*data)

	databyte := *data

	extHeader := util.BroadcastDataExtensionHeader{}
	extHeader.AppId = appId
	buf, err := json.Marshal(extHeader)
	if err != nil {
		util.Println(util.ERROR, "extHeader marshal error: ", err)
	} else {
		req.ReqParams.ExtHeaderLen = len(buf)
		databyte = append(buf, databyte...)
	}

	if appId == consts.AppIdMedia {
		req.ReqParams.Payload.PayloadType = consts.PayloadTypeOctetStream
	} else {
		req.ReqParams.Payload.PayloadType = consts.PayloadTypeText
	}

	self.BroadcastData(&req.ReqParams, nil, &databyte, nil, true, true)
}*/

func (self *WebrtcConnect) BroadcastCachingData(sourcePeerId *string, sequence int) {
	self.CachingBufferMapMutex.Lock()

	for _, buf := range self.CachingBufferMap {

		if *sourcePeerId != buf.SourcePeerId {
			continue
		}

		buf.BufferMutax.Lock()

		for _, dp := range buf.DataPackets {
			if sequence == dp.Sequence {
				self.BroadcastData(&dp.Payload.Header.ReqParams, nil, dp.Payload.Payload, buf.SourcePeerId, false, true)
				break
			}
		}

		buf.BufferMutax.Unlock()
	}

	self.CachingBufferMapMutex.Unlock()
}

/*
	func (self *WebrtcConnect) IoTData(data *util.IoTData) {
		util.Println(util.INFO, "IOT DATA:", data)

		self.UDPConnection = true
		self.AddConnectedAppIds(data.AppId)

		appdata, err := json.Marshal(data.AppData)
		if err != nil {
			util.Println(util.ERROR, "IOT Data to string error:", err)
		} else {
			self.SendData(&appdata, data.AppId, false)
			self.RecvIoTCallback(string(appdata))
		}
	}

	func (self *WebrtcConnect) BlockChainData(data *util.BlockChainData) {
		util.Println(util.INFO, "BlockChain DATA:", data)

		self.UDPConnection = true
		self.AddConnectedAppIds(data.AppId)

		appdata, err := json.Marshal(data.AppData)
		if err != nil {
			util.Println(util.ERROR, "IOT Data to string error:", err)
		} else {
			self.SendData(&appdata, data.AppId, false)
			//self.RecvBlockChainCallback(string(appdata))
		}
	}

	func (self *WebrtcConnect) MediaData(data *util.MediaAppData) {
		util.Println(util.INFO, "Media DATA:", data)

		self.UDPConnection = true
		self.AddConnectedAppIds(data.AppId)

		self.SendData(data.AppData, data.AppId, false)
	}
*/
func (self *WebrtcConnect) CheckSource(senderId string, channelId string) bool {
	senderOriginId := self.GetOriginId(senderId)

	for _, channel := range self.ServiceInfo().ChannelList {
		if channel.ChannelId == channelId {
			if channel.SourceList == nil {
				if self.ServiceInfo().SourceList != nil {
					for _, source := range *self.ServiceInfo().SourceList {
						if source == "*" || source == senderOriginId {
							return true
						}
					}
				}

				return false
			} else {
				for _, source := range *channel.SourceList {
					if source == "*" || source == senderOriginId {
						return true
					}
				}

				return false
			}
		}
	}

	return false
}

func (self *WebrtcConnect) SendData(channelId string, dataType string, data []byte) int32 {
	if !self.CheckSource(self.PeerOriginId(), channelId) {
		return util.RspCode_AuthFailed
	}

	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	payload := make([]byte, 0)
	payload = append(payload, data...)
	sign := self.SignData(data)
	payload = append(payload, sign...)

	var extHeader interface{} = nil
	if dataType == util.MediaTypeDataCache {
		ext := util.BroadcastDataExtensionHeaderDataCache{}
		ext.ChannelId = channelId
		ext.MediaType = dataType
		ext.Sequence = params.Peer.Sequence
		ext.SourceId = self.PeerId()

		self.CachingMedia(&ext, &data, &sign, self.PeerId())

		extHeader = ext
	} else if dataType == util.MediaTypeData {
		ext := util.BroadcastDataExtensionHeader{}
		ext.ChannelId = channelId
		ext.MediaType = dataType

		extHeader = ext
	} else {
		return util.RspCode_WrongRequest
	}

	self.BroadcastData(&params, extHeader, &payload, self.PeerId(), false, true)

	return util.RspCode_Success
}

func (self *WebrtcConnect) CachingMedia(header *util.BroadcastDataExtensionHeaderDataCache, media *[]byte, signMedia *[]byte, peerId string) bool {
	self.JoinPeersLock()
	defer self.JoinPeersUnlock()

	if header == nil {
		util.Println(util.ERROR, "CachingMedia header is nil")
		return false
	}

	if media == nil || len(*media) <= 0 {
		util.Println(util.ERROR, "CachingMedia media is nil")
		return false
	}

	newCache := false

	joinPeers := self.GetJoinPeers()

	for _, peer := range *joinPeers {
		if peer.PeerId == peerId {
			var find bool = false
			for _, cachingMedia := range peer.CachingMedia {
				if cachingMedia.ChannelId == header.ChannelId {
					if cachingMedia.Sequence < header.Sequence || cachingMedia.SignMedia == nil || len(cachingMedia.SignMedia) <= 0 {
						cachingMedia.Sequence = header.Sequence
						cachingMedia.Media = *media
						cachingMedia.SignMedia = *signMedia
						newCache = true
					}
					find = true
					break
				}
			}
			if !find {
				cachingMedia := util.JoinPeerInfoCachingMedia{}
				cachingMedia.ChannelId = header.ChannelId
				cachingMedia.Sequence = header.Sequence
				cachingMedia.Media = *media
				cachingMedia.SignMedia = *signMedia
				peer.CachingMedia = append(peer.CachingMedia, cachingMedia)
				newCache = true
			}
			break
		}
	}

	return newCache
}

func (self *WebrtcConnect) GetCachingMedia(peerId string, channelId string) *util.JoinPeerInfoCachingMedia {
	self.JoinPeersLock()
	defer self.JoinPeersUnlock()

	//peerOriginId := self.GetOriginId(peerId)

	joinPeers := self.GetJoinPeers()

	for _, peer := range *joinPeers {
		if peer.PeerId == peerId {
			for _, cachingMedia := range peer.CachingMedia {
				if cachingMedia.ChannelId == channelId {
					return &cachingMedia
				}
			}
		}
	}

	return nil
}

func (self *WebrtcConnect) BroadcastData(params *util.BroadcastDataParams, extHeader interface{}, data *[]byte, senderId string, caching bool, includeOGP bool) {

	req := new(util.BroadcastData)
	req.ReqCode = util.ReqCode_BroadcastData
	req.ReqParams = *params

	if caching {
		self.dataCaching(req, data, &params.Peer.PeerId)
	}

	msg := util.GetPPMessage(req, extHeader, *data)

	var peerlist *[]*Peer = nil

	if params.Operation.CandidatePath {
		peerlist = self.allPeer()
	} else {
		peerlist = self.allPrimary()
	}

	if peerlist == nil || len(*peerlist) <= 0 {
		util.Println(util.WORK, "No peer to send broadcast data.")
		return
	}

	for _, peer := range *peerlist {
		if senderId == peer.ToPeerId {
			continue
		}

		if req.ReqParams.Peer.PeerId == peer.ToPeerId {
			continue
		}

		if peer.Position >= connect.InComingCandidate && peer.Position <= connect.OutGoingPrimary {
			if peer.Position == connect.OutGoingPrimary && !includeOGP {
				continue
			}

			if peer.Position == connect.InComingCandidate || peer.Position == connect.OutGoingCandidate {
				if !params.Operation.CandidatePath {
					continue
				}
			}
		} else {
			continue
		}

		util.Println(util.WORK, peer.ToPeerId, "send broadcastdata:", req)
		peer.sendPPMessage(msg)
		util.Println(util.WORK, peer.ToPeerId, "sended broadcastdata:", req)
	}
}

func (self *WebrtcConnect) checkCachingBuffer(buf *util.CachingBuffer) {
	now := time.Now()

	buf.BufferMutax.Lock()

	for {
		if len(buf.DataPackets) > self.OverlayInfo().CrPolicy.MNCache {

			if self.OverlayInfo().CrPolicy.MDCache > 0 {
				packet := buf.DataPackets[0]
				ti := packet.DateTime.Add(time.Minute * time.Duration(self.OverlayInfo().CrPolicy.MDCache))
				if ti.Before(now) {
					buf.DataPackets = buf.DataPackets[1:]
				} else {
					break
				}
			} else {
				buf.DataPackets = buf.DataPackets[1:]
			}
		} else {
			break
		}
	}

	buf.BufferMutax.Unlock()
}

func (self *WebrtcConnect) GetBuffermap() *[]*util.PeerBuffermap {
	self.CachingBufferMapMutex.Lock()

	bufmaps := new([]*util.PeerBuffermap)

	for _, buf := range self.CachingBufferMap {
		self.checkCachingBuffer(buf)

		buf.BufferMutax.Lock()

		bufmap := util.PeerBuffermap{}
		bufmap.SourcePeerId = buf.SourcePeerId

		for _, dp := range buf.DataPackets {
			bufmap.Sequence = append(bufmap.Sequence, dp.Sequence)
		}

		*bufmaps = append(*bufmaps, &bufmap)

		buf.BufferMutax.Unlock()
	}

	self.CachingBufferMapMutex.Unlock()

	return bufmaps
}

func (self *WebrtcConnect) GetCachingData(sourceId string, sequence int) *util.DataPacketPayload {
	self.CachingBufferMapMutex.Lock()

	for _, buf := range self.CachingBufferMap {
		if buf.SourcePeerId == sourceId {
			for _, dp := range buf.DataPackets {
				if dp.Sequence == sequence {
					self.CachingBufferMapMutex.Unlock()
					return &dp.Payload
				}
			}
		}
	}

	self.CachingBufferMapMutex.Unlock()

	return nil
}

func (self *WebrtcConnect) dataCaching(header *util.BroadcastData, data *[]byte, senderId *string) {

	if 0 >= self.OverlayInfo().CrPolicy.MNCache {
		return
	}

	self.CachingBufferMapMutex.Lock()
	buf := self.CachingBufferMap[*senderId]

	if buf == nil {
		buf = &util.CachingBuffer{}
		buf.SourcePeerId = *senderId
		self.CachingBufferMap[*senderId] = buf
	}
	self.CachingBufferMapMutex.Unlock()

	buf.BufferMutax.Lock()
	for _, dp := range buf.DataPackets {
		if dp.Sequence == header.ReqParams.Peer.Sequence {
			buf.BufferMutax.Unlock()
			return
		}
	}

	dp := util.DataPacket{}
	dp.DateTime = time.Now()
	dp.Sequence = header.ReqParams.Peer.Sequence
	dp.Payload.Header = header
	dp.Payload.Payload = data

	buf.DataPackets = append(buf.DataPackets, dp)
	buf.BufferMutax.Unlock()

	self.checkCachingBuffer(buf)

	//util.PrintJson(util.WORK, "CachingBuffer", self.CachingBufferMap)
	util.Println(util.INFO, "cachingbuffer-", *senderId, "-size: ", len(buf.DataPackets))
}

func (self *WebrtcConnect) checkBuffermapForBroadcast(bufmap *[]*util.PeerBuffermap) {

	mybufmap := self.GetBuffermap()

	missingbuf := self.checkBuffermap(bufmap, mybufmap)

	for _, buf := range missingbuf {
		for _, seq := range buf.Sequence {
			self.BroadcastCachingData(&buf.SourcePeerId, seq)
		}
	}
}

func (self *WebrtcConnect) checkBuffermapForGetData(peer *Peer, bufmap *[]*util.PeerBuffermap) int {

	mybufmap := self.GetBuffermap()

	missingbuf := self.checkBuffermap(mybufmap, bufmap)

	var sendCnt int = 0

	for _, buf := range missingbuf {
		for _, seq := range buf.Sequence {
			//self.BroadcastBuffermap(&buf.SourcePeerId, seq)
			peer.sendGetData(buf.SourcePeerId, seq, "")
			sendCnt++
		}
	}

	return sendCnt
}

// target과 비교해서 source에 없는것 리턴
func (self *WebrtcConnect) checkBuffermap(source *[]*util.PeerBuffermap, target *[]*util.PeerBuffermap) []util.PeerBuffermap {

	rsltBuf := make([]util.PeerBuffermap, 0)

	if source != nil && len(*source) > 0 {
		if target != nil && len(*target) > 0 {
			for _, tbuf := range *target {
				buffind := false
				for _, buf := range *source {
					if buf.SourcePeerId != tbuf.SourcePeerId {
						continue
					}

					missing := new(util.PeerBuffermap)
					missing.SourcePeerId = buf.SourcePeerId

					buffind = true

					for _, targetSeq := range tbuf.Sequence {
						find := false
						for _, seq := range buf.Sequence {
							if targetSeq == seq {
								find = true
								break
							}
						}

						if !find {
							missing.Sequence = append(missing.Sequence, targetSeq)
						}
					}

					rsltBuf = append(rsltBuf, *missing)
				}

				if !buffind {
					rsltBuf = append(rsltBuf, *tbuf)
				}
			}
		}
	} else {
		if target != nil {
			for _, buf := range *target {
				rsltBuf = append(rsltBuf, *buf)
			}
		}
	}

	return rsltBuf
}

func (self *WebrtcConnect) recoveryDataForPull() {

	util.Println(util.WORK, "Start data recovery - Pull")

	resChan := make(chan *util.BuffermapResponse)

	/*type peerBuf struct {
		PeerId     string
		Positon    connect.PeerPosition
		Buffermaps *[]*connect.PeerBuffermap
	}*/

	candiBufmaps := make(map[*Peer]*[]*util.PeerBuffermap, 0)
	primaryBufmap := new([]*util.PeerBuffermap)
	var primaryPeer *Peer = nil

	self.PeerMapMux.Lock()
	for _, peer := range self.peerMap {

		if peer.Position != connect.OutGoingCandidate && peer.Position != connect.OutGoingPrimary {
			continue
		}

		pb := new([]*util.PeerBuffermap)

		peer.sendBuffermap(&resChan)

		select {
		case res := <-resChan:
			pb = res.RspParams.Buffermap

			if peer.Position == connect.OutGoingCandidate {
				candiBufmaps[peer] = pb
			} else if peer.Position == connect.OutGoingPrimary {
				primaryPeer = peer
				primaryBufmap = pb
			}
		case <-time.After(time.Second * 5):
			util.Println(util.ERROR, peer.ToPeerId, "Recv Buffermap response timeout!")
			continue
		}
	}
	self.PeerMapMux.Unlock()

	close(resChan)

	util.Println(util.INFO, "Pull recovery response candi count:", len(candiBufmaps))

	self.checkBuffermapForBroadcast(primaryBufmap)

	for peer, candibuf := range candiBufmaps {
		sendCnt := self.checkBuffermapForGetData(peer, candibuf)
		if sendCnt > 0 {
			<-time.After(time.Second * 3)
		}
	}

	self.checkBuffermapForGetData(primaryPeer, primaryBufmap)

}

func (self *WebrtcConnect) Release(done *chan *util.HybridOverlayLeaveResponse) {
	defer close(*done)

	util.Println(util.INFO, " !!! Release overlay !!!")

	res := new(util.HybridOverlayLeaveResponse)
	res.RspCode = 200
	res.Overlay.OverlayId = self.OverlayInfo().OverlayId

	if !self.LeaveOverlay {
		res = self.OverlayLeave()

		if res == nil || res.RspCode != 200 {
			*done <- res
			return
		}
	}

	self.LeaveOverlay = true

	self.sendLeaveNoti()

	for _, peer := range self.peerMap {
		peer.SendRelease()
		<-time.After(time.Millisecond * 100)
		self.DisconnectFrom(peer)
	}
	//<-time.After(time.Millisecond * 100)

	//close(*done)

	*done <- res
}

func (self *WebrtcConnect) sendLeaveNoti() {
	util.Println(util.INFO, "Send leave noti")
	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControlLeave{}
	extHeader.ChannelId = self.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeLeave
	extHeader.LeaverInfo.PeerId = self.PeerId()
	extHeader.LeaverInfo.DisplayName = *self.PeerInfo().DisplayName

	payload := self.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, "SignStruct error!! payload is nil")
		return
	}

	self.BroadcastData(&params, &extHeader, &payload, self.PeerId(), false, true)
}

func (self *WebrtcConnect) SendTransmissionNoti(hom *util.HybridOverlayModification) {
	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControlTransmission{}
	extHeader.ChannelId = self.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeTransmission

	var change bool = false

	if hom.Overlay.ServiceInfo != nil && hom.Overlay.ServiceInfo.SourceList != nil {
		change = true
		srcList := make([]util.BroadcastDataExtensionHeaderControlPeerList, 0)
		for _, peerId := range *hom.Overlay.ServiceInfo.SourceList {
			srcList = append(srcList,
				util.BroadcastDataExtensionHeaderControlPeerList{PeerId: peerId})
		}
		extHeader.SourceList.ServiceSourceList = &srcList
	}

	if hom.Overlay.ServiceInfo != nil && hom.Overlay.ServiceInfo.BlockList != nil {
		change = true
		blkList := make([]util.BroadcastDataExtensionHeaderControlPeerList, 0)
		for _, peerId := range *hom.Overlay.ServiceInfo.BlockList {
			blkList = append(blkList,
				util.BroadcastDataExtensionHeaderControlPeerList{PeerId: peerId})
		}
		extHeader.SourceList.ServiceBlockList = &blkList
	}

	if hom.Overlay.ServiceInfo != nil && hom.Overlay.ServiceInfo.ChannelList != nil {
		chnList := make([]util.BroadcastDataExtensionHeaderControlTransmissionSourceListChannelSourceList, 0)
		chnchange := false
		for _, channel := range hom.Overlay.ServiceInfo.ChannelList {
			chnchange = true
			chnsrc := util.BroadcastDataExtensionHeaderControlTransmissionSourceListChannelSourceList{}
			chnsrc.ChannelId = channel.ChannelId
			chnsrc.SourceList = make([]util.BroadcastDataExtensionHeaderControlPeerList, 0)
			for _, peerId := range *channel.SourceList {
				chnsrc.SourceList = append(chnsrc.SourceList,
					util.BroadcastDataExtensionHeaderControlPeerList{PeerId: peerId})
			}
			chnList = append(chnList, chnsrc)
		}

		if chnchange {
			change = true
			extHeader.SourceList.ChannelSourceList = &chnList
		}
	}

	if !change {
		return
	}

	self.ApplyTransmissionNoti(&extHeader, false)

	payload := self.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, "SignStruct error!! payload is nil")
		return
	}

	util.Println(util.INFO, "SendTransmissionNoti:", extHeader)
	self.BroadcastData(&params, &extHeader, &payload, self.PeerId(), false, true)
}

func (self *WebrtcConnect) ApplyTransmissionNoti(noti *util.BroadcastDataExtensionHeaderControlTransmission, notiToApp bool) {
	util.PrintJson(util.INFO, "TransmissionControlNoti:", noti)

	var srclist []string = nil
	if noti.SourceList.ServiceSourceList != nil {
		srclist = make([]string, 0)
		for _, peer := range *noti.SourceList.ServiceSourceList {
			srclist = append(srclist, peer.PeerId)
		}
		self.OverlayInfo().ServiceInfo.SourceList = &srclist
	}

	if noti.SourceList.ServiceBlockList != nil {
		blocklist := make([]string, 0)
		for _, peer := range *noti.SourceList.ServiceBlockList {
			blocklist = append(blocklist, peer.PeerId)
		}
		self.OverlayInfo().ServiceInfo.BlockList = &blocklist
	}

	var channellist []util.SessionChangeInfoChannel = nil
	if noti.SourceList.ChannelSourceList != nil {
		channellist = make([]util.SessionChangeInfoChannel, 0)
		for _, chnsrc := range *noti.SourceList.ChannelSourceList {
			for idx, channel := range self.OverlayInfo().ServiceInfo.ChannelList {
				if channel.ChannelId == chnsrc.ChannelId {
					srclist := make([]string, 0)
					for _, peer := range chnsrc.SourceList {
						srclist = append(srclist, peer.PeerId)
					}
					//channel.SourceList = &srclist
					self.OverlayInfo().ServiceInfo.ChannelList[idx].SourceList = &srclist

					channellist = append(channellist, util.SessionChangeInfoChannel{ChannelId: channel.ChannelId, SourceList: &srclist})

					break
				}
			}
		}
	}

	if !notiToApp {
		return
	}

	if self.SessionChangeCallback != nil && (srclist != nil || channellist != nil) {
		change := util.SessionChangeInfo{}
		change.OverlayId = self.OverlayInfo().OverlayId
		if srclist != nil {
			change.SourceList = &srclist
		}
		if channellist != nil {
			change.ChannelList = &channellist
		}
		util.PrintJson(util.WORK, "!!! SessionChangeCallback:", change)
		self.SessionChangeCallback(&change)
	}
}

func (self *WebrtcConnect) SendOwnershipNoti(hom *util.HybridOverlayModification) {
	if hom.Ownership == nil || hom.Ownership.OwnerId == nil {
		return
	}

	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControlOwnership{}
	extHeader.ChannelId = self.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeOwnership

	extHeader.OwnerInfo.PeerId = *hom.Ownership.OwnerId

	payload := self.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, "SignStruct error!! payload is nil")
		return
	}

	self.ApplyOwnershipNoti(&extHeader, false)

	util.Println(util.INFO, "SendOwnershipNoti:", extHeader)
	self.BroadcastData(&params, &extHeader, &payload, self.PeerId(), false, true)
}

func (self *WebrtcConnect) ApplyOwnershipNoti(noti *util.BroadcastDataExtensionHeaderControlOwnership, notiToApp bool) {
	util.PrintJson(util.INFO, "OwnershipControlNoti:", noti)

	if len(noti.OwnerInfo.PeerId) <= 0 {
		return
	}

	self.OverlayInfo().OwnerId = noti.OwnerInfo.PeerId

	if !notiToApp {
		return
	}

	if self.SessionChangeCallback != nil {
		change := util.SessionChangeInfo{}
		change.OverlayId = self.OverlayInfo().OverlayId
		change.OwnerId = &noti.OwnerInfo.PeerId
		util.PrintJson(util.WORK, "!!! SessionChangeCallback:", change)
		self.SessionChangeCallback(&change)
	}
}

func (self *WebrtcConnect) SendSessionInfoNoti(hom *util.HybridOverlayModification) {
	if hom.Overlay.Title == nil && hom.Overlay.Description == nil &&
		(hom.Overlay.ServiceInfo == nil || (hom.Overlay.ServiceInfo.StartDatetime == nil && hom.Overlay.ServiceInfo.EndDatetime == nil)) {
		return
	}

	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControlSessionInfo{}
	extHeader.ChannelId = self.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeSessionInfo

	if hom.Overlay.Title != nil {
		extHeader.SessionInfo.Title = hom.Overlay.Title
	}
	if hom.Overlay.Description != nil {
		extHeader.SessionInfo.Description = hom.Overlay.Description
	}
	if hom.Overlay.ServiceInfo != nil && hom.Overlay.ServiceInfo.StartDatetime != nil {
		extHeader.SessionInfo.StartDateTime = hom.Overlay.ServiceInfo.StartDatetime
	}
	if hom.Overlay.ServiceInfo != nil && hom.Overlay.ServiceInfo.EndDatetime != nil {
		extHeader.SessionInfo.EndDateTime = hom.Overlay.ServiceInfo.EndDatetime
	}

	payload := self.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, "SignStruct error!! payload is nil")
		return
	}

	util.Println(util.INFO, "SendSessionInfoNoti:", extHeader)
	self.BroadcastData(&params, &extHeader, &payload, self.PeerId(), false, true)

	self.ApplySessionInfoNoti(&extHeader, false)
}

func (self *WebrtcConnect) ApplySessionInfoNoti(noti *util.BroadcastDataExtensionHeaderControlSessionInfo, notiToApp bool) {
	util.PrintJson(util.INFO, "SessionInfoControlNoti:", noti)

	if noti.SessionInfo.Title != nil {
		self.OverlayInfo().Title = *noti.SessionInfo.Title
	}

	if noti.SessionInfo.Description != nil {
		self.OverlayInfo().Description = *noti.SessionInfo.Description
	}

	if noti.SessionInfo.StartDateTime != nil {
		self.OverlayInfo().ServiceInfo.StartDatetime = noti.SessionInfo.StartDateTime
	}

	if noti.SessionInfo.EndDateTime != nil {
		self.OverlayInfo().ServiceInfo.EndDatetime = noti.SessionInfo.EndDateTime
	}

	if !notiToApp {
		return
	}

	if self.SessionChangeCallback != nil {
		change := util.SessionChangeInfo{}
		change.OverlayId = self.OverlayInfo().OverlayId
		change.Title = noti.SessionInfo.Title
		change.Description = noti.SessionInfo.Description
		change.StartDateTime = noti.SessionInfo.StartDateTime
		change.EndDateTime = noti.SessionInfo.EndDateTime
		util.PrintJson(util.WORK, "!!! SessionChangeCallback:", change)
		self.SessionChangeCallback(&change)
	}
}

func (self *WebrtcConnect) SendExpulsionNoti(peers *[]string) {
	if peers == nil || len(*peers) <= 0 {
		return
	}

	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControlExpulsion{}
	extHeader.ChannelId = self.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeExpulsion

	extHeader.ExpulsionList = make([]util.BroadcastDataExtensionHeaderControlPeerList, 0)
	for _, peerId := range *peers {
		peer := util.BroadcastDataExtensionHeaderControlPeerList{}
		peer.PeerId = peerId
		extHeader.ExpulsionList = append(extHeader.ExpulsionList, peer)
	}

	payload := self.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, "SignStruct error!! payload is nil")
		return
	}

	util.Println(util.INFO, "SendExpulsionNoti:", extHeader)
	self.BroadcastData(&params, &extHeader, &payload, self.PeerId(), true, true)

	self.ApplyExpulsionNoti(peers, false)
}

func (self *WebrtcConnect) ApplyExpulsionNoti(peers *[]string, notiToApp bool) {
	util.PrintJson(util.INFO, "ExpulsionControlNoti:", peers)

	if peers == nil || len(*peers) <= 0 {
		return
	}

	var expelMe bool = false

	for _, peerId := range *peers {
		if self.PeerId() == peerId || self.PeerOriginId() == peerId {
			expelMe = true
			self.LeaveOverlay = true
			break
		}
	}

	peermap := self.allPeer()

	for _, peer := range *peermap {
		if expelMe {
			self.DisconnectFrom(peer)
		} else {
			for _, peerId := range *peers {
				if peer.ToPeerId == peerId || peer.ToOriginId == peerId {
					peer.SendRelease()
					self.DelConnectionInfo(peer.Position, peer.ToPeerId)
					self.DisconnectFrom(peer)

					if self.PeerChangeCallback != nil {
						util.Println(util.WORK, "PeerChangeCallback - Leave !!!!!!!!!!!!!", peer.ToPeerId)
						self.PeerChangeCallback(self.OverlayInfo().OverlayId, peer.ToPeerId, *peer.Info.PeerInfo.DisplayName, true)
					}
					break
				}
			}
		}
	}

	if notiToApp && expelMe {
		if self.ExpulsionCallback != nil {
			util.Println(util.WORK, "!!! ExpulsionCallback:", self.OverlayInfo().OverlayId, self.PeerOriginId())
			self.ExpulsionCallback(self.OverlayInfo().OverlayId, self.PeerOriginId())
		}
	}
}

func (self *WebrtcConnect) SendSessionTerminationNoti() {
	params := util.BroadcastDataParams{}
	params.Operation = &util.BroadcastDataParamsOperation{}
	params.Operation.Ack = false
	params.Operation.CandidatePath = false
	params.Peer.PeerId = self.PeerId()
	params.Peer.Sequence = self.getBroadcastDataSeq()

	extHeader := util.BroadcastDataExtensionHeaderControl{}
	extHeader.ChannelId = self.ControlChannelId
	extHeader.MediaType = util.MediaTypeControl
	extHeader.ControlType = util.ControlTypeTermination

	payload := self.SignStruct(extHeader)
	if payload == nil {
		util.Println(util.ERROR, "SignStruct error!! payload is nil")
		return
	}

	util.Println(util.INFO, "SendSessionTerminationNoti:", extHeader)
	self.BroadcastData(&params, &extHeader, &payload, self.PeerId(), true, true)

	<-time.After(time.Second * 2)

	self.ApplyTerminationNoti(false)
}

func (self *WebrtcConnect) ApplyTerminationNoti(notiToApp bool) {
	peermap := self.allPeer()

	for _, peer := range *peermap {
		self.DisconnectFrom(peer)
	}

	if notiToApp {
		if self.SessionTerminationCallback != nil {
			util.Println(util.WORK, "!!! SessionTerminationCallback:", self.OverlayInfo().OverlayId)
			self.SessionTerminationCallback(self.OverlayInfo().OverlayId, self.PeerOriginId())
		}
	}
}

func (self *WebrtcConnect) NotiData(overlayId string, channelId string, senderId string, dataType string, data *[]byte) {
	if self.DataCallback != nil {
		util.Println(util.INFO, "!!! NotiData:", overlayId, channelId, senderId, dataType, data)
		self.DataCallback(overlayId, channelId, senderId, dataType, data)
	}
}
