package config

import (
	"fmt"
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	WORK_DIR_BASENAME = "autograder"

	BACKUP_DIRNAME    = "backup"
	CACHE_DIRNAME     = "cache"
	CONFIG_DIRNAME    = "config"
	DATABASE_DIRNAME  = "database"
	LOGS_DIRNAME      = "logs"
	SOURCES_DIRNAME   = "sources"
	TEMPLATES_DIRNAME = "templates"

	TESTDATA_DIRNAME = "testdata"
)

func GetDefaultBaseDir() string {
	return xdg.DataHome
}

func GetWorkDir() string {
	dirname := WORK_DIR_BASENAME

	serverName := NAME.Get()
	if (serverName != "") && (dirname != serverName) {
		dirname = fmt.Sprintf("%s-%s", dirname, serverName)
	}

	return filepath.Join(BASE_DIR.Get(), dirname)
}

func GetBaseBackupDir() string {
	return filepath.Join(GetWorkDir(), BACKUP_DIRNAME)
}

// Get the backup directory, which will check BACKUP_DIR first,
// and then return GetBaseBackupDir() if the option is empty.
func GetBackupDir() string {
	dir := BACKUP_DIR.Get()
	if dir != "" {
		return dir
	}

	return GetBaseBackupDir()
}

func GetCacheDir() string {
	return filepath.Join(GetWorkDir(), CACHE_DIRNAME)
}

func GetConfigDir() string {
	return filepath.Join(GetWorkDir(), CONFIG_DIRNAME)
}

func GetTestdataDir() string {
	return filepath.Join(GetWorkDir(), TESTDATA_DIRNAME)
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

func GetTemplatesDir() string {
	return filepath.Join(GetWorkDir(), TEMPLATES_DIRNAME)
}
