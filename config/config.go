package config;

// For the defaulted getters, the defualt will be returned on ANY error
// (even if the key exists, but is of the wrong type).

import (
    "encoding/json"
    "fmt"
    "io"
    "math"
    "os"
    "strconv"
    "strings"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

const ENV_PREFIX = "AUTOGRADER__";
const ENV_DOT_REPLACEMENT = "__";

var configValues map[string]any = make(map[string]any);

func ToJSON() (string, error) {
    return util.ToJSONIndent(configValues);
}

func init() {
    LoadLoacalConfig();
    LoadEnv();
    InitLogging();
}

func InitLogging() {
    if (LOG_PRETTY.GetBool()) {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr});
    } else {
        log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger();
    }

    var rawLogLevel = LOG_LEVEL.GetString();
    level, err := zerolog.ParseLevel(rawLogLevel);
    if (err != nil) {
        log.Fatal().Err(err).Str("level", rawLogLevel).Msg("Failed to parse the logging level.");
    }

    if (DEBUG.GetBool() && (level > zerolog.DebugLevel)) {
        level = zerolog.DebugLevel;
    }

    zerolog.SetGlobalLevel(level);
}

func LoadLoacalConfig() error {
    path := LOCAL_CONFIG_PATH.GetString();
    if (util.PathExists(path)) {
        err := LoadFile(path);
        if (err != nil) {
            return fmt.Errorf("Could not load local config '%s': '%w'.", path, err);
        }
    }

    path = SECRETS_CONFIG_PATH.GetString();
    if (util.PathExists(path)) {
        err := LoadFile(path);
        if (err != nil) {
            return fmt.Errorf("Could not load secrets config '%s': '%w'.", path, err);
        }
    }

    return nil;
}

// See LoadReader().
func LoadFile(path string) error {
    file, err := os.Open(path);
    if (err != nil) {
        return fmt.Errorf("Could not open config file (%s): %w.", path, err);
    }
    defer file.Close();

    err = LoadReader(file);
    if (err != nil) {
        return fmt.Errorf("Unable to decode config file (%s): %w.", path, err);
    }

    return nil;
}

// See LoadReader().
func LoadString(text string) error {
    err := LoadReader(strings.NewReader(text));
    if (err != nil) {
        return fmt.Errorf("Unable to decode config from string: %w.", err);
    }

    return nil;
}

// Load data into the configuration.
// This will not clear out an existing configuration (so can load multiple files).
// If there are any key conflicts, the data loaded last will win.
// If you want to clear the config, use Reset().
func LoadReader(reader io.Reader) error {
    decoder := json.NewDecoder(reader);

    var fileOptions map[string]any;

    err := decoder.Decode(&fileOptions);
    if (err != nil) {
        return err;
    }

    for key, value := range fileOptions {
        // encoding/json uses float64 as its default numeric type.
        // Check if it is actually an integer.
        floatValue, ok := value.(float64);
        if (ok) {
            if (math.Trunc(floatValue) == floatValue) {
                value = int(floatValue);
            }
        }

        configValues[key] = value;
    }

    return nil;
}

// Load any config options from environmental variables.
// Config keys must start with ENV_PREFIX.
// Keys will them be trainsformed by
// removing the leading ENV_PREFIX, replacing ENV_DOT_REPLACEMENT with '.', and lowercasing.
func LoadEnv() {
    for _, envValue := range(os.Environ()) {
        if (!strings.HasPrefix(envValue, ENV_PREFIX)) {
            continue;
        }

        parts := strings.SplitN(envValue, "=", 2);

        key := parts[0];
        value := parts[1];

        key = strings.TrimPrefix(key, ENV_PREFIX);
        key = strings.ReplaceAll(key, ENV_DOT_REPLACEMENT, ".");
        key = strings.ToLower(key);

        Set(key, value);
    }
}

func Reset() {
    configValues = make(map[string]any);
}

func Has(key string) bool {
    _, present := configValues[key];
    return present;
}

func Set(key string, value any) {
    configValues[key] = value;
}

func Get(key string) any {
    value, present := configValues[key];
    if (!present) {
        log.Fatal().Str("key", key).Msg("Config key does not exist.");
    }

    return value;
}

func GetDefault(key string, defaultValue any) any {
    value, exists := configValues[key];
    if (exists) {
        return value;
    }

    return defaultValue;
}

func GetString(key string) string {
    return asString(Get(key));
}

func GetStringDefault(key string, defaultValue string) string {
    return asString(GetDefault(key, defaultValue));
}

func GetInt(key string) int {
    intValue, err := asInt(Get(key));
    if (err != nil) {
        log.Fatal().Err(err).Str("key", key).Msg("Could not get int option.");
    }

    return intValue;
}

func GetIntDefault(key string, defaultValue int) int {
    intValue, err := asInt(GetDefault(key, defaultValue));
    if (err != nil) {
        log.Warn().Err(err).Str("key", key).Int("default", defaultValue).Msg("Could not get int option, returning default.");
        return defaultValue;
    }

    return intValue;
}

func GetFloat(key string) float64 {
    floatValue, err := asFloat(Get(key));
    if (err != nil) {
        log.Fatal().Err(err).Str("key", key).Msg("Could not get float option.");
    }

    return floatValue;
}

func GetFloatDefault(key string, defaultValue float64) float64 {
    floatValue, err := asFloat(GetDefault(key, defaultValue));
    if (err != nil) {
        log.Warn().Err(err).Str("key", key).Float64("default", defaultValue).Msg("Could not get float option, returning default.");
        return defaultValue;
    }

    return floatValue;
}

func GetBool(key string) bool {
    boolValue, err := asBool(Get(key));
    if (err != nil) {
        log.Fatal().Err(err).Str("key", key).Msg("Could not get bool option.");
    }

    return boolValue;
}

func GetBoolDefault(key string, defaultValue bool) bool {
    boolValue, err := asBool(GetDefault(key, defaultValue));
    if (err != nil) {
        log.Warn().Err(err).Str("key", key).Bool("default", defaultValue).Msg("Could not get bool option, returning default.");
        return defaultValue;
    }

    return boolValue;
}

func asString(value any) string {
    stringValue, ok := value.(string);
    if (!ok) {
        return fmt.Sprintf("%v", value);
    }

    return stringValue;
}

func asInt(value any) (int, error) {
    strValue, ok := value.(string);
    if (ok) {
        intValue, err := strconv.Atoi(strValue);
        if (err != nil) {
            return 0, fmt.Errorf("Config value is a string ('%s'), but could not be converted to an int: %w.", strValue, err);
        }

        return intValue, nil;
    }

    intValue, ok := value.(int);
    if (!ok) {
        return 0, fmt.Errorf("Config value ('%v') is not an int.", value);
    }

    return intValue, nil;
}

func asFloat(value any) (float64, error) {
    strValue, ok := value.(string);
    if (ok) {
        floatValue, err := strconv.ParseFloat(strValue, 64);
        if (err != nil) {
            return 0.0, fmt.Errorf("Config value is a string ('%s'), but could not be converted to a float: %w.", strValue, err);
        }

        return floatValue, nil;
    }

    floatValue, ok := value.(float64);
    if (!ok) {
        return 0.0, fmt.Errorf("Config value ('%v') is not a float.", value);
    }

    return floatValue, nil;
}

func asBool(value any) (bool, error) {
    strValue, ok := value.(string);
    if (ok) {
        boolValue, err := strconv.ParseBool(strValue);
        if (err != nil) {
            return false, fmt.Errorf("Config value is a string ('%s'), but could not be converted to a bool: %w.", strValue, err);
        }

        return boolValue, nil;
    }

    boolValue, ok := value.(bool);
    if (!ok) {
        return false, fmt.Errorf("Config value ('%v') is not a bool.", value);
    }

    return boolValue, nil;
}
