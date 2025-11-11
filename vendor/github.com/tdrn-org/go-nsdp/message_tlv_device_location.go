// message_tlv_device_location.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"fmt"
)

// TLV to exchange the target device's location.
//
// Add an empty DeviceLocation TLV to a read request to get a filled one back.
type DeviceLocation struct {
	Location string // Device location text
}

func EmptyDeviceLocation() *DeviceLocation {
	return NewDeviceLocation("")
}

func NewDeviceLocation(location string) *DeviceLocation {
	return &DeviceLocation{Location: location}
}

func unmarshalDeviceLocation(bytes []byte) (*DeviceLocation, error) {
	return NewDeviceLocation(string(bytes)), nil
}

func (tlv *DeviceLocation) Type() Type {
	return TypeDeviceLocation
}

func (tlv *DeviceLocation) Length() uint16 {
	return uint16(len(tlv.Location))
}

func (tlv *DeviceLocation) Value() []byte {
	return []byte(tlv.Location)
}

func (tlv *DeviceLocation) String() string {
	return fmt.Sprintf("DeviceLocation(%04xh) '%s'", TypeDeviceLocation, tlv.Location)
}
