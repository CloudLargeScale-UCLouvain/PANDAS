package sample

import (
	"math/big"
	"math/rand"
)

// Block represents a block of samples in the network.
type Block struct {
	BlockSamples       [][]*Sample
	BlockID            int
	NumCols            int
	NumRows            int
	InterSampleGap     *big.Int // the gap between two consecutive sample in the ID space
	NetSize            int      // number of peers in the network
	IdByRowToSeqNumber map[string]int
	SeqNumberToIdByRow map[int]string
}

const NUM_ID_BITS = 256

var MaxKey *big.Int

func init() {
	MaxKey = new(big.Int).Lsh(big.NewInt(1), NUM_ID_BITS)
	MaxKey.Sub(MaxKey, big.NewInt(1))
}

// NewBlock creates a new Block instance with the given block ID.
func NewBlock(id int, numRows int, numCols int, netSize int) *Block {

	block := &Block{
		BlockID: id,
		NumRows: numRows,
		NumCols: numCols,
		NetSize: netSize,
	}
	block.IdByRowToSeqNumber = make(map[string]int)
	block.SeqNumberToIdByRow = make(map[int]string)
	numSamples := numRows * numCols
	block.InterSampleGap = new(big.Int).Div(MaxKey, big.NewInt(int64(numSamples)))

	block.BlockSamples = make([][]*Sample, block.NumRows)
	for i := 0; i < block.NumRows; i++ {
		block.BlockSamples[i] = make([]*Sample, block.NumCols)
	}

	for i := 0; i < block.NumRows; i++ {
		for j := 0; j < block.NumCols; j++ {
			s := NewSample(id, i+1, j+1, block)
			block.IdByRowToSeqNumber[s.IdByRow.String()] = s.SeqNumber
			block.SeqNumberToIdByRow[s.SeqNumber] = s.IdByRow.String()
			block.BlockSamples[i][j] = s
		}
	}

	return block
}

// ComputeRegionRadius calculates the radius of the region containing the desired number of copies of each sample.
func (b *Block) ComputeRegionRadius(numberOfCopiesPerSample int, numValidators int) *big.Int {
	radius := new(big.Int).Div(new(big.Int).Mul(MaxKey, big.NewInt(int64(numberOfCopiesPerSample))), big.NewInt(int64(numValidators*2)))
	return radius
}

// ComputeRegionRadiusWithNumNodes calculates the radius of the region containing the desired number of copies of each sample with the specified number of nodes.
func (b *Block) ComputeRegionRadiusWithNumNodes(numberOfCopiesPerSample int) *big.Int {
	radius := new(big.Int).Div(new(big.Int).Mul(MaxKey, big.NewInt(int64(numberOfCopiesPerSample))), big.NewInt(int64(b.NetSize*2)))
	return radius
}

// GetSamplesByRow returns the samples in a specific row.
func (b *Block) GetSamplesByRow(row int) []*Sample {
	return b.BlockSamples[row-1]
}

// GetSamplesByColumn returns the samples in a specific column.
func (b *Block) GetSamplesByColumn(column int) []*Sample {
	samples := make([]*Sample, b.NumRows)
	for i := 0; i < b.NumRows; i++ {
		samples[i] = b.BlockSamples[i][column-1]
	}
	return samples
}

// GetNRandomSamples returns n randomly selected samples.
func (b *Block) GetNRandomSamples(n int) []*Sample {
	samples := make([]*Sample, n)
	for i := 0; i < n; i++ {
		r := rand.Intn(b.NumRows)
		c := rand.Intn(b.NumCols)
		samples[i] = b.BlockSamples[r][c]
	}
	return samples
}

func (b *Block) GetRandomSample() *Sample {
	r := rand.Intn(b.NumRows)
	c := rand.Intn(b.NumCols)

	return b.BlockSamples[r][c]
}
