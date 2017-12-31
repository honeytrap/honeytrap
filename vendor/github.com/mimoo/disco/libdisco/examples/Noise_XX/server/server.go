package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {

	//
	// run `go run server.go gen` to generate the static key of the server
	//
	if len(os.Args) == 2 && os.Args[1] == "setup" {

		// generating the server's keypair
		serverKeyPair, err := libdisco.GenerateAndSaveDiscoKeyPair("./serverKeyPair")
		if err != nil {
			panic("couldn't generate and save the server's key pair")
		}

		// displaying the public part
		fmt.Println("generated the server static public key successfuly. server's public key:")
		fmt.Println(hex.EncodeToString(serverKeyPair.PublicKey[:]))

		return
	}

	//
	// run `go run server.go accept hex_proof hex_root_public_key` to connect to a server
	//
	if len(os.Args) == 4 && os.Args[1] == "accept" {

		// load the server's keypair
		serverKeyPair, err := libdisco.LoadDiscoKeyPair("./serverkeyPair")
		if err != nil {
			fmt.Println("couldn't load the server's key pair")
			return
		}

		// retrieve root key
		rootPublicKey, err := hex.DecodeString(os.Args[3])
		if err != nil || len(rootPublicKey) != 32 {
			fmt.Println("public root key passed is not a 32-byte value in hexadecimal (", len(rootPublicKey), ")")
			return
		}

		// create a verifier for when we will receive the server's public key
		verifier := libdisco.CreatePublicKeyVerifier(rootPublicKey)

		// retrieve signature/proof
		proof, err := hex.DecodeString(os.Args[2])
		if err != nil || len(proof) != 64 {
			fmt.Println("proof passed is not a 64-byte value in hexadecimal (", len(proof), ")")
			return
		}

		// configure the Disco connection with Noise_XX
		serverConfig := libdisco.Config{
			KeyPair:              serverKeyPair,
			HandshakePattern:     libdisco.Noise_XX,
			PublicKeyVerifier:    verifier,
			StaticPublicKeyProof: proof,
		}

		// listen on port 6666
		listener, err := libdisco.Listen("tcp", "127.0.0.1:6666", &serverConfig)
		if err != nil {
			fmt.Println("cannot setup a listener on localhost:", err)
			return
		}
		addr := listener.Addr().String()
		fmt.Println("listening on:", addr)

		for {
			// accept a connection
			server, err := listener.Accept()
			if err != nil {
				fmt.Println("server cannot accept()")
				server.Close()
				continue
			}
			fmt.Println("server accepted connection from", server.RemoteAddr())
			// read what the socket has to say until connection is closed
			go func(server net.Conn) {
				buf := make([]byte, 100)
				for {
					n, err := server.Read(buf)
					if err != nil {
						fmt.Println("server can't read on socket for", server.RemoteAddr(), ":", err)
						break
					}
					fmt.Println("received data from", server.RemoteAddr(), ":", string(buf[:n]))
				}
				fmt.Println("shutting down the connection with", server.RemoteAddr())
				server.Close()
			}(server)

		}

		return
	}

	// usage
	fmt.Println("read source code to find out usage")
	return
}
