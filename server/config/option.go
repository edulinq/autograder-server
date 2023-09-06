package config

import (
    "github.com/rs/zerolog/log"
)

var seenOptions = make(map[string]bool);

// Options are a way to access the general configuration in a structured way.
// Options do not themselves hold a value, but just information on how to access the config.
// Options will generally panic if you try to get the incorrect type from them.
type Option struct {
    Key string
    DefaultValue any
    Description string
}

// Create a new option, will panic on failure.
func newOption(key string, defaultValue any, description string) *Option {
    _, ok := seenOptions[key];
    if (ok) {
        log.Fatal().Str("key", key).Msg("Duplicate option key.");
    }

    option := Option{
        Key: key,
        DefaultValue: defaultValue,
        Description: description,
    }

    seenOptions[key] = true;
    return &option;
}

func (this *Option) Get() any {
    return GetDefault(this.Key, this.DefaultValue);
}

func (this *Option) GetString() string {
    return GetStringDefault(this.Key, this.DefaultValue.(string));
}

func (this *Option) GetInt() int {
    return GetIntDefault(this.Key, this.DefaultValue.(int));
}

func (this *Option) GetBool() bool {
    return GetBoolDefault(this.Key, this.DefaultValue.(bool));
}
