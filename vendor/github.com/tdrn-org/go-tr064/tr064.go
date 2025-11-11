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

//go:generate go run cmd/build/build.go generate

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// ErrDocNotFound indicates a specification document was not found.
var ErrDocNotFound = errors.New("document not found")

// XMLNameSpace defines the XML namespace to use for SOAP calls.
const XMLNameSpace = "http://schemas.xmlsoap.org/soap/envelope/"

// XMLEncodingStyle defines the XML encoding to use for SOAP calls.
const XMLEncodingStyle = "http://schemas.xmlsoap.org/soap/encoding/"

// ServiceSpec represents a well-known TR-064 specification document.
type ServiceSpec string

const (
	// DefaultServiceSpec defines the default TR-064 specification to be assumed available for
	// any TR-064 capabable device.
	DefaultServiceSpec ServiceSpec = "tr64desc"
	// IgdServiceSpec defines the TR-064 specification of Internet-Gateway-Devices (e.g. router).
	IgdServiceSpec ServiceSpec = "igddesc"
)

// Name gets the specification name.
func (spec ServiceSpec) Name() string {
	return string(spec)
}

// Path gets the specification path relative to the device URL.
func (spec ServiceSpec) Path() string {
	return "/" + string(spec) + ".xml"
}

func fetchServiceSpec(client *http.Client, deviceUrl *url.URL, spec ServiceSpec) (*tr64descDoc, error) {
	tr64descUrl := deviceUrl.JoinPath(spec.Path())
	tr64desc := &tr64descDoc{}
	err := unmarshalXMLDocument(client, tr64descUrl, tr64desc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch '%s' (cause: %w)", tr64descUrl, err)
	}
	tr64desc.bind(spec)
	return tr64desc, nil
}

func unmarshalXMLDocument(client *http.Client, docUrl *url.URL, v any) error {
	response, err := client.Get(docUrl.String())
	if err != nil {
		return fmt.Errorf("failed to access URL '%s' (cause: %w)", docUrl, err)
	}
	switch response.StatusCode {
	case http.StatusOK:
		// success, simply move on
	case http.StatusNotFound:
		return fmt.Errorf("failed to get URL '%s' (cause: %w)", docUrl, ErrDocNotFound)
	default:
		return fmt.Errorf("failed to get URL '%s' (status: %s)", docUrl, response.Status)
	}
	document := response.Body
	defer document.Close()
	documentBytes, err := io.ReadAll(document)
	if err != nil {
		return fmt.Errorf("failed to read URL '%s' (cause: %w)", docUrl, err)
	}
	err = xml.Unmarshal(documentBytes, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal URL '%s' (cause: %w)", docUrl, err)
	}
	return nil
}

type tr64descDoc struct {
	SpecVersion   specVersionDoc   `xml:"specVersion"`
	SystemVersion systemVersionDoc `xml:"systemVersion"`
	Device        deviceDoc        `xml:"device"`
}

func (doc *tr64descDoc) bind(spec ServiceSpec) {
	doc.Device.bind(spec)
}

type WalkServiceFunc func(*serviceDoc, *scpdDoc) error

func (doc *tr64descDoc) walk(client *http.Client, baseUrl *url.URL, f WalkServiceFunc) error {
	return doc.Device.walk(client, baseUrl, f)
}

type specVersionDoc struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}

func (doc *specVersionDoc) signature() string {
	return fmt.Sprintf("spec version: %d.%d", doc.Major, doc.Minor)
}

type systemVersionDoc struct {
	HW          int    `xml:"HW"`
	Major       int    `xml:"Major"`
	Minor       int    `xml:"Minor"`
	Patch       int    `xml:"Patch"`
	Buildnumber int    `xml:"Buildnumber"`
	Display     string `xml:"Display"`
}

type deviceListDoc struct {
	Devices []deviceDoc `xml:"device"`
}

type deviceDoc struct {
	DeviceType       string         `xml:"deviceType"`
	FriendlyName     string         `xml:"friendlyName"`
	Manufacturer     string         `xml:"manufacturer"`
	ManufacturerURL  string         `xml:"manufacturerURL"`
	ModelDescription string         `xml:"modelDescription"`
	ModelName        string         `xml:"modelName"`
	ModelNumber      string         `xml:"modelNumber"`
	ModelURL         string         `xml:"modelURL"`
	UDN              string         `xml:"UDN"`
	SerialNumber     string         `xml:"serialNumber"`
	OriginUDN        string         `xml:"originUDN"`
	ServiceList      serviceListDoc `xml:"serviceList"`
	DeviceList       deviceListDoc  `xml:"deviceList"`
	PresentationURL  string         `xml:"presentationURL"`
}

func (doc *deviceDoc) bind(spec ServiceSpec) {
	for serviceIndex := range doc.ServiceList.Services {
		doc.ServiceList.Services[serviceIndex].bind(spec)
	}
	for deviceIndex := range doc.DeviceList.Devices {
		doc.DeviceList.Devices[deviceIndex].bind(spec)
	}
}

func (doc *deviceDoc) walk(client *http.Client, baseUrl *url.URL, f WalkServiceFunc) error {
	for _, service := range doc.ServiceList.Services {
		if strings.HasPrefix(service.ServiceType, "urn:schemas-any-com:service:Any:") {
			continue
		}
		scpd, err := service.scpd(client, baseUrl)
		if err != nil {
			return err
		}
		err = f(&service, scpd)
		if err != nil {
			return err
		}
	}
	for _, device := range doc.DeviceList.Devices {
		err := device.walk(client, baseUrl, f)
		if err != nil {
			return err
		}
	}
	return nil
}

type serviceListDoc struct {
	Services []serviceDoc `xml:"service"`
}

type serviceDoc struct {
	ServiceType string      `xml:"serviceType"`
	ServiceId   string      `xml:"serviceId"`
	ControlURL  string      `xml:"controlURL"`
	EventSubURL string      `xml:"eventSubURL"`
	SCPDURL     string      `xml:"SCPDURL"`
	spec        ServiceSpec `xml:"-"`
	cachedSCPD  *scpdDoc    `xml:"-"`
}

func (doc *serviceDoc) bind(spec ServiceSpec) {
	doc.spec = spec
}

func (doc *serviceDoc) scpd(client *http.Client, baseUrl *url.URL) (*scpdDoc, error) {
	if doc.cachedSCPD == nil {
		scpdUrl := baseUrl.JoinPath(doc.SCPDURL)
		scpd := &scpdDoc{}
		err := unmarshalXMLDocument(client, scpdUrl, scpd)
		if err != nil {
			return nil, err
		}
		doc.cachedSCPD = scpd
	}
	return doc.cachedSCPD, nil
}

var serviceSCPDPattern = regexp.MustCompile(`^/(.+)SCPD.xml$`)

func (service *serviceDoc) scpdName() string {
	match := serviceSCPDPattern.FindStringSubmatch(service.SCPDURL)
	if match == nil {
		log.Fatal(fmt.Errorf("unexpected SCPD URL '%s'", service.SCPDURL))
	}
	return match[1]
}

func (service *serviceDoc) Spec() ServiceSpec {
	return service.spec
}

func (service *serviceDoc) Type() string {
	return service.ServiceType
}

func (service *serviceDoc) ShortType() string {
	return serviceShortType(service.ServiceType)
}

func (service *serviceDoc) Id() string {
	return service.ServiceId
}

func (service *serviceDoc) ShortId() string {
	return serviceShortId(service.ServiceId)
}

func (service *serviceDoc) ControlUrl() string {
	return service.ControlURL
}

type scpdDoc struct {
	SpecVersion       specVersionDoc       `xml:"specVersion"`
	ActionList        actionListDoc        `xml:"actionList"`
	ServiceStateTable serviceStateTableDoc `xml:"serviceStateTable"`
}

func (doc *scpdDoc) lookupVariable(name string) *stateVariableDoc {
	for _, variable := range doc.ServiceStateTable.StateVariables {
		if variable.Name == name {
			return &variable
		}
	}
	return nil
}

type actionListDoc struct {
	Actions []actionDoc `xml:"action"`
}

type actionDoc struct {
	Name         string          `xml:"name"`
	ArgumentList argumentListDoc `xml:"argumentList"`
}

type argumentListDoc struct {
	Arguments []argumentDoc `xml:"argument"`
}

func (doc *argumentListDoc) hasIn() bool {
	for _, argument := range doc.Arguments {
		if argument.Direction == "in" {
			return true
		}
	}
	return false
}

func (doc *argumentListDoc) hasOut() bool {
	for _, argument := range doc.Arguments {
		if argument.Direction == "out" {
			return true
		}
	}
	return false
}

type argumentDoc struct {
	Name                 string `xml:"name"`
	Direction            string `xml:"direction"`
	RelatedStateVariable string `xml:"relatedStateVariable"`
}

type serviceStateTableDoc struct {
	StateVariables []stateVariableDoc `xml:"stateVariable"`
}

type stateVariableDoc struct {
	Name          string              `xml:"name"`
	DataType      string              `xml:"dataType"`
	DefaultValue  string              `xml:"defaultValue"`
	AllowedValues allowedValueListDoc `xml:"allowedValueList"`
}

type allowedValueListDoc struct {
	AllowedValues []string `xml:"allowedValue"`
}

var serviceShortTypePattern = regexp.MustCompile(`^urn\:(.+)\:service\:(.+):\d+$`)

func serviceShortType(serviceType string) string {
	match := serviceShortTypePattern.FindStringSubmatch(serviceType)
	if match == nil {
		log.Fatal("Unexpected service type '", serviceType, "'")
	}
	mangledServiceType := mangleName(match[2])
	return mangledServiceType
}

var serviceShortIdPattern = regexp.MustCompile(`^urn\:(.+)\:serviceId\:(.+)$`)

func serviceShortId(serviceId string) string {
	match := serviceShortIdPattern.FindStringSubmatch(serviceId)
	if match == nil {
		log.Fatal("Unexpected service id '", serviceId, "'")
	}
	mangledServiceId := mangleName(match[2])
	return mangledServiceId
}
