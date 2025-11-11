// conn.go
//
// Copyright (C) 2022-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package nsdp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const defaultReceiveBufferSize uint = 8192
const defaultReceiveQueueLength uint = 16
const defaultReceiveDeviceLimit uint = 0
const defaultReceiveTimeout time.Duration = 2000 * time.Millisecond

// Conn represents a network connection used for sending and receiving NSDP messages.
type Conn struct {
	laddr              *net.UDPAddr
	taddr              *net.UDPAddr
	host               net.HardwareAddr
	conn               *net.UDPConn
	seq                Sequence
	ReceiveBufferSize  uint          // Receive buffer size (defaults to 8192)
	ReceiveQueueLength uint          // Receive queue length (defaults to 16)
	ReceiveDeviceLimit uint          // Receive device limit (defaults to 0; no limit)
	ReceiveTimeout     time.Duration // Receive timeout (defaults to 2s)
	Debug              bool          // Enables debug output via log.Printf
}

// NewConn establishes a new connection to the given remote target.
func NewConn(target string, debug bool) (*Conn, error) {
	if debug {
		log.Printf("NSDP setting up connection...")
		log.Printf("NSDP target address: '%s'", target)
	}
	taddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}
	tconn, err := net.Dial("udp", target)
	if err != nil {
		return nil, err
	}
	tconn.Close()
	lhost, _, err := net.SplitHostPort(tconn.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	lport := strconv.Itoa(int(taddr.AddrPort().Port() - 1))
	listen := net.JoinHostPort(lhost, lport)
	if debug {
		log.Printf("NSDP listen address: '%s'", listen)
	}
	laddr, err := net.ResolveUDPAddr("udp", listen)
	if err != nil {
		return nil, err
	}
	host, err := lookupHardwareAddr(laddr)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Printf("NSDP host MAC: '%s'", host)
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	return &Conn{
		laddr:              laddr,
		taddr:              taddr,
		host:               host,
		conn:               conn,
		seq:                Sequence(time.Now().UnixNano()),
		ReceiveBufferSize:  defaultReceiveBufferSize,
		ReceiveQueueLength: defaultReceiveQueueLength,
		ReceiveDeviceLimit: defaultReceiveDeviceLimit,
		ReceiveTimeout:     defaultReceiveTimeout,
		Debug:              debug,
	}, nil
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// SendReceiveMessage sends the given NSDP message and waits for responses.
//
// The submitted message's host address and sequence number are ignored. Instead the connection state
// is used to populate this info.
//
// If the message's device address is empty (00:00:00:00:00:00), an arbitrary number of response messages is returned.
// Furthermore the call will only finish after the receive timeout or the receive device limit is reached. Receiving no
// response is not considered an error.
//
// If the message's device address has been set, exactly one response message is returned. Furthermore the call will
// return as soon as a response is received. Receiving no response is considered an error.
//
// The returned map is build up using the responding device's hardware address string as the key and the corresponding
// response message as the value.
func (c *Conn) SendReceiveMessage(msg *Message) (map[string]*Message, error) {
	c.seq += 1
	c.conn.SetReadDeadline(time.Now().Add(c.ReceiveTimeout))
	if bytes.Equal(msg.Header.DeviceAddress, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		return c.sendReceiveBroadcastMessage(msg)
	}
	return c.sendReceiveUnicastMessage(msg)
}

type receiveQueueEntry struct {
	msg *Message
	err error
}

func (c *Conn) sendReceiveBroadcastMessage(msg *Message) (map[string]*Message, error) {
	receiveQueue := make(chan *receiveQueueEntry, c.ReceiveQueueLength)
	go func() {
		for {
			msg, err := c.receiveMessage()
			receiveQueue <- &receiveQueueEntry{
				msg: msg,
				err: err,
			}
			if err != nil {
				break
			}
		}
	}()
	err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}
	receivedMsgs := make(map[string]*Message, 0)
	for {
		received := <-receiveQueue
		if received.err != nil {
			if !isTimeoutErr(received.err) {
				return nil, received.err
			}
			break
		}
		receivedMsgs[received.msg.Header.DeviceAddress.String()] = received.msg
		if 0 < c.ReceiveDeviceLimit && c.ReceiveDeviceLimit <= uint(len(receivedMsgs)) {
			break
		}
	}
	return receivedMsgs, nil
}

func (c *Conn) sendReceiveUnicastMessage(msg *Message) (map[string]*Message, error) {
	receiveQueue := make(chan *receiveQueueEntry, 1)
	go func() {
		for {
			msg, err := c.receiveMessage()
			receiveQueue <- &receiveQueueEntry{
				msg: msg,
				err: err,
			}
			break
		}
	}()
	err := c.sendMessage(msg)
	if err != nil {
		return nil, err
	}
	receivedMsgs := make(map[string]*Message, 0)
	received := <-receiveQueue
	if received.err != nil {
		return nil, received.err
	}
	receivedMsgs[received.msg.Header.DeviceAddress.String()] = received.msg
	return receivedMsgs, nil
}

func (c *Conn) sendMessage(msg *Message) error {
	preparedMsg := msg.prepareMessage(c.host, c.seq)
	sendBuffer := preparedMsg.Marshal()
	if c.Debug {
		log.Printf("NSDP %s > %s:\n%s\n%s", c.laddr, c.taddr, hex.EncodeToString(sendBuffer), preparedMsg)
	}
	_, err := c.conn.WriteToUDP(sendBuffer, c.taddr)
	return err
}

func (c *Conn) receiveMessage() (*Message, error) {
	buffer := make([]byte, c.ReceiveBufferSize)
	for {
		len, addr, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			return nil, err
		}
		msg, err := c.unmarshalReceivedMessage(addr, buffer[:len])
		if err != nil {
			return nil, err
		}
		if !c.checkMessageSequence(addr, msg) {
			continue
		}
		if c.Debug {
			log.Printf("NSDP %s < %s:\n%s\n%s", c.laddr, addr, hex.EncodeToString(buffer[:len]), msg)
		}
		return msg, nil
	}
}

func (c *Conn) unmarshalReceivedMessage(addr *net.UDPAddr, received []byte) (*Message, error) {
	msg, err := UnmarshalMessage(received)
	if err != nil {
		if c.Debug {
			log.Printf("NSDP %s < %s:\n%s", c.laddr, addr, hex.EncodeToString(received))
			log.Printf("NSDP Error while unmarshaling message; cause: %v", err)
		}
		return nil, err
	}
	return msg, nil
}

func (c *Conn) checkMessageSequence(addr *net.UDPAddr, msg *Message) bool {
	if msg.Header.Sequence != c.seq {
		if c.Debug {
			log.Printf("NSDP %s < %s:\nIgnoring unsolicited message (sequence: %04xh)", c.laddr, addr, msg.Header.Sequence)
		}
		return false
	}
	return true
}

func isTimeoutErr(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

func lookupHardwareAddr(addr *net.UDPAddr) (net.HardwareAddr, error) {
	// lo has no real MAC; use 00:00:00:00:00:00 in this case
	if addr.IP.IsLoopback() {
		return make([]byte, 6), nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if isMatchingInterface(addr, &iface) {
			if len(iface.HardwareAddr) != 6 {
				return nil, fmt.Errorf("failed to lookup hardware address for interface %s", iface.Name)
			}
			return iface.HardwareAddr, nil
		}
	}
	return nil, fmt.Errorf("failed to lookup hardware address for address: %s", addr)
}

func isMatchingInterface(addr *net.UDPAddr, iface *net.Interface) bool {
	ifaceAddrs, err := iface.Addrs()
	if err == nil {
		for _, ifaceAddr := range ifaceAddrs {
			if strings.HasPrefix(ifaceAddr.String(), addr.IP.String()) {
				return true
			}
		}
	}
	return false
}
