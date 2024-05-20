package connect

import (
	"time"

	"hp2p.pb/hp2p_pb"
	"hp2p.util/util"
)

type ApiHandler struct {
	pbclient *hp2p_pb.PbClient
	recvChan *chan interface{}
	sendChan *chan interface{}

	connect *Connect

	isClosed bool
}

func NewApiHandler(grpcPort int, peerIndex string, connect *Connect) *ApiHandler {
	handler := new(ApiHandler)
	handler.pbclient = hp2p_pb.NewPbClient(grpcPort, peerIndex, (*connect).PeerInfo().PeerId)
	handler.recvChan = handler.pbclient.GetRecvChan()
	handler.sendChan = handler.pbclient.GetSendChan()
	handler.isClosed = false
	handler.connect = connect

	(*handler.connect).SetSessionTerminationCallback(handler.SessionTerminate)
	(*handler.connect).SetPeerChangeCallback(handler.PeerChange)
	(*handler.connect).SetSessionChangeCallback(handler.SessionChange)
	(*handler.connect).SetExpulsionCallback(handler.Expulsion)
	(*handler.connect).SetDataCallback(handler.Data)

	return handler
}

func (handler *ApiHandler) Close() {
	handler.isClosed = true
	handler.pbclient.Close()
}

func (handler *ApiHandler) Ready(index string) bool {
	return handler.pbclient.Ready(index)
}

func (handler *ApiHandler) SessionTerminate(overlayId string, peerId string) bool {
	return handler.pbclient.SessionTerminate(overlayId, peerId)
}

func (handler *ApiHandler) PeerChange(overlayId string, peerId string, displayName string, leave bool) bool {
	return handler.pbclient.PeerChange(overlayId, peerId, displayName, leave)
}

func (handler *ApiHandler) SessionChange(change *util.SessionChangeInfo) bool {
	return handler.pbclient.SessionChange(change)
}

func (handler *ApiHandler) Expulsion(overlayId string, peerId string) bool {
	return handler.pbclient.Expulsion(overlayId, peerId)
}

func (handler *ApiHandler) Data(overlayId string, channelId string, senderId string, dataType string, data *[]byte) bool {
	return handler.pbclient.Data(overlayId, channelId, senderId, dataType, data)
}

func (handler *ApiHandler) Recv() {
	go handler.pbclient.RecvRequest()
	go handler.pbclient.Heartbeat()

	for !handler.isClosed {
		select {
		case request := <-*handler.recvChan:
			var response interface{} = nil

			switch request.(type) {
			case *hp2p_pb.CreationRequest:
				response = *handler.handleCreationRequest(request.(*hp2p_pb.CreationRequest))
			case *hp2p_pb.JoinRequest:
				response = *handler.handleJoinRequest(request.(*hp2p_pb.JoinRequest))
			case *hp2p_pb.QueryRequest:
				response = *handler.handleQueryRequest(request.(*hp2p_pb.QueryRequest))
			case *hp2p_pb.ModificationRequest:
				response = *handler.handleModificationRequest(request.(*hp2p_pb.ModificationRequest))
			case *hp2p_pb.RemovalRequest:
				response = *handler.handleRemovalRequest(request.(*hp2p_pb.RemovalRequest))
			case *hp2p_pb.LeaveRequest:
				response = *handler.handleLeaveRequest(request.(*hp2p_pb.LeaveRequest))
			case *hp2p_pb.SearchPeerRequest:
				response = *handler.handleSearchPeerRequest(request.(*hp2p_pb.SearchPeerRequest))
			case *hp2p_pb.SendDataRequest:
				response = *handler.handleSendDataRequest(request.(*hp2p_pb.SendDataRequest))
			}

			if handler.isClosed {
				return
			}

			if response == nil || handler.sendChan == nil {
				util.Println(util.ERROR, "Failed to handle request in API handler. Response : ", response, ", SendChan : ", handler.sendChan)
				continue
			}

			select {
			case *handler.sendChan <- response:
			default:
				util.Println(util.ERROR, "Failed to send response in API handler.")
			}
		}
	}
}

func (handler *ApiHandler) handleCreationRequest(request *hp2p_pb.CreationRequest) *hp2p_pb.Response_Creation {
	conn := *handler.connect
	overlayCreation := new(util.HybridOverlayCreation)
	overlayCreation.Overlay.Title = request.GetTitle()
	overlayCreation.Overlay.Type = "core"
	overlayCreation.Overlay.SubType = "tree"
	overlayCreation.Overlay.OwnerId = request.GetOwnerId()
	overlayCreation.Overlay.Description = request.GetDescription()
	overlayCreation.Overlay.HeartbeatInterval = conn.GetHeartbeatInterval()
	overlayCreation.Overlay.HeartbeatTimeout = conn.GetHeartbeatTimeout()
	overlayCreation.Overlay.Auth.Type = "open"
	accessKey := request.GetAccessKey()
	if len(accessKey) > 0 {
		overlayCreation.Overlay.Auth.AccessKey = &accessKey
		overlayCreation.Overlay.Auth.Type = "closed"
	}
	overlayCreation.Overlay.Auth.AdminKey = request.GetAdminKey()
	peerList := request.GetPeerList()
	if peerList != nil && len(peerList) > 0 {
		overlayCreation.Overlay.Auth.PeerList = &peerList
		overlayCreation.Overlay.Auth.Type = "closed"
	}
	overlayCreation.Overlay.CrPolicy = conn.OverlayInfo().CrPolicy

	overlayCreation.Overlay.ServiceInfo.Title = &overlayCreation.Overlay.Title
	if len(overlayCreation.Overlay.Description) > 0 {
		overlayCreation.Overlay.ServiceInfo.Description = &overlayCreation.Overlay.Description
	}
	startDateTime := request.GetStartDateTime()
	if len(startDateTime) > 0 {
		overlayCreation.Overlay.ServiceInfo.StartDatetime = &startDateTime
	}
	endDateTime := request.GetEndDateTime()
	if len(endDateTime) > 0 {
		overlayCreation.Overlay.ServiceInfo.EndDatetime = &endDateTime
	}
	if request.GetUseSourceList() {
		sourceList := request.GetSourceList()
		overlayCreation.Overlay.ServiceInfo.SourceList = &sourceList
	}
	if request.GetUseBlockList() {
		blockList := request.GetBlockList()
		overlayCreation.Overlay.ServiceInfo.BlockList = &blockList
	}
	for _, channel := range request.ChannelList {
		ci := new(util.ChannelInfo)
		ci.ChannelId = channel.GetChannelId()
		ci.ChannelType = channel.GetChannelType()
		if channel.GetUseSourceList() {
			sourceList := channel.GetSourceList()
			ci.SourceList = &sourceList
		}

		if channel.ChannelType == "video/feature" {
			attr := new(util.ChannelAttributeVideoFeature)
			ci.ChannelAttribute = new(interface{})
			*ci.ChannelAttribute = &attr
			attr.Target = channel.GetVideoFeature().GetTarget()
			attr.Mode = channel.GetVideoFeature().GetMode()
			attr.Resolution = channel.GetVideoFeature().GetResolution()
			attr.FrameRate = channel.GetVideoFeature().GetFrameRate()
			attr.KeypointsType = channel.GetVideoFeature().GetKeyPointsType()
			attr.ModelUri = channel.GetVideoFeature().GetModelUri()
			attr.Hash = channel.GetVideoFeature().GetHash()
			attr.Dimension = channel.GetVideoFeature().GetDimension()
		} else if channel.ChannelType == "audio" {
			attr := new(util.ChannelAttributeAudio)
			ci.ChannelAttribute = new(interface{})
			*ci.ChannelAttribute = &attr
			attr.Codec = channel.GetAudio().GetCodec()
			attr.SampleRate = channel.GetAudio().GetSampleRate()
			attr.Bitrate = channel.GetAudio().GetBitrate()
			attr.ChannelType = channel.GetAudio().GetChannelType()
		} else if channel.ChannelType == "text" {
			attr := new(util.ChannelAttributeText)
			ci.ChannelAttribute = new(interface{})
			*ci.ChannelAttribute = &attr
			attr.Format = channel.GetText().GetFormat()
			attr.Encoding = channel.GetText().GetEncoding()
		}

		overlayCreation.Overlay.ServiceInfo.ChannelList = append(overlayCreation.Overlay.ServiceInfo.ChannelList, *ci)
	}

	ovinfo := conn.CreateOverlay(overlayCreation)

	if ovinfo == nil || len(ovinfo.OverlayId) <= 0 {
		util.Println(util.ERROR, "Failed to create overlay.")
		return &hp2p_pb.Response_Creation{
			Creation: &hp2p_pb.CreationResponse{
				RspCode: util.RspCode_Failed,
			},
		}
	}

	util.Println(util.INFO, "CreateOverlay ID : ", ovinfo.OverlayId)

	return &hp2p_pb.Response_Creation{
		Creation: &hp2p_pb.CreationResponse{
			RspCode:       util.RspCode_Success,
			OverlayId:     ovinfo.OverlayId,
			StartDateTime: *ovinfo.ServiceInfo.StartDatetime,
			EndDateTime:   *ovinfo.ServiceInfo.EndDatetime,
		},
	}
}

func (handler *ApiHandler) handleQueryRequest(request *hp2p_pb.QueryRequest) *hp2p_pb.Response_Query {
	conn := *handler.connect
	overlayId := request.GetOverlayId()
	title := request.GetTitle()
	description := request.GetDescription()

	hoq := conn.QueryOverlay(&overlayId, &title, &description)

	if hoq == nil {
		return &hp2p_pb.Response_Query{
			Query: &hp2p_pb.QueryResponse{
				RspCode: util.RspCode_Failed,
			},
		}
	}

	util.Println(util.INFO, "QueryOverlay : ", hoq)

	rslt := new(hp2p_pb.Response_Query)
	rslt.Query = new(hp2p_pb.QueryResponse)
	rslt.Query.OverlayList = make([]*hp2p_pb.QueryResponseOverlay, 0)
	rslt.Query.RspCode = util.RspCode_Success

	for _, ovinfo := range hoq.Overlay {
		overlay := new(hp2p_pb.QueryResponseOverlay)
		overlay.OverlayId = ovinfo.OverlayId
		overlay.OwnerId = ovinfo.OwnerId
		overlay.StartDateTime = *ovinfo.ServiceInfo.StartDatetime
		overlay.EndDateTime = *ovinfo.ServiceInfo.EndDatetime

		if ovinfo.Auth.Type == "open" {
			overlay.Closed = 0
		} else {
			overlay.Closed = 1
		}

		if ovinfo.ServiceInfo.Title != nil && len(*ovinfo.ServiceInfo.Title) > 0 {
			overlay.Title = *ovinfo.ServiceInfo.Title
		} else {
			overlay.Title = ovinfo.Title
		}

		if ovinfo.ServiceInfo.Description != nil && len(*ovinfo.ServiceInfo.Description) > 0 {
			overlay.Description = *ovinfo.ServiceInfo.Description
		} else if ovinfo.Description != nil && len(*ovinfo.Description) > 0 {
			overlay.Description = *ovinfo.Description
		} else {
			overlay.Description = ""
		}

		chlist := convertChannelList(&ovinfo.ServiceInfo.ChannelList)
		overlay.ChannelList = *chlist

		rslt.Query.OverlayList = append(rslt.Query.OverlayList, overlay)
	}

	return rslt
}

func convertChannelList(channelList *[]util.ChannelInfo) *[]*hp2p_pb.Channel {
	pbChannelList := make([]*hp2p_pb.Channel, 0)

	if channelList != nil && len(*channelList) > 0 {
		for _, channel := range *channelList {
			resch := new(hp2p_pb.Channel)
			resch.ChannelId = channel.ChannelId
			resch.ChannelType = channel.ChannelType
			resch.UseSourceList = channel.SourceList != nil //&& len(*channel.SourceList) > 0
			if resch.UseSourceList {
				resch.SourceList = *channel.SourceList
			}

			if channel.ChannelType == "control" {
				resch.ChannelAttribute = new(hp2p_pb.Channel_IsServiceChannel)
				resch.ChannelAttribute.(*hp2p_pb.Channel_IsServiceChannel).IsServiceChannel = true
			} else if channel.ChannelType == "video/feature" {
				pbattr := new(hp2p_pb.Channel_VideoFeature)
				pbattr.VideoFeature = new(hp2p_pb.ChannelAttributeVideoFeature)

				videoch := (*channel.ChannelAttribute).(util.ChannelAttributeVideoFeature)

				pbattr.VideoFeature.Target = videoch.Target
				pbattr.VideoFeature.Mode = videoch.Mode
				pbattr.VideoFeature.Resolution = videoch.Resolution
				pbattr.VideoFeature.FrameRate = videoch.FrameRate
				pbattr.VideoFeature.KeyPointsType = videoch.KeypointsType
				pbattr.VideoFeature.ModelUri = videoch.ModelUri
				pbattr.VideoFeature.Hash = videoch.Hash
				pbattr.VideoFeature.Dimension = videoch.Dimension

				resch.ChannelAttribute = pbattr
			} else if channel.ChannelType == "audio" {
				pbattr := new(hp2p_pb.Channel_Audio)
				pbattr.Audio = new(hp2p_pb.ChannelAttributeAudio)

				audioch := (*channel.ChannelAttribute).(util.ChannelAttributeAudio)

				pbattr.Audio.Codec = audioch.Codec
				pbattr.Audio.SampleRate = audioch.SampleRate
				pbattr.Audio.Bitrate = audioch.Bitrate
				pbattr.Audio.ChannelType = audioch.ChannelType

				resch.ChannelAttribute = pbattr
			} else if channel.ChannelType == "text" {
				pbattr := new(hp2p_pb.Channel_Text)
				pbattr.Text = new(hp2p_pb.ChannelAttributeText)

				textch := (*channel.ChannelAttribute).(util.ChannelAttributeText)

				pbattr.Text.Format = textch.Format
				pbattr.Text.Encoding = textch.Encoding

				resch.ChannelAttribute = pbattr
			}

			pbChannelList = append(pbChannelList, resch)
		}
	}

	return &pbChannelList
}

func (handler *ApiHandler) handleJoinRequest(request *hp2p_pb.JoinRequest) *hp2p_pb.Response_Join {
	conn := *handler.connect
	overlayId := request.GetOverlayId()
	accessKey := request.GetAccessKey()
	displayName := request.GetDisplayName()
	privateKeyPath := request.GetPrivateKeyPath()

	if !conn.SetPrivateKeyFromFile(privateKeyPath) {
		return &hp2p_pb.Response_Join{
			Join: &hp2p_pb.JoinResponse{
				RspCode: util.RspCode_PrivateKeyError,
			},
		}
	}

	conn.OverlayInfo().OverlayId = overlayId
	conn.OverlayInfo().Auth.AccessKey = &accessKey
	conn.PeerInfo().DisplayName = &displayName

	util.Println(util.INFO, "handleJoinRequest : ", request)

	hoj := conn.Join(false, accessKey)

	if hoj == nil {
		return &hp2p_pb.Response_Join{
			Join: &hp2p_pb.JoinResponse{
				RspCode: util.RspCode_Failed,
			},
		}
	}

	if hoj.RspCode == util.RspCode_AuthFailed2 {
		return &hp2p_pb.Response_Join{
			Join: &hp2p_pb.JoinResponse{
				RspCode: util.RspCode_AuthFailed,
			},
		}
	}

	if hoj.RspCode != util.RspCode_Success && hoj.RspCode != util.RspCode_TreeSuccess {
		return &hp2p_pb.Response_Join{
			Join: &hp2p_pb.JoinResponse{
				RspCode: int32(hoj.RspCode),
			},
		}
	}

	hoq := conn.QueryOverlay(&overlayId, nil, nil)
	if hoq == nil || len(hoq.Overlay) <= 0 {
		util.Println(util.ERROR, "Failed to query overlay.")
	} else {
		conn.OverlayInfo().Type = hoq.Overlay[0].Type
		conn.OverlayInfo().SubType = hoq.Overlay[0].SubType
		conn.OverlayInfo().OwnerId = hoq.Overlay[0].OwnerId
		conn.OverlayInfo().CrPolicy = hoq.Overlay[0].CrPolicy
	}

	if /*conn.CheckOwner(conn.PeerOriginId()) &&*/ len(hoj.Overlay.Status.PeerInfoList) <= 0 {
		conn.OverlayReport()
	} else {
		go conn.ConnectAfterJoin(&hoj.Overlay.Status.PeerInfoList, accessKey)
	}

	joinResponse := new(hp2p_pb.JoinResponse)
	joinResponse.RspCode = util.RspCode_Success
	joinResponse.OverlayId = conn.OverlayInfo().OverlayId
	joinResponse.Title = *conn.OverlayInfo().ServiceInfo.Title
	joinResponse.Description = *conn.OverlayInfo().ServiceInfo.Description
	svcInfo := conn.OverlayInfo().ServiceInfo
	joinResponse.StartDateTime = *svcInfo.StartDatetime
	joinResponse.EndDateTime = *svcInfo.EndDatetime
	joinResponse.SourceList = *svcInfo.SourceList
	chlist := convertChannelList(&svcInfo.ChannelList)
	joinResponse.ChannelList = *chlist

	return &hp2p_pb.Response_Join{
		Join: joinResponse,
	}
}

func (handler *ApiHandler) handleModificationRequest(request *hp2p_pb.ModificationRequest) *hp2p_pb.Response_Modification {
	conn := *handler.connect
	overlayId := request.GetOverlayId()
	ownerId := request.GetOwnerId()
	adminKey := request.GetAdminKey()

	hreq := util.HybridOverlayModification{}
	hreq.Overlay = util.HybridOverlayModificationOverlay{}
	hreq.Overlay.OverlayId = overlayId
	hreq.Overlay.OwnerId = ownerId
	hreq.Overlay.Auth.AdminKey = adminKey

	var expulsionList []string = make([]string, 0)

	if request.GetModificationTitle() {
		if hreq.Overlay.ServiceInfo == nil {
			hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
		}
		title := request.GetTitle()
		hreq.Overlay.ServiceInfo.Title = &title
	}

	if request.GetModificationDescription() {
		if hreq.Overlay.ServiceInfo == nil {
			hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
		}
		description := request.GetDescription()
		hreq.Overlay.ServiceInfo.Description = &description
	}

	if request.GetModificationAccessKey() {
		accessKey := request.GetAccessKey()
		hreq.Overlay.Auth.AccessKey = &accessKey
	}

	if request.GetModificationPeerList() {
		peerList := make([]string, 0)
		peerList = append(peerList, request.GetPeerList()...)

		// 기존 PeerList에서 제외되는 PeerList를 추출한다.
		if conn.OverlayInfo().Auth.PeerList != nil && len(*conn.OverlayInfo().Auth.PeerList) > 0 {
			if len(peerList) <= 0 {
				expulsionList = append(expulsionList, *conn.OverlayInfo().Auth.PeerList...)
			} else {
				for _, peer := range *conn.OverlayInfo().Auth.PeerList {
					found := false
					for _, newpeer := range peerList {
						if peer == newpeer {
							found = true
							break
						}
					}

					if !found {
						expulsionList = append(expulsionList, peer)
					}
				}
			}
		}

		hreq.Overlay.Auth.PeerList = &peerList
	}

	if request.GetModificationStartDateTime() {
		if hreq.Overlay.ServiceInfo == nil {
			hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
		}
		startDateTime := request.GetStartDateTime()
		hreq.Overlay.ServiceInfo.StartDatetime = &startDateTime
	}

	if request.GetModificationEndDateTime() {
		if hreq.Overlay.ServiceInfo == nil {
			hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
		}
		endDateTime := request.GetEndDateTime()
		hreq.Overlay.ServiceInfo.EndDatetime = &endDateTime
	}

	if request.GetModificationSourceList() {
		if hreq.Overlay.ServiceInfo == nil {
			hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
		}
		sourceList := make([]string, 0)
		sourceList = append(sourceList, request.GetSourceList()...)
		hreq.Overlay.ServiceInfo.SourceList = &sourceList
	}

	if request.GetModificationBlockList() {
		if hreq.Overlay.ServiceInfo == nil {
			hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
		}
		blockList := make([]string, 0)
		blockList = append(blockList, request.GetBlockList()...)
		hreq.Overlay.ServiceInfo.BlockList = &blockList

		expulsionList = append(expulsionList, blockList...)
	}

	if request.GetModificationChannelList() {
		chlist := request.GetChannelList()
		if chlist != nil && len(chlist) > 0 {
			if hreq.Overlay.ServiceInfo == nil {
				hreq.Overlay.ServiceInfo = new(util.ServiceInfo)
			}

			for _, ch := range chlist {
				ci := new(util.ChannelInfo)
				ci.ChannelId = ch.GetChannelId()
				sourceList := make([]string, 0)
				if ch.GetUseSourceList() {
					sourceList = append(sourceList, ch.GetSourceList()...)
					ci.SourceList = &sourceList
				}

				hreq.Overlay.ServiceInfo.ChannelList = append(hreq.Overlay.ServiceInfo.ChannelList, *ci)
			}
		}
	}

	if request.GetModificationOwnerId() {
		if hreq.Ownership == nil {
			hreq.Ownership = new(util.HybridOverlayModificationOwnership)
		}
		newowner := request.GetNewOwnerId()
		hreq.Ownership.OwnerId = &newowner
	}

	if request.GetModificationAdminKey() {
		if hreq.Ownership == nil {
			hreq.Ownership = new(util.HybridOverlayModificationOwnership)
		}
		newadmin := request.GetNewAdminKey()
		hreq.Ownership.AdminKey = &newadmin
	}

	hres := conn.OverlayModification(&hreq)

	if hres == nil {
		return &hp2p_pb.Response_Modification{
			Modification: &hp2p_pb.ModificationResponse{
				RspCode: util.RspCode_Failed,
			},
		}
	}

	if hres.RspCode == util.RspCode_Success {
		conn.SendTransmissionNoti(&hreq)
		conn.SendOwnershipNoti(&hreq)
		conn.SendSessionInfoNoti(&hreq)
		conn.SendExpulsionNoti(&expulsionList)
	}

	return &hp2p_pb.Response_Modification{
		Modification: &hp2p_pb.ModificationResponse{
			RspCode: int32(hres.RspCode),
		},
	}
}

func (handler *ApiHandler) handleSendDataRequest(request *hp2p_pb.SendDataRequest) *hp2p_pb.Response_SendData {
	conn := *handler.connect
	dataType := request.GetDataType()
	data := request.GetData()
	channelId := request.GetChannelId()

	if data == nil || len(data) <= 0 {
		return &hp2p_pb.Response_SendData{
			SendData: &hp2p_pb.SendDataResponse{
				RspCode: util.RspCode_WrongRequest,
			},
		}
	}

	if dataType != util.MediaTypeData && dataType != util.MediaTypeDataCache {
		return &hp2p_pb.Response_SendData{
			SendData: &hp2p_pb.SendDataResponse{
				RspCode: util.RspCode_WrongRequest,
			},
		}
	}

	if channelId == "" {
		return &hp2p_pb.Response_SendData{
			SendData: &hp2p_pb.SendDataResponse{
				RspCode: util.RspCode_WrongRequest,
			},
		}
	}

	res := conn.SendData(channelId, dataType, data)

	return &hp2p_pb.Response_SendData{
		SendData: &hp2p_pb.SendDataResponse{
			RspCode: res,
		},
	}
}

func (handler *ApiHandler) handleRemovalRequest(request *hp2p_pb.RemovalRequest) *hp2p_pb.Response_Removal {
	conn := *handler.connect
	overlayId := request.GetOverlayId()
	ownerId := request.GetOwnerId()
	adminKey := request.GetAdminKey()

	hreq := util.HybridOverlayRemoval{}
	hreq.Overlay = util.HybridOverlayRemovalOverlay{}
	hreq.Overlay.OverlayId = overlayId
	hreq.Overlay.OwnerId = ownerId
	hreq.Overlay.Auth.AdminKey = adminKey

	hres := conn.OverlayRemove(&hreq)

	if hres == nil {
		return &hp2p_pb.Response_Removal{
			Removal: &hp2p_pb.RemovalResponse{
				RspCode: util.RspCode_Failed,
			},
		}
	}

	conn.SendSessionTerminationNoti()

	return &hp2p_pb.Response_Removal{
		Removal: &hp2p_pb.RemovalResponse{
			RspCode: int32(hres.RspCode),
		},
	}
}

func (handler *ApiHandler) handleLeaveRequest(request *hp2p_pb.LeaveRequest) *hp2p_pb.Response_Leave {
	conn := *handler.connect
	overlayId := request.GetOverlayId()
	peerId := request.GetPeerId()
	accessKey := request.GetAccessKey()

	if conn.OverlayInfo().OverlayId != overlayId || conn.PeerOriginId() != peerId {
		return &hp2p_pb.Response_Leave{
			Leave: &hp2p_pb.LeaveResponse{
				RspCode: util.RspCode_NotJoined,
			},
		}
	}

	if len(accessKey) > 0 {
		conn.OverlayInfo().Auth.AccessKey = &accessKey
	}

	done := make(chan *util.HybridOverlayLeaveResponse)
	go conn.Release(&done)

	select {
	case hres := <-done:
		if hres == nil {
			return &hp2p_pb.Response_Leave{
				Leave: &hp2p_pb.LeaveResponse{
					RspCode: util.RspCode_Failed,
				},
			}
		}

		return &hp2p_pb.Response_Leave{
			Leave: &hp2p_pb.LeaveResponse{
				RspCode: int32(hres.RspCode),
			},
		}
	case <-time.After(time.Second * 10):
		return &hp2p_pb.Response_Leave{
			Leave: &hp2p_pb.LeaveResponse{
				RspCode: util.RspCode_Failed,
			},
		}
	}
}

func (handler *ApiHandler) handleSearchPeerRequest(request *hp2p_pb.SearchPeerRequest) *hp2p_pb.Response_SearchPeer {
	conn := *handler.connect
	overlayId := request.GetOverlayId()

	if conn.OverlayInfo().OverlayId != overlayId {
		return &hp2p_pb.Response_SearchPeer{
			SearchPeer: &hp2p_pb.SearchPeerResponse{
				RspCode: util.RspCode_NotJoined,
			},
		}
	}

	rslt := &hp2p_pb.Response_SearchPeer{
		SearchPeer: &hp2p_pb.SearchPeerResponse{
			PeerList: make([]*hp2p_pb.Peer, 0),
			RspCode:  util.RspCode_Success,
		},
	}

	conn.JoinPeersLock()
	peers := conn.GetJoinPeers()

	if peers != nil || len(*peers) > 0 {
		for _, peer := range *peers {
			if peer.PeerId != conn.PeerId() {
				rslt.SearchPeer.PeerList = append(rslt.SearchPeer.PeerList, &hp2p_pb.Peer{
					PeerId:      peer.PeerId,
					DisplayName: peer.DisplayName,
				})
			}
		}
	}
	conn.JoinPeersUnlock()

	return rslt
}
