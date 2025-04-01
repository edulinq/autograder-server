package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	Course     string `help:"ID of the course." arg:""`
	Assignment string `help:"ID of the assignment." arg:""`
	Grades     string `help:"Path to TSV file containing 'email<TAB>score'." arg:"" type:"existingfile"`
	Force      bool   `help:"Ignore when there are bad users and upload all the grades for good users." short:"f" default:"false"`
	DryRun     bool   `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Upload grades for an assignment to the coure's LMS from a TSV file."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	assignment := db.MustGetAssignment(args.Course, args.Assignment)
	course := assignment.GetCourse()

	if assignment.GetLMSID() == "" {
		log.Fatal("Assignment has no LMS ID.", assignment)
	}

	users, err := db.GetCourseUsers(course)
	if err != nil {
		log.Fatal("Failed to fetch autograder users.", err, assignment)
	}

	grades, err := loadGrades(args.Grades, users, args.Force)
	if err != nil {
		log.Fatal("Could not fetch grades.", err, assignment)
	}

	if len(grades) == 0 {
		fmt.Println("Found no grades to upload.")
	}

	if args.DryRun {
		fmt.Println("Dry Run: Skipping upload.")
	} else {
		err = lms.UpdateAssignmentScores(course, assignment.GetLMSID(), grades)
		if err != nil {
			log.Fatal("Could not upload grades.", err, assignment)
		}
	}

	fmt.Printf("Uploaded %d grades.\n", len(grades))
}

func loadGrades(path string, users map[string]*model.CourseUser, force bool) ([]*lmstypes.SubmissionScore, error) {
	grades := make([]*lmstypes.SubmissionScore, 0)

	rows, err := util.ReadSeparatedFile(path, "\t", 0)
	if err != nil {
		return nil, err
	}

	for i, row := range rows {
		if len(row) < 2 {
			return nil, fmt.Errorf("Row (%d) does not have enough values. Expecting 2, found %d.", i, len(row))
		}

		user := users[row[0]]
		if user == nil {
			message := fmt.Sprintf("Row (%d) has an unrecognized user: '%s'.", i, row[0])

			if force {
				fmt.Println(message)
				continue
			} else {
				return nil, fmt.Errorf("%s", message)
			}
		}

		lmsID := user.GetLMSID()
		if lmsID == "" {
			message := fmt.Sprintf("User '%s' (from row (%d)) has no LMS ID.", row[0], i)

			if force {
				fmt.Println(message)
				continue
			} else {
				return nil, fmt.Errorf("%s", message)
			}
		}

		grades = append(grades, &lmstypes.SubmissionScore{
			UserID: lmsID,
			Score:  util.MustStrToFloat(row[1]),
		})
	}

	return grades, nil
}
