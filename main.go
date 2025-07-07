// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/pflag"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog"
	"go.uber.org/fx"
)

const (
	applicationName = "parodus"
)

var (
	GitCommit = "undefined"
	Version   = "undefined"
	BuildTime = "undefined"
)

func main() {
	app := fx.New(
		xlog.Logger(),
		config.CommandLine{Name: applicationName}.Provide(SetupFlagSet),
		fx.Provide(
			ProvideVersionPrintFunc,
			Provide,
			config.ProvideViper(),
			xlog.Unmarshal("log"),
			StartUpstreamConnection,
		),
		fx.Invoke(
			StartParodus,
		),
	)

	err := app.Err()
	if errors.Is(err, pflag.ErrHelp) {
		return
	} else if errors.Is(err, nil) {
		app.Run()
	}

	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}

func ProvideVersionPrintFunc() func() {
	return func() {
		fmt.Fprintf(os.Stdout, "%s:\n", applicationName)
		fmt.Fprintf(os.Stdout, "  version: \t%s\n", Version)
		fmt.Fprintf(os.Stdout, "  go version: \t%s\n", runtime.Version())
		fmt.Fprintf(os.Stdout, "  built time: \t%s\n", BuildTime)
		fmt.Fprintf(os.Stdout, "  git commit: \t%s\n", GitCommit)
		fmt.Fprintf(os.Stdout, "  os/arch: \t%s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}
