/*
 * Copyright 2024 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mesh

import (
	"slices"
	"strings"
)

// List contains all nodes of a mesh.
type List struct {
	SchemaVersion string `json:"schema_version"`
	Nodes         []Node `json:"nodes"`
}

// Connections determines all unique connections within the mesh left to right.
// Means, left device of a connection is always either a master or a slave node
// of the mesh and right device is either a slave node or a client node.
func (list *List) Connections() []*Connection {
	nodeMap := make(map[string]*Node)
	for nodeIndex, node := range list.Nodes {
		nodeMap[node.Uid] = &list.Nodes[nodeIndex]
	}
	connections := make([]*Connection, 0)
	for _, left := range nodeMap {
		for _, leftInterface := range left.NodeInterfaces {
			for _, leftLink := range leftInterface.NodeLinks {
				if leftLink.Node1Uid != left.Uid || !leftLink.IsConnected() {
					continue
				}
				right := nodeMap[leftLink.Node2Uid]
				connection := &Connection{
					LeftMeshRole:    left.MeshRole,
					LeftDeviceName:  left.DeviceName,
					RightMeshRole:   right.MeshRole,
					RightDeviceName: right.DeviceName,
					InterfaceName:   leftInterface.Name,
					InterfaceType:   leftInterface.Type,
					MaxDataRateRx:   leftLink.MaxDataRateRx,
					MaxDataRateTx:   leftLink.MaxDataRateTx,
					CurDataRateRx:   leftLink.CurDataRateRx,
					CurDataRateTx:   leftLink.CurDataRateTx,
				}
				connections = append(connections, connection)
			}
		}
	}
	slices.SortFunc(connections, compareConnections)
	return connections
}

// Connection represents a single connection between two nodes within the mesh.
type Connection struct {
	// LeftMeshRole contains the role of the left device (master or slave).
	LeftMeshRole string
	// LeftDeviceName contains the name of the left device.
	LeftDeviceName string
	// RightMeshRole contains the role of the right device (slave or unknown)
	RightMeshRole string
	// RightDeviceName contains the name of the right device.
	RightDeviceName string
	// InterfaceName contains the logical name of the interface the devices are connected via.
	InterfaceName string
	// InterfaceType contains the type of the interface the devices are connected via (LAN or WLAN)
	InterfaceType string
	// MaxDataRateRx contains the maximum receiving bit rate between the devices.
	MaxDataRateRx int
	// MaxDataRateRx contains the maximum transmitting bit rate between the devices.
	MaxDataRateTx int
	// CurDataRateRx contains the current receiving bit rate between the devices.
	CurDataRateRx int
	// CurDataRateTx contains the current transmitting bit rate between the devices.
	CurDataRateTx int
}

func compareConnections(a *Connection, b *Connection) int {
	comparison := strings.Compare(a.LeftMeshRole, b.LeftMeshRole)
	if comparison != 0 {
		return comparison
	}
	comparison = strings.Compare(a.RightMeshRole, b.RightMeshRole)
	if comparison != 0 {
		return comparison
	}
	comparison = strings.Compare(a.InterfaceType, b.InterfaceType)
	if comparison != 0 {
		return comparison
	}
	comparison = strings.Compare(a.InterfaceName, b.InterfaceName)
	if comparison != 0 {
		return comparison
	}
	comparison = strings.Compare(a.LeftDeviceName, b.LeftDeviceName)
	if comparison != 0 {
		return comparison
	}
	comparison = strings.Compare(a.RightDeviceName, b.RightDeviceName)
	if comparison != 0 {
		return comparison
	}
	return 0
}

// Node represents a single node within the mesh.
type Node struct {
	Uid            string      `json:"uid"`
	DeviceName     string      `json:"device_name"`
	IsMeshed       bool        `json:"is_meshed"`
	MeshRole       string      `json:"mesh_role"`
	NodeInterfaces []Interface `json:"node_interfaces"`
}

// IsMaster determines whether this node is a master node.
func (node *Node) IsMaster() bool {
	return node.IsMeshed && node.MeshRole == "master"
}

// IsSlave determines whether this node is a slave node.
func (node *Node) IsSlave() bool {
	return node.IsMeshed && node.MeshRole == "slave"
}

// Interface represents a node's interface capable of handling multiple links to other nodes.
type Interface struct {
	Uid       string `json:"uid"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	NodeLinks []Link `json:"node_links"`
}

// Link represents a single connection between two nodes.
type Link struct {
	State             string `json:"state"`
	Node1Uid          string `json:"node_1_uid"`
	Node2Uid          string `json:"node_2_uid"`
	NodeInterface1Uid string `json:"node_interface_1_uid"`
	NodeInterface2Uid string `json:"node_interface_2_uid"`
	MaxDataRateRx     int    `json:"max_data_rate_rx"`
	MaxDataRateTx     int    `json:"max_data_rate_tx"`
	CurDataRateRx     int    `json:"cur_data_rate_rx"`
	CurDataRateTx     int    `json:"cur_data_rate_tx"`
}

// IsConnected whether a link is currently connected.
func (link *Link) IsConnected() bool {
	return link.State == "CONNECTED"
}
