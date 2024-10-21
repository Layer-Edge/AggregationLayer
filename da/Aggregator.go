package da

type Aggregator struct {
    data []byte
}

// Straight forward linear aggregation for now
func (aggr *Aggregator) Aggregate(data []byte) {
    aggr.data = append(aggr.data, data...)
}
