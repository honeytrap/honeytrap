package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {

	//
	// run `go run client.go gen` to generate the static key of the client
	//
	if len(os.Args) == 2 && os.Args[1] == "setup" {

		// generating the client's keypair
		clientKeyPair, err := libdisco.GenerateAndSaveDiscoKeyPair("./clientKeyPair")
		if err != nil {
			panic("couldn't generate and save the client's key pair")
		}

		// displaying the public part
		fmt.Println("generated the client static public key successfuly. Client's public key:")
		fmt.Println(hex.EncodeToString(clientKeyPair.PublicKey[:]))

		return
	}

	//
	// run `go run client.go connect hex_proof hex_server_static_public_key` to connect to a server
	//
	if len(os.Args) == 4 && os.Args[1] == "connect" {

		// load the client's keypair
		clientKeyPair, err := libdisco.LoadDiscoKeyPair("./clientkeyPair")
		if err != nil {
			fmt.Println("couldn't load the client's key pair")
			return
		}

		// retrieve server's static public key
		serverPublicKey, err := hex.DecodeString(os.Args[3])
		if err != nil || len(serverPublicKey) != 32 {
			fmt.Println("server's static public key passed is not a 32-byte value in hexadecimal (", len(serverPublicKey), ")")
			return
		}

		// retrieve signature/proof
		proof, err := hex.DecodeString(os.Args[2])
		if err != nil || len(proof) != 64 {
			fmt.Println("proof passed is not a 64-byte value in hexadecimal (", len(proof), ")")
			return
		}

		// configure the Disco connection with Noise_XX
		clientConfig := libdisco.Config{
			KeyPair:              clientKeyPair,
			RemoteKey:            serverPublicKey,
			HandshakePattern:     libdisco.Noise_X,
			StaticPublicKeyProof: proof,
		}

		// Dial the port 6666 of localhost
		client, err := libdisco.Dial("tcp", "127.0.0.1:6666", &clientConfig)
		if err != nil {
			fmt.Println("client can't connect to server:", err)
			return
		}
		defer client.Close()
		fmt.Println("connected to", client.RemoteAddr())

		// write whatever stdin has to say to the socket
		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			_, err = client.Write([]byte(scanner.Text()))
			if err != nil {
				fmt.Println("client can't write on socket:", err)
			}
		}

		return
	}

	// usage
	fmt.Println("read source code to find out usage")
	return
}
