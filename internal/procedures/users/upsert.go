package users

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type UpsertUsersOptions struct {
	RawUsers []*model.RawUserData

	SkipInserts bool
	SkipUpdates bool

	// Send any relevant email (usually about creation or password changing).
	SendEmails bool

	// Do not actually commit any changes or send any emails regardless of |SendEmails|.
	DryRun bool

	// The authority of the user making this add request.
	// The server role should always be set, but the course role can be unknown.
	ContextServerRole model.ServerUserRole
	ContextCourseRole model.CourseUserRole
}

// Call upsertUser() for multiple users.
// Every passed in user will get an index-matched entry in the returned value.
func UpsertUsers(options UpsertUsersOptions) []*model.UserOpResult {
	results := make([]*model.UserOpResult, 0, len(options.RawUsers))

	for i, _ := range options.RawUsers {
		results = append(results, upsertUser(options, i))
	}

	return results
}

func UpsertUser(options UpsertUsersOptions) *model.UserOpResult {
	return upsertUser(options, 0)
}

// Upsert the user at the given index according with the given options.
// Note that the returned errors are split into two categories in the UserOpResult structure: validation and system.
// Validation errors are errors coming from the data/user (like an invalid email).
// System errors are errors that should not happen regardless of the data.
// Essentially, system errors are out fault and validation errors are the user's fault.
func upsertUser(options UpsertUsersOptions, index int) *model.UserOpResult {
	if (options.ContextServerRole == model.ServerRoleUnknown) && (options.ContextCourseRole == model.RoleUnknown) {
		return model.NewSystemErrorUserOpResult("", fmt.Errorf("No authority/roles were provided when adding a user."))
	}

	if (index < 0) || (index >= len(options.RawUsers)) {
		return model.NewSystemErrorUserOpResult("", fmt.Errorf("User index out of bounds. Got %d, must be in [0, %d).", index, len(options.RawUsers)))
	}

	rawData := options.RawUsers[index]

	// Construct and validate a user with the raw data.
	newUser, err := rawData.ToServerUser()
	if err != nil {
		return model.NewValidationErrorUserOpResult(rawData.Email, err)
	}

	// Check that the context user has permissions to create the new user.
	result := checkUpsertPermissions(newUser, rawData.Course, options.ContextServerRole, options.ContextCourseRole)
	if result != nil {
		return result
	}

	// Check if this is an update or insert.
	oldUser, err := db.GetServerUser(newUser.Email, true)
	if err != nil {
		return model.NewSystemErrorUserOpResult(newUser.Email, err)
	}

	if oldUser == nil {
		// This is an add/insert.
		if !options.SkipInserts {
			result = insertUser(newUser, options, rawData)
		}
	} else {
		// This is an update.
		if !options.SkipUpdates {
			result = updateUser(newUser, oldUser, options, rawData)
		}
	}

	// This user was skipped.
	if result == nil {
		return &model.UserOpResult{
			Email:   newUser.Email,
			Skipped: true,
		}
	}

	// Send out any necessary emails.
	sendUserOpEmail(result, options)

	return result
}

func insertUser(user *model.ServerUser, options UpsertUsersOptions, rawData *model.RawUserData) *model.UserOpResult {
	result := &model.UserOpResult{
		Email: user.Email,
		Added: true,
	}

	// New users need authnetication information.
	user.Salt = util.StringPointer(model.ShouldNewRandomSalt())

	var err error = nil

	if rawData.Pass != "" {
		_, err = user.SetPassword(rawData.Pass)
	} else {
		result.CleartextPassword, err = user.SetRandomPassword()
	}

	if err != nil {
		return model.NewSystemErrorUserOpResult(user.Email, fmt.Errorf("Failed to set new password: '%w'.", err))
	}

	for courseID, _ := range user.Roles {
		result.Enrolled = append(result.Enrolled, courseID)
	}

	if !options.DryRun {
		err := db.UpsertUser(user)
		if err != nil {
			return model.NewSystemErrorUserOpResult(user.Email, fmt.Errorf("Failed to insert user: '%w'.", err))
		}
	}

	return result
}

func updateUser(newUser *model.ServerUser, user *model.ServerUser, options UpsertUsersOptions, rawData *model.RawUserData) *model.UserOpResult {
	var enrolledCourses []string = nil
	for course, _ := range newUser.Roles {
		_, exists := user.Roles[course]
		if !exists {
			enrolledCourses = append(enrolledCourses, course)
		}
	}

	changed, err := user.Merge(newUser)
	if err != nil {
		return model.NewValidationErrorUserOpResult(newUser.Email, err)
	}

	if rawData.Pass != "" {
		passChanged, err := user.SetPassword(rawData.Pass)
		if err != nil {
			return model.NewSystemErrorUserOpResult(newUser.Email, err)
		}

		changed = (changed || passChanged)
	}

	if !options.DryRun {
		err := db.UpsertUser(user)
		if err != nil {
			return model.NewSystemErrorUserOpResult(user.Email, fmt.Errorf("Failed to update user: '%w'.", err))
		}
	}

	result := &model.UserOpResult{
		Email:    user.Email,
		Modified: changed,
		Enrolled: enrolledCourses,
	}

	return result
}

// Ensure the context user can perform this add operation.
func checkUpsertPermissions(user *model.ServerUser, courseID string, contextServerRole model.ServerUserRole, contextCourseRole model.CourseUserRole) *model.UserOpResult {
	// Users must be known to the server before adding another user.
	if contextServerRole == model.ServerRoleUnknown {
		return model.NewSystemErrorUserOpResult(user.Email,
			fmt.Errorf("Users must have a server role to upsert users."))
	}

	// Regardless of course role, users with higher server permissions cannot be added.
	if contextServerRole < user.Role {
		return model.NewValidationErrorUserOpResult(user.Email,
			fmt.Errorf("User has a server role of '%s', which is not high enough to create a user with server role '%s'.", contextServerRole.String(), user.Role.String()))
	}

	// If the user has high enough server credentials, then there is no need to check any course role.
	if contextServerRole >= model.ServerRoleAdmin {
		return nil
	}

	// The user has no course role and insufficient server role.
	if contextCourseRole == model.RoleUnknown {
		return model.NewValidationErrorUserOpResult(user.Email,
			fmt.Errorf("User has an insufficient server role of '%s' and no course role to create users.", contextServerRole.String()))
	}

	// At this point, the context user's server-level credentials are not high enough to create any user.
	// Course-level credentials will need to be checked.
	courseRole := user.Roles[courseID]

	// The user's course-level credentials need to be high enough to create any user.
	if contextCourseRole < model.RoleAdmin {
		return model.NewValidationErrorUserOpResult(user.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to create users.", contextCourseRole.String()))
	}

	// Check the relative course credentials.
	if contextCourseRole < courseRole {
		return model.NewValidationErrorUserOpResult(user.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to create a user with course role '%s'.", contextCourseRole.String(), courseRole.String()))
	}

	return nil
}

func sendUserOpEmail(result *model.UserOpResult, options UpsertUsersOptions) {
	if !options.SendEmails {
		return
	}

	message := result.GetEmail()
	if message == nil {
		return
	}

	emailSent := true

	if !options.DryRun {
		err := email.SendMessage(message)
		if err != nil {
			result.SystemErrors = append(result.SystemErrors, fmt.Sprintf("Failed to send email: '%v'.", err))
			emailSent = false
		}
	}

	result.Emailed = emailSent
}
