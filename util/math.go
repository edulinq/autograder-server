package util

import (
    "math"
    "slices"
    "strconv"

    "github.com/rs/zerolog/log"
)

const EPSILON = 1e-5

func IsClose(a float64, b float64) bool {
    return math.Abs(a - b) < EPSILON;
}

func IsZero(a float64) bool {
    return IsClose(a, 0.0);
}

func FloatToStr(value float64) string {
    return strconv.FormatFloat(value, 'f', -1, 64);
}

func MustStrToFloat(value string) float64 {
    result, err := StrToFloat(value);
    if (err != nil) {
        log.Fatal().Err(err).Str("value", value).Msg("Failed to convert string to float.");
    }

    return result;
}

func StrToFloat(value string) (float64, error) {
    return strconv.ParseFloat(value, 64);
}

func MinMax(values []float64) (float64, float64) {
    var min float64;
    var max float64;

    for i, value := range values {
        if ((i == 0) || (value < min)) {
            min = value;
        }

        if ((i == 0) || (value > max)) {
            max = value;
        }
    }

    return min, max;
}

// Will sort the input slice.
func Median(values []float64) float64 {
    slices.Sort(values);

    length := len(values);
    if (length == 0) {
        return 0.0;
    } else if (length % 2 == 0) {
        return (values[length / 2] + values[(length / 2) - 1]) / 2.0
    } else {
        return values[length / 2];
    }
}
