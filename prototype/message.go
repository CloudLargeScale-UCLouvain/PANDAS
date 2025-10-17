package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net"
    "strconv"
    "strings"
    "time"
)

// ========== Struct definition ==========

const sampleSize = 56; //1 //in bytes

type MessageType int

const (
    SamplingRequest MessageType = iota
    SamplingResponse
    SeedingRequest
    SeedingResponse
    Ping
)

func (me MessageType) String() string {
    return [...]string{"SamplingRequest", "SamplingResponse", "SeedingRequest", "SeedingResponse", "Ping"}[me]
}

type Message struct {
    SenderID       string      `json:"SenderID"`
    BlockID        int         `json:"BlockID"`
    SampleIDsByRow []int       `json:"SampleIDsByRow"` // List of sample row IDs
    Samples        []byte      `json:"Samples"`        // List of samples in a response
    Peers          []string    `json:"Peers"`          // List of peers in a response
    MessageType    MessageType `json:"MessageType"`    // message type
}

// Create one or more messages (in order to not exceed MTU)
func sendUDPSamples(blockID int, sampleIDsByRow []int, sampleNum int, peers []string, messageType MessageType, destAddr string) error {

    var messages [][]byte
    log.Printf("Sending %d samples", sampleNum)

    messageCount := 1
    if sampleNum > 0 {
        messageCount = (sampleNum + MAX_SAMPLES_PER_PACKET - 1) / MAX_SAMPLES_PER_PACKET
    }
    log.Printf("Pushing %d messages", messageCount)

    for i := 0; i < messageCount; i++ {
        start := i * MAX_SAMPLES_PER_PACKET
        end := start + MAX_SAMPLES_PER_PACKET
        if end > sampleNum {
            end = sampleNum
        }

        messageSampleIDs := sampleIDsByRow[start:end]
        //message := createUDPMessage(myUDPAddr, blockID, messageSampleIDs, len(messageSampleIDs), peers, messageType)
				message := createUDPMessage(myUDPAddr, blockID, messageSampleIDs, len(messageSampleIDs), []string{destAddr}, messageType)
        // Append a newline character to the message
        //message = append(message, '\n')
        messages = append(messages, message)
    }

    dstIp, dstPort, err := parseIPPort(destAddr)
    if err != nil {
        log.Panicf("Error: %v parsing IP port from %s", err, destAddr)
    }
    err = sendUDPMessagesWithCachedConnections(messages, dstIp, dstPort)

    return err
}

func sendUDPRequests(blockID int, sampleIDsByRow [] int, messageType MessageType, dstIp string, dstPort int) error {
    var messages [][]byte
    numSamplesToRequest := len(sampleIDsByRow)
    log.Printf("Requesting %d samples", numSamplesToRequest)

    messageCount := 1
    if numSamplesToRequest > 0 {
        messageCount = (numSamplesToRequest + MAX_SAMPLE_REQUESTS_PER_PACKET - 1) / MAX_SAMPLE_REQUESTS_PER_PACKET
    }
    log.Printf("Sending %d sample request messages", messageCount)

    for i := 0; i < messageCount; i++ {
        start := i * MAX_SAMPLE_REQUESTS_PER_PACKET 
        end := start + MAX_SAMPLE_REQUESTS_PER_PACKET 
        if end >  numSamplesToRequest {
            end = numSamplesToRequest 
        }

        messageSampleIDs := sampleIDsByRow[start:end]
        message := createUDPMessage(myUDPAddr, blockID, messageSampleIDs, 0, nil, messageType)
        // Append a newline character to the message
        //message = append(message, '\n')
        messages = append(messages, message)
    }
    err := sendUDPMessagesWithCachedConnections(messages, dstIp, dstPort)

    return err
}

func sendUDPMessages(messages [][]byte, ip string, port int) error {
    addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, strconv.Itoa(port)))
    if err != nil {
        return fmt.Errorf("failed to resolve address: %v", err)
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        return fmt.Errorf("failed to dial UDP: %v", err)
    }
    defer conn.Close()

    for _, msg := range messages {
        _, err := conn.Write(msg)
        if err != nil {
            return fmt.Errorf("failed to send message: %v", err)
        }
        time.Sleep(1 * time.Microsecond)
    }

    return nil
}


func sendUDPMessagesWithCachedConnections(messages [][]byte, ip string, port int) error {
    key := ip + ":" + strconv.Itoa(port)

    conn, exists := udpConnCache[key]
    if !exists {
        addr, err := net.ResolveUDPAddr("udp", key)
        if err != nil {
            return fmt.Errorf("failed to resolve address: %v", err)
        }

        conn, err = net.DialUDP("udp", nil, addr)
        if err != nil {
            return fmt.Errorf("failed to dial UDP: %v", err)
        }

        udpConnCache[key] = conn
    }

    for _, msg := range messages {
        _, err := conn.Write(msg)
        if err != nil {
            return fmt.Errorf("failed to send message: %v", err)
        }
        time.Sleep(1 * time.Microsecond)
    }

    return nil
}


func createUDPMessage(senderID string, blockID int, sampleIDsByRow []int, sampleNum int, peers []string, messageType MessageType) []byte {

    m := &Message{
        SenderID:       senderID,
        BlockID:        blockID,
        SampleIDsByRow: sampleIDsByRow,
        Samples:        make([]byte, sampleNum*sampleSize),
        Peers:          peers,
        MessageType:    messageType,
    }

    jsonData, err := json.Marshal(m)
    if err != nil {
        log.Println("Error encoding struct to binary")
        panic(err)
    }
    // Append a newline character to the JSON data
    jsonData = append(jsonData, '\n')

    //fmt.Println(string(jsonData))
    addEvent(formatJSONLogMessageSend(m))
    return jsonData
}

func (m *Message) String() string {
    return fmt.Sprintf("Message - BlockID: %d, SenderID: %s, SampleIDsByRow: %v, MessageType: %s", m.BlockID, m.SenderID, m.SampleIDsByRow, m.MessageType)
}

func readMessage(message Message) {
    log.Println(message)
}

// sendMessageToPeer sends a UDP message to the specified IP address and port.
func sendUDPMessageToPeer(msg []byte, ip string, port int) error {
    // Resolve the address
    addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
    if err != nil {
        return fmt.Errorf("failed to resolve address: %v", err)
    }

    // Create a UDP connection
    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        return fmt.Errorf("failed to dial UDP: %v", err)
    }
    defer conn.Close()

    // Send the message
    _, err = conn.Write(msg)
    if err != nil {
        return fmt.Errorf("failed to send message: %v", err)
    }

    return nil
}

func createMessageParse(senderID string, sampleIDsByRow []int, sampleNum int, peers []string, messageType MessageType) []byte {
    m := &Message{
        SenderID:       senderID,
        BlockID:        0,
        SampleIDsByRow: sampleIDsByRow,
        Samples:        make([]byte, sampleNum*sampleSize),
        Peers:          peers,
        MessageType:    messageType,
    }

    jsonData, err := json.Marshal(m)
    if err != nil {
        log.Println("Error encoding struct to binary")
        panic(err)
    }
    //fmt.Println(string(jsonData))
    addEvent(formatJSONLogMessageSend(m))
    return jsonData
}

func parseIPPort(address string) (string, int, error) {
    // Split the address into IP and port parts
    parts := strings.Split(address, ":")
    if len(parts) != 2 {
        return "", 0, fmt.Errorf("invalid address format")
    }

    // Extract the IP part
    ip := parts[0]

    // Extract and convert the port part
    port, err := strconv.Atoi(parts[1])
    if err != nil {
        return "", 0, fmt.Errorf("invalid port format: %v", err)
    }

    // Validate the IP address
    if net.ParseIP(ip) == nil {
        return "", 0, fmt.Errorf("invalid IP address format")
    }

    return ip, port, nil
}

func handleMessageUDP(data []byte, messageChannel chan<- Message) {
    // Convert the byte slice to a string
    messageStr := string(data)
    /*
    // Ensure the message ends with a newline character (it should already be true)
    if !strings.HasSuffix(messageStr, "\n") {
        log.Println("Error: Message does not end with a newline: %s", messageStr)
        return
    }

    // Remove the newline character before unmarshaling
    messageStr = strings.TrimSuffix(messageStr, "\n")

    // Check if the message string is empty
    if len(messageStr) == 0 {
        log.Println("Error: Received empty message")
        return
    }*/

    // Unmarshal the message
    var message Message
    err := json.Unmarshal([]byte(messageStr), &message)
    if err != nil {
        log.Println("Error decoding Message:", err)
        log.Printf("Received erroneuos message: %s", messageStr)
        return
    }

    // Send the message over the messageChannel
    messageChannel <- message
}
