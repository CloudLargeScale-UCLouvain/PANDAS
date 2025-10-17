package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "IP address to listen on")
	port := flag.Int("port", 8000, "Port number to listen on")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *ip, *port)
	fmt.Printf("Listening for UDP packets on %s\n", addr)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Error listening for UDP: %v", err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024) // Buffer size

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading UDP: %v", err)
			continue
		}
		fmt.Printf("Received %d bytes from %s\n", n, udpAddr.String())
	}
}
