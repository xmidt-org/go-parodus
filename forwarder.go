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

	"github.com/go-kit/log"
	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/webpa-common/v2/logging" // nolint:staticcheck
	"github.com/xmidt-org/wrp-go/v3"
	"nanomsg.org/go/mangos/v2"
)

// Forwarder struct forwards messages coming from Talaria down to the libparouds clients
type Forwarder struct {
	Name      string
	URL       string
	LastAlive time.Time
	logger    log.Logger

	stopTicker chan struct{}
	sock       mangos.Socket
}

func CreateServiceForwarder(name string, url string, logger log.Logger) (*Forwarder, error) {
	sock, err := client.CreatePushSocket(url)
	if err != nil {
		return nil, err
	}

	quit := make(chan struct{})

	forwarder := &Forwarder{
		Name:       name,
		URL:        url,
		LastAlive:  time.Now(),
		sock:       sock,
		stopTicker: quit,
		logger:     log.WithPrefix(logger, "forwarder", name),
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
		logging.Info(logger).Log(logging.MessageKey(), fmt.Sprintf("%s push socket event", name), "event", event, "pipe", pipe)
	})
	return forwarder, nil
}

func (forwarder *Forwarder) HandleMessage(message *wrp.Message) *wrp.Message {
	logging.Debug(forwarder.logger).Log(logging.MessageKey(), "handling message", "wrp", *message)
	err := client.SendMessage(forwarder.sock, *message)
	if err != nil {
		logging.Error(forwarder.logger).Log(logging.MessageKey(), "failed to send message", logging.ErrorKey(), err)
		return kratos.CreateErrorWRP(message.TransactionUUID, message.Source, message.Destination, http.StatusServiceUnavailable, err)
	}
	return nil
}

func (forwarder *Forwarder) Close() {
	forwarder.stopTicker <- struct{}{}
	err := forwarder.sock.Close()
	if err != nil {
		logging.Error(forwarder.logger).Log(logging.MessageKey(), "failed to close socket", logging.ErrorKey(), err)
	}
}
