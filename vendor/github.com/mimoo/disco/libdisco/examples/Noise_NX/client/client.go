package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("usage: go run client.go hex_root_public_key")
	}

	// retrieve root public key
	rootPublicKey, err := hex.DecodeString(os.Args[1])
	if err != nil || len(rootPublicKey) != 32 {
		fmt.Println("public root key passed is not a 32-byte value in hexadecimal (", len(rootPublicKey), ")")
		return
	}

	// create a verifier for when we will receive the server's public key
	verifier := libdisco.CreatePublicKeyVerifier(rootPublicKey)

	// configure the Disco connection
	clientConfig := libdisco.Config{
		HandshakePattern:  libdisco.Noise_NX,
		PublicKeyVerifier: verifier,
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

}
