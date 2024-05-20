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
import queue
import threading

from .processHandler import ProcessHandler
from .classes import *
from .pb2 import api_pb2
#from .logger import print, printError, printDebug

class Peer:
    peerId = ""
    overlayId = ""
    startDateTime = None
    endDateTime = None

    isConnect = False
    processHandler = None

    streamQueue:queue.Queue = None

    readyEvent:threading.Event = None
    creationEvent:threading.Event = None
    queryEvent:threading.Event = None
    joinEvent:threading.Event = None
    modificationEvent:threading.Event = None
    removalEvent:threading.Event = None
    sendDataEvent:threading.Event = None
    leaveEvent:threading.Event = None
    searchPeerEvent:threading.Event = None

    creationResponse: CreationResponse = None
    queryResponse: QueryResponse = None
    joinResponse: JoinResponse = None
    modificationResponse: Response = None
    removalResponse: Response = None
    sendDataResponse: Response = None
    leaveResponse: Response = None
    searchPeerResponse: SearchPeerResponse = None

    def __init__(self, peerId: str):
        self.peerId = peerId
        self.overlayId = ""
        self.isConnect = False
        self.processHandler = ProcessHandler()

        self.readyEvent = threading.Event()
        self.creationEvent = threading.Event()
        self.queryEvent = threading.Event()
        self.joinEvent = threading.Event()
        self.modificationEvent = threading.Event()
        self.removalEvent = threading.Event()
        self.sendDataEvent = threading.Event()
        self.leaveEvent = threading.Event()
        self.searchPeerEvent = threading.Event()

        self.streamQueue = queue.Queue()

    def PeerStart(self, grpcPort: int, index: str, option: PeerOption) -> bool:
        self.processHandler.SetPeerOption(option)
        if not self.processHandler.PeerStart(self.peerId, grpcPort, index):
            return False
        return True
    
    def GetStreamQueue(self) -> queue.Queue:
        return self.streamQueue
    
    def  PeerStop(self):
        self.processHandler.PeerStop()

    def WaitReady(self) -> bool:
        rslt =  self.readyEvent.wait()
        self.readyEvent.clear()

        return rslt
    
    def SetReady(self):
        self.isConnect = True
        self.readyEvent.set()

    def WaitCreation(self) -> bool:
        rslt = self.creationEvent.wait()
        self.creationEvent.clear()

        return rslt
    
    def WaitQuery(self) -> bool:
        rslt = self.queryEvent.wait()
        self.queryEvent.clear()

        return rslt
    
    def WaitJoin(self) -> bool:
        rslt = self.joinEvent.wait()
        self.joinEvent.clear()

        return rslt
    
    def WaitModification(self) -> bool:
        rslt = self.modificationEvent.wait()
        self.modificationEvent.clear()

        return rslt
    
    def WaitRemoval(self) -> bool:
        rslt = self.removalEvent.wait()
        self.removalEvent.clear()

        return rslt
    
    def WaitSendData(self) -> bool:
        rslt = self.sendDataEvent.wait()
        self.sendDataEvent.clear()

        return rslt
    
    def WaitLeave(self) -> bool:
        rslt = self.leaveEvent.wait()
        self.leaveEvent.clear()

        return rslt

    def WaitSearchPeer(self) -> bool:
        rslt = self.searchPeerEvent.wait()
        self.searchPeerEvent.clear()

        return rslt
    
    def IsConnect(self) -> bool:
        return self.isConnect
    
    def GetCreationRequest(self, req: CreationRequest) -> api_pb2.CreationRequest:
        greq = api_pb2.CreationRequest(
            title = req.title,
            ownerId = req.ownerId,
            adminKey = req.adminKey,
            channelList = []
        )

        if req.description: greq.description = req.description
        if req.accessKey: greq.accessKey = req.accessKey
        if req.startDateTime: greq.startDateTime = req.startDateTime
        if req.endDateTime: greq.endDateTime = req.endDateTime
        if req.peerList: greq.peerList.extend(req.peerList)
        if req.blockList:
            greq.blockList.extend(req.blockList)
            greq.useBlockList = True
        else: greq.useBlockList = False
        if req.sourceList: 
            greq.sourceList.extend(req.sourceList)
            greq.useSourceList = True
        else: greq.useSourceList = False

        for channel in req.channelList:
            if channel is None: continue
            
            gchannel = api_pb2.Channel(
                channelId = channel.channelId,
                channelType = channel.channelType.value,
                useSourceList = False
            )

            if channel.sourceList:
                gchannel.sourceList.extend(channel.sourceList)
                gchannel.useSourceList = True

            if channel.channelType == ChannelType.ServiceControl:
                gchannel.isServiceChannel = True
            elif channel.channelType == ChannelType.FeatureBasedVideo:
                gchannel.videoFeature.target = channel.target.value
                gchannel.videoFeature.mode = channel.mode.value
                if channel.resolution: gchannel.videoFeature.resolution = channel.resolution
                if channel.framerate: gchannel.videoFeature.frameRate = channel.framerate
                if channel.keypointsType: gchannel.videoFeature.keyPointsType = channel.keypointsType
                if channel.modelUri: gchannel.videoFeature.modelUri = channel.modelUri
                if channel.hash: gchannel.videoFeature.hash = channel.hash
                if channel.dimension: gchannel.videoFeature.dimension = channel.dimension

            elif channel.channelType == ChannelType.Audio:
                gchannel.audio.codec = channel.codec.value
                gchannel.audio.sampleRate = channel.sampleRate.value
                gchannel.audio.bitrate = channel.bitrate.value
                gchannel.audio.channelType = channel.audioChannelType.value
            elif channel.channelType == ChannelType.Text:
                gchannel.text.format = channel.format.value
                gchannel.text.encoding = channel.encoding.value

            greq.channelList.append(gchannel)

        return greq

    def SetCreationResponse(self, res: api_pb2.CreationResponse) -> None:
        if res:
            self.overlayId = res.overlayId
            self.startDateTime = res.startDateTime
            self.endDateTime = res.endDateTime

            self.creationResponse = CreationResponse(ResponseCode(res.rspCode), res.overlayId, res.startDateTime, res.endDateTime)  

        self.creationEvent.set()

    def GetCreationResponse(self) -> CreationResponse:
        return self.creationResponse
    
    def SetQueryResponse(self, res: api_pb2.QueryResponse) -> None:
        self.queryResponse = QueryResponse(ResponseCode(res.rspCode))
        self.queryResponse.overlay = []
        if res.rspCode == ResponseCode.Success.value and res.overlayList:
            for goverlay in res.overlayList:
                overlay = Overlay()
                overlay.overlayId = goverlay.overlayId
                overlay.title = goverlay.title
                overlay.description = goverlay.description
                overlay.ownerId = goverlay.ownerId
                overlay.closed = OverlayClosed(goverlay.closed)
                overlay.startDateTime = goverlay.startDateTime
                overlay.endDateTime = goverlay.endDateTime

                if goverlay.channelList and len(goverlay.channelList) > 0:
                    overlay.channelList = []

                    for gchannel in goverlay.channelList:
                        channel = None

                        if gchannel.channelType == ChannelType.ServiceControl.value:
                            channel = ChannelServiceControl()
                            channel.channelId = gchannel.channelId
                        elif gchannel.channelType == ChannelType.FeatureBasedVideo.value:
                            channel = ChannelFeatureBasedVideo()
                            channel.target = FeatureBasedVideoTarget(gchannel.videoFeature.target)
                            channel.mode = FeatureBasedVideoMode(gchannel.videoFeature.mode)
                            channel.resolution = gchannel.videoFeature.resolution
                            channel.framerate = gchannel.videoFeature.frameRate
                            channel.keypointsType = gchannel.videoFeature.keyPointsType
                            channel.modelUri = gchannel.videoFeature.modelUri
                            channel.hash = gchannel.videoFeature.hash
                            channel.dimension = gchannel.videoFeature.dimension
                        elif gchannel.channelType == ChannelType.Audio.value:
                            channel = ChannelAudio()
                            channel.codec = AudioCodec(gchannel.audio.codec)
                            channel.sampleRate = AudioSampleRate(gchannel.audio.sampleRate)
                            channel.bitrate = AudioBitrate(gchannel.audio.bitrate)
                            channel.audioChannelType = AudioChannelType(gchannel.audio.channelType)
                        elif gchannel.channelType == ChannelType.Text.value:
                            channel = ChannelText()
                            channel.format = TextFormat(gchannel.text.format)
                            channel.encoding = TextEncoding(gchannel.text.encoding)

                        overlay.channelList.append(channel)

                self.queryResponse.overlay.append(overlay)

        self.queryEvent.set()

    def GetQueryResponse(self) -> QueryResponse:
        return self.queryResponse
    
    def SetJoinResponse(self, res: api_pb2.JoinResponse) -> None:
        self.joinResponse = JoinResponse(ResponseCode(res.rspCode))
        
        if res.rspCode == ResponseCode.Success.value:
            self.joinResponse.startDateTime = res.startDateTime
            self.joinResponse.endDateTime = res.endDateTime
            self.joinResponse.title = res.title
            self.joinResponse.description = res.description
            self.joinResponse.sourceList = res.sourceList
            
            if res.channelList and len(res.channelList) > 0:
                self.joinResponse.channelList = []
                for channel in res.channelList:
                    
                    resChannel = None

                    if channel.channelType == ChannelType.ServiceControl.value:
                        resChannel = ChannelServiceControl(channel.channelId)

                    elif channel.channelType == ChannelType.FeatureBasedVideo.value:
                        resChannel = ChannelFeatureBasedVideo(channel.channelId)
                        resChannel.target = FeatureBasedVideoTarget(channel.videoFeature.target)
                        resChannel.mode = FeatureBasedVideoMode(channel.videoFeature.mode)
                        resChannel.resolution = channel.videoFeature.resolution
                        resChannel.framerate = channel.videoFeature.frameRate
                        resChannel.keypointsType = channel.videoFeature.keyPointsType
                        resChannel.modelUri = channel.videoFeature.modelUri
                        resChannel.hash = channel.videoFeature.hash
                        resChannel.dimension = channel.videoFeature.dimension
                    elif channel.channelType == ChannelType.Audio.value:
                        resChannel = ChannelAudio(channel.channelId)
                        resChannel.codec = AudioCodec(channel.audio.codec)
                        resChannel.sampleRate = AudioSampleRate(channel.audio.sampleRate)
                        resChannel.bitrate = AudioBitrate(channel.audio.bitrate)
                        resChannel.audioChannelType = AudioChannelType(channel.audio.channelType)
                    elif channel.channelType == ChannelType.Text.value:
                        resChannel = ChannelText(channel.channelId)
                        resChannel.format = TextFormat(channel.text.format)
                        resChannel.encoding = TextEncoding(channel.text.encoding)
                    
                    if resChannel:
                        if channel.useSourceList:
                            resChannel.sourceList = channel.sourceList

                        self.joinResponse.channelList.append(resChannel)
        
        self.joinEvent.set()

    def GetJoinResponse(self) -> JoinResponse:
        return self.joinResponse
    
    def SetModificationResponse(self, res: api_pb2.ModificationResponse) -> None:
        self.modificationResponse = Response(ResponseCode(res.rspCode))
        self.modificationEvent.set()
    
    def GetModificationResponse(self) -> Response:
        return self.modificationResponse
    
    def SetRemovalResponse(self, res: api_pb2.RemovalResponse) -> None:
        self.removalResponse = Response(ResponseCode(res.rspCode))
        self.removalEvent.set()

    def GetRemovalResponse(self) -> Response:
        return self.removalResponse
    
    def SetSendDataResponse(self, res: api_pb2.SendDataResponse) -> None:
        self.sendDataResponse = Response(ResponseCode(res.rspCode))
        self.sendDataEvent.set()

    def GetSendDataResponse(self) -> Response:
        return self.sendDataResponse
    
    def SetLeaveResponse(self, res: api_pb2.LeaveResponse) -> None:
        self.leaveResponse = Response(ResponseCode(res.rspCode))
        self.leaveEvent.set()

    def GetLeaveResponse(self) -> Response:
        return self.leaveResponse
    
    def SetSearchPeerResponse(self, res: api_pb2.SearchPeerResponse) -> None:
        self.searchPeerResponse = SearchPeerResponse(ResponseCode(res.rspCode))
        if res.rspCode == ResponseCode.Success.value:
            self.searchPeerResponse.peerList = []
            for peer in res.peerList:
                self.searchPeerResponse.peerList.append(PeerInfo(peer.peerId, peer.displayName))

        self.searchPeerEvent.set()

    def GetSearchPeerResponse(self) -> SearchPeerResponse:
        return self.searchPeerResponse