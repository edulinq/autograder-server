package lmsusers

import (
    "errors"
    "fmt"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/lms/lmstypes"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Sync users with the provided LMS.
func SyncLMSUsers(course *model.Course, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
    lmsUsersSlice, err := lms.FetchUsers(course);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch LMS users: '%w'.", err);
    }

    lmsUsers := make(map[string]*lmstypes.User, len(lmsUsersSlice));
    for _, lmsUser := range lmsUsersSlice {
        lmsUsers[lmsUser.Email] = lmsUser;
    }

    return syncLMSUsers(course, dryRun, sendEmails, lmsUsers, nil);
}

func SyncLMSUser(course *model.Course, email string, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
    lmsUser, err := lms.FetchUser(course, email);
    if (err != nil) {
        return nil, err;
    }

    lmsUsers := map[string]*lmstypes.User{
        lmsUser.Email: lmsUser,
    };

    return syncLMSUsers(course, dryRun, sendEmails, lmsUsers, []string{email});
}

// Sync users.
// If |syncEmails| is not empty, then only emails in it will be checked/resolved.
// Otherwise, all emails from local and LMS users will be checked.
func syncLMSUsers(course *model.Course, dryRun bool, sendEmails bool, lmsUsers map[string]*lmstypes.User,
        syncEmails []string) (*model.UserSyncResult, error) {
    localUsers, err := db.GetUsers(course);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch local users: '%w'.", err);
    }

    if (len(syncEmails) == 0) {
        syncEmails = getAllEmails(localUsers, lmsUsers);
    }

    syncResult := model.NewUserSyncResult();

    for _, email := range syncEmails {
        resolveResult, err := resolveUserSync(course, localUsers, lmsUsers, email);
        if (err != nil) {
            return nil, err;
        }

        if (resolveResult != nil) {
            syncResult.AddResolveResult(resolveResult);
        }
    }

    if (dryRun) {
        return syncResult, nil;
    }

    err = db.SaveUsers(course, localUsers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    if (sendEmails) {
        err = nil;

        for _, newUser := range syncResult.Add {
            pass := syncResult.ClearTextPasswords[newUser.Email];
            err = errors.Join(err, model.SendUserAddEmail(newUser, pass, true, false, dryRun, true));
        }

        if (err != nil) {
            return nil, fmt.Errorf("Failed to send user email: '%w'.", err);
        }
    }

    return syncResult, nil;
}

func mergeUsers(localUser *model.User, lmsUser *lmstypes.User, mergeAttributes bool) bool {
    changed := false;

    if (localUser.LMSID != lmsUser.ID) {
        localUser.LMSID = lmsUser.ID;
        changed = true;
    }

    if (!mergeAttributes) {
        return changed;
    }

    if (localUser.Name != lmsUser.Name) {
        localUser.Name = lmsUser.Name;
        changed = true;
    }

    if (localUser.Role != lmsUser.Role) {
        localUser.Role = lmsUser.Role;
        changed = true;
    }

    return changed;
}

// Resolve differences between a local user and LMS user (linked using the provided email).
// The passed in local user map will be modified to reflect any resolution.
// The taken action will depend on the options set in the course's LMS adapter.
func resolveUserSync(course *model.Course, localUsers map[string]*model.User,
        lmsUsers map[string]*lmstypes.User, email string) (*model.UserResolveResult, error) {
    adapter := course.GetLMSAdapter();

    localUser := localUsers[email];
    lmsUser := lmsUsers[email];

    if ((localUser == nil) && lmsUser == nil) {
        return nil, nil;
    }

    // Add.
    if (localUser == nil) {
        if (!adapter.SyncAddUsers) {
            return nil, nil;
        }

        pass, err := util.RandHex(model.DEFAULT_PASSWORD_LEN);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to generate a default password: '%w'.", err);
        }

        localUser = &model.User{
            Email: email,
            Name: lmsUser.Name,
            Role: lmsUser.Role,
            LMSID: lmsUser.ID,
        };

        hashPass := util.Sha256HexFromString(pass);
        err = localUser.SetPassword(hashPass);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to set password: '%w'.", err);
        }

        localUsers[email] = localUser;

        return &model.UserResolveResult{Add: localUser, ClearTextPassword: pass}, nil;
    }

    // Del.
    if (lmsUser == nil) {
        if (!adapter.SyncRemoveUsers) {
            return &model.UserResolveResult{Unchanged: localUser}, nil;
        }

        delete(localUsers, email);
        return &model.UserResolveResult{Del: localUser}, nil;
    }

    // Mod.
    userChanged := mergeUsers(localUser, lmsUser, adapter.SyncUserAttributes);
    if (userChanged) {
        return &model.UserResolveResult{Mod: localUser}, nil;
    }

    return &model.UserResolveResult{Unchanged: localUser}, nil;
}

func getAllEmails(localUsers map[string]*model.User, lmsUsers map[string]*lmstypes.User) []string {
    emails := make([]string, 0, max(len(localUsers), len(lmsUsers)));

    for email, _ := range localUsers {
        emails = append(emails, email);
    }

    for email, _ := range lmsUsers {
        _, ok := localUsers[email];
        if (ok) {
            // Already added.
            continue;
        }

        emails = append(emails, email);
    }

    return emails;
}
