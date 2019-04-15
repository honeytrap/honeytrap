// Copyright 2016-2019 DutchSec (https://dutchsec.com/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package splunk

import (
	"crypto/tls"
	"errors"
	"net/url"
)

var (
	ErrEndpointsNotSet = errors.New("Endpoints has not been set")
	ErrTokenNotSet     = errors.New("Token has not been set")
)

// Config defines a struct which holds configuration values for a SearchBackend.
type Config struct {
	Endpoints []string
	Token     string

	tlsConfig *tls.Config
}

// UnmarshalTOML deserializes the giving data into the config.
func (c *Config) UnmarshalTOML(p interface{}) error {
	c.tlsConfig = &tls.Config{}

	data, _ := p.(map[string]interface{})

	if v, ok := data["endpoints"]; !ok {
		return ErrEndpointsNotSet
	} else if s, ok := v.([]interface{}); !ok {
		return ErrEndpointsNotSet
	} else {
		for _, e := range s {
			if _, ok := e.(string); !ok {
			} else if u, err := url.Parse(e.(string)); err != nil {
				return err
			} else {
				c.Endpoints = append(c.Endpoints, u.String())
			}
		}

	}

	if token, ok := data["token"]; !ok {
		return ErrTokenNotSet
	} else if v, ok := token.(string); !ok {
		return ErrTokenNotSet
	} else {
		c.Token = v
	}

	if v, ok := data["verify"]; !ok {
	} else if v, ok := v.(bool); !ok {
	} else {
		c.tlsConfig = &tls.Config{
			InsecureSkipVerify: !v,
		}
	}

	return nil
}
