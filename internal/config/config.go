package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const ENV_PREFIX = "AUTOGRADER__"
const ENV_DOT_REPLACEMENT = "__"

const CONFIG_FILENAME = "config.json"
const SECRETS_FILENAME = "secrets.json"

var configValues map[string]any = make(map[string]any)

// Init any system components using the current config.
// This should be called after the config has been set.
// For example, this will init the logging component with the current logging levels.
func InitSystemWithConfig() {
	initLoggingFromConfig()

	util.SetStoreHTTPDir(STORE_HTTP.Get())
}

func ToJSON() (string, error) {
	return util.ToJSONIndent(configValues)
}

// A mode intended for running unit tests.
func MustEnableUnitTestingMode() {
	err := EnableUnitTestingMode()
	if err != nil {
		log.Fatal("Failed to enable unit testing mode.", err)
	}
}

func EnableUnitTestingMode() error {
	return EnableUnitTestingModeFull(true)
}

// Sets system configurations for unit testing.
// Changes the base directory to a temp directory and copies over test data.
// If loadEnv is true, loads environmental variables.
func EnableUnitTestingModeFull(loadEnv bool) error {
	defer InitSystemWithConfig()

	UNIT_TESTING_MODE.Set(true)
	NO_TASKS.Set(true)
	LOAD_TEST_DATA.Set(true)

	tempWorkDir, err := util.MkDirTemp("autograder-unit-testing-")
	if err != nil {
		return fmt.Errorf("Failed to make temp unit testing work dir: '%w'.", err)
	}

	// Change the base dir to a temp dir.
	BASE_DIR.Set(tempWorkDir)

	// Upgrade the backend logging level to debug if no other level is set through config.
	if !LOG_BACKEND_LEVEL.IsSet() {
		LOG_BACKEND_LEVEL.Set(log.LevelDebug.String())
	}

	// Copy over test courses.
	testsDir := filepath.Join(util.RootDirForTesting(), TESTDATA_DIRNAME)

	err = util.CopyDir(testsDir, GetTestdataDir(), false)
	if err != nil {
		return fmt.Errorf("Failed to copy test data into working dir: '%w'.", err)
	}

	if loadEnv {
		LoadEnv()
	}

	return nil
}

func LoadConfigFromDir(dir string) error {
	path := filepath.Join(dir, CONFIG_FILENAME)
	if util.PathExists(path) {
		err := LoadFile(path)
		if err != nil {
			return fmt.Errorf("Could not load config '%s': '%w'.", path, err)
		}
	}

	path = filepath.Join(dir, SECRETS_FILENAME)
	if util.PathExists(path) {
		err := LoadFile(path)
		if err != nil {
			return fmt.Errorf("Could not load secrets '%s': '%w'.", path, err)
		}
	}

	return nil
}

// Load data into the configuration.
// This will not clear out an existing configuration (so can load multiple files).
// If there are any key conflicts, the data loaded last will win.
// If you want to clear the config, use Reset().
func LoadFile(path string) error {
	var fileOptions map[string]any
	err := util.JSONFromFile(path, &fileOptions)
	if err != nil {
		return fmt.Errorf("Could not read config from file (%s): %w.", path, err)
	}

	for key, value := range fileOptions {
		if key == BASE_DIR.Key {
			return fmt.Errorf("Cannot set key '%s' in config files, use the command-line.", key)
		}

		// encoding/json uses float64 as its default numeric type.
		// Check if it is actually an integer.
		floatValue, ok := value.(float64)
		if ok {
			if math.Trunc(floatValue) == floatValue {
				value = int(floatValue)
			}
		}

		configValues[key] = value
	}

	return nil
}

// Load any config options from environmental variables.
// Config keys must start with ENV_PREFIX.
// Keys will them be trainsformed by
// removing the leading ENV_PREFIX, replacing ENV_DOT_REPLACEMENT with '.', and lowercasing.
func LoadEnv() {
	for _, envValue := range os.Environ() {
		if !strings.HasPrefix(envValue, ENV_PREFIX) {
			continue
		}

		parts := strings.SplitN(envValue, "=", 2)

		key := parts[0]
		value := parts[1]

		key = strings.TrimPrefix(key, ENV_PREFIX)
		key = strings.ReplaceAll(key, ENV_DOT_REPLACEMENT, ".")
		key = strings.ToLower(key)

		Set(key, value)
	}
}

func Reset() {
	configValues = make(map[string]any)
}

func Has(key string) bool {
	_, present := configValues[key]
	return present
}

func Set(key string, value any) {
	configValues[key] = value
}

func GetDefault(key string, defaultValue any) any {
	value, exists := configValues[key]
	if exists {
		return value
	}

	return defaultValue
}

func GetStringDefault(key string, defaultValue string) string {
	return asString(GetDefault(key, defaultValue))
}

func GetIntDefault(key string, defaultValue int) int {
	intValue, err := asInt(GetDefault(key, defaultValue))
	if err != nil {
		log.Warn("Could not get int option, returning default.", err, log.NewAttr("key", key), log.NewAttr("default", defaultValue))
		return defaultValue
	}

	return intValue
}

func GetFloatDefault(key string, defaultValue float64) float64 {
	floatValue, err := asFloat(GetDefault(key, defaultValue))
	if err != nil {
		log.Warn("Could not get float option, returning default.", err, log.NewAttr("key", key), log.NewAttr("default", defaultValue))
		return defaultValue
	}

	return floatValue
}

func GetBoolDefault(key string, defaultValue bool) bool {
	boolValue, err := asBool(GetDefault(key, defaultValue))
	if err != nil {
		log.Warn("Could not get bool option, returning default.", err, log.NewAttr("key", key), log.NewAttr("default", defaultValue))
		return defaultValue
	}

	return boolValue
}

func asString(value any) string {
	stringValue, ok := value.(string)
	if !ok {
		return fmt.Sprintf("%v", value)
	}

	return stringValue
}

func asInt(value any) (int, error) {
	strValue, ok := value.(string)
	if ok {
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			return 0, fmt.Errorf("Config value is a string ('%s'), but could not be converted to an int: %w.", strValue, err)
		}

		return intValue, nil
	}

	intValue, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("Config value ('%v') is not an int.", value)
	}

	return intValue, nil
}

func asFloat(value any) (float64, error) {
	strValue, ok := value.(string)
	if ok {
		floatValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return 0.0, fmt.Errorf("Config value is a string ('%s'), but could not be converted to a float: %w.", strValue, err)
		}

		return floatValue, nil
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0.0, fmt.Errorf("Config value ('%v') is not a float.", value)
	}

	return floatValue, nil
}

func asBool(value any) (bool, error) {
	strValue, ok := value.(string)
	if ok {
		boolValue, err := strconv.ParseBool(strValue)
		if err != nil {
			return false, fmt.Errorf("Config value is a string ('%s'), but could not be converted to a bool: %w.", strValue, err)
		}

		return boolValue, nil
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("Config value ('%v') is not a bool.", value)
	}

	return boolValue, nil
}
