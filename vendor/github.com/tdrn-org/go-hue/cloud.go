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
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"time"
)

// NewCloudBridgeLocator creates a new [CloudBridgeLocator] for discovering local bridges via the Hue Cloud's [Discovery endpoint].
//
// Only bridges registered in the cloud are locatable via this [BridgeLocator].
//
// Example:
//
//	locator := hue.NewCloudBridgeLocator()
//	bridges, _ := locator.Locate(hue.DefaultTimeout)
//	for _, bridge := range bridges {
//		client, _ := bridge.NewClient(new hue.NewLocalAuthenticator(""), hue.DefaultTimeout)
//		// make sure linking button is pressed before invoking Authenticate
//		hostname, _ := os.Hostnamee()
//		deviceType := "MyApp#" + hostname
//		generateClientKey := true
//		request := hueapi.AuthenticateJSONRequestBody{
//			Devicetype:        &deviceType,
//			Generateclientkey: &generateClientKey,
//		}
//		response, _ := client.Authenticate(request)
//		if response.response.HTTPResponse.StatusCode == http.StatusOK {
//			success := (*rsp.JSON200)[0].Success
//			fmt.Println("Bridge id: ", bridge.BridgeId)
//			fmt.Println("Username: ", *rspSuccess.Username)
//		}
//		// Authentication username is automatically picked up by the client. All API calls are now possible.
//		getDevicesResponse, _ := client.GetDevices()
//	}
//
//	// If Bridge Id and Username are already known, this can be shortened to
//
//	locator := hue.NewCloudBridgeLocator()
//	bridge, _ := locator.Lookup("0123456789ABCDEF", hue.DefaultTimeout)
//	client, _ := bridge.NewClient(hue.NewLocalAuthenticator("secret username token"), hue.DefaultTimeout)
//	getDevicesResponse, _ := client.GetDevices()
//
// [Discovery endpoint]: https://developers.meethue.com/develop/application-design-guidance/hue-bridge-discovery/#Disocvery%20Endpoint
func NewCloudBridgeLocator() *CloudBridgeLocator {
	logger := slog.With(slog.String("locator", cloudBridgeLocatorName))
	return &CloudBridgeLocator{
		DiscoveryEndpointUrl: cloudDefaultDiscoveryEndpointUrl,
		logger:               logger,
	}
}

const cloudBridgeLocatorName string = "cloud"

// CloudBridgeLocator locates local bridges via the Hue Cloud's [Discovery endpoint].
//
// Use [NewCloudBridgeLocator] to create a new instance.
//
// [Discovery endpoint]: https://developers.meethue.com/develop/application-design-guidance/hue-bridge-discovery/#Disocvery%20Endpoint
type CloudBridgeLocator struct {
	// DiscoveryEndpointUrl defines the discovery endpoint URL to use. This URL defaults to https://discovery.meethue.com and may be
	// overwritten for local testing.
	DiscoveryEndpointUrl *url.URL
	// TlsConfig defines the TLS configuration to use for accessing the endpoint URL. If nil, the standard options are used.
	TlsConfig *tls.Config
	logger    *slog.Logger
}

func (locator *CloudBridgeLocator) Name() string {
	return cloudBridgeLocatorName
}

func (locator *CloudBridgeLocator) Query(timeout time.Duration) ([]*Bridge, error) {
	locator.logger.Info("discovering bridges...", slog.Any("discovery_endpoint", locator.DiscoveryEndpointUrl))
	discoveredEntries, err := locator.queryDiscoveryEndpoint(timeout)
	if err != nil {
		return nil, err
	}
	bridges := make([]*Bridge, 0, len(discoveredEntries))
	for _, discoveredEntry := range discoveredEntries {
		url, err := discoveredEntry.toUrl()
		if err != nil {
			locator.logger.Error("ignoring invalid response entry", slog.Any("entry", discoveredEntry), slog.Any("err", err))
			continue
		}
		config, err := queryAndValidateLocalBridgeConfig(url, discoveredEntry.Id, timeout)
		if err != nil {
			locator.logger.Error("ignoring response entry", slog.Any("entry", discoveredEntry), slog.Any("err", err))
			continue
		}
		bridge, err := config.newBridge(locator, url)
		if err != nil {
			return nil, err
		}
		locator.logger.Info("located bridge", slog.Any("bridge", bridge))
		bridges = append(bridges, bridge)
	}
	return bridges, nil
}

func (locator *CloudBridgeLocator) Lookup(bridgeId string, timeout time.Duration) (*Bridge, error) {
	locator.logger.Info("looking up bridge...", slog.String("bridge_id", bridgeId), slog.Any("discovery_endpoint", locator.DiscoveryEndpointUrl))
	discoveredEntries, err := locator.queryDiscoveryEndpoint(timeout)
	if err != nil {
		return nil, err
	}
	for _, discoveredEntry := range discoveredEntries {
		if discoveredEntry.Id != bridgeId {
			continue
		}
		url, err := discoveredEntry.toUrl()
		if err != nil {
			locator.logger.Info("bridge entry not valid", slog.Any("entry", discoveredEntry), slog.Any("err", err))
			return nil, ErrBridgeNotAvailable
		}
		config, err := queryAndValidateLocalBridgeConfig(url, discoveredEntry.Id, timeout)
		if err != nil {
			locator.logger.Info("bridge not available", slog.String("bridge_id", bridgeId), slog.Any("err", err))
			return nil, ErrBridgeNotAvailable
		}
		bridge, err := config.newBridge(locator, url)
		if err != nil {
			return nil, err
		}
		locator.logger.Info("located bridge", slog.Any("bridge", bridge))
		return bridge, nil
	}
	return nil, ErrBridgeNotAvailable
}

func (locator *CloudBridgeLocator) NewClient(bridge *Bridge, authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error) {
	return newLocalBridgeHueClient(bridge, authenticator, timeout)
}

func (locator *CloudBridgeLocator) queryDiscoveryEndpoint(timeout time.Duration) ([]cloudDiscoveryEndpointResponseEntry, error) {
	response := make([]cloudDiscoveryEndpointResponseEntry, 0)
	err := fetchJson(newDefaultClient(timeout, locator.TlsConfig), locator.DiscoveryEndpointUrl, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

var cloudDefaultDiscoveryEndpointUrl *url.URL = initCloudDefaultDiscoveryEndpointUrl()

func initCloudDefaultDiscoveryEndpointUrl() *url.URL {
	url, err := url.Parse("https://discovery.meethue.com/")
	if err != nil {
		log.Fatal(err)
	}
	return url
}

type cloudDiscoveryEndpointResponseEntry struct {
	Id                string `json:"id"`
	InternalIpAddress string `json:"internalipaddress"`
	Port              int    `json:"port"`
}

func (entry *cloudDiscoveryEndpointResponseEntry) toUrl() (*url.URL, error) {
	address := net.JoinHostPort(entry.InternalIpAddress, strconv.Itoa(entry.Port))
	url, err := url.Parse("https://" + address + "/")
	if err != nil {
		return nil, fmt.Errorf("invalid address '%s' (cause: %w)", address, err)
	}
	return url, err
}
