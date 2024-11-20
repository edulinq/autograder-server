package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/cmd"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
	cmd.CommonOptions

	Endpoint   string   `help:"Endpoint of the desired action." arg:""`
	Parameters []string `help:"Parameters for the endpoint" arg:"" optional:""`
}

type GeneralCMD struct {
	UserEmail        string   `json:"user-email"`
	UserPass         string   `json:"user-pass"`
	RootUserNonce    string   `json:"root-user-nonce"`
	CourseId         string   `json:"course-id"`
	AssignmentId     string   `json:"assignment-id"`
	Assignment       *string   `json:"Assignment"`
	MinCourseRoleStudent bool `json:"MinCourseRoleStudent"`
	TargetEmail      string   `json:"target-email"`
	TargetSubmission string   `json:"target-submission"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Perform an action with the desired endpoint."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Failed to load config options.", err)
	}

	var description core.EndpointDescription 

	describe := api.Describe(*api.GetRoutes())
	for endpoint, requestResponse := range describe.Endpoints {
		if endpoint == args.Endpoint {
			description = requestResponse
			break
		}
	}

	if description == (core.EndpointDescription{}) {
		log.Error("Failed to find the endpoint.", log.NewAttr("endpoint", args.Endpoint))
		fmt.Print(util.MustToJSONIndent(api.Describe(*api.GetRoutes())))
		return
	}
	// reqType := reflect.TypeOf(description.RequestType)
	resType := description.ResponseType

	// reqInstance := reflect.New(reqType.Elem()).Interface()



	// fmt.Println("req: ", reqInstance)
	fmt.Println("res: ", resType)



	// request := GeneralCMD {
	// 	TargetEmail: "course-student@test.edulinq.org",
	// 	CourseId: "course101",
	// 	AssignmentId: "hw0",
	// }

	// jsonData := util.MustToJSON(request)
	// fmt.Println("jsonData: ", jsonData)

	// json := util.MustToJSON(request)
	// fmt.Println(json)

	// fmt.Println("parameters: ", args.Parameters)

	// cmd.MustHandleCMDRequestAndExitFull(args.Endpoint, reqType, resType, args.CommonOptions, nil)
}