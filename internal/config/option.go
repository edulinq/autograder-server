package config

// Options are a way to access the general configuration in a structured way.
// Options do not themselves hold a value, but just information on how to access the config.
// Users should heavily prefer getting config values via Options rather than directly in the config.

import (
    "slices"
    "strings"

    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/util"
)

var seenOptions = make(map[string]*baseOption);

type baseOption struct {
    Key string
    DefaultValue any
    Description string
}

type StringOption struct {*baseOption}
type IntOption struct {*baseOption}
type FloatOption struct {*baseOption}
type BoolOption struct {*baseOption}

// Create a new option, will panic on failure.
func mustNewOption(key string, defaultValue any, description string) *baseOption {
    _, ok := seenOptions[key];
    if (ok) {
        log.Fatal("Duplicate option key.", log.NewAttr("key", key));
    }

    option := baseOption{
        Key: key,
        DefaultValue: defaultValue,
        Description: description,
    }

    seenOptions[key] = &option;
    return &option;
}

func MustNewStringOption(key string, defaultValue string, description string) *StringOption {
    return &StringOption{mustNewOption(key, defaultValue, description)};
}

func MustNewIntOption(key string, defaultValue int, description string) *IntOption {
    return &IntOption{mustNewOption(key, defaultValue, description)};
}

func MustNewFloatOption(key string, defaultValue float64, description string) *FloatOption {
    return &FloatOption{mustNewOption(key, defaultValue, description)};
}

func MustNewBoolOption(key string, defaultValue bool, description string) *BoolOption {
    return &BoolOption{mustNewOption(key, defaultValue, description)};
}

func (this *StringOption) Get() string {
    return GetStringDefault(this.Key, this.DefaultValue.(string));
}

func (this *IntOption) Get() int {
    return GetIntDefault(this.Key, this.DefaultValue.(int));
}

func (this *FloatOption) Get() float64 {
    return GetFloatDefault(this.Key, this.DefaultValue.(float64));
}

func (this *BoolOption) Get() bool {
    return GetBoolDefault(this.Key, this.DefaultValue.(bool));
}

func (this *baseOption) Set(value any) {
    Set(this.Key, value);
}

func OptionsToJSON() (string, error) {
    options := make([]*baseOption, 0, len(seenOptions));

    for _, option := range seenOptions {
        options = append(options, option);
    }

    slices.SortFunc(options, func(a *baseOption, b *baseOption) int {
        return strings.Compare(a.Key, b.Key);
    });

    return util.ToJSONIndent(options);
}
