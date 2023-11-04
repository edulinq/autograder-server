package sqlite

import (
    "database/sql"
    _ "embed"
    "fmt"
    "path/filepath"
    "sync"

    _ "github.com/mattn/go-sqlite3"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

//go:embed sql/create.sql
var SQL_CREATE_TABLES string;

type backend struct {
    db *sql.DB
    lock sync.Mutex;
    path string
}

func Open() (*backend, error) {
    path := config.DB_SQLITE_PATH.Get();
    if (path == "") {
        return nil, fmt.Errorf("SQLite path cannot be empty.");
    }

    if (!filepath.IsAbs(path)) {
        path = filepath.Join(config.WORK_DIR.Get(), path);
    }

    util.MkDir(filepath.Dir(path));

	db, err := sql.Open("sqlite3", path);
	if (err != nil) {
        return nil, fmt.Errorf("Failed to open SQLite3 database at '%s': %w.", path, err);
	}

    return &backend{db: db, path: path}, nil;
}

func (this *backend) Close() error {
    this.lock.Lock();
    defer this.lock.Unlock();

    return this.db.Close()
}

func (this *backend) EnsureTables() error {
    this.lock.Lock();
    defer this.lock.Unlock();

    var count int;
	err := this.db.QueryRow("SELECT COUNT(*) FROM sqlite_master").Scan(&count);
    if (err != nil) {
        return err;
    }

    if (count > 0) {
        // Tables exist, assume the database is fine.
        return nil;
    }

    // No tables exist, create them.
	_, err = this.db.Exec(SQL_CREATE_TABLES);
	if (err != nil) {
		return fmt.Errorf("Could not create tables: %w.", err);
	}

    return nil;
}

// TEST
func (this *backend) GetCourseUsers(courseID string) (map[string]*usr.User, error) {
    this.lock.Lock();
    defer this.lock.Unlock();

    // TEST
    return nil, nil;
}
