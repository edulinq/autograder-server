package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Course         string `help:"ID of the course." arg:""`
	Assignment     string `help:"ID of the assignment." arg:""`
	Submission     string `help:"Path to submission directory." required:"" type:"existingdir"`
	OutPath        string `help:"Option path to output a JSON grading result." type:"path"`
	User           string `help:"User email for the submission." default:"testuser"`
	Message        string `help:"Submission message." default:""`
	AllowLate      bool   `help:"Allow this submission to be graded, even if it is late." default:"false"`
	CheckRejection bool   `help:"Check if this submission should be rejected (bypassed by default)." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Perform a grading."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	assignment := db.MustGetAssignment(args.Course, args.Assignment)

	gradeOptions := grader.GetDefaultGradeOptions()
	gradeOptions.AllowLate = args.AllowLate

	result, reject, softError, err := grader.Grade(assignment, args.Submission, args.User, args.Message, args.CheckRejection, gradeOptions)
	if err != nil {
		if (result != nil) && result.HasTextOutput() {
			fmt.Println("Grading failed, but output was recovered:")
			fmt.Println(result.GetCombinedOutput())
		}
		log.Fatal("Failed to run grader.", assignment, err)
	}

	if reject != nil {
		log.Fatal("Submission was rejected.", assignment, log.NewAttr("reject-reason", reject.String()))
	}

	if softError != "" {
		log.Fatal("Submission got a soft error.", assignment, log.NewAttr("soft-error", softError))
	}

	if args.OutPath != "" {
		err = util.ToJSONFileIndent(result.Info, args.OutPath)
		if err != nil {
			log.Fatal("Failed to output JSON result.", assignment, log.NewAttr("outpath", args.OutPath), err)
		}
	}

	fmt.Println(result.Info.Report())
}
