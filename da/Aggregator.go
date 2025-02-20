package da

import (
 "encoding/hex"
)

type Aggregator struct {
    data string
}

// Straight forward linear aggregation for now
func (aggr *Aggregator) Aggregate(data []byte) {
    if len(aggr.data) > 0 {
	    aggr.data = aggr.data + ","
    }
    	
    aggr.data = aggr.data + hex.EncodeToString(data)
}
