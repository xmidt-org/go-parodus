// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog"
	"go.uber.org/fx"
	"os"
)

func main() {
	app := fx.New(
		xlog.Logger(),
		config.CommandLine{Name: "events"}.Provide(client.SetupFlagSet),
		fx.Provide(
			config.ProvideViper(),
			xlog.Unmarshal("log"),
			Provide,
			client.StartClient,
		),
		fx.Invoke(
			BeginMessaging,
		),
	)

	switch err := app.Err(); err {
	case pflag.ErrHelp:
		return
	case nil:
		app.Run()
	default:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
