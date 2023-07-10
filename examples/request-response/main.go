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
	"os"

	"github.com/spf13/pflag"
	"github.com/xmidt-org/go-parodus/client"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/config"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		sallust.WithLogger(),
		config.CommandLine{Name: "request-response"}.Provide(client.SetupFlagSet),
		fx.Provide(
			config.ProvideViper(),
			//TODO: add func for logger
			Provide,
			func(u config.Unmarshaller) (c sallust.Config, err error) {
				err = u.UnmarshalKey("log", &c)
				return
			},
		),
		fx.Invoke(
			client.StartClient,
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
