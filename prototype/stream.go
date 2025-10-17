package main

import (
	"bufio"
	"container/list"
	"context"
    "errors"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	proto "github.com/libp2p/go-libp2p/core/protocol"
)

type PeerStreamManager struct {
	host            host.Host
	myAddr          string
	myID            *big.Int
	streamMap       map[peer.ID]network.Stream
	mu              sync.RWMutex // Mutex for synchronization
	protocol        string
	pendingRequests map[string]map[int]MessageType
	streamList      *list.List                //LRU ordering of streams
	listElementMap  map[peer.ID]*list.Element // Map to store list elements corresponding to peerIDs
	openConnections int                       // Maximum number of open connections
	connectTimeout  time.Duration
	sendMutexMap    map[peer.ID]*sync.Mutex // mutex to block concurrent sends to same peer
}

//var sendMutex sync.Mutex

func NewPeerStreamManager(nick string, role string, h host.Host, my_multiaddr string, proto string, openConnections int, timeout time.Duration) *PeerStreamManager {
	id := MaddrToPeerID(my_multiaddr, nick, role)

	return &PeerStreamManager{
		host:            h,
		myAddr:          my_multiaddr,
		myID:            id,
		streamMap:       make(map[peer.ID]network.Stream),
		sendMutexMap:    make(map[peer.ID]*sync.Mutex),
		listElementMap:  make(map[peer.ID]*list.Element),
		protocol:        proto,
		pendingRequests: make(map[string]map[int]MessageType),
		streamList:      list.New(),
		openConnections: openConnections,
		connectTimeout:  timeout,
	}
}

func (pm *PeerStreamManager) AddPendingRequest(addr string, sampleID int, messageType MessageType) {
	s.AddPendingRequest(pm, addr, sampleID, messageType)
}

func (pm *PeerStreamManager) GetOrCreateStream(ctx context.Context, peerInfo *peer.AddrInfo) (network.Stream, error) {
	/*if peerInfo.ID.String() == pm.host.ID().String() {
		log.Panicln("Attempting to create a stream to itself")
	}*/
    if peerInfo.ID.String() == pm.host.ID().String() {
        log.Printf("Attempting to create a stream to itself")
        return nil, errors.New("cannot create a stream to itself")
    }

	stream := pm.lookupStream(peerInfo.ID)
	if stream != nil {
		return stream, nil
	}

	// Attempt to connect to the peer
	err := pm.host.Connect(ctx, *peerInfo)
	if err != nil {
		return nil, err
	}

	// Attempt to create a new stream
	stream, err = pm.host.NewStream(ctx, peerInfo.ID, proto.ID(pm.protocol))
	if err != nil {
		return nil, err
	}

	// Add the new stream to the map and list
	pm.addStream(peerInfo.ID, stream)

	pm.initSendMutex(peerInfo.ID)

	return stream, nil
}

func (pm *PeerStreamManager) initSendMutex(peerID peer.ID) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if the mutex for this peerID is already initialized
	if _, ok := pm.sendMutexMap[peerID]; !ok {
		pm.sendMutexMap[peerID] = &sync.Mutex{}
	}
}

func (pm *PeerStreamManager) lookupStream(peerID peer.ID) network.Stream {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	// Check if the peerID exists in the streamMap
	if stream, ok := pm.streamMap[peerID]; ok {
		// Move the corresponding element in the list to the front (recently used)
		if elem, ok := pm.listElementMap[peerID]; ok {
			pm.streamList.MoveToFront(elem)
		}
		return stream
	}
	return nil
}

func (pm *PeerStreamManager) removeStream(peerID peer.ID) {

	if elem, exists := pm.listElementMap[peerID]; exists {
		pm.streamList.Remove(elem)
		stream, _ := pm.streamMap[peerID]
		delete(pm.streamMap, peerID)
		log.Printf("Closing the stream with %s", peerID)
		err := stream.Close()
		if err != nil {
			log.Println("Error closing the stream ", err)
		}
		delete(pm.listElementMap, peerID)
	} else {
		log.Printf("Stream does not exist in removeStream")
	}
}

func (pm *PeerStreamManager) addStream(peerID peer.ID, stream network.Stream) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	//log.Printf("Stream list length: %d\n", pm.streamList.Len())
	if pm.streamList.Len() >= pm.openConnections {
		// If the cache is full, remove the least recently used stream
		back := pm.streamList.Back()
		if back != nil {
			// Remove the least recently used stream
			pm.removeStream(back.Value.(peer.ID))
		}
	}
	// Add the new stream to the front of the list (recently used)
	elem := pm.streamList.PushFront(peerID)
	pm.streamMap[peerID] = stream
	pm.listElementMap[peerID] = elem
}

func (pm *PeerStreamManager) sendMessageToPeer(msg []byte, peerInfo *peer.AddrInfo) {

	ctx := context.Background()
	stream, err := pm.GetOrCreateStream(ctx, peerInfo)
	if err == nil {
		err = pm.sendMessage(stream, msg, peerInfo.ID)
		if err == nil {
			log.Printf("Message sent")
		} else {
			log.Println("Error sending message: ", err)
		}
	} else {
		log.Println("Error creating a stream: ", err)
	}
}

// Send a message over a stream
func (pm *PeerStreamManager) sendMessage(stream network.Stream, msg []byte, peerID peer.ID) error {

	// Lock the mutex for the peer.ID in the sendMutexMap
	pm.lockSendMutex(peerID)
	defer pm.unlockSendMutex(peerID)

	w := bufio.NewWriter(stream)
	err := writeBinaryData(w, msg)
	if err != nil {
		log.Println("Error Sending Data:", err)
	}
	return err
}

func (pm *PeerStreamManager) lockSendMutex(peerID peer.ID) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Check if the mutex for this peerID exists
	if mutex, ok := pm.sendMutexMap[peerID]; ok {
		mutex.Lock()
	}
}

func (pm *PeerStreamManager) unlockSendMutex(peerID peer.ID) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Check if the mutex for this peerID exists
	if mutex, ok := pm.sendMutexMap[peerID]; ok {
		mutex.Unlock()
	}
}
