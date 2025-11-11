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
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/brutella/dnssd"
)

// NewMDNSBridgeLocator creates a new [MdnsBridgeLocator] for discovering bridges via [Multicast DNS (mDNS)].
//
// Example:
//
//	locator := NewMDNSBridgeLocator()
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
//	locator := hue.NewMDNSBridgeLocator()
//	bridge, _ := locator.Lookup("0123456789ABCDEF", hue.DefaultTimeout)
//	client, _ := bridge.NewClient(hue.NewLocalAuthenticator("secret username token"), hue.DefaultTimeout)
//	getDevicesResponse, _ := client.GetDevices()
//
// [Multicast DNS (mDNS)]: https://developers.meethue.com/develop/application-design-guidance/hue-bridge-discovery/#mDNS
func NewMDNSBridgeLocator() *MdnsBridgeLocator {
	logger := slog.With(slog.String("locator", mdnsBridgeLocatorName))
	return &MdnsBridgeLocator{
		Limit:  0,
		logger: logger,
	}
}

const mdnsBridgeLocatorName string = "mDNS"

// MdnsBridgeLocator locates local bridges via via [Multicast DNS (mDNS)].
//
// Use [NewMDNSBridgeLocator] to create a new instance.
//
// [Multicast DNS (mDNS)]: https://developers.meethue.com/develop/application-design-guidance/hue-bridge-discovery/#mDNS
type MdnsBridgeLocator struct {
	// Limit defines the maxmimum number of bridges to return during a [BridgeLocator.Query] call. 0 means no limit.
	// As mDNS is working asynchronously, a query normally continues until the given timeout is reached. If a limit is
	// set, a query is complete as soon as the limit is reached.
	Limit  int
	logger *slog.Logger
}

func (locator *MdnsBridgeLocator) Name() string {
	return mdnsBridgeLocatorName
}

const mdnsHueService string = "_hue._tcp.local."

func (locator *MdnsBridgeLocator) Query(timeout time.Duration) ([]*Bridge, error) {
	locator.logger.Info("discovering service...", slog.String("service", mdnsHueService))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	bridges := make([]*Bridge, 0)
	add := func(entry dnssd.BrowseEntry) {
		locator.logger.Info("detected service", slog.String("instance", entry.ServiceInstanceName()), slog.Any("text", entry.Text))
		url, config, err := locator.queryAndValidateBridgeConfig(&entry, timeout)
		if err != nil {
			locator.logger.Info("ignoring invalid service", slog.String("name", entry.Name), slog.Any("err", err))
			return
		}
		bridge, err := config.newBridge(locator, url)
		if err != nil {
			locator.logger.Info("failed to decode service", slog.String("name", entry.Name), slog.Any("err", err))
			return
		}
		locator.logger.Info("located bridge", slog.Any("bridge", bridge))
		bridges = append(bridges, bridge)
		if locator.Limit > 0 && len(bridges) >= locator.Limit {
			cancel()
		}
	}
	rmv := func(entry dnssd.BrowseEntry) {
		// nothing to do here
	}
	err := dnssd.LookupType(ctx, mdnsHueService, add, rmv)
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return nil, err
	}
	return bridges, nil
}

func (locator *MdnsBridgeLocator) Lookup(bridgeId string, timeout time.Duration) (*Bridge, error) {
	locator.logger.Info("looking up bridge...", slog.String("bridge_id", bridgeId), slog.String("service", mdnsHueService))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var bridge *Bridge
	add := func(entry dnssd.BrowseEntry) {
		locator.logger.Info("detected service", slog.String("instance", entry.ServiceInstanceName()), slog.Any("text", entry.Text))
		serviceBridgeId := locator.browseEntryBridgeId(&entry)
		if serviceBridgeId != bridgeId {
			return
		}
		url, config, err := locator.queryAndValidateBridgeConfig(&entry, timeout)
		if err != nil {
			locator.logger.Info("ignoring invalid service", slog.String("name", entry.Name), slog.Any("err", err))
			return
		}
		bridge, err = config.newBridge(locator, url)
		if err != nil {
			locator.logger.Info("failed to decode service", slog.String("name", entry.Name), slog.Any("err", err))
			return
		}
		locator.logger.Info("located bridge", slog.Any("bridge", bridge))
		cancel()
	}
	rmv := func(entry dnssd.BrowseEntry) {
		// nothing to do here
	}
	err := dnssd.LookupType(ctx, mdnsHueService, add, rmv)
	ctxErr := ctx.Err()
	if err != nil && !(errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded)) {
		return nil, err
	}
	return bridge, nil
}

func (locator *MdnsBridgeLocator) NewClient(bridge *Bridge, authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error) {
	return newLocalBridgeHueClient(bridge, authenticator, timeout)
}

func (locator *MdnsBridgeLocator) queryAndValidateBridgeConfig(entry *dnssd.BrowseEntry, timeout time.Duration) (*url.URL, *bridgeConfig, error) {
	if len(entry.IPs) == 0 {
		return nil, nil, fmt.Errorf("addressless service '%s'", entry.Name)
	}
	ip := entry.IPs[0]
	address := net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port))
	url, err := url.Parse("https://" + address + "/")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compose URL for address '%s' (cause: %w)", address, err)
	}
	bridgeId := locator.browseEntryBridgeId(entry)
	config, err := queryAndValidateLocalBridgeConfig(url, bridgeId, timeout)
	if err != nil {
		return nil, nil, err
	}
	return url, config, nil
}

func (locator *MdnsBridgeLocator) browseEntryBridgeId(entry *dnssd.BrowseEntry) string {
	return entry.Text["bridgeid"]
}
