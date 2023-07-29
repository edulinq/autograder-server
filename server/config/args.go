package config;

// A Kong-style struct for adding on all the config-related options to a CLI.
type ConfigArgs struct {
    ConfigPath []string `help:"Path to config file to load." type:"existingfile"`
    Config map[string]string `help:"Config options."`
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

    InitLogging();

    return nil;
}
