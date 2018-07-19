package ldap

import ber "github.com/honeytrap/honeytrap/services/asn1-ber"

// LDAP Application Codes
const (
	ApplicationBindRequest           = 0
	ApplicationBindResponse          = 1
	ApplicationUnbindRequest         = 2
	ApplicationSearchRequest         = 3
	ApplicationSearchResultEntry     = 4
	ApplicationSearchResultDone      = 5
	ApplicationModifyRequest         = 6
	ApplicationModifyResponse        = 7
	ApplicationAddRequest            = 8
	ApplicationAddResponse           = 9
	ApplicationDelRequest            = 10
	ApplicationDelResponse           = 11
	ApplicationModifyDNRequest       = 12
	ApplicationModifyDNResponse      = 13
	ApplicationCompareRequest        = 14
	ApplicationCompareResponse       = 15
	ApplicationAbandonRequest        = 16
	ApplicationSearchResultReference = 19
	ApplicationExtendedRequest       = 23
	ApplicationExtendedResponse      = 24
)

var appCodes = map[int]string{
	ApplicationBindRequest:           "BindRequest",
	ApplicationBindResponse:          "BindResponse",
	ApplicationUnbindRequest:         "UnbindRequest",
	ApplicationSearchRequest:         "SearchRequest",
	ApplicationSearchResultEntry:     "SearchResultEntry",
	ApplicationSearchResultDone:      "SearchResultDone",
	ApplicationModifyRequest:         "ModifyRequest",
	ApplicationModifyResponse:        "ModifyResponse",
	ApplicationAddRequest:            "AddRequest",
	ApplicationAddResponse:           "AddResponse",
	ApplicationDelRequest:            "DelRequest",
	ApplicationDelResponse:           "DelResponse",
	ApplicationModifyDNRequest:       "ModifyDNRequest",
	ApplicationModifyDNResponse:      "ModifyDNResponse",
	ApplicationCompareRequest:        "CompareRequest",
	ApplicationCompareResponse:       "CompareResponse",
	ApplicationAbandonRequest:        "AbandonRequest",
	ApplicationSearchResultReference: "SearchResultReference",
	ApplicationExtendedRequest:       "ExtendedRequest",
	ApplicationExtendedResponse:      "ExtendedResponse",
}

// Resultcodes
const (
	ResultSuccess                  = 0
	ResultOperationsError          = 1
	ResultProtocolError            = 2
	ResultTimeLimitExceeded        = 3
	ResultSizeLimitExceeded        = 4
	ResultCompareFalse             = 5
	ResultCompareTrue              = 6
	ResultAuthMethodNotSupported   = 7
	ResultStrongerAuthRequired     = 8
	ResultReferral                 = 10
	ResultNoSuchAttribute          = 16
	ResultNoSuchObject             = 32
	ResultInvalidCredentials       = 49
	ResultInsufficientAccessRights = 50
	ResultEntryAlreadyExist        = 68
)

// Message defines an LDAP message
type Message struct {
	id         int
	protocolOp int
	log        map[string]interface{}
	conn       *Conn
	control    []Control
}

// Control defines an optional LDAP control
type Control struct {
	controlType int
	criticality bool
	value       string
}

var (
	asn1Eoc = ber.Encode(
		ber.ClassContext,
		ber.TypePrimitive,
		ber.TagEOC,
		nil,
		"EOC")
	asn1Succes = ber.NewInteger(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagEnumerated,
		ResultSuccess,
		"Succes")
)

// NewMessage creates a new LDAP message
func NewMessage(c *Conn) (*Message, error) {

	m := &Message{
		id:         int(c.packet.Children[0].Value.(int64)),
		protocolOp: int(c.packet.Children[1].Tag),
		log:        make(map[string]interface{}),
		conn:       c,
	}

	m.log["ldap.id"] = m.id
	m.log["ldap.operation"] = m.protocolOp

	if op, ok := appCodes[m.protocolOp]; ok {
		m.log["ldap.operation"] = op
	}

	switch m.protocolOp {

	case ApplicationBindRequest:
		m.log["ldap.version"] = p.Children[1].Children[0].Value
	}
	return m, nil
}

// Response returns an ASN.1 BER/DER encoded LDAP response packet
func (m *Message) Response(authState int) (*ber.Packet, error) {
	var (
		err error
		pc  *ber.Packet
	)

	p := m.envelope()

	switch m.protocolOp {

	case ApplicationBindRequest:
		pc, err = m.simpleBind()
		p.AppendChild(pc)
	case ApplicationAddRequest:
		fallthrough
	case ApplicationSearchRequest:
		fallthrough
	case ApplicationModifyRequest:
		fallthrough
	case ApplicationDelRequest:
		fallthrough
	case ApplicationModifyDNRequest:
		fallthrough
	case ApplicationCompareRequest:
		fallthrough
	case ApplicationExtendedRequest:
		fallthrough
	default:
		// Just send a succes response
		// This is not bulletproof (protocolOp+1 is not always the response code)
		tag := ber.Tag(m.protocolOp + 1)
		p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, tag, nil, "Response")
		p.AppendChild(asn1Succes)
		p.AppendChild(ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", ""))
		p.AppendChild(asn1Eoc)
	}

	if err != nil {
		return nil, err
	}

	return p, nil
}

//Create the LDAP message envelope with the correct ID
func (m *Message) envelope() *ber.Packet {

	p := ber.NewSequence("LDAP Response")
	p.AppendChild(ber.NewInteger(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagInteger,
		m.id,
		"MessageID",
	))

	return p
}

func (m *Message) simpleBind() (*ber.Packet, error) {

	// Check Auth

	p := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationBindResponse, nil, "Bind Response")
	p.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, "Succes"))
	p.AppendChild(ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", ""))
	p.AppendChild(ber.Encode(ber.ClassContext, ber.TypePrimitive, ber.TagEOC, nil, "EOC"))

	return p, nil
}
