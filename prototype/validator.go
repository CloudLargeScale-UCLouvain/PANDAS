package main

import (
    "context"
    //"fmt"
    "log"
    "math/rand"
    "time"

    "github.com/libp2p-das/sample"
    "github.com/libp2p/go-libp2p/core/host"
    //"github.com/libp2p/go-libp2p/core/peer"
    //"github.com/multiformats/go-multiaddr"
)

func handleEventsValidator(exp_duration int, h host.Host, messageChannel <-chan Message, streamManager *PeerStreamManager, ctx context.Context, finished chan bool, LogDirectory string, NickFlag string) {

    log.Printf("Validator node handling events...")
    expeDurationTicker := time.NewTicker(time.Duration(exp_duration) * time.Second)
    defer expeDurationTicker.Stop()
    msgcount := 0
    
    //Create pubsub
    var pubMessagesChannel <-chan *HeaderMessage
    if config.EnableHeaderDis {
        pub, err := CreatePubSub(h, ctx, LogDirectory, NickFlag)
        if err != nil {
            log.Println("Error creating pubSub:", err)
            return
        }
        go pub.readLoop()
        pubMessagesChannel = pub.messages
    } else {
        pubMessagesChannel = nil
    }

    for {
        select {

        case m := <-pubMessagesChannel:
            log.Printf("\033[34m Received Block %d Header Publish \033[0m", m.BlockID)
            addEvent(formatJSONLogEvent(HeaderReceived, m.BlockID))
        case <-expeDurationTicker.C:
            log.Println("Experiment time exceded")
            log.Printf("I still miss %d samples from the previous block", s.GetSamplesIMissCount())
            finished <- true
            return
        case msg := <-messageChannel:
            addEvent(formatJSONLogMessageReceive(msg))
            processMessageValidator(streamManager, msg, msgcount)
            msgcount += 1
        default:
        }
    }

}

func processMessageValidator(streamManager *PeerStreamManager, msg Message, msgcount int) {
    
    switch msg.MessageType {

    case SeedingRequest:
        log.Printf("Received Seed Request %+v", msg)
        go validatorOnSamplingRequest(streamManager, msg, SeedingResponse)

    case SeedingResponse:
        //log.Printf("Received Seed Response with msg.BlockID: %d", msg.BlockID)
        if blockID != msg.BlockID {
            log.Printf("Received Seed samples for a new block msg.BlockID: %d", msg.BlockID)
            // seeds arrived for a new block
            blockID = msg.BlockID
            msgcount = 0
            s.SetSamplesIHave(make(map[int]struct{}))
            s.SetSamplesIMiss(make(map[int]struct{}))
            // Do not regenerate a block if it has been already generated
            if currBlock == nil {
                currBlock = sample.NewBlock(blockID, BLOCK_DIM, BLOCK_DIM, len(searchTable.validators))
            } else {
                currBlock.BlockID = blockID
                currBlock.NetSize = len(searchTable.validators)
            }
        }
        onBuilderSeedResponse(msg)
        ihavecount := s.GetSamplesIHaveCount()
        //log.Printf("I have count: %d and I need count %d", ihavecount, len(peerToSamples[myself.Id.String()]))
        if ihavecount == len(peerToSamples[myself.Id.String()]) {
                log.Printf("Got all the seed samples")
                //go doValidatorSampling(blockID)
								addEvent(formatJSONLogEvent(SamplingStarted, currBlock.BlockID))
                go doRegularSampling(streamManager, blockID, searchTable.validators)
                go respondToPendingRequests(streamManager, blockID)
        } else {
            numMissingSeeds := len(peerToSamples[myself.Id.String()]) - ihavecount
            if numMissingSeeds < 0 {
                log.Printf("Message: %+v", msg)
            }
            log.Printf("Still missing %d seed samples", numMissingSeeds)
        }

    case SamplingRequest:
        log.Printf("Sample request received")
        go validatorOnSamplingRequest(streamManager, msg, SamplingResponse)

    case SamplingResponse:

        log.Printf("Sample response received")

        go validatorOnSamplingResponse(msg, 0, blockID)
    case Ping:

        log.Printf("Ping received")
    default:

        log.Printf("Unknown MessageType: %s", msg.MessageType)
        //panic("Unknown MessageType")
    }
}

/*
func startSeedingFromBuilder(streamManager *PeerStreamManager) {
    //  time.Sleep(200 * time.Millisecond)
    builder := searchTable.builder
    numSamples := BLOCK_DIM * BLOCK_DIM * NUM_SAMPLE_COPIES / (len(searchTable.validators) + 1)
    log.Println("Requesting samples from builder:", builder)
    log.Printf("We'll request %d missing samples, num validators: %d", numSamples, len(searchTable.validators))

    msg := createMessageParse(streamManager.myAddr, nil, 0, nil, SeedingRequest)
    log.Printf("Sending seed request to builder")
    streamManager.sendMessageToPeer(msg, builder.PeerInfo)
    //sendMessage(stream, msg)
}*/

func onBuilderSeedResponse(msg Message) {

    //log.Printf("Received %d samples from Builder", len(msg.Samples)/sampleSize)
    //TODO seed samples may arrive in multiple messages
    s.SeedSamplesReceived(msg.SampleIDsByRow)
}

//var missingSampleIDs map[string]struct{}

func doValidatorSampling(blockID int) {

    log.Printf("~~~~~~~~~~~~~~~~Doing validator sampling~~~~~~~~~~~~~~~~~~~~~")
    //logger.Println(formatJSONLogEvent(2, blockID))
    //  time.Sleep(500 * time.Millisecond)
    //select random 2 rows and two columns
    // Simulate set with a map, we cannot use big.Int as keys, so we'll have to temporarily use strings (big.Int.String())
    rowsToSample := make(map[int]struct{})
    for len(rowsToSample) < NUM_ROWS_TO_SAMPLE {
        r := rand.Intn(BLOCK_DIM)
        r += 1 //rows and columns number starts from 1
        rowsToSample[r] = struct{}{}
    }

    log.Println("Chosen rows to sample:", rowsToSample)

    columnsToSample := make(map[int]struct{})
    for len(columnsToSample) < NUM_COLS_TO_SAMPLE {
        r := rand.Intn(BLOCK_DIM)
        r += 1 //rows and columns number starts from 1
        columnsToSample[r] = struct{}{}
    }

    log.Println("Chosen columns to sample:", columnsToSample)

    // Request only the first half of the row/column
    for rowNum := range rowsToSample {
        rowSamples := currBlock.GetSamplesByRow(rowNum)
        startIndex := rand.Intn(2) * (BLOCK_DIM / 2)
        endIndex := startIndex + (BLOCK_DIM / 2)

        s.SamplesWanted(rowSamples[startIndex:endIndex])
    }

    for colNum := range columnsToSample {
        colSamples := currBlock.GetSamplesByColumn(colNum)
        startIndex := rand.Intn(2) * (BLOCK_DIM / 2)
        endIndex := startIndex + (BLOCK_DIM / 2)

        s.SamplesWanted(colSamples[startIndex:endIndex])
    }

    if s.GetSamplesIMissCount() == 0 {
        addEvent(formatJSONLogEvent(SamplingFinished, currBlock.BlockID))
        //fmt.Println("No samples to fetch - I got everything from the builder!")
    } else {
        fetch_samples(currBlock, s, searchTable.validators, SamplingRequest)
    }
}

func validatorOnSamplingResponse(msg Message, msgCount int, blockId int) {

    log.Printf("Received %d samples from %s", len(msg.SampleIDsByRow), msg.SenderID)

    s.RandomSamplesReceived(msg.SampleIDsByRow)

    missingSamplesLen := s.GetSamplesIMissCount()
    if missingSamplesLen == 0 {
        addEvent(formatJSONLogEvent(SamplingFinished, currBlock.BlockID))
    }

    addEvent(formatJSONLogMessageCount(msgCount, currBlock.BlockID))
    addEvent(formatJSONLogMessageCount(msgCount, currBlock.BlockID))
    log.Printf("I'm still missing %d samples", missingSamplesLen)
    /*
    if missingSamplesLen > 0 {
        missingSamples := s.GetSamplesIMissList()
        log.Printf("Missing samples are: %v", missingSamples)
    }*/
}

func validatorOnSamplingRequest(streamManager *PeerStreamManager, msg Message, messageType MessageType) {

    log.Printf("Received sampling request for %d samples", len(msg.SampleIDsByRow))

    senderMultiAddr := msg.SenderID
    samplesIWillProvide := make([]int, 0)
    samplesIHave := s.GetSamplesIHaveMap()
    for _, sampleID := range msg.SampleIDsByRow {
        if _, exists := samplesIHave[sampleID]; exists {
            samplesIWillProvide = append(samplesIWillProvide, sampleID)
        } else {
            streamManager.AddPendingRequest(senderMultiAddr, sampleID, messageType)
        }
    }

    log.Printf("I will provide %d samples", len(samplesIWillProvide))
    if len(samplesIWillProvide) == 0 {
        return
    }

    /*
    multiAddr, err := multiaddr.NewMultiaddr(senderMultiAddr)
    if err != nil {
        log.Println("Error parsing peer address:", err)
        return
    }
    peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
    */

    err := sendUDPSamples(blockID, samplesIWillProvide, len(samplesIWillProvide), nil, messageType, senderMultiAddr)
    if err != nil {
        log.Println("Error sending samples:", err)
        return
    }
    //msgOut := createMessageParse(streamManager.myAddr, samplesIWillProvide, len(samplesIWillProvide), nil, messageType)
    log.Printf("Sending sampling response to a peer: %s", senderMultiAddr)

    //sendMessage(stream, msgOut)
    //streamManager.sendMessageToPeer(msgOut, peerInfo)
}

func respondToPendingRequests(streamManager *PeerStreamManager, blockID int) {

    var addressesToRemove []string

    pendingRequests := s.GetPendingRequestsMap(streamManager)
    for addr, samples := range pendingRequests {
        /*
        multiAddr, err := multiaddr.NewMultiaddr(addr)
        if err != nil {
            log.Panic("Error parsing peer address:", err)
        }
        peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
        if err != nil {
            log.Panic("Error converting multiaddr to peerInfo:", err)
        }*/
        samplesIWillProvide := make([]int, 0)
        var samplesToRemove []int
        log.Printf("Pending requests for %d samples from peer: %s", len(samples), addr)
        samplesIHave := s.GetSamplesIHaveMap()
        msgTypeToSamples := make(map[MessageType][]int)
        
        for sampleSeq := range samples {
            if _, exists := samplesIHave[sampleSeq]; exists {
                msgType := pendingRequests[addr][sampleSeq]
                msgTypeToSamples[msgType] = append(msgTypeToSamples[msgType], sampleSeq)
                samplesIWillProvide = append(samplesIWillProvide, sampleSeq)
                samplesToRemove = append(samplesToRemove, sampleSeq)
            }/* else {
                //log.Printf("Asked for a sample %s that I do not yet have (seeding is probably not complete from all the validators)", sampleID.String())
            }*/
        }

        // If there is nothing to provide, skip this address
        if len(samplesIWillProvide) == 0 {
            continue
        }

        for msgType, samples := range msgTypeToSamples {
            //msgOut := createMessageParse(streamManager.myAddr, samples, len(samples), nil, msgType)
            //log.Printf("Sending %d pending samples to a peer: %s", len(samples), peerInfo)
            //streamManager.sendMessageToPeer(msgOut, peerInfo)
            if msgType == SeedingResponse {
                log.Printf("Sending %d pending seed samples to a peer: %s", len(samples), addr)
            } else if msgType == SamplingResponse {
                log.Printf("Sending %d pending random samples to a peer: %s", len(samples), addr)
            } else {
                panic("Unexpected messageType in respondToPendingRequest")
            }
            err := sendUDPSamples(blockID, samples, len(samples), nil, msgType, addr)
            if err != nil {
                log.Println("Error: sending pending samples failed: ", err)
            }
        }

        // Remove the samples we sent from the pending samples
        for _, s := range samplesToRemove {
            delete(pendingRequests[addr], s)
        }
    }

    // Remove addresses for which there are no pending sample requests
    for addr, samples := range pendingRequests {
        if len(samples) == 0 {
            addressesToRemove = append(addressesToRemove, addr)
        }
    }

    for _, addr := range addressesToRemove {
        delete(pendingRequests, addr)
    }
    s.SetPendingRequestsMap(streamManager, pendingRequests)

}
