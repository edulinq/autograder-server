package common

import (
	"os"
	"path/filepath"
)

const (
	API_DESCRIPTION_FILENAME = "api.json"
	RESOURCES_OUTPUT_DIRNAME = "resources"
)

func GetResourcesDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, "../../", RESOURCES_OUTPUT_DIRNAME), nil
}

func GetAPIDescriptionFilepath() (string, error) {
	resourcesDir, err := GetResourcesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(resourcesDir, API_DESCRIPTION_FILENAME), nil
}
