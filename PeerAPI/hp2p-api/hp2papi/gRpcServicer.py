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
import socket
from .pb2 import api_pb2
from .pb2 import api_pb2_grpc
import grpc
from concurrent import futures
import threading
import queue
from .classes import GrpcMsgType
from .logger import print, printError, printDebug

class Hp2pServicer(api_pb2_grpc.Hp2pApiProtoServicer):
    notiQueue = None
    grpcServer = None
    grpcPort = 50051
    serverThread = None
    isStop = False
    isCustomPort = False
    peers = None

    peerIndexChangeEvent:threading.Event = None

    checkPorts = [i for i in range(50051, 50061)]

    def __init__(self):
        self.notiQueue = queue.Queue()
        self.grpcServer = None
        self.grpcPort = 50051
        self.serverThread = None
        self.peerIndexChangeEvent = threading.Event()

    def SetPeerIndexChange(self) -> None:
        self.peerIndexChangeEvent.set()
        

    def WaitPeerIndexChange(self) -> None:
        self.peerIndexChangeEvent.wait()
        self.peerIndexChangeEvent.clear()
        

    def GetNotiQueue(self) -> queue.Queue:
        return self.notiQueue
    
    def SetGrpcPort(self, port: int) -> None:
        self.grpcPort = port
        self.isCustomPort = True

    def GetGrpcPort(self) -> int:
        return self.grpcPort

    def check_port(self, port: int) -> bool:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            return s.connect_ex(('localhost', port)) != 0

    def serverRun(self) -> None:
        self.grpcServer = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
        api_pb2_grpc.add_Hp2pApiProtoServicer_to_server(self, self.grpcServer)
        self.grpcServer.add_insecure_port('[::]:' + str(self.grpcPort))
        self.grpcServer.start()
        print("Python gRPC Server is running on port " + str(self.grpcPort))
        self.grpcServer.wait_for_termination()

    def GrpcStart(self, peers : dict) -> bool:
        if self.isCustomPort:
            if not self.check_port(self.grpcPort):
                printError(f"GrpcServer start fail. Port {self.grpcPort} is already in use.")
                return False
        else:
            for port in self.checkPorts:
                if self.check_port(port):
                    self.grpcPort = port
                    break
            else:
                printError(f"GrpcServer start fail. All ports {self.checkPorts} are already in use.")
                return False

        self.serverThread = threading.Thread(target=self.serverRun)
        self.serverThread.start()

        self.peers = peers

        return True

    def GrpcStop(self) -> None:
        print("gRPC Server is stopping...")

        self.isStop = True

        if self.serverThread and self.serverThread.is_alive():
            if self.grpcServer:
                self.grpcServer.stop(5)
            self.serverThread.join()
        print("gRPC Server is stopped.")

    # def Creation(self, request, context):
    #     print("Creation")
    #     print(request)
    #     return api_pb2.CreationResponse(
    #         rspCode=404,
    #     )

    def Ready(self, request, context):
        printDebug(f"GRPC has received : Ready, {request}")
        self.notiQueue.put({"type": GrpcMsgType.Ready, "index": request.index, "context": context})

        return api_pb2.ReadyResponse(
            isOk=True,
        )
    
    def Heartbeat(self, request, context):
        return api_pb2.HeartbeatResponse(
            response=not self.isStop,
        )
    
    def SessionTermination(self, request, context):
        self.notiQueue.put({"type": GrpcMsgType.SessionTermination, "data": request})
        return api_pb2.SessionTerminateResponse(
            ack=True,
        )
    def PeerChange(self, request, context):
        self.notiQueue.put({"type": GrpcMsgType.PeerChange, "data": request})
        return api_pb2.PeerChangeResponse(
            ack=True,
        )
    
    def SessionChange(self, request, context):
        self.notiQueue.put({"type": GrpcMsgType.SessionChange, "data": request})
        return api_pb2.SessionChangeResponse(
            ack=True,
        )
    
    def Expulsion(self, request, context):
        self.notiQueue.put({"type": GrpcMsgType.Expulsion, "data": request})
        return api_pb2.ExpulsionResponse(
            ack=True,
        )
    
    def Data(self, request, context):
        self.notiQueue.put({"type": GrpcMsgType.Data, "data": request})
        return api_pb2.DataResponse(
            ack=True,
        )
    
    def getGrpcMsgType(self, response):
        if response.HasField("creation"):
            return GrpcMsgType.Creation
        elif response.HasField("join"):
            return GrpcMsgType.Join
        elif response.HasField("query"):
            return GrpcMsgType.Query
        elif response.HasField("modification"):
            return GrpcMsgType.Modification
        elif response.HasField("removal"):
            return GrpcMsgType.Removal
        elif response.HasField("leave"):
            return GrpcMsgType.Leave
        elif response.HasField("sendData"):
            return GrpcMsgType.SendData
        elif response.HasField("searchPeer"):
            return GrpcMsgType.SearchPeer
        elif response.HasField("notMine"):
            return GrpcMsgType.NotMine
        else:
            return None

    def Homp(self, request_iterator, context):
        try:
            for response in request_iterator:
                if self.isStop:
                    break
                if response.HasField("peerIndex"):
                    peerIndexRes = response.peerIndex
                    peerIndex = peerIndexRes.peerIndex
                    peerId = peerIndexRes.peerId

                    if peerIndex in self.peers:
                        peer = self.peers[peerIndex]

                        while context.is_active() and not self.isStop:
                            try: 
                                msg = peer.streamQueue.get(timeout=1)
                                printDebug(f"Request: {msg['type']}")
                                printDebug(msg["data"])
                                
                                yield msg["data"]
                                break
                            except queue.Empty:
                                continue
                            except Exception as e:
                                printError(f"!!!!!!!!!!!!!!!!!!!! Homp error: {e}")
                    else:
                        printError(f"!!!!!!!!!!!! Peer {peerIndex} is not exist !!!!!!!!!!!!!!!!")
                else:
                    self.notiQueue.put({"type": self.getGrpcMsgType(response), "data": response})
                    if response.HasField("creation") or response.HasField("join"):
                        self.WaitPeerIndexChange()
                        printDebug("Detect PeerIndexChange")

        except Exception as e:
            if context.is_active():
                printError(f"!!!!!!!!!!!!!!!!!!!!!! Homp error: {e}")

        '''
        while context.is_active() and not self.isStop:
            try:
                msg = self.streamQueue.get(timeout=1)
                printDebug(f"Request: {msg['type']}")
                printDebug(msg["data"])
                yield msg["data"]

                response = next(request_iterator)
                printDebug(f"Response: {response}")
                self.notiQueue.put({"type": self.getGrpcMsgType(response), "data": response})
                ##aaa = next(request_iterator)
                
            except queue.Empty:
                continue
            except Exception as e:
                if e:
                    printError(f"Homp error: {e}")
                continue
        '''