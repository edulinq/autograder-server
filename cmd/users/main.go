package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const DEFAULT_PASSWORD_LEN = 32;

type AddUser struct {
    Email string `help:"Email for the user." arg:"" required:""`
    Name string `help:"Display name for the user. Defaults to the user's email." short:"n"`
    Role string `help:"Role for the user. Defaults to student." short:"r" default:"student"`
    Pass string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
    Force bool `help:"Overwrite any existing user." short:"f" default:"false"`
}

func (this *AddUser) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    user := users[this.Email];
    if (!this.Force && (user != nil)) {
        return fmt.Errorf("User '%s' already exists, cannot add.", this.Email);
    }

    if (this.Name == "") {
        this.Name = this.Email;
    }

    role := model.GetRole(this.Role);
    if (role == model.Unknown) {
        return fmt.Errorf("Unknown role: '%s'.", this.Role);
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

    user = &model.User{
        Email: this.Email,
        DisplayName: this.Name,
        Role: role,
    };

    err = user.SetPassword(pass);
    if (err != nil) {
        return fmt.Errorf("Could not set password: '%w'.", err);
    }

    users[user.Email] = user;

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

type AddTSV struct {
    TSV string `help:"Path to the TSV file containing the new users." arg:"" required:""`
    SkipRows int `help:"Number of initial rows to skip." default:"0"`
    Force bool `help:"Overwrite any existing users." short:"f" default:"false"`
}

func (this *AddTSV) Run(path string) error {
    users, err := model.LoadUsersFile(path);
    if (err != nil) {
        return fmt.Errorf("Failed to load users file '%s': '%w'.", path, err);
    }

    newUsers, err := readUsersTSV(this.TSV, this.SkipRows);
    if (err != nil) {
        return err;
    }

    if (!this.Force) {
        // Make a pass and ensure that no new users are dups.
        dupCount := 0;
        for _, newUser := range newUsers {
            user := users[newUser.Email];
            if (user != nil) {
                fmt.Printf("Found a duplicate user: '%s'.\n", user.Email);
                dupCount++;
            }
        }

        if (dupCount > 0) {
            return fmt.Errorf("Found %d dupliate users.", dupCount);
        }
    }

    generatedPasswords := make([][2]string, 0);

    // Set passwords and add new users to the users map.
    for _, newUser := range newUsers {
        if (newUser.Salt == "") {
            newUser.Salt, err = util.RandHex(DEFAULT_PASSWORD_LEN);
            if (err != nil) {
                return fmt.Errorf("Failed to generate a default password.");
            }

            generatedPasswords = append(generatedPasswords, [2]string{newUser.Email, newUser.Salt});
        }

        err = newUser.SetPassword(util.Sha256Hex([]byte(newUser.Salt)));
        if (err != nil) {
            return fmt.Errorf("Could not set password: '%w'.", err);
        }

        users[newUser.Email] = newUser;
    }

    err = model.SaveUsersFile(path, users);
    if (err != nil) {
        return fmt.Errorf("Failed to save users file '%s': '%w'.", path, err);
    }

    // Wait to the very end to output generated passwords.
    for _, info := range generatedPasswords {
        fmt.Printf("Generated password for '%s': '%s'.\n", info[0], info[1]);
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
    Path string `help:"Option path to output a JSON grading result." type:"path" default:"users.json"`

    Add AddUser `cmd:"" help:"Add a user."`
    AddTSV AddTSV `cmd:"" help:"Add users from a TSV file formatted as: '<email>[\t<display name>[\t<role>[\t<password>]]]'. See add for default values."`
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

    err = context.Run(cli.Path);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to run command.");
    }
}

// Read users from a TSV formatted as: '<email>[\t<display name>[\t<role>[\t<password>]]]'.
// The users returned from this function are not official users yet.
// Their cleaartext password (not hash) will be stored in Salt, and it is up to the caller
// to decide what to do with them (and to set their password).
func readUsersTSV(path string, skipRows int) ([]*model.User, error) {
    users := make([]*model.User, 0);

    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open user TSV file '%s': '%w'.", path, err);
    }
    defer file.Close();

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

        user := &model.User{};

        if (len(parts) >= 1) {
            user.Email = parts[0];
        } else {
            return nil, fmt.Errorf("User file '%s' line %d does not have enough fields.", path, lineno);
        }

        if (len(parts) >= 2) {
            user.DisplayName = parts[1];
        } else {
            user.DisplayName = user.Email;
        }

        if (len(parts) >= 3) {
            user.Role = model.GetRole(parts[2]);
            if (user.Role == model.Unknown) {
                return nil, fmt.Errorf("User file '%s' line %d has an unknown role: '%s'.", path, lineno, parts[2]);
            }
        } else {
            user.Role = model.Student;
        }

        if (len(parts) >= 4) {
            user.Salt = parts[3];
        } else {
            user.Salt = "";
        }

        if (len(parts) >= 5) {
            return nil, fmt.Errorf("User file '%s' line %d contains too many fields. Found %d, expecting at most %d.", path, lineno, len(parts), 4);
        }

        users = append(users, user);
    }

    return users, nil;
}
