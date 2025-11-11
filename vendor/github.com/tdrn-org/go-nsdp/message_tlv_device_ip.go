// message_tlv_device_ip.go
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

// TLV to exchange the target device's IP address.
//
// Add an empty DeviceIP TLV to a read request to get a filled one back.
type DeviceIP struct {
	IP net.IP // Device IP
}

func EmptyDeviceIP() *DeviceIP {
	return NewDeviceIP(net.IP{})
}

func NewDeviceIP(ip net.IP) *DeviceIP {
	return &DeviceIP{IP: ip}
}

func unmarshalDeviceIP(bytes []byte) (*DeviceIP, error) {
	len := len(bytes)
	if len == 0 {
		return EmptyDeviceIP(), nil
	}
	if len != 4 && len != 16 {
		return nil, fmt.Errorf("unexpected device IP length: %d", len)
	}
	return NewDeviceIP(net.IP(bytes)), nil
}

func (tlv *DeviceIP) Type() Type {
	return TypeDeviceIP
}

func (tlv *DeviceIP) Length() uint16 {
	return uint16(len(tlv.IP))
}

func (tlv *DeviceIP) Value() []byte {
	return tlv.IP
}

func (tlv *DeviceIP) String() string {
	return fmt.Sprintf("DeviceIP(%04xh) %s", TypeDeviceIP, tlv.IP)
}
