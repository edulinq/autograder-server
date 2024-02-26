package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edulinq/autograder/log"
	"github.com/edulinq/autograder/model"
)

func GetUsers(course *model.Course) (map[string]*model.User, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetUsers(course)
}

func MustGetUsers(course *model.Course) map[string]*model.User {
	users, err := GetUsers(course)
	if err != nil {
		log.Fatal("Failed to get users.", err, course)
	}

	return users
}

func GetUsersFromID(courseID string) (map[string]*model.User, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	course, err := GetCourse(courseID)
	if err != nil {
		return nil, err
	}

	if course == nil {
		return nil, fmt.Errorf("Unknown course: '%s'.", courseID)
	}

	return GetUsers(course)
}

// Insert method that takes a course id, email [] string and resolves based on *, roles, emails to all emails
func ResolveUsers(course *model.Course, emails []string) ([]string, error) {
    if backend == nil {
        return nil, fmt.Errorf("Database has not been opened.")
    }
    emailSet := map[string]any{}
    roleSet := map[model.UserRole]any{}
    // check for emails, roles, and * (all)
    for _, email := range emails {
        if strings.Contains(email, "@") {
            emailSet[email] = nil
        } else {
            if email == "*" {
                allRoles := model.GetAllRoles()
                for role := range allRoles {
                    roleSet[role] = nil
                }
            } else {
                roleSet[model.GetRole(email)] = nil
            }
        }
    }
    if len(roleSet) > 0 {
        users, err := GetUsers(course)
        if err != nil {
            return nil, err
        }
        for _, user := range users {
            _, ok := roleSet[user.Role] // in the role set
            if ok {
                emailSet[user.Email] = nil
            }
        }
    }
    emailSlice := make([]string, 0, len(emailSet));
    for email := range emailSet {
        emailSlice = append(emailSlice, email)
    }
    return emails, nil
}

func GetUser(course *model.Course, email string) (*model.User, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	user, err := backend.GetUser(course, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return user, nil
}

// Insert the given users (overriding any conflicting users).
// For user merging (instead of overriding), user db.SyncUsers().
func SaveUsers(course *model.Course, users map[string]*model.User) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.SaveUsers(course, users)
}

// Convenience function for SaveUsers() with a single user.
func SaveUser(course *model.Course, user *model.User) error {
	users := map[string]*model.User{user.Email: user}
	return SaveUsers(course, users)
}

// Remove a user.
// Returns a boolean indicating if the user exists.
// If true, then the user exists and was removed.
// If false (and the error is nil), then the user did not exist.
func RemoveUser(course *model.Course, email string) (bool, error) {
	if backend == nil {
		return false, fmt.Errorf("Database has not been opened.")
	}

	user, err := GetUser(course, email)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return true, backend.RemoveUser(course, email)
}

// Sync a single user to the database.
// See db.SyncUsers().
func SyncUser(course *model.Course, user *model.User,
	merge bool, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	newUsers := map[string]*model.User{
		user.Email: user,
	}

	return SyncUsers(course, newUsers, merge, dryRun, sendEmails)
}

// Sync (merge) new users with existing users.
// The db takes ownership of the passed-in users (they may be modified).
// If |merge| is true, then existing users will be updated with non-empty fields.
// Otherwise existing users will be ignored.
// Any non-ignored user WILL have their password changed.
// Passwords should either be left empty (and they will be randomly generated),
// or set to the hash of the desired password.
func SyncUsers(course *model.Course, newUsers map[string]*model.User,
	merge bool, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	localUsers, err := GetUsers(course)
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
}
