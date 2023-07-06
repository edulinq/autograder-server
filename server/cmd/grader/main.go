package main

import (
    "github.com/eriq-augustine/autograder/grader"
)

func main() {
    // TEST
    config := grader.AssignmentConfig{
        ID: "HO0",
        DisplayName: "Hands-On 0",
        Image: grader.DockerImageConfig{
            ParentName: "autograder/python",
        },
    };

    imageName, err := grader.BuildAssignmentImage("cse140-s23", config);
    if (err != nil) {
        // TODO(eriq): Log
        panic(err);
    }

    grader.RunContainerGrader(imageName);
}
