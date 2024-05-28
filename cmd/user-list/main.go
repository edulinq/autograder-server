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

	Course string `help:"Only include user enrolled in this course."`
	Table  bool   `help:"Output data to stdout as a TSV." default:"false"`
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
		listServerUsers(args.Table)
	} else {
		listCourseUsers(args.Course, args.Table)
	}
}

func listServerUsers(table bool) {
	users, err := db.GetServerUsers()
	if err != nil {
		log.Fatal("Failed to get server users.", err)
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

func listCourseUsers(courseID string, table bool) {
	course := db.MustGetCourse(courseID)

	users, err := db.GetCourseUsers(course)
	if err != nil {
		log.Fatal("Failed to get course users.", err)
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
