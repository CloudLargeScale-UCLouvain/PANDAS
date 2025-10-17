package main

import (
    "context"
    "fmt"
    "log"
    "sort"
    "time"

    "github.com/libp2p-das/sample"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/multiformats/go-multiaddr"
)

// Global variables
// Ordered neighbors used by sampleOrderedPush
var orderedNeighbors []*Neighbor = nil

// Handle all the message processing for builder here
func handleEventsBuilder(exp_duration int, h host.Host, messageChannel <-chan Message, streamManager *PeerStreamManager, ctx context.Context, finished chan bool, LogDirectory string, NickFlag string) {
    var pub *Pub
    var err error

    //Create pubsub
    if config.EnableHeaderDis {
        pub, err = CreatePubSub(h, ctx, LogDirectory, NickFlag)
        if err != nil {
            log.Panicln("Error creating pubSub:", err)
        }
    }
    // setup timers for block generation and end of experiment
    expeDurationTicker := time.NewTicker(time.Duration(exp_duration) * time.Second)
    defer expeDurationTicker.Stop()
    log.Printf("Builder node will start publishing blocks in six minutesZ...")
    delayTimer := time.NewTimer(time.Duration(360) * time.Second)
    defer delayTimer.Stop()
    msgcount := 0
    // Set the ticker period to a high value now. It is later adjusted to 12 when delayTimer expires below
    blockTicker := time.NewTicker(time.Duration(BLOCK_TIME) * time.Second)
    defer blockTicker.Stop()
    log.Printf("Builder node handling events...")
    sleepTimeOver := false
    for {
        select {
        case <-delayTimer.C:
            log.Println("The initial sleep time is over")
            sleepTimeOver = true
        case <-expeDurationTicker.C:
            log.Println("Experiment time exceded")
            finished <- true
            return
        case msg := <-messageChannel:
            //log.Printf("Got a message %s", msg)
            addEvent(formatJSONLogMessageReceive(msg))
            msgcount += processMessageBuilder(streamManager, msg)
        case <-blockTicker.C:
            if sleepTimeOver == true {
                if blockID < 0 {
                    blockID = 0
                }
                addEvent(formatJSONLogMessageCount(msgcount, blockID))
                msgcount = 0
                /*
                   if currBlock == nil {
                       log.Println("About to generate a new block")
                       currBlock = sample.NewBlock(blockID, BLOCK_DIM, BLOCK_DIM, len(searchTable.validators))
                   } else {
                       currBlock.BlockID = blockID
                       currBlock.NetSize = len(searchTable.validators)
                   }
                   //currBlock = sample.NewBlock(blockID, BLOCK_DIM, BLOCK_DIM, len(searchTable.validators))
                   log.Println("About to check for unhosted samples")
                   CheckForUnHostedSamples(currBlock, searchTable.validators)
                */

                if config.EnableHeaderDis {
                    log.Printf("About to publish the block %d", blockID)
                    addEvent(formatJSONLogEvent(HeaderSent, blockID))

                    go pub.HeaderPublish(blockID)
                }
                go seedSamplesOrderedPush(blockID)
                log.Println("Done publishing the block")
                blockID += 1
            }
        default:
        }
    }
}

func CheckForUnHostedSamples(curreBlock *sample.Block, validators map[string]*Neighbor) map[int]struct{} {
    //numValidators := countPeersWithRole(peers, "validator")
    log.Printf("There are %d validators", len(validators))
    radius := currBlock.ComputeRegionRadius(NUM_SAMPLE_COPIES, len(validators))
    unhostedSamples := make(map[int]struct{})
    //log.Printf("Number of rows: %d Number of columns: %d", currBlock.NumRows, currBlock.NumCols)
    for i := 0; i < currBlock.NumRows; i++ {
        for j := 0; j < currBlock.NumCols; j++ {
            s := currBlock.BlockSamples[i][j]
            hosted := false
            //log.Printf("Sample %s", s.IdByRow.String())
            for _, peer := range validators {
                if peer.Role != "validator" {
                    continue
                }
                //log.Printf("\tPeer %s is in region: %t", peer.Id, s.IsInRegion(peer.Id, radius))
                if s.IsInRegion(peer.Id, radius) {
                    //log.Printf("\t\t hosted by %s, %s", peer.Id, peer.Role)
                    hosted = true
                }
            }
            if !hosted {
                unhostedSamples[s.SeqNumber] = struct{}{}
                //log.Printf("Sample %s is not hosted by anyone", s.IdByRow.String())
            }
        }
    }
    log.Printf("There are %d unhosted samples: %v", len(unhostedSamples), unhostedSamples)
    /*for sample, _ := range unhostedSamples {
        log.Printf("Sample %d is not hosted by anyone", sample)
    }*/
    return unhostedSamples

}

func processMessageBuilder(streamManager *PeerStreamManager, msg Message) int {
    switch msg.MessageType {
    case SeedingRequest:
        log.Printf("Handling SeedRequest message")
        onSeedRequest(streamManager, msg)
        return 0
    case Ping:
        log.Printf("Ping received")
        return 1
    default:
        log.Panicf("ERROR: Builder shouldn't receive MessageType: %s", msg.MessageType)
        return 0
    }
}

func onSeedRequest(streamManager *PeerStreamManager, msg Message) {
    numSamples := BLOCK_DIM * BLOCK_DIM * NUM_SAMPLE_COPIES / len(searchTable.validators)
    log.Printf("Validator %s requesting %d samples", msg.SenderID, numSamples)

    senderMultiAddr := msg.SenderID
    multiAddr, err := multiaddr.NewMultiaddr(senderMultiAddr)
    if err != nil {
        log.Println("Error parsing peer address:", err)
        return
    }
    peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
    if err != nil {
        log.Println("Error creating peer.AddrInfo:", err)
        return
    }
    //ctx := context.Background()
    //stream, err := streamManager.GetOrCreateStream(ctx, peerInfo)
    outmsg := createMessageParse(streamManager.myAddr, nil, numSamples, nil, SeedingResponse)
    streamManager.sendMessageToPeer(outmsg, peerInfo)

}

func seedSamplesOrderedPush(blockID int) {
    log.Printf("Start seeding samples for block %d", blockID)
    if orderedNeighbors == nil {
        keys := make([]string, 0, len(searchTable.validators))
        neighborsByIpAndPort := make(map[string]*Neighbor)

        for _, neighbor := range searchTable.validators {
            key := fmt.Sprintf("%s:%d", neighbor.Ip, neighbor.Port)
            neighborsByIpAndPort[key] = neighbor
            keys = append(keys, key)
        }

        sort.Strings(keys)

        for _, key := range keys {
            orderedNeighbors = append(orderedNeighbors, neighborsByIpAndPort[key])
        }
    }
		addEvent(formatJSONLogEvent(SeedingStart, blockID))
    for _, neighbor := range orderedNeighbors {
        samples := peerToSamples[neighbor.Id.String()]
				log.Printf("Block:%i Pushing samples to IP:%d Port: %d with UDP addr: %s", blockID, neighbor.Ip, neighbor.Port, neighbor.addrUDPStr)
        err := sendUDPSamples(blockID, samples, len(samples), nil, SeedingResponse, neighbor.addrUDPStr)
        if err != nil {
            log.Println("Error sending UDP samples: ", err)
        }
    }
		addEvent(formatJSONLogEvent(SeedingEnd, blockID))
		log.Printf("End seeding samples for block %d", blockID)
}

//func seedSamplesPush(streamManager *PeerStreamManager) {
func seedSamplesPush(blockID int) {

	neighborsByIp := make(map[string][]*Neighbor)

	for _, neighbor := range searchTable.validators {
		ip := neighbor.Ip
		neighborsByIp[ip] = append(neighborsByIp[ip], neighbor)
	}

    for _, neighbors := range neighborsByIp {
    
        for _, neighbor := range neighbors {
            samples := peerToSamples[neighbor.Id.String()]
            log.Printf("Pushing samples to IP:%s Port: %d with UDP addr: %s", neighbor.Ip, neighbor.Port, neighbor.addrUDPStr)
            err := sendUDPSamples(blockID, samples, len(samples), nil, SeedingResponse, neighbor.addrUDPStr)
            if err != nil {
                log.Println("Error sending UDP samples: ", err)
            }
            //sendSamplesToValidator(streamManager, neighbor)
            //time.Sleep(10 * time.Millisecond)
        }
    }

    /*
    for _, neighbor := range searchTable.validators {
        samples := peerToSamples[neighbor.Id.String()]
        log.Printf("Pushing samples to IP:%s Port: %d with UDP addr: %s", neighbor.Ip, neighbor.Port, neighbor.addrUDPStr)
        err := sendUDPSamples(blockID, samples, len(samples), nil, SeedingResponse, neighbor.addrUDPStr)
        if err != nil {
            log.Println("Error sending UDP samples: ", err)
        }
        //sendSamplesToValidator(streamManager, neighbor)
        time.Sleep(5 * time.Millisecond)
    }*/

}

func sendSamplesToValidator(streamManager *PeerStreamManager, peer *Neighbor) {
    numSamples := BLOCK_DIM * BLOCK_DIM * NUM_SAMPLE_COPIES / len(searchTable.validators)
    msg := createMessageParse(streamManager.myAddr, nil, numSamples, nil, SeedingResponse)
    streamManager.sendMessageToPeer(msg, peer.PeerInfo)
}
