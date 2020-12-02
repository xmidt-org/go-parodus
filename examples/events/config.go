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
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/fx"
	"net/http"
	"time"
)

func Provide() (client.ClientConfig, *App) {
	logger := logging.New(&logging.Options{
		File:  "stdout",
		JSON:  true,
		Level: "DEBUG",
	})

	app := &App{
		logger:      logger,
		stopSending: make(chan struct{}, 1),
	}

	return client.ClientConfig{
		Name:       "config",
		ParodusURL: "tcp://127.0.0.1:6666",
		ServiceURL: "tcp://127.0.0.1:13031",
		Debug:      true,
		Logger:     logger,
		MSGHandler: app,
		Register:   time.Minute,
	}, app
}

type App struct {
	logger      log.Logger
	out         client.SendMessageHandler
	stopSending chan struct{}
}

func (app *App) HandleMessage(msg *wrp.Message) *wrp.Message {
	return kratos.CreateErrorWRP(msg.TransactionUUID, msg.Source, msg.Destination, http.StatusBadRequest, errors.New("interface not implemented"))
}

func (app *App) Close() {
	// Do nothing, to make interface happy
}

func BeginMessaging(app *App, handler client.SendMessageHandler, lc fx.Lifecycle) {
	app.out = handler

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			go app.sendEvents()
			return nil
		},
		OnStop: func(c context.Context) error {
			app.stopSending <- struct{}{}
			return nil
		},
	})

}

func (app *App) sendEvents() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-app.stopSending:
			ticker.Stop()
			return
		case <-ticker.C:
			err := app.out.SendMessage(wrp.Message{
				Type:        wrp.SimpleEventMessageType,
				Source:      "events",
				Destination: "event:go-parodus",
				ContentType: "application/json",
				Payload:     []byte(fmt.Sprintf(`{"time":"%s"}`, time.Now())),
			}, context.Background())
			if err != nil {
				logging.Error(app.logger).Log(logging.MessageKey(), "Failed to send message", logging.ErrorKey(), err)
			} else {
				logging.Debug(app.logger).Log(logging.MessageKey(), "Sent message")
			}
		}
	}
}
