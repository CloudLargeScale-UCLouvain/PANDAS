package main

import (
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func main() {
	output_filename := os.Args[1]
	ip := os.Args[2]
	port := os.Args[3]

	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)

	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	tcpAddrString := fmt.Sprintf("/ip4/%s/tcp/%s/", ip, port)
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			tcpAddrString,
		),
	)

	//if err := os.WriteFile(fmt.Sprintf("./keys/%s.key", h.ID().String()), privBytes, 0666); err != nil {
	if err := os.WriteFile(fmt.Sprintf("%s", output_filename), privBytes, 0666); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		panic(err)
	}
	fmt.Print(fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", ip, port, h.ID().String()))

	//privBytes, err := os.ReadFile(output_filename)
	//priv, err := crypto.UnmarshalPrivateKey(privBytes)
	//if err != nil {
	//panic(err)
	//}
}
