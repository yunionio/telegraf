// test_responder.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"encoding/hex"
	"log"
	"net"
)

// TestResponder supports replay of static NSDP responses for testing.
//
// Multiple sets of responses can be added to a responder instance by
// invoking AddResponses. After the instance has been started, they
// are simply played back as soon as a request is received (1st request
// is handled by sending back the responses added by 1st AddResponses call,
// 2nd request by ... and so on).
type TestResponder struct {
	taddr          *net.UDPAddr
	responseChunks [][][]byte
	conn           *net.UDPConn
	started        chan bool
	stopped        chan bool
}

// NewTestResponder creates a new responder instance for the given target address.
//
// The target address should be the same, as submitted to NewConn.
func NewTestResponder(target string) (*TestResponder, error) {
	taddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}
	responder := &TestResponder{
		taddr:          taddr,
		responseChunks: make([][][]byte, 0),
		started:        make(chan bool, 1),
		stopped:        make(chan bool, 1),
	}
	return responder, nil
}

// AddResponses adds an arbitrary number string encoded NSDP messages to be send back on an incoming request.
func (responder *TestResponder) AddResponses(encodedResponses ...string) {
	responseChunk := make([][]byte, 0)
	for _, encodedResponse := range encodedResponses {
		response, err := hex.DecodeString(encodedResponse)
		if err != nil {
			log.Panicf("NSDP-TestResponder invalid response; cause: %v", err)
		}
		responseChunk = append(responseChunk, response)
	}
	responder.responseChunks = append(responder.responseChunks, responseChunk)
}

// Start starts this responder instance.
func (responder *TestResponder) Start() error {
	if responder.taddr.IP.IsLoopback() {
		log.Printf("NSDP-TestResponder starting on target %s", responder.taddr)
		conn, err := net.ListenUDP("udp", responder.taddr)
		if err != nil {
			return err
		}
		responder.conn = conn
		go responder.listen()
		<-responder.started
	}
	return nil
}

func (responder *TestResponder) listen() {
	defer responder.conn.Close()
	defer func() { responder.stopped <- true }()
	buffer := make([]byte, 8192)
	for _, responseChunk := range responder.responseChunks {
		log.Printf("NSDP-TestResponder listening on %s", responder.conn.LocalAddr().String())
		responder.started <- true
		len, addr, err := responder.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("NSDP-TestResponder listening failure; cause: %v", err)
			break
		}
		if len == 1 {
			break
		}
		log.Printf("NSDP-TestResponder %s < %s\n%s", responder.taddr, addr, hex.EncodeToString(buffer[:len]))
		err = responder.handleRequest(addr, buffer[:len], responseChunk)
		if err != nil {
			log.Printf("NSDP-TestResponder failed to handle message; cause: %v", err)
			break
		}
	}
}

func (responder *TestResponder) handleRequest(addr *net.UDPAddr, request []byte, responseChunk [][]byte) error {
	requestMsg, err := UnmarshalMessage(request)
	if err != nil {
		return err
	}
	for _, response := range responseChunk {
		responseMsg, err := UnmarshalMessage(response)
		if err != nil {
			return err
		}
		responseMsg.Header.HostAddress = requestMsg.Header.HostAddress
		responseMsg.Header.Sequence = requestMsg.Header.Sequence
		_, err = responder.conn.WriteToUDP(responseMsg.Marshal(), addr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop stops this responder instance.
func (responder *TestResponder) Stop() error {
	if responder.conn != nil {
		_, err := responder.conn.WriteTo([]byte{0x00}, responder.conn.LocalAddr())
		if err != nil {
			return err
		}
		<-responder.stopped
		log.Println("NSDP-TestResponder stopped")
	}
	return nil
}

// Target gets the actual address this responder instance is listening on.
func (responder *TestResponder) Target() string {
	if responder.conn != nil {
		return responder.conn.LocalAddr().String()
	}
	return responder.taddr.String()
}
