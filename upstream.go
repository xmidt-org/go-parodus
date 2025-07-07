// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/xmidt-org/kratos"
	"github.com/xmidt-org/webpa-common/v2/logging" // nolint:staticcheck
	"go.uber.org/fx"
)

func StartUpstreamConnection(config Config, lc fx.Lifecycle, logger log.Logger) (kratos.Client, error) {
	queueConfig := kratos.QueueConfig{
		MaxWorkers: 5,
		Size:       100,
	}

	client, err := kratos.NewClient(kratos.ClientConfig{
		DeviceName:           config.DeviceID,
		FirmwareName:         config.FirmwareName,
		ModelName:            config.HardwareModel,
		Manufacturer:         config.HardwareManufacturer,
		DestinationURL:       config.URL,
		OutboundQueue:        queueConfig,
		WRPEncoderQueue:      queueConfig,
		WRPDecoderQueue:      queueConfig,
		HandlerRegistryQueue: queueConfig,
		HandleMsgQueue:       queueConfig,
		Handlers:             []kratos.HandlerConfig{},
		HandlePingMiss: func() error {
			logging.Error(logger).Log(logging.MessageKey(), "Ping Miss")
			// TODO: handle reconnect
			return nil
		},
		ClientLogger: logger,
		PingConfig: kratos.PingConfig{
			PingWait:    time.Second * time.Duration(config.PingTimeout),
			MaxPingMiss: 3,
		},
	})
	if err != nil {
		logging.Error(logger).Log(logging.MessageKey(), "failed to create client", logging.ErrorKey(), err)
		if client != nil {
			closeErr := client.Close()
			logging.Info(logger).Log(logging.MessageKey(), "failed to close bad client", logging.ErrorKey(), closeErr)
		}
	}

	logging.Info(logger).Log(logging.MessageKey(), "kratos client created")
	lc.Append(fx.Hook{
		OnStart: func(context context.Context) error {
			return nil
		},
		OnStop: func(context context.Context) error {
			return client.Close()
		},
	})
	return client, nil
}
