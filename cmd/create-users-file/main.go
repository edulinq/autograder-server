package main

import (
    "fmt"
    "path/filepath"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/model"
    "github.com/edulinq/autograder/internal/util"
)

var args struct {
    config.ConfigArgs
    Path string `help:"Path to the new users JSON file." arg:"" type:"path"`

    Email string `help:"Email for the new user." arg:"" required:""`
    Name string `help:"Name for the new user." short:"n"`
    Role string `help:"Role for the new user. Defaults to student." short:"r" default:"student"`
    Pass string `help:"Password for the new user. Defaults to a random string (will be output)." short:"p"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Create a users file, which can be used to seed users in a course." +
            " A full user must be specified, which will be added to the file." +
            " If an existing users file is specified, then the user will be added to the file."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    role := model.GetRole(args.Role);
    if (role == model.RoleUnknown) {
        log.Fatal("Unknown role.", log.NewAttr("role", args.Role));
    }

    users := getUsers(args.Path);

    newUser, generatedPass := makeNewUser(args.Email, args.Name, args.Pass, role);

    users[newUser.Email] = newUser;

    err = util.ToJSONFileIndent(users, args.Path);
    if (err != nil) {
        log.Fatal("Failed to write users file.", log.NewAttr("path", args.Path), err);
    }

    fmt.Printf("Successfully wrote user's file: '%s'.\n", args.Path);

    if (generatedPass != "") {
        fmt.Printf("Generated Password: '%s'.\n", generatedPass);
    }
}

func getUsers(path string) map[string]*model.User {
    var err error;

    if (!util.PathExists(path)) {
        err = util.MkDir(filepath.Dir(path));
        if (err != nil) {
            log.Fatal("Failed to create dir.", log.NewAttr("path", path), err);
        }

        return make(map[string]*model.User, 0);
    }

    var users map[string]*model.User;
    err = util.JSONFromFile(path, &users);
    if (err != nil) {
        log.Fatal("Failed to read existing users file.", log.NewAttr("path", path), err);
    }

    return users;
}

func makeNewUser(email string, name string, pass string, role model.UserRole) (*model.User, string) {
    newUser := model.NewUser(email, name, role);

    var err error;
    generatedPass := "";

    // If set, the password comes in cleartext.
    if (pass != "") {
        hashPass := util.Sha256HexFromString(pass);
        err = newUser.SetPassword(hashPass);
        if (err != nil) {
            log.Fatal("Failed to set password.", err);
        }
    } else {
        generatedPass, err = newUser.SetRandomPassword();
        if (err != nil) {
            log.Fatal("Failed to set random password.", err);
        }
    }

    return newUser, generatedPass;
}
