package submissions

// All the API endpoints handled by this package.

import (
	"github.com/edulinq/autograder/internal/api/core"
)

var routes []core.Route = []core.Route{
	core.NewAPIRoute(`courses/assignments/submissions/fetch/course/attempts`, HandleFetchCourseAttempts,
		"Get all recent submissions and grading information for this assignment."),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/course/scores`, HandleFetchCourseScores,
		"Get a summary of the most recent scores for this assignment."),

	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/attempt`, HandleFetchUserAttempt,
		"Get a submission along with all grading information."),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/attempts`, HandleFetchUserAttempts,
		"Get all submission attempts made by a user along with all grading information."),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/history`, HandleFetchUserHistory,
		"Get a summary of the submissions for this assignment."),
	core.NewAPIRoute(`courses/assignments/submissions/fetch/user/peek`, HandleFetchUserPeek,
		"Get a copy of the grading report for the specified submission. Does not submit a new submission."),

	core.NewAPIRoute(`courses/assignments/submissions/remove`, HandleRemove,
		"Remove a specified submission. Defaults to the most recent submission."),
	core.NewAPIRoute(`courses/assignments/submissions/submit`, HandleSubmit,
		"Submit an assignment submission to the autograder."),
}

func GetRoutes() *[]core.Route {
	return &routes
}
