package db

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
)

func GetUsers(course *model.Course) (map[string]*usr.User, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetUsers(course);
}

func GetUsersFromID(courseID string) (map[string]*usr.User, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    course, err := GetCourse(courseID);
    if (err != nil) {
        return nil, err;
    }

    if (course == nil) {
        return nil, fmt.Errorf("Unknown course: '%s'.", courseID);
    }

    return GetUsers(course);
}

func GetUser(course *model.Course, email string) (*usr.User, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    user, err := backend.GetUser(course, email);
    if (err != nil) {
        return nil, err;
    }

    if (user == nil) {
        return nil, nil;
    }

    return user, nil;
}

// Insert the given users (overriding any conflicting users).
// For user merging (instead of overriding), user db.SyncUsers().
func SaveUsers(course *model.Course, users map[string]*usr.User) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return backend.SaveUsers(course, users);
}

// Remove a user.
// Returns a boolean indicating if the user exists.
// If true, then the user exists and was removed.
// If false (and the error is nil), then the user did not exist.
func RemoveUser(course *model.Course, email string) (bool, error) {
    if (backend == nil) {
        return false, fmt.Errorf("Database has not been opened.");
    }

    user, err := GetUser(course, email);
    if (err != nil) {
        return false, err;
    }

    if (user == nil) {
        return false, nil;
    }

    return true, backend.RemoveUser(course, email);
}

// Sync a single user to the database.
// See db.SyncUsers().
func SyncUser(course *model.Course, user *usr.User,
        merge bool, dryRun bool, sendEmails bool) (*usr.UserSyncResult, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    newUsers := map[string]*usr.User{
        user.Email: user,
    };

    return SyncUsers(course, newUsers, merge, dryRun, sendEmails);
}


// Sync (merge) new users with existing users.
// The db takes ownership of the passed-in users (they may be modified).
// If |merge| is true, then existing users will be updated with non-empty fields.
// Otherwise existing users will be ignored.
// Any non-ignored user WILL have their password changed.
// Passwords should either be left empty (and they will be randomly generated),
// or set to the hash of the desired password.
func SyncUsers(course *model.Course, newUsers map[string]*usr.User,
        merge bool, dryRun bool, sendEmails bool) (*usr.UserSyncResult, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    localUsers, err := GetUsers(course);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch local users: '%w'.", err);
    }

    // Summary of results.
    syncResult := usr.NewUserSyncResult();

    // Users that require saving in the DB.
    syncUsers := make(map[string]*usr.User);

    for _, newUser := range newUsers {
        localUser := localUsers[newUser.Email];

        if ((localUser != nil) && !merge) {
            // Skip.
            syncResult.AddResolveResult(&usr.UserResolveResult{Skip: localUser});
            continue;
        }

        if (newUser.Pass == "") {
            clearTextPass, err := newUser.SetRandomPassword();
            if (err != nil) {
                return nil, err;
            }

            syncResult.ClearTextPasswords[newUser.Email] = clearTextPass;
        } else {
            hashPass := newUser.Pass;

            err = newUser.SetPassword(hashPass);
            if (err != nil) {
                return nil, err;
            }
        }

        if (localUser == nil) {
            // New user.

            // Default role-less users to usr.Other.
            if (newUser.Role == usr.Unknown) {
                newUser.Role = usr.Other;
            }

            syncResult.AddResolveResult(&usr.UserResolveResult{Add: newUser});
            syncUsers[newUser.Email] = newUser;

            continue;
        }

        // Merge.
        changed := localUser.Merge(newUser);
        if (changed) {
            syncResult.AddResolveResult(&usr.UserResolveResult{Mod: localUser});
            syncUsers[newUser.Email] = newUser;
        }
    }

    if (dryRun) {
        return syncResult, nil;
    }

    err = SaveUsers(course, syncUsers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    if (sendEmails) {
        sleep := (len(newUsers) > 1);

        for _, newUser := range syncResult.Add {
            clearTextPass := syncResult.ClearTextPasswords[newUser.Email];
            usr.SendUserAddEmail(newUser, clearTextPass, (clearTextPass != ""), false, dryRun, sleep);
        }

        for _, newUser := range syncResult.Mod {
            clearTextPass := syncResult.ClearTextPasswords[newUser.Email];
            if (clearTextPass == "") {
                // Unlike new users, only send an email on generated passwords.
                continue;
            }

            usr.SendUserAddEmail(newUser, clearTextPass, true, true, dryRun, sleep);
        }
    }

    return syncResult, nil;
}
