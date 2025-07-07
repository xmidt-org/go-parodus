// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"time"

	"github.com/go-kit/log"
	libparodus "github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/webpa-common/v2/logging" // nolint:staticcheck
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/fx"
	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/pull"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

type Parodus struct {
	sock   mangos.Socket
	logger log.Logger

	client       kratos.Client
	stopHandling chan struct{}
}

func StartParodus(config Config, client kratos.Client, lc fx.Lifecycle, logger log.Logger) error {
	var sock mangos.Socket
	var err error

	if sock, err = pull.NewSocket(); err != nil {
		logging.Error(logger).Log(logging.MessageKey(), "can't get new pull socket", logging.ErrorKey(), err)
		return err
	}
	if err = sock.Listen(config.LocalURL); err != nil {
		logging.Error(logger).Log(logging.MessageKey(), "can't listen on new pull socket", logging.ErrorKey(), err, "url", config.LocalURL)
		return err
	}
	logging.Info(logger).Log(logging.MessageKey(), "Parodus Config", "config", config)
	sock.SetPipeEventHook(func(event mangos.PipeEvent, pipe mangos.Pipe) {
		logging.Info(logger).Log(logging.MessageKey(), "parodus pull socket event", "event", event, "pipe", pipe)
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
	logging.Debug(p.logger).Log(logging.MessageKey(), "Starting msgHandler")
	defer func() {
		logging.Debug(p.logger).Log(logging.MessageKey(), "msgHandler has stopped")
	}()
	for {
		select {
		case <-p.stopHandling:
			return
		case msg := <-wrpBus:
			switch msg.Type {
			case wrp.ServiceRegistrationMessageType:
				logging.Debug(p.logger).Log(logging.MessageKey(), "received service registration", "url", msg.URL, "name", msg.ServiceName)

				if handler, err := p.client.HandlerRegistry().GetHandler(msg.ServiceName); err != nil {
					// TODO: create timer for keep alive
					service, err := CreateServiceForwarder(msg.ServiceName, msg.URL, p.logger)
					if err != nil {
						logging.Error(p.logger).Log(logging.MessageKey(), "failed to send message to talaria", logging.ErrorKey(), err)
					}
					err = p.client.HandlerRegistry().Add(service.Name, service)
					if err != nil {
						logging.Error(p.logger).Log(logging.MessageKey(), "failed to add service to registry", logging.ErrorKey(), err)
					}
				} else {
					// update handler timestamp
					if forwarder, ok := handler.(*Forwarder); ok {
						forwarder.LastAlive = time.Now()
						logging.Debug(p.logger).Log(logging.MessageKey(), "updated registration timestamp", "url", msg.URL, "name", msg.ServiceName)
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
						logging.Debug(p.logger).Log(logging.MessageKey(), "updated registration timestamp", "url", msg.URL, "name", msg.ServiceName)
					}
				}
			default:
				logging.Error(p.logger).Log(logging.MessageKey(), "Unexpected WRP Message. Please file an issue at github.com/xmidt-org/go-parodus/issues", "wrp", msg)
			}
		}
	}
}
