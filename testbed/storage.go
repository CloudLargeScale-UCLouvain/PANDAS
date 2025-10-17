package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/libp2p-das/sample"

)

type Storage struct {
	samplesIHave    map[int]struct{}
	samplesIMiss    map[int]struct{}
	unhostedSamples map[int]struct{}

	// Mutexes
	samplesIHaveMutex    sync.RWMutex
	samplesIMissMutex    sync.RWMutex
	unhostedSamplesMutex sync.RWMutex
    pendingRequestMutex sync.RWMutex

	//Time spend waiting for each Mutex
	samplesIHaveTime    int64
	samplesIMissTime    int64
	unhostedSamplesTime int64
}

func NewStorage() *Storage {
	return &Storage{
		samplesIHave:        make(map[int]struct{}),
		samplesIMiss:        make(map[int]struct{}),
		unhostedSamples:     make(map[int]struct{}),
		samplesIHaveTime:    0,
		samplesIMissTime:    0,
		unhostedSamplesTime: 0,
	}
}

func (s *Storage) PrintTimeSpentOnMutexes() string {
	return "Time spent on samplesIHaveMutex: " + strconv.FormatInt(s.samplesIHaveTime, 10) + "ms\n" + "Time spent on samplesIMissMutex: " + strconv.FormatInt(s.samplesIMissTime, 10) + "ms\n" + "Time spent on unhostedSamplesMutex: " + strconv.FormatInt(s.unhostedSamplesTime, 10) + "ms\n"

}

func (s *Storage) SetUnhostedSamples(uh map[int]struct{}) {
	start := time.Now()
	s.unhostedSamplesMutex.Lock()
	stop := time.Now()
	s.unhostedSamplesTime += stop.Sub(start).Milliseconds()
	s.unhostedSamples = uh
	s.unhostedSamplesMutex.Unlock()
}

func (s *Storage) SetSamplesIHave(ih map[int]struct{}) {
	start := time.Now()
	s.samplesIHaveMutex.Lock()
	stop := time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()
	s.samplesIHave = ih
	s.samplesIHaveMutex.Unlock()
}

func (s *Storage) SetSamplesIMiss(im map[int]struct{}) {
	start := time.Now()
	s.samplesIMissMutex.Lock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()
	s.samplesIMiss = im
	s.samplesIMissMutex.Unlock()
}

func (s *Storage) AddSamplesIHave(samples []int) {
	start := time.Now()
	s.samplesIHaveMutex.Lock()
	stop := time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	for _, sample := range samples {
		s.samplesIHave[sample] = struct{}{}
	}
	s.samplesIHaveMutex.Unlock()
}

func (s *Storage) AddSamplesIMiss(samples []int) {
	start := time.Now()
	s.samplesIMissMutex.Lock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	start = time.Now()
	s.samplesIHaveMutex.RLock()
	stop = time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	start = time.Now()
	s.unhostedSamplesMutex.RLock()
	stop = time.Now()
	s.unhostedSamplesTime += stop.Sub(start).Milliseconds()

	for _, sample := range samples {
		if _, exists := s.unhostedSamples[sample]; exists {
			continue
		}
		if _, exists := s.samplesIHave[sample]; exists {
			continue
		}
		s.samplesIMiss[sample] = struct{}{}
	}
	s.unhostedSamplesMutex.RUnlock()
	s.samplesIHaveMutex.RUnlock()
	s.samplesIMissMutex.Unlock()
}

func (s *Storage) SeedSamplesReceived(samples []int) {

	start := time.Now()
	s.samplesIHaveMutex.Lock()
	stop := time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	for _, sample := range samples {
		s.samplesIHave[sample] = struct{}{}
	}

	s.samplesIHaveMutex.Unlock()
}

func (s *Storage) RandomSamplesReceived(samples []int) {

	start := time.Now()
	s.samplesIMissMutex.Lock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	for _, sample := range samples {
		delete(s.samplesIMiss, sample)
	}

	s.samplesIMissMutex.Unlock()
}

func (s *Storage) SamplesReceived(samples []int) {
	start := time.Now()
	s.samplesIMissMutex.Lock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	start = time.Now()
	s.samplesIHaveMutex.Lock()
	stop = time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	for _, sample := range samples {
		s.samplesIHave[sample] = struct{}{}
		delete(s.samplesIMiss, sample)
	}
	s.samplesIHaveMutex.Unlock()
	s.samplesIMissMutex.Unlock()
}

func (s *Storage) SamplesWanted(samples []*sample.Sample) {
	start := time.Now()
	s.samplesIMissMutex.Lock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	start = time.Now()
	s.samplesIHaveMutex.RLock()
	stop = time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	start = time.Now()
	s.unhostedSamplesMutex.RLock()
	stop = time.Now()
	s.unhostedSamplesTime += stop.Sub(start).Milliseconds()

	for _, sample := range samples {
		if _, exists := s.unhostedSamples[sample.SeqNumber]; exists {
			continue
		}
        //TODO add assert to check if this happens
		if _, exists := s.samplesIHave[sample.SeqNumber]; exists {
			continue
		}
		s.samplesIMiss[sample.SeqNumber] = struct{}{}
	}
	s.unhostedSamplesMutex.RUnlock()
	s.samplesIHaveMutex.RUnlock()

	s.samplesIMissMutex.Unlock()
}

func (s *Storage) GetSamplesIHaveCount() int {
	start := time.Now()
	s.samplesIHaveMutex.RLock()
	stop := time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	defer s.samplesIHaveMutex.RUnlock()
	return len(s.samplesIHave)
}

func (s *Storage) GetSamplesIMissCount() int {
	start := time.Now()
	s.samplesIMissMutex.RLock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	defer s.samplesIMissMutex.RUnlock()
	return len(s.samplesIMiss)
}

func (s *Storage) GetSamplesIMissList() []int {
	start := time.Now()
	s.samplesIMissMutex.RLock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	defer s.samplesIMissMutex.RUnlock()
	ids := make([]int, len(s.samplesIMiss))
	i := 0
	for k := range s.samplesIMiss {
		ids[i] = k
		i++
	}
	return ids
}

func (s *Storage) GetPendingRequestsMap(streamManager *PeerStreamManager) map[string]map[int]MessageType {

    s.pendingRequestMutex.RLock()
    defer s.pendingRequestMutex.RUnlock()

    
    // Create a new map to hold the copied data
    copiedMap := make(map[string]map[int]MessageType)

    // Iterate over the original map and copy its contents
    for key, value := range streamManager.pendingRequests {
        copiedValue := make(map[int]MessageType)
        for k, v := range value {
            copiedValue[k] = v
        }
        copiedMap[key] = copiedValue
    }

    // Return the copied map
    return copiedMap
}

func (s *Storage) SetPendingRequestsMap(streamManager *PeerStreamManager, newMap map[string]map[int]MessageType) {
    // Lock for writing to ensure exclusive access to pm.pendingRequests
    s.pendingRequestMutex.Lock()
    defer s.pendingRequestMutex.Unlock()

    // Assign the new map to pm.pendingRequests
    streamManager.pendingRequests = newMap
}

func (s *Storage) AddPendingRequest(streamManager *PeerStreamManager, addr string, sampleID int, messageType MessageType) {

    s.pendingRequestMutex.Lock()
    defer s.pendingRequestMutex.Unlock()
	if _, ok := streamManager.pendingRequests[addr]; !ok {
		// If not, initialize the slice for the addr
		streamManager.pendingRequests[addr] = make(map[int]MessageType)
	}
	streamManager.pendingRequests[addr][sampleID] = messageType
}


func (s *Storage) GetSamplesIMissMap() map[int]struct{} {
	start := time.Now()
	s.samplesIMissMutex.RLock()
	stop := time.Now()
	s.samplesIMissTime += stop.Sub(start).Milliseconds()

	defer s.samplesIMissMutex.RUnlock()
	//creare a copy of sampleIMiss
	samplesIMissCopy := make(map[int]struct{})
	for k, v := range s.samplesIMiss {
		samplesIMissCopy[k] = v
	}
	return samplesIMissCopy
}

func (s *Storage) GetSamplesIHaveMap() map[int]struct{} {
	start := time.Now()
	s.samplesIHaveMutex.RLock()
	stop := time.Now()
	s.samplesIHaveTime += stop.Sub(start).Milliseconds()

	defer s.samplesIHaveMutex.RUnlock()
	//creare a copy of sampleIHave
	samplesIHaveCopy := make(map[int]struct{})
	for k, v := range s.samplesIHave {
		samplesIHaveCopy[k] = v
	}
	return samplesIHaveCopy
}
