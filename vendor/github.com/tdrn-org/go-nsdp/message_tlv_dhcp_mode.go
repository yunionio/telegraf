// message_tlv_port_statistic.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"fmt"
)

// TLV to exchange the target device's DHCP mode.
//
// Add an empty DHCPMode TLV to a read request to get a filled one back.
type DHCPMode struct {
	Mode uint8 // DHCP mode (0: disabled, 1: enabled)
}

const dhcpModeLen uint16 = 1

func EmptyDHCPMode() *DHCPMode {
	return NewDHCPMode(0)
}

func NewDHCPMode(mode uint8) *DHCPMode {
	return &DHCPMode{Mode: mode}
}

func unmarshalDHCPMode(value []byte) (*DHCPMode, error) {
	len := len(value)
	if len == 0 {
		return EmptyDHCPMode(), nil
	}
	if len != int(dhcpModeLen) {
		return nil, fmt.Errorf("unexpected dhcp mode length: %d", len)
	}
	return NewDHCPMode(value[0]), nil
}

func (tlv *DHCPMode) Type() Type {
	return TypeDHCPMode
}

func (tlv *DHCPMode) Length() uint16 {
	return uint16(dhcpModeLen)
}

func (tlv *DHCPMode) Value() []byte {
	value := make([]byte, dhcpModeLen)
	value[0] = tlv.Mode
	return value
}

func (tlv *DHCPMode) String() string {
	return fmt.Sprintf("DHCPMode(%04xh) %s", TypeDHCPMode, tlv.ModeString())
}

// ModeString returns a textual representation of the mode value.
func (tlv *DHCPMode) ModeString() string {
	switch tlv.Mode {
	case 0:
		return "Disabled"
	case 1:
		return "Enabled"
	}
	return fmt.Sprintf("%02xh", tlv.Mode)
}
