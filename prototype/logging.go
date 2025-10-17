package main

import (
	"encoding/json"
	"log"
	"math/big"
	"time"
)

type EventCode int

const (
	HeaderSent EventCode = iota
	HeaderReceived
	SamplingFinished
	SamplingStarted
	SeedingStart
	SeedingEnd
)

type LogEvent struct {
	Timestamp string    `json:"timestamp"`
	EventType EventCode `json:"eventType"`
	BlockId   int       `json:"blockId"`
}

type LogMessageCount struct {
	Timestamp    string `json:"timestamp"`
	MessageCount int    `json:"messageCount"`
	BlockId      int    `json:"blockId"`
}

type LogEntry struct {
	Timestamp      string      `json:"timestamp"`
	SenderID       string      `json:"SenderID"`
    BlockID        int         `json:"BlockID"`
	SampleIDsByRow []big.Int   `json:"SampleIDsByRow"` // List of sample row IDs
	Samples        []byte      `json:"Samples"`        // List of samples in a response
	Peers          []string    `json:"Peers"`          // List of peers in a response
	MessageType    MessageType `json:"MessageType"`    // message type
}

func formatJSONLogMessageCount(messageCount int, blockId int) string {
	// Custom log entry struct
	logEntry := LogMessageCount{
		Timestamp:    time.Now().Format(time.RFC3339Nano),
		MessageCount: messageCount,
		BlockId:      blockId,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}

func formatJSONLogEvent(eventType EventCode, blockId int) string {
	// Custom log entry struct
	logEntry := LogEvent{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		EventType: eventType,
		BlockId:   blockId,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}

func formatJSONLogMessageReceive(m Message) string {
	// Custom log entry struct
	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		SenderID:  m.SenderID,
        BlockID:   m.BlockID, 
		//SampleIDsByRow: m.SampleIDsByRow,
		//Samples:        m.Samples,
		//Peers:          m.Peers,
		SampleIDsByRow: []big.Int{},
		Samples:        []byte{0, 0, 0},
		Peers:          []string{" "},
		MessageType:    m.MessageType,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}

func formatJSONLogMessageSend(m *Message) string {
	// Custom log entry struct
	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		SenderID:  m.SenderID,
        BlockID:   m.BlockID,
		//SampleIDsByRow: m.SampleIDsByRow,
		//Samples:        m.Samples,
		//Peers:          m.Peers,
		SampleIDsByRow: []big.Int{},
		Samples:        []byte{0, 0, 0},
		Peers:          []string{" "},
		MessageType:    m.MessageType,
	}

	// Marshal log entry to JSON
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return ""
	}

	return string(jsonData)
}
