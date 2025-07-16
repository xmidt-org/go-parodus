// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/xmidt-org/webpa-common/v2/logging" // nolint:staticcheck
	"github.com/xmidt-org/wrp-go/v3"
	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/push"
)

func ReadPump(pullSock mangos.Socket, msgBus chan []byte, logger log.Logger) {
	logging.Debug(logger).Log(logging.MessageKey(), "Starting out<-svc handler")
	for {
		data, err := pullSock.Recv()
		logging.Debug(logger).Log(logging.MessageKey(), "read pump received bytes", "len", len(data))
		if err != nil {
			logging.Error(logger).Log(logging.MessageKey(), "failed to receive message", logging.ErrorKey(), err)
			return
		}
		msgBus <- data
	}
	// where to close pullSock?
}

func WritePump(pushSock mangos.Socket, bus chan wrp.Message, stopWriting chan struct{}, logger log.Logger) {
	for {
		select {
		case <-stopWriting:
			logging.Debug(logger).Log(logging.MessageKey(), "writing pump stopping")
			// TODO: should I do more logic for error handling
			pushSock.Close()
			return
		case msg := <-bus:
			logging.Debug(logger).Log(logging.MessageKey(), "sending message", "wrp", msg.MessageType().String())
			data := wrp.MustEncode(&msg, wrp.Msgpack)
			err := pushSock.Send(data)
			if err != nil {
				logging.Error(logger).Log(logging.MessageKey(), "Failed to Send Message on socket", logging.ErrorKey(), err)
			}
		}
	}
}

func ParseBus(wrpBusOut chan wrp.Message, dataBusIn chan []byte, stopReading chan struct{}, logger log.Logger) {
	logging.Debug(logger).Log(logging.MessageKey(), "Starting HandleMSGBus")
	defer func() {
		logging.Debug(logger).Log(logging.MessageKey(), "HandleMSGBus has stopped")
	}()
	for {
		select {
		case <-stopReading:
			logging.Debug(logger).Log(logging.MessageKey(), "parse pump stopping")
			// TODO: should I do something else here? like close msgBus?
			return
		case data := <-dataBusIn:
			var msg wrp.Message
			err := wrp.NewDecoderBytes(data, wrp.Msgpack).Decode(&msg)
			if err != nil {
				logging.Error(logger).Log(logging.MessageKey(), "failed to decode message", logging.ErrorKey(), err)
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
