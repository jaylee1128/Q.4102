# hp2p-api
An Python API for HP2P Peer
## Usage

1. Install hp2p-api
```
$pip install hp2p-api
```
2. Edit or create 'clientconfig.json'. This file must be located in the same directory as the executable.   
```
{
    "OVERLAY_SERVER_ADDR" : "http://127.0.0.1:8081",
    "SIGNALING_SERVER_ADDR" : "ws://127.0.0.1:8082",
    "LOG_LEVEL" : 2,
    "FILE_LOG" : false,
    "STDOUT_LOG" : true
}
```
3. Import the 'hp2p-api' in your Python file using the following statement:
``` py
import hp2papi as api
```
4. Use the api   

## API   
#### SetPeerOption
Customize peer options as required.   
Must be done before calling StartGrpcServer().
``` py
api.SetPeerOption(peerOption: PeerOption)
```
#### SetGrpcPort
Set the port for the gRPC server to use.   
Must be done before calling StartGrpcServer().
``` py
api.SetGrpcPort(port: int)
```
#### StartGrpcServer
Start gRPC server. Returns false if the port is in use.
``` py
api.StartGrpcServer() -> bool
```
#### Cleanup
Stop the server and release all resources.
``` py
api.Cleanup()
```
#### SetNotificatonListene
Register a listener to receive changes in the session.
``` py
api.SetNotificatonListener(overlayId: str, peerId: str, func: Callable[[Notification], None]) -> bool
```
#### DelNotificatonListener
Remove the registered listener.
``` py
api.DelNotificatonListener(overlayId: str, peerId: str) -> bool
```
#### Creation
Create a new session.
``` py
api.Creation(request: api.CreationRequest) -> CreationResponse
```
#### Query
Search for the existing session.
``` py
api.Query(overlayId: str = None, title: str = None, description: str = None) -> QueryResponse
```
#### Modification
Modify the session.
``` py
api.Modification(request: api.ModificationRequest) -> Response
```
#### Join
Join the session.
``` py
api.Join(request: api.JoinRequest) -> JoinResponse
```
#### SearchPeer
Retrieve peers participating in the session.
``` py
api.SearchPeer(request: api.SearchPeerRequest) -> SearchPeerResponse
```
#### SendData
Send the data.
``` py
api.SendData(request: api.SendDataRequest) -> Response
```
#### Leave
Leave the session.
``` py
api.Leave(request: api.LeaveRequest) -> Response
```
#### Removal
Remove the session.
``` py
api.Removal(request: api.RemovalRequest) -> Response
```
  
## LICENSE

The MIT License

Copyright (c) 2022 ETRI

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
