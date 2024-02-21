from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HeartbeatRequest(_message.Message):
    __slots__ = ["dummy"]
    DUMMY_FIELD_NUMBER: _ClassVar[int]
    dummy: int
    def __init__(self, dummy: _Optional[int] = ...) -> None: ...

class HeartbeatResponse(_message.Message):
    __slots__ = ["response"]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    response: bool
    def __init__(self, response: bool = ...) -> None: ...

class SessionChangeChannel(_message.Message):
    __slots__ = ["channelId", "sourceList"]
    CHANNELID_FIELD_NUMBER: _ClassVar[int]
    SOURCELIST_FIELD_NUMBER: _ClassVar[int]
    channelId: str
    sourceList: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, channelId: _Optional[str] = ..., sourceList: _Optional[_Iterable[str]] = ...) -> None: ...

class SessionChangeRequest(_message.Message):
    __slots__ = ["overlayId", "titleChanged", "title", "descriptionChanged", "description", "ownerIdChanged", "ownerId", "accessKeyChanged", "accessKey", "startDateTimeChanged", "startDateTime", "endDateTimeChanged", "endDateTime", "sourceListChanged", "sourceList", "channelSourceChanged", "channelList"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    TITLECHANGED_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTIONCHANGED_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    OWNERIDCHANGED_FIELD_NUMBER: _ClassVar[int]
    OWNERID_FIELD_NUMBER: _ClassVar[int]
    ACCESSKEYCHANGED_FIELD_NUMBER: _ClassVar[int]
    ACCESSKEY_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIMECHANGED_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIMECHANGED_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    SOURCELISTCHANGED_FIELD_NUMBER: _ClassVar[int]
    SOURCELIST_FIELD_NUMBER: _ClassVar[int]
    CHANNELSOURCECHANGED_FIELD_NUMBER: _ClassVar[int]
    CHANNELLIST_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    titleChanged: bool
    title: str
    descriptionChanged: bool
    description: str
    ownerIdChanged: bool
    ownerId: str
    accessKeyChanged: bool
    accessKey: str
    startDateTimeChanged: bool
    startDateTime: str
    endDateTimeChanged: bool
    endDateTime: str
    sourceListChanged: bool
    sourceList: _containers.RepeatedScalarFieldContainer[str]
    channelSourceChanged: bool
    channelList: _containers.RepeatedCompositeFieldContainer[SessionChangeChannel]
    def __init__(self, overlayId: _Optional[str] = ..., titleChanged: bool = ..., title: _Optional[str] = ..., descriptionChanged: bool = ..., description: _Optional[str] = ..., ownerIdChanged: bool = ..., ownerId: _Optional[str] = ..., accessKeyChanged: bool = ..., accessKey: _Optional[str] = ..., startDateTimeChanged: bool = ..., startDateTime: _Optional[str] = ..., endDateTimeChanged: bool = ..., endDateTime: _Optional[str] = ..., sourceListChanged: bool = ..., sourceList: _Optional[_Iterable[str]] = ..., channelSourceChanged: bool = ..., channelList: _Optional[_Iterable[_Union[SessionChangeChannel, _Mapping]]] = ...) -> None: ...

class SessionChangeResponse(_message.Message):
    __slots__ = ["ack"]
    ACK_FIELD_NUMBER: _ClassVar[int]
    ack: bool
    def __init__(self, ack: bool = ...) -> None: ...

class SessionTerminateRequest(_message.Message):
    __slots__ = ["overlayId", "peerId"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    PEERID_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    peerId: str
    def __init__(self, overlayId: _Optional[str] = ..., peerId: _Optional[str] = ...) -> None: ...

class SessionTerminateResponse(_message.Message):
    __slots__ = ["ack"]
    ACK_FIELD_NUMBER: _ClassVar[int]
    ack: bool
    def __init__(self, ack: bool = ...) -> None: ...

class PeerChangeRequest(_message.Message):
    __slots__ = ["overlayId", "peerId", "displayName", "isLeave"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    PEERID_FIELD_NUMBER: _ClassVar[int]
    DISPLAYNAME_FIELD_NUMBER: _ClassVar[int]
    ISLEAVE_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    peerId: str
    displayName: str
    isLeave: bool
    def __init__(self, overlayId: _Optional[str] = ..., peerId: _Optional[str] = ..., displayName: _Optional[str] = ..., isLeave: bool = ...) -> None: ...

class PeerChangeResponse(_message.Message):
    __slots__ = ["ack"]
    ACK_FIELD_NUMBER: _ClassVar[int]
    ack: bool
    def __init__(self, ack: bool = ...) -> None: ...

class ExpulsionRequest(_message.Message):
    __slots__ = ["overlayId", "peerId"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    PEERID_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    peerId: str
    def __init__(self, overlayId: _Optional[str] = ..., peerId: _Optional[str] = ...) -> None: ...

class ExpulsionResponse(_message.Message):
    __slots__ = ["ack"]
    ACK_FIELD_NUMBER: _ClassVar[int]
    ack: bool
    def __init__(self, ack: bool = ...) -> None: ...

class DataRequest(_message.Message):
    __slots__ = ["overlayId", "dataType", "channelId", "sendPeerId", "data"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    DATATYPE_FIELD_NUMBER: _ClassVar[int]
    CHANNELID_FIELD_NUMBER: _ClassVar[int]
    SENDPEERID_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    dataType: str
    channelId: str
    sendPeerId: str
    data: bytes
    def __init__(self, overlayId: _Optional[str] = ..., dataType: _Optional[str] = ..., channelId: _Optional[str] = ..., sendPeerId: _Optional[str] = ..., data: _Optional[bytes] = ...) -> None: ...

class DataResponse(_message.Message):
    __slots__ = ["ack"]
    ACK_FIELD_NUMBER: _ClassVar[int]
    ack: bool
    def __init__(self, ack: bool = ...) -> None: ...

class ReadyRequest(_message.Message):
    __slots__ = ["index"]
    INDEX_FIELD_NUMBER: _ClassVar[int]
    index: str
    def __init__(self, index: _Optional[str] = ...) -> None: ...

class ReadyResponse(_message.Message):
    __slots__ = ["isOk"]
    ISOK_FIELD_NUMBER: _ClassVar[int]
    isOk: bool
    def __init__(self, isOk: bool = ...) -> None: ...

class ChannelAttributeVideoFeature(_message.Message):
    __slots__ = ["target", "mode", "resolution", "frameRate", "keyPointsType", "modelUri", "hash", "dimension"]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    MODE_FIELD_NUMBER: _ClassVar[int]
    RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    FRAMERATE_FIELD_NUMBER: _ClassVar[int]
    KEYPOINTSTYPE_FIELD_NUMBER: _ClassVar[int]
    MODELURI_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    DIMENSION_FIELD_NUMBER: _ClassVar[int]
    target: str
    mode: str
    resolution: str
    frameRate: str
    keyPointsType: str
    modelUri: str
    hash: str
    dimension: str
    def __init__(self, target: _Optional[str] = ..., mode: _Optional[str] = ..., resolution: _Optional[str] = ..., frameRate: _Optional[str] = ..., keyPointsType: _Optional[str] = ..., modelUri: _Optional[str] = ..., hash: _Optional[str] = ..., dimension: _Optional[str] = ...) -> None: ...

class ChannelAttributeAudio(_message.Message):
    __slots__ = ["codec", "sampleRate", "bitrate", "channelType"]
    CODEC_FIELD_NUMBER: _ClassVar[int]
    SAMPLERATE_FIELD_NUMBER: _ClassVar[int]
    BITRATE_FIELD_NUMBER: _ClassVar[int]
    CHANNELTYPE_FIELD_NUMBER: _ClassVar[int]
    codec: str
    sampleRate: str
    bitrate: str
    channelType: str
    def __init__(self, codec: _Optional[str] = ..., sampleRate: _Optional[str] = ..., bitrate: _Optional[str] = ..., channelType: _Optional[str] = ...) -> None: ...

class ChannelAttributeText(_message.Message):
    __slots__ = ["format", "encoding"]
    FORMAT_FIELD_NUMBER: _ClassVar[int]
    ENCODING_FIELD_NUMBER: _ClassVar[int]
    format: str
    encoding: str
    def __init__(self, format: _Optional[str] = ..., encoding: _Optional[str] = ...) -> None: ...

class Channel(_message.Message):
    __slots__ = ["channelId", "channelType", "useSourceList", "sourceList", "isServiceChannel", "videoFeature", "audio", "text"]
    CHANNELID_FIELD_NUMBER: _ClassVar[int]
    CHANNELTYPE_FIELD_NUMBER: _ClassVar[int]
    USESOURCELIST_FIELD_NUMBER: _ClassVar[int]
    SOURCELIST_FIELD_NUMBER: _ClassVar[int]
    ISSERVICECHANNEL_FIELD_NUMBER: _ClassVar[int]
    VIDEOFEATURE_FIELD_NUMBER: _ClassVar[int]
    AUDIO_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    channelId: str
    channelType: str
    useSourceList: bool
    sourceList: _containers.RepeatedScalarFieldContainer[str]
    isServiceChannel: bool
    videoFeature: ChannelAttributeVideoFeature
    audio: ChannelAttributeAudio
    text: ChannelAttributeText
    def __init__(self, channelId: _Optional[str] = ..., channelType: _Optional[str] = ..., useSourceList: bool = ..., sourceList: _Optional[_Iterable[str]] = ..., isServiceChannel: bool = ..., videoFeature: _Optional[_Union[ChannelAttributeVideoFeature, _Mapping]] = ..., audio: _Optional[_Union[ChannelAttributeAudio, _Mapping]] = ..., text: _Optional[_Union[ChannelAttributeText, _Mapping]] = ...) -> None: ...

class CreationRequest(_message.Message):
    __slots__ = ["title", "description", "ownerId", "adminKey", "accessKey", "peerList", "useBlockList", "blockList", "useSourceList", "sourceList", "startDateTime", "endDateTime", "channelList"]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    OWNERID_FIELD_NUMBER: _ClassVar[int]
    ADMINKEY_FIELD_NUMBER: _ClassVar[int]
    ACCESSKEY_FIELD_NUMBER: _ClassVar[int]
    PEERLIST_FIELD_NUMBER: _ClassVar[int]
    USEBLOCKLIST_FIELD_NUMBER: _ClassVar[int]
    BLOCKLIST_FIELD_NUMBER: _ClassVar[int]
    USESOURCELIST_FIELD_NUMBER: _ClassVar[int]
    SOURCELIST_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    CHANNELLIST_FIELD_NUMBER: _ClassVar[int]
    title: str
    description: str
    ownerId: str
    adminKey: str
    accessKey: str
    peerList: _containers.RepeatedScalarFieldContainer[str]
    useBlockList: bool
    blockList: _containers.RepeatedScalarFieldContainer[str]
    useSourceList: bool
    sourceList: _containers.RepeatedScalarFieldContainer[str]
    startDateTime: str
    endDateTime: str
    channelList: _containers.RepeatedCompositeFieldContainer[Channel]
    def __init__(self, title: _Optional[str] = ..., description: _Optional[str] = ..., ownerId: _Optional[str] = ..., adminKey: _Optional[str] = ..., accessKey: _Optional[str] = ..., peerList: _Optional[_Iterable[str]] = ..., useBlockList: bool = ..., blockList: _Optional[_Iterable[str]] = ..., useSourceList: bool = ..., sourceList: _Optional[_Iterable[str]] = ..., startDateTime: _Optional[str] = ..., endDateTime: _Optional[str] = ..., channelList: _Optional[_Iterable[_Union[Channel, _Mapping]]] = ...) -> None: ...

class CreationResponse(_message.Message):
    __slots__ = ["rspCode", "overlayId", "startDateTime", "endDateTime"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    overlayId: str
    startDateTime: str
    endDateTime: str
    def __init__(self, rspCode: _Optional[int] = ..., overlayId: _Optional[str] = ..., startDateTime: _Optional[str] = ..., endDateTime: _Optional[str] = ...) -> None: ...

class QueryRequest(_message.Message):
    __slots__ = ["overlayId", "title", "description"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    title: str
    description: str
    def __init__(self, overlayId: _Optional[str] = ..., title: _Optional[str] = ..., description: _Optional[str] = ...) -> None: ...

class QueryResponseOverlay(_message.Message):
    __slots__ = ["overlayId", "title", "description", "ownerId", "closed", "startDateTime", "endDateTime", "channelList"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    OWNERID_FIELD_NUMBER: _ClassVar[int]
    CLOSED_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    CHANNELLIST_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    title: str
    description: str
    ownerId: str
    closed: int
    startDateTime: str
    endDateTime: str
    channelList: _containers.RepeatedCompositeFieldContainer[Channel]
    def __init__(self, overlayId: _Optional[str] = ..., title: _Optional[str] = ..., description: _Optional[str] = ..., ownerId: _Optional[str] = ..., closed: _Optional[int] = ..., startDateTime: _Optional[str] = ..., endDateTime: _Optional[str] = ..., channelList: _Optional[_Iterable[_Union[Channel, _Mapping]]] = ...) -> None: ...

class QueryResponse(_message.Message):
    __slots__ = ["rspCode", "overlayList"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    OVERLAYLIST_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    overlayList: _containers.RepeatedCompositeFieldContainer[QueryResponseOverlay]
    def __init__(self, rspCode: _Optional[int] = ..., overlayList: _Optional[_Iterable[_Union[QueryResponseOverlay, _Mapping]]] = ...) -> None: ...

class JoinRequest(_message.Message):
    __slots__ = ["overlayId", "accessKey", "peerId", "displayName", "privateKeyPath"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    ACCESSKEY_FIELD_NUMBER: _ClassVar[int]
    PEERID_FIELD_NUMBER: _ClassVar[int]
    DISPLAYNAME_FIELD_NUMBER: _ClassVar[int]
    PRIVATEKEYPATH_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    accessKey: str
    peerId: str
    displayName: str
    privateKeyPath: str
    def __init__(self, overlayId: _Optional[str] = ..., accessKey: _Optional[str] = ..., peerId: _Optional[str] = ..., displayName: _Optional[str] = ..., privateKeyPath: _Optional[str] = ...) -> None: ...

class JoinResponse(_message.Message):
    __slots__ = ["rspCode", "overlayId", "startDateTime", "endDateTime", "sourceList", "channelList"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    SOURCELIST_FIELD_NUMBER: _ClassVar[int]
    CHANNELLIST_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    overlayId: str
    startDateTime: str
    endDateTime: str
    sourceList: _containers.RepeatedScalarFieldContainer[str]
    channelList: _containers.RepeatedCompositeFieldContainer[Channel]
    def __init__(self, rspCode: _Optional[int] = ..., overlayId: _Optional[str] = ..., startDateTime: _Optional[str] = ..., endDateTime: _Optional[str] = ..., sourceList: _Optional[_Iterable[str]] = ..., channelList: _Optional[_Iterable[_Union[Channel, _Mapping]]] = ...) -> None: ...

class ModificationRequest(_message.Message):
    __slots__ = ["overlayId", "ownerId", "adminKey", "modificationTitle", "title", "modificationDescription", "description", "modificationAccessKey", "accessKey", "modificationPeerList", "peerList", "modificationBlockList", "blockList", "modificationSourceList", "sourceList", "modificationStartDateTime", "startDateTime", "modificationEndDateTime", "endDateTime", "modificationChannelList", "channelList", "modificationOwnerId", "newOwnerId", "modificationAdminKey", "newAdminKey"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    OWNERID_FIELD_NUMBER: _ClassVar[int]
    ADMINKEY_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONTITLE_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONDESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONACCESSKEY_FIELD_NUMBER: _ClassVar[int]
    ACCESSKEY_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONPEERLIST_FIELD_NUMBER: _ClassVar[int]
    PEERLIST_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONBLOCKLIST_FIELD_NUMBER: _ClassVar[int]
    BLOCKLIST_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONSOURCELIST_FIELD_NUMBER: _ClassVar[int]
    SOURCELIST_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONSTARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    STARTDATETIME_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    ENDDATETIME_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONCHANNELLIST_FIELD_NUMBER: _ClassVar[int]
    CHANNELLIST_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONOWNERID_FIELD_NUMBER: _ClassVar[int]
    NEWOWNERID_FIELD_NUMBER: _ClassVar[int]
    MODIFICATIONADMINKEY_FIELD_NUMBER: _ClassVar[int]
    NEWADMINKEY_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    ownerId: str
    adminKey: str
    modificationTitle: bool
    title: str
    modificationDescription: bool
    description: str
    modificationAccessKey: bool
    accessKey: str
    modificationPeerList: bool
    peerList: _containers.RepeatedScalarFieldContainer[str]
    modificationBlockList: bool
    blockList: _containers.RepeatedScalarFieldContainer[str]
    modificationSourceList: bool
    sourceList: _containers.RepeatedScalarFieldContainer[str]
    modificationStartDateTime: bool
    startDateTime: str
    modificationEndDateTime: bool
    endDateTime: str
    modificationChannelList: bool
    channelList: _containers.RepeatedCompositeFieldContainer[Channel]
    modificationOwnerId: bool
    newOwnerId: str
    modificationAdminKey: bool
    newAdminKey: str
    def __init__(self, overlayId: _Optional[str] = ..., ownerId: _Optional[str] = ..., adminKey: _Optional[str] = ..., modificationTitle: bool = ..., title: _Optional[str] = ..., modificationDescription: bool = ..., description: _Optional[str] = ..., modificationAccessKey: bool = ..., accessKey: _Optional[str] = ..., modificationPeerList: bool = ..., peerList: _Optional[_Iterable[str]] = ..., modificationBlockList: bool = ..., blockList: _Optional[_Iterable[str]] = ..., modificationSourceList: bool = ..., sourceList: _Optional[_Iterable[str]] = ..., modificationStartDateTime: bool = ..., startDateTime: _Optional[str] = ..., modificationEndDateTime: bool = ..., endDateTime: _Optional[str] = ..., modificationChannelList: bool = ..., channelList: _Optional[_Iterable[_Union[Channel, _Mapping]]] = ..., modificationOwnerId: bool = ..., newOwnerId: _Optional[str] = ..., modificationAdminKey: bool = ..., newAdminKey: _Optional[str] = ...) -> None: ...

class ModificationResponse(_message.Message):
    __slots__ = ["rspCode"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    def __init__(self, rspCode: _Optional[int] = ...) -> None: ...

class RemovalRequest(_message.Message):
    __slots__ = ["overlayId", "ownerId", "adminKey"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    OWNERID_FIELD_NUMBER: _ClassVar[int]
    ADMINKEY_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    ownerId: str
    adminKey: str
    def __init__(self, overlayId: _Optional[str] = ..., ownerId: _Optional[str] = ..., adminKey: _Optional[str] = ...) -> None: ...

class RemovalResponse(_message.Message):
    __slots__ = ["rspCode"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    def __init__(self, rspCode: _Optional[int] = ...) -> None: ...

class LeaveRequest(_message.Message):
    __slots__ = ["overlayId", "peerId", "accessKey"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    PEERID_FIELD_NUMBER: _ClassVar[int]
    ACCESSKEY_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    peerId: str
    accessKey: str
    def __init__(self, overlayId: _Optional[str] = ..., peerId: _Optional[str] = ..., accessKey: _Optional[str] = ...) -> None: ...

class LeaveResponse(_message.Message):
    __slots__ = ["rspCode"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    def __init__(self, rspCode: _Optional[int] = ...) -> None: ...

class Peer(_message.Message):
    __slots__ = ["peerId", "displayName"]
    PEERID_FIELD_NUMBER: _ClassVar[int]
    DISPLAYNAME_FIELD_NUMBER: _ClassVar[int]
    peerId: str
    displayName: str
    def __init__(self, peerId: _Optional[str] = ..., displayName: _Optional[str] = ...) -> None: ...

class SearchPeerRequest(_message.Message):
    __slots__ = ["overlayId"]
    OVERLAYID_FIELD_NUMBER: _ClassVar[int]
    overlayId: str
    def __init__(self, overlayId: _Optional[str] = ...) -> None: ...

class SearchPeerResponse(_message.Message):
    __slots__ = ["rspCode", "peerList"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    PEERLIST_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    peerList: _containers.RepeatedCompositeFieldContainer[Peer]
    def __init__(self, rspCode: _Optional[int] = ..., peerList: _Optional[_Iterable[_Union[Peer, _Mapping]]] = ...) -> None: ...

class SendDataRequest(_message.Message):
    __slots__ = ["dataType", "channelId", "data"]
    DATATYPE_FIELD_NUMBER: _ClassVar[int]
    CHANNELID_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    dataType: str
    channelId: str
    data: bytes
    def __init__(self, dataType: _Optional[str] = ..., channelId: _Optional[str] = ..., data: _Optional[bytes] = ...) -> None: ...

class SendDataResponse(_message.Message):
    __slots__ = ["rspCode"]
    RSPCODE_FIELD_NUMBER: _ClassVar[int]
    rspCode: int
    def __init__(self, rspCode: _Optional[int] = ...) -> None: ...

class Request(_message.Message):
    __slots__ = ["id", "creation", "query", "join", "modification", "removal", "leave", "searchPeer", "sendData"]
    ID_FIELD_NUMBER: _ClassVar[int]
    CREATION_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    JOIN_FIELD_NUMBER: _ClassVar[int]
    MODIFICATION_FIELD_NUMBER: _ClassVar[int]
    REMOVAL_FIELD_NUMBER: _ClassVar[int]
    LEAVE_FIELD_NUMBER: _ClassVar[int]
    SEARCHPEER_FIELD_NUMBER: _ClassVar[int]
    SENDDATA_FIELD_NUMBER: _ClassVar[int]
    id: str
    creation: CreationRequest
    query: QueryRequest
    join: JoinRequest
    modification: ModificationRequest
    removal: RemovalRequest
    leave: LeaveRequest
    searchPeer: SearchPeerRequest
    sendData: SendDataRequest
    def __init__(self, id: _Optional[str] = ..., creation: _Optional[_Union[CreationRequest, _Mapping]] = ..., query: _Optional[_Union[QueryRequest, _Mapping]] = ..., join: _Optional[_Union[JoinRequest, _Mapping]] = ..., modification: _Optional[_Union[ModificationRequest, _Mapping]] = ..., removal: _Optional[_Union[RemovalRequest, _Mapping]] = ..., leave: _Optional[_Union[LeaveRequest, _Mapping]] = ..., searchPeer: _Optional[_Union[SearchPeerRequest, _Mapping]] = ..., sendData: _Optional[_Union[SendDataRequest, _Mapping]] = ...) -> None: ...

class Response(_message.Message):
    __slots__ = ["id", "creation", "query", "join", "modification", "removal", "leave", "searchPeer", "sendData"]
    ID_FIELD_NUMBER: _ClassVar[int]
    CREATION_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    JOIN_FIELD_NUMBER: _ClassVar[int]
    MODIFICATION_FIELD_NUMBER: _ClassVar[int]
    REMOVAL_FIELD_NUMBER: _ClassVar[int]
    LEAVE_FIELD_NUMBER: _ClassVar[int]
    SEARCHPEER_FIELD_NUMBER: _ClassVar[int]
    SENDDATA_FIELD_NUMBER: _ClassVar[int]
    id: str
    creation: CreationResponse
    query: QueryResponse
    join: JoinResponse
    modification: ModificationResponse
    removal: RemovalResponse
    leave: LeaveResponse
    searchPeer: SearchPeerResponse
    sendData: SendDataResponse
    def __init__(self, id: _Optional[str] = ..., creation: _Optional[_Union[CreationResponse, _Mapping]] = ..., query: _Optional[_Union[QueryResponse, _Mapping]] = ..., join: _Optional[_Union[JoinResponse, _Mapping]] = ..., modification: _Optional[_Union[ModificationResponse, _Mapping]] = ..., removal: _Optional[_Union[RemovalResponse, _Mapping]] = ..., leave: _Optional[_Union[LeaveResponse, _Mapping]] = ..., searchPeer: _Optional[_Union[SearchPeerResponse, _Mapping]] = ..., sendData: _Optional[_Union[SendDataResponse, _Mapping]] = ...) -> None: ...
