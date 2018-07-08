package asn1

import (
	"math/big"
	"reflect"
	"testing"
)

// isBytesEqual compares two byte arrays/slices.
func isBytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// testCase represents a decoding/encoding test case with a target object and
// a expected sequence of bytes.
type testCase struct {
	value    interface{}
	expected []byte
}

// testEncode encodes an object and compares with the expected bytes.
func testEncode(t *testing.T, ctx *Context, options string, tests ...testCase) {
	for _, test := range tests {
		data, err := ctx.EncodeWithOptions(test.value, options)
		if err != nil {
			t.Fatal(err)
		}
		if !isBytesEqual(data, test.expected) {
			t.Fatalf("Failed to encode \"%v\".\n Expected: %#v.\n Got:      %#v",
				test.value, test.expected, data)
		}
	}
}

func checkEqual(t *testing.T, obj1 interface{}, obj2 interface{}) {
	equal := false
	switch val1 := obj1.(type) {
	case *big.Int:
		val2, ok := obj2.(*big.Int)
		equal = ok && val1.Cmp(val2) == 0
	default:
		equal = reflect.DeepEqual(obj1, obj2)
	}
	if !equal {
		t.Fatalf("Decoded value does not match.\n Got \"%v\" (%T)\n When decoding \"%v\" (%T)",
			obj1, obj1, obj2, obj2)
	}
}

// testEncode decodes a sequence of bytes and compares with the target object.
func testDecode(t *testing.T, ctx *Context, options string, tests ...testCase) {
	for _, test := range tests {
		value := reflect.New(reflect.TypeOf(test.value))
		rest, err := ctx.DecodeWithOptions(test.expected, value.Interface(), options)
		if err != nil {
			t.Fatal(err)
		}
		if len(rest) > 0 {
			t.Fatalf("Unexpected remaining bytes when decoding \"%v\": %#v\n",
				test.value, rest)
		}
		checkEqual(t, value.Elem().Interface(), test.value)
	}
}

// testEncodeDecode does testEncode and testDecode.
func testEncodeDecode(t *testing.T, ctx *Context, options string, tests ...testCase) {
	for _, test := range tests {
		testEncode(t, ctx, options, test)
		testDecode(t, ctx, options, test)
	}
}

// testSimple encodes an object, decodes the resulting bytes and then compared the
// two objects.
func testSimple(t *testing.T, ctx *Context, options string, objs ...interface{}) {
	for _, obj := range objs {
		data, err := ctx.EncodeWithOptions(obj, options)
		if err != nil {
			t.Fatal(err)
		}
		value := reflect.New(reflect.TypeOf(obj))
		rest, err := ctx.DecodeWithOptions(data, value.Interface(), options)
		if err != nil {
			t.Fatal(err)
		}
		if len(rest) > 0 {
			t.Fatalf("Unexpected remaining bytes when decoding \"%v\": %#v\n",
				obj, rest)
		}
		checkEqual(t, obj, value.Elem().Interface())
	}
}

func TestContext(t *testing.T) {
	target := 0
	// Without options
	data, err := Encode(0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decode(data, &target)
	if err != nil {
		t.Fatal(err)
	}
	// With options
	data, err = EncodeWithOptions(0, "tag:1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecodeWithOptions(data, &target, "tag:1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleBool(t *testing.T) {
	ctx := NewContext()
	// Simple test
	testEncodeDecode(t, ctx, "", testCase{false, []byte{0x01, 0x01, 0x00}})
	testEncodeDecode(t, ctx, "", testCase{true, []byte{0x01, 0x01, 0xff}})
	testDecode(t, ctx, "", testCase{true, []byte{0x01, 0x01, 0x01}})
	// Der test
	ctx.SetDer(true, true)
	testEncodeDecode(t, ctx, "", testCase{false, []byte{0x01, 0x01, 0x00}})
	testEncodeDecode(t, ctx, "", testCase{true, []byte{0x01, 0x01, 0xff}})
	// Test Der invalid boolean
	boolean := false
	_, err := ctx.Decode([]byte{0x01, 0x01, 0x01}, &boolean)
	if err == nil {
		t.Fatal("DER boolean should accept only the simplest form.")
	} else {
		if _, ok := err.(*ParseError); !ok {
			t.Fatal(err)
		}
	}
}

func TestSimpleInteger(t *testing.T) {
	tests := []testCase{
		// int
		{0, []byte{0x02, 0x01, 0x00}},
		{1, []byte{0x02, 0x01, 0x01}},
		{255, []byte{0x02, 0x02, 0x00, 0xff}},
		{-1, []byte{0x02, 0x01, 0xff}},
		{1000, []byte{0x02, 0x02, 0x03, 0xe8}},
		{-1000, []byte{0x02, 0x02, 0xfc, 0x18}},
		// uint
		{uint(0), []byte{0x02, 0x01, 0x00}},
		{uint(1), []byte{0x02, 0x01, 0x01}},
		{uint(127), []byte{0x02, 0x01, 0x7f}},
		{uint(256 - 1), []byte{0x02, 0x02, 0x00, 0xff}},
		{uint(256*256 - 1), []byte{0x02, 0x03, 0x00, 0xff, 0xff}},
		// big.Int
		{big.NewInt(0), []byte{0x02, 0x01, 0x00}},
		{big.NewInt(1), []byte{0x02, 0x01, 0x01}},
		{big.NewInt(-1), []byte{0x02, 0x01, 0xff}},
		{
			big.NewInt(0).SetBit(big.NewInt(0), 128, 1),
			[]byte{
				0x02, 0x11, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00,
			},
		},
		{
			big.NewInt(0).Neg(big.NewInt(0).SetBit(big.NewInt(0), 128, 1)),
			[]byte{
				0x02, 0x11, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00,
			},
		},
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "", tests...)
}

func TestSimpleString(t *testing.T) {
	lenStr := func(length int, lenBytes ...byte) testCase {
		buf := make([]byte, length)
		return testCase{
			string(buf),
			append(append([]byte{0x04}, lenBytes...), buf...),
		}
	}
	tests := []testCase{
		// int
		{"", []byte{0x04, 0x00}},
		{"a", []byte{0x04, 0x01, 0x61}},
		{"abc", []byte{0x04, 0x03, 0x61, 0x62, 0x63}},
		lenStr(127, 0x7f),
		lenStr(255, 0x81, 0xff),
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "", tests...)
}

func TestUtf8String(t *testing.T) {
	ctx := NewContext()
	testEncodeDecode(t, ctx, "universal,tag:12", testCase{
		"test",
		[]byte{0x0c, 0x04, 0x74, 0x65, 0x73, 0x74},
	})
}

func TestOctetString(t *testing.T) {
	tests := []testCase{
		{
			[]byte{},
			[]byte{0x04, 0x00},
		},
		{
			[]byte{0x00},
			[]byte{0x04, 0x01, 0x00},
		},
		{
			[]byte{0x01, 0x02, 0x03},
			[]byte{0x04, 0x03, 0x01, 0x02, 0x03},
		},
		{
			[...]byte{},
			[]byte{0x04, 0x00},
		},
		{
			[...]byte{0x00},
			[]byte{0x04, 0x01, 0x00},
		},
		{
			[...]byte{0x01, 0x02, 0x03},
			[]byte{0x04, 0x03, 0x01, 0x02, 0x03},
		},
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "", tests...)
	// Tests that should fail
	tests = []testCase{
		{
			[0]byte{},
			[]byte{0x04, 0x01, 0x00},
		},
		{
			[2]byte{},
			[]byte{0x04, 0x01, 0x00},
		},
	}
	for _, test := range tests {
		_, err := ctx.Decode(test.expected, test)
		if err == nil {
			t.Fatal("OctetString with length different from array should have failed.")
		}
	}
}

func TestSimpleOid(t *testing.T) {
	// Cases that encoding and decoding do not match
	tests := []testCase{
		{Oid{}, []byte{0x06, 0x01, 0x00}},
		{Oid{0}, []byte{0x06, 0x01, 0x00}},
		{Oid{1}, []byte{0x06, 0x01, 0x28}},
		{Oid{2}, []byte{0x06, 0x01, 0x50}},
	}
	ctx := NewContext()
	testEncode(t, ctx, "", tests...)

	// Cases that de/encoding match:
	tests = []testCase{
		{Oid{0, 0}, []byte{0x06, 0x01, 0x00}},
		{Oid{0, 39}, []byte{0x06, 0x01, 0x27}},
		{Oid{1, 0}, []byte{0x06, 0x01, 0x28}},
		{Oid{1, 39}, []byte{0x06, 0x01, 0x4f}},
		{Oid{2, 0}, []byte{0x06, 0x01, 0x50}},
		{Oid{2, 39}, []byte{0x06, 0x01, 0x77}},
		{Oid{0, 0, 0}, []byte{0x06, 0x02, 0x00, 0x00}},
		{Oid{0, 0, 1}, []byte{0x06, 0x02, 0x00, 0x01}},
		{Oid{0, 0, 127}, []byte{0x06, 0x02, 0x00, 0x07f}},
		{Oid{0, 0, 128}, []byte{0x06, 0x03, 0x00, 0x81, 0x00}},
		{Oid{0, 0, 255}, []byte{0x06, 0x03, 0x00, 0x81, 0x7f}},
		{Oid{0, 0, 1000}, []byte{0x06, 0x03, 0x00, 0x87, 0x68}},
		{Oid{0, 0, ^uint(0)}, []byte{0x06, 0x0b, 0x00, 0x81, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
	}
	testEncodeDecode(t, ctx, "", tests...)

	// Cases that should fail
	oids := []Oid{
		{3, 0}, {4, 0}, {5, 0}, {256, 0},
		{0, 40}, {0, 41}, {0, 256},
	}
	for _, oid := range oids {
		_, err := ctx.Encode(oid)
		if err == nil {
			t.Fatalf("OID %v is invalid and encoding should have failed.", oid)
		}
	}
}

func TestCmpOid(t *testing.T) {
	tests := []struct {
		s    int
		a, b Oid
	}{
		{1, Oid{1, 1, 10, 10}, Oid{0}},
		{1, Oid{1, 1, 10, 10}, Oid{0, 1}},
		{1, Oid{1, 1, 10, 10}, Oid{0, 0, 0, 0, 0, 0, 0}},
		{1, Oid{1, 1, 10, 10}, Oid{0, 10}},
		{1, Oid{1, 1, 10, 10}, Oid{1, 0}},
		{1, Oid{1, 1, 10, 10}, Oid{1, 1}},
		{1, Oid{1, 1, 10, 10}, Oid{1, 1, 1}},
		{1, Oid{1, 1, 10, 10}, Oid{1, 1, 10}},
		{1, Oid{1, 1, 10, 10}, Oid{1, 1, 10, 1}},
		{1, Oid{1, 1, 10, 10}, Oid{1, 1, 10, 9}},
		{0, Oid{1, 1, 10, 10}, Oid{1, 1, 10, 10}},
		{-1, Oid{1, 1, 10, 10}, Oid{1, 1, 10, 11}},
		{-1, Oid{1, 1, 10, 10}, Oid{1, 1, 10, 10, 0}},
		{-1, Oid{1, 1, 10, 10}, Oid{1, 1, 11}},
		{-1, Oid{1, 1, 10, 10}, Oid{1, 1, 11, 0}},
		{-1, Oid{1, 1, 10, 10}, Oid{2}},
	}
	op := map[int]string{1: ">", 0: "=", -1: "<"}
	for _, test := range tests {
		cmp := test.a.Cmp(test.b)
		if cmp > 0 {
			cmp = 1
		} else if cmp < 0 {
			cmp = -1
		}
		if cmp != test.s {
			t.Fatalf("Wrong comparison.\n\tExpected %s %s %s\n\tGot: %s %s %s",
				test.a, op[test.s], test.b, test.a, op[cmp], test.b)
		}
	}
}

func TestSimpleNull(t *testing.T) {
	tests := []testCase{
		{Null{}, []byte{0x05, 0x00}},
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "", tests...)
}

func TestSequence(t *testing.T) {
	type Type struct {
		A int
		B string
		C bool
	}
	obj := Type{1, "abc", true}
	ctx := NewContext()
	testSimple(t, ctx, "", obj)
}

func TestSet(t *testing.T) {
	type Type struct {
		A int
		B string
		C bool
	}
	obj := Type{1, "abc", true}
	ctx := NewContext()
	testSimple(t, ctx, "set", obj)
}

func TestBerSet(t *testing.T) {
	type Type struct {
		A int
		B string
		C bool
	}
	elem := Type{1, "abc", true}
	ctx := NewContext()
	data, err := ctx.EncodeWithOptions(elem, "set")
	if err != nil {
		t.Fatal(err)
	}
	type RevType struct {
		C bool
		B string
		A int
	}
	revElem := RevType{}
	_, err = ctx.DecodeWithOptions(data, &revElem, "set")
	if err != nil {
		t.Fatal(err)
	}
	if elem.A != revElem.A ||
		elem.B != revElem.B ||
		elem.C != revElem.C {
		t.Fatalf("Sets does not match:\n %#v\n %#v", elem, revElem)
	}
}

func TestDerSet(t *testing.T) {
	// Thos test uses BER for encoding and DER for decoding. When the library
	// encodes SETs in BER mode it keeps the order of the fields since it allows
	// SET elements to appear in any order. However DER forces SETs elements to be
	// encoded in ascending order of their tag numbers.
	ctx := NewContext()
	ctx.SetDer(false, true)

	type Type struct {
		A int    // tag = 2
		B string // tag = 3
		C bool   // tag = 1
	}

	// Encode with BER.
	elem := Type{A: 1, B: "abc", C: true}
	data, err := ctx.EncodeWithOptions(elem, "set")
	if err != nil {
		t.Fatal(err)
	}

	// Encode with DER. Should return an error.
	elem = Type{}
	_, err = ctx.DecodeWithOptions(data, &elem, "set")
	if err == nil {
		t.Fatal("Der decoding of a SET with non sorted element should have failed.")
	}
	if _, ok := err.(*ParseError); !ok {
		t.Fatal("Unexpected error:", err)
	}
}

func TestOptional(t *testing.T) {
	type Type struct {
		A int `asn1:"optional"`
		B string
		C bool
	}
	test := testCase{
		Type{0, "abc", true},
		[]byte{
			// SEQ LEN=8
			0x30, 0x08,
			// OCTETSTRING LEN=3
			0x04, 0x03,
			// "abc"
			0x61, 0x62, 0x63,
			// BOOLEAN LEN=1
			0x01, 0x01,
			// true
			0xff,
		},
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "", test)
}

func TestEncodeDefaultBer(t *testing.T) {
	type Type struct {
		A1 int  `asn1:"default:-1"`
		A2 uint `asn1:"default:127"`
		B  string
		C  bool
	}
	test := testCase{
		Type{0, 0, "abc", true},
		[]byte{
			// SEQ LEN=14
			0x30, 0x0e,
			// INTEGER LEN=1
			0x02, 0x01,
			// -1
			0xff,
			// INTEGER LEN=1
			0x02, 0x01,
			// 127
			0x7f,
			// OCTETSTRING LEN=3
			0x04, 0x03,
			// "abc"
			0x61, 0x62, 0x63,
			// BOOLEAN LEN=1
			0x01, 0x01,
			// true
			0xff,
		},
	}
	ctx := NewContext()
	ctx.SetDer(false, false)
	testEncode(t, ctx, "", test)
}

func TestEncodeDefaultDer(t *testing.T) {
	type Type struct {
		A int `asn1:"default:127"`
		B string
		C bool
	}
	test := testCase{
		Type{0, "abc", true},
		[]byte{
			// SEQ LEN=8
			0x30, 0x08,
			// OCTETSTRING LEN=3
			0x04, 0x03,
			// "abc"
			0x61, 0x62, 0x63,
			// BOOLEAN LEN=1
			0x01, 0x01,
			// true
			0xff,
		},
	}
	ctx := NewContext()
	ctx.SetDer(true, true)
	testEncode(t, ctx, "", test)
}

func TestTag(t *testing.T) {
	ctx := NewContext()
	testEncodeDecode(t, ctx, "tag:1", testCase{
		false,
		[]byte{0x81, 0x01, 0x00},
	})
	testEncodeDecode(t, ctx, "explicit,tag:1", testCase{
		true,
		[]byte{0xa1, 0x03, 0x01, 0x01, 0xff},
	})
	testEncodeDecode(t, ctx, "explicit,application,tag:1", testCase{
		false,
		[]byte{0x61, 0x03, 0x01, 0x01, 0x00},
	})
	testEncodeDecode(t, ctx, "explicit,tag:1000", testCase{
		true,
		[]byte{0xbf, 0x87, 0x68, 0x3, 0x1, 0x1, 0xff},
	})

}

func TestIndefinite(t *testing.T) {
	type Type struct {
		Flag bool
		Num  int
	}
	type NestedType struct {
		Nested Type `asn1:"indefinite"`
	}
	testCases := []testCase{
		{
			Type{true, 0},
			[]byte{
				0x30, 0x80,
				0x01, 0x01, 0xff,
				0x02, 0x01, 0x00,
				0x00, 0x00,
			},
		},
		{
			NestedType{Type{true, 0}},
			[]byte{
				0x30, 0x80,
				0x30, 0x80,
				0x01, 0x01, 0xff,
				0x02, 0x01, 0x00,
				0x00, 0x00,
				0x00, 0x00,
			},
		},
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "indefinite", testCases...)
}

func TestChoice(t *testing.T) {

	type Type struct {
		Num int
		Msg interface{} `asn1:"choice:msg"`
	}

	ctx := NewContext()
	ctx.AddChoice("msg", []Choice{
		{reflect.TypeOf(""), "tag:0"},
		{reflect.TypeOf(int(0)), "tag:1"},
	})

	choice := "abc"
	obj := Type{1, choice}
	data, err := ctx.Encode(obj)
	if err != nil {
		t.Fatal(err)
	}

	decodedObj := Type{}
	_, err = ctx.Decode(data, &decodedObj)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decodedObj.Msg, choice) {
		t.Fatalf("Incorrect choice value.\n Expected: %v\t Got: %v", choice, decodedObj.Msg)
	}
}

func TestExplicitChoice(t *testing.T) {
	ctx := NewContext()
	ctx.AddChoice("msg", []Choice{
		{reflect.TypeOf(int(0)), "explicit,tag:0"},
		{reflect.TypeOf(""), "explicit,tag:1"},
	})

	obj := "abc"
	data, err := ctx.EncodeWithOptions(obj, "choice:msg")
	if err != nil {
		t.Fatal(err)
	}

	var decodedObj interface{}
	_, err = ctx.DecodeWithOptions(data, &decodedObj, "choice:msg")
	if err != nil {
		t.Fatal("Error:", err)
	}

	if !reflect.DeepEqual(decodedObj, obj) {
		t.Fatalf("Incorrect choice value.\n Expected: %v\t Got: %v", obj, decodedObj)
	}
}

func TestNestedExplicitChoice(t *testing.T) {
	type Type struct {
		Num int
		Msg interface{} `asn1:"choice:msg"`
	}

	ctx := NewContext()
	ctx.AddChoice("msg", []Choice{
		{reflect.TypeOf(int(0)), "explicit,tag:0"},
		{reflect.TypeOf(""), "explicit,tag:1"},
	})

	choice := "abc"
	obj := Type{1, choice}
	data, err := ctx.Encode(obj)
	if err != nil {
		t.Fatal(err)
	}

	decodedObj := Type{}
	_, err = ctx.Decode(data, &decodedObj)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decodedObj.Msg, choice) {
		t.Fatalf("Incorrect choice value.\n Expected: %v\t Got: %v", choice, decodedObj.Msg)
	}
}

func TestArraySlice(t *testing.T) {
	testCases := []testCase{
		{
			[]int{0, 1, 2},
			[]byte{0x30, 0x09, 0x02, 0x01, 0x00, 0x02, 0x01, 0x01, 0x02, 0x01, 0x02},
		},
		{
			[...]int{0, 1, 2},
			[]byte{0x30, 0x09, 0x02, 0x01, 0x00, 0x02, 0x01, 0x01, 0x02, 0x01, 0x02},
		},
	}
	ctx := NewContext()
	testEncodeDecode(t, ctx, "", testCases...)
}
