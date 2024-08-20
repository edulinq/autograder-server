package config

import (
	"fmt"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	WORK_DIR_BASENAME = "autograder"

	PID_FILENAME         = "autograder.pid"
	UNIX_SOCKET_FILENAME = "autograder.sock"

	BACKUP_DIRNAME        = "backup"
	CACHE_DIRNAME         = "cache"
	CONFIG_DIRNAME        = "config"
	COURSE_IMPORT_DIRNAME = "course_import"
	DATABASE_DIRNAME      = "database"
	LOGS_DIRNAME          = "logs"
	SOURCES_DIRNAME       = "sources"
)

func GetDefaultBaseDir() string {
	return xdg.DataHome
}

func GetWorkDir() string {
	dirname := WORK_DIR_BASENAME

	serverName := NAME.Get()
	if serverName != "" {
		dirname = fmt.Sprintf("%s-%s", dirname, serverName)
	}

	return filepath.Join(BASE_DIR.Get(), serverName)
}

func GetBackupDir() string {
	return filepath.Join(GetWorkDir(), BACKUP_DIRNAME)
}

// Get the backup directory for a task, which will check TASK_BACKUP_DIR first,
// and then return GetBackupDir() if the option is empty.
func GetTaskBackupDir() string {
	dir := TASK_BACKUP_DIR.Get()
	if dir != "" {
		return dir
	}

	return GetBackupDir()
}

func GetCacheDir() string {
	return filepath.Join(GetWorkDir(), CACHE_DIRNAME)
}

func GetConfigDir() string {
	return filepath.Join(GetWorkDir(), CONFIG_DIRNAME)
}

func GetCourseImportDir() string {
	return filepath.Join(GetWorkDir(), COURSE_IMPORT_DIRNAME)
}

func GetDatabaseDir() string {
	return filepath.Join(GetWorkDir(), DATABASE_DIRNAME)
}

func GetLogsDir() string {
	return filepath.Join(GetWorkDir(), LOGS_DIRNAME)
}

func GetSourcesDir() string {
	return filepath.Join(GetWorkDir(), SOURCES_DIRNAME)
}

func GetUnixSocketDir() string {
	unixSocketDir := filepath.Join(GetWorkDir(), "tmp")

	err := util.MkDir(unixSocketDir)
	if err != nil {
		log.Error("Error creating Unix Socket directory.", err)
	}

	return filepath.Join(unixSocketDir, UNIX_SOCKET_FILENAME)
}

func GetPidDir() string {
	pidDir := filepath.Join(GetWorkDir(), "tmp")

	err := util.MkDir(pidDir)
	if err != nil {
		log.Error("Error creating PID directory.", err)
	}

	return filepath.Join(pidDir, PID_FILENAME)
}
