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
from abc import ABCMeta, abstractmethod
from dataclasses import dataclass, field
from enum import Enum, unique
from typing import List

__all__ = [
    "PeerOption",
    "ResponseCode",
    "GrpcMsgType",
    "DataType",
    "ChannelType",
    "FeatureBasedVideoMode",
    "FeatureBasedVideoTarget",
    "AudioCodec",
    "AudioSampleRate",
    "AudioBitrate",
    "AudioChannelType",
    "TextFormat",
    "TextEncoding",
    "Response",
    "OverlayClosed",
    "Overlay",
    "QueryRequest",
    "QueryResponse",
    "ChannelServiceControl",
    "ChannelFeatureBasedVideo",
    "ChannelAudio",
    "ChannelText",
    "Request",
    "CreationRequest",
    "CreationResponse",
    "ModificationRequest",
    "JoinRequest",
    "JoinResponse",
    "SearchPeerRequest",
    "PeerInfo",
    "SearchPeerResponse",
    "SendDataRequest",
    "LeaveRequest",
    "RemovalRequest",
    "Notification",
    "NotificationType",
    "SessionChangeNotification",
    "SessionTerminationNotification",
    "PeerChangeNotification",
    "ExpulsionNotification",
    "DataNotification",
    "Channel"
]

class PeerOption():
    HeartbeatInterval = 10
    HeartbeatTimeout = 15
    MNCache = 0
    MDCache = 0
    RecoveryBy = "push" # "push" or "pull"
    RateControlQuantity = 0 # count
    RateControlBitrate = 0 # kbps


@unique
class ResponseCode(Enum):
    Success = 200
    WrongRequest = 400
    NotJoined = 401
    AuthenticationError = 403
    NotFound = 404
    Fail = 500

@unique
class GrpcMsgType(int, Enum):
    Ready = 1
    Creation = 2
    Query = 3
    Join = 4
    Modification = 5
    Leave = 6
    Removal = 7
    SearchPeer = 8
    Heartbeat = 9
    SessionTermination = 10
    PeerChange = 11
    SessionChange = 12
    Expulsion = 13
    SendData = 14
    Data = 15
    NotMine = 16

@unique
class DataType(Enum):
    DataCache = "data-cache"
    Data = "data"

@unique
class ChannelType(Enum):
    ServiceControl = "control"
    FeatureBasedVideo = "video/feature"
    Audio = "audio"
    Text = "text"

@unique
class FeatureBasedVideoTarget(Enum):
    Face = "face"
    Body = "body"
    Hands = "hands"
    Pose = "pose"

@unique
class FeatureBasedVideoMode(Enum):
    KeypointsDescriptionMode = "KDM"
    SharedNeuralNetworkMode = "SNNM"

@unique
class AudioCodec(Enum):
    AAC = "AAC"
    ALAC = "ALAC"
    AMR = "AMR"
    FLAC = "FLAC"
    G711 = "G711"
    G722 = "G722"
    MP3 = "MP3"
    Opus = "Opus"
    Vorbis = "Vorbis"

@unique
class AudioSampleRate(Enum):
    Is44100 = "44100"
    Is48000 = "48000"

@unique
class AudioBitrate(Enum):
    Is128kbps = "128kbps"
    Is192kbps = "192kbps"
    Is256kbps = "256kbps"
    Is320kbps = "320kbps"

@unique
class AudioChannelType(Enum):
    Mono = "mono"
    Stereo = "stereo"

@unique
class TextFormat(Enum):
    Plain = "text/plain"
    Json = "application/json"

@unique
class TextEncoding(Enum):
    UTF8 = "UTF-8"

@unique
class OverlayClosed(Enum):
    Open = 0
    Closed = 1

@unique
class NotificationType(Enum):
    SessionChangeNotification = 0
    SessionTerminationNotification = 1
    PeerChangeNotification = 2
    ExpulsionNotification = 3
    DataNotification = 4

def _check(var) -> bool:
    if not var:
        return False
    
    return True

@dataclass
class Request(metaclass=ABCMeta):
    @abstractmethod
    def mandatoryCheck(self) -> bool:
        pass

@dataclass
class Response(metaclass=ABCMeta):
    code: ResponseCode = None

@dataclass
class Channel:
    channelId: str = None
    channelType: ChannelType = None
    sourceList: List[str] = None

    def mandatoryCheck(self) -> bool:
        if not _check(self.channelId) or len(self.channelId) <= 0:
            return False
        if not _check(self.channelType):
            return False
        
        return True

@dataclass
class Overlay:
    overlayId: str = None
    title: str = None
    description: str = None
    ownerId: str = None
    closed: OverlayClosed = OverlayClosed.Open
    startDateTime: str = None #YYYYMMDDHHmmSS
    endDateTime: str = None #YYYYMMDDHHmmSS
    channelList: List[Channel] = None

@dataclass
class QueryResponse(Response):
    overlay: List[Overlay] = None

@dataclass
class QueryRequest(Request):
    """
    서비스 세션 목록 조회 요청 Class
        optional:
            overlayId, title, description
    """
    overlayId: str = None
    title: str = None
    description: str = None

    def mandatoryCheck(self) -> bool:
        return True
@dataclass
class ChannelServiceControl(Channel):
    """
    제어 신호 메시지를 위한 채널
    """
    def __init__(self, channelId: str = None):
        self.channelId = channelId
        self.channelType = ChannelType.ServiceControl
    
    def mandatoryCheck(self) -> bool:
        if not super().mandatoryCheck():
            return False
            
        return True

@dataclass
class ChannelFeatureBasedVideo(Channel):
    """
    feature를 위한 채널
    """
    def __init__(
        self,
        channelId: str = None,
        target: FeatureBasedVideoTarget = None,
        mode: FeatureBasedVideoMode = None,
        resolution: str = None,
        framerate: str = None,
        keypointsType: str = None,
        modelUri: str = None,
        hash: str = None,
        dimension: str = None,
    ):
        self.channelId = channelId
        self.target = target
        self.channelType = ChannelType.FeatureBasedVideo
        self.mode = mode
        self.resolution = resolution
        self.framerate = framerate
        self.keypointsType = keypointsType
        self.modelUri = modelUri
        self.hash = hash
        self.dimension = dimension

    target: FeatureBasedVideoTarget = None
    mode: FeatureBasedVideoMode = None
    resolution: str = None
    framerate: str = None
    keypointsType: str = None
    modelUri: str = None
    hash: str = None
    dimension: str = None

    def mandatoryCheck(self) -> bool:
        if not super().mandatoryCheck():
            return False
        
        if not _check(self.target) or self.target not in FeatureBasedVideoTarget:
            return False

        if not _check(self.mode) or self.mode not in FeatureBasedVideoMode:
            return False
        
        if self.mode is FeatureBasedVideoMode.KeypointsDescriptionMode:
            if not _check(self.resolution):
                return False
            if not _check(self.framerate):
                return False
            if not _check(self.keypointsType):
                return False
            
        elif self.mode is FeatureBasedVideoMode.SharedNeuralNetworkMode:
            if not _check(self.modelUri):
                return False
            if not _check(self.hash):
                return False
            if not _check(self.dimension):
                return False
        
        return True

@dataclass
class ChannelAudio(Channel):
    """
    Audio를 위한 채널
    """
    def __init__(
        self,
        channelId: str = None,
        codec: AudioCodec = None,
        sampleRate: AudioSampleRate = None,
        bitrate: AudioBitrate = None,
        audioChannelType: AudioChannelType = None
    ):
        self.channelId = channelId
        self.channelType = ChannelType.Audio
        self.codec = codec
        self.sampleRate = sampleRate
        self.bitrate = bitrate
        self.audioChannelType = audioChannelType

    codec: AudioCodec = None
    sampleRate: AudioSampleRate = None
    bitrate: AudioBitrate = None
    audioChannelType: AudioChannelType = None

    def mandatoryCheck(self) -> bool:
        if not super().mandatoryCheck():
            return False
        
        if not _check(self.codec):
            return False
        if not _check(self.sampleRate):
            return False
        if not _check(self.bitrate):
            return False
        if not _check(self.audioChannelType):
            return False
        
        return True

@dataclass
class ChannelText(Channel):
    """
    text를 위한 채널
    """
    def __init__(self, channelId: str = None, format: TextFormat = TextFormat.Plain, encoding: TextEncoding = TextEncoding.UTF8):
        self.channelId = channelId
        self.channelType = ChannelType.Text
        self.format = format
        self.encoding = encoding

    format: TextFormat = None
    encoding: TextEncoding = None
    
    def mandatoryCheck(self) -> bool:
        if not super().mandatoryCheck():
            return False
        
        if not _check(self.format):
            return False
        
        if not _check(self.encoding):
            return False
        
        return True

@dataclass
class CreationRequest(Request):
    """
    신규 서비스 생성 요청 Class
        mandatory:
            title, ownerId, adminKey, channelList
    """
    title: str = None
    description: str = None
    ownerId: str = None
    adminKey: str = None
    accessKey: str = None
    peerList: List[str] = None
    blockList: List[str] = None
    startDateTime: str = None #YYYYMMDDHHmmSS
    endDateTime: str = None #YYYYMMDDHHmmSS
    sourceList: List[str] = None
    channelList: List[Channel] = None

    def mandatoryCheck(self) -> bool:
        if not _check(self.title):
            return False
        if not _check(self.ownerId):
            return False
        if not _check(self.adminKey):
            return False
        if not _check(self.channelList) or len(self.channelList) <= 0:
            return False
        for channel in self.channelList:
            if channel is not None and not channel.mandatoryCheck():
                return False
        
        return True

@dataclass
class CreationResponse(Response):
    overlayId: str = None
    startDateTime: str = None
    endDateTime: str = None

@dataclass
class ModificationRequest(Request):
    """
    서비스 세션 변경 요청 Class
        mandatory:
            overlayId, ownerId, adminKey
            3개의 mandatory 는 변경할 서비스 세션을 식별하기 위한 키값으로 
            기존 서비스 세션의 값을 그대로 넣어야 하며 변경이 불가하다.

        optional:
            나머지 다른 값들은 변경이 필요한 값만 변경할 값으로 넣는다.
            넣지 않으면(None이면) 변경하지 않는다.

            ** channelList:
                channel의 channelId는 변경할 수 없으며 변경할 channel의 키값이다.
                sourceList만 변경 가능하다.
    """
    overlayId: str = None
    ownerId: str = None
    adminKey: str = None
    title: str = None
    description: str = None
    newOwnerId: str = None
    newAdminKey: str = None
    startDateTime: str = None #YYYYMMDDHHmmSS
    endDateTime: str = None #YYYYMMDDHHmmSS
    accessKey: str = None
    peerList: List[str] = None
    blockList: List[str] = None
    sourceList: List[str] = None
    channelList: List[Channel] = None

    def mandatoryCheck(self) -> bool:
        if not _check(self.overlayId):
            return False
        if not _check(self.ownerId):
            return False
        if not _check(self.adminKey):
            return False
        
        return True

@dataclass
class JoinRequest(Request):
    """
    서비스 세션 참가 요청 Class
        mandatory:
            overlayId, peerId, displayName, privateKeyPath
            
        optional:
            참가하려는 서비스 세션에 accessKey가 설정된 경우 accessKey 필수
    """
    overlayId: str = None
    accessKey: str = None
    peerId: str = None
    displayName: str = None
    privateKeyPath: str = None
    
    def mandatoryCheck(self) -> bool:
        if not _check(self.overlayId):
            return False
        if not _check(self.peerId):
            return False
        if not _check(self.displayName):
            return False
        if not _check(self.privateKeyPath):
            return False
        
        return True

@dataclass
class JoinResponse(Response):
    title: str = None
    description: str = None
    startDateTime: str = None #YYYYMMDDHHmmSS
    endDateTime: str = None #YYYYMMDDHHmmSS
    sourceList: List[str] = None
    channelList: List[Channel] = None

@dataclass
class PeerInfo:
    """
    Peer 정보 Class
    """
    peerId: str = None
    displayName: str = None

    def __init__(self, peerId: str = None, displayName: str = None):
        self.peerId = peerId
        self.displayName = displayName

@dataclass
class SearchPeerRequest(Request):
    """
    서비스 세션에 참가한 Peer 정보 요청 Class
        mandatory:
            overlayId
    """
    overlayId: str = None

    def __init__(self, overlayId: str = None):
        self.overlayId = overlayId
        
    def mandatoryCheck(self) -> bool:
        if not _check(self.overlayId):
            return False
        
        return True

@dataclass
class SearchPeerResponse(Response):
    peerList: List[PeerInfo] = None

@dataclass
class SendDataRequest(Request):
    """
    Data broadcast 요청 Class
        mandatory:
            dataType, overlayId, dataLength, data
    """
    dataType: DataType = None
    overlayId: str = None
    channelId: str = None
    data: bytes = None

    def __init__(self, dataType: DataType = None, overlayId: str = None, channelId: str = None, data: bytes = None):
        self.dataType = dataType
        self.overlayId = overlayId
        self.channelId = channelId
        self.data = data
        
    def mandatoryCheck(self) -> bool:
        if not _check(self.dataType):
            return False
        if not _check(self.overlayId):
            return False
        if not _check(self.channelId):
            return False
        if not _check(self.data):
            return False
        
        return True

@dataclass
class LeaveRequest(Request):
    """
    서비스 세션 탈퇴 요청 Class
        mandatory:
            overlayId, peerId
        optional:
            accessKey - 설정된 경우 필수
    """
    overlayId: str = None
    peerId: str = None
    accessKey: str = None
        
    def mandatoryCheck(self) -> bool:
        if not _check(self.overlayId):
            return False
        if not _check(self.peerId):
            return False
        
        return True

@dataclass
class RemovalRequest(Request):
    """
    서비스 세션 삭제 요청 Class
        mandatory:
            overlayId, ownerId, adminKey
    """
    overlayId: str = None
    ownerId: str = None
    adminKey: str = None
        
    def mandatoryCheck(self) -> bool:
        if not _check(self.overlayId):
            return False
        if not _check(self.ownerId):
            return False
        if not _check(self.adminKey):
            return False
        
        return True

@dataclass
class Notification:
    notificationType: NotificationType = None
    overlayId: str = None
    peerId: str = None

@dataclass
class SessionChangeNotification(Notification):
    """
    세션 변경 알림 Class
    """
    def __init__(
        self,
        overlayId: str = None,
        title: str = None,
        description: str = None,
        startDateTime: str = None, #YYYYMMDDHHmmSS
        endDateTime: str = None, #YYYYMMDDHHmmSS
        ownerId: str = None,
        accessKey: str = None,
        sourceList: List[str] = None,
        channelList: List[Channel] = None
    ):
        self.notificationType = NotificationType.SessionChangeNotification
        self.overlayId = overlayId
        self.title = title
        self.description = description
        self.startDateTime = startDateTime
        self.endDateTime = endDateTime
        self.ownerId = ownerId
        self.accessKey = accessKey
        self.sourceList = sourceList
        self.channelList = channelList

    title: str = None
    description: str = None
    startDateTime: str = None #YYYYMMDDHHmmSS
    endDateTime: str = None #YYYYMMDDHHmmSS
    ownerId: str = None
    accessKey: str = None
    sourceList: List[str] = None
    channelList: List[Channel] = None

@dataclass
class SessionTerminationNotification(Notification):
    """
    세션 종료 알림 Class
    """
    def __init__(self, overlayId: str = None, peerId: str = None):
        self.notificationType = NotificationType.SessionTerminationNotification
        self.overlayId = overlayId
        self.peerId = peerId

@dataclass
class PeerChangeNotification(Notification):
    """
    참가, 탈퇴하는 Peer 알림 Class
    """
    def __init__(
        self,
        overlayId: str = None,
        peerId: str = None,
        changePeerId: str = None,
        displayName: str = None,
        leave: bool = None
    ):
        self.notificationType = NotificationType.PeerChangeNotification
        self.overlayId = overlayId
        self.peerId = peerId
        self.changePeerId = changePeerId
        self.displayName = displayName
        self.leave = leave
    
    changePeerId: str = None
    displayName: str = None
    leave: bool = None

@dataclass
class ExpulsionNotification(Notification):
    """
    방출 알림 Class
    """
    def __init__(self, overlayId: str = None, peerId: str = None):
        self.notificationType = NotificationType.ExpulsionNotification
        self.overlayId = overlayId
        self.peerId = peerId

@dataclass
class DataNotification(Notification):
    """
    Broadcast data 수신 Class
    """
    def __init__(
        self,
        overlayId: str = None,
        channelId: str = None,
        sendPeerId: str = None,
        dataType: DataType = None,
        peerId: str = None,
        data: bytes = None,
        dataLength: int = None
    ):
        self.notificationType = NotificationType.DataNotification
        self.overlayId = overlayId
        self.channelId = channelId
        self.sendPeerId = sendPeerId
        self.dataType = dataType
        self.peerId = peerId
        self.data = data
        self.dataLength = dataLength

    dataType: DataType = None
    data: bytes = None
    channelId: str = None
    sendPeerId: str = None
    dataLength: int = None