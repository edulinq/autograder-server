package types

import (
    "path/filepath"

    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const USERS_FILENAME = "users.json"

// Load users from a file adjacent to the course config (if it exists).
func loadStaticUsers(courseConfigPath string) (map[string]*usr.User, error) {
    users := make(map[string]*usr.User);

    path := filepath.Join(filepath.Dir(courseConfigPath), USERS_FILENAME);
    if (!util.PathExists(path)) {
        return users, nil;
    }

    err := util.JSONFromFile(path, &users);
    if (err != nil) {
        return nil, err;
    }

    return users, nil;
}
