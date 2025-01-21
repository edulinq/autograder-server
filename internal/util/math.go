package util

import (
	"math"
	"slices"
	"strconv"

	"github.com/edulinq/autograder/internal/log"
)

const EPSILON = 1e-5

type AggregateValues struct {
	Count  int     `json:"count"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

func IsClose(a float64, b float64) bool {
	return math.Abs(a-b) < EPSILON
}

func IsZero(a float64) bool {
	return IsClose(a, 0.0)
}

func FloatToStr(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// If the value is NaN, return the default.
func DefaultNaN(value float64, defaultValue float64) float64 {
	if math.IsNaN(value) {
		return defaultValue
	}

	return value
}

func MustStrToFloat(value string) float64 {
	result, err := StrToFloat(value)
	if err != nil {
		log.Fatal("Failed to convert string to float.", err, log.NewAttr("value", value))
	}

	return result
}

func StrToFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func MinMax(values []float64) (float64, float64) {
	var min float64
	var max float64

	for i, value := range values {
		if (i == 0) || (value < min) {
			min = value
		}

		if (i == 0) || (value > max) {
			max = value
		}
	}

	return min, max
}

func Median(values []float64) float64 {
	slices.Sort(values)

	length := len(values)
	if length == 0 {
		return 0.0
	} else if length%2 == 0 {
		return (values[length/2] + values[(length/2)-1]) / 2.0
	} else {
		return values[length/2]
	}
}

// Input values will be sorted (in place).
func ComputeAggregates(values []float64) AggregateValues {
	if (values == nil) || (len(values) == 0) {
		return AggregateValues{}
	}

	median := Median(values)
	min := 0.0
	max := 0.0
	mean := 0.0

	for i, value := range values {
		if (i == 0) || (value < min) {
			min = value
		}

		if (i == 0) || (value > max) {
			max = value
		}

		mean += value
	}

	mean /= float64(len(values))

	return AggregateValues{
		Count:  len(values),
		Mean:   mean,
		Median: median,
		Min:    min,
		Max:    max,
	}
}

func (this AggregateValues) Equals(other AggregateValues) bool {
	return ((this.Count == other.Count) &&
		IsClose(this.Mean, other.Mean) &&
		IsClose(this.Median, other.Median) &&
		IsClose(this.Min, other.Min) &&
		IsClose(this.Max, other.Max))
}

// Round a float to |precision| number of decimal places,
// e.g., RoundWithPrecision(3.14159, 2) == 3.14.
func RoundWithPrecision(value float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(value*ratio) / ratio
}
