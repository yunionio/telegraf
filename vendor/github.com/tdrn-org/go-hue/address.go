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
	"fmt"
	"log/slog"
	"net/url"
	"time"
)

// NewAddressBridgeLocator creates a new [AddressBridgeLocator] for accessing a local bridge via a [well-known address].
//
// Example:
//
//	locator, _ := hue.NewAddressBridgeLocator("196.168.1.127")
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
//	locator, _ := hue.NewAddressBridgeLocator("196.168.1.127")
//	bridge, _ := locator.Lookup("0123456789ABCDEF", hue.DefaultTimeout)
//	client, _ := bridge.NewClient(hue.NewLocalAuthenticator("secret username token"), hue.DefaultTimeout)
//	getDevicesResponse, _ := client.GetDevices()
//
// [well-known address]: https://developers.meethue.com/develop/application-design-guidance/hue-bridge-discovery/#Manual-ip
func NewAddressBridgeLocator(address string) (*AddressBridgeLocator, error) {
	logger := slog.With(slog.String("locator", addressBridgeLocatorName))
	url, err := url.Parse("https://" + address + "/")
	if err != nil {
		return nil, fmt.Errorf("invalid address '%s' (cause: %w)", address, err)
	}
	return &AddressBridgeLocator{
		url:    url,
		logger: logger,
	}, nil
}

const addressBridgeLocatorName string = "address"

// AddressBridgeLocator locates a local bridge via a well-known address.
//
// Use [NewAddressBridgeLocator] to create a new instance. As this locator is looking at exactly one bridge,
// a Query call will return not more than one brigde.
type AddressBridgeLocator struct {
	url    *url.URL
	logger *slog.Logger
}

func (locator *AddressBridgeLocator) Name() string {
	return addressBridgeLocatorName
}

func (locator *AddressBridgeLocator) Query(timeout time.Duration) ([]*Bridge, error) {
	bridge, err := locator.Lookup("", timeout)
	if err != nil {
		return []*Bridge{}, nil
	}
	return []*Bridge{bridge}, nil
}

func (locator *AddressBridgeLocator) Lookup(bridgeId string, timeout time.Duration) (*Bridge, error) {
	locator.logger.Info("probing bridge...", slog.Any("url", locator.url))
	config, err := queryAndValidateLocalBridgeConfig(locator.url, bridgeId, timeout)
	if err != nil {
		locator.logger.Info("bridge not available", slog.String("bridge_id", bridgeId), slog.Any("err", err))
		return nil, ErrBridgeNotAvailable
	}
	bridge, err := config.newBridge(locator, locator.url)
	if err != nil {
		return nil, err
	}
	locator.logger.Info("located bridge", slog.Any("bridge", bridge))
	return bridge, nil
}

func (locator *AddressBridgeLocator) NewClient(bridge *Bridge, authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error) {
	return newLocalBridgeHueClient(bridge, authenticator, timeout)
}
