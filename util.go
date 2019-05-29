package inflight

import "math"

func min(a, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}

func ACS(pl PriorityBand, queues []*Queue) int {
	assuredconcurrencyshares := 0
	for _, queue := range queues {
		if queue.Priority == pl {
			assuredconcurrencyshares += queue.SharedQuota
		}
	}
	return assuredconcurrencyshares
}

func ACV(pl PriorityBand, queues []*Queue) int {
	// ACV(l) = ceil( SCL * ACS(l) / ( sum[priority levels k] ACS(k) ) )
	denom := 0
	for _, prioritylvl := range Priorities {
		denom += ACS(prioritylvl, queues)
	}
	return int(math.Ceil(float64(SCL * ACS(pl, queues) / denom)))
}
