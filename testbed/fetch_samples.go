package main

import (
	//"context"
	"log"
	"time"

	"github.com/libp2p-das/sample"
	//"github.com/libp2p-das/storage"
)

//var samplesBeingFetchedFromMutex sync.RWMutex

func requestSamples(blockID int, s *Storage, peer *Neighbor, samplesToRequest []int, messageType MessageType) {
	// Simulate fetching sample
	//TODO: Add code to request samples from the target peer

	//ctx := context.Background()
	//TODO: if opening a stream fails, maybe retry again before returning
    /*
	stream, err := streamManager.GetOrCreateStream(ctx, peer.PeerInfo)
	if err != nil {
		log.Println("Error starting a stream for sampling: ", err)
		return
	}*/
	missingSamples := s.GetSamplesIMissList()

	if len(missingSamples) == 0 {
		log.Printf("No samples left to request from peer: %s", peer.PeerInfo)
		return
	}
    // Compute the intersection of samplesToRequest and missingSamples

    // Create a map to store the elements of missingSamples
    sampleMap := make(map[int]struct{})
    for _, sample := range missingSamples {
        sampleMap[sample] = struct{}{}
    }

    // Find the intersection by checking which elements of samplesToRequest are in sampleMap
    var intersection []int
    for _, sample := range samplesToRequest {
        if _, exists := sampleMap[sample]; exists {
            intersection = append(intersection, sample)
        }
    }

	//log.Printf("Sending sampling request to peer: %s for %d samples: %v", peer.PeerInfo, len(missingSamples), intersection)
	log.Printf("Sending sampling request to peer: %s for %d samples", peer.PeerInfo, len(missingSamples))

	//msg := createMessageParse(streamManager.myAddr, missingSamples, 0, nil, messageType)
    //msg := createUDPMessage(myUDPAddr, blockID, missingSamples, 0, nil, messageType)
    //sendUDPMessageToPeer(msg, peer.Ip, peer.Port)
    sendUDPRequests(blockID, intersection, messageType, peer.Ip, peer.Port)
	//streamManager.sendMessageToPeer(msg, peer.PeerInfo)

}

func getPeerScore(peer *Neighbor, block *sample.Block, s *Storage, peers map[string]*Neighbor, samplesBeingFetchedFrom map[int]map[string]struct{}) (int, []int) {
	score := 0
	//radius := block.ComputeRegionRadius(NUM_SAMPLE_COPIES, len(peers))
	missingSamples := s.GetSamplesIMissMap()
    samplesToRequest := []int{}
	for _, sample := range peerToSamples[peer.Id.String()] {
		//log.Println("loop")
		if _, exists := missingSamples[sample]; exists {
            samplesToRequest = append(samplesToRequest, sample)
			if _, sample_exists := samplesBeingFetchedFrom[sample]; !sample_exists || len(samplesBeingFetchedFrom[sample]) < aggressivness {
				score += 1
			}
		}
	}

	return score, samplesToRequest
}

func getBestPeerToAsk(block *sample.Block, s *Storage, peers map[string]*Neighbor, samplesBeingFetchedFrom map[int]map[string]struct{}, askedNodes map[string]struct{}) (*Neighbor, int, []int) {

	maxScore := 0
	var bestPeer *Neighbor = nil
    bestSamplesToRequest := []int{}
	for k, v := range peers {
		if _, exists := askedNodes[k]; exists {
			continue
		}
		//this can happen for regular nodes
		//as they use fetch_samples to get samples they should host
		if myself.Id.String() == v.Id.String() {
			continue
		}
        score, samplesToRequest := getPeerScore(v, block, s, peers, samplesBeingFetchedFrom)

		// /log.Printf("Score for %s: %d", v.PeerInfo.ID.String(), score)
		if score > maxScore {
			maxScore = score
			bestPeer = v
            bestSamplesToRequest = samplesToRequest
		}
	}
	return bestPeer, maxScore, bestSamplesToRequest
}

// var samplesBeingFetchedFrom map[string]map[string]struct{}
var aggressivness int

func fetch_samples(block *sample.Block, s *Storage, peers map[string]*Neighbor, messageType MessageType) {
	//this is added in case some nodes have our samples but didn't respond for some reason
	//we clean the askedNodes map and try everything again
	for retry := 0; retry < 1; retry++ {
		numRounds := 3
		//samplesBeingFetchedFromMutex.Lock()
		samplesBeingFetchedFrom := make(map[int]map[string]struct{})
		//samplesBeingFetchedFromMutex.Unlock()
		askedNodes := make(map[string]struct{})
		aggressivness = 1

		for round := 1; round <= numRounds; round++ {
			if s.GetSamplesIMissCount() == 0 {
				log.Printf("All samples fetched")
				break
			}

			log.Printf("Round %d\n", round)

			for {
				//get current time
				start := time.Now()
				peer, score, samplesToRequest := getBestPeerToAsk(block, s, peers, samplesBeingFetchedFrom, askedNodes)
				if peer == nil || score == 0 {
					break
				}
				stop := time.Now()
				log.Printf("getBestPeerToAsk took %v", stop.Sub(start))
				//samplesBeingFetchedFromMutex.Lock()
				for _, sample := range peerToSamples[peer.Id.String()] {
					if _, exists := samplesBeingFetchedFrom[sample]; !exists {
						samplesBeingFetchedFrom[sample] = make(map[string]struct{})
					}
					samplesBeingFetchedFrom[sample][peer.Id.String()] = struct{}{}
				}
				//samplesBeingFetchedFromMutex.Unlock()
				go requestSamples(block.BlockID, s, peer, samplesToRequest, messageType)
				askedNodes[peer.Id.String()] = struct{}{}
			}
			time.Sleep(1 * time.Second)
			aggressivness += 1
		}

		//print samples I still miss
		if s.GetSamplesIMissCount() > 0 {
			log.Printf("I still miss samples after finishing fetching")
		}
	}
}
