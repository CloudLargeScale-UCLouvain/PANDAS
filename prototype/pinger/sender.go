package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: sender <csv_file> <num_packets> <packet_size>")
		return
	}

	csvFileName := os.Args[1]
	numPackets, err := strconv.Atoi(os.Args[2])
	packetSize, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Invalid number of packets: %v", err)
	}

	// Read CSV file
	file, err := os.Open(csvFileName)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV file: %v", err)
	}

	time.Sleep(15 * time.Second)
	// Iterate over each receiver
	for _, line := range lines {
		if len(line) != 6 {
			log.Printf("Skipping invalid line in CSV: %v", line)
			continue
		}

		role := line[5]
		ip := line[3]
		portStr := line[1]
		port, err := strconv.Atoi(portStr)
		fmt.Printf("IP: %s Port: %s role: %s\n", ip, port, role)
		if err != nil {
			log.Printf("Invalid port number in CSV: %v", err)
			continue
		}

		// Skip non-validator nodes
		if role != "validator" {
			continue
		}

		addr := fmt.Sprintf("%s:%d", ip, port)
		fmt.Printf("Sending %d packets of size %d bytes to %s\n", numPackets, packetSize, addr)
		// Send UDP packets
		for i := 0; i < numPackets; i++ {
			err := sendUDPPacket(addr, packetSize)
			if err != nil {
				log.Printf("Error sending UDP packet: %v", err)
			}
		}
	}
}

func sendUDPPacket(addr string, packetSize int) error {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	message := make([]byte, packetSize)
	_, err = conn.Write(message)
	if err != nil {
		return err
	}

	return nil
}
