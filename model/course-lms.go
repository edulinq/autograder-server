package model

import (
    "fmt"
)

// Sync IDs with matching LMS users (does not add/remove users).
func (this *Course) SyncLMSUsers() (int, error) {
    users, err := this.GetUsers();
    if (err != nil) {
        return 0, fmt.Errorf("Failed to fetch local users: '%w'.", err);
    }

    lmsUsers, err := this.LMSAdapter.FetchUsers();
    if (err != nil) {
        return 0, fmt.Errorf("Failed to fetch LMS users: '%w'.", err);
    }

    count := 0
    for _, lmsUser := range lmsUsers {
        user := users[lmsUser.Email]
        if (user == nil) {
            continue;
        }

        if ((user.LMSID == lmsUser.ID) && (user.DisplayName == lmsUser.Name)) {
            continue;
        }

        user.LMSID = lmsUser.ID;
        user.DisplayName = lmsUser.Name;
        count++;
    }

    err = this.SaveUsersFile(users);
    if (err != nil) {
        return 0, fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    return count, nil;
}
