package disk

import (
    "fmt"
    "path/filepath"
    "sync"

    _ "github.com/mattn/go-sqlite3"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

type backend struct {
    baseDir string
    lock sync.Mutex;
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

    util.MkDir(filepath.Dir(baseDir));

    return &backend{baseDir: baseDir}, nil;
}

func (this *backend) Close() error {
    return nil;
}

func (this *backend) EnsureTables() error {
    return nil;
}

// TEST
func (this *backend) GetCourseUsers(courseID string) (map[string]*usr.User, error) {
    this.lock.Lock();
    defer this.lock.Unlock();

    // TEST
    return nil, nil;
}
