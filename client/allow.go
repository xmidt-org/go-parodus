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

package client

import (
	"errors"
	"github.com/xmidt-org/wrp-go/wrp"
)

var (
	ErrInvalidPartnerID = errors.New("partner id does not match")
)

// AllowMessage is the interface to block request that come from talaria.
// If an error is returned, an unauthorized message will be sent with the error payload
type AllowMessage interface {
	Allow(msg wrp.Message) error
}

// AllowMessageFunc is the Allow function that a AllowMessage has.
type AllowMessageFunc func(msg wrp.Message) error

// Allow runs the allowMessageFunc, making an AllowMessageFunc also an AllowMessage.
func (allowFunc AllowMessageFunc) Allow(msg wrp.Message) error {
	return allowFunc(msg)
}

// BlockByPartnerID is a AllowMessageFunc to block request that don't have the associated device partnerID with
// the wrp request.
func BlockByPartnerID(partnerID string) AllowMessageFunc {
	return func(msg wrp.Message) error {
		for _, id := range msg.PartnerIDs {
			if id == partnerID {
				return nil
			}
		}
		return ErrInvalidPartnerID
	}
}
