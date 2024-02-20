package main

import (
    "fmt"

    "github.com/alecthomas/kong"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/docker"
    "github.com/edulinq/autograder/model"
)

var args struct {
    config.ConfigArgs
    docker.BuildOptions
    Course string `help:"ID of the course." arg:"" optional:""`
    Assignment string `help:"ID of the assignment." arg:"" optional:""`
    Force bool `help:"Force images build commands to be sent to docker even if the image is up-to-date." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Build images from all known assignments, or from the specified assignment."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal("Could not load config options.", err);
    }

    db.MustOpen();
    defer db.MustClose();

    var assignments []*model.Assignment;

    if (args.Assignment != "") {
        assignments = append(assignments, db.MustGetAssignment(args.Course, args.Assignment));
    } else {
        for _, course := range db.MustGetCourses() {
            for _, assignment := range course.GetAssignments() {
                assignments = append(assignments, assignment);
            }
        }
    }

    imageNames := buildImages(assignments);

    fmt.Printf("Successfully built %d images:\n", len(imageNames));
    for _, imageName := range imageNames {
        fmt.Printf("    %s\n", imageName);
    }
}

func buildImages(assignments []*model.Assignment) []string {
    imageNames := make([]string, 0);

    for _, assignment := range assignments {
        err := docker.BuildImageFromSource(assignment, args.Force, false, &args.BuildOptions);
        if (err != nil) {
            log.Fatal("Failed to build image.", assignment, err);
        }

        imageNames = append(imageNames, assignment.ImageName());
    }

    return imageNames;
}
