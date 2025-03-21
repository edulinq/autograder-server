package core

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

// API Description will be empty until RunServer() is called.
var apiDescription APIDescription

type APIDescription struct {
	Endpoints map[string]EndpointDescription `json:"endpoints"`
}

type EndpointDescription struct {
	RequestType  string            `json:"request-type"`
	ResponseType string            `json:"response-type"`
	Description  string            `json:"description"`
	InputFields  map[string]string `json:"input-fields"`
	OutputFields map[string]string `json:"output-fields"`
}

type TypeDescription map[string]string

func SetAPIDescription(description APIDescription) {
	apiDescription = description
}

func GetAPIDescription() APIDescription {
	return apiDescription
}

func GetDescriptionFromHandler(basePath string) (string, error) {
	absPath := getHandlerFilePath(basePath)
	if !util.IsFile(absPath) {
		return "", fmt.Errorf("Unable to find file path to API Handler. Endpoint: '%s'. Expected path: '%s'.", basePath, absPath)
	}

	handlePattern := regexp.MustCompile(`Handle`)
	description, err := util.GetDescriptionFromFunction(absPath, handlePattern)
	if err != nil {
		return "", fmt.Errorf("Error while getting description from function: '%v'. Endpoint: '%s'.", err, basePath)
	}

	return description, nil
}

func getHandlerFilePath(basePath string) string {
	if strings.HasPrefix(basePath, "/") {
		basePath = strings.TrimPrefix(basePath, "/")
	}

	return util.ShouldAbs(filepath.Join(util.ShouldGetThisDir(), "..", basePath)) + ".go"
}
