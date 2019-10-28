# go-parodus

[![Build Status](https://travis-ci.com/xmidt-org/go-parodus.svg?branch=master)](https://travis-ci.com/xmidt-org/go-parodus)
[![codecov.io](http://codecov.io/github/xmidt-org/go-parodus/coverage.svg?branch=master)](http://codecov.io/github/xmidt-org/go-parodus?branch=master)
[![Code Climate](https://codeclimate.com/github/xmidt-org/go-parodus/badges/gpa.svg)](https://codeclimate.com/github/xmidt-org/go-parodus)
[![Issue Count](https://codeclimate.com/github/xmidt-org/go-parodus/badges/issue_count.svg)](https://codeclimate.com/github/xmidt-org/go-parodus)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/go-parodus)](https://goreportcard.com/report/github.com/xmidt-org/go-parodus)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/go-parodus/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/go-parodus.svg)](CHANGELOG.md)


## Summary

go-parodus is a golang implementation of the [parodus client](https://github.com/xmidt-org/parodus)


## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Build](#build)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). 
By participating, you agree to this Code.

## Details

### parodus
go-parodus has two main functions: 1)maintain the websocket connection with [talaria](https://github.com/xmidt-org/talaria).
Managing the websocket layer is handled via the [kratos library](https://github.com/xmidt-org/kratos) which was originally developed for testing purposes.
And 2) handle the nanomsg server with its clients. When a request comes from talaria, the wrp message is routed to the
clients. For more information on how Parodus work refer to the [Wiki](https://github.com/xmidt-org/parodus/wiki/Parodus-In-Detail)

Available Tags:
_note_: not all flags have been implemented yet
```
Usage of parodus:
  -b, --boot-time int                  the boot time in unix time (default 1571960392)
      --debug                          enables debug logging
  -4, --force-ipv4                     forcefully connect parodus to ipv4 address
  -6, --force-ipv6                     forcefully connect parodus to ipv6 address
  -n, --fw-name string                 firmware name and version currently running
  -r, --hw-last-reboot-reason string   the last known reboot reason
  -d, --hw-mac string                  the MAC address used to manage the device (default "unknown")
  -f, --hw-manufacturer string         the device manufacturer
  -m, --hw-model string                the hardware model name
  -s, --hw-serial-number string        the serial number
  -l, --parodus-local-url string       Parodus local server url (default "tcp://127.0.0.1:6666")
  -p, --partner-id string              partner ID of iot/gateway device
  -c, --ssl-cert-path string           provide the certs for establishing secure upstream
  -v, --version                        print version and exit
  -o, --xmidt-backoff-max int          the maximum value in seconds for the backoff algorithm (default 60)
  -i, --xmidt-interface-used string    the device interface being used to connect to the cloud (default "eth0")
  -t, --xmidt-ping-timeout int         the maximum time to wait between pings before assuming the upstream is broken (default 60)
  -u, --xmidt-url string               the hardware model name

```

### parodus clients
For creating a parodus client most of the work has already been done for you in the `libparodus` package by maintaining
the nanomsg client to parodus. The consumer of the package will need to implement the `kratos.DownstreamHandler` interface

#### Examples
For the following examples the XMiDT cluster must be up and running. For local testing I recommend standing up a [local
docker cluster](https://github.com/xmidt-org/xmidt/tree/master/deploy).
- request-response[examples/request-response/README.md] -> set and get information from a map
- event[examples/request-response/README.md] -> spam talaria with events generated from a client

## Build

### Source

In order to build from the source, you need a working Go environment with 
version 1.11 or greater. Find more information on the [Go website](https://golang.org/doc/install).

You can directly use `go get` to put the go-parodus binary into your `GOPATH`:
```bash
GO111MODULE=on go get github.com/xmidt-org/go-parodus
```

You can also clone the repository yourself and build using make:

```bash
mkdir -p $GOPATH/src/github.com/xmidt-org
cd $GOPATH/src/github.com/xmidt-org
git clone git@github.com:xmidt-org/go-parodus.git
cd go-parodus
go build .
```

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).
