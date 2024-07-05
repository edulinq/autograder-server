package lmssync

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/users"
)

// Sync users with the provided LMS.
func SyncAllLMSUsers(course *model.Course, dryRun bool, sendEmails bool) ([]*model.UserOpResult, error) {
	lmsUsersSlice, err := lms.FetchUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch LMS users: '%w'.", err)
	}

	lmsUsers := make(map[string]*lmstypes.User, len(lmsUsersSlice))
	for _, lmsUser := range lmsUsersSlice {
		if lmsUser.Email != "" {
			lmsUsers[lmsUser.Email] = lmsUser
		}
	}

	return syncLMSUsers(course, dryRun, sendEmails, false, lmsUsers)
}

func SyncLMSUserEmail(course *model.Course, email string, dryRun bool, sendEmails bool) ([]*model.UserOpResult, error) {
	return SyncLMSUserEmails(course, []string{email}, dryRun, sendEmails)
}

func SyncLMSUserEmails(course *model.Course, emails []string, dryRun bool, sendEmails bool) ([]*model.UserOpResult, error) {
	lmsUsers := make(map[string]*lmstypes.User)

	for _, email := range emails {
		lmsUser, err := lms.FetchUser(course, email)
		if err != nil {
			return nil, err
		}

		if lmsUser != nil {
			lmsUsers[lmsUser.Email] = lmsUser
		}
	}

	return syncLMSUsers(course, dryRun, sendEmails, true, lmsUsers)
}

// Sync LMS users.
// Note that |skipMissing| makes it so that only users in |lmsUsers| will be considered.
// This means that deletes will never be processed (since they are always in the LMS).
func syncLMSUsers(course *model.Course, dryRun bool, sendEmails bool, skipMissing bool, lmsUsers map[string]*lmstypes.User) ([]*model.UserOpResult, error) {
	adapter := course.GetLMSAdapter()
	if adapter == nil {
		return make([]*model.UserOpResult, 0), nil
	}

	if !adapter.SyncUsers() {
		results := make([]*model.UserOpResult, 0, len(lmsUsers))
		for email, _ := range lmsUsers {
			results = append(results, &model.UserOpResult{Email: email, Skipped: true})
		}
		return results, nil
	}

	courseUsers, err := db.GetCourseUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch course users: '%w'.", err)
	}

	usersData := make([]*model.RawUserData, 0, len(lmsUsers))
	for _, lmsUser := range lmsUsers {
		usersData = append(usersData, lmsUser.ToRawUserData(course.GetID()))
	}

	upsertOptions := users.UpsertUsersOptions{
		RawUsers:          usersData,
		SkipInserts:       !adapter.SyncUserAdds,
		SkipUpdates:       !adapter.SyncUserAttributes,
		SendEmails:        sendEmails,
		DryRun:            dryRun,
		ContextServerRole: model.ServerRoleRoot,
	}

	results := users.UpsertUsers(upsertOptions)

	// Remove any remaining users from the course.
	if skipMissing {
		return results, nil
	}

	removeEmails := make([]string, 0)
	for email, _ := range courseUsers {
		_, exists := lmsUsers[email]
		if !exists {
			removeEmails = append(removeEmails, email)
		}
	}

	for _, email := range removeEmails {
		if adapter.SyncUserRemoves {
			_, _, err := db.RemoveUserFromCourse(course, email)
			if err != nil {
				results = append(results, &model.UserOpResult{
					Email:        email,
					SystemErrors: []string{err.Error()},
				})
			} else {
				results = append(results, &model.UserOpResult{
					Email:    email,
					Modified: true,
					Dropped:  []string{course.GetID()},
				})
			}
		} else {
			results = append(results, &model.UserOpResult{
				Email:   email,
				Skipped: true,
			})
		}
	}

	return results, nil
}
