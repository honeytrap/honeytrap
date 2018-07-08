package snmp

import (
	"fmt"
	"reflect"

	"github.com/Logicalis/asn1"
)

// SNMP error codes.
const (
	NoError             = 0
	TooBig              = 1
	NoSuchName          = 2
	BadValue            = 3
	ReadOnly            = 4
	GenErr              = 5
	NoAccess            = 6
	WrongType           = 7
	WrongLength         = 8
	WrongEncoding       = 9
	WrongValue          = 10
	NoCreation          = 11
	InconsistentValue   = 12
	ResourceUnavailable = 13
	CommitFailed        = 14
	UndoFailed          = 15
	AuthorizationError  = 16
	NotWritable         = 17
	InconsistentName    = 18
)

// Message is the top level element of the SNMP protocol.
type Message struct {
	Version   int
	Community string
	Pdu       interface{} `asn1:"choice:pdu"`
}

// Pdu is a generic type for other Protocol Data Units.
type Pdu struct {
	Identifier  int
	ErrorStatus int
	ErrorIndex  int
	Variables   []Variable
}

// BulkPdu is a generic type for other Protocol Data Units.
type BulkPdu struct {
	Identifier     int
	NonRepeaters   int
	MaxRepetitions int
	Variables      []Variable
}

// GetRequestPdu is used to request data.
type GetRequestPdu Pdu

// GetNextRequestPdu works similarly to GetRequestPdu, but it's returned the
// value for the next valid Oid.
type GetNextRequestPdu Pdu

// GetResponsePdu is used in responses to SNMP requests.
type GetResponsePdu Pdu

// SetRequestPdu is used to request data to be updated.
type SetRequestPdu Pdu

// V1TrapPdu is used when sending a trap in SNMPv1.
type V1TrapPdu struct {
	Enterprise   asn1.Oid
	AgentAddr    IPAddress
	GenericTrap  int
	SpecificTrap int
	Timestamp    TimeTicks
	Variables    []Variable
}

// GetBulkRequestPdu is used for bulk requests.
type GetBulkRequestPdu BulkPdu

// InformRequestPdu is used for inform requests.
type InformRequestPdu Pdu

// V2TrapPdu is used when sending a trap in SNMPv2.
type V2TrapPdu Pdu

// Variable represents an entry of the variable bindings
type Variable struct {
	Name  asn1.Oid
	Value interface{} `asn1:"choice:val"`
}

// Types available for Variable.Value

// IPAddress is a IPv4 address.
type IPAddress [4]byte

// String returns a representation of IPAddress in dot notation.
func (ip IPAddress) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

// Counter32 is a counter type.
type Counter32 uint32

// Unsigned32 is an integer type.
type Unsigned32 uint32

// TimeTicks is a type for time.
type TimeTicks uint32

// Opaque is a type for blobs.
type Opaque []byte

// Counter64 is a counter type.
type Counter64 uint64

// Exceptions available for Variable.Value

// NoSuchObject exception.
type NoSuchObject asn1.Null

func (e NoSuchObject) String() string { return "NoSuchObject" }

// NoSuchInstance exception.
type NoSuchInstance asn1.Null

func (e NoSuchInstance) String() string { return "NoSuchInstance" }

// EndOfMibView exception.
type EndOfMibView asn1.Null

func (e EndOfMibView) String() string { return "EndOfMibView" }

// Asn1Context returns a new allocated asn1.Context and registers all the
// choice types necessary for SNMPv1 and SNMPv2.
func Asn1Context() *asn1.Context {
	ctx := asn1.NewContext()
	ctx.AddChoice("pdu", []asn1.Choice{
		{
			Type:    reflect.TypeOf(GetRequestPdu{}),
			Options: "tag:0",
		},
		{
			Type:    reflect.TypeOf(GetNextRequestPdu{}),
			Options: "tag:1",
		},
		{
			Type:    reflect.TypeOf(GetResponsePdu{}),
			Options: "tag:2",
		},
		{
			Type:    reflect.TypeOf(SetRequestPdu{}),
			Options: "tag:3",
		},
		{
			Type:    reflect.TypeOf(V1TrapPdu{}),
			Options: "tag:4",
		},
		{
			Type:    reflect.TypeOf(GetBulkRequestPdu{}),
			Options: "tag:5",
		},
		{
			Type:    reflect.TypeOf(InformRequestPdu{}),
			Options: "tag:6",
		},
		{
			Type:    reflect.TypeOf(V2TrapPdu{}),
			Options: "tag:7",
		},
	})
	ctx.AddChoice("val", []asn1.Choice{
		// Simple syntax
		{
			Type: reflect.TypeOf(asn1.Null{}),
		},
		{
			Type: reflect.TypeOf(int(0)),
		},
		{
			Type: reflect.TypeOf(""),
		},
		{
			Type: reflect.TypeOf(asn1.Oid{}),
		},
		// Application wide
		{
			Type:    reflect.TypeOf(IPAddress{}),
			Options: "application,tag:0",
		},
		{
			Type:    reflect.TypeOf(Counter32(0)),
			Options: "application,tag:1",
		},
		{
			Type:    reflect.TypeOf(Unsigned32(0)),
			Options: "application,tag:2",
		},
		{
			Type:    reflect.TypeOf(TimeTicks(0)),
			Options: "application,tag:3",
		},
		{
			Type:    reflect.TypeOf(Opaque("")),
			Options: "application,tag:4",
		},
		// [APPLICATION 5] does not exist.
		{
			Type:    reflect.TypeOf(Counter64(0)),
			Options: "application,tag:6",
		},
		// Exceptions
		{
			Type:    reflect.TypeOf(NoSuchObject{}),
			Options: "tag:0",
		},
		{
			Type:    reflect.TypeOf(NoSuchInstance{}),
			Options: "tag:1",
		},
		{
			Type:    reflect.TypeOf(EndOfMibView{}),
			Options: "tag:2",
		},
	})
	return ctx
}
