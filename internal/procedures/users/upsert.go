package users

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Data for upserting users.
// Upserting should only be done by server or course admins,
// it should not be used for self creation.
type UpsertUsersOptions struct {
	RawUsers []*model.RawUserData `json:"raw-users"`

	SkipInserts bool `json:"skip-inserts"`
	SkipUpdates bool `json:"skip-updates"`

	// Send any relevant email (usually about creation or password changing).
	SendEmails bool `json:"send-emails"`

	// Do not actually commit any changes or send any emails regardless of |SendEmails|.
	DryRun bool `json:"dry-run"`

	// The authority of the user making this add request.
	// The server role should always be set, but the course role can be unknown.
	// The email is necessary to lower the permissions for a self update.
	ContextEmail      string               `json:"-"`
	ContextServerRole model.ServerUserRole `json:"-"`
	ContextCourseRole model.CourseUserRole `json:"-"`
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
// Essentially, system errors are our fault and validation errors are the user's fault.
func upsertUser(options UpsertUsersOptions, index int) *model.UserOpResult {
	if (options.ContextServerRole == model.ServerRoleUnknown) && (options.ContextCourseRole == model.CourseRoleUnknown) {
		return model.NewUserOpResultSystemError("-1001", "", fmt.Errorf("No authority/roles were provided when adding a user."))
	}

	if (index < 0) || (index >= len(options.RawUsers)) {
		return model.NewUserOpResultSystemError("-1002", "", fmt.Errorf("User index out of bounds. Got %d, must be in [0, %d).", index, len(options.RawUsers)))
	}

	rawData := options.RawUsers[index]

	// Construct and validate a user with the raw data.
	newUser, err := rawData.ToServerUser()
	if err != nil {
		return model.NewUserOpResultValidationError("-1003", rawData.Email, err)
	}

	// Check that the context user has permissions to create the new user.
	result := checkBaseUpsertPermissions(newUser, options)
	if result != nil {
		return result
	}

	// Check if this is an update or insert.
	oldUser, err := db.GetServerUser(newUser.Email)
	if err != nil {
		return model.NewUserOpResultSystemError("-1004", newUser.Email, err)
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
			BaseUserOpResult: model.BaseUserOpResult{
				Email:   newUser.Email,
				Skipped: true,
			},
		}
	}

	// Send out any necessary emails.
	sendUserOpEmail(result, options)

	return result
}

func insertUser(user *model.ServerUser, options UpsertUsersOptions, rawData *model.RawUserData) *model.UserOpResult {
	result := checkInsertPermissions(user, rawData, options)
	if result != nil {
		return result
	}

	result = &model.UserOpResult{
		BaseUserOpResult: model.BaseUserOpResult{
			Email: user.Email,
			Added: true,
		},
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
		return model.NewUserOpResultSystemError("-1005", user.Email, fmt.Errorf("Failed to set new password: '%w'.", err))
	}

	for courseID, _ := range user.CourseInfo {
		result.Enrolled = append(result.Enrolled, courseID)
	}

	if !options.DryRun {
		err := db.UpsertUser(user)
		if err != nil {
			return model.NewUserOpResultSystemError("-1006", user.Email, fmt.Errorf("Failed to insert user: '%w'.", err))
		}
	}

	return result
}

func updateUser(newUser *model.ServerUser, user *model.ServerUser, options UpsertUsersOptions, rawData *model.RawUserData) *model.UserOpResult {
	result := checkUpdatePermissions(newUser, user, rawData, options)
	if result != nil {
		return result
	}

	var enrolledCourses []string = nil
	for course, _ := range newUser.CourseInfo {
		_, exists := user.CourseInfo[course]
		if !exists {
			enrolledCourses = append(enrolledCourses, course)
		}
	}

	changed, err := user.Merge(newUser)
	if err != nil {
		return model.NewUserOpResultValidationError("-1007", newUser.Email, err)
	}

	if rawData.Pass != "" {
		passChanged, err := user.SetPassword(rawData.Pass)
		if err != nil {
			return model.NewUserOpResultSystemError("-1008", newUser.Email, err)
		}

		changed = (changed || passChanged)
	}

	if !options.DryRun {
		err := db.UpsertUser(user)
		if err != nil {
			return model.NewUserOpResultSystemError("-1009", user.Email, fmt.Errorf("Failed to update user: '%w'.", err))
		}
	}

	result = &model.UserOpResult{
		BaseUserOpResult: model.BaseUserOpResult{
			Email:    user.Email,
			Modified: changed,
			Enrolled: enrolledCourses,
		},
	}

	return result
}

// Check the common upsert permissions (shared with insert and update).
// Will return nil if there is no permissions issue.
func checkBaseUpsertPermissions(user *model.ServerUser, options UpsertUsersOptions) *model.UserOpResult {
	// Users must be known to the server before insert.
	if options.ContextServerRole == model.ServerRoleUnknown {
		return model.NewUserOpResultSystemError("-1010", user.Email,
			fmt.Errorf("Users must have a server role to upsert users."))
	}

	// Regardless of course role, users with higher server permissions cannot be inserted.
	if options.ContextServerRole < user.Role {
		return model.NewUserOpResultValidationError("-1011", user.Email,
			fmt.Errorf("User has a server role of '%s', which is not high enough to upsert a user with server role of '%s'.", options.ContextServerRole.String(), user.Role.String()))
	}

	return nil
}

// Check permissions specific to an insert.
// This assumes that checkBaseUpsertPermissions() has already been called and passed.
func checkInsertPermissions(user *model.ServerUser, rawData *model.RawUserData, options UpsertUsersOptions) *model.UserOpResult {
	// After relative server roles are checked (in common permissions check), admins can do whatever.
	if options.ContextServerRole >= model.ServerRoleAdmin {
		return nil
	}

	// The user has no course role and insufficient server role.
	if options.ContextCourseRole == model.CourseRoleUnknown {
		return model.NewUserOpResultValidationError("-1012", user.Email,
			fmt.Errorf("User has an insufficient server role of '%s' and no course role to insert users.", options.ContextServerRole.String()))
	}

	// At this point, the context user's server-level credentials are not high enough to insert any user.
	// Course-level credentials will need to be checked.
	courseRole := user.GetCourseRole(rawData.Course)

	// Check the relative course credentials.
	if options.ContextCourseRole < courseRole {
		return model.NewUserOpResultValidationError("-1013", user.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to insert a user with course role of '%s'.", options.ContextCourseRole.String(), courseRole.String()))
	}

	// The user's course-level credentials need to be high enough to insert.
	if options.ContextCourseRole < model.CourseRoleAdmin {
		return model.NewUserOpResultValidationError("-1014", user.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to insert users.", options.ContextCourseRole.String()))
	}

	return nil
}

// Check permissions specific to an update.
// This assumes that checkBaseUpsertPermissions() has already been called and passed.
// To check permissions, we will split the update into "server changes" and "course changes".
// For example, a user may be a server admin, but they can srill be edited by a course admin if only course changes are being made.
func checkUpdatePermissions(newUser *model.ServerUser, oldUser *model.ServerUser, rawData *model.RawUserData, options UpsertUsersOptions) *model.UserOpResult {
	// Check permissions for server-level changes.
	hasServerChanges := rawData.HasServerInfo()
	if hasServerChanges {
		result := checkServerUpdatePermissions(newUser, oldUser, rawData, options)
		if result != nil {
			return result
		}
	}

	// Check permissions for course-level changes.
	hasCourseChanges := rawData.HasCourseInfo()
	if hasCourseChanges {
		result := checkCourseUpdatePermissions(newUser, oldUser, rawData, options)
		if result != nil {
			return result
		}
	}

	return nil
}

// Check permissions for updates on server-level data.
func checkServerUpdatePermissions(newUser *model.ServerUser, oldUser *model.ServerUser, rawData *model.RawUserData, options UpsertUsersOptions) *model.UserOpResult {
	// Server roles can only be modified by server admins.
	hasServerRoleChange := ((newUser.Role != model.ServerRoleUnknown) && (newUser.Role != oldUser.Role))
	if hasServerRoleChange && (options.ContextServerRole < model.ServerRoleAdmin) {
		return model.NewUserOpResultValidationError("-1015", newUser.Email,
			fmt.Errorf("User has a server role of '%s', which is not high enough to modify server roles.", options.ContextServerRole.String()))
	}

	// Cannot modify server data on a user that has higher server role.
	if options.ContextServerRole < oldUser.Role {
		return model.NewUserOpResultValidationError("-1016", newUser.Email,
			fmt.Errorf("User has a server role of '%s', which is not high enough to update a user with server role of '%s'.", options.ContextServerRole.String(), oldUser.Role.String()))
	}

	// Cannot modify server data unless you are an admin or self.
	if (oldUser.Email != options.ContextEmail) && (options.ContextServerRole < model.ServerRoleAdmin) {
		return model.NewUserOpResultValidationError("-1017", newUser.Email,
			fmt.Errorf("User has a server role of '%s', which is not high enough to update server-level information for another user.", options.ContextServerRole.String()))
	}

	return nil
}

// Check permissions for updates on course-level data.
func checkCourseUpdatePermissions(newUser *model.ServerUser, oldUser *model.ServerUser, rawData *model.RawUserData, options UpsertUsersOptions) *model.UserOpResult {
	// Server admins can do whatever they want on course information.
	if options.ContextServerRole >= model.ServerRoleAdmin {
		return nil
	}

	oldCourseRole := oldUser.GetCourseRole(rawData.Course)
	newCourseRole := newUser.GetCourseRole(rawData.Course)

	// Course roles can only be modified by course admins.
	hasCourseRoleChange := ((newCourseRole != model.CourseRoleUnknown) && (oldCourseRole != newCourseRole))
	if hasCourseRoleChange && (options.ContextCourseRole < model.CourseRoleAdmin) {
		return model.NewUserOpResultValidationError("-1018", newUser.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to modify course roles.", options.ContextCourseRole.String()))
	}

	// Cannot update course data on a user that has higher course role.
	if options.ContextCourseRole < oldCourseRole {
		return model.NewUserOpResultValidationError("-1019", newUser.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to update a user with course role of '%s'.", options.ContextCourseRole.String(), oldCourseRole.String()))
	}

	// Cannot modify course data unless you are an admin or self.
	if (oldUser.Email != options.ContextEmail) && (options.ContextCourseRole < model.CourseRoleAdmin) {
		return model.NewUserOpResultValidationError("-1020", newUser.Email,
			fmt.Errorf("User has a course role of '%s', which is not high enough to update course-level information for another user.", options.ContextCourseRole.String()))
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
			result.CommunicationError = model.NewLocatableError("-1021", false, "", fmt.Sprintf("Failed to send email: '%v'.", err))
			emailSent = false
		}
	}

	result.Emailed = emailSent
}
