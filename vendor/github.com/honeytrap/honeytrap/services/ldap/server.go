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
package ldap

import "crypto/tls"

//Server ldap server data
type Server struct {
	Handlers []requestHandler

	Credentials []string `toml:"credentials"`

	tlsConfig *tls.Config

	*DSE

	login string // username of logged in user
}

func (s Server) isLogin() bool {
	return s.login != ""
}

type DSE struct {
	NamingContexts       []string `toml:"naming-contexts"`
	SupportedLDAPVersion []string `toml:"supported-ldap-version"`
	SupportedExtension   []string `toml:"supported-extension"`
	VendorName           []string `toml:"vendor-name"`
	VendorVersion        []string `toml:"vendor-version"`
	Description          []string `toml:"description"`
	ObjectClass          []string `toml:"objectclass"`
}

// Get return the rootDSE as search result
func (d *DSE) Get() *SearchResultEntry {
	var dn string

	if len(d.NamingContexts) > 0 {
		dn = d.NamingContexts[0]
	}

	return &SearchResultEntry{
		DN: dn,
		Attrs: AttributeMap{
			"namingContexts":       d.NamingContexts,
			"supportedLDAPVersion": d.SupportedLDAPVersion,
			"supportedExtension":   d.SupportedExtension,
			"vendorName":           d.VendorName,
			"vendorVersion":        d.VendorVersion,
			"description":          d.Description,
			"objectClass":          d.ObjectClass,
		},
	}
}
