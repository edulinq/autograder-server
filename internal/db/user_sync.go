package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
)

// Sync a single user to the database.
// See db.SyncUsers().
func SyncCourseUser(course *model.Course, user *model.CourseUser, merge bool, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	newUsers := map[string]*model.CourseUser{
		user.Email: user,
	}

	return SyncCourseUsers(course, newUsers, merge, dryRun, sendEmails)
}

// TEST - Password semantics changed!

// TEST - Do we want this, or more specific functions for each use case?

// TEST - check comment.
// Sync (merge) new course users with existing course users.
// The db takes ownership of the passed-in users (they may be modified).
// If |merge| is true, then existing users will be updated with non-empty fields.
// Otherwise existing users will be ignored.
// Any non-ignored user WILL have their password changed.
// Passwords should either be left empty (and they will be randomly generated),
// or set to the hash of the desired password.
func SyncCourseUsers(course *model.Course, newUsers map[string]*model.CourseUser, merge bool, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
	// TEST
	return model.NewUserSyncResult(), nil

	/* TEST
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	localUsers, err := GetCourseUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch local users: '%w'.", err)
	}

	// Summary of results.
	syncResult := model.NewUserSyncResult()

	// Users that require saving in the DB.
	syncUsers := make(map[string]*model.User)

	for _, newUser := range newUsers {
		localUser := localUsers[newUser.Email]

		if (localUser != nil) && !merge {
			// Skip.
			syncResult.AddResolveResult(&model.UserResolveResult{Skip: localUser})
			continue
		}

		if newUser.Pass == "" {
			clearTextPass, err := newUser.SetRandomPassword()
			if err != nil {
				return nil, err
			}

			syncResult.ClearTextPasswords[newUser.Email] = clearTextPass
		} else {
			hashPass := newUser.Pass

			err = newUser.SetPassword(hashPass)
			if err != nil {
				return nil, err
			}
		}

		if localUser == nil {
			// New user.

			// Default role-less users to model.RoleOther.
			if newUser.Role == model.RoleUnknown {
				newUser.Role = model.RoleOther
			}

			syncResult.AddResolveResult(&model.UserResolveResult{Add: newUser})
			syncUsers[newUser.Email] = newUser

			continue
		}

		// Merge.
		changed := localUser.Merge(newUser)
		if changed {
			syncResult.AddResolveResult(&model.UserResolveResult{Mod: localUser})
			syncUsers[newUser.Email] = localUser
		} else {
			syncResult.AddResolveResult(&model.UserResolveResult{Unchanged: localUser})
		}
	}

	if dryRun {
		return syncResult, nil
	}

	err = SaveUsers(course, syncUsers)
	if err != nil {
		return nil, fmt.Errorf("Failed to save users file: '%w'.", err)
	}

	if sendEmails {
		sleep := (len(newUsers) > 1)

		err = nil

		for _, newUser := range syncResult.Add {
			clearTextPass := syncResult.ClearTextPasswords[newUser.Email]
			err = errors.Join(err, model.SendUserAddEmail(course, newUser, clearTextPass, (clearTextPass != ""), false, dryRun, sleep))
		}

		for _, newUser := range syncResult.Mod {
			clearTextPass := syncResult.ClearTextPasswords[newUser.Email]
			if clearTextPass == "" {
				// Unlike new users, only send an email on generated passwords.
				continue
			}

			err = errors.Join(err, model.SendUserAddEmail(course, newUser, clearTextPass, true, true, dryRun, sleep))
		}

		if err != nil {
			return nil, fmt.Errorf("Failed to send user email: '%w'.", err)
		}
	}

	return syncResult, nil
	*/
}
