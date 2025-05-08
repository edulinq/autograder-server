package jplag

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const (
	NAME         = "jplag"
	VERSION      = "5.1.0.4"
	DOCKER_IMAGE = "ghcr.io/edulinq/jplag-docker"

	MAX_RUNTIME_SECS = 2 * 60

	OUT_DIRNAME       = "out"
	OUT_FILENAME      = "results.csv"
	TEMPLATE_FILENAME = "template"

	DEFAULT_MIN_TOKENS = 12
)

var (
	hasImage  bool = false
	imageLock sync.Mutex
)

type JPlagEngine struct {
	MinTokens int
}

func GetEngine() *JPlagEngine {
	return &JPlagEngine{
		MinTokens: DEFAULT_MIN_TOKENS,
	}
}

func (this *JPlagEngine) GetName() string {
	return NAME
}

func (this *JPlagEngine) IsAvailable() bool {
	return docker.CanAccessDocker()
}

func (this *JPlagEngine) ComputeFileSimilarity(paths [2]string, templatePath string, ctx context.Context) (*model.FileSimilarity, error) {
	err := ensureImage()
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure JPlag docker image exists: '%w'.", err)
	}

	tempDir, err := util.MkDirTemp("jplag-")
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	err = util.MkDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp src dir: '%w'.", err)
	}

	templateDir := filepath.Join(tempDir, "template")
	err = util.MkDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp template dir: '%w'.", err)
	}

	originalFilename := filepath.Base(paths[0])
	language := getLanguage(originalFilename)

	extension := ""
	tempFilenames := make([]string, 0, 2)
	for i, path := range paths {
		extension = filepath.Ext(path)
		if language == DEFAULT_LANGUAGE {
			// JPlag is very strict about file extensions.
			// If we are using the fallback language, change the extension.
			extension = FALLBACK_EXTENSION
		}

		tempFilename := fmt.Sprintf("%d%s", i, extension)
		tempPath := filepath.Join(srcDir, tempFilename)
		err = util.CopyFile(path, tempPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to copy file to temp dir: '%w'.", err)
		}

		tempFilenames = append(tempFilenames, tempFilename)
	}

	templateFilename := fmt.Sprintf("%s%s", TEMPLATE_FILENAME, extension)
	tempTemplatePath := filepath.Join(templateDir, templateFilename)
	if templatePath != "" {
		err = util.CopyFile(templatePath, tempTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to copy template file to temp dir: '%w'.", err)
		}
	}

	// Ensure permissions are very open because UID/GID will not be properly aligned.
	err = util.RecursiveChmod(tempDir, 0666, 0777)
	if err != nil {
		return nil, fmt.Errorf("Failed to set recursive permissions for temp dir: '%w'.", err)
	}

	mounts := []docker.MountInfo{
		docker.MountInfo{
			Source:   tempDir,
			Target:   "/jplag",
			ReadOnly: false,
		},
	}

	arguments := []string{
		"--mode", "RUN",
		"--csv-export",
		"--language", getLanguage(tempFilenames[0]),
		"--min-tokens", fmt.Sprintf("%d", this.MinTokens),
		"/jplag/src",
	}

	if templatePath != "" {
		arguments = append(arguments, "--base-code", "template")
	}

	stdout, stderr, _, _, err := docker.RunContainer(ctx, this, getImageName(), mounts, arguments, NAME, MAX_RUNTIME_SECS)
	if err != nil {
		log.Debug("Failed to run JPlag container.", err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr))
		return nil, fmt.Errorf("Failed to run JPlag container: '%w'.", err)
	}

	score, err := fetchResults(tempDir)
	if err != nil {
		log.Debug("Failed to read output from JPlag.", err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr))
		return nil, fmt.Errorf("Failed to read output from JPlag: '%w'.", err)
	}

	result := model.FileSimilarity{
		Filename: originalFilename,
		Tool:     NAME,
		Version:  VERSION,
		Score:    score,
	}

	return &result, nil
}

func getImageName() string {
	return fmt.Sprintf("%s:%s", DOCKER_IMAGE, VERSION)
}

// Ensure that the correct docker image exists.
func ensureImage() error {
	imageLock.Lock()
	defer imageLock.Unlock()

	return docker.EnsureImage(getImageName())
}

func fetchResults(tempDir string) (float64, error) {
	path := filepath.Join(tempDir, OUT_DIRNAME, OUT_FILENAME)
	if !util.PathExists(path) {
		return 0.0, fmt.Errorf("JPlag output file does not exist: '%s'.", path)
	}

	rows, err := util.ReadSeparatedFile(path, ",", 1)
	if err != nil {
		return 0.0, fmt.Errorf("Failed to read JPlag output file: '%w'.", err)
	}

	numRows := len(rows)
	numCols := -1
	if numRows > 0 {
		numCols = len(rows[0])
	}

	if (numRows != 1) || (numCols != 4) {
		return 0.0, fmt.Errorf("Shape of JPlag output is not correct. Expected (1 x 4), found (%d x %d).", numRows, numCols)
	}

	valueString := rows[0][2]
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0.0, fmt.Errorf("Failed to parse JPlag similarity value to a float '%s': '%w'.", valueString, err)
	}

	return value, nil
}

func (this *JPlagEngine) LogValue() []*log.Attr {
	return []*log.Attr{
		log.NewAttr("similarity-engine", NAME),
		log.NewAttr("version", VERSION),
	}
}
