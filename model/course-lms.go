package model

import (
    "fmt"

    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

type UserSyncResult struct {
    Add []*usr.User
    Mod []*usr.User
    Del []*usr.User

    ClearTextPasswords map[string]string
}

type UserResolveResult struct {
    Add *usr.User
    Mod *usr.User
    Del *usr.User

    ClearTextPassword string
}

// Sync users with the provided LMS.
func (this *Course) SyncLMSUsers(dryRun bool, sendEmails bool) (*UserSyncResult, error) {
    if (this.LMSAdapter == nil) {
        return nil, fmt.Errorf("Course '%s' has no adapter, cannot sync users.", this.ID);
    }

    localUsers, err := this.GetUsers();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch local users: '%w'.", err);
    }

    lmsUsersSlice, err := this.LMSAdapter.FetchUsers();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch LMS users: '%w'.", err);
    }

    lmsUsers := make(map[string]*lms.User, len(lmsUsersSlice));
    for _, lmsUser := range lmsUsersSlice {
        lmsUsers[lmsUser.Email] = lmsUser;
    }

    syncResult := UserSyncResult{
        Add: make([]*usr.User, 0),
        Mod: make([]*usr.User, 0),
        Del: make([]*usr.User, 0),
        ClearTextPasswords: make(map[string]string),
    }

    for _, email := range getAllEmails(localUsers, lmsUsers) {
        resolveResult, err := this.resolveUserSync(localUsers, lmsUsers, email);
        if (err != nil) {
            return nil, err;
        }

        if (resolveResult != nil) {
            syncResult.AddResolveResult(resolveResult);
        }
    }

    if (dryRun) {
        return &syncResult, nil;
    }

    err = this.SaveUsersFile(localUsers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    if (sendEmails) {
        for _, localUser := range localUsers {
            pass := syncResult.ClearTextPasswords[localUser.Email];
            usr.SendUserAddEmail(localUser, pass, true, false, dryRun, true);
        }
    }

    return &syncResult, nil;
}

// TEST
func (this *Course) SyncUserWithLMS(user *usr.User) error {
    if (this.LMSAdapter == nil) {
        return nil;
    }

    userInfo, err := this.LMSAdapter.FetchUser(user.Email)
    if (err != nil) {
        return err;
    }

    if (userInfo == nil) {
        return nil;
    }

    user.LMSID = userInfo.ID;

    if (userInfo.Name != "") {
        user.DisplayName = userInfo.Name;
    }

    return nil;
}

func mergeUsers(localUser *usr.User, lmsUser *lms.User, mergeAttributes bool) bool {
    changed := false;

    if (localUser.LMSID != lmsUser.ID) {
        localUser.LMSID = lmsUser.ID;
        changed = true;
    }

    if (!mergeAttributes) {
        return changed;
    }

    if (localUser.DisplayName != lmsUser.Name) {
        localUser.DisplayName = lmsUser.Name;
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
func (this *Course) resolveUserSync(localUsers map[string]*usr.User, lmsUsers map[string]*lms.User, email string) (
        *UserResolveResult, error) {
    localUser := localUsers[email];
    lmsUser := lmsUsers[email];

    if ((localUser == nil) && lmsUser == nil) {
        return nil, nil;
    }

    // Add.
    if (localUser == nil) {
        if (!this.LMSAdapter.SyncAddUsers) {
            return nil, nil;
        }

        pass, err := util.RandHex(usr.DEFAULT_PASSWORD_LEN);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to generate a default password.");
        }

        localUser = &usr.User{
            Email: email,
            DisplayName: lmsUser.Name,
            Role: lmsUser.Role,
            LMSID: lmsUser.ID,
        };

        hashPass := util.Sha256HexFromString(pass);
        localUser.SetPassword(hashPass);

        localUsers[email] = localUser;

        return &UserResolveResult{Add: localUser, ClearTextPassword: pass}, nil;
    }

    // Del.
    if (lmsUser == nil) {
        if (!this.LMSAdapter.SyncRemoveUsers) {
            return nil, nil;
        }

        delete(localUsers, email);
        return &UserResolveResult{Del: localUser}, nil;
    }

    // Mod.
    userChanged := mergeUsers(localUser, lmsUser, this.LMSAdapter.SyncUserAttributes);
    if (userChanged) {
        return &UserResolveResult{Mod: localUser}, nil;
    }

    return nil, nil;
}

func getAllEmails(localUsers map[string]*usr.User, lmsUsers map[string]*lms.User) []string {
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

func (this *UserSyncResult) Count() int {
    return len(this.Add) + len(this.Mod) + len(this.Del);
}

func (this *UserSyncResult) PrintReport() {
    groups := []struct{operation string; users []*usr.User}{
        {"Added", this.Add},
        {"Modified", this.Mod},
        {"Deleted", this.Del},
    };

    for i, group := range groups {
        if (i != 0) {
            fmt.Println();
        }

        fmt.Printf("%s %d users.\n", group.operation, len(group.users));
        for _, user := range group.users {
            fmt.Println("    " + user.ToRow(", "));
        }
    }
}

func (this *UserSyncResult) AddResolveResult(resolveResult *UserResolveResult) {
    if (resolveResult == nil) {
        return;
    }

    if (resolveResult.Add != nil) {
        this.Add = append(this.Add, resolveResult.Add);
        this.ClearTextPasswords[resolveResult.Add.Email] = resolveResult.ClearTextPassword;
    }

    if (resolveResult.Mod != nil) {
        this.Mod = append(this.Mod, resolveResult.Mod);
    }

    if (resolveResult.Del != nil) {
        this.Del = append(this.Del, resolveResult.Del);
    }
}
