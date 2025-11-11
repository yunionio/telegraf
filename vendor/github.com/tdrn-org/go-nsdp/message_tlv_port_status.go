// message_tlv_port_statistic.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"bytes"
	"fmt"
)

// TLV to exchange the target device's port status.
//
// Add an empty PortStatus TLV to a read request to receive a filled one for each of the device's port.
type PortStatus struct {
	Port     uint8 // The number of the port this status refers to
	Status   uint8 // The port's status (0: down, otherwise up)
	Unknown1 uint8
}

const portStatusLen uint16 = 3

func EmptyPortStatus() *PortStatus {
	return &PortStatus{}
}

func NewPortStatus(port uint8, status uint8) *PortStatus {
	return &PortStatus{
		Port:   port,
		Status: status,
	}
}

func unmarshalPortStatus(value []byte) (*PortStatus, error) {
	len := len(value)
	if len == 0 {
		return EmptyPortStatus(), nil
	}
	if len != int(portStatusLen) {
		return nil, fmt.Errorf("unexpected port status length: %d", len)
	}
	buffer := bytes.NewBuffer(value)
	tlv := EmptyPortStatus()
	tlv.Port, _ = buffer.ReadByte()
	tlv.Status, _ = buffer.ReadByte()
	tlv.Unknown1, _ = buffer.ReadByte()
	return tlv, nil
}

func (tlv *PortStatus) Type() Type {
	return TypePortStatus
}

func (tlv *PortStatus) Length() uint16 {
	return uint16(portStatusLen)
}

func (tlv *PortStatus) Value() []byte {
	buffer := &bytes.Buffer{}
	buffer.Grow(int(portStatusLen))
	buffer.WriteByte(tlv.Port)
	buffer.WriteByte(tlv.Status)
	buffer.WriteByte(tlv.Unknown1)
	return buffer.Bytes()
}

func (tlv *PortStatus) String() string {
	return fmt.Sprintf("PortStatus(%04xh) Port%d Status: %s Unknown1: %02xh", TypePortStatus, tlv.Port, tlv.StatusString(), tlv.Unknown1)
}

// StatusString returns a textual representation of the status value.
func (tlv *PortStatus) StatusString() string {
	switch tlv.Status {
	case 0:
		return "Disconnected"
	case 1:
		return "10Mbit/half-duplex"
	case 2:
		return "10Mbit/full-duplex"
	case 3:
		return "100Mbit/half-duplex"
	case 4:
		return "100Mbit/full-duplex"
	case 5:
		return "1Gbit/full-duplex"
	}
	return fmt.Sprintf("%02xh", tlv.Status)
}
