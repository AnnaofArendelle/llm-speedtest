package metrics

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestQuantile(t *testing.T) {
	if Quantile(nil, 0.5) != 0 {
		t.Fatal("empty quantile should be 0")
	}
	if Quantile([]float64{7}, 0.9) != 7 {
		t.Fatal("single sample")
	}
	xs := []float64{5, 1, 3, 2, 4} // unsorted
	if !near(Median(xs), 3) {
		t.Fatalf("median = %v", Median(xs))
	}
	if !near(Quantile(xs, 0.25), 2) || !near(Quantile(xs, 0.75), 4) {
		t.Fatalf("quartiles = %v %v", Quantile(xs, 0.25), Quantile(xs, 0.75))
	}
	if !near(Quantile([]float64{1, 2}, 0.5), 1.5) {
		t.Fatal("interpolated median")
	}
	if !near(Quantile(xs, 0), 1) || !near(Quantile(xs, 1), 5) || !near(Quantile(xs, 2), 5) {
		t.Fatal("bounds")
	}
}

func TestSummarise(t *testing.T) {
	st := Summarise([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if st.Count != 8 || st.Min != 2 || st.Max != 9 || !near(st.Mean, 5) || !near(st.Median, 4.5) {
		t.Fatalf("stats = %+v", st)
	}
	if !near(st.StdDev, math.Sqrt(32.0/7.0)) {
		t.Fatalf("stddev = %v", st.StdDev)
	}
	if Summarise(nil).Count != 0 {
		t.Fatal("empty")
	}
}

func TestLogScoreAndRound(t *testing.T) {
	if LogScore(0.5, 0.5, 30) != 100 || LogScore(30, 0.5, 30) != 0 || LogScore(100, 0.5, 30) != 0 {
		t.Fatal("bounds")
	}
	mid := LogScore(1, 0.5, 30)
	if mid < 82 || mid > 84 {
		t.Fatalf("1s score = %v", mid)
	}
	if LogScore(1, 0, 30) != 0 || LogScore(1, 30, 0.5) != 0 {
		t.Fatal("invalid bounds")
	}
	if Round(1.23456, 3) != 1.235 || Round(2.5, 0) != 3 {
		t.Fatal("round")
	}
	if Ratio(1, 0) != 0 || !near(Ratio(1, 4), 0.25) {
		t.Fatal("ratio")
	}
}
