package hp2p_pb

import (
	context "context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"hp2p.util/util"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PbClient struct {
	ServerPort  int
	clientConn  *grpc.ClientConn
	protoClient Hp2PApiProtoClient
	isClosed    bool

	peerIndex string

	recvChan chan interface{}
	sendChan chan interface{}
}

func NewPbClient(serverPort int, peerIndex string) *PbClient {
	client := new(PbClient)
	client.ServerPort = serverPort
	client.clientConn = nil
	client.protoClient = nil
	client.isClosed = false
	client.peerIndex = peerIndex

	client.recvChan = make(chan interface{})
	client.sendChan = make(chan interface{})

	return client
}

func (client *PbClient) Close() {
	client.isClosed = true

	close(client.recvChan)
	close(client.sendChan)

	if client.clientConn != nil {
		client.clientConn.Close()
	}
}

func (client *PbClient) SetPeerIndex(peerIndex string) {
	client.peerIndex = peerIndex
}

func (client *PbClient) GetRecvChan() *chan interface{} {
	return &client.recvChan
}

func (client *PbClient) GetSendChan() *chan interface{} {
	return &client.sendChan
}

func (client *PbClient) getProtoClient() *Hp2PApiProtoClient {
	if client.protoClient == nil {
		util.Println(util.INFO, "Connecting to API...")

		var err error = nil
		client.clientConn, err = grpc.Dial("localhost:"+strconv.Itoa(client.ServerPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			util.Println(util.ERROR, "Did not connect to API: ", err)
			return nil
		}

		client.protoClient = NewHp2PApiProtoClient(client.clientConn)

		util.Println(util.INFO, "Connected to API")
	}

	return &client.protoClient
}

func (client *PbClient) SendResponse(pstream *Hp2PApiProto_HompClient) {
	stream := *pstream
	for !client.isClosed {
		select {
		case response := <-client.sendChan:
			if client.isClosed {
				return
			}
			util.Println(util.INFO, "SendResponse: ", response)

			resWithId := Response{}
			resWithId.Id = client.peerIndex

			switch response.(type) {
			case Response_Creation:
				res := response.(Response_Creation)
				resWithId.Response = &res

				if res.Creation.RspCode == util.RspCode_Success {
					client.peerIndex = res.Creation.OverlayId
				}
			case Response_Join:
				res := response.(Response_Join)
				resWithId.Response = &res

				if res.Join.RspCode == util.RspCode_Success {
					client.peerIndex = (*res.Join).OverlayId
				}
			case Response_Query:
				res := response.(Response_Query)
				resWithId.Response = &res
			case Response_Modification:
				res := response.(Response_Modification)
				resWithId.Response = &res
			case Response_Removal:
				res := response.(Response_Removal)
				resWithId.Response = &res
			case Response_Leave:
				res := response.(Response_Leave)
				resWithId.Response = &res
			case Response_SearchPeer:
				res := response.(Response_SearchPeer)
				resWithId.Response = &res
			case Response_SendData:
				res := response.(Response_SendData)
				resWithId.Response = &res
			default:
				util.Println(util.ERROR, "Failed to send response. Response : ", response)
				continue
			}

			stream.Send(&resWithId)

			//default:
			//	continue
		}
	}
}

func (client *PbClient) RecvRequest() {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		util.Println(util.ERROR, "protoClient is nil")
		return
	}
	protoClient := *p_protoClient

	ctx := context.Background()
	stream, err := protoClient.Homp(ctx)
	if err != nil {
		util.Println(util.ERROR, "Could not create stream: ", err)
		return
	}

	go client.SendResponse(&stream)

	for !client.isClosed {
		request, err := stream.Recv()

		if client.isClosed {
			return
		}

		if err == nil && request != nil {
			util.Println(util.INFO, "RecvRequest: ", request)

			id := request.GetId()
			if id != client.peerIndex {
				util.Println(util.ERROR, "Received request is not for me: ", id)
				continue
			}

			if request.GetCreation() != nil {
				creationReq := request.GetCreation()
				util.Println(util.INFO, "CreationRequest: ", creationReq)

				select {
				case client.recvChan <- creationReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					creationResp := Response_Creation{
						Creation: &CreationResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &creationResp})
				}

			} else if request.GetJoin() != nil {
				joinReq := request.GetJoin()
				util.Println(util.INFO, "JoinRequest: ", joinReq)

				select {
				case client.recvChan <- joinReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					joinResp := Response_Join{
						Join: &JoinResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &joinResp})
				}
			} else if request.GetQuery() != nil {
				queryReq := request.GetQuery()
				util.Println(util.INFO, "QueryRequest: ", queryReq)

				select {
				case client.recvChan <- queryReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					queryResp := Response_Query{
						Query: &QueryResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &queryResp})
				}
			} else if request.GetModification() != nil {
				modReq := request.GetModification()
				util.Println(util.INFO, "ModificationRequest: ", modReq)

				select {
				case client.recvChan <- modReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					modResp := Response_Modification{
						Modification: &ModificationResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &modResp})
				}
			} else if request.GetSendData() != nil {
				sendDataReq := request.GetSendData()
				util.Println(util.INFO, "SendDataRequest: ", sendDataReq)

				select {
				case client.recvChan <- sendDataReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					sendDataResp := Response_SendData{
						SendData: &SendDataResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &sendDataResp})
				}
			} else if request.GetRemoval() != nil {
				remReq := request.GetRemoval()
				util.Println(util.INFO, "RemovalRequest: ", remReq)

				select {
				case client.recvChan <- remReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					remResp := Response_Removal{
						Removal: &RemovalResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &remResp})
				}
			} else if request.GetLeave() != nil {
				leaveReq := request.GetLeave()
				util.Println(util.INFO, "LeaveRequest: ", leaveReq)

				select {
				case client.recvChan <- leaveReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					leaveResp := Response_Leave{
						Leave: &LeaveResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &leaveResp})
				}
			} else if request.GetSearchPeer() != nil {
				searchPeerReq := request.GetSearchPeer()
				util.Println(util.INFO, "SearchPeerRequest: ", searchPeerReq)

				select {
				case client.recvChan <- searchPeerReq:
				default:
					util.Println(util.ERROR, "recvChan is full or nil or closed or no receiver")

					searchPeerResp := Response_SearchPeer{
						SearchPeer: &SearchPeerResponse{RspCode: util.RspCode_Failed},
					}

					stream.Send(&Response{Id: client.peerIndex, Response: &searchPeerResp})
				}
			} else {
				util.Println(util.ERROR, "Received request is unknown: ", request)
			}
		} else {
			if err == io.EOF {
				util.Println(util.INFO, "Received EOF")
				break
			}

			util.Println(util.ERROR, "Error receiving request: ", err)
		}
	}

	return
}

func (client *PbClient) Ready(index string) bool {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		return false
	}
	protoClient := *p_protoClient

	req := ReadyRequest{Index: index}
	util.Println(util.INFO, "ReadyRequest: ", &req)

	r, err := protoClient.Ready(context.Background(), &req)
	if err != nil {
		util.Println(util.ERROR, "Could not send ready to API: ", err)
		return false
	}

	util.Println(util.INFO, "Success to send ready to API: ", r.IsOk)

	return r.IsOk
}

func (client *PbClient) Heartbeat() {
	for !client.isClosed {
		<-time.After(3 * time.Second)

		p_protoClient := client.getProtoClient()
		if p_protoClient == nil {
			util.Println(util.ERROR, "Could not get protoClient")
			return
		}
		protoClient := *p_protoClient

		req := HeartbeatRequest{Dummy: 1}
		r, err := protoClient.Heartbeat(context.Background(), &req)
		if err != nil {
			//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			client.killProcess()
			return
		}

		if !r.Response {
			//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			client.killProcess()
			return
		}
	}
}

func (client *PbClient) killProcess() {
	pid := os.Getpid()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	} else {
		cmd = exec.Command("kill", "-9", strconv.Itoa(pid))
	}

	err := cmd.Run()
	if err != nil {
		util.Println(util.ERROR, "Failed to kill process: ", err)
		os.Exit(1)
	}
}

func (client *PbClient) SessionTerminate(overlayId string, peerId string) bool {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		util.Println(util.ERROR, "Could not get protoClient")
		return false
	}
	protoClient := *p_protoClient

	req := SessionTerminateRequest{OverlayId: overlayId, PeerId: peerId}
	r, err := protoClient.SessionTermination(context.Background(), &req)
	if err != nil {
		util.Println(util.ERROR, "Could not send termination to API: ", err)
	}

	if !r.Ack {
		util.Println(util.ERROR, "Termination Response is FALSE")
	}

	util.Println(util.INFO, "PBClient SessionTerminate")
	return r.Ack
	//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}

func (client *PbClient) PeerChange(overlayId string, peerId string, displayName string, leave bool) bool {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		util.Println(util.ERROR, "Could not get protoClient")
		return false
	}
	protoClient := *p_protoClient

	req := PeerChangeRequest{OverlayId: overlayId, PeerId: peerId, DisplayName: displayName, IsLeave: leave}
	r, err := protoClient.PeerChange(context.Background(), &req)
	if err != nil {
		util.Println(util.ERROR, "Could not send peer change to API: ", err)
	}

	if !r.Ack {
		util.Println(util.ERROR, "Peer Change Response is FALSE")
	}

	return r.Ack
}

func (client *PbClient) SessionChange(change *util.SessionChangeInfo) bool {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		util.Println(util.ERROR, "Could not get protoClient")
		return false
	}
	protoClient := *p_protoClient

	if change == nil || change.OverlayId == "" {
		util.Println(util.ERROR, "SessionChangeInfo is nil or OverlayId is empty")
		return false
	}

	req := SessionChangeRequest{OverlayId: change.OverlayId}
	if change.Title != nil {
		req.TitleChanged = true
		req.Title = *change.Title
	}
	if change.Description != nil {
		req.DescriptionChanged = true
		req.Description = *change.Description
	}
	if change.OwnerId != nil {
		req.OwnerIdChanged = true
		req.OwnerId = *change.OwnerId
	}
	if change.AccessKey != nil {
		req.AccessKeyChanged = true
		req.AccessKey = *change.AccessKey
	}
	if change.StartDateTime != nil {
		req.StartDateTimeChanged = true
		req.StartDateTime = *change.StartDateTime
	}
	if change.EndDateTime != nil {
		req.EndDateTimeChanged = true
		req.EndDateTime = *change.EndDateTime
	}
	if change.SourceList != nil {
		req.SourceListChanged = true
		req.SourceList = *change.SourceList
	}
	if change.ChannelList != nil {
		req.ChannelSourceChanged = true
		req.ChannelList = make([]*SessionChangeChannel, 0)
		for _, channel := range *change.ChannelList {
			req.ChannelList = append(req.ChannelList, &SessionChangeChannel{
				ChannelId:  channel.ChannelId,
				SourceList: *channel.SourceList,
			})
		}
	}

	r, err := protoClient.SessionChange(context.Background(), &req)
	if err != nil {
		util.Println(util.ERROR, "Could not send session change to API: ", err)
	}

	if !r.Ack {
		util.Println(util.ERROR, "Session Change Response is FALSE")
	}

	return r.Ack
}

func (client *PbClient) Expulsion(overlayId string, peerId string) bool {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		util.Println(util.ERROR, "Could not get protoClient")
		return false
	}
	protoClient := *p_protoClient

	req := ExpulsionRequest{OverlayId: overlayId, PeerId: peerId}
	r, err := protoClient.Expulsion(context.Background(), &req)
	if err != nil {
		util.Println(util.ERROR, "Could not send expulsion to API: ", err)
	}

	if !r.Ack {
		util.Println(util.ERROR, "Expulsion Response is FALSE")
	}

	return r.Ack
}

func (client *PbClient) Data(overlayId string, channelId string, senderId string, dataType string, data *[]byte) bool {
	p_protoClient := client.getProtoClient()
	if p_protoClient == nil {
		util.Println(util.ERROR, "Could not get protoClient")
		return false
	}
	protoClient := *p_protoClient

	req := DataRequest{OverlayId: overlayId, ChannelId: channelId, SendPeerId: senderId, DataType: dataType, Data: *data}
	r, err := protoClient.Data(context.Background(), &req)
	if err != nil {
		util.Println(util.ERROR, "Could not send data to API: ", err)
	}

	if !r.Ack {
		util.Println(util.ERROR, "Data Response is FALSE")
	}

	return r.Ack
}
