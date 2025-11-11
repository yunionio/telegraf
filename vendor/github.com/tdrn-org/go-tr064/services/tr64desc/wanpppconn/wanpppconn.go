// generated from spec version: 1.0
package wanpppconn

import (
	"encoding/xml"
	"github.com/tdrn-org/go-tr064"
)

type ServiceClient struct {
	TR064Client *tr064.Client
	Service     tr064.ServiceDescriptor
}

type GetInfoRequest struct {
	XMLName      xml.Name `xml:"u:GetInfoRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetInfoResponse struct {
	XMLName                    xml.Name `xml:"GetInfoResponse"`
	NewEnable                  bool     `xml:"NewEnable"`
	NewConnectionStatus        string   `xml:"NewConnectionStatus"`
	NewPossibleConnectionTypes string   `xml:"NewPossibleConnectionTypes"`
	NewConnectionType          string   `xml:"NewConnectionType"`
	NewName                    string   `xml:"NewName"`
	NewUptime                  uint32   `xml:"NewUptime"`
	NewUpstreamMaxBitRate      uint32   `xml:"NewUpstreamMaxBitRate"`
	NewDownstreamMaxBitRate    uint32   `xml:"NewDownstreamMaxBitRate"`
	NewLastConnectionError     string   `xml:"NewLastConnectionError"`
	NewIdleDisconnectTime      uint32   `xml:"NewIdleDisconnectTime"`
	NewRSIPAvailable           bool     `xml:"NewRSIPAvailable"`
	NewUserName                string   `xml:"NewUserName"`
	NewNATEnabled              bool     `xml:"NewNATEnabled"`
	NewExternalIPAddress       string   `xml:"NewExternalIPAddress"`
	NewDNSServers              string   `xml:"NewDNSServers"`
	NewMACAddress              string   `xml:"NewMACAddress"`
	NewConnectionTrigger       string   `xml:"NewConnectionTrigger"`
	NewLastAuthErrorInfo       string   `xml:"NewLastAuthErrorInfo"`
	NewMaxCharsUsername        uint16   `xml:"NewMaxCharsUsername"`
	NewMinCharsUsername        uint16   `xml:"NewMinCharsUsername"`
	NewAllowedCharsUsername    string   `xml:"NewAllowedCharsUsername"`
	NewMaxCharsPassword        uint16   `xml:"NewMaxCharsPassword"`
	NewMinCharsPassword        uint16   `xml:"NewMinCharsPassword"`
	NewAllowedCharsPassword    string   `xml:"NewAllowedCharsPassword"`
	NewTransportType           string   `xml:"NewTransportType"`
	NewRouteProtocolRx         string   `xml:"NewRouteProtocolRx"`
	NewPPPoEServiceName        string   `xml:"NewPPPoEServiceName"`
	NewRemoteIPAddress         string   `xml:"NewRemoteIPAddress"`
	NewPPPoEACName             string   `xml:"NewPPPoEACName"`
	NewDNSEnabled              bool     `xml:"NewDNSEnabled"`
	NewDNSOverrideAllowed      bool     `xml:"NewDNSOverrideAllowed"`
}

func (client *ServiceClient) GetInfo(out *GetInfoResponse) error {
	in := &GetInfoRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetInfo", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetConnectionTypeInfoRequest struct {
	XMLName      xml.Name `xml:"u:GetConnectionTypeInfoRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetConnectionTypeInfoResponse struct {
	XMLName                    xml.Name `xml:"GetConnectionTypeInfoResponse"`
	NewConnectionType          string   `xml:"NewConnectionType"`
	NewPossibleConnectionTypes string   `xml:"NewPossibleConnectionTypes"`
}

func (client *ServiceClient) GetConnectionTypeInfo(out *GetConnectionTypeInfoResponse) error {
	in := &GetConnectionTypeInfoRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetConnectionTypeInfo", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type SetConnectionTypeRequest struct {
	XMLName           xml.Name `xml:"u:SetConnectionTypeRequest"`
	XMLNameSpace      string   `xml:"xmlns:u,attr"`
	NewConnectionType string   `xml:"NewConnectionType"`
}

type SetConnectionTypeResponse struct {
	XMLName xml.Name `xml:"SetConnectionTypeResponse"`
}

func (client *ServiceClient) SetConnectionType(in *SetConnectionTypeRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &SetConnectionTypeResponse{}
	return client.TR064Client.InvokeService(client.Service, "SetConnectionType", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetStatusInfoRequest struct {
	XMLName      xml.Name `xml:"u:GetStatusInfoRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetStatusInfoResponse struct {
	XMLName                xml.Name `xml:"GetStatusInfoResponse"`
	NewConnectionStatus    string   `xml:"NewConnectionStatus"`
	NewLastConnectionError string   `xml:"NewLastConnectionError"`
	NewUptime              uint32   `xml:"NewUptime"`
}

func (client *ServiceClient) GetStatusInfo(out *GetStatusInfoResponse) error {
	in := &GetStatusInfoRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetStatusInfo", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetUserNameRequest struct {
	XMLName      xml.Name `xml:"u:GetUserNameRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetUserNameResponse struct {
	XMLName     xml.Name `xml:"GetUserNameResponse"`
	NewUserName string   `xml:"NewUserName"`
}

func (client *ServiceClient) GetUserName(out *GetUserNameResponse) error {
	in := &GetUserNameRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetUserName", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type SetUserNameRequest struct {
	XMLName      xml.Name `xml:"u:SetUserNameRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
	NewUserName  string   `xml:"NewUserName"`
}

type SetUserNameResponse struct {
	XMLName xml.Name `xml:"SetUserNameResponse"`
}

func (client *ServiceClient) SetUserName(in *SetUserNameRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &SetUserNameResponse{}
	return client.TR064Client.InvokeService(client.Service, "SetUserName", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type SetPasswordRequest struct {
	XMLName      xml.Name `xml:"u:SetPasswordRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
	NewPassword  string   `xml:"NewPassword"`
}

type SetPasswordResponse struct {
	XMLName xml.Name `xml:"SetPasswordResponse"`
}

func (client *ServiceClient) SetPassword(in *SetPasswordRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &SetPasswordResponse{}
	return client.TR064Client.InvokeService(client.Service, "SetPassword", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetNATRSIPStatusRequest struct {
	XMLName      xml.Name `xml:"u:GetNATRSIPStatusRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetNATRSIPStatusResponse struct {
	XMLName          xml.Name `xml:"GetNATRSIPStatusResponse"`
	NewRSIPAvailable bool     `xml:"NewRSIPAvailable"`
	NewNATEnabled    bool     `xml:"NewNATEnabled"`
}

func (client *ServiceClient) GetNATRSIPStatus(out *GetNATRSIPStatusResponse) error {
	in := &GetNATRSIPStatusRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetNATRSIPStatus", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type SetConnectionTriggerRequest struct {
	XMLName              xml.Name `xml:"u:SetConnectionTriggerRequest"`
	XMLNameSpace         string   `xml:"xmlns:u,attr"`
	NewConnectionTrigger string   `xml:"NewConnectionTrigger"`
}

type SetConnectionTriggerResponse struct {
	XMLName xml.Name `xml:"SetConnectionTriggerResponse"`
}

func (client *ServiceClient) SetConnectionTrigger(in *SetConnectionTriggerRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &SetConnectionTriggerResponse{}
	return client.TR064Client.InvokeService(client.Service, "SetConnectionTrigger", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type ForceTerminationRequest struct {
	XMLName      xml.Name `xml:"u:ForceTerminationRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type ForceTerminationResponse struct {
	XMLName xml.Name `xml:"ForceTerminationResponse"`
}

func (client *ServiceClient) ForceTermination() error {
	in := &ForceTerminationRequest{XMLNameSpace: client.Service.Type()}
	out := &ForceTerminationResponse{}
	return client.TR064Client.InvokeService(client.Service, "ForceTermination", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type RequestConnectionRequest struct {
	XMLName      xml.Name `xml:"u:RequestConnectionRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type RequestConnectionResponse struct {
	XMLName xml.Name `xml:"RequestConnectionResponse"`
}

func (client *ServiceClient) RequestConnection() error {
	in := &RequestConnectionRequest{XMLNameSpace: client.Service.Type()}
	out := &RequestConnectionResponse{}
	return client.TR064Client.InvokeService(client.Service, "RequestConnection", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetGenericPortMappingEntryRequest struct {
	XMLName             xml.Name `xml:"u:GetGenericPortMappingEntryRequest"`
	XMLNameSpace        string   `xml:"xmlns:u,attr"`
	NewPortMappingIndex uint16   `xml:"NewPortMappingIndex"`
}

type GetGenericPortMappingEntryResponse struct {
	XMLName                   xml.Name `xml:"GetGenericPortMappingEntryResponse"`
	NewRemoteHost             string   `xml:"NewRemoteHost"`
	NewExternalPort           uint16   `xml:"NewExternalPort"`
	NewProtocol               string   `xml:"NewProtocol"`
	NewInternalPort           uint16   `xml:"NewInternalPort"`
	NewInternalClient         string   `xml:"NewInternalClient"`
	NewEnabled                bool     `xml:"NewEnabled"`
	NewPortMappingDescription string   `xml:"NewPortMappingDescription"`
	NewLeaseDuration          uint32   `xml:"NewLeaseDuration"`
}

func (client *ServiceClient) GetGenericPortMappingEntry(in *GetGenericPortMappingEntryRequest, out *GetGenericPortMappingEntryResponse) error {
	in.XMLNameSpace = client.Service.Type()
	return client.TR064Client.InvokeService(client.Service, "GetGenericPortMappingEntry", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetSpecificPortMappingEntryRequest struct {
	XMLName         xml.Name `xml:"u:GetSpecificPortMappingEntryRequest"`
	XMLNameSpace    string   `xml:"xmlns:u,attr"`
	NewRemoteHost   string   `xml:"NewRemoteHost"`
	NewExternalPort uint16   `xml:"NewExternalPort"`
	NewProtocol     string   `xml:"NewProtocol"`
}

type GetSpecificPortMappingEntryResponse struct {
	XMLName                   xml.Name `xml:"GetSpecificPortMappingEntryResponse"`
	NewInternalPort           uint16   `xml:"NewInternalPort"`
	NewInternalClient         string   `xml:"NewInternalClient"`
	NewEnabled                bool     `xml:"NewEnabled"`
	NewPortMappingDescription string   `xml:"NewPortMappingDescription"`
	NewLeaseDuration          uint32   `xml:"NewLeaseDuration"`
}

func (client *ServiceClient) GetSpecificPortMappingEntry(in *GetSpecificPortMappingEntryRequest, out *GetSpecificPortMappingEntryResponse) error {
	in.XMLNameSpace = client.Service.Type()
	return client.TR064Client.InvokeService(client.Service, "GetSpecificPortMappingEntry", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type AddPortMappingRequest struct {
	XMLName                   xml.Name `xml:"u:AddPortMappingRequest"`
	XMLNameSpace              string   `xml:"xmlns:u,attr"`
	NewRemoteHost             string   `xml:"NewRemoteHost"`
	NewExternalPort           uint16   `xml:"NewExternalPort"`
	NewProtocol               string   `xml:"NewProtocol"`
	NewInternalPort           uint16   `xml:"NewInternalPort"`
	NewInternalClient         string   `xml:"NewInternalClient"`
	NewEnabled                bool     `xml:"NewEnabled"`
	NewPortMappingDescription string   `xml:"NewPortMappingDescription"`
	NewLeaseDuration          uint32   `xml:"NewLeaseDuration"`
}

type AddPortMappingResponse struct {
	XMLName xml.Name `xml:"AddPortMappingResponse"`
}

func (client *ServiceClient) AddPortMapping(in *AddPortMappingRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &AddPortMappingResponse{}
	return client.TR064Client.InvokeService(client.Service, "AddPortMapping", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type DeletePortMappingRequest struct {
	XMLName         xml.Name `xml:"u:DeletePortMappingRequest"`
	XMLNameSpace    string   `xml:"xmlns:u,attr"`
	NewRemoteHost   string   `xml:"NewRemoteHost"`
	NewExternalPort uint16   `xml:"NewExternalPort"`
	NewProtocol     string   `xml:"NewProtocol"`
}

type DeletePortMappingResponse struct {
	XMLName xml.Name `xml:"DeletePortMappingResponse"`
}

func (client *ServiceClient) DeletePortMapping(in *DeletePortMappingRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &DeletePortMappingResponse{}
	return client.TR064Client.InvokeService(client.Service, "DeletePortMapping", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetExternalIPAddressRequest struct {
	XMLName      xml.Name `xml:"u:GetExternalIPAddressRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetExternalIPAddressResponse struct {
	XMLName              xml.Name `xml:"GetExternalIPAddressResponse"`
	NewExternalIPAddress string   `xml:"NewExternalIPAddress"`
}

func (client *ServiceClient) GetExternalIPAddress(out *GetExternalIPAddressResponse) error {
	in := &GetExternalIPAddressRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetExternalIPAddress", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_GetDNSServersRequest struct {
	XMLName      xml.Name `xml:"u:X_GetDNSServersRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_GetDNSServersResponse struct {
	XMLName       xml.Name `xml:"X_GetDNSServersResponse"`
	NewDNSServers string   `xml:"NewDNSServers"`
}

func (client *ServiceClient) X_GetDNSServers(out *X_GetDNSServersResponse) error {
	in := &X_GetDNSServersRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_GetDNSServers", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetLinkLayerMaxBitRatesRequest struct {
	XMLName      xml.Name `xml:"u:GetLinkLayerMaxBitRatesRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetLinkLayerMaxBitRatesResponse struct {
	XMLName                 xml.Name `xml:"GetLinkLayerMaxBitRatesResponse"`
	NewUpstreamMaxBitRate   uint32   `xml:"NewUpstreamMaxBitRate"`
	NewDownstreamMaxBitRate uint32   `xml:"NewDownstreamMaxBitRate"`
}

func (client *ServiceClient) GetLinkLayerMaxBitRates(out *GetLinkLayerMaxBitRatesResponse) error {
	in := &GetLinkLayerMaxBitRatesRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetLinkLayerMaxBitRates", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetPortMappingNumberOfEntriesRequest struct {
	XMLName      xml.Name `xml:"u:GetPortMappingNumberOfEntriesRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetPortMappingNumberOfEntriesResponse struct {
	XMLName                       xml.Name `xml:"GetPortMappingNumberOfEntriesResponse"`
	NewPortMappingNumberOfEntries uint16   `xml:"NewPortMappingNumberOfEntries"`
}

func (client *ServiceClient) GetPortMappingNumberOfEntries(out *GetPortMappingNumberOfEntriesResponse) error {
	in := &GetPortMappingNumberOfEntriesRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetPortMappingNumberOfEntries", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type SetRouteProtocolRxRequest struct {
	XMLName            xml.Name `xml:"u:SetRouteProtocolRxRequest"`
	XMLNameSpace       string   `xml:"xmlns:u,attr"`
	NewRouteProtocolRx string   `xml:"NewRouteProtocolRx"`
}

type SetRouteProtocolRxResponse struct {
	XMLName xml.Name `xml:"SetRouteProtocolRxResponse"`
}

func (client *ServiceClient) SetRouteProtocolRx(in *SetRouteProtocolRxRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &SetRouteProtocolRxResponse{}
	return client.TR064Client.InvokeService(client.Service, "SetRouteProtocolRx", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type SetIdleDisconnectTimeRequest struct {
	XMLName               xml.Name `xml:"u:SetIdleDisconnectTimeRequest"`
	XMLNameSpace          string   `xml:"xmlns:u,attr"`
	NewIdleDisconnectTime uint32   `xml:"NewIdleDisconnectTime"`
}

type SetIdleDisconnectTimeResponse struct {
	XMLName xml.Name `xml:"SetIdleDisconnectTimeResponse"`
}

func (client *ServiceClient) SetIdleDisconnectTime(in *SetIdleDisconnectTimeRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &SetIdleDisconnectTimeResponse{}
	return client.TR064Client.InvokeService(client.Service, "SetIdleDisconnectTime", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetAutoDisconnectTimeSpanRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetAutoDisconnectTimeSpanRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetAutoDisconnectTimeSpanResponse struct {
	XMLName                                xml.Name `xml:"X_AVM-DE_GetAutoDisconnectTimeSpanResponse"`
	NewX_AVM_DE_DisconnectPreventionEnable bool     `xml:"NewX_AVM-DE_DisconnectPreventionEnable"`
	NewX_AVM_DE_DisconnectPreventionHour   uint16   `xml:"NewX_AVM-DE_DisconnectPreventionHour"`
}

func (client *ServiceClient) X_AVM_DE_GetAutoDisconnectTimeSpan(out *X_AVM_DE_GetAutoDisconnectTimeSpanResponse) error {
	in := &X_AVM_DE_GetAutoDisconnectTimeSpanRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetAutoDisconnectTimeSpan", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetAutoDisconnectTimeSpanRequest struct {
	XMLName                                xml.Name `xml:"u:X_AVM-DE_SetAutoDisconnectTimeSpanRequest"`
	XMLNameSpace                           string   `xml:"xmlns:u,attr"`
	NewX_AVM_DE_DisconnectPreventionEnable bool     `xml:"NewX_AVM-DE_DisconnectPreventionEnable"`
	NewX_AVM_DE_DisconnectPreventionHour   uint16   `xml:"NewX_AVM-DE_DisconnectPreventionHour"`
}

type X_AVM_DE_SetAutoDisconnectTimeSpanResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetAutoDisconnectTimeSpanResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetAutoDisconnectTimeSpan(in *X_AVM_DE_SetAutoDisconnectTimeSpanRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetAutoDisconnectTimeSpanResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetAutoDisconnectTimeSpan", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}
