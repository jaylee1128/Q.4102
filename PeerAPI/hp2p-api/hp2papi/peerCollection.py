#
# The MIT License
#
# Copyright (c) 2022 ETRI
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#
from collections import namedtuple
import threading
import time
from typing import Callable, List
from .classes import *
from .peer import Peer
from .gRpcServicer import Hp2pServicer
from .pb2 import api_pb2
from .logger import print, printError, printDebug

class PeerCollection:
    __readyPeerIdx = 1
    __isStop = False
    __servicer = None
    __notiThread = None
    __peerOption = None

    __NotiCallback = None

    __peers = dict()

    def __init__(self):
        printDebug("PeerCollection init.")
        self.__readyPeerIdx = 1
        self.__servicer = Hp2pServicer()
        self.__isStop = False
        self.__notiThread = None
        self.__peerOption = PeerOption()
        self.__peerOption.HeartbeatInterval = 10
        self.__peerOption.HeartbeatTimeout = 15
        self.__peerOption.MNCache = 0
        self.__peerOption.MDCache = 0 # minute
        self.__peerOption.RecoveryBy = "push" # "push" or "pull"
        self.__peerOption.RateControlQuantity = 0 # count
        self.__peerOption.RateControlBitrate = 0 # kbps
        self.__NotiCallback = None

    def SetGrpcPort(self, port: int) -> None:
        self.__servicer.SetGrpcPort(port)

    def GrpcStart(self) -> bool:
        if not self.__servicer.GrpcStart(self.__peers):
            return False
        
        self.__notiThread = threading.Thread(target=self.__notiWorker)
        self.__notiThread.start()

        return True

    def SetNotiCallback(self, callback: Callable[[Notification], None]) -> None:
        self.__NotiCallback = callback
        
    def __notiWorker(self) -> None:
        queue = self.__servicer.GetNotiQueue()
        while not self.__isStop:
            noti = queue.get()

            if not noti:
                queue.task_done()
                break

            printDebug("notiWorker: ")
            printDebug(noti)

            NotiObj = namedtuple("NotiObj", noti.keys())
            notiobj = NotiObj(**noti)

            threading.Thread(target=self.__notiProcess, args=(notiobj,)).start()
            
            queue.task_done()

    def __notiProcess(self, notiobj: any) -> None:
        if notiobj.type == GrpcMsgType.Ready:
            self.RecvReady(notiobj.index)
        elif notiobj.type == GrpcMsgType.Creation:
            self.RecvCreation(notiobj.data)
        elif notiobj.type == GrpcMsgType.Join:
            self.RecvJoin(notiobj.data)
        elif notiobj.type == GrpcMsgType.Query:
            self.RecvQuery(notiobj.data)
        elif notiobj.type == GrpcMsgType.Modification:
            self.RecvModification(notiobj.data)
        elif notiobj.type == GrpcMsgType.Leave:
            self.RecvLeave(notiobj.data)
        elif notiobj.type == GrpcMsgType.Removal:
            self.RecvRemoval(notiobj.data)
        elif notiobj.type == GrpcMsgType.SendData:
            self.RecvSendData(notiobj.data)
        elif notiobj.type == GrpcMsgType.SearchPeer:
            self.RecvSearchPeer(notiobj.data)
        elif notiobj.type == GrpcMsgType.SessionTermination:
            self.RecvSessionTermination(notiobj.data)
        elif notiobj.type == GrpcMsgType.PeerChange:
            self.RecvPeerChange(notiobj.data)
        elif notiobj.type == GrpcMsgType.SessionChange:
            self.RecvSessionChange(notiobj.data)
        elif notiobj.type == GrpcMsgType.Expulsion:
            self.RecvExpulsion(notiobj.data)
        elif notiobj.type == GrpcMsgType.Data:
            self.RecvData(notiobj.data)
        else:
            printError("Unknown noti type: " + str(notiobj.type))

    def NewPeerStart(self, peerId) -> str:
        peer = Peer(peerId)
        tempId = str(self.__readyPeerIdx)
        self.__readyPeerIdx += 1

        self.__peers[tempId] = peer

        if not peer.PeerStart(self.__servicer.GetGrpcPort(), tempId, self.__peerOption):
            printError("PeerCollection NewPeerStart fail.")
            del self.__peers[tempId]
            return None
        
        printDebug("PeerCollection Connect")
        
        peer.WaitReady()
        
        if not peer.IsConnect():
            printError("PeerCollection peer connect fail.")
            peer.PeerStop()
            del self.__peers[tempId]
            return None

        time.sleep(1)

        return tempId
    
    def getPeer(self, pid: str) -> Peer:
        if pid in self.__peers:
            return self.__peers[pid]
        else:
            return None

    def SetPeerOption(self, peerOption: PeerOption) -> None:
        self.__peerOption = peerOption
    
    def RecvReady(self, index: str):
        peer = self.getPeer(index)

        if not peer: return False

        #if not peer.Ready(port): return False
        peer.SetReady()
        return True
    
    def RecvSessionTermination(self, req: api_pb2.SessionTerminateRequest):
        peer = self.getPeer(req.overlayId)

        if not peer: return False

        self.__NotiCallback(SessionTerminationNotification(overlayId=req.overlayId, peerId=req.peerId))

        peer.PeerStop()

        try:
            del self.__peers[req.overlayId]
        except Exception as e:
            pass

        return True
    
    def RecvPeerChange(self, req: api_pb2.PeerChangeRequest):
        peer = self.getPeer(req.overlayId)

        if not peer: return False

        self.__NotiCallback(PeerChangeNotification(overlayId=req.overlayId, peerId=peer.peerId, changePeerId=req.peerId, displayName=req.displayName, leave=req.isLeave))

        return True
    
    def RecvSessionChange(self, req: api_pb2.SessionChangeRequest):
        peer = self.getPeer(req.overlayId)

        if not peer: return False

        sessionChange = SessionChangeNotification()
        sessionChange.overlayId = req.overlayId
        sessionChange.peerId = peer.peerId
        if req.titleChanged:
            sessionChange.title = req.title
        if req.descriptionChanged:
            sessionChange.description = req.description
        if req.ownerIdChanged:
            sessionChange.ownerId = req.ownerId
        if req.accessKeyChanged:
            sessionChange.accessKey = req.accessKey
        if req.startDateTimeChanged:
            sessionChange.startDateTime = req.startDateTime
        if req.endDateTimeChanged:
            sessionChange.endDateTime = req.endDateTime
        if req.sourceListChanged:
            sessionChange.sourceList = req.sourceList
        if req.channelSourceChanged:
            sessionChange.channelList = list()
            for channel in req.channelList:
                channelObj = Channel()
                channelObj.channelId = channel.channelId
                channelObj.sourceList = list()
                channelObj.sourceList.append(channel.sourceList)
                sessionChange.channelList.append(channelObj)

        self.__NotiCallback(sessionChange)

        return True
        
    def RecvExpulsion(self, req: api_pb2.ExpulsionRequest):
        peer = self.getPeer(req.overlayId)

        if not peer: return False

        self.__NotiCallback(ExpulsionNotification(overlayId=req.overlayId, peerId=req.peerId))

        peer.PeerStop()
        try:
            del self.__peers[req.overlayId]
        except Exception as e:
            printError("PeerCollection RecvSessionTermination error. " + str(e))

        return True
    
    def RecvData(self, req: api_pb2.DataRequest):
        peer = self.getPeer(req.overlayId)

        if not peer: return False

        self.__NotiCallback(DataNotification(overlayId=req.overlayId, channelId=req.channelId, sendPeerId=req.sendPeerId, dataType=req.dataType, peerId=peer.peerId, data=req.data, dataLength=len(req.data)))

        return True

    def Creation(self, pid: str, req: CreationRequest) -> CreationResponse:
        printDebug("PeerCollection Creation")

        peer = self.getPeer(pid)
        if not peer:
            printError("PeerCollection Creation fail. peer not found.")
            return CreationResponse(code=ResponseCode.Fail)
        
        request:api_pb2.CreationRequest = peer.GetCreationRequest(req)
        reqWithId = api_pb2.Request(
            id=pid,
            creation=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.Creation, "data": reqWithId})

        peer.WaitCreation()

        resp = peer.GetCreationResponse()
        if not resp:
            printError("PeerCollection Creation fail. Unknown error.")
            return CreationResponse(code=ResponseCode.Fail)
        
        if resp.code == ResponseCode.Success:
            del self.__peers[pid]
            self.__peers[resp.overlayId] = peer

        self.__servicer.SetPeerIndexChange()

        return resp
    
    def RecvCreation(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvCreation")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvCreation fail. peer not found.")
            return
        
        peer.SetCreationResponse(req.creation)

    def Query(self, pid: str, req: QueryRequest) -> QueryResponse:
        printDebug("PeerCollection Query")

        peer = self.getPeer(pid)
        if not peer:
            printError("PeerCollection Query fail. peer not found.")
            return QueryResponse(code=ResponseCode.Fail)
        
        request:api_pb2.QueryRequest = api_pb2.QueryRequest(
            overlayId=req.overlayId, 
            title=req.title,
            description=req.description
        )
        reqWithId = api_pb2.Request(
            id=pid,
            query=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.Query, "data": reqWithId})

        peer.WaitQuery()

        resp = peer.GetQueryResponse()

        del self.__peers[pid]
        peer.PeerStop()

        return resp
    
    def RecvQuery(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvQuery")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvQuery fail. peer not found.")
            return
        
        peer.SetQueryResponse(req.query)

    def Join(self, req: JoinRequest) -> JoinResponse:
        printDebug("PeerCollection Join")

        pid = req.overlayId
        peer = self.getPeer(pid)
        if not peer:
            pid = self.NewPeerStart(req.peerId)
            if not pid:
                return JoinResponse(code=ResponseCode.Fail)
            
            peer = self.getPeer(pid)

        request:api_pb2.JoinRequest = api_pb2.JoinRequest(
            overlayId=req.overlayId,
            accessKey=req.accessKey,
            peerId=req.peerId,
            displayName=req.displayName,
            privateKeyPath=req.privateKeyPath,
        )

        reqWithId = api_pb2.Request(
            id=pid,
            join=request
        )
        
        peer.GetStreamQueue().put({"type": GrpcMsgType.Join, "data": reqWithId})

        peer.WaitJoin()

        resp = peer.GetJoinResponse()

        if resp.code == ResponseCode.Success:
            if pid != req.overlayId:
                del self.__peers[pid]
                self.__peers[req.overlayId] = peer
        else:
            peer.PeerStop()
            del self.__peers[pid]

        self.__servicer.SetPeerIndexChange()

        return resp
    
    def RecvJoin(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvJoin")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvJoin fail. peer not found.")
            return
        
        peer.SetJoinResponse(req.join)

    def Modification(self, req: ModificationRequest) -> Response:
        printDebug("PeerCollection Modification")

        pid = req.overlayId
        peer = self.getPeer(pid)
        if not peer:
            pid = self.NewPeerStart(req.ownerId)
            if not pid:
                return Response(code=ResponseCode.Fail)
            
            peer = self.getPeer(pid)

        request:api_pb2.ModificationRequest = api_pb2.ModificationRequest(
            overlayId=req.overlayId,
            ownerId=req.ownerId,
            adminKey=req.adminKey,
        )

        if req.title:
            request.modificationTitle = True
            request.title = req.title
        else:
            request.modificationTitle = False
        if req.description:
            request.modificationDescription = True
            request.description = req.description
        else:
            request.modificationDescription = False
        if req.accessKey is not None:
            request.modificationAccessKey = True
            request.accessKey = req.accessKey
        else:
            request.modificationAccessKey = False
        if req.startDateTime:
            request.modificationStartDateTime = True
            request.startDateTime = req.startDateTime
        else:
            request.modificationStartDateTime = False
        if req.endDateTime:
            request.modificationEndDateTime = True
            request.endDateTime = req.endDateTime
        else:
            request.modificationEndDateTime = False
        if req.newAdminKey:
            request.modificationAdminKey = True
            request.newAdminKey = req.newAdminKey
        else:
            request.modificationAdminKey = False
        if req.newOwnerId:
            request.modificationOwnerId = True
            request.newOwnerId = req.newOwnerId
        else:
            request.modificationOwnerId = False
        if req.peerList is not None:
            request.modificationPeerList = True
            request.peerList.extend(req.peerList)
        else:
            request.modificationPeerList = False
        if req.sourceList:
            request.modificationSourceList = True
            request.sourceList.extend(req.sourceList)
        else:
            request.modificationSourceList = False
        if req.blockList:
            request.modificationBlockList = True
            request.blockList.extend(req.blockList)
        else:
            request.modificationBlockList = False
        if req.channelList:
            request.modificationChannelList = True
            for channel in req.channelList:
                pbchannel = api_pb2.Channel(channelId=channel.channelId)
                if channel.sourceList is not None:
                    pbchannel.useSourceList = True
                    pbchannel.sourceList.extend(channel.sourceList)
                else:
                    pbchannel.useSourceList = False
                request.channelList.append(pbchannel)
        else:
            request.modificationChannelList = False

        reqWithId = api_pb2.Request(
            id=pid,
            modification=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.Modification, "data": reqWithId})

        peer.WaitModification()

        resp = peer.GetModificationResponse()

        return resp           

    def RecvModification(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvModification")
        #printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvModification fail. peer not found.")
            return
        
        peer.SetModificationResponse(req.modification)

    def Removal(self, req: RemovalRequest) -> Response:
        printDebug("PeerCollection Removal")

        pid = req.overlayId
        peer = self.getPeer(pid)
        if not peer:
            pid = self.NewPeerStart(req.ownerId)
            if not pid:
                return Response(code=ResponseCode.Fail)
            
            peer = self.getPeer(pid)

        request:api_pb2.RemovalRequest = api_pb2.RemovalRequest(
            overlayId=req.overlayId,
            ownerId=req.ownerId,
            adminKey=req.adminKey,
        )

        reqWithId = api_pb2.Request(
            id=pid,
            removal=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.Removal, "data": reqWithId})

        peer.WaitRemoval()

        resp = peer.GetRemovalResponse()

        if resp.code == ResponseCode.Success:
            peer.PeerStop()
            del self.__peers[pid]

        return resp
    
    def RecvRemoval(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvRemoval")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvRemoval fail. peer not found.")
            return
        
        peer.SetRemovalResponse(req.removal)

    def SendData(self, req: SendDataRequest) -> Response:
        printDebug("PeerCollection SendData")

        pid = req.overlayId
        peer = self.getPeer(pid)
        if not peer:
            return Response(code=ResponseCode.NotFound)

        request:api_pb2.SendDataRequest = api_pb2.SendDataRequest(
            dataType=req.dataType.value,
            channelId=req.channelId,
            data=req.data,
        )

        reqWithId = api_pb2.Request(
            id=pid,
            sendData=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.SendData, "data": reqWithId})

        peer.WaitSendData()

        resp = peer.GetSendDataResponse()

        return resp
    
    def RecvSendData(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvSendData")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvSendData fail. peer not found.")
            return
        
        peer.SetSendDataResponse(req.sendData)
    
    def RecvRemoval(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvRemoval")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvRemoval fail. peer not found.")
            return
        
        peer.SetRemovalResponse(req.removal)
    
    def Leave(self, req: LeaveRequest) -> Response:
        printDebug("PeerCollection Leave")

        pid = req.overlayId
        peer = self.getPeer(pid)
        if not peer:
            printError("PeerCollection Leave fail. peer not found.")
            return Response(code=ResponseCode.Fail)

        request:api_pb2.Leave = api_pb2.LeaveRequest(
            overlayId=req.overlayId,
            peerId=req.peerId,
            accessKey=req.accessKey,
        )

        reqWithId = api_pb2.Request(
            id=pid,
            leave=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.Leave, "data": reqWithId})

        peer.WaitLeave()

        resp = peer.GetLeaveResponse()

        if resp.code == ResponseCode.Success:
            peer.PeerStop()
            del self.__peers[pid]

        return resp
    
    def RecvLeave(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvLeave")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvLeave fail. peer not found.")
            return
        
        peer.SetLeaveResponse(req.leave)

    def SearchPeer(self, req: SearchPeerRequest) -> SearchPeerResponse:
        printDebug("PeerCollection SearchPeer")

        pid = req.overlayId
        peer = self.getPeer(pid)
        if not peer:
            printError("PeerCollection SearchPeer fail. peer not found.")
            return Response(code=ResponseCode.NotJoined)

        request:api_pb2.Leave = api_pb2.SearchPeerRequest(
            overlayId=req.overlayId,
        )

        reqWithId = api_pb2.Request(
            id=pid,
            searchPeer=request
        )

        peer.GetStreamQueue().put({"type": GrpcMsgType.SearchPeer, "data": reqWithId})

        peer.WaitSearchPeer()

        resp = peer.GetSearchPeerResponse()

        return resp
    
    def RecvSearchPeer(self, req: api_pb2.Response) -> None:
        printDebug("PeerCollection RecvSearchPeer")
        printDebug(req)

        peer = self.getPeer(req.id)
        if not peer:
            printError("PeerCollection RecvSearchPeer fail. peer not found.")
            return
        
        peer.SetSearchPeerResponse(req.searchPeer)

    def Cleanup(self):
        self.__isStop = True
        self.__servicer.GetNotiQueue().put(None)
        try:
            self.__notiThread.join()
        except Exception as e:
            printError("PeerCollection Cleanup error. " + str(e))
        self.__servicer.GrpcStop()

        for peer in self.__peers.values():
            peer.PeerStop()

        print("PeerCollection Cleanup")
        pass