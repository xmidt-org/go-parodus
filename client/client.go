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
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/pflag"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/pull"
	"nanomsg.org/go/mangos/v2/protocol/push"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

type SendMessageHandler interface {
	SendMessage(msg wrp.Message, c context.Context) error
}

type SendMessageHandlerFunc func(msg wrp.Message, c context.Context) error

func (sendMessageHandlerFunc SendMessageHandlerFunc) SendMessage(msg wrp.Message, c context.Context) error {
	return sendMessageHandlerFunc(msg, c)
}

type client struct {
	name string
	url  string

	logger      *zap.Logger
	parodusSock mangos.Socket
	serviceSock mangos.Socket

	stopParsing  chan struct{}
	stopSending  chan struct{}
	stopHandling chan struct{}

	msgHandler      kratos.DownstreamHandler
	parodusUpstream chan wrp.Message
}

type ClientConfig struct {
	Name       string
	ParodusURL string
	ServiceURL string
	Debug      bool
	Logger     *zap.Logger
	MSGHandler kratos.DownstreamHandler
	Register   time.Duration
}

func validateConfig(config *ClientConfig) error {
	if config.Name == "" {
		return errors.New("name must be set in config")
	}
	if config.ParodusURL == "" {
		return errors.New("parodusURL must be set in config")
	} else {
		u, err := url.Parse(config.ParodusURL)
		if err != nil {
			return err
		}
		if u.Scheme != "tcp" {
			return fmt.Errorf("invalid ParodusURL scheme: %s - only tcp:// urls are supported", u.Scheme)
		}
	}
	if config.ServiceURL == "" {
		return errors.New("serviceURL must be set in config")
	}
	if config.MSGHandler == nil {
		return errors.New("handler must be defined")
	}
	if config.Logger == nil {
		config.Logger = config.Logger.WithOptions(zap.Fields(zap.String("component", "libparodus")))
	}
	if config.Register == 0 {
		config.Register = time.Minute
	}

	return nil
}

func StartClient(config ClientConfig, lc fx.Lifecycle) (SendMessageHandler, error) {
	if err := validateConfig(&config); err != nil {
		return nil, err
	}
	client := client{
		name:            config.Name,
		url:             config.ServiceURL,
		logger:          config.Logger.WithOptions(zap.Fields(zap.String("component", "libparodus"))),
		stopParsing:     make(chan struct{}),
		stopSending:     make(chan struct{}),
		stopHandling:    make(chan struct{}),
		msgHandler:      config.MSGHandler,
		parodusUpstream: make(chan wrp.Message, 100),
	}
	// create push socket
	if parodusSock, err := push.NewSocket(); err != nil {
		return nil, fmt.Errorf("can't get new push socket: %s", err)
	} else {
		if err := parodusSock.DialOptions(config.ParodusURL, map[string]interface{}{mangos.OptionDialAsynch: false}); err != nil {
			return nil, fmt.Errorf("can't dial on push socket: %s", err)
		}
		client.parodusSock = parodusSock
	}
	// create pull socket
	if serviceSock, err := pull.NewSocket(); err != nil {
		return nil, fmt.Errorf("can't get new pull socket: %s", err)
	} else {
		client.logger.Info(fmt.Sprintf("listing on %s", config.ServiceURL))
		if err := serviceSock.Listen(config.ServiceURL); err != nil {
			return nil, fmt.Errorf("can't listen on pull socket: %s", err)
		}
		client.serviceSock = serviceSock
	}
	dataBus := make(chan []byte, 100)
	wrpBusRead := make(chan wrp.Message, 100)
	ticker := time.NewTicker(config.Register)
	stopTicker := make(chan struct{}, 1)
	lc.Append(fx.Hook{
		OnStart: func(context context.Context) error {
			go ReadPump(client.serviceSock, dataBus, client.logger)
			go WritePump(client.parodusSock, client.parodusUpstream, client.stopSending, client.logger)
			go ParseBus(wrpBusRead, dataBus, client.stopParsing, client.logger)
			go client.handleMSG(wrpBusRead, client.parodusUpstream)
			client.sendRegistration()
			go func() { // Send alive every tick
				for {
					select {
					case <-stopTicker:
						ticker.Stop()
						return
					case <-ticker.C:
						client.sendRegistration()
					}
				}
			}()
			return nil
		},
		OnStop: func(context context.Context) error {
			client.logger.Info("stopping client")
			close(client.stopSending)
			close(client.stopHandling)
			close(client.stopParsing)
			close(stopTicker)
			return nil
		},
	})
	return &client, nil
}

func (c *client) handleMSG(wrpBusIn chan wrp.Message, wrpBusOut chan wrp.Message) {
	for {
		select {
		case <-c.stopHandling:
			// TODO: should I do more stop logic?
			c.logger.Debug("stopping handleMSG")
			return
		case msg := <-wrpBusIn:
			c.logger.Debug("received msg", zap.String("UUID", msg.TransactionUUID))

			switch msg.Type {
			case wrp.ServiceAliveMessageType:
				wrpBusOut <- wrp.Message{
					Type:            wrp.ServiceAliveMessageType,
					Source:          msg.Destination,
					TransactionUUID: msg.TransactionUUID,
					Destination:     msg.Source,
					Payload:         []byte("I'm here!"),
					Headers:         msg.Headers,
					ContentType:     "text/plain",
					Spans:           msg.Spans,
					ServiceName:     msg.ServiceName,
				}
			default:
				wrpBusOut <- *c.msgHandler.HandleMessage(&msg)
			}

		}
	}
}

func (client *client) sendRegistration() {
	msg := wrp.Message{
		Type:        wrp.ServiceRegistrationMessageType,
		ServiceName: client.name,
		URL:         client.url,
	}
	client.parodusUpstream <- msg
}
func (client *client) SendMessage(msg wrp.Message, c context.Context) error {
	select {
	case <-c.Done():
		return context.Canceled
	case client.parodusUpstream <- msg:
		return nil
	}
}

func CreateResponseWRP(msg *wrp.Message) *wrp.Message {
	responseMSG := *msg
	source := responseMSG.Destination
	responseMSG.Destination = msg.Source
	responseMSG.Source = source
	return &responseMSG
}

// SetupFlagSet sets up some initial helper flags
func SetupFlagSet(fs *pflag.FlagSet) error {
	fs.StringP("parodus-local-url", "l", "tcp://127.0.0.1:6666", "Parodus local server url")
	fs.StringP("service-url", "s", "tcp://127.0.0.1:13032", "service local url to be used by parodus")

	fs.BoolP("debug", "", false, "enables debug logging")
	fs.BoolP("version", "v", false, "print version and exit")
	return nil
}
