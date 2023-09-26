package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const DEFAULT_PASSWORD_LEN = 32;
const EMAIL_SLEEP_TIME = int64(0.5 * float64(time.Second));

type AddUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Name string `help:"Display name for the user. Defaults to the user's email." short:"n"`
    Role string `help:"Role for the user. Defaults to student." short:"r" default:"student"`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
    Force bool `help:"Overwrite any existing user." short:"f" default:"false"`
    SendEmail bool `help:"Send an email to the user about adding them. Errors sending emails will be noted, but will not halt operations." default:"false"`
    DryRun bool `help:"Do not actually write out the user's file or send emails, just state what you would do." default:"false"`
}

func (this *AddUser) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    generatedPass := false;
    if (this.Pass == "") {
        this.Pass, err = util.RandHex(DEFAULT_PASSWORD_LEN);
        if (err != nil) {
            return fmt.Errorf("Failed to generate a default password.");
        }

        generatedPass = true;
    }

    user, userExists, err := newOrMergeUser(users, this.Email, this.Name, this.Role, this.Pass, this.Force);
    if (err != nil) {
        return err;
    }

    users[user.Email] = user;

    if (this.DryRun) {
        fmt.Printf("Doing a dry run, users file '%s' will not be written to.\n", path);
    } else {
        err = model.SaveUsersFile(path, users);
        if (err != nil) {
            return fmt.Errorf("Failed to save users file '%s': '%w'.", path, err);
        }
    }

    // Wait until the users file has been written to output the generated password.
    if (generatedPass) {
        fmt.Printf("Generated password for '%s': '%s'.\n", this.Email, this.Pass);
    }

    if (this.SendEmail) {
        sendUserAddEmail(user, this.Pass, generatedPass, userExists, this.DryRun, false);
    }

    return nil;
}

type AddTSV struct {
    TSV string `help:"Path to the TSV file containing the new users." arg:"" required:""`
    SkipRows int `help:"Number of initial rows to skip." default:"0"`
    Force bool `help:"Overwrite any existing users." short:"f" default:"false"`
    SendEmail bool `help:"Send an email to the user about adding them. Errors sending emails will be noted, but will not halt operations." default:"false"`
    DryRun bool `help:"Do not actually write out the user's file, just state what you would do." default:"false"`
}

func (this *AddTSV) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    newUsers, err := readUsersTSV(users, this.TSV, this.SkipRows, this.Force);
    if (err != nil) {
        return err;
    }

    if (!this.Force) {
        // Make a pass and ensure that no new users are dups.
        dupCount := 0;
        for _, newUser := range newUsers {
            user := users[newUser.User.Email];
            if (user != nil) {
                fmt.Printf("Found a duplicate user: '%s'.\n", user.Email);
                dupCount++;
            }
        }

        if (dupCount > 0) {
            return fmt.Errorf("Found %d dupliate users.", dupCount);
        }
    }

    // Set all the new users.
    for _, newUser := range newUsers {
        users[newUser.User.Email] = newUser.User;
    }

    if (this.DryRun) {
        fmt.Printf("Doing a dry run, users file '%s' will not be written to.\n", path);
    } else {
        err = model.SaveUsersFile(path, users);
        if (err != nil) {
            return fmt.Errorf("Failed to save users file '%s': '%w'.", path, err);
        }
    }

    // Wait until the users file has been written to output the generated passwords.
    for _, newUser := range newUsers {
        if (!newUser.GeneratedPass) {
            continue;
        }

        fmt.Printf("Generated password for '%s': '%s'.\n", newUser.User.Email, newUser.CleartextPass);
    }

    if (this.SendEmail) {
        fmt.Println("Sending out registration emails.");
        for _, newUser := range newUsers {
            sendUserAddEmail(newUser.User, newUser.CleartextPass, newUser.GeneratedPass, newUser.UserExists, this.DryRun, true);
        }
    }

    return nil;
}

type AuthUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
}

func (this *AuthUser) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    user := users[this.Email];
    if (user == nil) {
        return fmt.Errorf("User '%s' does not exists, cannot auth.", this.Email);
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

func (this *GetUser) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    user := users[this.Email];
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

func (this *ListUsers) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
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

func (this *ChangePassword) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    user := users[this.Email];
    if (user == nil) {
        return fmt.Errorf("No user found with email '%s'.\n", this.Email);
    }

    generatedPass := false;
    if (this.Pass == "") {
        this.Pass, err = util.RandHex(DEFAULT_PASSWORD_LEN);
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

    err = model.SaveUsersFile(path, users);
    if (err != nil) {
        return fmt.Errorf("Failed to save users file '%s': '%w'.", path, err);
    }

    // Wait to the very end to output the generated password.
    if (generatedPass) {
        fmt.Printf("Generated password: '%s'.\n", this.Pass);
    }

    return nil
}

var cli struct {
    config.ConfigArgs
    UsersPath string `help:"Optional path to a users JSON file (or where one will be created)." type:"path" default:"users.json"`

    Add AddUser `cmd:"" help:"Add a user."`
    AddTSV AddTSV `cmd:"" help:"Add users from a TSV file formatted as: '<email>[\t<display name>[\t<role>[\t<password>]]]'. See add for default values."`
    Auth AuthUser `cmd:"" help:"Authenticate as a user."`
    Get GetUser `cmd:"" help:"Get a user."`
    Ls ListUsers `cmd:"" help:"List users."`
    Pass ChangePassword `cmd:"" help:"Change a user's password."`
}

func main() {
    context := kong.Parse(&cli,
        kong.Description("Manage users."),
    );

    err := config.HandleConfigArgs(cli.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    err = context.Run(cli.UsersPath);
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

// Read users from a TSV formatted as: '<email>[\t<display name>[\t<role>[\t<password>]]]'.
// The users returned from this function are not official users yet.
// Their cleaartext password (not hash) will be stored in Salt, and it is up to the caller
// to decide what to do with them (and to set their password).
func readUsersTSV(users map[string]*model.User, path string, skipRows int, force bool) ([]*TSVUser, error) {
    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open user TSV file '%s': '%w'.", path, err);
    }
    defer file.Close();

    newUsers := make([]*TSVUser, 0);

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
        var role string = model.GetRoleString(model.Student);
        var pass string = "";
        var generatedPass bool = false;

        if (len(parts) >= 1) {
            email = parts[0];
        } else {
            return nil, fmt.Errorf("User file '%s' line %d does not have enough fields.", path, lineno);
        }

        if (len(parts) >= 2) {
            name = parts[1];
        }

        if (len(parts) >= 3) {
            role = parts[2];
        }

        if (len(parts) >= 4) {
            pass = parts[3];
            generatedPass = false;
        } else {
            pass, err = util.RandHex(DEFAULT_PASSWORD_LEN);
            if (err != nil) {
                return nil, fmt.Errorf("Failed to generate a default password.");
            }

            generatedPass = true;
        }

        if (len(parts) >= 5) {
            return nil, fmt.Errorf("User file '%s' line %d contains too many fields. Found %d, expecting at most %d.", path, lineno, len(parts), 4);
        }

        user, userExists, err := newOrMergeUser(users, email, name, role, pass, force);
        if (err != nil) {
            return nil, err;
        }

        newUser := &TSVUser{
            User: user,
            UserExists: userExists,
            GeneratedPass: generatedPass,
            CleartextPass: pass,
        };

        newUsers = append(newUsers, newUser);
    }

    return newUsers, nil;
}

func composeUserAddEmail(address string, pass string, generatedPass bool, userExists bool) (string, string) {
    var subject string;
    var body string;

    if (userExists) {
        subject = "Autograder -- User Password Changed";
        body = fmt.Sprintf("Hello,\n\nThe password for '%s' has been changed.\n", address);
    } else {
        subject = "Autograder -- User Account Created";
        body = fmt.Sprintf("Hello,\n\nAn account with the username/email '%s' has been created.\n", address);
    }

    if (generatedPass) {
        body += fmt.Sprintf("The new password is '%s' (no quotes).\n", pass);
    }

    return subject, body;
}

// Return a user that is either new or a merged with the existing user (depending on force).
// If a user exists (and force is true), then the user will be updated.
// New users will not be added to |users|.
func newOrMergeUser(users map[string]*model.User, email string, name string, stringRole string, pass string, force bool) (*model.User, bool, error) {
    user := users[email];
    userExists := (user != nil);

    if (userExists && !force) {
        return nil, false, fmt.Errorf("User '%s' already exists, cannot add.", email);
    }

    if (!userExists) {
        user = &model.User{Email: email};
    }

    if (name != "") {
        user.DisplayName = name;
    } else  if (user.DisplayName == "") {
        user.DisplayName = email;
    }

    // Note the slightly tricky conditions here.
    // Only error if the string role is bad and there is not an existing good role.
    role := model.GetRole(stringRole);
    if (role != model.Unknown) {
        user.Role = role;
    } else if (user.Role == model.Unknown) {
        return nil, false, fmt.Errorf("Unknown role: '%s'.", stringRole);
    }

    hashPass := util.Sha256Hex([]byte(pass));

    err := user.SetPassword(hashPass);
    if (err != nil) {
        return nil, false, fmt.Errorf("Could not set password: '%w'.", err);
    }

    return user, userExists, nil;
}

func sendUserAddEmail(user *model.User, pass string, generatedPass bool, userExists bool, dryRun bool, sleep bool) {
    subject, body := composeUserAddEmail(user.Email, pass, generatedPass, userExists);

    if (dryRun) {
        fmt.Printf("Doing a dry run, user '%s' will not be emailed.\n", user.Email);
        log.Debug().Str("address", user.Email).Str("subject", subject).Str("body", body).Msg("Email not sent because of dry run.");
    } else {
        err := email.Send([]string{user.Email}, subject, body);
        if (err != nil) {
            log.Error().Err(err).Str("email", user.Email).Msg("Failed to send email.");
        }
        fmt.Printf("Registration email send to '%s'.\n", user.Email);

        if (sleep) {
            time.Sleep(time.Duration(EMAIL_SLEEP_TIME));
        }
    }
}
