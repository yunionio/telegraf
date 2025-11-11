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
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tdrn-org/go-hue/hueapi"
)

// NewLocalBridgeAuthenticator creates a new [LocalBridgeAuthenticator] suitable for authenticating towards a local bridge.
//
// The user name must be empty or previously been created via a successful [Authenticate] API call. Everytime a
// successful [Authenticate] API call is performed, the user name will be overwritten by the returned user name.
//
// [Authenticate]: https://developers.meethue.com/develop/hue-api/7-configuration-api/#create-user
func NewLocalBridgeAuthenticator(userName string) *LocalBridgeAuthenticator {
	logger := slog.With(slog.String("authenticator", "local"))
	return &LocalBridgeAuthenticator{
		UserName: userName,
		logger:   logger,
	}
}

// LocalBridgeAuthenticator contains the necessary attributes to authenticate towards a local bridge.
type LocalBridgeAuthenticator struct {
	// ClientKey contains the client key returned by a successful Authenticate API call. This attribute is not required
	// for the actual authentication. After a successful Authenticate API call it is provied for informational purposes.
	ClientKey string
	// UserName contains used to authenticate towards a bridge. This user name must be previously been created via a successful [Authenticate] API call.
	//
	// [Authenticate]: https://developers.meethue.com/develop/hue-api/7-configuration-api/#create-user
	UserName string
	logger   *slog.Logger
}

func (authenticator *LocalBridgeAuthenticator) AuthenticateRequest(ctx context.Context, req *http.Request) error {
	if authenticator.UserName != "" {
		authenticator.logger.Debug("authenticating request", slog.Any("url", req.URL))
		req.Header.Add(hueapi.ApplicationKeyHeader, authenticator.UserName)
	}
	return nil
}

func (authenticator *LocalBridgeAuthenticator) Authenticated(rsp *hueapi.AuthenticateResponse) {
	if rsp.StatusCode() == http.StatusOK {
		rspSuccess := (*rsp.JSON200)[0].Success
		rspError := (*rsp.JSON200)[0].Error
		if rspSuccess != nil {
			authenticator.ClientKey = *rspSuccess.Clientkey
			authenticator.UserName = *rspSuccess.Username
			authenticator.logger.Info("updating authentication", slog.String("client", authenticator.ClientKey))
		}
		if rspError != nil {
			authenticator.logger.Warn("authentication failed", slog.Int("error_type", *rspError.Type), slog.String("error_description", *rspError.Description))
		}
	}
}

func (authenticator *LocalBridgeAuthenticator) Authentication() (string, error) {
	if authenticator.UserName == "" {
		return authenticator.UserName, ErrNotAuthenticated
	}
	return authenticator.UserName, nil
}

func queryAndValidateLocalBridgeConfig(url *url.URL, bridgeId string, timeout time.Duration) (*bridgeConfig, error) {
	httpClient := localBridgeHttpClient(bridgeId, timeout)
	configUrl := configUrl(url)
	config := &bridgeConfig{}
	err := fetchJson(&httpClient.Client, configUrl, config)
	if err != nil {
		return nil, err
	}
	if httpClient.CertificateBridgeId == "" {
		return nil, fmt.Errorf("failed to receive bridge id from '%s'", url)
	}
	if bridgeId != "" && !strings.EqualFold(httpClient.CertificateBridgeId, config.BridgeId) {
		return nil, fmt.Errorf("bridge id mismatch (received '%s' from '%s' and expected '%s')", httpClient.CertificateBridgeId, url, config.BridgeId)
	}
	return config, nil
}

type localBridgeClient struct {
	http.Client
	CertificateBridgeId string
}

func (client *localBridgeClient) verifyBridgeCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(rawCerts) != 1 {
		return fmt.Errorf("unexpected number of bridge certificate: %d", len(rawCerts))
	}
	bridgeCert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("failed to parse bridge certificate (cause: %w)", err)
	}
	bridgeId := bridgeCert.Subject.CommonName
	if client.CertificateBridgeId != "" && !strings.EqualFold(client.CertificateBridgeId, bridgeId) {
		return fmt.Errorf("received bridge id (%s) does not match expected bridge id (%s)", bridgeId, client.CertificateBridgeId)
	}
	client.CertificateBridgeId = bridgeId
	roots := x509.NewCertPool()
	roots.AddCert(hueCACert)
	roots.AddCert(bridgeCert)
	_, err = bridgeCert.Verify(x509.VerifyOptions{Roots: roots, CurrentTime: bridgeCert.NotBefore})
	if err != nil {
		return fmt.Errorf("invalid bridge certificate (cause: %w)", err)
	}
	return nil
}

func localBridgeHttpClient(bridgeId string, timeout time.Duration) *localBridgeClient {
	client := &localBridgeClient{
		Client:              http.Client{Timeout: timeout},
		CertificateBridgeId: strings.ToLower(bridgeId),
	}
	tlsClientconfig := &tls.Config{
		VerifyPeerCertificate: client.verifyBridgeCertificate,
		InsecureSkipVerify:    true,
	}
	client.Transport = &http.Transport{
		ResponseHeaderTimeout: timeout,
		TLSClientConfig:       tlsClientconfig,
	}
	return client
}

func newLocalBridgeHueClient(bridge *Bridge, authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error) {
	httpClient := localBridgeHttpClient(bridge.BridgeId, timeout)
	httpClientOpt := func(c *hueapi.Client) error {
		c.Client = httpClient
		return nil
	}
	apiClient, err := hueapi.NewClientWithResponses(bridge.Url.String(), httpClientOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to create Hue API client (cause: %w)", err)
	}
	return &bridgeClient{
		bridge:        bridge,
		url:           bridge.Url,
		httpClient:    &httpClient.Client,
		apiClient:     apiClient,
		authenticator: authenticator,
	}, nil
}

//go:embed hueCA.pem
var hueCACertRaw []byte
var hueCACert *x509.Certificate

func init() {
	hueCACert = initDecodeCert(hueCACertRaw)
}

func initDecodeCert(data []byte) *x509.Certificate {
	block, rest := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" || len(rest) > 0 {
		log.Fatal("invalid certificate data")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse Hue CA certificate (cause: %w)", err))
	}
	return cert
}
