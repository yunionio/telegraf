// message_tlv_router_ip.go
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

// TLV to exchange the target device's gateway address.
type RouterIP struct {
	IP net.IP
}

func EmptyRouterIP() *RouterIP {
	return NewRouterIP(net.IP{})
}

func NewRouterIP(ip net.IP) *RouterIP {
	return &RouterIP{IP: ip}
}

func unmarshalRouterIP(bytes []byte) (*RouterIP, error) {
	len := len(bytes)
	if len == 0 {
		return EmptyRouterIP(), nil
	}
	if len != 4 && len != 16 {
		return nil, fmt.Errorf("unexpected router IP length: %d", len)
	}
	return NewRouterIP(net.IP(bytes)), nil
}

func (tlv *RouterIP) Type() Type {
	return TypeRouterIP
}

func (tlv *RouterIP) Length() uint16 {
	return uint16(len(tlv.IP))
}

func (tlv *RouterIP) Value() []byte {
	return tlv.IP
}

func (tlv *RouterIP) String() string {
	return fmt.Sprintf("RouterIP(%04xh) %s", TypeRouterIP, tlv.IP)
}
