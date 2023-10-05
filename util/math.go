package util

import (
    "strconv"

    "github.com/rs/zerolog/log"
)

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
