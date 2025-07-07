// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"time"

	"github.com/xmidt-org/kratos" // nolint:staticcheck
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func StartUpstreamConnection(config Config, lc fx.Lifecycle, logger *zap.Logger) (kratos.Client, error) {
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
			logger.Error("msg", zap.Any("error", "Ping Miss"))
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
		logger.Error("failed to create client", zap.Error(err))
		if client != nil {
			closeErr := client.Close()
			logger.Info("failed to close bad client", zap.Error(closeErr))
		}
	}

	logger.Info("kratos client created")
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
