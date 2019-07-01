package asn1_test

import (
	"log"
	"reflect"

	"github.com/PromonLogicalis/asn1"
)

func Example() {
	ctx := asn1.NewContext()

	// Use BER for encoding and decoding.
	ctx.SetDer(false, false)

	// Add a CHOICE
	ctx.AddChoice("value", []asn1.Choice{
		{
			Type:    reflect.TypeOf(""),
			Options: "tag:0",
		},
		{
			Type:    reflect.TypeOf(int(0)),
			Options: "tag:1",
		},
	})

	type Message struct {
		Id    int
		Value interface{} `asn1:"choice:value"`
	}

	// Encode
	msg := Message{
		Id:    1000,
		Value: "this is a value",
	}
	data, err := ctx.Encode(msg)
	if err != nil {
		log.Fatal(err)
	}

	// Decode
	decodedMsg := Message{}
	_, err = ctx.Decode(data, &decodedMsg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n", decodedMsg)
}
