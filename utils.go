// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import "regexp"

func validateMAC(mac string) bool {
	var (
		macAddrRE    *regexp.Regexp
		macAddrSlice []string
	)

	macAddrRE = regexp.MustCompile(`(([0-9A-Fa-f]{2}(?:[:-]?)){5}[0-9A-Fa-f]{2})|(([0-9A-Fa-f]{4}\.){2}[0-9A-Fa-f]{4})`)
	macAddrSlice = macAddrRE.FindAllString(mac, -1)

	return len(macAddrSlice) >= 1
}
