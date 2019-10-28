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
	"fmt"
	"github.com/spf13/pflag"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog"
	"go.uber.org/fx"
	"os"
	"runtime"
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
