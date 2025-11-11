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

package mock

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
)

// Mock server user name
const User = "user"

// Mock server password
const Password = "password"

// ServiceMock is used to plugin service mocks.
type ServiceMock struct {
	// Path sets the URL path used to invoke this service.
	Path string
	// HandleFunc provides the actual mock functionality (which is called as
	// soon the URL path defined is accessed).
	HandleFunc func(http.ResponseWriter, *http.Request)
}

// ServiceMockFromFile creates a [ServiceMock] instance sending the given file
// everytime the associated path is accessed.
func ServiceMockFromFile(path string, file string) *ServiceMock {
	return &ServiceMock{
		Path: path,
		HandleFunc: func(w http.ResponseWriter, _ *http.Request) {
			responseBytes, err := os.ReadFile(file)
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/xml")
			_, err = w.Write(responseBytes)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		},
	}
}

// TR064Server interface is used to interact with a mock server instantiated via [Start].
type TR064Server interface {
	// Server gets the HTTP URL the mock server is listenting on.
	Server() *url.URL
	// Server gets the HTTPS URL the mock server is listenting on.
	SecureServer() *url.URL
	// Ping checks whether the mock server is up and running (on the HTTP address).
	Ping() error
	// Ping checks whether the mock server is up and running (on the HTTPS address).
	SecurePing() error
	// Shutdown terminates the mock server gracefully.
	Shutdown()
}

// Start setup and starts a new mock server.
//
// The mock server establishes a HTTP as well as a HTTPS listener using dynamic ports.
// Use [Server] and [SecureServer] to get the actual addresses.
func Start(docsDir string, mocks ...*ServiceMock) TR064Server {
	httpListener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	httpServerUrl, err := url.Parse("http://" + User + ":" + Password + "@" + httpListener.Addr().String() + "/")
	if err != nil {
		log.Fatal(err)
	}
	httpsListener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	httpsServerUrl, err := url.Parse("https://" + User + ":" + Password + "@" + httpsListener.Addr().String() + "/")
	if err != nil {
		log.Fatal(err)
	}
	mock := &mockServer{
		docsDir:        docsDir,
		httpListener:   httpListener,
		httpServerUrl:  httpServerUrl,
		httpsListener:  httpsListener,
		httpsServerUrl: httpsServerUrl,
	}
	mock.setupAndStartServer(mocks...)
	return mock
}

type mockServer struct {
	docsDir        string
	httpListener   net.Listener
	httpServerUrl  *url.URL
	httpServer     *http.Server
	httpsListener  net.Listener
	httpsServerUrl *url.URL
	httpsServer    *http.Server
	stoppedWG      sync.WaitGroup
}

// Server gets the HTTP URL the mock server is listenting on.
func (mock *mockServer) Server() *url.URL {
	return mock.httpServerUrl
}

// Server gets the HTTPS URL the mock server is listenting on.
func (mock *mockServer) SecureServer() *url.URL {
	return mock.httpsServerUrl
}

// Ping checks whether the mock server is up and running (on the HTTP address).
func (mock *mockServer) Ping() error {
	return mock.ping(http.DefaultClient, mock.httpServerUrl)
}

func (mock *mockServer) ping(client *http.Client, url *url.URL) error {
	pingUrl := url.JoinPath("/ping")
	log.Println("Pinging '", pingUrl, "'...")
	response, err := client.Get(pingUrl.String())
	if err != nil {
		return fmt.Errorf("failed to access URL '%s' (cause: %w)", url, err)
	}
	log.Println("Ping status: ", response.Status)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to ping URL '%s' (status: %s)", url, response.Status)
	}
	return nil
}

// Ping checks whether the mock server is up and running (on the HTTPS address).
func (mock *mockServer) SecurePing() error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	return mock.ping(client, mock.httpsServerUrl)
}

// Shutdown terminates the mock server gracefully.
func (mock *mockServer) Shutdown() {
	log.Println("Shutting down mock server...")
	err := mock.httpServer.Shutdown(context.Background())
	if err != nil {
		log.Println("Failed to shutdown HTTPS server: ", err)
	}
	err = mock.httpsServer.Shutdown(context.Background())
	if err != nil {
		log.Println("Failed to shutdown HTTP server: ", err)
	}
	mock.stoppedWG.Wait()
	log.Println("Mock server stopped")
}

func (mock *mockServer) handlePing(w http.ResponseWriter, req *http.Request) {
	log.Println("Mock: ", req.URL)
	w.WriteHeader(http.StatusOK)
}

type soapRequest struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"Body"`
		Action  string   `xml:",innerxml"`
	}
}

var soapActionPattern = regexp.MustCompile(`^\s*<u:([0-9a-zA-z_-]+)Request .*`)

// UnmarshalSoapAction unmarshals a SOAP body from the given HTTP request and determines the contained action.
//
// In case of an error, the corresponding response status code is automatically written.
func UnmarshalSoapAction(w http.ResponseWriter, req *http.Request) (string, error) {
	requestBody, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", err
	}
	soapRequest := &soapRequest{}
	err = xml.Unmarshal(requestBody, soapRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return "", err
	}
	match := soapActionPattern.FindStringSubmatch(soapRequest.Body.Action)
	if match == nil {
		w.WriteHeader(http.StatusBadRequest)
		return "", fmt.Errorf("unexpected soap request body")
	}
	return match[1], nil
}

// WriteSoapResponse writes a SOAP response by wrapping the given output object into the necessary SOAP envelope.
//
// In case of an error, the corresponding response status code is automatically written.
func WriteSoapResponse(w http.ResponseWriter, out any) error {
	response := &soapResponse{
		XMLNameSpace:     "http://schemas.xmlsoap.org/soap/envelope/",
		XMLEncodingStyle: "http://schemas.xmlsoap.org/soap/encoding/",
		Body: &soapResponseBody{
			Out: out,
		},
	}
	responseBody, err := xml.MarshalIndent(response, "", "\t")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "text/xml")
	_, err = w.Write(responseBody)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

type soapResponse struct {
	XMLName          xml.Name `xml:"s:Envelope"`
	XMLNameSpace     string   `xml:"xmlns:s,attr"`
	XMLEncodingStyle string   `xml:"s:encodingStyle,attr"`
	Body             *soapResponseBody
}

type soapResponseBody struct {
	XMLName xml.Name `xml:"s:Body"`
	Out     any
}

func (mock *mockServer) setupAndStartServer(mocks ...*ServiceMock) {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /ping", mock.handlePing)
	for _, mock := range mocks {
		handler.HandleFunc("POST "+mock.Path, mock.HandleFunc)
	}
	docsFS := http.FileServer(http.Dir(mock.docsDir))
	handler.Handle("/", docsFS)
	//handler.HandleFunc("/", mock.handleDocs)
	mock.httpServer = &http.Server{
		Handler: handler,
	}
	mock.httpsServer = &http.Server{
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: mock.getServerCertificate,
		},
	}
	go mock.listenAndServe()
	go mock.listenAndServeTLS()
}

func (mock *mockServer) listenAndServe() {
	log.Println("Starting HTTP server...")
	mock.stoppedWG.Add(1)
	defer mock.stoppedWG.Done()
	err := mock.httpServer.Serve(mock.httpListener)
	if !errors.Is(err, http.ErrServerClosed) {
		log.Println("HTTP server failure: ", err)
		return
	}
	log.Println("HTTP server stopped")
}

func (mock *mockServer) listenAndServeTLS() {
	log.Println("Starting HTTPS server...")
	mock.stoppedWG.Add(1)
	defer mock.stoppedWG.Done()
	err := mock.httpsServer.ServeTLS(mock.httpsListener, "", "")
	if !errors.Is(err, http.ErrServerClosed) {
		log.Println("HTTPS server failure: ", err)
		return
	}
	log.Println("HTTPS server stopped")
}

func (mock *mockServer) getServerCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certificate, err := tls.X509KeyPair(mockCertificatePEM(), mockKeyPEM())
	return &certificate, err
}
