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

import (
	"errors"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
)

var (
	//ErrNotASearchRequest error if it is not a search request
	ErrNotASearchRequest = errors.New("not a search request")

	mapScope = map[int64]string{
		0: "baseObject",
		1: "singleLevel",
		2: "wholeSubtree",
	}

	mapDerefAliases = map[int64]string{
		0: "never",
		1: "inSearching",
		2: "FindingBaseObj",
		3: "always",
	}
)

//SearchRequest a simplified ldap search request
type SearchRequest struct {
	Packet       *ber.Packet
	BaseDN       string // DN under which to start searching
	Scope        int64  // baseObject(0), singleLevel(1), wholeSubtree(2)
	DerefAliases int64  // neverDerefAliases(0),derefInSearching(1),derefFindingBaseObj(2),derefAlways(3)
	SizeLimit    int64  // max number of results to return
	TimeLimit    int64  // max time in seconds to spend processing
	TypesOnly    bool   // if true client is expecting only type info
	FilterAttr   string // filter attribute name (assumed to be an equality match with just this one attribute)
	FilterValue  string // filter attribute value
}

func parseSearchRequest(p *ber.Packet, el eventLog) (*SearchRequest, error) {

	ret := &SearchRequest{}

	if len(p.Children) < 2 {
		return nil, ErrNotASearchRequest
	}

	err := checkPacket(p.Children[1], ber.ClassApplication, ber.TypeConstructed, 0x3)
	if err != nil {
		return nil, ErrNotASearchRequest
	}

	rps := p.Children[1].Children

	if len(rps) > 0 {
		ret.BaseDN = string(rps[0].ByteValue)
		el["ldap.search-basedn"] = ret.BaseDN
	}
	if len(rps) > 5 {
		ret.Scope = forceInt64(rps[1].Value)
		ret.DerefAliases = forceInt64(rps[2].Value)
		ret.SizeLimit = forceInt64(rps[3].Value)
		ret.TimeLimit = forceInt64(rps[4].Value)
		ret.TypesOnly = rps[5].Value.(bool)
	}

	// is this a present filter like (objectClass=*)
	if err := checkPacket(rps[6], ber.ClassContext, ber.TypePrimitive, 0x7); err == nil {
		ret.FilterAttr = string(rps[6].ByteValue)
		if len(rps[7].Children) == 0 {
			ret.FilterValue = "*"
		}
	} else if err := checkPacket(rps[6], ber.ClassContext, ber.TypeConstructed, 0x3); err == nil {
		// It is simple, return the attribute and value
		ret.FilterAttr = string(rps[6].Children[0].ByteValue)
		ret.FilterValue = string(rps[6].Children[1].ByteValue)
	} else {
		// This is likely some sort of complex search criteria.
		// Try to generate a searchFingerPrint based on the values
		ret.FilterAttr = "#search-fingerprint"

		var getContextValue func(p *ber.Packet) string

		getContextValue = func(p *ber.Packet) string {

			var sb strings.Builder

			if len(p.ByteValue) > 0 {
				_, err = sb.Write(p.ByteValue)
			}

			for _, child := range p.Children {
				childVal := getContextValue(child)
				if sb.Len() > 0 && len(childVal) > 0 {
					_, err = sb.WriteRune(',')
				}
				_, err = sb.WriteString(childVal)
			}

			if err != nil {
				log.Debugf("ldap-search: writing search-fingerprint failed: %s", err)
				return ""
			}
			return sb.String()
		}

		var buf strings.Builder

		for index := 6; index < len(rps); index++ {

			str := getContextValue(rps[index])

			if buf.Len() > 0 {
				_, err = buf.WriteRune(',')
			}

			if str != "" {
				_, err = buf.WriteString(str)
			}

			if err != nil {
				log.Debugf("ldap-search: writing search-fingerprint failed: %s", err)
			}
		}
		ret.FilterValue = buf.String()
	}

	// set event values
	el["ldap.search-basedn"] = ret.BaseDN
	el["ldap.search-filter"] = ret.FilterAttr
	el["ldap.search-filtervalue"] = ret.FilterValue
	el["ldap.search-timelimit"] = ret.TimeLimit
	el["ldap.search-sizelimit"] = ret.SizeLimit

	if s, ok := mapScope[ret.Scope]; ok {
		el["ldap.search-scope"] = s
	} else {
		el["ldap.search-scope"] = ret.Scope
	}

	if s, ok := mapDerefAliases[ret.DerefAliases]; ok {
		el["ldap.search-derefaliases"] = s
	} else {
		el["ldap.search-derefaliases"] = ret.DerefAliases
	}

	return ret, nil
}

// AttributeMap holding attributes for a search result entry
type AttributeMap map[string][]string

//SearchResultEntry a simplified ldap search response
type SearchResultEntry struct {
	DN    string // DN of this search result
	Attrs AttributeMap
}

func (e *SearchResultEntry) makePacket(msgid int64) *ber.Packet {

	replypacket := replyEnvelope(msgid)

	searchResult := ber.Encode(
		ber.ClassApplication, ber.TypeConstructed, ber.Tag(4), nil, "Response")
	searchResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, e.DN, "DN"))
	attrs := ber.Encode(
		ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attrs")

	for k, v := range e.Attrs {

		attr := ber.Encode(
			ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attr")
		attr.AppendChild(
			ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, k, "Key"))
		attrvals := ber.Encode(
			ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "Values")

		for _, v1 := range v {
			attrvals.AppendChild(
				ber.NewString(
					ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v1, "String Value"))
		}

		attr.AppendChild(attrvals)
		attrs.AppendChild(attr)

	}
	searchResult.AppendChild(attrs)

	replypacket.AppendChild(searchResult)

	return replypacket
}

func makeSearchResultDonePacket(msgid int64, resultcode int) *ber.Packet {

	replypacket := replyEnvelope(msgid)

	// tag 5 is SearchResultDone
	searchResult := ber.Encode(
		ber.ClassApplication, ber.TypeConstructed, ber.Tag(5), nil, "Response")
	searchResult.AppendChild(
		ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, resultcode, "Result Code"))
	// per the spec these are "matchedDN" and "diagnosticMessage", but we don't need them for this
	searchResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "Unused"))
	searchResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "Unused"))

	replypacket.AppendChild(searchResult)

	return replypacket

}

// a callback function to produce search results; should return nil to mean
// we chose not to attempt to search (i.e. this request is not for us);
// or return empty slice to mean 0 results (or slice with data for results)
type searchFunc func(*SearchRequest) []*SearchResultEntry

type searchFuncHandler struct {
	searchFunc searchFunc
}

func (h *searchFuncHandler) handle(p *ber.Packet, el eventLog) []*ber.Packet {

	req, err := parseSearchRequest(p, el)
	if err == ErrNotASearchRequest {
		return nil
	} else if err != nil {
		log.Debugf("Error while trying to parse search request: %v", err)
		return nil
	}

	el["ldap.request-type"] = "search"

	msgid, err := messageID(p)
	if err != nil {
		log.Debugf("Failed to extract message id: %s", err)
		return nil
	}

	// do callback
	res := h.searchFunc(req)

	// no results
	if len(res) < 1 {
		return []*ber.Packet{makeSearchResultDonePacket(msgid, ResNoSuchObject)}
	}

	// format each result
	ret := make([]*ber.Packet, 0)
	for _, resitem := range res {
		resultPacket := resitem.makePacket(msgid)
		ret = append(ret, resultPacket)
	}

	// end with a done packet
	ret = append(ret, makeSearchResultDonePacket(msgid, ResSuccess))

	return ret
}
