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
