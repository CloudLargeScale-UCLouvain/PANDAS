package main

import (
    "context"
    "log"
    "math/big"
    "time"

    "github.com/libp2p-das/sample"
    "github.com/libp2p/go-libp2p/core/host"
)

func handleEventsRegular(exp_duration int, h host.Host, messageChannel <-chan Message, streamManager *PeerStreamManager, ctx context.Context, finished chan bool, LogDirectory string, NickFlag string) {

    log.Printf("Regular node handling events...")
    msgcount := 0
    //TODO
    expeDurationTicker := time.NewTicker(time.Duration(exp_duration) * time.Second)
    defer expeDurationTicker.Stop()

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
            msgcount = 0
            addEvent(formatJSONLogEvent(HeaderReceived, m.BlockID))
            if m.BlockID > 0 {
                log.Printf("I still miss %d samples from the previous block %d", s.GetSamplesIMissCount(), currBlock.BlockID)
            }

            // Do not regenerate a block if it has been already generated
            if currBlock == nil {
                currBlock = sample.NewBlock(m.BlockID, BLOCK_DIM, BLOCK_DIM, len(searchTable.neighbors))
            } else {
                currBlock.BlockID = m.BlockID
                currBlock.NetSize = len(searchTable.neighbors)
            }

            s.SetSamplesIHave(make(map[int]struct{}))
            s.SetSamplesIMiss(make(map[int]struct{}))

            go startSeedingFromValidators(streamManager)
            randomSamplingStarted = false
        case <-expeDurationTicker.C:
            log.Printf("I still miss %d samples from the previous block", s.GetSamplesIMissCount())
            log.Println("Experiment time exceded")
            finished <- true
            return
        case msg := <-messageChannel:

            //log.Printf("Got a message %", msg)

            addEvent(formatJSONLogMessageReceive(msg))
            processMessageRegular(streamManager, msg, currBlock.BlockID, msgcount)
            msgcount += 1
        default:
        }
    }
}

func processMessageRegular(streamManager *PeerStreamManager, msg Message, blockID int, msgcount int) {

    switch msg.MessageType {
    case SeedingResponse:

        log.Printf("Received Seed Response")

        onValidatorSeedResponse(msg)
        ihavecount := s.GetSamplesIHaveCount()
        if randomSamplingStarted == false && ihavecount == len(peerToSamples[myself.Id.String()]){
            randomSamplingStarted = true
            go doRegularSampling(streamManager, blockID, searchTable.neighbors)
        } else {
            numMissingSeeds := len(peerToSamples[myself.Id.String()]) - ihavecount
            log.Printf("Still missing %d seed samples", numMissingSeeds)
        }
        //go respondToPendingRequests(streamManager)
    case SamplingRequest:
        //TODO: handle sampling request

        log.Printf("Sample request received")

        go validatorOnSamplingRequest(streamManager, msg, SamplingResponse)
    case SamplingResponse:

        log.Printf("Sample response received")

        go validatorOnSamplingResponse(msg, msgcount, blockID)
    case Ping:

        log.Printf("Ping received")
    default:

        log.Printf("Unknown MessageType: %s", msg.MessageType)
        //panic("Unknown MessageType")
    }
}

func startSeedingFromValidators(streamManager *PeerStreamManager) {

    // Sleep a bit to give time for validators to finish voting
    time.Sleep(VOTING_DEADLINE * time.Second)
    // samples to obtain through seeding
    missingSamplesByRow := make(map[int]struct{})

    s.AddSamplesIMiss(peerToSamples[streamManager.myID.String()])

    for _, sample := range peerToSamples[streamManager.myID.String()] {
        missingSamplesByRow[sample] = struct{}{}
    }
    //log.Printf("We'll request %d missing seed samples: %v", len(missingSamplesByRow), missingSamplesByRow)
    log.Printf("We'll request %d missing seed samples", len(missingSamplesByRow))

    fetch_samples(currBlock, s, searchTable.validators, SeedingRequest)
}

func onValidatorSeedResponse(msg Message) {

    log.Printf("Received %d samples from %s", len(msg.Samples)/sampleSize, msg.SenderID)

    // This needs to update both samplesIMiss and samplesIHave because we
    // use fetch_samples to retrieve seed samples unlike validators that get seed
    // samples pushed to them
    s.SamplesReceived(msg.SampleIDsByRow)
    if s.GetSamplesIMissCount() == 0 {
        addEvent(formatJSONLogEvent(SamplingFinished, currBlock.BlockID))
    }
}

func doRegularSampling(streamManager *PeerStreamManager, blockID int, nghbrs map[string]*Neighbor) {
    //Wait for other regular nodes to get their samples
    //time.Sleep(1500 * time.Millisecond)

    log.Printf("~~~~~~~~~~~~~~~~Doing regular sampling~~~~~~~~~~~~~~~~~~~~~")
    numAllNodes := len(nghbrs)
    radius := currBlock.ComputeRegionRadius(NUM_SAMPLE_COPIES, numAllNodes)
    randomSamples := []*sample.Sample{}

    numIterations := 0
    for len(randomSamples) < NUM_RANDOM_SAMPLES {
        s := currBlock.GetRandomSample()
        if numIterations > 20*NUM_RANDOM_SAMPLES {
            log.Panic("Too many iterations to pick random samples")
        }
        numIterations += 1
        // We need to check the below condition because some of the validators
        // might have not responded to the seeding request so we can't
        // simply omit samples in SamplesIHave
        if s.IsInRegion(streamManager.myID, radius) {
            continue
        } else {
            randomSamples = append(randomSamples, s)
            //log.Printf("Random sample to fetch: %d", s.SeqNumber)
        }
    }

    s.SetSamplesIMiss(make(map[int]struct{}))
    s.SamplesWanted(randomSamples)

    if s.GetSamplesIMissCount() == 0 {
        addEvent(formatJSONLogEvent(SamplingFinished, currBlock.BlockID))
    } else {
        fetch_samples(currBlock, s, nghbrs, SamplingRequest)
    }

}

func removeElementByValue(slice []big.Int, valueToRemove big.Int) []big.Int {
    var result []big.Int

    for _, v := range slice {
        if v.Cmp(&valueToRemove) != 0 {
            result = append(result, v)
        }
    }

    return result
}
