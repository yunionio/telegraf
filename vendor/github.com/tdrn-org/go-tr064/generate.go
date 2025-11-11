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
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"go/format"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Generate generates Go code suitable for invoking the services defined in the given specification.
//
// The base URL points towards the device providing the service.
// The spec argument defines the specification use for code generation.
// The code is generated within the given directory. Already existing files are
// overwritten without notice.
func Generate(baseUrl *url.URL, spec ServiceSpec, dir string) {
	specUrl := baseUrl.JoinPath(spec.Path())
	log.Println("Reading '", specUrl.Redacted(), "'...")
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	tr64desc, err := fetchServiceSpec(httpClient, baseUrl, spec)
	if err != nil {
		log.Fatal("Failed to fetch ", specUrl, " cause: ", err)
	}
	log.Println("Generating...")
	gc := &generateContext{
		httpClient:     httpClient,
		baseUrl:        baseUrl,
		spec:           spec,
		serviceClients: make(map[string]*bytes.Buffer),
		serviceTypes:   make(map[string]string),
		serviceTests:   make(map[string]*bytes.Buffer),
	}
	err = gc.generate(tr64desc, dir)
	if err != nil {
		log.Fatal("Failed to generate:", err)
	}
}

type generateContext struct {
	httpClient     *http.Client
	baseUrl        *url.URL
	spec           ServiceSpec
	serviceClients map[string]*bytes.Buffer
	serviceTypes   map[string]string
	serviceTests   map[string]*bytes.Buffer
	err            error
}

func (gc *generateContext) generate(tr64desc *tr64descDoc, dir string) error {
	gc.err = tr64desc.walk(gc.httpClient, gc.baseUrl, gc.generateServiceClient)
	if gc.err != nil {
		return gc.err
	}
	gc.err = tr64desc.walk(gc.httpClient, gc.baseUrl, gc.generateServiceTest)
	if gc.err != nil {
		return gc.err
	}
	gc.flushFiles(dir)
	return gc.err
}

func (gc *generateContext) generateServiceClient(service *serviceDoc, scpd *scpdDoc) error {
	packageName := service.scpdName()
	gc.serviceTypes[service.ShortType()] = packageName
	buffer := gc.serviceClients[packageName]
	if buffer == nil {
		buffer = &bytes.Buffer{}
		gc.serviceClients[packageName] = buffer
		gc.emit(buffer, "// generated from %s\n", scpd.SpecVersion.signature())
		gc.emit(buffer, "package %s\n", packageName)
		gc.emit(buffer, "import (\n")
		gc.emit(buffer, "\"encoding/xml\"\n")
		gc.emit(buffer, "\"github.com/tdrn-org/go-tr064\"\n")
		gc.emit(buffer, ")\n")
		gc.emit(buffer, "\ntype ServiceClient struct {\n")
		gc.emit(buffer, "TR064Client *tr064.Client\n")
		gc.emit(buffer, "Service tr064.ServiceDescriptor\n")
		gc.emit(buffer, "}\n")
		for _, action := range scpd.ActionList.Actions {
			hasRequest := gc.generateServiceClientActionArgument(buffer, scpd, &action, "in")
			hasResponse := gc.generateServiceClientActionArgument(buffer, scpd, &action, "out")
			gc.generateServiceClientAction(buffer, &action, hasRequest, hasResponse)
		}
	}
	return gc.err
}

func (gc *generateContext) generateServiceClientActionArgument(buffer *bytes.Buffer, scpd *scpdDoc, action *actionDoc, direction string) bool {
	if gc.err != nil {
		return false
	}
	actionName := mangleName(action.Name)
	var typeName string
	if direction == "in" {
		typeName = actionName + "Request"
	} else if direction == "out" {
		typeName = actionName + "Response"
	} else {
		log.Fatal("Unexpected direction '", direction, "'")
	}
	gc.emit(buffer, "\ntype %s struct {\n", typeName)
	if direction == "in" {
		gc.emit(buffer, "XMLName xml.Name `xml:\"u:%sRequest\"`\n", action.Name)
		gc.emit(buffer, "XMLNameSpace string `xml:\"xmlns:u,attr\"`\n")
	} else {
		gc.emit(buffer, "XMLName xml.Name `xml:\"%sResponse\"`\n", action.Name)
	}
	variableCount := 0
	for _, argument := range action.ArgumentList.Arguments {
		if argument.Direction == direction {
			variable := scpd.lookupVariable(argument.RelatedStateVariable)
			if variable == nil {
				gc.err = fmt.Errorf("unknown state variable '%s'", argument.Name)
				return false
			}
			variableName := mangleName(argument.Name)
			variableType := variableType(variable)
			gc.emit(buffer, "%s %s `xml:\"%s\"`\n", variableName, variableType, argument.Name)
			variableCount++
		}
	}
	gc.emit(buffer, "}\n")
	return variableCount > 0
}

func (gc *generateContext) generateServiceClientAction(buffer *bytes.Buffer, action *actionDoc, hasRequest bool, hasResponse bool) {
	if gc.err != nil {
		return
	}
	actionName := mangleName(action.Name)
	requestTypeName := actionName + "Request"
	responseTypeName := actionName + "Response"
	if hasRequest && hasResponse {
		gc.emit(buffer, "\nfunc (client *ServiceClient) %s(in *%s, out *%s) error {\n", actionName, requestTypeName, responseTypeName)
	} else if hasRequest {
		gc.emit(buffer, "\nfunc (client *ServiceClient) %s(in *%s) error {\n", actionName, requestTypeName)
	} else if hasResponse {
		gc.emit(buffer, "\nfunc (client *ServiceClient) %s(out *%s) error {\n", actionName, responseTypeName)
	} else {
		gc.emit(buffer, "\nfunc (client *ServiceClient) %s() error {\n", actionName)
	}
	if hasRequest {
		gc.emit(buffer, "in.XMLNameSpace = client.Service.Type()\n")
	} else {
		gc.emit(buffer, "in := &%s{XMLNameSpace: client.Service.Type() }\n", requestTypeName)
	}
	if !hasResponse {
		gc.emit(buffer, "out := &%s{}\n", responseTypeName)
	}
	gc.emit(buffer, "return client.TR064Client.InvokeService(client.Service, \"%s\", tr064.NewSOAPRequest(in), tr064.NewSOAPResponse(out))\n", action.Name)
	gc.emit(buffer, "}\n")
}

func (gc *generateContext) generateServiceTest(service *serviceDoc, scpd *scpdDoc) error {
	packageName := service.scpdName()
	buffer := gc.serviceTests[packageName]
	if buffer == nil {
		buffer = &bytes.Buffer{}
		gc.serviceTests[packageName] = buffer
		gc.emit(buffer, "// generated from %s\n", scpd.SpecVersion.signature())
		gc.emit(buffer, "package services_test\n")
		gc.emit(buffer, "import (\n")
		gc.emit(buffer, "\"log\"\n")
		gc.emit(buffer, "\"net/http\"\n")
		gc.emit(buffer, "\"testing\"\n")
		gc.emit(buffer, "\"github.com/stretchr/testify/require\"\n")
		gc.emit(buffer, "\"github.com/tdrn-org/go-tr064\"\n")
		gc.emit(buffer, "\"github.com/tdrn-org/go-tr064/mock\"\n")
		gc.emit(buffer, "\"github.com/tdrn-org/go-tr064/services/%s/%s\"\n", gc.spec.Name(), packageName)
		gc.emit(buffer, ")\n")
		gc.emit(buffer, "var %sMock = &mock.ServiceMock {\n", packageName)
		gc.emit(buffer, "Path: \"%s\",\n", service.ControlURL)
		gc.emit(buffer, "HandleFunc: %sHandler,\n", packageName)
		gc.emit(buffer, "}\n")
		gc.emit(buffer, "\nfunc Test%s(t *testing.T) {\n", service.ShortType())
		gc.emit(buffer, "// Start mock server\n")
		gc.emit(buffer, "tr064Mock := mock.Start(\"testdata\", %sMock)\n", packageName)
		gc.emit(buffer, "defer tr064Mock.Shutdown()\n")
		gc.emit(buffer, "// Actual test\n")
		gc.emit(buffer, "client := tr064.NewClient(tr064Mock.Server())\n")
		gc.emit(buffer, "client.Debug = true\n")
		gc.emit(buffer, "serviceClient := &%s.ServiceClient{\n", packageName)
		gc.emit(buffer, "TR064Client: client,\n")
		gc.emit(buffer, "Service: &tr064.StaticServiceDescriptor{\n")
		gc.emit(buffer, "ServiceSpec: tr064.ServiceSpec(\"%s\"),\n", gc.spec.Name())
		gc.emit(buffer, "ServiceType: \"%s\",\n", service.Type())
		gc.emit(buffer, "ServiceId: \"%s\",\n", service.Id())
		gc.emit(buffer, "ServiceControlUrl:  \"%s\",\n", service.ControlURL)
		gc.emit(buffer, "},\n")
		gc.emit(buffer, "}\n")
		gc.generateServiceTestBlocks(buffer, packageName, scpd)
		gc.emit(buffer, "}\n")
		gc.generateServiceTestMock(buffer, packageName, scpd)
	}
	return gc.err
}

func (gc *generateContext) generateServiceTestBlocks(buffer *bytes.Buffer, packageName string, scpd *scpdDoc) {
	for _, action := range scpd.ActionList.Actions {
		actionName := mangleName(action.Name)
		hasRequest := action.ArgumentList.hasIn()
		hasResponse := action.ArgumentList.hasOut()
		gc.emit(buffer, "{\n")
		if hasRequest {
			gc.emit(buffer, "in := &%s.%sRequest{}\n", packageName, actionName)
		}
		if hasResponse {
			gc.emit(buffer, "out := &%s.%sResponse{}\n", packageName, actionName)
		}
		if hasRequest && hasResponse {
			gc.emit(buffer, "require.NoError(t, serviceClient.%s(in,out))\n", actionName)
		} else if hasRequest {
			gc.emit(buffer, "require.NoError(t, serviceClient.%s(in))\n", actionName)
		} else if hasResponse {
			gc.emit(buffer, "require.NoError(t, serviceClient.%s(out))\n", actionName)
		} else {
			gc.emit(buffer, "require.NoError(t, serviceClient.%s())\n", actionName)
		}
		gc.emit(buffer, "}\n")
	}
}

func (gc *generateContext) generateServiceTestMock(buffer *bytes.Buffer, packageName string, scpd *scpdDoc) {
	gc.emit(buffer, "\nfunc %sHandler(w http.ResponseWriter, req *http.Request) {\n", packageName)
	gc.emit(buffer, "log.Println(\"Mock: \", req.URL)\n")
	gc.emit(buffer, "action, err := mock.UnmarshalSoapAction(w, req)\n")
	gc.emit(buffer, "if err != nil {\n")
	gc.emit(buffer, "log.Println(err)\n")
	gc.emit(buffer, "return\n")
	gc.emit(buffer, "}\n")
	gc.emit(buffer, "switch action {\n")
	for _, action := range scpd.ActionList.Actions {
		actionName := mangleName(action.Name)
		gc.emit(buffer, "case \"%s\":\n", action.Name)
		gc.emit(buffer, "%s_%s(w)\n", packageName, actionName)
	}
	gc.emit(buffer, "\n")
	gc.emit(buffer, "\n")
	gc.emit(buffer, "\n")
	gc.emit(buffer, "default:\n")
	gc.emit(buffer, "log.Println(\"Unknown action: \", action)\n")
	gc.emit(buffer, "w.WriteHeader(http.StatusBadRequest)\n")
	gc.emit(buffer, "}\n")
	gc.emit(buffer, "}\n")
	for _, action := range scpd.ActionList.Actions {
		actionName := mangleName(action.Name)
		gc.emit(buffer, "\nfunc %s_%s(w http.ResponseWriter) {\n", packageName, actionName)
		gc.emit(buffer, "out := %s.%sResponse{}\n", packageName, actionName)
		gc.emit(buffer, "err := mock.WriteSoapResponse(w,out)\n")
		gc.emit(buffer, "if err != nil {\n")
		gc.emit(buffer, "log.Println(err)\n")
		gc.emit(buffer, "}\n")
		gc.emit(buffer, "}\n")
	}
}

func (gc *generateContext) emit(buffer *bytes.Buffer, format string, a ...any) {
	if gc.err != nil {
		return
	}
	_, gc.err = buffer.WriteString(fmt.Sprintf(format, a...))
}

func (gc *generateContext) flushFiles(dir string) {
	gc.flushServiceClientFiles(dir)
	gc.flushServiceTestFiles(dir)
}

func (gc *generateContext) flushServiceClientFiles(dir string) {
	if gc.err != nil {
		return
	}
	for packageName, buffer := range gc.serviceClients {
		log.Println("Writing service client '", packageName, "'...")
		packageDir := filepath.Join(dir, "services", gc.spec.Name(), packageName)
		code, err := format.Source(buffer.Bytes())
		if err != nil {
			gc.err = fmt.Errorf("failed to format generated service client code (cause: %w)", err)
			return
		}
		err = os.MkdirAll(packageDir, 0777)
		if err != nil {
			gc.err = fmt.Errorf("failed to create service client directory '%s' (cause: %w)", packageName, err)
			return
		}
		file := filepath.Join(packageDir, packageName+".go")
		err = os.WriteFile(file, code, 0666)
		if err != nil {
			gc.err = fmt.Errorf("failed to write service client file '%s' (cause: %w)", file, err)
			return
		}
	}
	for serviceType, packageName := range gc.serviceTypes {
		log.Println("Writing service name '", packageName, "'/'", serviceType, "'...")
		buffer := &bytes.Buffer{}
		gc.emit(buffer, "// %s\n", serviceType)
		gc.emit(buffer, "package %s\n", packageName)
		gc.emit(buffer, "const ServiceShortType = \"%s\"\n", serviceType)
		code, err := format.Source(buffer.Bytes())
		if err != nil {
			gc.err = fmt.Errorf("failed to format generated service name code (cause: %w)", err)
			return
		}
		file := filepath.Join(dir, "services", gc.spec.Name(), packageName, "name.go")
		err = os.WriteFile(file, code, 0666)
		if err != nil {
			gc.err = fmt.Errorf("failed to write service name file '%s' (cause: %w)", file, err)
			return
		}
	}
}

func (gc *generateContext) flushServiceTestFiles(dir string) {
	if gc.err != nil {
		return
	}
	for packageName, buffer := range gc.serviceTests {
		log.Println("Writing service test '", packageName, "'...")
		packageDir := filepath.Join(dir, "services", gc.spec.Name())
		code, err := formatSource(buffer.Bytes())
		if err != nil {
			gc.err = fmt.Errorf("failed to format generated service test code (cause: %w)", err)
			return
		}
		err = os.MkdirAll(packageDir, 0777)
		if err != nil {
			gc.err = fmt.Errorf("failed to create service test directory '%s' (cause: %w)", packageName, err)
			return
		}
		file := filepath.Join(packageDir, packageName+"_test.go")
		err = os.WriteFile(file, code, 0666)
		if err != nil {
			gc.err = fmt.Errorf("failed to write service test file '%s' (cause: %w)", file, err)
			return
		}
	}
}

func formatSource(src []byte) ([]byte, error) {
	code, err := format.Source(src)
	if err != nil {
		scanner := bufio.NewScanner(bytes.NewReader(src))
		line := 1
		for scanner.Scan() {
			fmt.Println(line, ": ", scanner.Text())
			line++
		}
	}
	return code, err
}

func mangleName(name string) string {
	mangled := strings.ReplaceAll(name, "-", "_")
	mangled = strings.ReplaceAll(mangled, ".", "_")
	return mangled
}

var dataTypeMap map[string]string = map[string]string{
	"i1":       "int8",
	"i2":       "int16",
	"i4":       "int32",
	"ui1":      "uint8",
	"ui2":      "uint16",
	"ui4":      "uint32",
	"boolean":  "bool",
	"string":   "string",
	"dateTime": "string",
	"uuid":     "string",
}

func variableType(variable *stateVariableDoc) string {
	mappedType := dataTypeMap[variable.DataType]
	if mappedType != "" {
		return mappedType
	}
	log.Println("Unknown variable data type '", variable.DataType, "' resolved to any")
	return "any"
}
