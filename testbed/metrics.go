package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

type ParcelLog struct {
	timestamp int64
	colRow    int
	block     int
	size      int
	first     int
}

type MessageGlobalMetrics struct {
	fileName             string
	messageLogFile        string
	numberSend           int
	errorSend            int
	numberReceived       int
	errorReceived        int
	successBlockSampling int
	failedBlockSampling  int
	blocklogHashMap      map[int][]ParcelLog
}

func InitMessageMetrics(nodeRole string, nick string) *MessageGlobalMetrics {
	m := &MessageGlobalMetrics{
		fileName:             nodeRole + "-" + nick + "-MessageGlobal.csv",
		messageLogFile:        nodeRole + "-" + nick + "-MessageLog.csv",
		numberSend:           0,
		errorSend:            0,
		numberReceived:       0,
		errorReceived:        0,
		successBlockSampling: 0,
		failedBlockSampling:  0,
		blocklogHashMap:      make(map[int][]ParcelLog),
	}
	return m
}

func CreateParcelLog(colRow int, block int, size int, first int, timestamp int64) *ParcelLog {
	p := &ParcelLog{
		timestamp: timestamp,
		colRow:    colRow,
		block:     block,
		size:      size,
		first:     first,
	}
	return p
}

func (m *MessageGlobalMetrics) AddSend() {
	m.numberSend += 1
}

func (m *MessageGlobalMetrics) AddReceived() {
	m.numberReceived += 1
}

func (m *MessageGlobalMetrics) AddErrorSend() {
	m.errorSend += 1
}

func (m *MessageGlobalMetrics) AddErrorReceived() {
	m.errorReceived += 1
}

/*
	func (m *MessageGlobalMetrics) logHashMapElement(message *Message) {
		colRow, block, size, first := readMessage(message.Message)
		value, exists := m.blocklogHashMap[block]
		newEntry := CreateParcelLog(colRow, block, size, first, time.Now().UnixNano()/int64(time.Millisecond))
		if exists {
			value = append(value, *newEntry)
			m.blocklogHashMap[block] = value
		} else {
			logList := make([]ParcelLog, 0)
			logList = append(logList, *newEntry)
			m.blocklogHashMap[block] = logList
		}
	}
*/
func (m *MessageGlobalMetrics) WriteMessageGlobalCSV() {

	//Write global data
	file, err := os.Create(m.fileName)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	data := [][]string{
		{"# Message send", "# Send error", "# Message received", "# Received error"},
		{strconv.Itoa(m.numberSend), strconv.Itoa(m.errorSend), strconv.Itoa(m.numberReceived), strconv.Itoa(m.errorReceived)},
	}
	err = writer.WriteAll(data)
	if err != nil {
		log.Fatal("Error writing CSV:", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing CSV writer:", err)
	}

	//Write log messages
	file, err = os.Create(m.messageLogFile)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	writer = csv.NewWriter(file)

	data = [][]string{
		{"Block #", "# samples", "duration"},
	}

	for key, value := range m.blocklogHashMap {
		data = append(data, []string{strconv.Itoa(key), strconv.Itoa(len(value)), strconv.Itoa(int(value[len(value)-1].timestamp - value[0].timestamp))})
	}

	err = writer.WriteAll(data)
	if err != nil {
		log.Fatal("Error writing CSV:", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing CSV writer:", err)
	}

}
