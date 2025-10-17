package main

import (
	//"context"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p-das/sample"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	//QUIC implementations:
	//"github.com/quic-go/quic-go"
)

const BLOCK_DIM = 512
const NUM_SAMPLE_COPIES = 1
const NUM_RANDOM_SAMPLES = 73 // number of random samples
const NUM_ROWS_TO_SAMPLE = 2  // number of columns to sample
const NUM_COLS_TO_SAMPLE = 2  // number of rows to sample
const BLOCK_TIME = 12         // block time = 12 seconds
const MAX_SAMPLES_PER_PACKET = 20
const MAX_SAMPLE_REQUESTS_PER_PACKET = 70
const VOTING_DEADLINE = 4

// Global variables
type Config struct {
	IP                 string
	ListenPort         int
	ListenUDPPort      int
	ProtocolID         string
	NickFlag           string
	NodeType           string
	ExperimentDuration int
	DebugMode          bool
	PerfMode           bool
	Key                string
	PeerID             string
	LogDirectory       string
	NodeFile           string
	OpenConnLimit      int
	ConnectTimeout     int
    EnableHeaderDis   bool
}
var config Config
var searchTable *SearchTable
var randomSamplingStarted bool = false
var blockID int = -1
var myself *Neighbor
var currBlock *sample.Block = nil
var myUDPAddr string = ""
var BANDWIDTH_LIMIT_BYTES_PER_SEC = 125000000 // 1Gbps in bytes per second

// used to calculate uniform validatorIDs
var validatorPositions map[string]int

// cached UDP connections
var udpConnCache = make(map[string]*net.UDPConn)

// Map to optimise peer to samples mapping (NOTE: assumes that mapping is static across blocks)
var peerToSamples = make(map[string][]int)
var peerToSamplesBigInt = make(map[string][]big.Int)

// Cached samples that validators/regular nodes store
var sampleCacheByRow map[string]*sample.Sample
var sampleCacheByColumn map[string]*sample.Sample

var events []string

var eventMutex sync.Mutex
var s = NewStorage()

func addEvent(event string) {
	eventMutex.Lock()
	events = append(events, event)
	eventMutex.Unlock()
}

//var eventLogger *log.Logger

func main() {
	//========== Experiment arguments ==========
	config = Config{}
	flag.StringVar(&config.IP, "ip", "127.0.0.1", "IP to bind to.")
	flag.StringVar(&config.NickFlag, "nick", "", "nickname for node")
	flag.StringVar(&config.NodeType, "nodeType", "builder", "type of node: builder, nonvalidator, validator")
	flag.BoolVar(&config.DebugMode, "debug", false, "debug mode")
	flag.BoolVar(&config.PerfMode, "pref", false, "perf")
    flag.BoolVar(&config.EnableHeaderDis, "HeaderDis", false, "enable (false) or disable (true) gossip header dissemination")

    flag.IntVar(&config.ExperimentDuration, "duration", 80, "Experiment duration (in seconds).")
    flag.StringVar(&config.ProtocolID, "pid", "/pandas/0.1", "Sets a protocol id for stream headers")
    flag.IntVar(&config.ListenPort, "port", 9000, "Specifies a port number to listen")
    flag.IntVar(&config.ListenUDPPort, "UDPport", 12000, "Specifies a port number to listen")
    flag.StringVar(&config.Key, "key", "", "Specify key file to use for the node")
    flag.StringVar(&config.LogDirectory, "log", "./log/", "log directory")
    flag.StringVar(&config.NodeFile, "node", "./nodes.csv", "node directory")
    flag.IntVar(&config.OpenConnLimit, "connLimit", 5000, "Limit on the open connections")
    flag.IntVar(&config.ConnectTimeout, "connTimeout", 30, "Timeout in connection retries in msec")
    flag.Parse()

	log.SetPrefix(config.NickFlag + ": ")
	log.SetFlags(log.Lmicroseconds) //print time in microseconds
	//ctx := context.Background()
	log.Printf("Running PANDAS, nickname: %s, type: %s, ip: %s/%d\n", config.NickFlag, config.NodeType, config.IP, config.ListenPort)

	//Generate the block – for now pass network size = 1, but update this later
	currBlock = sample.NewBlock(blockID, BLOCK_DIM, BLOCK_DIM, 1)

	if !(config.DebugMode) {
		log.SetOutput(io.Discard)
	}
	//log.SetOutput(io.Discard)
	//========== Initialise pubsub service ==========
	// create a new libp2p Host

	messageChannel := make(chan Message, 20000)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Create host
	go makeHostUDP(config, ctx, messageChannel)

	validatorPositions = countValidators(config.NodeFile, config.NickFlag)
	//FIXME this is a hack to get the peer ID
	//I wanted to always pass through NewNeigbour in case we change the method of computing the peerID
	h, err := makeHost(config, ctx)
	if err != nil {
		log.Println(err)
		return
	}
	multiAddrStr := fmt.Sprintf("%s/p2p/%s", h.Addrs()[0], h.ID().String())
	myUDPAddr = fmt.Sprintf("%s:%d", config.IP, config.ListenUDPPort)
	multiAddr, err := multiaddr.NewMultiaddr(multiAddrStr)
	if err != nil {
		log.Println("Error parsing peer address:", err)
		return
	}
	//myself = NewNeighbour(h.ID().String(), multiAddr, config.NodeType, false)
	myself = NewNeighbour(config.NickFlag, h.ID().String(), multiAddr, config.NodeType, false, config.IP, config.ListenUDPPort)
	log.Printf("my own ID:%s, multiaddr: %s\n", myself.Id, multiAddr)

	//global and defined in node.go
	searchTable = NewSearchTable(myself)

	log.Println("Adding peers from file")
	for _, peer := range readPeersFromFile(config.NodeFile) {
		log.Println(peer)
		searchTable.AddNeighbor(peer)
	}

	// Set up a stream handler with a closure that has access to the messageChannel
	h.SetStreamHandler(protocol.ID(config.ProtocolID), func(stream network.Stream) {
		handleStream(stream, messageChannel)
	})
	streamManager := NewPeerStreamManager(config.NickFlag, config.NodeType, h, multiAddrStr, config.ProtocolID, config.OpenConnLimit, time.Duration(config.ConnectTimeout)*time.Millisecond)
	//log.Printf("%s waiting for 1s for other nodes to get up ", h.ID().String())

	calculateUnhostedSamples(streamManager.myID)
	go pingRandomPeers(streamManager)

	//channel used to wait for goroutines handling events
	finished := make(chan bool)

	//Start time for load metrics
	if config.NodeType == "builder" {
		time.Sleep(200 * time.Millisecond)

		log.Printf("Builder here...\n")

		go handleEventsBuilder(config.ExperimentDuration, h, messageChannel, streamManager, ctx, finished, config.LogDirectory, config.NickFlag)
		<-finished

	} else if config.NodeType == "regular" {
		time.Sleep(4 * time.Second)

		log.Printf("Regular node here...\n")

		go handleEventsRegular(config.ExperimentDuration, h, messageChannel, streamManager, ctx, finished, config.LogDirectory, config.NickFlag)
		<-finished
	} else if config.NodeType == "validator" {
		time.Sleep(1 * time.Second)

		log.Printf("Validator node here...\n")

		go handleEventsValidator(config.ExperimentDuration, h, messageChannel, streamManager, ctx, finished, config.LogDirectory, config.NickFlag)
		<-finished
	} else {
		log.Printf("Error: Invalid node type..\n")
	}
	//========== Initialise Logger ==========
	//Create Log file
	file, err := os.OpenFile(config.LogDirectory+config.NickFlag+".log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer file.Close()
	for event := range events {
		_, err = file.WriteString(events[event] + "\n")
		if err != nil {
			log.Fatal("Error writing to log file:", err)
		}
	}
	log.Println(s.PrintTimeSpentOnMutexes())
	log.Println("Main done - shutting down")
}

func defaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

func shortID(p peer.ID) string {
	pretty := p.ShortString()
	return pretty[len(pretty)-8:]
}

/*
func getPeerIDFromKey(config Config) (peer.ID, error) {
    // Get the RSA key for the node.
    privBytes, err := os.ReadFile(config.Key)
    priv, err := crypto.UnmarshalPrivateKey(privBytes)
    if err != nil {
        return nil, err
    }

    // Get peer ID from public key
    peerID, err := peer.IDFromPublicKey(libp2pPrivKey.GetPublic())
    if err != nil {
        fmt.Println("Error generating peer ID:", err)
        return nil, err
    }

    return peerID, nil

}*/

func makeHost(config Config, ctx context.Context) (host.Host, error) {
	// Get the RSA key for the node.
	privBytes, err := os.ReadFile(config.Key)
	priv, err := crypto.UnmarshalPrivateKey(privBytes)
	if err != nil {
		panic(err)
	}

	tcpAddrString := fmt.Sprintf("/ip4/%s/tcp/%d", config.IP, config.ListenPort)
	//quicAddrString := fmt.Sprintf("/ip4/%s/udp/%d/quic", config.IP, config.ListenPort)

	/*
	   connmgr, err := connmgr.NewConnManager(
	       5000,    // Lowwater
	       3000000, // HighWater,
	       connmgr.WithGracePeriod(10*time.Minute),
	   )*/

	h, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			tcpAddrString, // regular tcp connections
			//quicAddrString, // a UDP endpoint for the QUIC transport
		),
		libp2p.DisableMetrics(),
		libp2p.NoSecurity,
		libp2p.DisableRelay(),
		//libp2p.ConnectionManager(connmgr),
		//libp2p.NATPortMap(),
		//libp2p.EnableNATService(),
	)

	/*
	   h, err := libp2p.New(
	       // Use the keypair we generated
	       libp2p.Identity(priv),
	       // Multiple listen addresses
	       libp2p.ListenAddrStrings(
	            tcpAddrString, // regular tcp connections
	       //     //quicAddrString, // a UDP endpoint for the QUIC transport
	       ),
	       // libp2p.NATPortMap(),
	       // libp2p.EnableNATService(),
	       libp2p.ConnectionManager(connmgr.NewConnManager(
	           -1,       // MaxConcurrentStreamsPerConnection: Unlimited concurrent streams per connection
	           -1,       // MaxConnections: Unlimited connections
	           0,        // GracePeriod: No grace period (immediate removal of idle connections)
	       )),
	   )*/

	if err != nil {
		return nil, err
	}

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	return h, err
}

func makeHostUDP(config Config, ctx context.Context, messageChannel chan<- Message) error {
	addr := net.JoinHostPort(config.IP, strconv.Itoa(config.ListenUDPPort))

	// Start UDP server
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Println("Error listening: ", err)
		return err
	}
	defer conn.Close()

	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read incoming messages
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				log.Println("Error receiving message:", err)
				continue
			}

			// Make a copy of the message data
			completeMessage := make([]byte, n)
			copy(completeMessage, buffer[:n])

			// Check if the complete message is not empty
			/*
			   if len(completeMessage) == 0 {
			       continue
			   }*/

			// Handle the complete message
			go handleMessageUDP(completeMessage, messageChannel)
		}
	}
}
