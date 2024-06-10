package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms/lmssync"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type AddTSV struct {
	TSV       string `help:"Path to the TSV file containing the new users." arg:"" required:""`
	SkipRows  int    `help:"Number of initial rows to skip." default:"0"`
	Force     bool   `help:"Overwrite any existing users." short:"f" default:"false"`
	SendEmail bool   `help:"Send an email to the user about adding them. Errors sending emails will be noted, but will not halt operations." default:"false"`
	DryRun    bool   `help:"Do not actually write out the user's file, just state what you would do." default:"false"`
	SyncLMS   bool   `help:"After adding users, sync the course users (all of them) with the course's LMS." default:"false"`
}

func (this *AddTSV) Run(course *model.Course) error {
	newUsers, err := readUsersTSV(this.TSV, this.SkipRows)
	if err != nil {
		return err
	}

	result, err := db.SyncUsers(course, newUsers, this.Force, this.DryRun, this.SendEmail)
	if err != nil {
		return err
	}

	if this.DryRun {
		fmt.Println("Doing a dry run, users file will not be written to.")
		fmt.Println(util.MustToJSONIndent(result))
	}

	if this.SyncLMS {
		emails := make([]string, 0, len(newUsers))
		for _, newUser := range newUsers {
			emails = append(emails, newUser.Email)
		}

		result, err = lmssync.SyncLMSUserEmails(course, emails, this.DryRun, this.SendEmail)
		if err != nil {
			return err
		}

		if this.DryRun {
			fmt.Println("LMS sync report:")
			fmt.Println(util.MustToJSONIndent(result))
		}
	}

	return nil
}

type ChangePassword struct {
	Email     string `help:"Email for the user." arg:"" required:""`
	Pass      string `help:"Password for the user. Defaults to a random string (will be output)." short:"p"`
	SendEmail bool   `help:"Send an email to the user." default:"false"`
}

func (this *ChangePassword) Run(course *model.Course) error {
	user, err := db.GetCourseUser(course, this.Email)
	if err != nil {
		return fmt.Errorf("Failed to get user: '%w'.", err)
	}

	if user == nil {
		return fmt.Errorf("User '%s' does not exist.", this.Email)
	}

	user.Pass = this.Pass

	result, err := db.SyncUser(course, user, true, false, this.SendEmail)
	if err != nil {
		return fmt.Errorf("Failed to sync user: '%w'.", err)
	}

	// Wait to the very end to output the generated password.
	if len(result.ClearTextPasswords) > 0 {
		fmt.Printf("Generated password: '%s'.\n", result.ClearTextPasswords[user.Email])
	}

	return nil
}

type RmUser struct {
	Email string `help:"Email for the user to be removed." arg:"" required:""`
}

func (this *RmUser) Run(course *model.Course) error {
	exists, err := db.RemoveUser(course, this.Email)
	if err != nil {
		return fmt.Errorf("Failed to remove user '%s': '%w'.", this.Email, err)
	}

	if !exists {
		return fmt.Errorf("User does not exist '%s'.", this.Email)
	}

	fmt.Printf("User '%s' removed.\n", this.Email)

	return nil
}

var cli struct {
	config.ConfigArgs
	Course string `help:"ID of the course."`

	AddTSV AddTSV `cmd:"" help:"Add users from a TSV file formatted as: '<email>[\t<name>[\t<role>[\t<password>]]]'. See add for default values."`
	Rm     RmUser `cmd:"" help:"Remove a user."`
}

func main() {
	context := kong.Parse(&cli,
		kong.Description("Manage users."),
	)

	err := config.HandleConfigArgs(cli.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	course := db.MustGetCourse(cli.Course)

	err = context.Run(course)
	if err != nil {
		log.Fatal("Failed to run command.", err, course)
	}
}

type TSVUser struct {
	User          *model.User
	UserExists    bool
	GeneratedPass bool
	CleartextPass string
}

// Read users from a TSV formatted as: '<email>[\t<name>[\t<role>[\t<cleartext password>]]]'.
// The users returned from this function are not official users yet.
// The users returned from here can be sent straight to db.SyncUsers() without any modifications.
func readUsersTSV(path string, skipRows int) (map[string]*model.User, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to open user TSV file '%s': '%w'.", path, err)
	}
	defer file.Close()

	newUsers := make(map[string]*model.User)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	lineno := 0
	for scanner.Scan() {
		lineno++
		if skipRows > 0 {
			skipRows--
			continue
		}

		parts := strings.Split(scanner.Text(), "\t")

		var email string
		var name string = ""
		var role model.CourseUserRole = model.RoleStudent

		if len(parts) >= 1 {
			email = parts[0]
		} else {
			return nil, fmt.Errorf("User file '%s' line %d does not have enough fields.", path, lineno)
		}

		if len(parts) >= 2 {
			name = parts[1]
		}

		if len(parts) >= 3 {
			role = model.GetCourseUserRole(parts[2])
			if role == model.RoleUnknown {
				return nil, fmt.Errorf("User file '%s' line %d has unknwon role '%s'.", path, lineno, parts[2])
			}
		}

		newUser := model.NewUser(email, name, role)

		if len(parts) >= 4 {
			hashPass := util.Sha256HexFromString(parts[3])
			newUser.Pass = hashPass
		}

		if len(parts) >= 5 {
			return nil, fmt.Errorf("User file '%s' line %d contains too many fields. Found %d, expecting at most %d.", path, lineno, len(parts), 4)
		}

		newUsers[email] = newUser
	}

	return newUsers, nil
}
