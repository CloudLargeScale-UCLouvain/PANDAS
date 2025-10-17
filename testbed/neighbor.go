package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Neighbor represents a node neighbor in the network.
type Neighbor struct {
	Id         *big.Int
	LastSeen   time.Time
	Role       string
	IsEvil     bool
	Addr       ma.Multiaddr
	Ip         string
	Port       int
	addrUDPStr string
	PeerInfo   *peer.AddrInfo
}

func MaddrToPeerID(ma string, nick string, role string) *big.Int {
	id := new(big.Int)
    /*
	if role == "validator" {
		offset := new(big.Int)
		max_val := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
		max_val = max_val.Sub(max_val, big.NewInt(1))
		divisor := big.NewInt(int64(len(validatorPositions)))
		offset.Div(max_val, divisor)
		id.Mul(offset, big.NewInt(int64(validatorPositions[nick])))
	} else {*/
		data := []byte(ma)
		hash := sha256.Sum256(data)
		id.SetBytes(hash[:])
	//}
	return id
}

func NewNeighbour(nick string, id_s string, addr ma.Multiaddr, role string, isEvil bool, ip string, port int) *Neighbor {
	id := MaddrToPeerID(id_s, nick, role)
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Println("Error creating peer.AddrInfo:", addr, err)
		return nil
	}
	udpAddr := fmt.Sprintf("%s:%d", ip, port)

	return &Neighbor{
		Id:         id,
		LastSeen:   time.Now(),
		Role:       role,
		IsEvil:     isEvil,
		Addr:       addr,
		Ip:         ip,
		Port:       port,
		addrUDPStr: udpAddr,
		PeerInfo:   peerInfo,
	}
}

func (n *Neighbor) AddrUDP() string {
	addr := fmt.Sprintf("%s:%d", n.Ip, n.Port)
	return addr
}

// NewNeighbour creates a new Neighbour instance with the given parameters.
/*
func NewNeighbour(id_s string, addr ma.Multiaddr, role string, isEvil bool) *Neighbor {
	//FIXME: it was just easier to hash the ID for now :)
	id := MaddrToPeerID(id_s)

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Println("Error creating peer.AddrInfo:", addr, err)
		return nil
	}
	return &Neighbor{
		Id:       id,
		LastSeen: time.Now(),
		Addr:     addr,
		Role:     role,
		IsEvil:   isEvil,
		PeerInfo: peerInfo,
	}
}*/

func (n *Neighbor) String() string {
	return n.Id.String() + " addr:" + n.AddrUDP() + ", role:" + n.Role
}

// UpdateLastSeen updates the last seen time of the Neighbour.
func (n *Neighbor) UpdateLastSeen() {
	n.LastSeen = time.Now()
}

// Expired checks if the Neighbour has expired based on the TTL.
func (n *Neighbor) Expired(ttl time.Duration) bool {
	return time.Since(n.LastSeen) >= ttl
}

// CompareTo compares two Neighbours based on their last seen time.
func (n *Neighbor) IsFresherThan(other *Neighbor) bool {
	if n.LastSeen.Before(other.LastSeen) {
		return false
	} else {
		return true
	}
}

// Equals checks if the Neighbour is equal to another object.
func (n *Neighbor) Equals(other interface{}) bool {
	if otherNeighbour, ok := other.(*Neighbor); ok {
		return n.Id.Cmp(otherNeighbour.Id) == 0
	}
	return false
}
