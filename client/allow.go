// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"github.com/xmidt-org/wrp-go/v3"
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
