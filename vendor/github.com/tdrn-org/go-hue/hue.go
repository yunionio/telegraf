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

package hue

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/tdrn-org/go-hue/hueapi"
)

// ErrBridgeNotAvailable indicates a bridge is currently not available.
var ErrBridgeNotAvailable = errors.New("bridge not available")

// ErrBridgeClientFailure indicates a system error while invoking the bridge client.
var ErrBridgeClientFailure = errors.New("bridge client call failure")

// ErrNotAuthenticated indicates bridge access has not yet been authenticated
var ErrNotAuthenticated = errors.New("not authenticated")

// ErrHueAPIFailure indicates an API call has failed with an API error.
var ErrHueAPIFailure = errors.New("api failure")

// DefaultTimeout defines a suitable default for timeout related functions.
const DefaultTimeout time.Duration = 60 * time.Second

// Bridge contains the general bridge attributes as well the [BridgeLocator] instance used to locate this bridge.
type Bridge struct {
	// Locator refers to the locator instance used to identify this bridge.
	Locator BridgeLocator
	// Name contains the name of the bridge.
	Name string
	// SoftwareVersion contains the version of the bridge SW.
	SoftwareVersion string
	// ApiVersion contains the version of the bridge API.
	ApiVersion string
	// HardwareAddress contains the Hardware (MAC) address of the bridge.
	HardwareAddress net.HardwareAddr
	// BridgeId contains the id of the bridge.
	BridgeId string
	// ReplacesBridgeId contains optionally the id of the bridge replaced by this one.
	ReplacesBridgeId string
	// ModelId contains the id of the bridge model.
	ModelId string
	// Url contains the URL used to access the bridge.
	Url *url.URL
}

// NewClient creates a new [BridgeClient] suitable for accessing the bridge services.
func (bridge *Bridge) NewClient(authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error) {
	return bridge.Locator.NewClient(bridge, authenticator, timeout)
}

// String gets the bridge's signature string.
func (bridge *Bridge) String() string {
	return fmt.Sprintf("%s:%s (Name: '%s', SW: %s, API: %s, MAC: %s, URL: %s)", bridge.Locator.Name(), bridge.BridgeId, bridge.Name, bridge.SoftwareVersion, bridge.ApiVersion, bridge.HardwareAddress.String(), bridge.Url)
}

// BridgeAuthenticator injects the necessary authentication credentials into an bridge API call.
type BridgeAuthenticator interface {
	// AuthenticateRequest authenticates the given request.
	AuthenticateRequest(ctx context.Context, req *http.Request) error
	// Authenticated is called with the response of an Authenticate API call and updates this instance's authentication credentials.
	Authenticated(rsp *hueapi.AuthenticateResponse)
	// Authentication returns the user name used to authenticate towards the bridge. Error ErrNotAuthenticated indicates, bridge access
	// has not yet been authenticated.
	Authentication() (string, error)
}

// BridgeLocator provides the necessary functions to identify and access bridge instances.
type BridgeLocator interface {
	// Name gets the name of the locator.
	Name() string
	// Query locates all accessible bridges.
	//
	// An empty collection is returned, in case no bridge could be located.
	Query(timeout time.Duration) ([]*Bridge, error)
	// Lookup locates the bridge for the given bridge id.
	//
	// An error is returned in case the bridge is not available.
	Lookup(bridgeId string, timeout time.Duration) (*Bridge, error)
	// NewClient create a new bridge client for accessing the given bridge's services.
	NewClient(bridge *Bridge, authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error)
}

type bridgeConfig struct {
	Name             string `json:"name"`
	SwVersion        string `json:"swversion"`
	ApiVersion       string `json:"apiversion"`
	Mac              string `json:"mac"`
	BridgeId         string `json:"bridgeid"`
	FactoryNew       bool   `json:"factorynew"`
	ReplacesBridgeId string `json:"replacesbridgeid"`
	ModelId          string `json:"modelid"`
}

func (config *bridgeConfig) newBridge(locator BridgeLocator, url *url.URL) (*Bridge, error) {
	hardwareAddress, err := net.ParseMAC(config.Mac)
	if err != nil {
		return nil, fmt.Errorf("invalid hardware address '%s' in bridge config (cause: %w)", config.Mac, err)
	}
	return &Bridge{
		Locator:          locator,
		Name:             config.Name,
		SoftwareVersion:  config.SwVersion,
		ApiVersion:       config.ApiVersion,
		HardwareAddress:  hardwareAddress,
		BridgeId:         config.BridgeId,
		ReplacesBridgeId: config.ReplacesBridgeId,
		ModelId:          config.ModelId,
		Url:              url,
	}, nil
}

func configUrl(url *url.URL) *url.URL {
	return url.JoinPath("/api/0/config")
}

func fetchJson(client *http.Client, url *url.URL, v interface{}) error {
	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to prepare request for URL '%s' (cause: %w)", url, err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to query URL '%s' (cause: %w)", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to query URL '%s' (status: %s)", url, response.Status)
	}
	err = json.NewDecoder(response.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("failed to decode response body (cause: %w)", err)
	}
	return nil
}

func newDefaultClient(timeout time.Duration, tlsConfig *tls.Config) *http.Client {
	transport := &http.Transport{
		TLSClientConfig:       tlsConfig.Clone(),
		ResponseHeaderTimeout: timeout,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// BridgeClient provides the Hue API functions provided by a bridge.
type BridgeClient interface {
	// Bridge gets the bridge instance this client accesses.
	Bridge() *Bridge
	// Url gets URL used to access the bridge services.
	Url() *url.URL
	// HttpClient gets the underlying http client used to access the bridge.
	HttpClient() *http.Client
	// Authenticate API call.
	Authenticate(request hueapi.AuthenticateJSONRequestBody) (*hueapi.AuthenticateResponse, error)
	// GetResources API call.
	GetResources() (*hueapi.GetResourcesResponse, error)
	// GetBridges API call.
	GetBridges() (*hueapi.GetBridgesResponse, error)
	// GetBridge API call.
	GetBridge(bridgeId string) (*hueapi.GetBridgeResponse, error)
	// UpdateBridge API call.
	UpdateBridge(bridgeId string, body hueapi.UpdateBridgeJSONRequestBody) (*hueapi.UpdateBridgeResponse, error)
	// GetBridgeHomes API call.
	GetBridgeHomes() (*hueapi.GetBridgeHomesResponse, error)
	// GetBridgeHome API call.
	GetBridgeHome(bridgeHomeId string) (*hueapi.GetBridgeHomeResponse, error)
	// GetDevices API call.
	GetDevices() (*hueapi.GetDevicesResponse, error)
	// DeleteDevice API call.
	DeleteDevice(deviceId string) (*hueapi.DeleteDeviceResponse, error)
	// GetDevice API call.
	GetDevice(deviceId string) (*hueapi.GetDeviceResponse, error)
	// UpdateDevice API call.
	UpdateDevice(deviceId string, body hueapi.UpdateDeviceJSONRequestBody) (*hueapi.UpdateDeviceResponse, error)
	// GetDevicePowers API call.
	GetDevicePowers() (*hueapi.GetDevicePowersResponse, error)
	// GetDevicePower API call.
	GetDevicePower(deviceId string) (*hueapi.GetDevicePowerResponse, error)
	// GetGroupedLights API call.
	GetGroupedLights() (*hueapi.GetGroupedLightsResponse, error)
	// GetGroupedLight API call.
	GetGroupedLight(groupedLightId string) (*hueapi.GetGroupedLightResponse, error)
	// UpdateGroupedLight API call.
	UpdateGroupedLight(groupedLightId string, body hueapi.UpdateGroupedLightJSONRequestBody) (*hueapi.UpdateGroupedLightResponse, error)
	// GetLights API call.
	GetLights() (*hueapi.GetLightsResponse, error)
	// GetLight API call.
	GetLight(lightId string) (*hueapi.GetLightResponse, error)
	// UpdateLight API call.
	UpdateLight(lightId string, body hueapi.UpdateLightJSONRequestBody) (*hueapi.UpdateLightResponse, error)
	// GetLightLevels API call.
	GetLightLevels() (*hueapi.GetLightLevelsResponse, error)
	// GetLightLevel API call.
	GetLightLevel(lightId string) (*hueapi.GetLightLevelResponse, error)
	// UpdateLightLevel API call.
	UpdateLightLevel(lightId string, body hueapi.UpdateLightLevelJSONRequestBody) (*hueapi.UpdateLightLevelResponse, error)
	// GetMotionSensors API call.
	GetMotionSensors() (*hueapi.GetMotionSensorsResponse, error)
	// GetMotionSensor API call.
	GetMotionSensor(motionId string) (*hueapi.GetMotionSensorResponse, error)
	// UpdateMotionSensor API call.
	UpdateMotionSensor(motionId string, body hueapi.UpdateMotionSensorJSONRequestBody) (*hueapi.UpdateMotionSensorResponse, error)
	// GetRooms API call.
	GetRooms() (*hueapi.GetRoomsResponse, error)
	// CreateRoom API call.
	CreateRoom(body hueapi.CreateRoomJSONRequestBody) (*hueapi.CreateRoomResponse, error)
	// DeleteRoom API call.
	DeleteRoom(roomId string) (*hueapi.DeleteRoomResponse, error)
	// GetRoom API call.
	GetRoom(roomId string) (*hueapi.GetRoomResponse, error)
	// UpdateRoom API call.
	UpdateRoom(roomId string, body hueapi.UpdateRoomJSONRequestBody) (*hueapi.UpdateRoomResponse, error)
	// GetScenes API call.
	GetScenes() (*hueapi.GetScenesResponse, error)
	// CreateScene API call.
	CreateScene(body hueapi.CreateSceneJSONRequestBody) (*hueapi.CreateSceneResponse, error)
	// DeleteScene API call.
	DeleteScene(sceneId string) (*hueapi.DeleteSceneResponse, error)
	// GetScene API call.
	GetScene(sceneId string) (*hueapi.GetSceneResponse, error)
	// UpdateScene API call.
	UpdateScene(sceneId string, body hueapi.UpdateSceneJSONRequestBody) (*hueapi.UpdateSceneResponse, error)
	// GetSmartScenes API call.
	GetSmartScenes() (*hueapi.GetSmartScenesResponse, error)
	// CreateSmartScene API call.
	CreateSmartScene(body hueapi.CreateSmartSceneJSONRequestBody) (*hueapi.CreateSmartSceneResponse, error)
	// DeleteSmartScene API call.
	DeleteSmartScene(sceneId string) (*hueapi.DeleteSmartSceneResponse, error)
	// GetSmartScene API call.
	GetSmartScene(sceneId string) (*hueapi.GetSmartSceneResponse, error)
	// UpdateSmartScene API call.
	UpdateSmartScene(sceneId string, body hueapi.UpdateSmartSceneJSONRequestBody) (*hueapi.UpdateSmartSceneResponse, error)
	// GetTemperatures API call.
	GetTemperatures() (*hueapi.GetTemperaturesResponse, error)
	// GetTemperature API call.
	GetTemperature(temperatureId string) (*hueapi.GetTemperatureResponse, error)
	// UpdateTemperature API call.
	UpdateTemperature(temperatureId string, body hueapi.UpdateTemperatureJSONRequestBody) (*hueapi.UpdateTemperatureResponse, error)
	// GetZones API call.
	GetZones() (*hueapi.GetZonesResponse, error)
	// CreateZone API call.
	CreateZone(body hueapi.CreateZoneJSONRequestBody) (*hueapi.CreateZoneResponse, error)
	// DeleteZone API call.
	DeleteZone(zoneId string) (*hueapi.DeleteZoneResponse, error)
	// GetZone API call.
	GetZone(zoneId string) (*hueapi.GetZoneResponse, error)
	// UpdateZone API call.
	UpdateZone(zoneId string, body hueapi.UpdateZoneJSONRequestBody) (*hueapi.UpdateZoneResponse, error)
}

type bridgeClient struct {
	bridge        *Bridge
	url           *url.URL
	httpClient    *http.Client
	apiClient     hueapi.ClientWithResponsesInterface
	authenticator BridgeAuthenticator
}

func (client *bridgeClient) Bridge() *Bridge {
	return client.bridge
}

func (client *bridgeClient) Url() *url.URL {
	return client.url
}

func (client *bridgeClient) HttpClient() *http.Client {
	return client.httpClient
}

func bridgeClientApiName() string {
	pc, _, _, _ := runtime.Caller(2)
	rawCaller := runtime.FuncForPC(pc).Name()
	return rawCaller[strings.LastIndex(rawCaller, ".")+1:]
}

func bridgeClientWrapSystemError(err error) error {
	return fmt.Errorf("%w %s (cause: %w)", ErrBridgeClientFailure, bridgeClientApiName(), err)
}

func bridgeClientApiError(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden:
		return fmt.Errorf("%w %s(...) (status: %s)", ErrNotAuthenticated, bridgeClientApiName(), response.Status)
	default:
		return fmt.Errorf("%w %s(...) (status: %s)", ErrHueAPIFailure, bridgeClientApiName(), response.Status)
	}
}

func (client *bridgeClient) Authenticate(request hueapi.AuthenticateJSONRequestBody) (*hueapi.AuthenticateResponse, error) {
	response, err := client.apiClient.AuthenticateWithResponse(context.Background(), request)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	client.authenticator.Authenticated(response)
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetResources() (*hueapi.GetResourcesResponse, error) {
	response, err := client.apiClient.GetResourcesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetBridges() (*hueapi.GetBridgesResponse, error) {
	response, err := client.apiClient.GetBridgesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetBridge(bridgeId string) (*hueapi.GetBridgeResponse, error) {
	response, err := client.apiClient.GetBridgeWithResponse(context.Background(), bridgeId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateBridge(bridgeId string, body hueapi.UpdateBridgeJSONRequestBody) (*hueapi.UpdateBridgeResponse, error) {
	response, err := client.apiClient.UpdateBridgeWithResponse(context.Background(), bridgeId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetBridgeHomes() (*hueapi.GetBridgeHomesResponse, error) {
	response, err := client.apiClient.GetBridgeHomesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetBridgeHome(bridgeHomeId string) (*hueapi.GetBridgeHomeResponse, error) {
	response, err := client.apiClient.GetBridgeHomeWithResponse(context.Background(), bridgeHomeId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetDevices() (*hueapi.GetDevicesResponse, error) {
	response, err := client.apiClient.GetDevicesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) DeleteDevice(deviceId string) (*hueapi.DeleteDeviceResponse, error) {
	response, err := client.apiClient.DeleteDeviceWithResponse(context.Background(), deviceId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetDevice(deviceId string) (*hueapi.GetDeviceResponse, error) {
	response, err := client.apiClient.GetDeviceWithResponse(context.Background(), deviceId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateDevice(deviceId string, body hueapi.UpdateDeviceJSONRequestBody) (*hueapi.UpdateDeviceResponse, error) {
	response, err := client.apiClient.UpdateDeviceWithResponse(context.Background(), deviceId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetDevicePowers() (*hueapi.GetDevicePowersResponse, error) {
	response, err := client.apiClient.GetDevicePowersWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetDevicePower(deviceId string) (*hueapi.GetDevicePowerResponse, error) {
	response, err := client.apiClient.GetDevicePowerWithResponse(context.Background(), deviceId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetGroupedLights() (*hueapi.GetGroupedLightsResponse, error) {
	response, err := client.apiClient.GetGroupedLightsWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetGroupedLight(groupedLightId string) (*hueapi.GetGroupedLightResponse, error) {
	response, err := client.apiClient.GetGroupedLightWithResponse(context.Background(), groupedLightId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateGroupedLight(groupedLightId string, body hueapi.UpdateGroupedLightJSONRequestBody) (*hueapi.UpdateGroupedLightResponse, error) {
	response, err := client.apiClient.UpdateGroupedLightWithResponse(context.Background(), groupedLightId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetLights() (*hueapi.GetLightsResponse, error) {
	response, err := client.apiClient.GetLightsWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetLight(lightId string) (*hueapi.GetLightResponse, error) {
	response, err := client.apiClient.GetLightWithResponse(context.Background(), lightId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateLight(lightId string, body hueapi.UpdateLightJSONRequestBody) (*hueapi.UpdateLightResponse, error) {
	response, err := client.apiClient.UpdateLightWithResponse(context.Background(), lightId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetLightLevels() (*hueapi.GetLightLevelsResponse, error) {
	response, err := client.apiClient.GetLightLevelsWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetLightLevel(lightId string) (*hueapi.GetLightLevelResponse, error) {
	response, err := client.apiClient.GetLightLevelWithResponse(context.Background(), lightId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateLightLevel(lightId string, body hueapi.UpdateLightLevelJSONRequestBody) (*hueapi.UpdateLightLevelResponse, error) {
	response, err := client.apiClient.UpdateLightLevelWithResponse(context.Background(), lightId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetMotionSensors() (*hueapi.GetMotionSensorsResponse, error) {
	response, err := client.apiClient.GetMotionSensorsWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetMotionSensor(motionId string) (*hueapi.GetMotionSensorResponse, error) {
	response, err := client.apiClient.GetMotionSensorWithResponse(context.Background(), motionId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateMotionSensor(motionId string, body hueapi.UpdateMotionSensorJSONRequestBody) (*hueapi.UpdateMotionSensorResponse, error) {
	response, err := client.apiClient.UpdateMotionSensorWithResponse(context.Background(), motionId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetRooms() (*hueapi.GetRoomsResponse, error) {
	response, err := client.apiClient.GetRoomsWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) CreateRoom(body hueapi.CreateRoomJSONRequestBody) (*hueapi.CreateRoomResponse, error) {
	response, err := client.apiClient.CreateRoomWithResponse(context.Background(), body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) DeleteRoom(roomId string) (*hueapi.DeleteRoomResponse, error) {
	response, err := client.apiClient.DeleteRoomWithResponse(context.Background(), roomId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetRoom(roomId string) (*hueapi.GetRoomResponse, error) {
	response, err := client.apiClient.GetRoomWithResponse(context.Background(), roomId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateRoom(roomId string, body hueapi.UpdateRoomJSONRequestBody) (*hueapi.UpdateRoomResponse, error) {
	response, err := client.apiClient.UpdateRoomWithResponse(context.Background(), roomId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetScenes() (*hueapi.GetScenesResponse, error) {
	response, err := client.apiClient.GetScenesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) CreateScene(body hueapi.CreateSceneJSONRequestBody) (*hueapi.CreateSceneResponse, error) {
	response, err := client.apiClient.CreateSceneWithResponse(context.Background(), body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) DeleteScene(sceneId string) (*hueapi.DeleteSceneResponse, error) {
	response, err := client.apiClient.DeleteSceneWithResponse(context.Background(), sceneId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetScene(sceneId string) (*hueapi.GetSceneResponse, error) {
	response, err := client.apiClient.GetSceneWithResponse(context.Background(), sceneId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateScene(sceneId string, body hueapi.UpdateSceneJSONRequestBody) (*hueapi.UpdateSceneResponse, error) {
	response, err := client.apiClient.UpdateSceneWithResponse(context.Background(), sceneId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetSmartScenes() (*hueapi.GetSmartScenesResponse, error) {
	response, err := client.apiClient.GetSmartScenesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) CreateSmartScene(body hueapi.CreateSmartSceneJSONRequestBody) (*hueapi.CreateSmartSceneResponse, error) {
	response, err := client.apiClient.CreateSmartSceneWithResponse(context.Background(), body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) DeleteSmartScene(sceneId string) (*hueapi.DeleteSmartSceneResponse, error) {
	response, err := client.apiClient.DeleteSmartSceneWithResponse(context.Background(), sceneId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetSmartScene(sceneId string) (*hueapi.GetSmartSceneResponse, error) {
	response, err := client.apiClient.GetSmartSceneWithResponse(context.Background(), sceneId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateSmartScene(sceneId string, body hueapi.UpdateSmartSceneJSONRequestBody) (*hueapi.UpdateSmartSceneResponse, error) {
	response, err := client.apiClient.UpdateSmartSceneWithResponse(context.Background(), sceneId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetTemperatures() (*hueapi.GetTemperaturesResponse, error) {
	response, err := client.apiClient.GetTemperaturesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetTemperature(temperatureId string) (*hueapi.GetTemperatureResponse, error) {
	response, err := client.apiClient.GetTemperatureWithResponse(context.Background(), temperatureId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateTemperature(temperatureId string, body hueapi.UpdateTemperatureJSONRequestBody) (*hueapi.UpdateTemperatureResponse, error) {
	response, err := client.apiClient.UpdateTemperatureWithResponse(context.Background(), temperatureId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetZones() (*hueapi.GetZonesResponse, error) {
	response, err := client.apiClient.GetZonesWithResponse(context.Background(), client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) CreateZone(body hueapi.CreateZoneJSONRequestBody) (*hueapi.CreateZoneResponse, error) {
	response, err := client.apiClient.CreateZoneWithResponse(context.Background(), body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) DeleteZone(zoneId string) (*hueapi.DeleteZoneResponse, error) {
	response, err := client.apiClient.DeleteZoneWithResponse(context.Background(), zoneId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) GetZone(zoneId string) (*hueapi.GetZoneResponse, error) {
	response, err := client.apiClient.GetZoneWithResponse(context.Background(), zoneId, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}

func (client *bridgeClient) UpdateZone(zoneId string, body hueapi.UpdateZoneJSONRequestBody) (*hueapi.UpdateZoneResponse, error) {
	response, err := client.apiClient.UpdateZoneWithResponse(context.Background(), zoneId, body, client.authenticator.AuthenticateRequest)
	if err != nil {
		return nil, bridgeClientWrapSystemError(err)
	}
	return response, bridgeClientApiError(response.HTTPResponse)
}
