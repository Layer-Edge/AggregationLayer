package da

import (
    "encoding/binary"
)

type Aggregator struct {
    data []byte
}

// Straight forward linear aggregation for now
func (aggr *Aggregator) Aggregate(data []byte) {
    bs := make([]byte, 2)
    binary.LittleEndian.PutUint16(bs, uint16(len(data)))
    aggr.data = append(aggr.data, bs...)
    aggr.data = append(aggr.data, data...)
}
