package config;

import (
    "fmt"
)

// A Kong-style struct for adding on all the config-related options to a CLI.
type ConfigArgs struct {
    ConfigPath []string `help:"Path to config file to load." type:"existingfile"`
    Config map[string]string `help:"Config options." short:"c"`
    Debug bool `help:"Enable general debugging. Shortcut for '-c debug=true'." default:"false"`
    Testing bool `help:"Enable all options for general testing. Shortcut for '-c debug=true -c api.noauth=true -c grader.nostore=true -c tasks.disable=true'. Not compatible with --unit-testing." default:"false"`
    UnitTesting bool `help:"Enable all options for unit testing and load test data/courses. Not compatible with --testing." default:"false"`
}

func HandleConfigArgs(args ConfigArgs) error {
    for _, path := range args.ConfigPath {
        err := LoadFile(path);
        if (err != nil) {
            return err;
        }
    }

    for key, value := range args.Config {
        Set(key, value);
    }

    if (args.Debug) {
        DEBUG.Set(true);
    }

    if (args.Testing && args.UnitTesting) {
        return fmt.Errorf("--testing and --unit-testing cannot both be supplied.");
    }

    if (args.Testing) {
        EnableTestingMode();
    }

    if (args.UnitTesting) {
        err := EnableUnitTestingMode();
        if (err != nil) {
            return err;
        }
    }

    InitLogging();

    return nil;
}
