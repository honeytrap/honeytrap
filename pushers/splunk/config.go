/*
* Honeytrap
* Copyright (C) 2016-2017 DutchSec (https://dutchsec.com/)
*
* This program is free software; you can redistribute it and/or modify it under
* the terms of the GNU Affero General Public License version 3 as published by the
* Free Software Foundation.
*
* This program is distributed in the hope that it will be useful, but WITHOUT
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
* FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
* details.
*
* You should have received a copy of the GNU Affero General Public License
* version 3 along with this program in the file "LICENSE".  If not, see
* <http://www.gnu.org/licenses/agpl-3.0.txt>.
*
* See https://honeytrap.io/ for more details. All requests should be sent to
* licensing@honeytrap.io
*
* The interactive user interfaces in modified source and object code versions
* of this program must display Appropriate Legal Notices, as required under
* Section 5 of the GNU Affero General Public License version 3.
*
* In accordance with Section 7(b) of the GNU Affero General Public License version 3,
* these Appropriate Legal Notices must retain the display of the "Powered by
* Honeytrap" logo and retain the original copyright notice. If the display of the
* logo is not reasonably feasible for technical reasons, the Appropriate Legal Notices
* must display the words "Powered by Honeytrap" and retain the original copyright notice.
 */
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
