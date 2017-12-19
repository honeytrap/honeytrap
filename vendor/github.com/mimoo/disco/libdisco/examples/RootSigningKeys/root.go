package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {

	//
	// run `go run root.go gen` to generate the root key
	//
	if len(os.Args) == 2 && os.Args[1] == "gen" {
		// generating the root signing key
		if err := libdisco.GenerateAndSaveDiscoRootKeyPair("./privateRoot", "./publicRoot"); err != nil {
			panic("cannot generate and save a root key")
		}

		// loading the public part
		pubkey, err := libdisco.LoadDiscoRootPublicKey("./publicRoot")
		if err != nil {
			fmt.Println("cannot load the disco root pubkey")
			return
		}

		// displaying the public part
		fmt.Println("generated the root signing key successfuly. Public root key:")
		fmt.Println(hex.EncodeToString(pubkey))

		return
	}

	//
	// run `go run root.go sign hex_pubkey` to sign a public key
	//
	if len(os.Args) == 3 && os.Args[1] == "sign" {
		// what do we sign?
		toSign, err := hex.DecodeString(os.Args[2])
		if err != nil || len(toSign) != 32 {
			fmt.Println("public key passed is not a 32-byte value in hexadecimal (", len(toSign), ")")
			return
		}

		// load the private root key
		privkey, err := libdisco.LoadDiscoRootPrivateKey("./privateRoot")
		if err != nil {
			fmt.Println("couldn't load the private root key")
			return
		}

		// create proof
		proof := libdisco.CreateStaticPublicKeyProof(privkey, toSign)

		// display the proof
		fmt.Println("proof successfuly created:")
		fmt.Println(hex.EncodeToString(proof))

		return
	}

	// usage
	fmt.Println("read source code to find out usage")
	return

}
