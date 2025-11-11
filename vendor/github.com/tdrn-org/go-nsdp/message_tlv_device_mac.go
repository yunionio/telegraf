// message_tlv_device_mac.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"fmt"
	"net"
)

// TLV to exchange the target device's MAC address.
//
// Add an empty DeviceMAC TLV to a read request to get a filled one back.
type DeviceMAC struct {
	MAC net.HardwareAddr // Device MAC
}

func EmptyDeviceMAC() *DeviceMAC {
	return NewDeviceMAC(net.HardwareAddr{})
}

func NewDeviceMAC(mac net.HardwareAddr) *DeviceMAC {
	return &DeviceMAC{MAC: mac}
}

func unmarshalDeviceMAC(bytes []byte) (*DeviceMAC, error) {
	len := len(bytes)
	if len == 0 {
		return EmptyDeviceMAC(), nil
	}
	if len != 6 {
		return nil, fmt.Errorf("unexpected device MAC length: %d", len)
	}
	return NewDeviceMAC(net.HardwareAddr(bytes)), nil
}

func (tlv *DeviceMAC) Type() Type {
	return TypeDeviceMAC
}

func (tlv *DeviceMAC) Length() uint16 {
	return uint16(len(tlv.MAC))
}

func (tlv *DeviceMAC) Value() []byte {
	return tlv.MAC
}

func (tlv *DeviceMAC) String() string {
	return fmt.Sprintf("DeviceMAC(%04xh) %s", TypeDeviceMAC, tlv.MAC)
}
