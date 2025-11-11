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

// TLV to exchange the target device's firmware version for firmware slot 2.
//
// Add an empty FWVersionSlot2 TLV to a read request to get a filled one back.
type FWVersionSlot2 struct {
	Version string // Slot 2 version (e.g. 2.06.17)
}

func EmptyFWVersionSlot2() *FWVersionSlot2 {
	return NewFWVersionSlot2("")
}

func NewFWVersionSlot2(version string) *FWVersionSlot2 {
	return &FWVersionSlot2{Version: version}
}

func unmarshalFWVersionSlot2(bytes []byte) (*FWVersionSlot2, error) {
	return NewFWVersionSlot2(string(bytes)), nil
}

func (tlv *FWVersionSlot2) Type() Type {
	return TypeFWVersionSlot2
}

func (tlv *FWVersionSlot2) Length() uint16 {
	return uint16(len(tlv.Version))
}

func (tlv *FWVersionSlot2) Value() []byte {
	return []byte(tlv.Version)
}

func (tlv *FWVersionSlot2) String() string {
	return fmt.Sprintf("FWVersionSlot2(%04xh) '%s'", TypeFWVersionSlot2, tlv.Version)
}
