package common

// Filespecs are specifications for file-like objects.
// They could be for a plain file/dir (just a path),
// or for something like a git repo.

import (
    "encoding/json"
    "fmt"
    "path/filepath"
    "strings"

    "github.com/eriq-augustine/autograder/util"
)

type FileSpecType string;

const (
    FILESPEC_TYPE_EMPTY FileSpecType = "empty"
    FILESPEC_TYPE_NIL = "nil"
    FILESPEC_TYPE_PATH = "path"
    FILESPEC_TYPE_GIT = "git"
)

type FileSpec struct {
    Type FileSpecType `json:"type"`
    Path string `json:"path,omitempty"`
    Dest string `json:"dest,omitempty"`
    Reference string `json:"reference,omitempty"`
    Username string `json:"username,omitempty"`
    Token string `json:"token,omitempty"`
}

func (this *FileSpec) Validate() error {
    this.Type = FileSpecType(strings.ToLower(strings.TrimSpace(string(this.Type))));

    switch (this.Type) {
        case FILESPEC_TYPE_EMPTY, FILESPEC_TYPE_NIL:
            if ((this.Path != "") || (this.Dest != "") || (this.Reference != "") || (this.Username != "") || (this.Token != "")) {
                return fmt.Errorf("An empty/nil FileSpec should have no other fields set.");
            }
        case FILESPEC_TYPE_PATH:
            if (this.Path == "") {
                return fmt.Errorf("A path FileSpec cannot have an empty path.");
            }
        case FILESPEC_TYPE_GIT:
            if (this.Path == "") {
                return fmt.Errorf("A git FileSpec cannot have an empty path.");
            }

            if (this.Dest == "") {
                urlParts := strings.Split(this.Path, "/")
                this.Dest = strings.TrimSuffix(urlParts[len(urlParts) - 1], ".git");
            }
        default:
            return fmt.Errorf("Unknown FileSpec type: '%s'.", this.Type);
    }

    return nil;
}

func (this *FileSpec) UnmarshalJSON(data []byte) error {
    rawText := strings.TrimSpace(string(data));

    if ((rawText == "") || (rawText == "null") || rawText == `""`) {
        this.Type = FILESPEC_TYPE_EMPTY;
        return nil;
    }

    // Check for a string (path FileSpec).
    if (strings.HasPrefix(rawText, `"`)) {
        var text string;
        err := json.Unmarshal(data, &text);
        if (err != nil) {
            return err;
        }

        this.Type = FILESPEC_TYPE_PATH;
        this.Path = strings.TrimSpace(text);
        return nil;
    }

    // If not a string, this should be an object.
    if (!strings.HasPrefix(rawText, "{")) {
        return fmt.Errorf("Could not deserialize FileSpec. Should be a JSON string or object, found '%s'.", rawText);
    }

    var values map[string]string;
    err := json.Unmarshal(data, &values);
    if (err != nil) {
        return err;
    }

    this.Type = FileSpecType(values["type"]);
    this.Path = values["path"];
    this.Dest = values["dest"];
    this.Reference = values["reference"];
    this.Username = values["username"];
    this.Token = values["token"];

    return nil;
}

// Parse a filespec from a string, but allow for strings to not be quoted.
// Returned FileSpec will be validated.
func ParseFileSpec(contents string) (*FileSpec, error) {
    contents = strings.TrimSpace(contents);
    if (!strings.HasPrefix(contents, `"`) && !strings.HasPrefix(contents, "{")) {
        contents = `"` + contents + `"`;
    }

    var spec FileSpec;
    err := util.JSONFromString(contents, &spec);
    if (err != nil) {
        return nil, err;
    }

    err = spec.Validate();
    if (err != nil) {
        return nil, err;
    }

    return &spec, nil;
}

func GetEmptyFileSpec() *FileSpec {
    return &FileSpec{Type: FILESPEC_TYPE_EMPTY};
}

func GetNilFileSpec() *FileSpec {
    return &FileSpec{Type: FILESPEC_TYPE_NIL};
}

func GetPathFileSpec(path string) *FileSpec {
    return &FileSpec{Type: FILESPEC_TYPE_PATH, Path: path};
}

func (this *FileSpec) IsEmpty() bool {
    if (this == nil) {
        return true;
    }

    return this.Type == FILESPEC_TYPE_EMPTY;
}

func (this *FileSpec) IsPath() bool {
    if (this == nil) {
        return false;
    }

    return this.Type == FILESPEC_TYPE_PATH;
}

func (this *FileSpec) IsGit() bool {
    if (this == nil) {
        return false;
    }

    return this.Type == FILESPEC_TYPE_GIT;
}

func (this *FileSpec) IsNil() bool {
    if (this == nil) {
        return false;
    }

    return this.Type == FILESPEC_TYPE_NIL;
}

func (this *FileSpec) IsAbs() bool {
    return this.IsPath() && filepath.IsAbs(this.Path);
}

func (this *FileSpec) GetPath() string {
    return this.Path;
}

// Copy the target of this FileSpec in the specified location.
// If the FileSpec has a dest, then that will be the name of the resultant dirent within destDir.
// If the filespec is a path, then copy a dirent.
// If the filespec is a git repo, then ensure it is cloned/updated.
// Empty and Nil FileSpecs are no-ops.
// |onlyContents| applies to paths that are dirs and insists that only the contents of dir
// and not the base dir itself is copied.
func (this *FileSpec) CopyTarget(baseDir string, destDir string, onlyContents bool) error {
    switch (this.Type) {
        case FILESPEC_TYPE_EMPTY, FILESPEC_TYPE_NIL:
            // no-op.
            return nil;
        case FILESPEC_TYPE_PATH:
            return this.copyPath(baseDir, destDir, onlyContents);
        case FILESPEC_TYPE_GIT:
            return this.copyGit(destDir);
        default:
            return fmt.Errorf("Unknown filespec type: '%s'.", this.Type);
    }
}

func (this *FileSpec) copyPath(baseDir string, destDir string, onlyContents bool) error {
    sourcePath := this.Path;
    if (!filepath.IsAbs(sourcePath) && (baseDir != "")) {
        sourcePath = filepath.Join(baseDir, this.Path);
    }

    destPath := "";
    if (onlyContents) {
        if (this.Dest == "") {
            destPath = destDir;
        } else {
            destPath = filepath.Join(destDir, this.Dest);
        }
    } else {
        filename := this.Dest;
        if (filename == "") {
            filename = filepath.Base(this.Path);
        }

        destPath = filepath.Join(destDir, filename);
    }

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

func (this *FileSpec) copyGit(destDir string) error {
    destPath := filepath.Join(destDir, this.Dest);

    if (util.PathExists(destPath)) {
        return fmt.Errorf("Destination for git FileSpec ('%s') already exists.", destPath);
    }

    err := util.MkDir(destDir);
    if (err != nil) {
        return fmt.Errorf("Failed to make dir for git FileSpec ('%s'): '%w'.", destDir, err);
    }

    _, err = util.GitEnsureRepo(this.Path, destPath, true, this.Reference, this.Username, this.Token);
    return err;
}
