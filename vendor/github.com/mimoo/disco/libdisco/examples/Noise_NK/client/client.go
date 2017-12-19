package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/mimoo/disco/libdisco"
)

func main() {
	// retrieve the server's public key from an argument
	serverPublicKey, _ := hex.DecodeString(os.Args[1])

	// configure the Disco connection with Noise_NK
	// meaning the client knows the key (retrieved from the CLI)
	clientConfig := libdisco.Config{
		HandshakePattern: libdisco.Noise_NK,
		RemoteKey:        serverPublicKey,
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
