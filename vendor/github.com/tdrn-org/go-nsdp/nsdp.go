// nsdp.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

/*
This package provides support for the NSDP (Netgear Switch Discovery Protocol).

This protocol is udp based and uses a request response message flow. E.g. to query
the name of all NSDP capable switches in the current network segment the
following code snippet may be used:

	...
	conn, err := nsdp.NewConn("255.255.255.255:63322", true)
	defer conn.Close()
	requestMsg := nsdp.NewMessage(ReadRequest)
	requestMsg.AppendTLV(EmptyDeviceName())
	responseMsgs, err := conn.SendReceiveMessage(requestMsg)
	for _, responseMsg := range responseMsgs {
		...
	}
	...

This snippet broadcasts a read request containing the DeviceName TLV. NSDP aware devices
will respond with a read-response message containing a filled DeviceName TLV.
*/
package nsdp

// IPv4 broadcast address
const IPv4BroadcastTarget = "255.255.255.255:63322"
