package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
}

func main() {
	kong.Parse(&args,
		kong.Description("Show all the current config values."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	description, err := core.DescribeRoutes(*api.GetRoutes())
	if err != nil {
		log.Fatal("Failed to describe API endpoints.", err)
	}

	fmt.Printf("%s\n", util.MustToJSONIndent(description))
}
