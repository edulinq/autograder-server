package model

const USER_REFERENCE_DELIM = "::"

// A flexible way to reference server users.
// Server user references can be represented as follows (in the order the are evaluated):
//
// - An email address
// - An email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
// - A server role (which will include all server users with that role)
// - A literal "*" (which includes all users on the server)
// TODO: Update final part of this comment (show examples of various forms)
// - A course user reference
type ServerUserReferenceInput string

// Course user references can be represented as follows (in the order they are evaluated):
//
// - An email address
// - An email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
// - A course role (which will include all course users with that role)
// - A literal "*" (which includes all users in the course)
// - A course user reference
type CourseUserReferenceInput string

type ServerUserReference struct {
	// Signals to include all server users.
	AllUsers bool

	// Signals to exclude all server users.
	ExcludeAllUsers bool

	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of server roles to include.
	ServerUserRoles map[string]ServerUserRole

	// The set of server roles to exclude.
	ExcludeServerUserRoles map[string]ServerUserRole

	// The courses and list of roles to include.
	// Keyed on the course ID.
	CourseUserReferences map[string]*CourseUserReference

	// The set of courses to exclude.
	ExcludeCourseUserReferences map[string]any
}

type CourseUserReference struct {
	// The course that orients the reference.
	Course *Course

	// Signals to include all course users.
	AllUsers bool

	// Signals to exclude all course users.
	ExcludeAllUsers bool

	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of course roles to include.
	CourseUserRoles map[string]CourseUserRole

	// The set of course roles to exclude.
	ExcludeCourseUserRoles map[string]CourseUserRole
}

func (this *ServerUserReference) AddCourseUserReference(courseUserReference *CourseUserReference) {
	if this == nil {
		return
	}

	if courseUserReference.Course == nil {
		return
	}

	currentCourseUserReference, ok := this.CourseUserReferences[courseUserReference.Course.GetID()]
	if !ok {
		this.CourseUserReferences[courseUserReference.Course.GetID()] = courseUserReference
		return
	}

	if currentCourseUserReference == nil {
		this.CourseUserReferences[courseUserReference.Course.GetID()] = courseUserReference
		return
	}

	if courseUserReference.AllUsers {
		currentCourseUserReference.AllUsers = true
	}

	for email, _ := range courseUserReference.Emails {
		currentCourseUserReference.Emails[email] = nil
	}

	for email, _ := range courseUserReference.ExcludeEmails {
		currentCourseUserReference.ExcludeEmails[email] = nil
	}

	for roleString, role := range courseUserReference.CourseUserRoles {
		currentCourseUserReference.CourseUserRoles[roleString] = role
	}

	for roleString, role := range courseUserReference.ExcludeCourseUserRoles {
		currentCourseUserReference.ExcludeCourseUserRoles[roleString] = role
	}

	return
}

func (this *CourseUserReference) ToServerUserReference() *ServerUserReference {
	if this == nil {
		return nil
	}

	return &ServerUserReference{
		AllUsers:               this.AllUsers,
		ExcludeAllUsers:        this.ExcludeAllUsers,
		Emails:                 this.Emails,
		ExcludeEmails:          this.ExcludeEmails,
		ServerUserRoles:        make(map[string]ServerUserRole, 0),
		ExcludeServerUserRoles: make(map[string]ServerUserRole, 0),
		// Clear the emails and exclude emails to reduce memory usage.
		// These fields are transferred to the new ServerUserReference.
		CourseUserReferences: map[string]*CourseUserReference{
			this.Course.GetID(): &CourseUserReference{
				Course:                 this.Course,
				AllUsers:               this.AllUsers,
				Emails:                 make(map[string]any, 0),
				ExcludeEmails:          make(map[string]any, 0),
				CourseUserRoles:        this.CourseUserRoles,
				ExcludeCourseUserRoles: this.ExcludeCourseUserRoles,
			},
		},
		ExcludeCourseUserReferences: make(map[string]any, 0),
	}
}
