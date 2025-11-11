// message_tlv_port_statistic.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TLV to exchange the target device's port statistic.
//
// Add an empty PortStatistic TLV to a read request to receive a filled one for each of the device's port.
type PortStatistic struct {
	Port       uint8  // The number of the port this statistic refers to
	Received   uint64 // Number of received bytes
	Sent       uint64 // Number of sent bytes
	Packets    uint64 // Number of processed packets
	Broadcasts uint64 // Number of processed broadcasts
	Multicasts uint64 // Number of processed multicasts
	Errors     uint64 // Number of encountered errors
}

const portStatisticLen uint16 = 49

func EmptyPortStatistic() *PortStatistic {
	return &PortStatistic{}
}

func NewPortStatistic(port uint8, received uint64, sent uint64, packets uint64, broadcasts uint64, multicasts uint64, errors uint64) *PortStatistic {
	return &PortStatistic{
		Port:       port,
		Received:   received,
		Sent:       sent,
		Packets:    packets,
		Broadcasts: broadcasts,
		Multicasts: multicasts,
		Errors:     errors,
	}
}

func unmarshalPortStatistic(value []byte) (*PortStatistic, error) {
	len := len(value)
	if len == 0 {
		return EmptyPortStatistic(), nil
	}
	if len != int(portStatisticLen) {
		return nil, fmt.Errorf("unexpected port statistic length: %d", len)
	}
	buffer := bytes.NewBuffer(value)
	tlv := EmptyPortStatistic()
	tlv.Port, _ = buffer.ReadByte()
	binary.Read(buffer, binary.BigEndian, &tlv.Received)
	binary.Read(buffer, binary.BigEndian, &tlv.Sent)
	binary.Read(buffer, binary.BigEndian, &tlv.Packets)
	binary.Read(buffer, binary.BigEndian, &tlv.Broadcasts)
	binary.Read(buffer, binary.BigEndian, &tlv.Multicasts)
	binary.Read(buffer, binary.BigEndian, &tlv.Errors)
	return tlv, nil
}

func (tlv *PortStatistic) Type() Type {
	return TypePortStatistic
}

func (tlv *PortStatistic) Length() uint16 {
	return uint16(portStatisticLen)
}

func (tlv *PortStatistic) Value() []byte {
	buffer := &bytes.Buffer{}
	buffer.Grow(int(portStatisticLen))
	buffer.WriteByte(tlv.Port)
	binary.Write(buffer, binary.BigEndian, tlv.Received)
	binary.Write(buffer, binary.BigEndian, tlv.Sent)
	binary.Write(buffer, binary.BigEndian, tlv.Packets)
	binary.Write(buffer, binary.BigEndian, tlv.Broadcasts)
	binary.Write(buffer, binary.BigEndian, tlv.Multicasts)
	binary.Write(buffer, binary.BigEndian, tlv.Errors)
	return buffer.Bytes()
}

func (tlv *PortStatistic) String() string {
	return fmt.Sprintf("PortStatistic(%04xh) Port%d Received: %d, Sent: %d, Packets: %d, Broadcasts: %d, Multicasts: %d, Errors: %d", TypePortStatistic, tlv.Port, tlv.Received, tlv.Sent, tlv.Packets, tlv.Broadcasts, tlv.Multicasts, tlv.Errors)
}
