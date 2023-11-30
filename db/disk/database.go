// A database backend that just exists on disk without any external tools,
// the data just exists in flat files.
// Meant mostly for testing and small deployments.
// This database will lock when writing.
package disk

import (
    "fmt"
    "path/filepath"
    "sync"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

type backend struct {
    baseDir string
    lock sync.RWMutex;
}

func Open() (*backend, error) {
    baseDir := config.DB_DISK_DIR.Get();
    if (baseDir == "") {
        return nil, fmt.Errorf("Disk database dir cannot be empty.");
    }

    if (!filepath.IsAbs(baseDir)) {
        baseDir = filepath.Join(config.WORK_DIR.Get(), baseDir);
    }

    baseDir = util.ShouldAbs(baseDir);

    err := util.MkDir(baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to make db dir '%s': '%w'.", baseDir, err);
    }

    log.Debug().Str("base-dir", baseDir).Msg("Opened disk database.");

    return &backend{baseDir: baseDir}, nil;
}

func (this *backend) Close() error {
    return nil;
}

func (this *backend) EnsureTables() error {
    return nil;
}

func (this *backend) Clear() error {
    err := util.RemoveDirent(this.baseDir);
    if (err != nil) {
        return err;
    }

    err = util.MkDir(this.baseDir);
    if (err != nil) {
        return fmt.Errorf("Failed to make db dir '%s': '%w'.", this.baseDir, err);
    }

    return nil;
}
