package calc

import "math"

func NsToSeconds(ns int64) float64 {
	return math.Round(float64(ns)/(1000*1000*1000)*100) / 100
}
