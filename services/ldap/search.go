/*
* Honeytrap
* Copyright (C) 2016-2018 DutchSec (https://dutchsec.com/)
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
package ldap

import (
	"fmt"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
)

var (
	//ErrNotASearchRequest error if it is not a search request
	ErrNotASearchRequest = fmt.Errorf("not a search request")

	//ErrSearchRequestTooComplex error if the search request is to complex
	ErrSearchRequestTooComplex = fmt.Errorf("search too complex to be parsed")
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
		el["ldap.search.basedn"] = ret.BaseDN
	}

	if len(rps) < 6 {
		return nil, ErrNotASearchRequest
	}

	ret.Scope = forceInt64(rps[1].Value)
	ret.DerefAliases = forceInt64(rps[2].Value)
	ret.SizeLimit = forceInt64(rps[3].Value)
	ret.TimeLimit = forceInt64(rps[4].Value)
	ret.TypesOnly = rps[5].Value.(bool)

	el["ldap.search-scope"] = ret.Scope
	el["ldap.search-derefaliases"] = ret.DerefAliases
	el["ldap.search-timelimit"] = ret.TimeLimit
	el["ldap.search-sizelimit"] = ret.SizeLimit

	// Check to see if it looks like a simple search criteria
	err = checkPacket(rps[6], ber.ClassContext, ber.TypeConstructed, 0x3)
	if err == nil {
		// It is simple, return the attribute and value
		ret.FilterAttr = string(rps[6].Children[0].ByteValue)
		ret.FilterValue = string(rps[6].Children[1].ByteValue)
	} else {
		// This is likely some sort of complex search criteria.
		// Try to generate a searchFingerPrint based on the values
		var getContextValue func(p *ber.Packet) string

		getContextValue = func(p *ber.Packet) string {
			var err error
			var sb strings.Builder

			if p.Value != nil {
				_, err = sb.WriteString(fmt.Sprint(p.Value))
			}

			for _, child := range p.Children {
				childVal := getContextValue(child)
				_, err = sb.WriteRune(',')
				_, err = sb.WriteString(childVal)
			}

			if err != nil {
				log.Debugf("ldap-search: writing search-fingerprint failed: %s", err)
				return ""
			}

			return sb.String()
		}

		var buf strings.Builder
		_, err = buf.WriteRune('\'')

		ret.FilterAttr = "#search-fingerprint"
		_, err = buf.WriteString(getContextValue(rps[6]))

		for index := 7; index < len(rps); index++ {
			if buf.Len() > 0 {
				_, err = buf.WriteRune(',')
			}
			_, err = buf.WriteString(getContextValue(rps[index]))
		}

		_, err = buf.WriteRune('\'')
		ret.FilterValue = buf.String()
	}

	el["ldap.search-filter"] = ret.FilterAttr
	el["ldap.search-filtervalue"] = ret.FilterValue

	return ret, nil
}

//SearchResultEntry a simplified ldap search response
type SearchResultEntry struct {
	DN    string                 // DN of this search result
	Attrs map[string]interface{} // map of attributes
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

		switch v := v.(type) {
		case string:
			attrvals.AppendChild(
				ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v, "String Value"))
		case []string:
			for _, v1 := range v {
				attrvals.AppendChild(
					ber.NewString(
						ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v1, "String Value"))
			}
		default:
			log.Debugf("skipping value for key '%s', I can't process type '%t'", k, v)
			continue
		}

		attr.AppendChild(attrvals)
		attrs.AppendChild(attr)

	}
	searchResult.AppendChild(attrs)

	replypacket.AppendChild(searchResult)

	return replypacket

}

func makeSearchResultDonePacket(msgid int64) *ber.Packet {

	replypacket := replyEnvelope(msgid)

	// tag 5 is SearchResultDone
	searchResult := ber.Encode(
		ber.ClassApplication, ber.TypeConstructed, ber.Tag(5), nil, "Response")
	searchResult.AppendChild(
		ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, "Result Code"))
	// per the spec these are "matchedDN" and "diagnosticMessage", but we don't need them for this
	searchResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "Unused"))
	searchResult.AppendChild(
		ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "Unused"))

	replypacket.AppendChild(searchResult)

	return replypacket

}

func makeSearchResultNoSuchObjectPacket(msgid int64) *ber.Packet {

	replypacket := replyEnvelope(msgid)

	searchResult := ber.Encode(
		ber.ClassApplication, ber.TypeConstructed, ber.Tag(5), nil, "Response")
	// 32 is "noSuchObject"
	searchResult.AppendChild(
		ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 32, "Result Code"))
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
	} else if err == ErrSearchRequestTooComplex {
		el["ldap.request-type"] = "search"
		return nil
	} else if err != nil {
		log.Debugf("Error while trying to parse search request: %v", err)
		return nil
	}

	el["ldap.request-type"] = "search"

	res := h.searchFunc(req)
	if res == nil {
		return nil
	}

	msgid, err := messageID(p)
	if err != nil {
		log.Debugf("Failed to extract message id")
		return nil
	}

	// no results
	if len(res) < 1 {
		return []*ber.Packet{makeSearchResultNoSuchObjectPacket(msgid)}
	}

	// format each result
	ret := make([]*ber.Packet, 0)
	for _, resitem := range res {
		resultPacket := resitem.makePacket(msgid)
		ret = append(ret, resultPacket)
	}

	// end with a done packet
	ret = append(ret, makeSearchResultDonePacket(msgid))

	return ret
}
