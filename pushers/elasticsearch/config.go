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
	// ErrElasticsearchNoURL will be returned if no url has been set in configuration
	ErrElasticsearchNoURL = errors.New("Elasticsearch url has not been set")
	// ErrElasticsearchNoIndex will be returned if no path has been set in the url which is being used as index
	ErrElasticsearchNoIndex = errors.New("Elasticsearch index has not been set")
)

// Config defines a struct which holds configuration values for a SearchBackend.
type Config struct {
	options []elastic.ClientOptionFunc

	// URL configures the Elasticsearch server and index to send messages to
	URL *url.URL `toml:"url"`

	// Insecure configures if the client should not verify tls configuration
	InsecureSkipVerify bool `toml:"insecure"`

	// Sniff defines if the client should find all nodes
	Sniff bool `toml:"sniff"`

	index string
}

// UnmarshalTOML deserializes the giving data into the config.
func (c *Config) UnmarshalTOML(p interface{}) error {
	tlsConfig := &tls.Config{}

	c.options = []elastic.ClientOptionFunc{
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

	c.index = parts[1]

	// remove path
	u.Path = ""
	c.URL = u

	c.options = append(c.options, elastic.SetURL(u.String()))
	c.options = append(c.options, elastic.SetScheme(u.Scheme))

	log.Debugf("Using URL: %s with index: %s", u.String(), c.index)

	if username, ok := data["username"]; !ok {
	} else if password := data["password"]; !ok {
	} else {
		username := username.(string)
		password := password.(string)
		c.options = append(c.options, elastic.SetBasicAuth(username, password))

		log.Debugf("Using authentication with username: %s and password.", username)
	}

	c.InsecureSkipVerify = false

	if v, ok := data["insecure"]; !ok {
	} else if b, ok := v.(bool); !ok {
	} else {
		c.InsecureSkipVerify = b
	}

	tlsConfig.InsecureSkipVerify = c.InsecureSkipVerify

	c.Sniff = false

	if v, ok := data["sniff"]; !ok {
	} else if b, ok := v.(bool); !ok {
	} else {
		c.Sniff = b
	}

	c.options = append(c.options, elastic.SetSniff(c.Sniff))

	c.options = append(c.options, elastic.SetHttpClient(&http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
			TLSClientConfig:     tlsConfig,
		},
		Timeout: time.Duration(20) * time.Second,
	}))

	return nil
}
