package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {
	// generating the client key pair
	clientKeyPair := libdisco.GenerateKeypair(nil)
	fmt.Println("client's public key:", clientKeyPair.ExportPublicKey())

	// configure the Disco connection
	clientConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_K,
		KeyPair:          clientKeyPair,
	}

	// retrieve the server's public key from an argument
	fmt.Println("please enter the server's public key in hexadecimal")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	serverKey, _ := hex.DecodeString(scanner.Text())
	clientConfig.RemoteKey = serverKey

	// Dial the port 6666 of localhost
	client, err := libdisco.Dial("tcp", "127.0.0.1:6666", &clientConfig)
	if err != nil {
		fmt.Println("client can't connect to server:", err)
		return
	}
	defer client.Close()
	fmt.Println("connected to", client.RemoteAddr())

	// write whatever stdin has to say to the socket
	for {
		scanner.Scan()
		_, err = client.Write([]byte(scanner.Text()))
		if err != nil {
			fmt.Println("client can't write on socket:", err)
		}
	}
}
