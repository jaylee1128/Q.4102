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

package util

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

var LEVEL int = 2

var INFO int = 2
var WORK int = 1
var ERROR int = 0

var FileLog bool = false
var StdoutLog bool = true

var ilogger *log.Logger
var wlogger *log.Logger
var elogger *log.Logger

var logPath string = ""

func LogInit() {
	writer := []io.Writer{}

	if FileLog {
		if logPath != "" && logPath[len(logPath)-1:] != "/" {
			logPath = logPath + "/"
		}

		logpath := logPath + "hp2p_" + time.Now().Format("2006-01-02_15_04_05") + ".log"
		file, err := os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatal(err)
		}

		writer = append(writer, file)
	}

	if StdoutLog {
		writer = append(writer, os.Stdout)
	}

	mw := io.MultiWriter(writer...)

	ilogger = log.New(mw, "[INFO] ", log.LstdFlags)
	wlogger = log.New(mw, "[WORK] ", log.LstdFlags)
	elogger = log.New(mw, "[ERROR] ", log.LstdFlags)
}

func getLevelLogger(level int) *log.Logger {

	if ilogger == nil || wlogger == nil || elogger == nil {
		LogInit()
	}

	switch level {
	case INFO:
		return ilogger
	case WORK:
		return wlogger
	case ERROR:
		return elogger
	default:
		return ilogger
	}
}

func Printf(level int, format string, v ...interface{}) {
	if level <= LEVEL {
		getLevelLogger(level).Printf(format, v...)
	}
}

func Println(level int, v ...interface{}) {
	if level <= LEVEL {
		getLevelLogger(level).Println(v...)
	}
}

func PrintJson(level int, header string, jsonv interface{}) {
	buf, err := json.MarshalIndent(jsonv, "", "   ")
	if err != nil {
		getLevelLogger(level).Println(level, header, "json marshal err: ", err)
	} else {
		getLevelLogger(level).Println(level, header, ": ", string(buf))
	}
}

func SetLevel(level int) {
	LEVEL = level
}

func SetWriter(fileLog bool, stdoutLog bool) {
	FileLog = fileLog
	StdoutLog = stdoutLog
}

func SetLogPath(path string) {
	logPath = path
}
