// message_tlv_device_model.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"fmt"
)

// TLV to exchange the target device's model name.
//
// Add an empty DeviceModel TLV to a read request to get a filled one back.
type DeviceModel struct {
	Model string // Model name (e.g. GS108Ev3)
}

func EmptyDeviceModel() *DeviceModel {
	return NewDeviceModel("")
}

func NewDeviceModel(model string) *DeviceModel {
	return &DeviceModel{Model: model}
}

func unmarshalDeviceModel(bytes []byte) (*DeviceModel, error) {
	return NewDeviceModel(string(bytes)), nil
}

func (tlv *DeviceModel) Type() Type {
	return TypeDeviceModel
}

func (tlv *DeviceModel) Length() uint16 {
	return uint16(len(tlv.Model))
}

func (tlv *DeviceModel) Value() []byte {
	return []byte(tlv.Model)
}

func (tlv *DeviceModel) String() string {
	return fmt.Sprintf("DeviceModel(%04xh) '%s'", TypeDeviceModel, tlv.Model)
}
