package main

import (
	"log"
	"math/big"
	"sort"
	"time"

	"github.com/libp2p-das/sample"
)

// SearchTable represents a table for managing neighbours.
type SearchTable struct {
	myself     *Neighbor            // The node itself
	builder    *Neighbor            // The builder node
	validators map[string]*Neighbor // Map of validators with ID as the key
	neighbors  map[string]*Neighbor // Map of nodes (validators + regulars) with ID as the key
}

// NewSearchTable creates a new SearchTable instance.
func NewSearchTable(myself *Neighbor) *SearchTable {
	return &SearchTable{
		myself:     myself,
		validators: make(map[string]*Neighbor),
		neighbors:  make(map[string]*Neighbor),
	}
}

// GetFreshestKNeighbours returns up to K neighbours with the most recent LastSeen value.
func (st *SearchTable) GetFreshestKNeighbours(k int) []*Neighbor {
	neighbours := make([]*Neighbor, 0, len(st.neighbors))
	for _, n := range st.neighbors {
		neighbours = append(neighbours, n)
	}

	sort.Slice(neighbours, func(i, j int) bool {
		return neighbours[i].LastSeen.After(neighbours[j].LastSeen)
	})

	if k < len(neighbours) {
		return neighbours[:k]
	}
	return neighbours
}

// GetNeighborsBySample returns neighbours within the specified radius of the sampleId.
func (st *SearchTable) GetNeighborsBySample(sampleId *big.Int, radius *big.Int) []*Neighbor {
	neighbours := make([]*Neighbor, 0, len(st.neighbors))

	// Calculate the lower bound for comparison
	lowerBound := new(big.Int).Sub(sampleId, radius)
	if lowerBound.Cmp(big.NewInt(0)) < 0 {
		lowerBound.SetUint64(0) // Set to 0 if lowerBound is negative
	}

	// Calculate the upper bound for comparison
	upperBound := new(big.Int).Add(sampleId, radius)
	if upperBound.Cmp(sample.MaxKey) > 0 {
		upperBound.Set(sample.MaxKey) // Set to MAX_KEY if upperBound exceeds MAX_KEY
	}

	for _, n := range st.neighbors {
		if n.Id.Cmp(lowerBound) >= 0 && n.Id.Cmp(upperBound) <= 0 {
			neighbours = append(neighbours, n)
		}
	}

	return neighbours
}

// AddNeighbor adds a neighbour to the SearchTable, avoiding duplicates.
func (st *SearchTable) AddNeighbor(n *Neighbor) {
    /*
	if n.Id.String() == st.myself.Id.String() {
	    log.Printf("Attempting to add myself to the search table - ignoring")
	    return
	}*/
	switch n.Role {
	case "builder":
		//log.Println("Adding builder node:", n)
		st.builder = n
	case "validator":
		//log.Println("Adding validator node:", n)
		//add to validators
		if existing, ok := st.validators[n.Id.String()]; ok {
			if n.IsFresherThan(existing) {
				existing.UpdateLastSeen()
			}
		} else {
			st.validators[n.Id.String()] = n
		}
		//add to neighbors
		if existing, ok := st.neighbors[n.Id.String()]; ok {
			if n.IsFresherThan(existing) {
				existing.UpdateLastSeen()
			}
		} else {
			st.neighbors[n.Id.String()] = n
		}
	case "regular":
		//log.Println("Adding regular node:", n)
		if existing, ok := st.neighbors[n.Id.String()]; ok {
			if n.IsFresherThan(existing) {
				existing.UpdateLastSeen()
			}
		} else {
			st.neighbors[n.Id.String()] = n
		}
	case "bootstrap":

		log.Println("Trying to add a bootstra nodes:", n, " - ignoring")
		//do nothing
	default:
		log.Panicf("Unknown role: %s", n.Role)

	}

}

// RemoveExpired removes neighbours that have expired.
func (st *SearchTable) RemoveExpired(ttl time.Duration) {
	for id, n := range st.neighbors {
		if n.Expired(ttl) {
			delete(st.neighbors, id)
		}
	}
}
