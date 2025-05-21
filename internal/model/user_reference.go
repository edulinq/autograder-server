package model

const USER_REFERENCE_DELIM = "::"

// A flexible way to reference server users.
// Server user references can be represented as follows (in the order the are evaluated):
//
// - An email address
// - A server role (which will include all server users with that role)
// - A literal "*" (which includes all users on the server)
// - Any of the above options preceded by a dash ("-") (which indicates that the user or group will NOT be included in the final results)
// - A course user reference
//
// A course user reference can be represented as follows:
//
// - <course-id>::<course-role> (which will include all users in the target course with the target course role)
// - *::<course-role> (which will include all users with the target course role in any course)
// - <course-id>::* (which includes all users in the target course)
// - *::* (which includes all users enrolled in at least one course)
// - Any of the above can be preceded by a dash ("-") (which indicates that that group will NOT be included in the final results)
type ServerUserReferenceInput string

// Course user references can be represented as follows (in the order the are evaluated):
//
// - An email address
// - A course role (which will include all course users with that role)
// - A literal "*" (which includes all users in the course)
// - Any of the above options preceded by a dash ("-") (which indicates that the user or group will NOT be included in the final results)
type CourseUserReferenceInput string

type ServerUserReference struct {
	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of server roles to include.
	ServerUserRoles map[string]any

	// The set of server roles to exclude.
	ExcludeServerUserRoles map[string]any

	// The courses and list of roles to include.
	// Keyed on the course ID.
	CourseUserReferences map[string]*CourseUserReference
}

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

func (this *ServerUserReference) AddCourseUserReference(courseUserReference *CourseUserReference) {
	if this == nil {
		return
	}

	// Transfer the emails to the ServerUserReference to reduce memory usage.
	for email, _ := range courseUserReference.Emails {
		this.Emails[email] = nil
	}

	courseUserReference.Emails = make(map[string]any, 0)

	for email, _ := range courseUserReference.ExcludeEmails {
		this.ExcludeEmails[email] = nil
	}

	courseUserReference.ExcludeEmails = make(map[string]any, 0)

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

	for roleString, _ := range courseUserReference.CourseUserRoles {
		currentCourseUserReference.CourseUserRoles[roleString] = nil
	}

	for roleString, _ := range courseUserReference.ExcludeCourseUserRoles {
		currentCourseUserReference.ExcludeCourseUserRoles[roleString] = nil
	}

	return
}

func (this *CourseUserReference) ToServerUserReference() *ServerUserReference {
	if this == nil {
		return nil
	}

	return &ServerUserReference{
		// Transfer Emails and ExcludeEmails to the ServerUserReference to reduce memory usage.
		Emails:                 this.Emails,
		ExcludeEmails:          this.ExcludeEmails,
		ServerUserRoles:        make(map[string]any, 0),
		ExcludeServerUserRoles: make(map[string]any, 0),
		CourseUserReferences: map[string]*CourseUserReference{
			this.Course.GetID(): &CourseUserReference{
				Course:                 this.Course,
				Emails:                 make(map[string]any, 0),
				ExcludeEmails:          make(map[string]any, 0),
				CourseUserRoles:        this.CourseUserRoles,
				ExcludeCourseUserRoles: this.ExcludeCourseUserRoles,
			},
		},
	}
}
