package sample

import (
	"math/big"
)

// Sample represents a sample in the DAS system.
type Sample struct {
	Data        []byte
	Row, Column int
	SeqNumber   int
	BlockID     int
	IdByRow     *big.Int
	IdByColumn  *big.Int
	Block       *Block
}

// NewSample creates a new Sample instance.
func NewSample(blockID int, row, column int, b *Block) *Sample {
	s := &Sample{
		BlockID:    blockID,
		Row:        row,
		Column:     column,
		SeqNumber:  (row-1)*b.NumCols + (column - 1),
		Block:      b,
		IdByRow:    nil,
		IdByColumn: nil,
		Data:       nil,
	}
	s.computeID()
	return s
}

// SampleNumberByRow calculates the sample number by row.
func (s *Sample) SampleNumberByRow() int {
	return (s.Row-1)*s.Block.NumCols + (s.Column - 1)
}

// SampleNumberByColumn calculates the sample number by column.
func (s *Sample) SampleNumberByColumn() int {
	return (s.Column-1)*s.Block.NumRows + (s.Row - 1)
}

// ComputeID maps the sample to the DHT keyspace.
func (s *Sample) computeID() {
	s.IdByRow = new(big.Int).Set(
		new(big.Int).Add(
			new(big.Int).Mul(
				s.Block.InterSampleGap,
				new(big.Int).SetInt64(int64(s.SampleNumberByRow())),
			),
			new(big.Int).SetInt64(int64(s.Block.BlockID)),
		),
	)

	s.IdByColumn = new(big.Int).Set(
		new(big.Int).Add(
			new(big.Int).Mul(
				s.Block.InterSampleGap,
				new(big.Int).SetInt64(int64(s.SampleNumberByColumn())),
			),
			new(big.Int).Add(new(big.Int).SetInt64(int64(s.Block.BlockID)), big.NewInt(1)),
		),
	)
}

// IsInRegion checks if the sample falls within the region of the node.
func (s *Sample) IsInRegion(peerID, radius *big.Int) bool {

	lowerBound := new(big.Int).Sub(peerID, radius)
	upperBound := new(big.Int).Add(peerID, radius)
	if lowerBound.Cmp(big.NewInt(0)) < 0 {
		lowerBound.Add(MaxKey, lowerBound)
	}
	if upperBound.Cmp(MaxKey) > 0 {
		upperBound.Sub(upperBound, MaxKey)
	}
	return s.IsInRegionByRow(lowerBound, upperBound) || s.IsInRegionByColumn(lowerBound, upperBound)
	//return s.IsInRegionByRow(lowerBound, upperBound)
}

// IsInRegionByColumn checks if the sample falls within the region of the node using the column ID.
func (s *Sample) IsInRegionByColumn(lowerBound, upperBound *big.Int) bool {

	if lowerBound.Cmp(upperBound) < 0 {
		return s.IdByColumn.Cmp(lowerBound) >= 0 && s.IdByColumn.Cmp(upperBound) <= 0
	} else {
		return s.IdByColumn.Cmp(lowerBound) >= 0 || s.IdByColumn.Cmp(upperBound) <= 0
	}
}

// IsInRegionByRow checks if the sample falls within the region of the node using the row ID.
func (s *Sample) IsInRegionByRow(lowerBound, upperBound *big.Int) bool {

	if lowerBound.Cmp(upperBound) < 0 {
		return s.IdByRow.Cmp(lowerBound) >= 0 && s.IdByRow.Cmp(upperBound) <= 0
	} else {
		return s.IdByRow.Cmp(lowerBound) >= 0 || s.IdByRow.Cmp(upperBound) <= 0
	}
}

// GetIDByRow returns the computed identifier of the sample using rows as a reference.
func (s *Sample) GetIDByRow() *big.Int {
	return s.IdByRow
}

// GetIDByColumn returns the computed identifier of the sample using columns as a reference.
func (s *Sample) GetIDByColumn() *big.Int {
	return s.IdByColumn
}
