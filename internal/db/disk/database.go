// A database backend that just exists on disk without any external tools,
// the data just exists in flat files.
// Meant mostly for testing and small deployments.
// This database will lock when writing.
package disk

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const DB_DIRNAME = "disk-database"

type backend struct {
	baseDir string
	lock    sync.RWMutex
	logLock sync.RWMutex
}

func Open() (*backend, error) {
	baseDir := util.ShouldAbs(filepath.Join(config.GetDatabaseDir(), DB_DIRNAME))

	err := util.MkDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to make db dir '%s': '%w'.", baseDir, err)
	}

	log.Debug("Opened disk database.", log.NewAttr("base-dir", baseDir))

	return &backend{baseDir: baseDir}, nil
}

func (this *backend) Close() error {
	return nil
}

func (this *backend) EnsureTables() error {
	return nil
}

func (this *backend) Clear() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.logLock.Lock()
	defer this.logLock.Unlock()

	err := util.RemoveDirent(this.baseDir)
	if err != nil {
		return err
	}

	err = util.MkDir(this.baseDir)
	if err != nil {
		return fmt.Errorf("Failed to make db dir '%s': '%w'.", this.baseDir, err)
	}

	return nil
}
