package common

// Filespecs are string specifications for file-like objects.
// They could be for a plain file/dir (just a path),
// or for something like a git repo.
// Filespecs often appear in assignment configurations.

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/eriq-augustine/autograder/util"
)

type FileSpec string;
type FileSpecType int;

const (
    FILESPEC_TYPE_PATH FileSpecType = iota
    FILESPEC_TYPE_GIT
    FILESPEC_TYPE_NIL
)

const (
    FILESPEC_DELIM = "::"
    FILESPEC_GIT = "git"
    FILESPEC_GIT_PREFIX = FILESPEC_GIT + FILESPEC_DELIM
    FILESPEC_NIL = "nil"
    FILESPEC_NIL_PREFIX = FILESPEC_NIL + FILESPEC_DELIM
)

func (this FileSpec) GetType() FileSpecType {
    if (strings.HasPrefix(string(this), FILESPEC_GIT_PREFIX)) {
        return FILESPEC_TYPE_GIT;
    } else if (strings.HasPrefix(string(this), FILESPEC_NIL_PREFIX)) {
        return FILESPEC_TYPE_NIL;
    } else {
        return FILESPEC_TYPE_PATH;
    }
}

func (this FileSpec) IsEmpty() bool {
    return (string(this) == "");
}

func (this FileSpec) IsPath() bool {
    return this.GetType() == FILESPEC_TYPE_PATH;
}

func (this FileSpec) IsGit() bool {
    return this.GetType() == FILESPEC_TYPE_GIT;
}

func (this FileSpec) IsNil() bool {
    return this.GetType() == FILESPEC_TYPE_NIL;
}

func (this FileSpec) IsAbs() bool {
    return this.IsPath() && filepath.IsAbs(string(this));
}

func (this FileSpec) GetPath() string {
    return string(this);
}

// Copy the target of this FileSpec in the specified location.
// If the filespec is a path, then copy a dirent.
// If the filespec is a git repo, then ensure it is cloned/updated.
// |onlyContents| applies to paths that are dirs and insists that only the contents of dir
// and not the base dir itself is copied.
func (this FileSpec) CopyTarget(baseDir string, destDir string, onlyContents bool) error {
    switch (this.GetType()) {
        case FILESPEC_TYPE_PATH:
            return this.copyPath(baseDir, destDir, onlyContents);
        case FILESPEC_TYPE_GIT:
            return this.copyGit(destDir);
        case FILESPEC_TYPE_NIL:
            // noop.
            return nil;
        default:
            return fmt.Errorf("Unknown filespec type ('%s'): '%v'.", this, this.GetType());
    }
}

func (this FileSpec) copyPath(baseDir string, destDir string, onlyContents bool) error {
    rawPath := string(this);

    sourcePath := rawPath;
    if (!filepath.IsAbs(sourcePath) && (baseDir != "")) {
        sourcePath = filepath.Join(baseDir, rawPath);
    }

    destPath := filepath.Join(destDir, filepath.Base(rawPath));

    var err error;

    if (onlyContents) {
        err = util.CopyDirContents(sourcePath, destPath);
    } else {
        err = util.CopyDirent(sourcePath, destPath, false);
    }

    if (err != nil) {
        return fmt.Errorf("Failed to copy path filespec '%s' to '%s': '%w'.", sourcePath, destPath, err);
    }

    return nil;
}

func (this FileSpec) ParseParts() []string {
    return strings.Split(string(this), FILESPEC_DELIM);
}

// Return: (url, dirname, ref, error).
func (this FileSpec) ParseGitParts() (string, string, string, error) {
    parts := this.ParseParts();

    if ((len(parts) < 2) || (len(parts) > 4) || (parts[0] != FILESPEC_GIT)) {
        return "", "", "", fmt.Errorf("Malformed git filespec: '%s'.", this);
    }

    url := parts[1]
    var dirname string;
    var ref string;

    if (len(parts) >= 3) {
        dirname = parts[2];
    } else {
        urlParts := strings.Split(url, "/")
        dirname = strings.TrimSuffix(urlParts[len(urlParts) - 1], ".git");
    }

    if (len(parts) >= 4) {
        ref = parts[3]
    }

    return url, dirname, ref, nil;
}

func (this FileSpec) copyGit(destDir string) error {
    url, dirname, ref, err := this.ParseGitParts();
    if (err != nil) {
        return err;
    }

    destPath := filepath.Join(destDir, dirname);

    _, err = util.GitEnsureRepo(url, destPath, true, ref);
    return err;
}
