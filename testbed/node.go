package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// create storage instance

func calculateUnhostedSamples(myID *big.Int) {

	// Pre-compute the mapping from each peer to the samples it hosts
	numValidators := len(searchTable.validators)
	numRegularsPlusValidators := len(searchTable.neighbors)
	radiusValidator := currBlock.ComputeRegionRadius(NUM_SAMPLE_COPIES, numValidators)
	radiusRegular := currBlock.ComputeRegionRadius(NUM_SAMPLE_COPIES, numRegularsPlusValidators)
	computePeerToSamples(myID, searchTable.neighbors, radiusValidator, radiusRegular)
	s.SetUnhostedSamples(CheckForUnHostedSamples(currBlock, searchTable.validators))

}

func pingRandomPeers(streamManager *PeerStreamManager) {
	//builder doesn't send ping as everyone ping it anyway
	if myself.Role == "builder" {
		return
	}

	// Wait for all the nodes to start
	time.Sleep((BLOCK_TIME / 3) * time.Second)

	//wait for a random amount of time to avoid overloading the builder
	randomWaitTime := time.Duration(rand.Intn((BLOCK_TIME / 3) * 1000))
	time.Sleep(randomWaitTime * time.Millisecond)

	//Send ping to the builder and some random nodes for the gossipsub header dissemination
	peers := make([]*Neighbor, 0)
	for _, v := range searchTable.neighbors {
		fmt.Println("v.Id: ", v.Id, "streamManager.myID: ", streamManager.myID)
		if v.Id.Cmp(streamManager.myID) != 0 {
			fmt.Println("Adding peer to peers")
			peers = append(peers, v)
		}
	}

	peers = append(peers, searchTable.builder)
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	if len(peers) > 20 {
		peers = peers[:20]
	}

	pingSuccess := 0
	for _, peer := range peers {
		/*
		   ctx := context.Background()
		   stream, err := streamManager.GetOrCreateStream(ctx, peer.PeerInfo)
		   if err != nil {
		       log.Println("Error starting a ping stream", err)
		       continue
		   }*/
		out_msg := createMessageParse(streamManager.myAddr, nil, 0, nil, Ping)

		log.Printf("Sending ping to node: %s", peer.PeerInfo.ID)
		pingSuccess += 1
		streamManager.sendMessageToPeer(out_msg, peer.PeerInfo)
	}
	log.Printf("Successfully pinged %d peers", pingSuccess)
}

// writeBinaryData writes binary data to a bufio.ReadWriter
func writeBinaryData(writer *bufio.Writer, data []byte) error {
	// Write the binary data to the ReadWriter
	_, err := writer.Write(data)
	if err != nil {
		return err
	}
	_, err = writer.WriteString("\n")
	if err != nil {
		return err
	}
	// Flush the buffer to ensure the data is written
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func computePeerToSamples(ownId *big.Int, peers map[string]*Neighbor, radiusValidator *big.Int, radiusRegular *big.Int) {
	log.Printf("computePeerToSamples")
	log.Printf("Number of validators: %d and radiusValidator: %s", len(peers), radiusValidator.String())
	for _, peer := range peers {
		radius := radiusValidator
		if peer.Role == "regular" {
			radius = radiusRegular
		}
		for i := 0; i < currBlock.NumRows; i++ {
			for j := 0; j < currBlock.NumCols; j++ {
				s := currBlock.BlockSamples[i][j]
				if s.IsInRegion(peer.Id, radius) {
					peerToSamples[peer.Id.String()] = append(peerToSamples[peer.Id.String()], s.SeqNumber)
					peerToSamplesBigInt[peer.Id.String()] = append(peerToSamplesBigInt[peer.Id.String()], *s.GetIDByRow())
				}
			}
		}
	}

}

func handleStream(s network.Stream, messageChannel chan<- Message) {

	//log.Println("Got a new stream!")

	rw := bufio.NewReader(s)

	for {
		binaryData, err := rw.ReadString('\n') // Adjust the delimiter based on your encoding

		if len(binaryData) == 0 || binaryData[len(binaryData)-1] != '\n' {
		} else {
			if err != nil {
				log.Println("Error Receiving Stream: ", err)
			}

			var message Message

			// Step 2: Unmarshal JSON into the struct
			err = json.Unmarshal([]byte(binaryData), &message)

			if err != nil {
				log.Println("Error decoding Message: ", err)
				//log.Println("Raw JSON Data:", binaryData)
			}

			log.Printf("Got message type %s from %s\n", message.MessageType, message.SenderID)

			messageChannel <- message
		}
	}
}

// TODO looks like this func is not used anymore, so remove
func connectToPeer(ctx context.Context, host host.Host, peerAddress string) {
	// Parse the peer multiaddress
	multiAddr, err := multiaddr.NewMultiaddr(peerAddress)
	if err != nil {
		log.Println("Error parsing peer address:", err)
		return
	}
	// Create a peer.AddrInfo from the multiaddress

	peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		log.Println("Error creating peer.AddrInfo:", err)
		return
	}
	// Connect to the peer
	err = host.Connect(ctx, *peerInfo)
	if err != nil {
		log.Println("Error connecting to peer:", err)
		return
	}

	log.Println("Connected to peer:", peerInfo.ID)

}

func generateTemporaryPeerID(ip string, port int) (peer.ID, error) {
	data := []byte(fmt.Sprintf("%s:%d", ip, port))
	hash := sha256.Sum256(data)
	hashString := hex.EncodeToString(hash[:])
	return peer.Decode(hashString)
}

func countValidators(filename string, myNick string) map[string]int {
	//log.Println("myNick:", myNick)
	positions := make(map[string]int)
	file, err := os.Open(filename)
	if err != nil {
		log.Panic("Error opening file:", err)
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV records: %s", err)
	}
	validatorCount := 0

	for _, record := range records {
		if record[len(record)-1] == "validator" {
			positions[record[0]] = validatorCount
		}
		validatorCount++
	}
	return positions
}

func readPeersFromFile(filename string) []*Neighbor {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	peers := make([]*Neighbor, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "-n") {
			// Skip the line containing "-n"
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) != 6 {
			log.Println("Invalid line format:", line)
			continue
		}

		nick := fields[0]
		role := fields[5]
		mhString := fields[4]
		ip := fields[3]
		portStr := fields[2]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Panicln("Error converting port:", err)
		}

		multiAddr, err := multiaddr.NewMultiaddr(mhString)
		if err != nil {
			log.Println("Error parsing peer address:", err)
			return nil
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
		if err != nil {
			log.Println("Error creating peer.AddrInfo:", err)
			return nil
		}

		/*log.Println("PeerInfo: ", peerInfo)
		  log.Printf("Multiaddr: %s, peerID: %s", mhString, peerInfo.ID.String())*/

		peer := NewNeighbour(nick, peerInfo.ID.String(), multiAddr, role, false, ip, port)
		peers = append(peers, peer)
	}
	return peers
}
