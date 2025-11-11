// message_eom.go
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
	"strings"
)

type MessageMarker uint32

const (
	EOMMarker MessageMarker = 0xffff0000
)

// EOM (end-of-message) marker terminating any NSDP message.
type EOM struct {
	Marker uint32
}

func newEOM() *EOM {
	return &EOM{
		Marker: uint32(EOMMarker),
	}
}

func (m *EOM) writeString(builder *strings.Builder) {
	fmt.Fprintf(builder, "EOM   : %08xh", m.Marker)
}

func (m *EOM) marshalBuffer(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, m.Marker)
}
