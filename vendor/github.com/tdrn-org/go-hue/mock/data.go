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
	_ "embed"
	"encoding/json"
	"log"

	"github.com/tdrn-org/go-hue/hueapi"
)

// Data contains a Bridge's state as returned by the different Get*s API calls.
type Data struct {
	GetResources *struct {
		Data   *[]hueapi.ResourceGet `json:"data,omitempty"`
		Errors *[]hueapi.Error       `json:"errors,omitempty"`
	} `json:"resources"`
	GetBridges *struct {
		Data   *[]hueapi.BridgeGet `json:"data,omitempty"`
		Errors *[]hueapi.Error     `json:"errors,omitempty"`
	} `json:"bridges"`
	GetBridgeHomes *struct {
		Data   *[]hueapi.BridgeHomeGet `json:"data,omitempty"`
		Errors *[]hueapi.Error         `json:"errors,omitempty"`
	} `json:"bridge_homes"`
	GetDevices *struct {
		Data   *[]hueapi.DeviceGet `json:"data,omitempty"`
		Errors *[]hueapi.Error     `json:"errors,omitempty"`
	} `json:"devices"`
	GetDevicePowers *struct {
		Data   *[]hueapi.DevicePowerGet `json:"data,omitempty"`
		Errors *[]hueapi.Error          `json:"errors,omitempty"`
	} `json:"device_powers"`
	GetGroupedLights *struct {
		Data   *[]hueapi.GroupedLightGet `json:"data,omitempty"`
		Errors *[]hueapi.Error           `json:"errors,omitempty"`
	} `json:"grouped_lights"`
	GetLights *struct {
		Data   *[]hueapi.LightGet `json:"data,omitempty"`
		Errors *[]hueapi.Error    `json:"errors,omitempty"`
	} `json:"lights"`
	GetLightLevels *struct {
		Data   *[]hueapi.LightLevelGet `json:"data,omitempty"`
		Errors *[]hueapi.Error         `json:"errors,omitempty"`
	} `json:"light_levels"`
	GetMotionSensors *struct {
		Data   *[]hueapi.MotionGet `json:"data,omitempty"`
		Errors *[]hueapi.Error     `json:"errors,omitempty"`
	} `json:"motion_sensors"`
	GetRooms *struct {
		Data   *[]hueapi.RoomGet `json:"data,omitempty"`
		Errors *[]hueapi.Error   `json:"errors,omitempty"`
	} `json:"rooms"`
	GetScenes *struct {
		Data   *[]hueapi.SceneGet `json:"data,omitempty"`
		Errors *[]hueapi.Error    `json:"errors,omitempty"`
	} `json:"scenes"`
	GetSmartScenes *struct {
		Data   *[]hueapi.SmartSceneGet `json:"data,omitempty"`
		Errors *[]hueapi.Error         `json:"errors,omitempty"`
	} `json:"smart_scenes"`
	GetTemperatures *struct {
		Data   *[]hueapi.TemperatureGet `json:"data,omitempty"`
		Errors *[]hueapi.Error          `json:"errors,omitempty"`
	} `json:"temperatures"`
	GetZones *struct {
		Data   *[]hueapi.RoomGet `json:"data,omitempty"`
		Errors *[]hueapi.Error   `json:"errors,omitempty"`
	} `json:"zones"`
}

//go:embed "mock.json"
var mockDataBytes []byte
var mockData *Data = &Data{}

func init() {
	err := json.Unmarshal(mockDataBytes, mockData)
	if err != nil {
		log.Fatal("mock.json: ", err)
	}
}
