package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/cmd"
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
	Short  bool     `help:"Use short form output."`
}

func main() {
	kong.Parse(&args,
		kong.Description("List users on the server (default) or from the specified course."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	if args.Course == "" {
		listServerUsers(args.Emails, args.Table)
	} else {
		db.MustOpen()
		defer db.MustClose()
		listCourseUsers(args.Emails, args.Course, args.Table)
	}
}

func listServerUsers(emails []string, table bool) {
	request := users.ListRequest{}

	response, err := cmd.SendCMDRequest(`users/list`, request)
	if err != nil {
		log.Fatal("Failed to send the list server users CMD request.", err)
	}

	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if len(emails) > 0 {
		response.Content = filterUsersByEmail(responseContent, emails)
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
	}

	if table {
		fmt.Println(strings.Join(model.COURSE_USER_ROW_COLUMNS, "\t"))
		for _, user := range responseContent.Users {
			fmt.Println(strings.Join(user.MustToRow(), "\t"))
		}
	} else {
		cmd.PrintCMDResponse(response, users.ListResponse{}, args.Short)
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

func filterUsersByEmail(userList users.ListResponse, emails []string) users.ListResponse {
	emailSet := make(map[string]struct{}, len(emails))
	for _, email := range emails {
		emailSet[email] = struct{}{}
	}

	var filteredUsers []*core.ServerUserInfo
	for _, user := range userList.Users {
		if _, exists := emailSet[user.Email]; exists {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return users.ListResponse{Users: filteredUsers}
}
