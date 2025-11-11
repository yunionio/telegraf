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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tdrn-org/go-hue/hueapi"
	"golang.org/x/oauth2"
)

// ErrNotAuthorized indicates the necessary authorization for remote access is either missing or has expired.
var ErrNotAuthorized = errors.New("authorization missing or expired")

// RemoteSession represents the necessary remote authorization required to access a bridge remotely.
type RemoteSession interface {
	// Authorized determines whether a remote authorization has been completed and is still valid (not expired).
	Authorized() bool
	// AuthorizationToken gets the current authorization token. The returned string can be stored to file
	// and used to restore a restoration during a call to [NewRemoteBridgeLocator]. [ErrNotAuthorized] is returned
	// in case there is not valid authorization in place.
	Authorization() (string, error)
	// AuthCodeURL gets the URL to invoke to start the [authorization workflow]. The workflow requires manual
	// interacation (e.g. login into device account and acknowledging device access) and therefore must executed
	// within a browser.
	//
	// [authorization workflow]: https://developers.meethue.com/develop/hue-api/remote-authentication-oauth/
	AuthCodeURL() string
	setAuthHeader(req *http.Request) error
	authHttpClient(timeout time.Duration) *http.Client
	handleOauth2Authorized(w http.ResponseWriter, req *http.Request, code string)
}

var remoteDefaultEndpointUrl *url.URL = safeParseUrl("https://api.meethue.com/")

// NewRemoteBridgeLocator creates a new [RemoteBridgeLocator] for discovering a remote bridge via the Hue [Cloud API].
//
// The given client id and secret are obtaining a Hue developer account and registering a [Remote Hue API app].
// The redirect URL must match the callback URL registered during app creation. The [RemoteSession] associated with
// the newly created locator is listening on this URL to receive the authorization credentials.
//
// If redirect URL is nil, a localhost based URL is created dynamically. Such dynamic redirect URLs are suitable
// for local testing only.
//
// If tokenFile is not empty, it must point to a file for storing authorization credentials. Such credentials
// are automatically restored during next start avoiding the need for a new interactive authorization workflow (unless
// the stored credentials have expired since last use).
//
// Example:
//
//	clientId := "..." // as defined in app registration
//	clientSecret := "..." // as defined in app registration
//	redirectUrl, _ := url.Parse("http://127.0.0.1:9123/authorized") // as defined in app registration
//	locator, _ := hue.NewRemoteBridgeLocator(clientId, clientSecret, redirectUrl, "token.json")
//	fmt.Println("Start authorization workflow in browser via: ", locator.AuthCodeURL())
//	// After a successful authorization workflow token.json contains a valid token for remote access
//	bridges, _ := locator.Locate(hue.DefaultTimeout)
//	for _, bridge := range bridges {
//		authenticator := new hue.NewremoteAuthenticator(locator, "")
//		client, _ := bridge.NewClient(authenticator, hue.DefaultTimeout)
//		authenticator.EnableLinking()
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
//	// If Bridge Id and Username are already known and a valid token file is in place, this can be shortened to
//
//	locator, _ := hue.NewRemoteBridgeLocator(clientId, clientSecret, redirectUrl, "token.json")
//	bridge, _ := locator.Lookup("0123456789ABCDEF", hue.DefaultTimeout)
//	client, _ := bridge.NewClient(hue.NewRemoteAuthenticator(locator, "secret username token"), hue.DefaultTimeout)
//	getDevicesResponse, _ := client.GetDevices()
//
// [Cloud API]: https://developers.meethue.com/develop/hue-api/remote-authentication-oauth/
// [Remote Hue API app]: https://developers.meethue.com/my-apps/
func NewRemoteBridgeLocator(clientId string, clientSecret string, redirectUrl *url.URL, tokenFile string) (*RemoteBridgeLocator, error) {
	tokenSource, err := loadRemoteTokenSource(tokenFile)
	if err != nil {
		return nil, err
	}
	logger := slog.With(slog.String("locator", remoteBridgeLocatorName))
	locator := &RemoteBridgeLocator{
		EndpointUrl:       remoteDefaultEndpointUrl,
		ClientId:          clientId,
		ClientSecret:      clientSecret,
		oauth2TokenSource: tokenSource,
		logger:            logger,
	}
	callback, err := remoteOauth2.listen(redirectUrl)
	if err != nil {
		return nil, err
	}
	locator.oauth2Callback = callback
	return locator, nil
}

const remoteBridgeLocatorName string = "remote"

// RemoteBridgeLocator locates a remote bridge via the Hue [Cloud API].
//
// Use [NewRemoteBridgeLocator] to create a new instance.
//
// [Cloud API]: https://developers.meethue.com/develop/hue-api/remote-authentication-oauth/
type RemoteBridgeLocator struct {
	// EndpointUrl defines the [Cloud API] endpoint to use. This URL defaults to https://api.meethue.com and may be
	// overwritten for local testing.
	//
	// [Cloud API]: https://developers.meethue.com/develop/hue-api/remote-authentication-oauth/
	EndpointUrl *url.URL
	// TlsConfig defines the TLS configuration to use for accessing the endpoint URL. If nil, the standard options are used.
	TlsConfig *tls.Config
	// ClientId defines the client id of the [Remote Hue API app] to use for remote access.
	//
	// [Remote Hue API app]: https://developers.meethue.com/my-apps/
	ClientId string
	// ClientSecret defines the client secret of the [Remote Hue API app] to use for remote access.
	//
	// [Remote Hue API app]: https://developers.meethue.com/my-apps/
	ClientSecret string
	// ReferrerUrl defines the URL to redirect to after an authorization workflow has been completed. The default value nil
	// disables the redirect.
	ReferrerUrl         *url.URL
	oauth2Callback      *remoteOauth2Callback
	oauth2TokenSource   *cachedTokenSource
	cachedOauthConfig   *oauth2.Config
	cachedOauth2Context context.Context
	logger              *slog.Logger
}

func (locator *RemoteBridgeLocator) Name() string {
	return remoteBridgeLocatorName
}

func (locator *RemoteBridgeLocator) Query(timeout time.Duration) ([]*Bridge, error) {
	bridge, err := locator.Lookup("", timeout)
	if err != nil {
		return []*Bridge{}, nil
	}
	return []*Bridge{bridge}, nil
}

func (locator *RemoteBridgeLocator) Lookup(bridgeId string, timeout time.Duration) (*Bridge, error) {
	client := locator.authHttpClient(timeout)
	url := locator.EndpointUrl.JoinPath("/route")
	locator.logger.Info("probing remote endpoint...", slog.Any("url", url))
	configUrl := configUrl(url)
	config := &bridgeConfig{}
	err := fetchJson(client, configUrl, config)
	if err != nil {
		return nil, err
	}
	if bridgeId != "" && !strings.EqualFold(bridgeId, config.BridgeId) {
		return nil, fmt.Errorf("bridge id mismatch (received '%s' and expected '%s')", bridgeId, config.BridgeId)
	}
	bridge, err := config.newBridge(locator, url)
	if err != nil {
		return nil, err
	}
	locator.logger.Info("located bridge", slog.Any("bridge", bridge))
	return bridge, nil
}

func (locator *RemoteBridgeLocator) NewClient(bridge *Bridge, authenticator BridgeAuthenticator, timeout time.Duration) (BridgeClient, error) {
	httpClient := locator.authHttpClient(timeout)
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
		httpClient:    httpClient,
		apiClient:     apiClient,
		authenticator: authenticator,
	}, nil
}

func (locator *RemoteBridgeLocator) Authorized() bool {
	token, _ := locator.oauth2TokenSource.Token()
	return token.Valid()
}

func (locator *RemoteBridgeLocator) Authorization() (string, error) {
	token, _ := locator.oauth2TokenSource.Token()
	if !token.Valid() {
		return "", ErrNotAuthorized
	}
	tokenBytes, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal authorization token (cause: %w)", err)
	}
	return string(tokenBytes), nil
}

func (locator *RemoteBridgeLocator) AuthCodeURL() string {
	oauth2Config := locator.oauth2Config()
	state := remoteOauth2.authCodeState(locator)
	return oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (locator *RemoteBridgeLocator) setAuthHeader(req *http.Request) error {
	token, err := locator.oauth2TokenSource.Token()
	if err != nil {
		return errors.Join(ErrNotAuthorized, err)
	}
	if !token.Valid() {
		return ErrNotAuthorized
	}
	token.SetAuthHeader(req)
	return nil
}

func (locator *RemoteBridgeLocator) authHttpClient(timeout time.Duration) *http.Client {
	var transport http.RoundTripper
	transport = &http.Transport{
		ResponseHeaderTimeout: timeout,
		TLSClientConfig:       locator.TlsConfig.Clone(),
	}
	token, _ := locator.oauth2TokenSource.Token()
	if token.Valid() {
		config := locator.oauth2Config()
		ctx := locator.oauth2Context()
		locator.oauth2TokenSource.Reset(config.TokenSource(ctx, token))
		transport = &oauth2.Transport{
			Source: locator.oauth2TokenSource,
			Base:   transport,
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout}
}

func (locator *RemoteBridgeLocator) handleOauth2Authorized(w http.ResponseWriter, req *http.Request, code string) {
	if code != "" {
		config := locator.oauth2Config()
		ctx := locator.oauth2Context()
		token, err := config.Exchange(ctx, code)
		if err != nil {
			locator.logger.Error("failed to retrieve token", slog.Any("err", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		locator.oauth2TokenSource.Reset(config.TokenSource(ctx, token))
	}
	if locator.ReferrerUrl != nil {
		http.Redirect(w, req, locator.ReferrerUrl.String(), http.StatusSeeOther)
	}
}

func (locator *RemoteBridgeLocator) oauth2Config() *oauth2.Config {
	if locator.cachedOauthConfig == nil {
		locator.cachedOauthConfig = &oauth2.Config{
			ClientID:     locator.ClientId,
			ClientSecret: locator.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  locator.EndpointUrl.JoinPath("/v2/oauth2/authorize").String(),
				TokenURL: locator.EndpointUrl.JoinPath("/v2/oauth2/token").String(),
			},
			RedirectURL: locator.oauth2Callback.redirectUrl.String(),
			Scopes:      []string{},
		}
	}
	return locator.cachedOauthConfig
}

func (locator *RemoteBridgeLocator) oauth2Context() context.Context {
	if locator.cachedOauth2Context == nil {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: locator.TlsConfig.Clone(),
			},
		}
		locator.cachedOauth2Context = context.WithValue(context.Background(), oauth2.HTTPClient, client)
	}
	return locator.cachedOauth2Context
}

// NewRemoteBridgeAuthenticator creates a new [RemoteBridgeAuthenticator] suitable for authenticating towards a remote bridge.
//
// The user name must be previously been created via a successful [Authenticate] API call. In difference to a local [Authenticate]
// calls, where the bridge's link button is pressed physically to acknowledge acccess, the remote variant is acknowledged by invoking
// [RemoteBridgeAuthenticator.EnableLinking] prior to the [Authenticate] API call.
//
// The given [RemoteSession] argument represents the authorization to use for accessing the [Cloud API]. The [RemoteBridgeLocator]
// used to locate the remote bridge and authorize the remote access provides this [RemoteSession].
//
// The user name must be empty or previously been created via a successful [Authenticate] API call. Everytime a
// successful [Authenticate] API call is performed, the user name will be overwritten by the returned user name.
//
// [Authenticate]: https://developers.meethue.com/develop/hue-api/7-configuration-api/#create-user
// [Cloud API]: https://developers.meethue.com/develop/hue-api/remote-authentication-oauth/
func NewRemoteBridgeAuthenticator(remoteSession RemoteSession, userName string) *RemoteBridgeAuthenticator {
	logger := slog.With(slog.String("authenticator", "remote"))
	return &RemoteBridgeAuthenticator{
		remoteSession: remoteSession,
		UserName:      userName,
		logger:        logger,
	}
}

// RemoteBridgeAuthenticator is used to authenticate towards a remote bridge.
type RemoteBridgeAuthenticator struct {
	remoteSession RemoteSession
	ClientKey     string
	UserName      string
	logger        *slog.Logger
}

func (authenticator *RemoteBridgeAuthenticator) Authorization() (string, error) {
	return authenticator.remoteSession.Authorization()
}

func (authenticator *RemoteBridgeAuthenticator) AuthenticateRequest(ctx context.Context, req *http.Request) error {
	err := authenticator.remoteSession.setAuthHeader(req)
	if err == nil {
		authenticator.logger.Debug("authorizing remote request", slog.Any("url", req.URL))
		if authenticator.UserName != "" {
			authenticator.logger.Debug("authenticating request", slog.Any("url", req.URL))
			req.Header.Add(hueapi.ApplicationKeyHeader, authenticator.UserName)
		}
	}
	return nil
}

func (authenticator *RemoteBridgeAuthenticator) Authenticated(rsp *hueapi.AuthenticateResponse) {
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

func (authenticator *RemoteBridgeAuthenticator) Authentication() (string, error) {
	if authenticator.UserName == "" {
		return authenticator.UserName, ErrNotAuthenticated
	}
	return authenticator.UserName, nil
}

// EnableLinking must be called prior to a [Authenticate] API call to acknoledge the user registration.
//
// [Authenticate]: https://developers.meethue.com/develop/hue-api/7-configuration-api/#create-user
func (authenticator *RemoteBridgeAuthenticator) EnableLinking(bridge *Bridge) error {
	configUrl := configUrl(bridge.Url)
	body := bytes.NewBuffer([]byte(`{"linkbutton":true}`))
	req, err := http.NewRequest(http.MethodPut, configUrl.String(), body)
	if err != nil {
		return fmt.Errorf("failed to prepare linking request (cause: %w)", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := authenticator.remoteSession.authHttpClient(DefaultTimeout)
	rsp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send enable linking request (cause: %w)", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to enable linking (status: %s)", rsp.Status)
	}
	return nil
}

var remoteOauth2 = &remoteOauth2Callbacks{
	entries: make(map[string]*remoteOauth2Callback),
	states:  make(map[string]*remoteOauth2State),
}

type remoteOauth2Callbacks struct {
	mutex   sync.Mutex
	entries map[string]*remoteOauth2Callback
	states  map[string]*remoteOauth2State
}

type remoteOauth2Callback struct {
	redirectUrl *url.URL
	listener    net.Listener
	httpServer  *http.Server
}

type remoteOauth2State struct {
	session RemoteSession
	expiry  time.Time
}

func (callbacks *remoteOauth2Callbacks) logger(redirectUrl *url.URL) *slog.Logger {
	logger := slog.With(slog.Any("oauth2-callback", redirectUrl))
	return logger
}

func (callbacks *remoteOauth2Callbacks) listen(redirectUrl *url.URL) (*remoteOauth2Callback, error) {
	callbacks.mutex.Lock()
	defer callbacks.mutex.Unlock()
	var callbackKey string
	var callback *remoteOauth2Callback
	if redirectUrl != nil {
		callbackKey = redirectUrl.String()
		callback = callbacks.entries[callbackKey]
	}
	if callback == nil {
		listenAndRedirectUrl, listener, err := callbacks.listen0(redirectUrl)
		if err != nil {
			return nil, err
		}
		handler := http.NewServeMux()
		handler.HandleFunc("GET "+listenAndRedirectUrl.Path, callbacks.handleOauth2Authorized)
		callbackKey = listenAndRedirectUrl.String()
		callback = &remoteOauth2Callback{
			redirectUrl: listenAndRedirectUrl,
			listener:    listener,
			httpServer:  &http.Server{Handler: handler},
		}
		go func() {
			logger := callbacks.logger(listenAndRedirectUrl)
			logger.Info("http server starting...")
			err := callback.httpServer.Serve(callback.listener)
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("http server failure", slog.Any("err", err))
			}
		}()
		callbacks.entries[callbackKey] = callback
	}
	return callback, nil
}

func (callbacks *remoteOauth2Callbacks) listen0(redirectUrl *url.URL) (*url.URL, net.Listener, error) {
	var address string
	var path string
	if redirectUrl != nil {
		scheme := redirectUrl.Scheme
		if scheme != "http" {
			return nil, nil, fmt.Errorf("unsupported redirect URL scheme '%s'", scheme)
		}
		hostname := redirectUrl.Hostname()
		port := redirectUrl.Port()
		if port == "" {
			port = "80"
		}
		address = net.JoinHostPort(hostname, port)
		path = redirectUrl.Path
	} else {
		address = "localhost:0"
		path = "/authorized"
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s (cause: %w)", address, err)
	}
	rawListenUrl := "http://" + listener.Addr().String() + "/"
	listenUrl, err := url.Parse(rawListenUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse listen URL '%s' (cause: %w)", rawListenUrl, err)
	}
	listenAndRedirectUrl := listenUrl.JoinPath(path)
	return listenAndRedirectUrl, listener, nil
}

func (callbacks *remoteOauth2Callbacks) authCodeState(session RemoteSession) string {
	callbacks.mutex.Lock()
	defer callbacks.mutex.Unlock()
	state := uuid.New().String()
	expiry := time.Now().Add(60 * time.Second)
	callbacks.states[state] = &remoteOauth2State{
		session: session,
		expiry:  expiry,
	}
	return state
}

func (callbacks *remoteOauth2Callbacks) stateSession(state string) RemoteSession {
	callbacks.mutex.Lock()
	defer callbacks.mutex.Unlock()
	var session RemoteSession
	now := time.Now()
	for state0, resolvedState0 := range callbacks.states {
		if resolvedState0.expiry.Before(now) {
			delete(callbacks.states, state0)
		} else if state0 == state {
			session = resolvedState0.session
			delete(callbacks.states, state0)
		}
	}
	return session
}

func (callbacks *remoteOauth2Callbacks) handleOauth2Authorized(w http.ResponseWriter, req *http.Request) {
	logger := callbacks.logger(req.URL)
	reqParams, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		logger.Error("failed to decode callback request parameters", slog.String("query", req.URL.RawQuery), slog.Any("err", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	code := reqParams.Get("code")
	state := reqParams.Get("state")
	session := callbacks.stateSession(state)
	if session == nil {
		logger.Error("authorization workflow failed", slog.String("query", req.URL.RawQuery))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	session.handleOauth2Authorized(w, req, code)
}

func loadRemoteTokenSource(tokenFile string) (*cachedTokenSource, error) {
	validatedTokenFile := ""
	if tokenFile != "" {
		absoluteTokenFile, err := filepath.Abs(tokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve token file '%s' (cause: %w)", tokenFile, err)
		}
		absoluteTokenDir := filepath.Dir(absoluteTokenFile)
		err = os.MkdirAll(absoluteTokenDir, os.ModeDir|0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create token directory '%s' (cause: %w)", absoluteTokenDir, err)
		}
		validatedTokenFile = absoluteTokenFile
	}
	logger := slog.With(slog.String("token", validatedTokenFile))
	var cachedToken *oauth2.Token
	var liveSource oauth2.TokenSource
	if validatedTokenFile != "" {
		logger.Info("using token file")
		tokenBytes, err := os.ReadFile(validatedTokenFile)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read token file '%s' (cause: %w)", validatedTokenFile, err)
		}
		if err == nil {
			logger.Info("reading token file...")
			cachedToken = &oauth2.Token{}
			err = json.Unmarshal(tokenBytes, cachedToken)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal token file '%s' (cause: %w)", validatedTokenFile, err)
			}
			liveSource = oauth2.StaticTokenSource(cachedToken)
		} else {
			logger.Info("token file not yet available")
		}
	}
	tokenSource := &cachedTokenSource{
		tokenFile:   validatedTokenFile,
		cachedToken: cachedToken,
		liveSource:  liveSource,
		logger:      logger,
	}
	return tokenSource, nil
}

type cachedTokenSource struct {
	tokenFile   string
	cachedToken *oauth2.Token
	liveSource  oauth2.TokenSource
	logger      *slog.Logger
}

func (tokenSource *cachedTokenSource) Token() (*oauth2.Token, error) {
	if tokenSource.liveSource == nil {
		return nil, ErrNotAuthorized
	}
	token, err := tokenSource.liveSource.Token()
	if err != nil {
		return token, err
	}
	if tokenSource.cachedToken == nil || tokenSource.cachedToken.AccessToken != token.AccessToken || tokenSource.cachedToken.TokenType != token.TokenType || tokenSource.cachedToken.RefreshToken != token.RefreshToken {
		tokenSource.cachedToken = token
		if tokenSource.tokenFile != "" {
			tokenSource.logger.Info("updating token file...", slog.String("file", tokenSource.tokenFile))
			tokenFileDir := filepath.Dir(tokenSource.tokenFile)
			err := os.MkdirAll(tokenFileDir, 0700)
			if err != nil {
				tokenSource.logger.Error("failed to create token directory", slog.String("dir", tokenFileDir), slog.Any("err", err))
				return token, nil
			}
			tokenBytes, err := json.Marshal(token)
			if err != nil {
				tokenSource.logger.Error("failed to marshal token", slog.Any("err", err))
				return token, nil
			}
			err = os.WriteFile(tokenSource.tokenFile, tokenBytes, 0600)
			if err != nil {
				tokenSource.logger.Error("failed to write token file", slog.String("file", tokenSource.tokenFile), slog.Any("err", err))
				return token, nil
			}
		}
	}
	return token, nil
}

func (tokenSource *cachedTokenSource) Reset(liveSource oauth2.TokenSource) {
	tokenSource.liveSource = liveSource
}
