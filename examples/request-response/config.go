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
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/wrp-go/v3"
	"go.uber.org/zap"

	"net/http"
)

func Provide(fs *pflag.FlagSet, logger *zap.Logger) client.ClientConfig {
	app := &App{
		Data:   map[string]ConfigSet{},
		logger: logger,
	}
	parodusURL, err := fs.GetString("parodus-local-url")
	if err != nil {
		parodusURL = "tcp://127.0.0.1:6666"
	}
	serviceURL, err := fs.GetString("service-url")
	if err != nil {
		serviceURL = "tcp://127.0.0.1:6666"
	}
	debug, _ := fs.GetBool("debug")

	return client.ClientConfig{
		Name:       "config",
		ParodusURL: parodusURL,
		ServiceURL: serviceURL,
		Debug:      debug,
		Logger:     logger,
		MSGHandler: app,
		Register:   time.Minute,
	}
}

// ConfigSet holds the in-memory configuration
type ConfigSet struct {
	Value    interface{} `json:"value,omitempty"`
	DataType int         `json:"dataType,omitempty"`
}

type App struct {
	Data   map[string]ConfigSet
	logger *zap.Logger
}

func (app *App) HandleMessage(msg *wrp.Message) *wrp.Message {
	app.logger.Debug("Received Message", zap.Any("wrp", msg))
	switch msg.Type {
	case wrp.SimpleRequestResponseMessageType:
		app.logger.Debug("working on message", zap.String("uuid", msg.TransactionUUID), zap.Any("source", msg.Source))
		var request ConfigRequest
		err := json.Unmarshal(msg.Payload, &request)
		if err != nil {
			return kratos.CreateErrorWRP(msg.TransactionUUID, msg.Source, msg.Destination, http.StatusBadRequest, err)
		}

		switch request.Command {
		case SETOPER:
			response := client.CreateResponseWRP(msg)
			configResponse := app.handleSet(request)
			response.ContentType = "application/json"
			response.SetStatus(int64(configResponse.StatusCode))
			payload, _ := json.Marshal(configResponse)
			response.Payload = payload
			return response
		case GETOPER:
			response := client.CreateResponseWRP(msg)
			configResponse := app.handleGet(request)
			response.ContentType = "application/json"
			response.SetStatus(int64(configResponse.StatusCode))
			payload, _ := json.Marshal(&configResponse)
			response.Payload = payload
			return response
		default:
			return kratos.CreateErrorWRP(msg.TransactionUUID, msg.Source, msg.Destination, http.StatusBadRequest, fmt.Errorf("unknown command %s", request.Command))
		}
	default:
		return kratos.CreateErrorWRP(msg.TransactionUUID, msg.Source, msg.Destination, http.StatusBadRequest, fmt.Errorf("unhandled msg type request %s", msg.Type))
	}
}
func (app *App) Close() {
	// Do nothing to make interface happy
}

func (app *App) handleSet(request ConfigRequest) ConfigResponse {
	if len(request.Parameters) == 0 {
		return ConfigResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "no parameters to set",
		}
	}

	for _, p := range request.Parameters {
		var set ConfigSet

		set.DataType = p.DataType
		set.Value = p.Value
		app.Data[p.Name] = set
	}
	return ConfigResponse{
		StatusCode: http.StatusOK,
		Message:    "Success",
	}
}

func (app *App) handleGet(request ConfigRequest) ConfigResponse {
	response := ConfigResponse{
		StatusCode: http.StatusOK,
		Parameters: []ConfigParameters{},
		Message:    "Success",
	}

	for _, n := range request.Names {
		app.logger.Debug("working names", zap.String("name", n))

		cfgSet, ok := app.Data[n]

		if !ok {
			configParam := ConfigParameters{}
			configParam.Name = n
			configParam.Message = fmt.Sprintf("Value for %s is null", n)
			response.Parameters = append(response.Parameters, configParam)
			continue
		}

		response.Parameters = append(response.Parameters, ConfigParameters{
			Name:     n,
			Value:    cfgSet.Value,
			DataType: cfgSet.DataType,
			Message:  "Success",
		})
	}

	app.logger.Debug("done", zap.Int("params", len(response.Parameters)), zap.Any("map", app.Data))

	return response
}
