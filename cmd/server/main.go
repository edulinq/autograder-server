package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/procedures/server"
)

var args struct {
	config.ConfigArgs
}

func main() {
	kong.Parse(&args)
	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	err = server.RunAndBlock(common.PRIMARY_SERVER)
	if err != nil {
		log.Fatal("Failed to start the server.", err)
	}
}
