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
	"context"
	"time"

	libparodus "github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/pull"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

type Parodus struct {
	sock   mangos.Socket
	logger *zap.Logger

	client       kratos.Client
	stopHandling chan struct{}
}

func StartParodus(config Config, client kratos.Client, lc fx.Lifecycle, logger *zap.Logger) error {
	var sock mangos.Socket
	var err error

	if sock, err = pull.NewSocket(); err != nil {
		logger.Error("can't get new pull socket", zap.Error(err))
		return err
	}
	if err = sock.Listen(config.LocalURL); err != nil {
		logger.Error("can't listen on new pull socket", zap.String("url", config.LocalURL))
		return err
	}
	logger.Info("Parodus Config", zap.Any("config", config))  // would reflection be ok here or should we implement an object marshaler interface?
	sock.SetPipeEventHook(func(event mangos.PipeEvent, pipe mangos.Pipe) {
		logger.Info("parodus pull socket event", zap.Any("event", event), zap.Any("pipe", pipe))
	})

	parodus := &Parodus{
		sock:         sock,
		logger:       logger,
		client:       client,
		stopHandling: make(chan struct{}),
	}

	dataBus := make(chan []byte, 100)
	wrpBus := make(chan wrp.Message, 100)
	stopReading := make(chan struct{})
	stopParsing := make(chan struct{})
	stopRouting := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(context context.Context) error {
			go libparodus.ReadPump(parodus.sock, dataBus, logger)
			go libparodus.ParseBus(wrpBus, dataBus, stopParsing, logger)
			go parodus.msgHandler(wrpBus)
			return nil
		},
		OnStop: func(context context.Context) error {
			close(stopReading)
			close(stopParsing)
			close(stopRouting)
			close(parodus.stopHandling)
			return parodus.sock.Close()
		},
	})

	return nil
}

func (p *Parodus) msgHandler(wrpBus chan wrp.Message) {
	p.logger.Debug("Starting msgHandler")
	defer func() {
		p.logger.Debug("msgHandler has stopped")
	}()
	for {
		select {
		case <-p.stopHandling:
			return
		case msg := <-wrpBus:
			switch msg.Type {
			case wrp.ServiceRegistrationMessageType:
				p.logger.Debug("received service registration", zap.String("url", msg.URL), zap.String("name", msg.ServiceName))

				if handler, err := p.client.HandlerRegistry().GetHandler(msg.ServiceName); err != nil {
					// TODO: create timer for keep alive
					service, err := CreateServiceForwarder(msg.ServiceName, msg.URL, p.logger)
					if err != nil {
						p.logger.Debug("failed to send message to talaria", zap.Error(err))
					}
					err = p.client.HandlerRegistry().Add(service.Name, service)
					if err != nil {
						p.logger.Error("failed to add service to registry", zap.Error(err))
					}
				} else {
					// update handler timestamp
					if forwarder, ok := handler.(*Forwarder); ok {
						forwarder.LastAlive = time.Now()
						p.logger.Debug("updated registration timestamp", zap.String("url", msg.URL), zap.String("name", msg.ServiceName))
					}
				}
			case wrp.SimpleRequestResponseMessageType:
				// Send message to Talaria
				p.client.Send(&msg)
			case wrp.SimpleEventMessageType:
				// Send event to Talaria
				p.client.Send(&msg)
			case wrp.ServiceAliveMessageType:
				// TODO: reset timer(timer should also be created
				if handler, err := p.client.HandlerRegistry().GetHandler(msg.ServiceName); err == nil {
					if forwarder, ok := handler.(*Forwarder); ok {
						forwarder.LastAlive = time.Now()
						p.logger.Debug("updated registration timestamp", zap.String("url", msg.URL), zap.String("name", msg.ServiceName))
					}
				}
			default:
				p.logger.Error("Unexpected WRP Message. Please file an issue at github.com/xmidt-org/go-parodus/issues", zap.Any("wrp", msg))
			}
		}
	}
}
