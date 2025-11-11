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

package hueapi

//go:generate go run ../cmd/build/build.go fetch https://github.com/openhue/openhue-api/releases/download/0.17/openhue.yaml openhue.gen.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi-codegen-openhue.yaml openhue.gen.yaml

import (
	_ "embed"
)

// The openhue-api spec yaml file used to generate the API code.
//
//go:embed openhue.gen.yaml
var OpenHueApiSpecYaml []byte

// The Hue API authentication header key.
const ApplicationKeyHeader = "hue-application-key"
