package cmd

import (
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/util"
)

func ListServerUsersTable(response core.APIResponse) string {
	var responseContent users.ListResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	var serverUsersTable strings.Builder

	headers := []string{"email", "name", "server-role", "courses"}
	serverUsersTable.WriteString(strings.Join(headers, "\t") + "\n")

	for i, user := range responseContent.Users {
		if i > 0 {
			serverUsersTable.WriteString("\n")
		}

		serverUsersTable.WriteString(user.Email)
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(user.Name)
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(user.Role.String())
		serverUsersTable.WriteString("\t")
		serverUsersTable.WriteString(util.MustToJSONIndent(user.Courses))
	}

	return serverUsersTable.String()
}
