package cmd

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func ListServerUsersTable(users []*core.ServerUserInfo) {
	headers := []string{"email", "name", "server-role", "courses"}
	fmt.Println(strings.Join(headers, "\t"))

	for _, user := range users {
		coursesJSON := util.MustToJSONIndent(user.Courses)

		row := []string{user.Email, user.Name, user.Role.String(), string(coursesJSON)}

		fmt.Println(strings.Join(row, "\t"))
	}
}
