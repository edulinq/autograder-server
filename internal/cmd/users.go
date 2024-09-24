package cmd

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
)

const INDENT = "    "

var rows [][]string

func ListUsers(users []*core.ServerUserInfo, courseUsers bool, table bool) {
	if table {
		listServerUsersTable(users)
	} else {
		listServerUsers(users)
	}
}

func listServerUsers(users []*core.ServerUserInfo) {
	for i, user := range users {
		if i != 0 {
			fmt.Println()
		}

		fmt.Println("Email: " + user.Email)
		fmt.Println("Name: " + user.Name)
		fmt.Println("Role: ", user.Role.String())
		fmt.Println("Courses:")

		courseIndex := 0
		for _, course := range user.Courses {
			if courseIndex != 0 {
				fmt.Println()
			}

			fmt.Println(INDENT + "course: " + course.CourseID)
			fmt.Println(INDENT + "Name: " + course.CourseName)
			fmt.Println(INDENT + "Role: " + course.Role.String())

			courseIndex++
		}
	}
}

func listServerUsersTable(users []*core.ServerUserInfo) {
	for _, user := range users {
		row := []string{user.Email, user.Name, user.Role.String()}

		if len(user.Courses) == 0 {
			emptyCourseInfo := make([]string, len(core.SERVER_USER_INFO_ROW_COLUMNS))
			row = append(row, emptyCourseInfo...)
			rows = append(rows, row)
		} else {
			for _, course := range user.Courses {
				courseRow := append(row, course.CourseID, course.CourseName, course.Role.String())
				rows = append(rows, courseRow)
			}
		}
	}

	printTSV(rows, core.SERVER_USER_INFO_ROW_COLUMNS)
}

func printTSV(rows [][]string, headerKeys []string) {
	var lines []string

	fmt.Println(strings.Join(headerKeys, "\t"))

	for _, row := range rows {
		lines = append(lines, strings.Join(row, "\t"))
	}

	fmt.Println(strings.Join(lines, "\n"))
}
