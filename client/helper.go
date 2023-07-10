/*
 *  Copyright 2019 Comcast Cable Communications Management, LLC
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package client

import (
	"fmt"

	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/zap"
	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/push"
)

func ReadPump(pullSock mangos.Socket, msgBus chan []byte, logger *zap.Logger) {
	logger.Debug("Starting out<-svc handler")
	for {
		data, err := pullSock.Recv()
		logger.Debug("read pump received bytes", zap.Int("len", len(data)))
		if err != nil {
			logger.Error("failed to receive message", zap.Error(err))
			return
		}
		msgBus <- data
	}
	// where to close pullSock?
}

func WritePump(pushSock mangos.Socket, bus chan wrp.Message, stopWriting chan struct{}, logger *zap.Logger) {
	for {
		select {
		case <-stopWriting:
			logger.Debug("writing pump stopping")
			// TODO: should I do more logic for error handling
			pushSock.Close()
			return
		case msg := <-bus:
			logger.Debug("sending message", zap.String("wrp", msg.MessageType().String()))
			data := wrp.MustEncode(&msg, wrp.Msgpack)
			err := pushSock.Send(data)
			if err != nil {
				logger.Error("Failed to Send Message on socket", zap.Error(err))
			}
		}
	}
}

func ParseBus(wrpBusOut chan wrp.Message, dataBusIn chan []byte, stopReading chan struct{}, logger *zap.Logger) {
	logger.Debug("Starting HandleMSGBus")
	defer func() {
		logger.Debug("HandleMSGBus has stopped")
	}()
	for {
		select {
		case <-stopReading:
			logger.Debug("parse pump stopping")
			// TODO: should I do something else here? like close msgBus?
			return
		case data := <-dataBusIn:
			var msg wrp.Message
			err := wrp.NewDecoderBytes(data, wrp.Msgpack).Decode(&msg)
			if err != nil {
				logger.Error("failed to decode message", zap.Error(err))
				continue
			}
			wrpBusOut <- msg
		}
	}
}

func CreatePushSocket(url string) (mangos.Socket, error) {
	if sock, err := push.NewSocket(); err != nil {
		return nil, fmt.Errorf("can't get new push socket: %s", err)
	} else {
		if err := sock.DialOptions(url, map[string]interface{}{mangos.OptionDialAsynch: false}); err != nil {
			return nil, fmt.Errorf("can't dial on push socket: %s", err)
		}
		return sock, nil
	}
}

// SendMessage will encode the message as msgpack and send the message.
// Note this is async. I couldn't find a way to make this synchronous.
func SendMessage(sock mangos.Socket, msg wrp.Message) error {
	var buffer []byte
	if err := wrp.NewEncoderBytes(&buffer, wrp.Msgpack).Encode(&msg); err != nil {
		return err
	}
	if err := sock.Send(buffer); err != nil {
		return err
	}
	return nil
}
