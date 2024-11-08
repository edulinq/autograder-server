package main

import (
	"github.com/alecthomas/kong"
	"github.com/edulinq/autograder/internal/config"
)

var args struct {
	config.ConfigArgs

	Endpoint string `help:"Endpoint of the desired action." arg:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Perform an action with the desired endpoint."),
	)

	
}