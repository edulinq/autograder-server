package main

import (
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/courses/users"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	CourseID string `help:"ID of the course." arg:""`
	Table    bool   `help:"Output data as a TSV." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("List the users in the course."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	request := users.ListRequest{
		APIRequestCourseUserContext: core.APIRequestCourseUserContext{
			CourseID: args.CourseID,
		},
	}

	var printFunc cmd.CustomResponseFormatter = nil
	if args.Table {
		printFunc = listCourseUsersTable
	}

	cmd.MustHandleCMDRequestAndExitFull(`courses/users/list`, request, users.ListResponse{}, args.CommonOptions, printFunc)
}

func listCourseUsersTable(response core.APIResponse) (string, bool) {
	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	var courseUsersTable strings.Builder

	headers := []string{"email", "name", "role", "lms-id"}
	courseUsersTable.WriteString(strings.Join(headers, "\t") + "\n")

	for i, user := range responseContent.Users {
		if i > 0 {
			courseUsersTable.WriteString("\n")
		}

		courseUsersTable.WriteString(user.Email)
		courseUsersTable.WriteString("\t")
		courseUsersTable.WriteString(user.Name)
		courseUsersTable.WriteString("\t")
		courseUsersTable.WriteString(user.Role.String())
		courseUsersTable.WriteString("\t")
		courseUsersTable.WriteString(user.LMSID)
	}

	return courseUsersTable.String(), true
}
