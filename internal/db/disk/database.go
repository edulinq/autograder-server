// A database backend that just exists on disk without any external tools,
// the data just exists in flat files.
// Meant mostly for testing and small deployments.
// This database will lock when writing.
package disk

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const DB_DIRNAME = "disk-database"

type backend struct {
	baseDir string

	// A general lock for contextual locking.
	contextualLock sync.RWMutex

	// Specific locks.
	logLock   sync.RWMutex
	userLock  sync.RWMutex
	statsLock sync.RWMutex
	tasksLock sync.RWMutex
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

// Acquire a contextual lock.
// For the DB, which also involves obtaining a read lock on a general lock,
// this will allow us to obtain all the locks when clearning the db.
func (this *backend) contextLock(id string) {
	this.contextualLock.RLock()
	common.Lock(id)
}

func (this *backend) contextReadLock(id string) {
	this.contextualLock.RLock()
	common.ReadLock(id)
}

func (this *backend) contextUnlock(id string) {
	this.contextualLock.RUnlock()
	common.Unlock(id)
}

func (this *backend) contextReadUnlock(id string) {
	this.contextualLock.RUnlock()
	common.ReadUnlock(id)
}

func (this *backend) Clear() error {
	this.contextualLock.Lock()
	defer this.contextualLock.Unlock()

	this.logLock.Lock()
	defer this.logLock.Unlock()

	this.userLock.Lock()
	defer this.userLock.Unlock()

	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	this.tasksLock.Lock()
	defer this.tasksLock.Unlock()

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
