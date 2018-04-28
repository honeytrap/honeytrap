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
package elasticsearch

import (
	"crypto/tls"
	"errors"
	"strings"

	"net/http"
	"net/url"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	ErrElasticsearchNoURL   = errors.New("Elasticsearch url has not been set")
	ErrElasticsearchNoIndex = errors.New("Elasticsearch index has not been set")
)

// Config defines a struct which holds configuration values for a SearchBackend.
type Config struct {
	options []elastic.ClientOptionFunc

	Index string
}

// UnmarshalTOML deserializes the giving data into the config.
func (c *Config) UnmarshalTOML(p interface{}) error {
	c.options = []elastic.ClientOptionFunc{
		elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 5,
				TLSClientConfig:     &tls.Config{},
			},
			Timeout: time.Duration(20) * time.Second,
		}),
		elastic.SetSniff(false),
		elastic.SetRetrier(&Retrier{}),
	}

	data, _ := p.(map[string]interface{})

	v, ok := data["url"]
	if !ok {
		return ErrElasticsearchNoURL
	}
	s, _ := v.(string)
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) != 2 {
		return ErrElasticsearchNoIndex
	}

	c.Index = parts[1]

	// remove path
	u.Path = ""
	c.options = append(c.options, elastic.SetURL(u.String()))

	log.Debugf("Using URL: %s with index: %s", u.String(), c.Index)

	if username, ok := data["username"]; !ok {
	} else if password := data["password"]; !ok {
	} else {
		username := username.(string)
		password := password.(string)
		c.options = append(c.options, elastic.SetBasicAuth(username, password))

		log.Debugf("Using authentication with username: %s and password.", username)
	}

	return nil
}
