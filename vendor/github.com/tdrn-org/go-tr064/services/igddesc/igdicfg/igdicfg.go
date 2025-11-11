// generated from spec version: 1.0
package igdicfg

import (
	"encoding/xml"
	"github.com/tdrn-org/go-tr064"
)

type ServiceClient struct {
	TR064Client *tr064.Client
	Service     tr064.ServiceDescriptor
}

type GetCommonLinkPropertiesRequest struct {
	XMLName      xml.Name `xml:"u:GetCommonLinkPropertiesRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetCommonLinkPropertiesResponse struct {
	XMLName                       xml.Name `xml:"GetCommonLinkPropertiesResponse"`
	NewWANAccessType              string   `xml:"NewWANAccessType"`
	NewLayer1UpstreamMaxBitRate   uint32   `xml:"NewLayer1UpstreamMaxBitRate"`
	NewLayer1DownstreamMaxBitRate uint32   `xml:"NewLayer1DownstreamMaxBitRate"`
	NewPhysicalLinkStatus         string   `xml:"NewPhysicalLinkStatus"`
}

func (client *ServiceClient) GetCommonLinkProperties(out *GetCommonLinkPropertiesResponse) error {
	in := &GetCommonLinkPropertiesRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetCommonLinkProperties", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetTotalBytesSentRequest struct {
	XMLName      xml.Name `xml:"u:GetTotalBytesSentRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetTotalBytesSentResponse struct {
	XMLName           xml.Name `xml:"GetTotalBytesSentResponse"`
	NewTotalBytesSent uint32   `xml:"NewTotalBytesSent"`
}

func (client *ServiceClient) GetTotalBytesSent(out *GetTotalBytesSentResponse) error {
	in := &GetTotalBytesSentRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetTotalBytesSent", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetTotalBytesReceivedRequest struct {
	XMLName      xml.Name `xml:"u:GetTotalBytesReceivedRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetTotalBytesReceivedResponse struct {
	XMLName               xml.Name `xml:"GetTotalBytesReceivedResponse"`
	NewTotalBytesReceived uint32   `xml:"NewTotalBytesReceived"`
}

func (client *ServiceClient) GetTotalBytesReceived(out *GetTotalBytesReceivedResponse) error {
	in := &GetTotalBytesReceivedRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetTotalBytesReceived", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetTotalPacketsSentRequest struct {
	XMLName      xml.Name `xml:"u:GetTotalPacketsSentRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetTotalPacketsSentResponse struct {
	XMLName             xml.Name `xml:"GetTotalPacketsSentResponse"`
	NewTotalPacketsSent uint32   `xml:"NewTotalPacketsSent"`
}

func (client *ServiceClient) GetTotalPacketsSent(out *GetTotalPacketsSentResponse) error {
	in := &GetTotalPacketsSentRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetTotalPacketsSent", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetTotalPacketsReceivedRequest struct {
	XMLName      xml.Name `xml:"u:GetTotalPacketsReceivedRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetTotalPacketsReceivedResponse struct {
	XMLName                 xml.Name `xml:"GetTotalPacketsReceivedResponse"`
	NewTotalPacketsReceived uint32   `xml:"NewTotalPacketsReceived"`
}

func (client *ServiceClient) GetTotalPacketsReceived(out *GetTotalPacketsReceivedResponse) error {
	in := &GetTotalPacketsReceivedRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetTotalPacketsReceived", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetAddonInfosRequest struct {
	XMLName      xml.Name `xml:"u:GetAddonInfosRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetAddonInfosResponse struct {
	XMLName                          xml.Name `xml:"GetAddonInfosResponse"`
	NewByteSendRate                  uint32   `xml:"NewByteSendRate"`
	NewByteReceiveRate               uint32   `xml:"NewByteReceiveRate"`
	NewPacketSendRate                uint32   `xml:"NewPacketSendRate"`
	NewPacketReceiveRate             uint32   `xml:"NewPacketReceiveRate"`
	NewTotalBytesSent                uint32   `xml:"NewTotalBytesSent"`
	NewTotalBytesReceived            uint32   `xml:"NewTotalBytesReceived"`
	NewAutoDisconnectTime            uint32   `xml:"NewAutoDisconnectTime"`
	NewIdleDisconnectTime            uint32   `xml:"NewIdleDisconnectTime"`
	NewDNSServer1                    string   `xml:"NewDNSServer1"`
	NewDNSServer2                    string   `xml:"NewDNSServer2"`
	NewVoipDNSServer1                string   `xml:"NewVoipDNSServer1"`
	NewVoipDNSServer2                string   `xml:"NewVoipDNSServer2"`
	NewUpnpControlEnabled            bool     `xml:"NewUpnpControlEnabled"`
	NewRoutedBridgedModeBoth         uint8    `xml:"NewRoutedBridgedModeBoth"`
	NewX_AVM_DE_TotalBytesSent64     string   `xml:"NewX_AVM_DE_TotalBytesSent64"`
	NewX_AVM_DE_TotalBytesReceived64 string   `xml:"NewX_AVM_DE_TotalBytesReceived64"`
	NewX_AVM_DE_WANAccessType        string   `xml:"NewX_AVM_DE_WANAccessType"`
}

func (client *ServiceClient) GetAddonInfos(out *GetAddonInfosResponse) error {
	in := &GetAddonInfosRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetAddonInfos", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetDsliteStatusRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM_DE_GetDsliteStatusRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetDsliteStatusResponse struct {
	XMLName                  xml.Name `xml:"X_AVM_DE_GetDsliteStatusResponse"`
	NewX_AVM_DE_DsliteStatus bool     `xml:"NewX_AVM_DE_DsliteStatus"`
}

func (client *ServiceClient) X_AVM_DE_GetDsliteStatus(out *X_AVM_DE_GetDsliteStatusResponse) error {
	in := &X_AVM_DE_GetDsliteStatusRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM_DE_GetDsliteStatus", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetIPTVInfosRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM_DE_GetIPTVInfosRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetIPTVInfosResponse struct {
	XMLName                   xml.Name `xml:"X_AVM_DE_GetIPTVInfosResponse"`
	NewX_AVM_DE_IPTV_Enabled  bool     `xml:"NewX_AVM_DE_IPTV_Enabled"`
	NewX_AVM_DE_IPTV_Provider string   `xml:"NewX_AVM_DE_IPTV_Provider"`
	NewX_AVM_DE_IPTV_URL      string   `xml:"NewX_AVM_DE_IPTV_URL"`
}

func (client *ServiceClient) X_AVM_DE_GetIPTVInfos(out *X_AVM_DE_GetIPTVInfosResponse) error {
	in := &X_AVM_DE_GetIPTVInfosRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM_DE_GetIPTVInfos", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}
