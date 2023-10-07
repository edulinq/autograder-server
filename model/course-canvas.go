package model

import (
    "fmt"

    "github.com/eriq-augustine/autograder/canvas"
)

// Sync IDs with matching canvas users (does not add/remove users).
func (this *Course) SyncCanvasUsers() (int, error) {
    users, err := this.GetUsers();
    if (err != nil) {
        return 0, fmt.Errorf("Failed to fetch local users: '%w'.", err);
    }

    canvasUsers, err := canvas.FetchUsers(this.CanvasInstanceInfo);
    if (err != nil) {
        return 0, fmt.Errorf("Failed to fetch canvas users: '%w'.", err);
    }

    count := 0
    for _, canvasUser := range canvasUsers {
        user := users[canvasUser.Email]
        if (user == nil) {
            continue;
        }

        if (user.CanvasID == canvasUser.ID) {
            continue;
        }

        user.CanvasID = canvasUser.ID;
        count++;
    }

    err = this.SaveUsersFile(users);
    if (err != nil) {
        return 0, fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    return count, nil;
}
