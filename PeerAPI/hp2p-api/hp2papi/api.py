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
"""
참조 문서 : https://docs.google.com/document/d/1Yz60FiMxJi8upZrMvffDTRNW4mUrCpUf

python3.8

@version 0.1
@author jaylee@jayb.kr
"""

from typing import Callable
from .classes import *
from .logger import SetLogToFile, SetLogLevel
from .api_func import (
    __CheckRequest,
    __Creation,
    __Query,
    __Modification,
    __Join,
    __SearchPeer,
    __SendData,
    __Leave,
    __Removal,
    __SetNotificatonListener,
    __DelNotificatonListener,
    __SetGrpcPort,
    __Cleanup,
    __SetPeerOption,
    __GrpcStart
)

__all__ = [
    "SetLogLevel",
    "SetLogToFile",
    "SetGrpcPort",
    "StartGrpcServer",
    "SetPeerOption",
    "Cleanup",
    "Creation",
    "Query",
    "Modification",
    "Join",
    "SearchPeer",
    "SendData",
    "Leave",
    "Removal",
    "SetNotificatonListener",
]
'''
def SetLogToFile() -> None:
    logger.SetLogToFile()

def SetLogLevel(level: str) -> None:
    logger.SetLogLevel(level)
'''
def SetPeerOption(peerOption: PeerOption) -> None:
    """
    Peer 설정
    """
    __SetPeerOption(peerOption)

def SetGrpcPort(port: int) -> None:
    """
    gRPC 포트 설정
    """
    __SetGrpcPort(port)

def StartGrpcServer() -> bool:
    """
    gRPC 서버 시작
    """
    return __GrpcStart()

def Cleanup() -> None:
    """
    종료
    """
    __Cleanup()

def Creation(req: CreationRequest) -> CreationResponse:
    """
    신규 서비스 세션 생성
    """
    return __CheckRequest(req, __Creation)

def Query(overlayId: str = None, title: str = None, description: str = None) -> QueryResponse:
    """
    서비스 세션 목록 조회
    """
    return __Query(overlayId, title, description)

def Modification(req: ModificationRequest) -> Response:
    """
    서비스 세션 변경
    """
    return __CheckRequest(req, __Modification)

def Join(req: JoinRequest) -> JoinResponse:
    """
    서비스 세션 참가
    """
    return __CheckRequest(req, __Join)

def SearchPeer(req: SearchPeerRequest) -> SearchPeerResponse:
    """
    서비스 세션에 참가한 Peer 목록 조회
    """
    return __CheckRequest(req, __SearchPeer)

def SendData(req: SendDataRequest) -> Response:
    """
    Data broadcast
    """
    return __CheckRequest(req, __SendData)

def Leave(req: LeaveRequest) -> Response:
    """
    Data broadcast
    """
    return __CheckRequest(req, __Leave)

def Removal(req: RemovalRequest) -> Response:
    """
    서비스 세션 삭제
    """
    return __CheckRequest(req, __Removal)

def SetNotificatonListener(overlayId: str, peerId: str, func: Callable[[Notification], None]) -> bool:
    """
    세션 변경 내용 수신 설정
    """
    return __SetNotificatonListener(overlayId, peerId, func)

def DelNotificatonListener(overlayId: str, peerId: str) -> bool:
    """
    세션 변경 내용 수신 설정 해제
    """
    return __DelNotificatonListener(overlayId, peerId)