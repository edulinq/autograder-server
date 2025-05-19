package model

const USER_REFERENCE_DELIM = "::"

type ServerUserReference struct {
	// Signals to include all server users.
	AllUsers bool

	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of server roles to include.
	ServerUserRoles map[ServerUserRole]any

	// The set of server roles to exclude.
	ExcludeServerUserRoles map[ServerUserRole]any

	// The courses and list of roles to include.
	// Keyed on the course ID.
	CourseReferences map[string]CourseUserReference

	// The set of courses to exclude.
	ExcludeCourseReferences map[string]any
}

type CourseUserReference struct {
	// The course that orients the reference.
	Course *Course

	// Signals to include all course users.
	AllUsers bool

	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of course roles to include.
	CourseUserRoles map[CourseUserRole]any

	// The set of course roles to exclude.
	ExcludeCourseUserRoles map[CourseUserRole]any
}

// TODO: Name is a bit off since we are not adding the reference directly.
func (this *ServerUserReference) AddCourseUserReference(courses map[string]*Course, courseRoles []CourseUserRole, exclude bool) {
	// TODO
	if this == nil {
		return
	}

	return
}

func (this *CourseUserReference) ToServerUserReference() *ServerUserReference {
	if this == nil {
		return nil
	}

	return &ServerUserReference{
		AllUsers:               this.AllUsers,
		Emails:                 this.Emails,
		ExcludeEmails:          this.ExcludeEmails,
		ServerUserRoles:        make(map[ServerUserRole]any, 0),
		ExcludeServerUserRoles: make(map[ServerUserRole]any, 0),
		// Clear the emails and exclude emails to reduce memory usage.
		// These fields are transferred to the new ServerUserReference.
		CourseReferences: map[string]CourseUserReference{
			this.Course.GetID(): CourseUserReference{
				Course:                 this.Course,
				AllUsers:               this.AllUsers,
				Emails:                 map[string]any{},
				ExcludeEmails:          map[string]any{},
				CourseUserRoles:        this.CourseUserRoles,
				ExcludeCourseUserRoles: this.ExcludeCourseUserRoles,
			},
		},
		ExcludeCourseReferences: map[string]any{},
	}
}
