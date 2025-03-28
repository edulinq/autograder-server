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
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const DB_DIRNAME = "disk-database"

type backend struct {
	baseDir string

	// A general lock for contextual locking.
	contextualLock sync.RWMutex

	// Specific locks.
	logLock                sync.RWMutex
	userLock               sync.RWMutex
	statsLock              sync.RWMutex
	systemStatsLock        sync.RWMutex
	tasksLock              sync.RWMutex
	analysisIndividualLock sync.RWMutex
	analysisPairwiseLock   sync.RWMutex
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
// When procedures require a contextual lock, we will also get a read lock on contextualLock.
// This allows us to obtain a write lock on contextualLock when we need to do database-wide operations,
// which will block out contextual lockers.
func (this *backend) contextLock(id string) {
	this.contextualLock.RLock()
	lockmanager.Lock(id)
}

func (this *backend) contextReadLock(id string) {
	this.contextualLock.RLock()
	lockmanager.ReadLock(id)
}

func (this *backend) contextUnlock(id string) {
	this.contextualLock.RUnlock()
	lockmanager.Unlock(id)
}

func (this *backend) contextReadUnlock(id string) {
	this.contextualLock.RUnlock()
	lockmanager.ReadUnlock(id)
}

func (this *backend) Clear() error {
	// Before clearing the DB, obtain all locks.

	this.contextualLock.Lock()
	defer this.contextualLock.Unlock()

	this.logLock.Lock()
	defer this.logLock.Unlock()

	this.userLock.Lock()
	defer this.userLock.Unlock()

	this.systemStatsLock.Lock()
	defer this.systemStatsLock.Unlock()

	this.statsLock.Lock()
	defer this.statsLock.Unlock()

	this.tasksLock.Lock()
	defer this.tasksLock.Unlock()

	this.analysisIndividualLock.Lock()
	defer this.analysisIndividualLock.Unlock()

	this.analysisPairwiseLock.Lock()
	defer this.analysisPairwiseLock.Unlock()

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
