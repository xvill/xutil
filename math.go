package xutil

import "math"

// Sum 和
func Sum(arr []float64) (sum float64) {
	for _, v := range arr {
		sum = sum + v
	}
	return
}

// Mean 均值
func Mean(arr []float64) float64 {
	var N, sum float64
	for _, v := range arr {
		N++
		sum = sum + v
	}
	return sum / N
}

func DiffSqrtMean(a []float64) []float64 {
	set := make([]float64, len(a))
	meanVal := Mean(a)
	var d float64
	for i, v := range a {
		d = v - meanVal
		set[i] = d * d
	}
	return set
}

func StdDev(a []float64) float64 {
	return math.Sqrt(Mean(DiffSqrtMean(a)))
}
