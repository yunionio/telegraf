// Package zabbix implements the sender protocol to send values to zabbix
// Taken from github.com/blacked/go-zabbix (discontinued)
package zabbix

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	defaultConnectTimeout = 5 * time.Second
	defaultWriteTimeout   = 5 * time.Second
	// A heavy loaded Zabbix server processing several metrics,
	// containing LLDs, could take several seconds to respond.
	defaultReadTimeout = 15 * time.Second
)

var (
	// ErrActive is returned when fails to send active agent metrics.
	ErrActive = errors.New("error active agent")
	// ErrTrapper is returned when fails to send trapper metrics.
	ErrTrapper = errors.New("error trapper")
	// ErrInvalidResponse is returned when the response from zabbix is not valid.
	ErrInvalidResponse = errors.New("zabbix response is not valid")
)

// Metric represents the JSON object send for each value to Zabbix.
type Metric struct {
	Host   string `json:"host"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Clock  int64  `json:"clock,omitempty"`
	Active bool   `json:"-"`
}

// NewMetric return a zabbix Metric with the values specified.
// agentActive should be set to true if we are sending to a Zabbix Agent (active) item.
func NewMetric(host, key, value string, agentActive bool, clock ...int64) *Metric {
	m := &Metric{Host: host, Key: key, Value: value, Active: agentActive}
	// do not send clock if not defined
	if len(clock) > 0 {
		m.Clock = int64(clock[0])
	}
	return m
}

// Packet represents the JSON object send to Zabbix with the values.
type Packet struct {
	Request      string    `json:"request"`
	Data         []*Metric `json:"data,omitempty"`
	Clock        int64     `json:"clock,omitempty"`
	Host         string    `json:"host,omitempty"`
	HostMetadata string    `json:"host_metadata,omitempty"`
}

// Response is the data returned from zabbix-server.
type Response struct {
	Response string
	Info     string
}

// ResponseInfo contains the specific number of items processed, failed, total and the time spent.
type ResponseInfo struct {
	Processed int
	Failed    int
	Total     int
	Spent     time.Duration
}

// GetInfo tries to parse the info string from the response data.
func (r *Response) GetInfo() (*ResponseInfo, error) {
	ret := ResponseInfo{}
	if r.Response != "success" {
		return &ret, fmt.Errorf("can not process info if response is not 'Success' (%s)", r.Response)
	}

	sp := strings.Split(r.Info, ";")
	if len(sp) != 4 {
		return &ret, fmt.Errorf("invalid format, expected 4 sections divided by ';', got %d for data (%s)", len(sp), r.Info)
	}
	for _, s := range sp {
		sp2 := strings.Split(s, ":")
		if len(sp2) != 2 {
			return &ret, fmt.Errorf("invalid format, expected key:value, data (%s)", s)
		}
		key := strings.TrimSpace(sp2[0])
		value := strings.TrimSpace(sp2[1])
		var err error
		switch key {
		case "processed":
			if ret.Processed, err = strconv.Atoi(value); err != nil {
				return &ret, fmt.Errorf("invalid parsing 'processed' value [%s] error: %s", value, err)
			}
		case "failed":
			if ret.Failed, err = strconv.Atoi(value); err != nil {
				return &ret, fmt.Errorf("invalid parsing 'failed' value [%s] error: %s", value, err)
			}
		case "total":
			if ret.Total, err = strconv.Atoi(value); err != nil {
				return &ret, fmt.Errorf("invalid parsing 'total' value [%s] error: %s", value, err)
			}
		case "seconds spent":
			var f float64
			if f, err = strconv.ParseFloat(value, 64); err != nil {
				return &ret, fmt.Errorf("invalid parsing 'seconds spent' value [%s] error: %s", value, err)
			}
			ret.Spent = time.Duration(int64(f * 1000000000.0))
		}
	}

	return &ret, nil
}

// NewPacket return a zabbix packet with a list of metrics.
func NewPacket(data []*Metric, agentActive bool, clock ...int64) *Packet {
	var request string
	if agentActive {
		request = "agent data"
	} else {
		request = "sender data"
	}

	p := &Packet{Request: request, Data: data}

	// do not send clock if not defined
	if len(clock) > 0 {
		p.Clock = int64(clock[0])
	}
	return p
}

// DataLen Packet method, return 8 bytes with packet length in little endian order.
func (p *Packet) DataLen() []byte {
	dataLen := make([]byte, 8)
	JSONData, _ := json.Marshal(p)
	binary.LittleEndian.PutUint32(dataLen, uint32(len(JSONData)))
	return dataLen
}

// Sender object with data to connect to Zabbix.
type Sender struct {
	ServerAddr     string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// NewSender return a sender object to send metrics using default values for timeouts.
// Server could be a hostname or ip address with, optional, port. Same format as https://pkg.go.dev/net#Dial.
func NewSender(addr string) *Sender {
	return &Sender{
		ServerAddr:     addr,
		ConnectTimeout: defaultConnectTimeout,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
	}
}

// NewSenderTimeout return a sender object to send metrics defining values for timeouts.
// Server could be a hostname or ip address with, optional, port. Same format as https://pkg.go.dev/net#Dial.
func NewSenderTimeout(
	server string,
	connectTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
) *Sender {
	return &Sender{
		ServerAddr:     server,
		ConnectTimeout: connectTimeout,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
	}
}

// getHeader return zabbix header.
// https://www.zabbix.com/documentation/4.0/manual/appendix/protocols/header_datalen.
func (s *Sender) getHeader() []byte {
	return []byte("ZBXD\x01")
}

// read data from connection.
func (s *Sender) read(conn net.Conn) ([]byte, error) {
	res, err := ioutil.ReadAll(conn)
	if err != nil {
		return res, fmt.Errorf("receiving data: %s", err.Error())
	}

	return res, nil
}

// SendMetrics send an array of metrics, making different packets for trapper and active items.
// Response for active and trappers items are returned separately.
// The errors are returned together, using errors.Join.
func (s *Sender) SendMetrics(metrics []*Metric) (resActive Response, resTrapper Response, err error) {
	var trapperMetrics []*Metric
	var activeMetrics []*Metric
	var errTrapper error
	var errActive error

	for i := 0; i < len(metrics); i++ {
		if metrics[i].Active {
			activeMetrics = append(activeMetrics, metrics[i])
		} else {
			trapperMetrics = append(trapperMetrics, metrics[i])
		}
	}

	if len(trapperMetrics) > 0 {
		packetTrapper := NewPacket(trapperMetrics, false)
		resTrapper, errTrapper = s.Send(packetTrapper)
		if errTrapper != nil {
			errTrapper = fmt.Errorf("%w: %w", ErrTrapper, errTrapper)
		}
	}

	if len(activeMetrics) > 0 {
		packetActive := NewPacket(activeMetrics, true)
		resActive, errActive = s.Send(packetActive)
		if errActive != nil {
			errActive = fmt.Errorf("%w: %w", ErrActive, errActive)
		}
	}

	return resActive, resTrapper, errors.Join(errActive, errTrapper)
}

// Send connects to Zabbix, sends the data, returns the response and closes the connection.
func (s *Sender) Send(packet *Packet) (res Response, err error) {
	// Timeout to resolve and connect to the server
	conn, err := net.DialTimeout("tcp", s.ServerAddr, s.ConnectTimeout)
	if err != nil {
		return res, fmt.Errorf("connecting to server (timeout=%v): %w", s.ConnectTimeout, err)
	}
	defer conn.Close()

	dataPacket, err := json.Marshal(packet)
	if err != nil {
		return res, fmt.Errorf("marshaling packet: %w", err)
	}

	// Fill buffer
	buffer := append(s.getHeader(), packet.DataLen()...)
	buffer = append(buffer, dataPacket...)

	// Write timeout
	conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))

	// Send packet to zabbix
	_, err = conn.Write(buffer)
	if err != nil {
		return res, fmt.Errorf("sending the data (timeout=%v): %w", s.WriteTimeout, err)
	}

	// Read timeout
	conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))

	// Read response from server
	response, err := s.read(conn)
	if err != nil {
		return res, fmt.Errorf("reading the response (timeout=%v): %w", s.ReadTimeout, err)
	}

	if len(response) < 14 {
		return res, fmt.Errorf("invalid response length: %s", response)
	}
	header := response[:5]
	data := response[13:]

	if !bytes.Equal(header, s.getHeader()) {
		return res, fmt.Errorf("got no valid header [%+v] , expected [%+v]", header, s.getHeader())
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return res, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}

	return res, nil
}

// RegisterHost provides a register a Zabbix's host with Autoregister method.
func (s *Sender) RegisterHost(host, hostmetadata string) error {
	p := &Packet{Request: "active checks", Host: host, HostMetadata: hostmetadata}

	res, err := s.Send(p)
	if err != nil {
		return fmt.Errorf("sending packet: %v", err)
	}

	if res.Response == "success" {
		return nil
	}

	// The autoregister process always return fail the first time
	// We retry the process to get success response to verify the host registration properly
	p = &Packet{Request: "active checks", Host: host, HostMetadata: hostmetadata}

	res, err = s.Send(p)
	if err != nil {
		return fmt.Errorf("sending packet: %v", err)
	}

	if res.Response == "failed" {
		return fmt.Errorf("autoregistration failed, verify hostmetadata")
	}

	return nil
}
