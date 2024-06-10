package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs

	Email string `help:"Email for the user." arg:"" required:""`
	Token string `help:"Token for the user." arg:"" required:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Authenticate as a user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	os.Exit(run())
}

func run() int {
	user, err := db.GetServerUser(args.Email, true)
	if err != nil {
		log.Fatal("Failed to get user.", err)
	}

	if user == nil {
		fmt.Printf("User '%s' does not exist, cannot auth.\n", args.Email)
		return 2
	}

	passHash := util.Sha256HexFromString(args.Token)

	auth, err := user.Auth(passHash)
	if err != nil {
		log.Fatal("Failed to auth user.", err)
	}

	if auth {
		fmt.Println("Authentication Successful")
		return 0
	} else {
		fmt.Println("Authentication Failed, Bad Password")
		return 1
	}
}
