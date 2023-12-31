package main

import (
    "bufio"
    "fmt"
    "os"
    "slices"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/procedures"
    "github.com/eriq-augustine/autograder/util"
)

type AddUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Name string `help:"Name for the user. Defaults to the user's email." short:"n"`
    Role string `help:"Role for the user. Defaults to student." short:"r" default:"student"`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
    Force bool `help:"Overwrite any existing user." short:"f" default:"false"`
    SendEmail bool `help:"Send an email to the user about adding them. Errors sending emails will be noted, but will not halt operations." default:"false"`
    DryRun bool `help:"Do not actually write out the user's file or send emails, just state what you would do." default:"false"`
    SyncLMS bool `help:"After adding users, sync the course users (all of them) with the course's LMS." default:"false"`
}

func (this *AddUser) Run(course *model.Course) error {
    role := model.GetRole(this.Role);
    if (role == model.RoleUnknown) {
        return fmt.Errorf("Unknown role: '%s'.", this.Role)
    }

    newUser := model.NewUser(this.Email, this.Name, role);

    // If set, the password comes in cleartext.
    if (this.Pass != "") {
        hashPass := util.Sha256HexFromString(this.Pass);
        newUser.Pass = hashPass;
    }

    result, err := db.SyncUser(course, newUser, this.Force, this.DryRun, this.SendEmail);
    if (err != nil) {
        return err;
    }

    if (this.DryRun) {
        fmt.Println("Doing a dry run, users file will not be written to.");
    }

    fmt.Println("Add Report:");
    fmt.Println(util.MustToJSONIndent(result));

    if (this.SyncLMS) {
        result, err = procedures.SyncLMSUserEmail(course, this.Email, this.DryRun, this.SendEmail);
        if (err != nil) {
            return err;
        }

        fmt.Println("\nLMS sync report:");
        fmt.Println(util.MustToJSONIndent(result));
    }

    return nil;
}

type AddTSV struct {
    TSV string `help:"Path to the TSV file containing the new users." arg:"" required:""`
    SkipRows int `help:"Number of initial rows to skip." default:"0"`
    Force bool `help:"Overwrite any existing users." short:"f" default:"false"`
    SendEmail bool `help:"Send an email to the user about adding them. Errors sending emails will be noted, but will not halt operations." default:"false"`
    DryRun bool `help:"Do not actually write out the user's file, just state what you would do." default:"false"`
    SyncLMS bool `help:"After adding users, sync the course users (all of them) with the course's LMS." default:"false"`
}

func (this *AddTSV) Run(course *model.Course) error {
    newUsers, err := readUsersTSV(this.TSV, this.SkipRows);
    if (err != nil) {
        return err;
    }

    result, err := db.SyncUsers(course, newUsers, this.Force, this.DryRun, this.SendEmail);
    if (err != nil) {
        return err;
    }

    if (this.DryRun) {
        fmt.Println("Doing a dry run, users file will not be written to.");
        fmt.Println(util.MustToJSONIndent(result));
    }

    if (this.SyncLMS) {
        emails := make([]string, 0, len(newUsers));
        for _, newUser := range newUsers {
            emails = append(emails, newUser.Email);
        }

        result, err = procedures.SyncLMSUserEmails(course, emails, this.DryRun, this.SendEmail);
        if (err != nil) {
            return err;
        }

        if (this.DryRun) {
            fmt.Println("LMS sync report:");
            fmt.Println(util.MustToJSONIndent(result));
        }
    }

    return nil;
}

type AuthUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
}

func (this *AuthUser) Run(course *model.Course) error {
    user, err := db.GetUser(course, this.Email);
    if (err != nil) {
        return fmt.Errorf("Failed to get user: '%w'.", err);
    }

    if (user == nil) {
        return fmt.Errorf("User '%s' does not exist, cannot auth.", this.Email);
    }

    passHash := util.Sha256Hex([]byte(this.Pass));

    if (user.CheckPassword(passHash)) {
        fmt.Println("Authentication Successful");
    } else {
        fmt.Println("Authentication Failed, Bad Password");
    }

    return nil;
}

type GetUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
}

func (this *GetUser) Run(course *model.Course) error {
    user, err := db.GetUser(course, this.Email);
    if (err != nil) {
        return fmt.Errorf("Failed to get user: '%w'.", err);
    }

    if (user == nil) {
        fmt.Printf("No user found with email '%s'.\n", this.Email);
    } else {
        fmt.Printf("Email: '%s', Name: '%s', Role: '%s'.\n", user.Email, user.Name, user.Role);
    }

    return nil
}

type ListUsers struct {
    All bool `help:"Show more info about each user." short:"a" default:"false"`
}

func (this *ListUsers) Run(course *model.Course) error {
    users, err := db.GetUsers(course);
    if (err != nil) {
        return fmt.Errorf("Failed to load users: '%w'.", err);
    }

    if (this.All) {
        fmt.Printf("%s\t%s\t%s\n", "Email", "Name", "Role");
    }

    emailList := make([]string, 0, len(users));
    for email, _ := range users {
        emailList = append(emailList, email);
    }
    slices.Sort(emailList);

    for _, email := range emailList {
        user := users[email];

        if (this.All) {
            fmt.Printf("%s\t%s\t%s\n", user.Email, user.Name, user.Role);
        } else {
            fmt.Println(user.Email);
        }
    }

    return nil
}

type ChangePassword struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
    SendEmail bool `help:"Send an email to the user." default:"false"`
}

func (this *ChangePassword) Run(course *model.Course) error {
    user, err := db.GetUser(course, this.Email);
    if (err != nil) {
        return fmt.Errorf("Failed to get user: '%w'.", err);
    }

    if (user == nil) {
        return fmt.Errorf("User '%s' does not exist.", this.Email);
    }

    user.Pass = this.Pass;

    result, err := db.SyncUser(course, user, true, false, this.SendEmail);
    if (err != nil) {
        return fmt.Errorf("Failed to sync user: '%w'.", err);
    }

    // Wait to the very end to output the generated password.
    if (len(result.ClearTextPasswords) > 0) {
        fmt.Printf("Generated password: '%s'.\n", result.ClearTextPasswords[user.Email]);
    }

    return nil
}

type RmUser struct {
    Email string `help:"Email for the user to be removed." arg:"" required:""`
}

func (this *RmUser) Run(course *model.Course) error {
    exists, err := db.RemoveUser(course, this.Email);
    if (err != nil) {
        return fmt.Errorf("Failed to remove user '%s': '%w'.", this.Email, err);
    }

    if (!exists) {
        return fmt.Errorf("User does not exist '%s'.", this.Email);
    }

    fmt.Printf("User '%s' removed.\n", this.Email);

    return nil;
}

var cli struct {
    config.ConfigArgs
    Course string `help:"ID of the course."`

    Add AddUser `cmd:"" help:"Add a user."`
    AddTSV AddTSV `cmd:"" help:"Add users from a TSV file formatted as: '<email>[\t<name>[\t<role>[\t<password>]]]'. See add for default values."`
    Auth AuthUser `cmd:"" help:"Authenticate as a user."`
    Get GetUser `cmd:"" help:"Get a user."`
    Ls ListUsers `cmd:"" help:"List users."`
    Rm RmUser `cmd:"" help:"Remove a user."`
}

func main() {
    context := kong.Parse(&cli,
        kong.Description("Manage users."),
    );

    err := config.HandleConfigArgs(cli.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(cli.Course);

    err = context.Run(course);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run command.");
    }
}

type TSVUser struct {
    User *model.User
    UserExists bool;
    GeneratedPass bool;
    CleartextPass string
}

// Read users from a TSV formatted as: '<email>[\t<name>[\t<role>[\t<cleartext password>]]]'.
// The users returned from this function are not official users yet.
// The users returned from here can be sent straight to db.SyncUsers() without any modifications.
func readUsersTSV(path string, skipRows int) (map[string]*model.User, error) {
    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open user TSV file '%s': '%w'.", path, err);
    }
    defer file.Close();

    newUsers := make(map[string]*model.User);

    scanner := bufio.NewScanner(file);
    scanner.Split(bufio.ScanLines);

    lineno := 0
    for scanner.Scan() {
        lineno++;
        if (skipRows > 0) {
            skipRows--;
            continue;
        }

        parts := strings.Split(scanner.Text(), "\t");

        var email string;
        var name string = "";
        var role model.UserRole = model.RoleStudent;

        if (len(parts) >= 1) {
            email = parts[0];
        } else {
            return nil, fmt.Errorf("User file '%s' line %d does not have enough fields.", path, lineno);
        }

        if (len(parts) >= 2) {
            name = parts[1];
        }

        if (len(parts) >= 3) {
            role = model.GetRole(parts[2]);
            if (role == model.RoleUnknown) {
                return nil, fmt.Errorf("User file '%s' line %d has unknwon role '%s'.", path, lineno, parts[2]);
            }
        }

        newUser := model.NewUser(email, name, role);

        if (len(parts) >= 4) {
            hashPass := util.Sha256HexFromString(parts[3]);
            newUser.Pass = hashPass;
        }

        if (len(parts) >= 5) {
            return nil, fmt.Errorf("User file '%s' line %d contains too many fields. Found %d, expecting at most %d.", path, lineno, len(parts), 4);
        }

        newUsers[email] = newUser;
    }

    return newUsers, nil;
}
