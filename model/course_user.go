package model

import (
    "fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
)

func (this *Course) GetUsers() (map[string]*usr.User, error) {
    path := filepath.Join(filepath.Dir(this.SourcePath), this.UsersFile);

    users, err := usr.LoadUsersFile(path);
    if (err != nil) {
        return nil, fmt.Errorf("Faile to deserialize users file '%s': '%w'.", path, err);
    }

    return users, nil;
}

func (this *Course) GetUser(email string) (*usr.User, error) {
    users, err := this.GetUsers();
    if (err != nil) {
        return nil, err;
    }

    user := users[email];
    if (user != nil) {
        return user, nil;
    }

    return nil, nil;
}

func (this *Course) SaveUsersFile(users map[string]*usr.User) error {
    path := filepath.Join(filepath.Dir(this.SourcePath), this.UsersFile);

    // Do not save user files in testing mode.
    if (config.TESTING_MODE.GetBool()) {
        return nil;
    }

    return usr.SaveUsersFile(path, users);
}

func (this *Course) AddUser(user *usr.User, merge bool, dryRun bool, sendEmails bool) (*usr.UserSyncResult, error) {
    newUsers := map[string]*usr.User{
        user.Email: user,
    };

    return this.SyncNewUsers(newUsers, merge, dryRun, sendEmails);
}

// Sync users with the new users.
// The course takes ownership of the passed-in users (they may be modified).
// If |merge| is true, then existing users will be updated with non-empty fields.
// Otherwise existing users will be ignored.
// Any non-ignored user WILL have their password changed.
// Passwords should either be left empty (and they will be randomly generated),
// or set to the hash of the desired password.
func (this *Course) SyncNewUsers(newUsers map[string]*usr.User, merge bool, dryRun bool, sendEmails bool) (*usr.UserSyncResult, error) {
    localUsers, err := this.GetUsers();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch local users: '%w'.", err);
    }

    syncResult := usr.NewUserSyncResult();

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
            localUsers[newUser.Email] = newUser;
            syncResult.AddResolveResult(&usr.UserResolveResult{Add: newUser});
            continue;
        }

        // Merge.
        changed := localUser.Merge(newUser);
        if (changed) {
            syncResult.AddResolveResult(&usr.UserResolveResult{Mod: localUser});
        }
    }

    if (dryRun) {
        return syncResult, nil;
    }

    err = this.SaveUsersFile(localUsers);
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
