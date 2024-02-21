//
// The MIT License
//
// Copyright (c) 2022 ETRI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package webrtc

import (
	"encoding/json"

	"hp2p.util/util"

	"github.com/gorilla/websocket"
)

type SignalHandler struct {
	end     chan struct{}
	message chan []byte
	result  *chan interface{}
	conn    *websocket.Conn
}

func (self *SignalHandler) handler() {

	for {
		select {
		case <-self.end:
			util.Println(util.INFO, "signal server end!!!")
			return
		case msg := <-self.message:
			//log.Printf("signalhandler: %s", string(msg))

			tmp := util.TypeGetter{}
			err := json.Unmarshal(msg, &tmp)
			if err != nil {
				util.Println(util.ERROR, "signal parsing error :", err)
			} else {
				switch tmp.Type {
				case "offer", "answer":
					sdp := util.RTCSessionDescription{}
					json.Unmarshal(msg, &sdp)
					*self.result <- sdp
				case "candidate":
					ice := util.RTCIceCandidate{}
					json.Unmarshal(msg, &ice)
					*self.result <- ice
				}
			}
		}
	}
}

func (self *SignalHandler) read() {
	defer close(self.end)

	for {
		_, message, err := self.conn.ReadMessage()
		if err != nil {
			util.Println(util.ERROR, "signal recv error:", err)
			return
		}
		util.Println(util.INFO, "signal recv :", string(message))
		self.message <- message
	}
}

func (self *SignalHandler) Start(addr string, rsltchan *chan interface{}) {

	self.result = rsltchan

	self.end = make(chan struct{})
	self.message = make(chan []byte)

	util.Println(util.INFO, "signal server start :", addr)
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		util.Println(util.ERROR, "signal server error :", err)
		return
	}
	//defer c.Close()

	self.conn = c

	go self.handler()
	go self.read()
}

func (self *SignalHandler) Send(msg []byte) {
	if self.conn == nil {
		return
	}

	err := self.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		util.Println(util.ERROR, "signal send error:", err)
		return
	}

	//util.Println(util.INFO, "signal send :", string(msg))
}
