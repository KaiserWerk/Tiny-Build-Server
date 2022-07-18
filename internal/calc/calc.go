package calc

import "math"

func MsToSeconds(ms int64) float64 {
	return math.Round(float64(ms)/(1000*1000*1000)*100) / 100
}
