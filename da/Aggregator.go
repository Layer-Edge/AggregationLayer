package da

type Aggregator struct {
	data string
}

// Straight forward linear aggregation for now
func (aggr *Aggregator) Aggregate(data string) {
	if len(aggr.data) > 0 {
		aggr.data = aggr.data + ","
	}

	aggr.data = aggr.data + data
}
