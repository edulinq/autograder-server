package dolos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const (
	NAME         = "dolos"
	VERSION      = "2.9.0"
	DOCKER_IMAGE = "ghcr.io/dodona-edu/dolos-cli"

	MAX_RUNTIME_SECS = 2 * 60

	OUT_DIRNAME  = "out"
	OUT_FILENAME = "pairs.csv"
)

var (
	hasImage  bool = false
	imageLock sync.Mutex
)

type dolosEngine struct{}

func GetEngine() *dolosEngine {
	return &dolosEngine{}
}

func (this *dolosEngine) GetName() string {
	return NAME
}

func (this *dolosEngine) IsAvailable() bool {
	return docker.CanAccessDocker()
}

func (this *dolosEngine) ComputeFileSimilarity(paths [2]string, baseLockKey string) (*model.FileSimilarity, error) {
	lockKey := fmt.Sprintf("dolos-%s", baseLockKey)
	common.Lock(lockKey)
	defer common.Unlock(lockKey)

	err := ensureImage()
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure Dolos docker image exists: '%w'.", err)
	}

	tempDir, err := util.MkDirTemp("dolos-")
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	tempFilenames := make([]string, 0, 2)

	for i, path := range paths {
		tempFilename := fmt.Sprintf("%d%s", i, filepath.Ext(path))
		tempPath := filepath.Join(tempDir, tempFilename)
		err = util.CopyFile(path, tempPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to copy file to temp dir: '%w'.", err)
		}

		err = os.Chmod(tempPath, 0664)
		if err != nil {
			return nil, fmt.Errorf("Failed to update permissions for temp file: '%w'.", err)
		}

		err = os.Chown(tempPath, 1000, 1000)
		if err != nil {
			return nil, fmt.Errorf("Failed to update ownership for temp file: '%w'.", err)
		}

		tempFilenames = append(tempFilenames, tempFilename)
	}

	mounts := []docker.MountInfo{
		docker.MountInfo{
			Source:   tempDir,
			Target:   "/dolos",
			ReadOnly: false,
		},
	}
	arguments := []string{
		"run",
		"--output-format", "csv",
		tempFilenames[0],
		tempFilenames[1],
		"--output-destination", OUT_DIRNAME,
	}

	stdout, stderr, _, _, err := docker.RunContainer(context.Background(), this, getImageName(), mounts, arguments, NAME, MAX_RUNTIME_SECS)
	if err != nil {
		log.Debug("Failed to run Dolos container.", err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr))
		return nil, fmt.Errorf("Failed to run Dolos container: '%w'.", err)
	}

	score, err := fetchResults(tempDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to read output from Dolos: '%w'.", err)
	}

	result := model.FileSimilarity{
		Filename: filepath.Base(paths[0]),
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
		return 0.0, fmt.Errorf("Dolos output file does not exist: '%s'.", path)
	}

	rows, err := util.ReadSeparatedFile(path, ",", 1)
	if err != nil {
		return 0.0, fmt.Errorf("Failed to read Dolos output file: '%w'.", err)
	}

	numRows := len(rows)
	numCols := -1
	if numRows > 0 {
		numCols = len(rows[0])
	}

	if (numRows != 1) || (numCols != 10) {
		return 0.0, fmt.Errorf("Shape of Dolos output is not correct. Expected (1 x 10), found (%d x %d).", numRows, numCols)
	}

	valueString := rows[0][5]
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0.0, fmt.Errorf("Failed to parse Dolos similarity value to a float '%s': '%w'.", valueString, err)
	}

	return value, nil
}

func (this *dolosEngine) LogValue() []*log.Attr {
	return []*log.Attr{
		log.NewAttr("similarity-engine", NAME),
		log.NewAttr("version", VERSION),
	}
}