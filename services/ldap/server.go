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
