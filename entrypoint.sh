# SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
# SPDX-License-Identifier: Apache-2.0
#!/bin/sh

set -e

parodus_port=16014
aker_port=16015
mocktr181_port=16016
req_res_port=13032

if [[ -z "${URL}" ]]; then
    URL="http://petasos:6400"
fi

if [[ -z "${FIRMWARE}" ]]; then
    FIRMWARE="mock-rdkb-firmware"
fi

if [[ -z "${BOOT_TIME}" ]]; then
    BOOT_TIME=`date +%s`
fi

if [[ -z "${HW_MANUFACTURER}" ]]; then
    HW_MANUFACTURER="Example Inc."
fi

if [[ -z "${REBOOT_REASON}" ]]; then
    REBOOT_REASON="unknown"
fi

if [[ -z "${SERIAL_NUMBER}" ]]; then
    SERIAL_NUMBER="mock-rdkb-simulator"
fi

if [[ -z "${PARTNER_ID}" ]]; then
    PARTNER_ID="comcast"
fi

if [[ -z "${CMAC}" ]]; then
    CMAC="112233445566"

    if [[ ! -z "${RANDOM_MAC}" ]]; then
        CMAC=`hexdump -n 6 -v -e '"" 12/1 "%02X" "\n"' /dev/urandom`
    fi
fi

if [[ -z "${INTERFACE}" ]]; then
    INTERFACE="eth0"
fi

if [[ -z "${TIMEOUT}" ]]; then
  TIMEOUT=130
fi

/parodus --hw-model=aker-testing \
        --hw-serial-number=$SERIAL_NUMBER \
        --hw-manufacturer=$HW_MANUFACTURER \
        --hw-mac=$CMAC \
        --hw-last-reboot-reason=$REBOOT_REASON \
        --fw-name=$FIRMWARE \
        --boot-time=$BOOT_TIME \
        --partner-id=$PARTNER_ID \
        --parodus-local-url=tcp://127.0.0.1:$parodus_port \
        --xmidt-url=$URL \
        --xmidt-ping-timeout=$TIMEOUT \
        --xmidt-backoff-max=2 \
        --force-ipv4  &

P1=$!

sleep 10
aker -p tcp://127.0.0.1:$parodus_port \
     -c tcp://127.0.0.1:$aker_port \
     -w echo \
     -d /tmp/aker-data.msgpack \
     -f /tmp/aker-data.msgpack.md5 \
     -m 128 > /dev/null &
P2=$!

mock_tr181 -p $parodus_port \
           -c $mocktr181_port \
           -d /etc/mock_tr181.json > /dev/null &
P3=$!
wait $P1
