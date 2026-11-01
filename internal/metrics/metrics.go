// Package metrics provides small, deterministic statistics helpers shared by
// the benchmark engines and the analysis layer.
package metrics

import (
	"math"
	"sort"
)

// Stats summarises a sample of non-negative measurements (seconds or tokens).
type Stats struct {
	Count  int     `json:"count"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P75    float64 `json:"p75"`
	P90    float64 `json:"p90"`
	P95    float64 `json:"p95"`
	P99    float64 `json:"p99"`
	StdDev float64 `json:"stddev"`
}

// Quantile returns the q-quantile (0..1) of xs using linear interpolation
// between closest ranks (the R-7 / NumPy default). xs need not be sorted.
// It returns 0 for an empty sample.
func Quantile(xs []float64, q float64) float64 {
	n := len(xs)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return xs[0]
	}
	s := make([]float64, n)
	copy(s, xs)
	sort.Float64s(s)
	return quantileSorted(s, q)
}

func quantileSorted(s []float64, q float64) float64 {
	n := len(s)
	if n == 0 {
		return 0
	}
	if q <= 0 {
		return s[0]
	}
	if q >= 1 {
		return s[n-1]
	}
	pos := q * float64(n-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return s[lo]
	}
	frac := pos - float64(lo)
	return s[lo] + (s[hi]-s[lo])*frac
}

// Median returns the median of xs (0 for an empty sample).
func Median(xs []float64) float64 { return Quantile(xs, 0.5) }

// Summarise computes descriptive statistics for xs.
func Summarise(xs []float64) Stats {
	st := Stats{Count: len(xs)}
	if len(xs) == 0 {
		return st
	}
	s := make([]float64, len(xs))
	copy(s, xs)
	sort.Float64s(s)
	st.Min = s[0]
	st.Max = s[len(s)-1]
	var sum float64
	for _, v := range s {
		sum += v
	}
	st.Mean = sum / float64(len(s))
	st.Median = quantileSorted(s, 0.5)
	st.P75 = quantileSorted(s, 0.75)
	st.P90 = quantileSorted(s, 0.90)
	st.P95 = quantileSorted(s, 0.95)
	st.P99 = quantileSorted(s, 0.99)
	if len(s) > 1 {
		var ss float64
		for _, v := range s {
			d := v - st.Mean
			ss += d * d
		}
		st.StdDev = math.Sqrt(ss / float64(len(s)-1))
	}
	return st
}

// Round rounds x to the given number of decimals; it is used before sorting
// so that float noise never changes an ordering.
func Round(x float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(x*p) / p
}

// Ratio returns num/den, or 0 when den is 0.
func Ratio(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
}

// LogScore maps x onto 0..100 using a logarithmic scale where x <= lo scores
// 100 and x >= hi scores 0. It is used for the speed component of the
// composite ranking.
func LogScore(x, lo, hi float64) float64 {
	if lo <= 0 || hi <= lo {
		return 0
	}
	if x <= lo {
		return 100
	}
	if x >= hi {
		return 0
	}
	return 100 * (math.Log(hi) - math.Log(x)) / (math.Log(hi) - math.Log(lo))
}
