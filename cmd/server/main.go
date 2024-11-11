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

	ok, err := common.CheckServerCreator(common.CmdServer)
	if err != nil {
		log.Fatal("Failed to start the server.", err)
	}

	if !ok {
		log.Fatal("A CMD server is running.")
	}

	err = server.Start(common.PrimaryServer)
	if err != nil {
		log.Fatal("Failed to start the server.", err)
	}
}
