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

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"connect"
	"connect/webrtc"

	"hp2p.util/util"
)

func main() {
	peerId := flag.String("id", "", "Peer ID")
	displayName := flag.String("dn", "", "Display Name")
	title := flag.String("t", "", "Title")
	desc := flag.String("d", "", "Description")
	create := flag.Bool("c", false, "Create Overlay")
	join := flag.Bool("j", false, "Join Overlay")
	gapi := flag.Bool("g", false, "Use Grpc API")
	hbInterval := flag.Int("hi", 10, "Heartbeat interval")
	hbTimeout := flag.Int("ht", 15, "Heartbeat timeout")
	//authType := flag.String("a", "open", "Auth type [open|closed]")
	adminKey := flag.String("ak", "admin", "Admin key")
	accessKey := flag.String("ac", "", "Access key")
	peerList := flag.String("pl", "", "Peer list [peer1,peer2,...]")
	mnCache := flag.Int("mn", 0, "mN Cache")
	mdCache := flag.Int("md", 0, "mD Cache(minute)")
	recoveryBy := flag.String("r", "push", "Recovery by [push|pull]")
	rateControlQuantity := flag.Int("rq", 0, "Rate control quantity(count). 0 is unlimit.")
	ratecontrolBitrate := flag.Int("rb", 0, "Rate control bitrate(Kbps). 0 is unlimit.")
	authList := flag.String("al", "", "Auth list [peer1,peer2,...]")
	ovId := flag.String("o", "", "Overlay ID for join")
	grpcPort := flag.Int("gp", 50051, "gRpc server port. default 50051.")
	peerIndex := flag.String("pi", "", "Peer index.")
	configPath := flag.String("cp", "", "Config file path.")

	flag.Parse()

	if len(*peerId) <= 0 {
		util.Println(util.ERROR, "Need Peer ID. Use -h for usage.")
		return
	}

	if *create && len(*title) <= 0 {
		util.Println(util.ERROR, "Need Title. Use -h for usage.")
		return
	}

	if *join && len(*ovId) <= 0 && len(*title) <= 0 && len(*desc) <= 0 {
		util.Println(util.ERROR, "Need Overlay ID or Title or Description. Use -h for usage.")
		return
	}

	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)

	os.Chdir(exPath)
	var conn connect.Connect = &webrtc.WebrtcConnect{}
	conn.Init(*peerId, *configPath)
	conn.SetHeartbeatInterval(*hbInterval)
	conn.SetHeartbeatTimeout(*hbTimeout)
	conn.SetPeerIndex(*peerIndex)
	conn.PeerInfo().DisplayName = displayName
	crPolicy := new(util.CrPolicy)
	crPolicy.MNCache = *mnCache
	crPolicy.MDCache = *mdCache
	crPolicy.RecoveryBy = *recoveryBy
	conn.OverlayInfo().CrPolicy = crPolicy

	util.Println(util.INFO, "Peer Id :", conn.PeerId())
	util.Println(util.INFO, "Display Name :", *displayName)
	util.Println(util.INFO, "Title :", *title)
	util.Println(util.INFO, "Description :", *desc)
	util.Println(util.INFO, "Create :", *create)
	util.Println(util.INFO, "Heartbeat interval :", *hbInterval)
	util.Println(util.INFO, "Heartbeat timeout :", *hbTimeout)
	//util.Println(util.INFO, "Auth type :", *authType)
	util.Println(util.INFO, "Admin key :", *adminKey)
	util.Println(util.INFO, "Access key :", *accessKey)
	util.Println(util.INFO, "Peer List :", *peerList)
	util.Println(util.INFO, "mN Cache :", *mnCache)
	util.Println(util.INFO, "mD Cache(minute) :", *mdCache)
	util.Println(util.INFO, "Recovery by :", *recoveryBy)
	util.Println(util.INFO, "Rate control quantity(count) :", *rateControlQuantity)
	util.Println(util.INFO, "Rate control bitrate(Kbps) :", *ratecontrolBitrate)
	util.Println(util.INFO, "Auth List :", *authList)
	util.Println(util.INFO, "Overlay ID :", *ovId)
	util.Println(util.INFO, "gRPC Server port :", *grpcPort)
	util.Println(util.INFO, "Config file path:", *configPath)
	util.Println(util.INFO, "Peer index:", *peerIndex)
	util.Println(util.INFO, "Use Grpc API:", *gapi)

	util.Printf(util.INFO, "\n\nHP2P.Go start...\n\n")

	runtime.GOMAXPROCS(runtime.NumCPU())
	util.Printf(util.INFO, "Use %v processes.\n\n", runtime.GOMAXPROCS(0))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var apihandler *connect.ApiHandler = nil

	if *gapi {
		apihandler = connect.NewApiHandler(*grpcPort, *peerIndex, &conn)

		if !apihandler.Ready(*peerIndex) {
			util.Println(util.ERROR, "Failed to connect to API.")
			return
		}

		go apihandler.Recv()
	}

	if *create {
		var overlayCreation *util.HybridOverlayCreation = nil

		overlayCreation = new(util.HybridOverlayCreation)
		overlayCreation.Overlay.Title = *title
		overlayCreation.Overlay.Type = "core"
		overlayCreation.Overlay.SubType = "tree"
		overlayCreation.Overlay.OwnerId = *peerId
		overlayCreation.Overlay.Description = *desc
		if len(overlayCreation.Overlay.Description) <= 0 {
			overlayCreation.Overlay.Description = "no description"
		}
		overlayCreation.Overlay.HeartbeatInterval = *hbInterval
		overlayCreation.Overlay.HeartbeatTimeout = *hbTimeout

		auth := new(util.OverlayAuth)
		auth.AdminKey = *adminKey
		auth.AccessKey = nil
		auth.PeerList = nil
		auth.Type = "open"
		if len(*peerList) > 0 {
			plist := strings.Split(*peerList, ",")
			auth.PeerList = &plist

			if len(*accessKey) > 0 {
				auth.AccessKey = accessKey
			}
			auth.Type = "closed"
		} else {
			if len(*accessKey) > 0 {
				auth.AccessKey = accessKey
				auth.Type = "closed"
			}
		}
		overlayCreation.Overlay.Auth = *auth

		crPolicy := new(util.CrPolicy)
		crPolicy.MNCache = *mnCache
		crPolicy.MDCache = *mdCache
		crPolicy.RecoveryBy = *recoveryBy
		overlayCreation.Overlay.CrPolicy = crPolicy

		serviceInfo := new(util.ServiceInfo)
		// st := "20230824171200"
		// serviceInfo.StartDatetime = &st
		//et := "20231123233100"
		//serviceInfo.EndDatetime = &et
		serviceInfo.SourceList = new([]string)
		*serviceInfo.SourceList = append(*serviceInfo.SourceList, "*")
		//*serviceInfo.SourceList = append(*serviceInfo.SourceList, "peer1")
		//*serviceInfo.SourceList = append(*serviceInfo.SourceList, "peer2")
		// serviceInfo.BlockList = new([]string)
		// *serviceInfo.BlockList = append(*serviceInfo.BlockList, "peer3")
		// *serviceInfo.BlockList = append(*serviceInfo.BlockList, "peer4")
		// *serviceInfo.BlockList = append(*serviceInfo.BlockList, "peer5")

		channelInfo := new(util.ChannelInfo)
		channelInfo.ChannelId = "channel1"
		channelInfo.ChannelType = "control"
		serviceInfo.ChannelList = append(serviceInfo.ChannelList, *channelInfo)

		channelInfo = new(util.ChannelInfo)
		channelInfo.ChannelId = "channel2"
		channelInfo.ChannelType = "video/feature"
		channelAttributeVideoFeature := new(util.ChannelAttributeVideoFeature)
		channelInfo.ChannelAttribute = new(interface{})
		*channelInfo.ChannelAttribute = &channelAttributeVideoFeature
		channelAttributeVideoFeature.Target = "face"
		channelAttributeVideoFeature.Mode = "KDM"
		channelAttributeVideoFeature.Resolution = "256x256"
		channelAttributeVideoFeature.FrameRate = "30fps"
		channelAttributeVideoFeature.KeypointsType = "face/68-multipi"
		//channelInfo.SourceList = new([]string)
		//*channelInfo.SourceList = append(*channelInfo.SourceList, "peer3")
		//*channelInfo.SourceList = append(*channelInfo.SourceList, "peer4")
		serviceInfo.ChannelList = append(serviceInfo.ChannelList, *channelInfo)

		channelInfo = new(util.ChannelInfo)
		channelInfo.ChannelId = "channel3"
		channelInfo.ChannelType = "audio"
		channelAudio := new(util.ChannelAttributeAudio)
		channelInfo.ChannelAttribute = new(interface{})
		*channelInfo.ChannelAttribute = &channelAudio
		channelAudio.Bitrate = "128kbps"
		channelAudio.SampleRate = "44100"
		channelAudio.Codec = "AAC"
		channelAudio.ChannelType = "stereo"
		//channelInfo.SourceList = new([]string)
		//*channelInfo.SourceList = append(*channelInfo.SourceList, "peer5")
		//*channelInfo.SourceList = append(*channelInfo.SourceList, "peer6")
		serviceInfo.ChannelList = append(serviceInfo.ChannelList, *channelInfo)

		channelInfo = new(util.ChannelInfo)
		channelInfo.ChannelId = "channel4"
		channelInfo.ChannelType = "text"
		channelText := new(util.ChannelAttributeText)
		channelInfo.ChannelAttribute = new(interface{})
		*channelInfo.ChannelAttribute = &channelText
		channelText.Encoding = "UTF-8"
		channelText.Format = "text/plain"
		serviceInfo.ChannelList = append(serviceInfo.ChannelList, *channelInfo)

		overlayCreation.Overlay.ServiceInfo = *serviceInfo

		conn.CreateOverlay(overlayCreation)

		if conn.OverlayInfo() == nil || len(conn.OverlayInfo().OverlayId) <= 0 {
			util.Println(util.ERROR, "Failed to create overlay.")
			return
		}

		util.Println(util.INFO, "CreateOverlay ID : ", conn.OverlayInfo().OverlayId)

		genprivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Println("Error generating RSA private key:", err)
			os.Exit(1)
		}

		if !conn.SetPrivateKey(genprivateKey) {
			util.Println(util.ERROR, "Failed to set private key.")
			return
		}

		ovinfo := conn.Join(false, *accessKey)

		if ovinfo == nil {
			return
		}

		conn.OverlayReport()
	} else if *join {
		conn.OverlayQuery(ovId, title, desc)

		genprivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Println("Error generating RSA private key:", err)
			os.Exit(1)
		}

		if !conn.SetPrivateKey(genprivateKey) {
			util.Println(util.ERROR, "Failed to set private key.")
			return
		}

		go conn.JoinAndConnect(false, "asdf")
	}

	//time.Sleep(time.Minute * 10)

	<-interrupt
	util.Println(util.INFO, "interrupt!!!!!")
	done := make(chan *util.HybridOverlayLeaveResponse)

	if *gapi {
		apihandler.Close()
		//pbclient.Close()
	}
	go conn.Release(&done)

	select {
	case res := <-done:
		util.Println(util.INFO, "Release response : ", res)
	case <-time.After(time.Second * 10):
	}
}
