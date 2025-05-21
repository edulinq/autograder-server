package model

// Course user references can be represented as follows:
//
// - An email address
// - A literal "*" (which includes all users in the course)
// - A course role (which will include all course users with that role)
// - Any of the above options preceded by a dash ("-") (which indicates that the user or group will NOT be included in the final results)
type CourseUserReferenceInput string

type CourseUserReference struct {
	// The course that orients the reference.
	Course *Course

	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of course roles to include.
	CourseUserRoles map[string]any

	// The set of course roles to exclude.
	ExcludeCourseUserRoles map[string]any
}
