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
from .classes import *
from typing import Callable
from .peerCollection import PeerCollection
from .logger import print, printDebug

__notificationListeners = dict()

__peerCollection = PeerCollection()

def __SetPeerOption(option: PeerOption) -> None:
    global __peerCollection
    __peerCollection.SetPeerOption(option)

def __SetGrpcPort(port: int) -> None:
    global __peerCollection
    __peerCollection.SetGrpcPort(port)

def __Init():
    global __peerCollection
    # __peerCollection.GrpcStart()

    __peerCollection.SetNotiCallback(__NotificationCallback)

    print("Peer API init.")

def __GrpcStart() -> bool:
    global __peerCollection
    return __peerCollection.GrpcStart()

def __Cleanup():
    global __peerCollection, __notificationListeners

    __notificationListeners.clear()
    __peerCollection.Cleanup()

    print("Peer API cleanup.")

def __CheckRequest(req: Request, func: Callable[[Request], Response]) -> Response:
    if not req or not req.mandatoryCheck():
        return Response(ResponseCode.WrongRequest)
    
    return func(req)

def __Creation(req: CreationRequest) -> CreationResponse:
    global __peerCollection

    pid = __peerCollection.NewPeerStart(req.ownerId)
    if not pid:
        return CreationResponse(code=ResponseCode.Fail)

    return __peerCollection.Creation(pid, req)

def __Query(overlayId: str = None, title: str = None, description: str = None) -> QueryResponse:
    global __peerCollection

    req = QueryRequest(overlayId=overlayId, title=title, description=description)
    pid = __peerCollection.NewPeerStart("temp")
    if not pid:
        return QueryResponse(code=ResponseCode.Fail)

    return __peerCollection.Query(pid, req)

def __Modification(req: ModificationRequest) -> Response:
    global __peerCollection

    return __peerCollection.Modification(req)

def __Join(req: JoinRequest) -> JoinResponse:
    global __peerCollection

    return __peerCollection.Join(req)

def __SearchPeer(req: SearchPeerRequest) -> SearchPeerResponse:
    global __peerCollection

    return __peerCollection.SearchPeer(req)

def __SendData(req: SendDataRequest) -> Response:
    global __peerCollection

    return __peerCollection.SendData(req)

def __Leave(req: LeaveRequest) -> Response:
    global __peerCollection

    resp = __peerCollection.Leave(req)

    if resp.code == ResponseCode.Success:
        __DelNotificatonListener(req.overlayId, req.peerId)

    return resp

def __Removal(req: RemovalRequest) -> Response:
    global __peerCollection

    resp = __peerCollection.Removal(req)

    if resp.code == ResponseCode.Success:
        __DelNotificatonListener(req.overlayId, req.ownerId)

    return resp

def __NotificationCallback(noti: Notification) -> None:
    global __notificationListeners

    if not noti:
        return

    try:
        peerid = noti.peerId.split(";")[0]
        notikey = noti.overlayId + peerid
        __notificationListeners[notikey](noti)

        if type(noti) is SessionTerminationNotification or type(noti) is ExpulsionNotification:
            __DelNotificatonListener(noti.overlayId, peerid)
    except Exception as e:
        printDebug(f"notification listener error. {e}")

def __SetNotificatonListener(overlayId: str, peerId: str, func: Callable[[Notification], None]) -> bool:
    global __notificationListeners

    if not overlayId or not peerId or not func:
        return False

    if not callable(func):
        return False

    __notificationListeners[overlayId + peerId] = func

    printDebug(__notificationListeners)

    return True

def __DelNotificatonListener(overlayId: str, peerId: str) -> bool:
    global __notificationListeners

    if not overlayId or not peerId:
        return False

    try:
        peerkey = peerId.split(";")[0]
        notikey = overlayId + peerkey

        del __notificationListeners[notikey]
        printDebug("\nDelete notification listener.")
    except:
        return False

    return True
