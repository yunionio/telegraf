// generated from spec version: 1.0
package hosts

import (
	"encoding/xml"
	"github.com/tdrn-org/go-tr064"
)

type ServiceClient struct {
	TR064Client *tr064.Client
	Service     tr064.ServiceDescriptor
}

type GetHostNumberOfEntriesRequest struct {
	XMLName      xml.Name `xml:"u:GetHostNumberOfEntriesRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type GetHostNumberOfEntriesResponse struct {
	XMLName                xml.Name `xml:"GetHostNumberOfEntriesResponse"`
	NewHostNumberOfEntries uint16   `xml:"NewHostNumberOfEntries"`
}

func (client *ServiceClient) GetHostNumberOfEntries(out *GetHostNumberOfEntriesResponse) error {
	in := &GetHostNumberOfEntriesRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "GetHostNumberOfEntries", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetSpecificHostEntryRequest struct {
	XMLName       xml.Name `xml:"u:GetSpecificHostEntryRequest"`
	XMLNameSpace  string   `xml:"xmlns:u,attr"`
	NewMACAddress string   `xml:"NewMACAddress"`
}

type GetSpecificHostEntryResponse struct {
	XMLName               xml.Name `xml:"GetSpecificHostEntryResponse"`
	NewIPAddress          string   `xml:"NewIPAddress"`
	NewAddressSource      string   `xml:"NewAddressSource"`
	NewLeaseTimeRemaining int32    `xml:"NewLeaseTimeRemaining"`
	NewInterfaceType      string   `xml:"NewInterfaceType"`
	NewActive             bool     `xml:"NewActive"`
	NewHostName           string   `xml:"NewHostName"`
}

func (client *ServiceClient) GetSpecificHostEntry(in *GetSpecificHostEntryRequest, out *GetSpecificHostEntryResponse) error {
	in.XMLNameSpace = client.Service.Type()
	return client.TR064Client.InvokeService(client.Service, "GetSpecificHostEntry", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type GetGenericHostEntryRequest struct {
	XMLName      xml.Name `xml:"u:GetGenericHostEntryRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
	NewIndex     uint16   `xml:"NewIndex"`
}

type GetGenericHostEntryResponse struct {
	XMLName               xml.Name `xml:"GetGenericHostEntryResponse"`
	NewIPAddress          string   `xml:"NewIPAddress"`
	NewAddressSource      string   `xml:"NewAddressSource"`
	NewLeaseTimeRemaining int32    `xml:"NewLeaseTimeRemaining"`
	NewMACAddress         string   `xml:"NewMACAddress"`
	NewInterfaceType      string   `xml:"NewInterfaceType"`
	NewActive             bool     `xml:"NewActive"`
	NewHostName           string   `xml:"NewHostName"`
}

func (client *ServiceClient) GetGenericHostEntry(in *GetGenericHostEntryRequest, out *GetGenericHostEntryResponse) error {
	in.XMLNameSpace = client.Service.Type()
	return client.TR064Client.InvokeService(client.Service, "GetGenericHostEntry", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetInfoRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetInfoRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetInfoResponse struct {
	XMLName                          xml.Name `xml:"X_AVM-DE_GetInfoResponse"`
	NewX_AVM_DE_FriendlynameMinChars uint16   `xml:"NewX_AVM-DE_FriendlynameMinChars"`
	NewX_AVM_DE_FriendlynameMaxChars uint16   `xml:"NewX_AVM-DE_FriendlynameMaxChars"`
	NewX_AVM_DE_HostnameMinChars     uint16   `xml:"NewX_AVM-DE_HostnameMinChars"`
	NewX_AVM_DE_HostnameMaxChars     uint16   `xml:"NewX_AVM-DE_HostnameMaxChars"`
	NewX_AVM_DE_HostnameAllowedChars string   `xml:"NewX_AVM-DE_HostnameAllowedChars"`
}

func (client *ServiceClient) X_AVM_DE_GetInfo(out *X_AVM_DE_GetInfoResponse) error {
	in := &X_AVM_DE_GetInfoRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetInfo", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetChangeCounterRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetChangeCounterRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetChangeCounterResponse struct {
	XMLName                   xml.Name `xml:"X_AVM-DE_GetChangeCounterResponse"`
	NewX_AVM_DE_ChangeCounter uint32   `xml:"NewX_AVM-DE_ChangeCounter"`
}

func (client *ServiceClient) X_AVM_DE_GetChangeCounter(out *X_AVM_DE_GetChangeCounterResponse) error {
	in := &X_AVM_DE_GetChangeCounterRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetChangeCounter", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetHostNameByMACAddressRequest struct {
	XMLName       xml.Name `xml:"u:X_AVM-DE_SetHostNameByMACAddressRequest"`
	XMLNameSpace  string   `xml:"xmlns:u,attr"`
	NewMACAddress string   `xml:"NewMACAddress"`
	NewHostName   string   `xml:"NewHostName"`
}

type X_AVM_DE_SetHostNameByMACAddressResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetHostNameByMACAddressResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetHostNameByMACAddress(in *X_AVM_DE_SetHostNameByMACAddressRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetHostNameByMACAddressResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetHostNameByMACAddress", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetAutoWakeOnLANByMACAddressRequest struct {
	XMLName       xml.Name `xml:"u:X_AVM-DE_GetAutoWakeOnLANByMACAddressRequest"`
	XMLNameSpace  string   `xml:"xmlns:u,attr"`
	NewMACAddress string   `xml:"NewMACAddress"`
}

type X_AVM_DE_GetAutoWakeOnLANByMACAddressResponse struct {
	XMLName           xml.Name `xml:"X_AVM-DE_GetAutoWakeOnLANByMACAddressResponse"`
	NewAutoWOLEnabled bool     `xml:"NewAutoWOLEnabled"`
}

func (client *ServiceClient) X_AVM_DE_GetAutoWakeOnLANByMACAddress(in *X_AVM_DE_GetAutoWakeOnLANByMACAddressRequest, out *X_AVM_DE_GetAutoWakeOnLANByMACAddressResponse) error {
	in.XMLNameSpace = client.Service.Type()
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetAutoWakeOnLANByMACAddress", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetAutoWakeOnLANByMACAddressRequest struct {
	XMLName           xml.Name `xml:"u:X_AVM-DE_SetAutoWakeOnLANByMACAddressRequest"`
	XMLNameSpace      string   `xml:"xmlns:u,attr"`
	NewMACAddress     string   `xml:"NewMACAddress"`
	NewAutoWOLEnabled bool     `xml:"NewAutoWOLEnabled"`
}

type X_AVM_DE_SetAutoWakeOnLANByMACAddressResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetAutoWakeOnLANByMACAddressResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetAutoWakeOnLANByMACAddress(in *X_AVM_DE_SetAutoWakeOnLANByMACAddressRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetAutoWakeOnLANByMACAddressResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetAutoWakeOnLANByMACAddress", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_WakeOnLANByMACAddressRequest struct {
	XMLName       xml.Name `xml:"u:X_AVM-DE_WakeOnLANByMACAddressRequest"`
	XMLNameSpace  string   `xml:"xmlns:u,attr"`
	NewMACAddress string   `xml:"NewMACAddress"`
}

type X_AVM_DE_WakeOnLANByMACAddressResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_WakeOnLANByMACAddressResponse"`
}

func (client *ServiceClient) X_AVM_DE_WakeOnLANByMACAddress(in *X_AVM_DE_WakeOnLANByMACAddressRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_WakeOnLANByMACAddressResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_WakeOnLANByMACAddress", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetSpecificHostEntryByIPRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetSpecificHostEntryByIPRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
	NewIPAddress string   `xml:"NewIPAddress"`
}

type X_AVM_DE_GetSpecificHostEntryByIPResponse struct {
	XMLName                             xml.Name `xml:"X_AVM-DE_GetSpecificHostEntryByIPResponse"`
	NewMACAddress                       string   `xml:"NewMACAddress"`
	NewActive                           bool     `xml:"NewActive"`
	NewHostName                         string   `xml:"NewHostName"`
	NewInterfaceType                    string   `xml:"NewInterfaceType"`
	NewX_AVM_DE_Port                    uint32   `xml:"NewX_AVM-DE_Port"`
	NewX_AVM_DE_Speed                   uint32   `xml:"NewX_AVM-DE_Speed"`
	NewX_AVM_DE_UpdateAvailable         bool     `xml:"NewX_AVM-DE_UpdateAvailable"`
	NewX_AVM_DE_UpdateSuccessful        string   `xml:"NewX_AVM-DE_UpdateSuccessful"`
	NewX_AVM_DE_InfoURL                 string   `xml:"NewX_AVM-DE_InfoURL"`
	NewX_AVM_DE_MACAddressList          string   `xml:"NewX_AVM-DE_MACAddressList"`
	NewX_AVM_DE_Model                   string   `xml:"NewX_AVM-DE_Model"`
	NewX_AVM_DE_URL                     string   `xml:"NewX_AVM-DE_URL"`
	NewX_AVM_DE_Guest                   bool     `xml:"NewX_AVM-DE_Guest"`
	NewX_AVM_DE_RequestClient           bool     `xml:"NewX_AVM-DE_RequestClient"`
	NewX_AVM_DE_VPN                     bool     `xml:"NewX_AVM-DE_VPN"`
	NewX_AVM_DE_WANAccess               string   `xml:"NewX_AVM-DE_WANAccess"`
	NewX_AVM_DE_Disallow                bool     `xml:"NewX_AVM-DE_Disallow"`
	NewX_AVM_DE_IsMeshable              bool     `xml:"NewX_AVM-DE_IsMeshable"`
	NewX_AVM_DE_Priority                bool     `xml:"NewX_AVM-DE_Priority"`
	NewX_AVM_DE_FriendlyName            string   `xml:"NewX_AVM-DE_FriendlyName"`
	NewX_AVM_DE_FriendlyNameIsWriteable bool     `xml:"NewX_AVM-DE_FriendlyNameIsWriteable"`
}

func (client *ServiceClient) X_AVM_DE_GetSpecificHostEntryByIP(in *X_AVM_DE_GetSpecificHostEntryByIPRequest, out *X_AVM_DE_GetSpecificHostEntryByIPResponse) error {
	in.XMLNameSpace = client.Service.Type()
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetSpecificHostEntryByIP", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_HostsCheckUpdateRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_HostsCheckUpdateRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_HostsCheckUpdateResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_HostsCheckUpdateResponse"`
}

func (client *ServiceClient) X_AVM_DE_HostsCheckUpdate() error {
	in := &X_AVM_DE_HostsCheckUpdateRequest{XMLNameSpace: client.Service.Type()}
	out := &X_AVM_DE_HostsCheckUpdateResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_HostsCheckUpdate", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_HostDoUpdateRequest struct {
	XMLName       xml.Name `xml:"u:X_AVM-DE_HostDoUpdateRequest"`
	XMLNameSpace  string   `xml:"xmlns:u,attr"`
	NewMACAddress string   `xml:"NewMACAddress"`
}

type X_AVM_DE_HostDoUpdateResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_HostDoUpdateResponse"`
}

func (client *ServiceClient) X_AVM_DE_HostDoUpdate(in *X_AVM_DE_HostDoUpdateRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_HostDoUpdateResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_HostDoUpdate", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetPrioritizationByIPRequest struct {
	XMLName              xml.Name `xml:"u:X_AVM-DE_SetPrioritizationByIPRequest"`
	XMLNameSpace         string   `xml:"xmlns:u,attr"`
	NewIPAddress         string   `xml:"NewIPAddress"`
	NewX_AVM_DE_Priority bool     `xml:"NewX_AVM-DE_Priority"`
}

type X_AVM_DE_SetPrioritizationByIPResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetPrioritizationByIPResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetPrioritizationByIP(in *X_AVM_DE_SetPrioritizationByIPRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetPrioritizationByIPResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetPrioritizationByIP", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetHostListPathRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetHostListPathRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetHostListPathResponse struct {
	XMLName                  xml.Name `xml:"X_AVM-DE_GetHostListPathResponse"`
	NewX_AVM_DE_HostListPath string   `xml:"NewX_AVM-DE_HostListPath"`
}

func (client *ServiceClient) X_AVM_DE_GetHostListPath(out *X_AVM_DE_GetHostListPathResponse) error {
	in := &X_AVM_DE_GetHostListPathRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetHostListPath", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetMeshListPathRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetMeshListPathRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetMeshListPathResponse struct {
	XMLName                  xml.Name `xml:"X_AVM-DE_GetMeshListPathResponse"`
	NewX_AVM_DE_MeshListPath string   `xml:"NewX_AVM-DE_MeshListPath"`
}

func (client *ServiceClient) X_AVM_DE_GetMeshListPath(out *X_AVM_DE_GetMeshListPathResponse) error {
	in := &X_AVM_DE_GetMeshListPathRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetMeshListPath", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_GetFriendlyNameRequest struct {
	XMLName      xml.Name `xml:"u:X_AVM-DE_GetFriendlyNameRequest"`
	XMLNameSpace string   `xml:"xmlns:u,attr"`
}

type X_AVM_DE_GetFriendlyNameResponse struct {
	XMLName                  xml.Name `xml:"X_AVM-DE_GetFriendlyNameResponse"`
	NewX_AVM_DE_FriendlyName string   `xml:"NewX_AVM-DE_FriendlyName"`
}

func (client *ServiceClient) X_AVM_DE_GetFriendlyName(out *X_AVM_DE_GetFriendlyNameResponse) error {
	in := &X_AVM_DE_GetFriendlyNameRequest{XMLNameSpace: client.Service.Type()}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_GetFriendlyName", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetFriendlyNameRequest struct {
	XMLName                  xml.Name `xml:"u:X_AVM-DE_SetFriendlyNameRequest"`
	XMLNameSpace             string   `xml:"xmlns:u,attr"`
	NewX_AVM_DE_FriendlyName string   `xml:"NewX_AVM-DE_FriendlyName"`
}

type X_AVM_DE_SetFriendlyNameResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetFriendlyNameResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetFriendlyName(in *X_AVM_DE_SetFriendlyNameRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetFriendlyNameResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetFriendlyName", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetFriendlyNameByIPRequest struct {
	XMLName                  xml.Name `xml:"u:X_AVM-DE_SetFriendlyNameByIPRequest"`
	XMLNameSpace             string   `xml:"xmlns:u,attr"`
	NewIPAddress             string   `xml:"NewIPAddress"`
	NewX_AVM_DE_FriendlyName string   `xml:"NewX_AVM-DE_FriendlyName"`
}

type X_AVM_DE_SetFriendlyNameByIPResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetFriendlyNameByIPResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetFriendlyNameByIP(in *X_AVM_DE_SetFriendlyNameByIPRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetFriendlyNameByIPResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetFriendlyNameByIP", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}

type X_AVM_DE_SetFriendlyNameByMACRequest struct {
	XMLName                  xml.Name `xml:"u:X_AVM-DE_SetFriendlyNameByMACRequest"`
	XMLNameSpace             string   `xml:"xmlns:u,attr"`
	NewMACAddress            string   `xml:"NewMACAddress"`
	NewX_AVM_DE_FriendlyName string   `xml:"NewX_AVM-DE_FriendlyName"`
}

type X_AVM_DE_SetFriendlyNameByMACResponse struct {
	XMLName xml.Name `xml:"X_AVM-DE_SetFriendlyNameByMACResponse"`
}

func (client *ServiceClient) X_AVM_DE_SetFriendlyNameByMAC(in *X_AVM_DE_SetFriendlyNameByMACRequest) error {
	in.XMLNameSpace = client.Service.Type()
	out := &X_AVM_DE_SetFriendlyNameByMACResponse{}
	return client.TR064Client.InvokeService(client.Service, "X_AVM-DE_SetFriendlyNameByMAC", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))
}
