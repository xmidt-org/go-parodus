/*
 *  Copyright 2019 Comcast Cable Communications Management, LLC
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"encoding/json"
	"errors"
)

type ParamType int

const (
	STRING ParamType = iota
	INTEGER
	UNSIGNEDINT
	BOOLEAN
	DATETIME
	BASE64
	LONG
	UNSIGNEDLONG
	FLOAT
	DOUBLE
	BYTE
)

const (
	SETOPER = "SET"
	GETOPER = "GET"
)

type Parameter struct {
	name     string      `json:"name"`
	value    interface{} `json:"value"`
	dateType ParamType   `json:"dataType"`
}

func NewBoolParam(name string, value bool) Parameter {
	return Parameter{
		name:     name,
		value:    value,
		dateType: BOOLEAN,
	}
}

func NewIntParam(name string, value int) Parameter {
	return Parameter{
		name:     name,
		value:    value,
		dateType: INTEGER,
	}
}

func NewStringParam(name string, value string) Parameter {
	return Parameter{
		name:     name,
		value:    value,
		dateType: STRING,
	}
}

func (param *Parameter) SetBool(value bool) {
	param.value = value
	param.dateType = BOOLEAN
}

func (param *Parameter) SetInt(value int) {
	param.value = value
	param.dateType = INTEGER
}

func (param *Parameter) SetString(value string) {
	param.value = value
	param.dateType = STRING
}

func (param *Parameter) GetValue() interface{} {
	return param.value
}

func (param *Parameter) UnmarshalJSON(data []byte) error {
	x := map[string]interface{}{}

	err := json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	if len(x) != 3 {
		return errors.New("unknown fields")
	}
	if name, ok := x["name"]; ok {
		param.name = name.(string)
	} else {
		return errors.New("no name param")
	}

	if dataType, ok := x["dataType"]; ok {
		tempValue := int(dataType.(float64))
		param.dateType = ParamType(tempValue)
	} else {
		return errors.New("no data type param")
	}
	if value, ok := x["value"]; ok {
		param.value = value
	} else {
		return errors.New("no value param")
	}
	return nil
}

func (param *Parameter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":     param.name,
		"value":    param.value,
		"dataType": param.dateType,
	})
}

type ConfigRequest struct {
	Command    string             `json:"command"`
	Names      []string           `json:"names"`
	Parameters []ConfigParameters `json:"parameters"`
}

func (req ConfigRequest) AsMap() map[string]Parameter {
	data := make(map[string]Parameter)
	for _, x := range req.Parameters {
		data[x.Name] = Parameter{
			name:     x.Name,
			dateType: ParamType(x.DataType),
			value:    x.Value,
		}
	}
	return data
}

type ConfigResponse struct {
	StatusCode int                `json:"statusCode"`
	Parameters []ConfigParameters `json:"parameters,omitempty"`
	Message    string             `json:"message,omitempty"`
}

func (res ConfigResponse) AsMap() map[string]Parameter {
	data := make(map[string]Parameter)
	for _, x := range res.Parameters {
		data[x.Name] = Parameter{
			name:     x.Name,
			dateType: ParamType(x.DataType),
			value:    x.Value,
		}
	}
	return data
}

type ConfigParameters struct {
	Name     string      `json:"name,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	DataType int         `json:"dataType,omitempty"`
	Message  string      `json:"message,omitempty"`
}
