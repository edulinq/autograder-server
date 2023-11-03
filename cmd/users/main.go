package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms/lmsusers"
    "github.com/eriq-augustine/autograder/model2"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

type AddUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Name string `help:"Display name for the user. Defaults to the user's email." short:"n"`
    Role string `help:"Role for the user. Defaults to student." short:"r" default:"student"`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
    Force bool `help:"Overwrite any existing user." short:"f" default:"false"`
    SendEmail bool `help:"Send an email to the user about adding them. Errors sending emails will be noted, but will not halt operations." default:"false"`
    DryRun bool `help:"Do not actually write out the user's file or send emails, just state what you would do." default:"false"`
    SyncLMS bool `help:"After adding users, sync the course users (all of them) with the course's LMS." default:"false"`
}

func (this *AddUser) Run(course model2.Course) error {
    role := usr.GetRole(this.Role);
    if (role == usr.Unknown) {
        return fmt.Errorf("Unknown role: '%s'.", this.Role)
    }

    newUser := usr.NewUser(this.Email, this.Name, role);

    // If set, the password comes in cleartext.
    if (this.Pass != "") {
        hashPass := util.Sha256HexFromString(this.Pass);
        newUser.Pass = hashPass;
    }

    result, err := course.AddUser(newUser, this.Force, this.DryRun, this.SendEmail);
    if (err != nil) {
        return err;
    }

    if (this.DryRun) {
        fmt.Println("Doing a dry run, users file will not be written to.");
    }

    fmt.Println("Add Report:");
    result.PrintReport();

    if (this.SyncLMS) {
        result, err = lmsusers.SyncLMSUser(course, this.Email, this.DryRun, this.SendEmail);
        if (err != nil) {
            return err;
        }

        fmt.Println("\nLMS sync report:");
        result.PrintReport();
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

func (this *AddTSV) Run(course model2.Course) error {
    newUsers, err := readUsersTSV(this.TSV, this.SkipRows);
    if (err != nil) {
        return err;
    }

    result, err := course.SyncNewUsers(newUsers, this.Force, this.DryRun, this.SendEmail);
    if (err != nil) {
        return err;
    }

    if (this.DryRun) {
        fmt.Println("Doing a dry run, users file will not be written to.");
        result.PrintReport();
    }

    if (this.SyncLMS) {
        result, err = lmsusers.SyncLMSUsers(course, this.DryRun, this.SendEmail);
        if (err != nil) {
            return err;
        }

        if (this.DryRun) {
            fmt.Println("LMS sync report:");
            result.PrintReport();
        }
    }

    return nil;
}

type AuthUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
}

func (this *AuthUser) Run(course model2.Course) error {
    user, err := course.GetUser(this.Email);
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

func (this *GetUser) Run(course model2.Course) error {
    user, err := course.GetUser(this.Email);
    if (err != nil) {
        return fmt.Errorf("Failed to get user: '%w'.", err);
    }

    if (user == nil) {
        fmt.Printf("No user found with email '%s'.\n", this.Email);
    } else {
        fmt.Printf("Email: '%s', Name: '%s', Role: '%s'.\n", user.Email, user.DisplayName, user.Role);
    }

    return nil
}

type ListUsers struct {
    All bool `help:"Show more info about each user." short:"a" default:"false"`
}

func (this *ListUsers) Run(course model2.Course) error {
    users, err := course.GetUsers();
    if (err != nil) {
        return fmt.Errorf("Failed to load users: '%w'.", err);
    }

    if (this.All) {
        fmt.Printf("%s\t%s\t%s\n", "Email", "Display Name", "Role");
    }

    for _, user := range users {
        if (this.All) {
            fmt.Printf("%s\t%s\t%s\n", user.Email, user.DisplayName, user.Role);
        } else {
            fmt.Println(user.Email);
        }
    }

    return nil
}

type ChangePassword struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
}

func (this *ChangePassword) Run(course model2.Course) error {
    users, err := course.GetUsers();
    if (err != nil) {
        return fmt.Errorf("Failed to get users: '%w'.", err);
    }

    user := users[this.Email];
    if (user == nil) {
        return fmt.Errorf("No user found with email '%s'.", this.Email);
    }

    generatedPass := false;
    if (this.Pass == "") {
        this.Pass, err = util.RandHex(usr.DEFAULT_PASSWORD_LEN);
        if (err != nil) {
            return fmt.Errorf("Failed to generate a default password.");
        }

        generatedPass = true;
    }

    pass := util.Sha256Hex([]byte(this.Pass));

    err = user.SetPassword(pass);
    if (err != nil) {
        return fmt.Errorf("Could not set password: '%w'.", err);
    }

    err = course.SaveUsers(users);
    if (err != nil) {
        return fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    // Wait to the very end to output the generated password.
    if (generatedPass) {
        fmt.Printf("Generated password: '%s'.\n", this.Pass);
    }

    return nil
}

type RmUser struct {
    Email string `help:"Email for the user to be removed." arg:"" required:""`
}

func (this *RmUser) Run(course model2.Course) error {
    users, err := course.GetUsers();
    if (err != nil) {
        return fmt.Errorf("Failed to get users: '%w'.", err);
    }

    user := users[this.Email];
    if (user == nil) {
        return fmt.Errorf("User '%s' does not exist, cannot remove.", this.Email);
    }

    delete(users, this.Email);

    err = course.SaveUsers(users);
    if (err != nil) {
        return fmt.Errorf("Failed to save users file: '%w'.", err);
    }

    fmt.Printf("User '%s' removed.\n", this.Email);

    return nil;
}

var cli struct {
    config.ConfigArgs
    CoursePath string `help:"Path to course JSON file." type:"existingfile"`

    Add AddUser `cmd:"" help:"Add a user."`
    AddTSV AddTSV `cmd:"" help:"Add users from a TSV file formatted as: '<email>[\t<display name>[\t<role>[\t<password>]]]'. See add for default values."`
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

    if (cli.CoursePath == "") {
        log.Fatal().Msg("--course-path must be supplied.");
    }

    course := db.MustLoadCourseConfig(cli.CoursePath);

    err = context.Run(course);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run command.");
    }
}

type TSVUser struct {
    User *usr.User
    UserExists bool;
    GeneratedPass bool;
    CleartextPass string
}

// Read users from a TSV formatted as: '<email>[\t<display name>[\t<role>[\t<cleartext password>]]]'.
// The users returned from this function are not official users yet.
// The users returned from here can be sent straight to model2.Course.SyncUsers() without any modifications.
func readUsersTSV(path string, skipRows int) (map[string]*usr.User, error) {
    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open user TSV file '%s': '%w'.", path, err);
    }
    defer file.Close();

    newUsers := make(map[string]*usr.User);

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
        var role usr.UserRole = usr.Student;

        if (len(parts) >= 1) {
            email = parts[0];
        } else {
            return nil, fmt.Errorf("User file '%s' line %d does not have enough fields.", path, lineno);
        }

        if (len(parts) >= 2) {
            name = parts[1];
        }

        if (len(parts) >= 3) {
            role = usr.GetRole(parts[2]);
            if (role == usr.Unknown) {
                return nil, fmt.Errorf("User file '%s' line %d has unknwon role '%s'.", path, lineno, parts[2]);
            }
        }

        newUser := usr.NewUser(email, name, role);

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
