package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs

	Emails []string `help:"Optional list of users to limit the results to (leave empty for all users)." arg:"" optional:""`
	Course string   `help:"Only include user enrolled in this course."`
	Table  bool     `help:"Output data to stdout as a TSV." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("List users on the server (default) or from the specified course."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	if args.Course == "" {
		listServerUsers(args.Emails, args.Table)
	} else {
		listCourseUsers(args.Emails, args.Course, args.Table)
	}
}

func listServerUsers(emails []string, table bool) {
	users, err := db.GetServerUsers()
	if err != nil {
		log.Fatal("Failed to get server users.", err)
	}

	if len(emails) > 0 {
		newUsers := make(map[string]*model.ServerUser, len(emails))
		for _, email := range emails {
			newUsers[email] = users[email]
		}

		users = newUsers
	}

	if table {
		fmt.Println(strings.Join(model.SERVER_USER_ROW_COLUMNS, "\t"))
		for _, user := range users {
			fmt.Println(strings.Join(user.MustToRow(), "\t"))
		}
	} else {
		fmt.Println(util.MustToJSONIndent(users))
	}
}

func listCourseUsers(emails []string, courseID string, table bool) {
	course := db.MustGetCourse(courseID)

	users, err := db.GetCourseUsers(course)
	if err != nil {
		log.Fatal("Failed to get course users.", err)
	}

	if len(emails) > 0 {
		newUsers := make(map[string]*model.CourseUser, len(emails))
		for _, email := range emails {
			newUsers[email] = users[email]
		}

		users = newUsers
	}

	if table {
		fmt.Println(strings.Join(model.COURSE_USER_ROW_COLUMNS, "\t"))
		for _, user := range users {
			fmt.Println(strings.Join(user.MustToRow(), "\t"))
		}
	} else {
		fmt.Println(util.MustToJSONIndent(users))
	}
}
