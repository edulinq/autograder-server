package lmssync

import (
    "errors"
    "fmt"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/lms/lmstypes"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Sync all available aspects of the course with their LMS.
// Will return nil (with no error) if the course has no LMS.
func SyncLMS(course *model.Course, dryRun bool, sendEmails bool) (*model.LMSSyncResult, error) {
    if (!course.HasLMSAdapter()) {
        return nil, nil;
    }

    userSync, err := SyncAllLMSUsers(course, dryRun, sendEmails);
    if (err != nil) {
        return nil, err;
    }

    assignmentSync, err := syncAssignments(course, dryRun);
    if (err != nil) {
        return nil, err;
    }

    result := &model.LMSSyncResult{
        UserSync: userSync,
        AssignmentSync: assignmentSync,
    };

    return result, nil;
}

// Sync users with the provided LMS.
func SyncAllLMSUsers(course *model.Course, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
    lmsUsersSlice, err := lms.FetchUsers(course);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch LMS users: '%w'.", err);
    }

    lmsUsers := make(map[string]*lmstypes.User, len(lmsUsersSlice));
    for _, lmsUser := range lmsUsersSlice {
        if (lmsUser.Email != "") {
            lmsUsers[lmsUser.Email] = lmsUser;
        }
    }

    return syncLMSUsers(course, dryRun, sendEmails, lmsUsers, nil);
}

func SyncLMSUserEmail(course *model.Course, email string, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
    return SyncLMSUserEmails(course, []string{email}, dryRun, sendEmails);
}

func SyncLMSUserEmails(course *model.Course, emails []string, dryRun bool, sendEmails bool) (*model.UserSyncResult, error) {
    lmsUsers := make(map[string]*lmstypes.User);

    for _, email := range emails {
        lmsUser, err := lms.FetchUser(course, email);
        if (err != nil) {
            return nil, err;
        }

        lmsUsers[lmsUser.Email] = lmsUser;
    }

    return syncLMSUsers(course, dryRun, sendEmails, lmsUsers, emails);
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
            err = errors.Join(err, model.SendUserAddEmail(course, newUser, pass, true, false, dryRun, true));
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
        if (!adapter.SyncUserAdds) {
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
        if (!adapter.SyncUserRemoves) {
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
        if (email == "") {
            continue;
        }

        _, ok := localUsers[email];
        if (ok) {
            // Already added.
            continue;
        }

        emails = append(emails, email);
    }

    return emails;
}

func syncAssignments(course *model.Course, dryRun bool) (*model.AssignmentSyncResult, error) {
    result := model.NewAssignmentSyncResult();

    adapter := course.GetLMSAdapter();
    if (!adapter.SyncAssignments) {
        return result, nil;
    }

    lmsAssignments, err := lms.FetchAssignments(course);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to get assignments: '%w'.", err);
    }

    localAssignments := course.GetAssignments();

    // Match local assignments to LMS assignments.
    matches := make(map[string]int);
    for _, localAssignment := range localAssignments {
        localID := localAssignment.GetID();
        localName := localAssignment.GetName();
        lmsID := localAssignment.GetLMSID();

        for i, lmsAssignment := range lmsAssignments {
            matchIndex := -1;

            if (lmsID != "") {
                // Exact ID match.
                if (lmsID == lmsAssignment.ID) {
                    matchIndex = i;
                }
            } else {
                // Name match.
                if ((localName != "") && (localName == lmsAssignment.Name)) {
                    matchIndex = i;
                }
            }

            if (matchIndex != -1) {
                _, exists := matches[localID];
                if (exists) {
                    delete(matches, localID);
                    result.AmbiguousMatches = append(result.AmbiguousMatches, model.AssignmentInfo{localID, localName});
                    break;
                }

                matches[localID] = matchIndex;
            }
        }

        _, exists := matches[localID];
        if (!exists) {
            result.NonMatchedAssignments = append(result.NonMatchedAssignments, model.AssignmentInfo{localID, localName});
        }
    }

    for localID, lmsIndex := range matches {
        localName := localAssignments[localID].GetName();
        changed := mergeAssignment(localAssignments[localID], lmsAssignments[lmsIndex]);
        if (changed) {
            result.SyncedAssignments = append(result.SyncedAssignments, model.AssignmentInfo{localID, localName});
        } else {
            result.UnchangedAssignments = append(result.UnchangedAssignments, model.AssignmentInfo{localID, localName});
        }
    }

    if (!dryRun) {
        err = db.SaveCourse(course);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to save course: '%w'.", err);
        }
    }

    return result, nil;
}

func mergeAssignment(localAssignment *model.Assignment, lmsAssignment *lmstypes.Assignment) bool {
    changed := false;

    if (localAssignment.LMSID == "") {
        localAssignment.LMSID = lmsAssignment.ID;
        changed = true;
    }

    if ((localAssignment.Name == "") && (lmsAssignment.Name != "")) {
        localAssignment.Name = lmsAssignment.Name;
        changed = true;
    }

    if (localAssignment.DueDate.IsZero() && (lmsAssignment.DueDate != nil) && !lmsAssignment.DueDate.IsZero()) {
        localAssignment.DueDate = common.TimestampFromTime(*lmsAssignment.DueDate);
        changed = true;
    }

    if (util.IsZero(localAssignment.MaxPoints) && !util.IsZero(lmsAssignment.MaxPoints)) {
        localAssignment.MaxPoints = lmsAssignment.MaxPoints;
        changed = true;
    }

    return changed;
}
