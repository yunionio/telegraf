// message_tlv_device_name.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"fmt"
)

// TLV to exchange the target device's name.
//
// Add an empty DeviceName TLV to a read request to get a filled one back.
type DeviceName struct {
	Name string // Device Name
}

func EmptyDeviceName() *DeviceName {
	return NewDeviceName("")
}

func NewDeviceName(name string) *DeviceName {
	return &DeviceName{Name: name}
}

func unmarshalDeviceName(bytes []byte) (*DeviceName, error) {
	return NewDeviceName(string(bytes)), nil
}

func (tlv *DeviceName) Type() Type {
	return TypeDeviceName
}

func (tlv *DeviceName) Length() uint16 {
	return uint16(len(tlv.Name))
}

func (tlv *DeviceName) Value() []byte {
	return []byte(tlv.Name)
}

func (tlv *DeviceName) String() string {
	return fmt.Sprintf("DeviceName(%04xh) '%s'", TypeDeviceName, tlv.Name)
}
