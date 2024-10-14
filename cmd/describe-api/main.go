package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
}

func main() {
	kong.Parse(&args,
		kong.Description("Describe all API endpoints."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	os.Exit(run())
}

func run() int {
	description, err := api.Describe()
	if err != nil {
		log.Fatal("Unable to describe API endpoints.", err)
	}

	fmt.Print(description)

	return 0
}
