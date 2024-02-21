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
import subprocess, os
import platform
import threading
from .classes import PeerOption
from .logger import peerprint, print, printError

class ProcessHandler:
    __process = None
    __workerThread = None
    __peerOption: PeerOption = None

    def __preexec_function(self):
        # 자식 프로세스를 그룹화하여 부모 프로세스 종료 시 자식 프로세스도 종료되도록 함 -> 하지만 안되네??
        os.setpgrp()

    def __getPeerPath(self):
        os_name = platform.system()

        peerexec = ""

        if os_name == "Windows":
            peerexec = "/peer/hp2p.go.peer.exe"
        elif os_name == "Linux":
            peerexec = "/peer/hp2p.go.peer"
        else:
            peerexec = "/peer/hp2p.go.peer.mac"

        current_file_path = os.path.abspath(__file__)
        current_directory = os.path.dirname(current_file_path)
        peerpath = current_directory + peerexec
        
        return peerpath

    def SetPeerOption(self, option: PeerOption):
        self.__peerOption = option

    def PeerStart(self, peerId :str, grpcPort :int, index: str) -> bool:
        try:
            peerpath = self.__getPeerPath()

            execstr = [
                peerpath,
                "-id", peerId,
                "-gp", str(grpcPort),
                "-pi", index,
                "-cp", os.getcwd(),
                "-g"
            ]

            if self.__peerOption:
                execstr.append("-hi")
                execstr.append(str(self.__peerOption.HeartbeatInterval))
                execstr.append("-ht")
                execstr.append(str(self.__peerOption.HeartbeatTimeout))
                execstr.append("-mn")
                execstr.append(str(self.__peerOption.MNCache))
                execstr.append("-md")
                execstr.append(str(self.__peerOption.MDCache))
                execstr.append("-r")
                execstr.append(self.__peerOption.RecoveryBy)
                execstr.append("-rq")
                execstr.append(str(self.__peerOption.RateControlQuantity))
                execstr.append("-rb")
                execstr.append(str(self.__peerOption.RateControlBitrate))

            # preexec_fn=self.__preexec_function,
            self.__process = subprocess.Popen(execstr, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            self.__workerThread = threading.Thread(target=self.__peerWorker)
            self.__workerThread.start()
            return True
            
        except Exception as e:
            printError("Error: Peer execution failed.")
            printError(e)
            return False
        
    def __peerWorker(self):
        for line in self.__process.stdout:
            msg = line.decode("utf-8")
            # 마지막 뉴라인 제거
            msg = msg[:-1]
            peerprint(msg)
            
        if self.__process:
            self.__process.wait()

    def PeerStop(self):
        try:
            if self.__process:
                print(f"Terminating peer... procId:{self.__process.pid}")
                self.__process.terminate()
                self.__process = None
                self.__workerThread.join()
                self.__workerThread = None
                print(f"Peer is terminated.")
            else:
                print("Peer is not running.")
        except:
            printError("Error: Peer termination failed.")