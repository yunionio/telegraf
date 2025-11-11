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

package tr064

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

// ServiceDescriptor represents a concrete service provided by a TR-064 server.
//
// The available services of a TR-064 server are determined by calling [Client.Services].
// Via [Client.InvokeService] a identified service can be invoked. The latter is normally
// not used directly. Instead the generated ServiceClient for the service is instantiated
// and invoked. See [Client] for further details.
type ServiceDescriptor interface {
	// Spec returns the TR-064 specification describing this service.
	Spec() ServiceSpec
	// Type returns the full service type as defined in the specification document.
	Type() string
	// ShortType returns the short type name of this service.
	ShortType() string
	// Id returns the full service id as defined in the specification document.
	Id() string
	// ShortId returns the short id of this service.
	ShortId() string
	// ControlUrl returns the control URL to use for accessing this service.
	ControlUrl() string
}

// StaticServiceDescriptor represents a statically defined [ServiceDescriptor].
//
// Normally a service should be identified dynamically by calling [Client.ServicesByType].
// In case the presence of a service is well-known, this static descriptor can be used.
type StaticServiceDescriptor struct {
	// ServiceSpec receives the TR-064 specification this service is normally defined in.
	ServiceSpec ServiceSpec
	// ServiceType receives the full service type of the service.
	ServiceType string
	// ServiceId receives the full service id of the service.
	ServiceId string
	// ServiceControlUrl receives the URL to use for accessing the service.
	ServiceControlUrl string
}

// Spec returns the TR-064 specification describing this service.
func (service *StaticServiceDescriptor) Spec() ServiceSpec {
	return service.ServiceSpec
}

// Type returns the full service type as defined in the specification document.
func (service *StaticServiceDescriptor) Type() string {
	return service.ServiceType
}

// ShortType returns the short type name of this service.
func (service *StaticServiceDescriptor) ShortType() string {
	return serviceShortType(service.ServiceType)
}

// Id returns the full service id as defined in the specification document.
func (service *StaticServiceDescriptor) Id() string {
	return service.ServiceId
}

// ShortId returns the short id of this service.
func (service *StaticServiceDescriptor) ShortId() string {
	return serviceShortId(service.ServiceId)
}

// ControlUrl returns the control URL to use for accessing this service.
func (service *StaticServiceDescriptor) ControlUrl() string {
	return service.ServiceControlUrl
}

// NewClient instantiates a new TR-064 client for accessing the given URL.
//
// If the given URL contains a user info, the contained username and password are automatically used for authentication.
func NewClient(deviceUrl *url.URL) *Client {
	anonymousDeviceUrl := *deviceUrl
	anonymousDeviceUrl.User = nil
	username := deviceUrl.User.Username()
	password, _ := deviceUrl.User.Password()
	client := &Client{
		DeviceUrl:             &anonymousDeviceUrl,
		Username:              username,
		Password:              password,
		mutex:                 &sync.Mutex{},
		cachedServices:        make(map[ServiceSpec][]ServiceDescriptor),
		cachedAuthentications: make(map[string]string),
	}
	client.cachedHttpClient = sync.OnceValue(client.httpClient)
	return client
}

// Client provides the necessary parameters to access a TR-064 capable server and perform service discovery.
//
// To access an actual service, this client as well as the desired service descriptor is combined into a
// service specific service client:
//
//	client := tr064.NewClient(deviceUrl)
//	services, _ := client.ServicesByName(tr064.DefaultServiceSpec, deviceinfo.ServiceName)
//	serviceClient := deviceinfo.ServiceClient {
//		TR064Client: client,
//		Service:     services[0],
//	}
//	info := &deviceinfo.GetInfoResponse{}
//	_ = serviceClient.GetInfo(info)
//
// The service client is then used to access the individual service functions.
type Client struct {
	// Url defines the URL to access the TR-064 server.
	DeviceUrl *url.URL
	// Username is set to the login to use for accessing restricted services.
	Username string
	// Password is set to the password to use for accessing restricted services.
	Password string
	// Timeout sets the timeout for HTTP(S) communication.
	Timeout time.Duration
	// TlsConfig defines the TLS options to use for HTTPS communication. May be nil.
	TlsConfig *tls.Config
	// Debug enables debug logging while accessing the TR-064 server.
	Debug                 bool
	mutex                 *sync.Mutex
	cachedServices        map[ServiceSpec][]ServiceDescriptor
	cachedAuthentications map[string]string
	cachedHttpClient      func() *http.Client
}

// Services fetches and parses the given specification and returns the defined services.
func (client *Client) Services(spec ServiceSpec) ([]ServiceDescriptor, error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	services := client.cachedServices[spec]
	if services == nil {
		services = make([]ServiceDescriptor, 0)
		httpClient := client.cachedHttpClient()
		tr64desc, err := fetchServiceSpec(httpClient, client.DeviceUrl, spec)
		if err != nil {
			return nil, err
		}
		collector := &serviceCollector{serviceMap: make(map[string]ServiceDescriptor)}
		err = tr64desc.walk(httpClient, client.DeviceUrl, collector.collectService)
		if err != nil {
			return nil, err
		}
		services = append(services, slices.Collect(maps.Values(collector.serviceMap))...)
		slices.SortFunc(services, func(a ServiceDescriptor, b ServiceDescriptor) int { return strings.Compare(a.Type(), b.Type()) })
		client.cachedServices[spec] = services
	}
	return services, nil
}

type serviceCollector struct {
	serviceMap map[string]ServiceDescriptor
}

func (collector *serviceCollector) collectService(service *serviceDoc, scpd *scpdDoc) error {
	collector.serviceMap[service.ServiceId] = service
	return nil
}

// ServicesByType fetches and parses the TR-064 server's service specifications
// like [Services], but returns only the services matching service type.
func (client *Client) ServicesByType(spec ServiceSpec, serviceType string) ([]ServiceDescriptor, error) {
	all, err := client.Services(spec)
	if err != nil {
		return nil, err
	}
	services := make([]ServiceDescriptor, 0)
	for _, service := range all {
		if service.Type() == serviceType || service.ShortType() == serviceType {
			services = append(services, service)
		}
	}
	return services, nil
}

// Get performs a simple GET request towards the TR-064 server using the given path reference.
func (client *Client) Get(ref string) (*http.Response, error) {
	refUrl, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: '%s' (cause: %w)", ref, err)
	}
	targetUrl := client.DeviceUrl.ResolveReference(refUrl)
	return client.cachedHttpClient().Get(targetUrl.String())
}

// NewSOAPRequest constructs a new SOAP request object wrapping the given input argument.
// The constructed SOAP request is suitable for invoking [Client.InvokeService]
func NewSOAPRequest[T any](in *T) *SOAPRequest[T] {
	return &SOAPRequest[T]{
		XMLNameSpace:     XMLNameSpace,
		XMLEncodingStyle: XMLEncodingStyle,
		Body: &SOAPRequestBody[T]{
			In: in,
		},
	}
}

// SOAPRequest defines XML based SOAP request object.
type SOAPRequest[T any] struct {
	XMLName          xml.Name `xml:"s:Envelope"`
	XMLNameSpace     string   `xml:"xmlns:s,attr"`
	XMLEncodingStyle string   `xml:"s:encodingStyle,attr"`
	Body             *SOAPRequestBody[T]
}

// SOAPRequestBody defines the Body element for a SOAP request.
type SOAPRequestBody[T any] struct {
	XMLName xml.Name `xml:"s:Body"`
	In      *T
}

// NewSOAPResponse constructs a new SOAP response object wrapping the given output argument.
// The constructed SOAP response is suitable for invoking [Client.InvokeService]
func NewSOAPResponse[T any](out *T) *SOAPResponse[T] {
	return &SOAPResponse[T]{
		Body: &SOAPResponseBody[T]{
			Out: out,
		},
	}
}

// SOAPResponse defines XML based SOAP response object.
type SOAPResponse[T any] struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    *SOAPResponseBody[T]
}

// SOAPResponseBody defines the Body element for a SOAP response.
type SOAPResponseBody[T any] struct {
	XMLName xml.Name `xml:"Body"`
	Out     *T
}

// InvokeService invokes the SOAP service identifed via the given service descriptor using the given input and output objects.
//
// If needed, the function performs the required authentication using the client's username and password attributes.
func (client *Client) InvokeService(service ServiceDescriptor, actionName, in any, out any) error {
	controlUrl, err := url.Parse(service.ControlUrl())
	if err != nil {
		return fmt.Errorf("failed to parse control URL '%s' (cause: %w)", service.ControlUrl(), err)
	}
	endpoint := client.DeviceUrl.ResolveReference(controlUrl).String()
	soapAction := fmt.Sprintf("%s#%s", service.Type(), actionName)
	requestBody, err := xml.MarshalIndent(in, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal request (cause: %w)", err)
	}
	if client.Debug {
		log.Println("Request:\n", string(requestBody))
	}
	authentication := client.authentication(service.Type())
	response, err := client.postSoapActionRequest(endpoint, soapAction, requestBody, authentication)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusUnauthorized {
		authentication, err := client.authenticate(response, service.Type())
		if err == nil {
			response, err = client.postSoapActionRequest(endpoint, soapAction, requestBody, authentication)
			if err != nil {
				return err
			}
		}
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("service call failure (status: %s)", response.Status)
	}
	responseBody, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response body (cause: %w)", err)
	}
	if client.Debug {
		log.Println("Response:\n", string(responseBody))
	}
	err = xml.Unmarshal(responseBody, out)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body (cause: %w)", err)
	}
	return nil
}

func (client *Client) authentication(serviceType string) string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.cachedAuthentications[serviceType]
}

func (client *Client) authenticate(challenge *http.Response, serviceType string) (string, error) {
	challengeHeader := challenge.Header["Www-Authenticate"]
	if len(challengeHeader) != 1 {
		return "", fmt.Errorf("missing or unexpected WWW-Authenticate header")
	}
	challengeValues := make(map[string]string)
	for _, challengeHeaderValue := range strings.Split(challengeHeader[0], ",") {
		splitChallengeHeaderValue := strings.Split(challengeHeaderValue, "=")
		if len(splitChallengeHeaderValue) == 2 {
			key := splitChallengeHeaderValue[0]
			value := splitChallengeHeaderValue[1]
			if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				value = value[1 : len(value)-1]
			}
			challengeValues[key] = value
		}
	}
	digestRealm := challengeValues["Digest realm"]
	ha1 := client.md5Hash(fmt.Sprintf("%s:%s:%s", client.Username, digestRealm, client.Password))
	ha2 := client.md5Hash(fmt.Sprintf("%s:%s", http.MethodPost, serviceType))
	nonce := challengeValues["nonce"]
	qop := challengeValues["qop"]
	cnonce := client.newCNonce()
	nc := "1"
	response := client.md5Hash(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, nonce, nc, cnonce, qop, ha2))
	authentication := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		client.Username, digestRealm, nonce, serviceType, cnonce, nc, qop, response)
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.cachedAuthentications == nil {
		client.cachedAuthentications = make(map[string]string)
	}
	client.cachedAuthentications[serviceType] = authentication
	return authentication, nil
}

func (client *Client) md5Hash(s string) string {
	hash := md5.New()
	_, err := hash.Write([]byte(s))
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (client *Client) newCNonce() string {
	cnonceBytes := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, cnonceBytes)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%016x", cnonceBytes)
}
func (client *Client) postSoapActionRequest(endpoint string, action string, requestBody []byte, authentication string) (*http.Response, error) {
	if client.Debug {
		log.Printf("Invoking action %s on endpoint %s ...\n", action, endpoint)
	}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request (cause: %w)", err)
	}
	request.Header.Add("Content-Type", "text/xml")
	request.Header.Add("SoapAction", action)
	if authentication != "" {
		request.Header.Add("Authorization", authentication)
	}
	response, err := client.cachedHttpClient().Do(request)
	if err != nil {
		return response, fmt.Errorf("failed to post request (cause: %w)", err)
	}
	if client.Debug {
		log.Println("Status: ", response.Status)
	}
	return response, nil
}

func (client *Client) httpClient() *http.Client {
	tlsClientConfig := client.TlsConfig.Clone()
	transport := &http.Transport{
		TLSClientConfig: tlsClientConfig,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   client.Timeout,
	}
}
