package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/lms"
    "github.com/edulinq/autograder/log"
)

var args struct {
    config.ConfigArgs
    Course string `help:"ID of the course." arg:""`
}

func main() {
    kong.Parse(&args,
        kong.Description("Fetch users for a specific LMS course."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    course := db.MustGetCourse(args.Course);

    users, err := lms.FetchUsers(course);
    if (err != nil) {
        log.Fatal("Could not fetch users.", err, course);
    }

    fmt.Println("id\temail\tname\trole");
    for _, user := range users {
        fmt.Printf("%s\t%s\t%s\t%s\n", user.ID, user.Email, user.Name, user.Role.String());
    }
}
