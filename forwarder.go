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

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/zap"
	"nanomsg.org/go/mangos/v2"
)

// Forwarder struct forwards messages coming from Talaria down to the libparouds clients
type Forwarder struct {
	Name      string
	URL       string
	LastAlive time.Time
	logger    *zap.Logger

	stopTicker chan struct{}
	sock       mangos.Socket
}

func CreateServiceForwarder(name string, url string, logger *zap.Logger) (*Forwarder, error) {
	sock, err := client.CreatePushSocket(url)
	if err != nil {
		return nil, err
	}
	logger = logger.With(zap.String("forwarder", name))
	quit := make(chan struct{})

	forwarder := &Forwarder{
		Name:       name,
		URL:        url,
		LastAlive:  time.Now(),
		sock:       sock,
		stopTicker: quit,
		logger:     logger,
	}
	ticker := time.NewTicker(5 * time.Second)
	message := &wrp.Message{Type: wrp.ServiceAliveMessageType}
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				forwarder.HandleMessage(message)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	sock.SetPipeEventHook(func(event mangos.PipeEvent, pipe mangos.Pipe) {
		logger.Info(fmt.Sprintf("%s push socket event", name), zap.Any("event", event), zap.Any("pipe", pipe))
	})
	return forwarder, nil
}

func (forwarder *Forwarder) HandleMessage(message *wrp.Message) *wrp.Message {
	forwarder.logger.Debug("handling message", zap.Any("wrp", *message))
	err := client.SendMessage(forwarder.sock, *message)
	if err != nil {
		forwarder.logger.Error("failed to send message", zap.Error(err))
		return kratos.CreateErrorWRP(message.TransactionUUID, message.Source, message.Destination, http.StatusServiceUnavailable, err)
	}
	return nil
}

func (forwarder *Forwarder) Close() {
	forwarder.stopTicker <- struct{}{}
	err := forwarder.sock.Close()
	if err != nil {
		forwarder.logger.Error("failed to close socket", zap.Error(err))
	}
}
