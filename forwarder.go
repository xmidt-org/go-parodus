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
	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/wrp-go/wrp"
	"nanomsg.org/go/mangos/v2"
	"net/http"
	"time"
)

// Forwarder struct forwards messages coming from Talaria down to the libparouds clients
//
type Forwarder struct {
	Name      string
	URL       string
	LastAlive time.Time
	logger    log.Logger

	sock mangos.Socket
}

func CreateServiceForwarder(name string, url string, logger log.Logger) (*Forwarder, error) {
	sock, err := client.CreatePushSocket(url)
	if err != nil {
		return nil, err
	}
	return &Forwarder{
		Name:      name,
		URL:       url,
		LastAlive: time.Now(),
		sock:      sock,
		logger:    log.WithPrefix(logger, "forwarder", name),
	}, nil
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
	err := forwarder.sock.Close()
	if err != nil {
		logging.Error(forwarder.logger).Log(logging.MessageKey(), "failed to close socket", logging.ErrorKey(), err)
	}
}
