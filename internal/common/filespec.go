package common

// Filespecs are specifications for file-like objects.
// They could be for a plain file/dir (just a path),
// or for something like a git repo.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

type FileSpecType string

const (
	FILESPEC_TYPE_EMPTY FileSpecType = "empty"
	FILESPEC_TYPE_NIL                = "nil"
	FILESPEC_TYPE_PATH               = "path"
	FILESPEC_TYPE_GIT                = "git"
	FILESPEC_TYPE_URL                = "url"
)

type FileSpec struct {
	Type      FileSpecType `json:"type"`
	Path      string       `json:"path,omitempty"`
	Dest      string       `json:"dest,omitempty"`
	Reference string       `json:"reference,omitempty"`
	Username  string       `json:"username,omitempty"`
	Token     string       `json:"token,omitempty"`
}

func (this *FileSpec) Validate() error {
	this.Type = FileSpecType(strings.ToLower(strings.TrimSpace(string(this.Type))))
	var err error

	switch this.Type {
	case FILESPEC_TYPE_EMPTY, FILESPEC_TYPE_NIL:
		if (this.Path != "") || (this.Dest != "") || (this.Reference != "") || (this.Username != "") || (this.Token != "") {
			return fmt.Errorf("An empty/nil FileSpec should have no other fields set.")
		}
	case FILESPEC_TYPE_PATH:
		if this.Path == "" {
			return fmt.Errorf("A path FileSpec cannot have an empty path.")
		}

		_, err := filepath.Match(this.Path, "")
		if err != nil {
			return fmt.Errorf("Invalid path pattern '%s': '%w'.", this.Path, err)
		}
	case FILESPEC_TYPE_GIT:
		if this.Path == "" {
			return fmt.Errorf("A git FileSpec cannot have an empty path.")
		}

		if this.Dest == "" {
			this.Dest, err = getURLBaseName(this.Path, true)
			if err != nil {
				return fmt.Errorf("Failed to parse git URL: '%w'.", err)
			}
		}
	case FILESPEC_TYPE_URL:
		if this.Path == "" {
			return fmt.Errorf("A url FileSpec cannot have an empty path.")
		}

		if this.Dest == "" {
			this.Dest, err = getURLBaseName(this.Path, false)
			if err != nil {
				return fmt.Errorf("Failed to parse git URL: '%w'.", err)
			}
		}
	default:
		return fmt.Errorf("Unknown FileSpec type: '%s'.", this.Type)
	}

	return nil
}

func (this *FileSpec) UnmarshalJSON(data []byte) error {
	rawText := strings.TrimSpace(string(data))

	if (rawText == "") || (rawText == "null") || rawText == `""` {
		this.Type = FILESPEC_TYPE_EMPTY
		return nil
	}

	// Check for a string (path or URL FileSpec).
	if strings.HasPrefix(rawText, `"`) {
		var text string
		err := json.Unmarshal(data, &text)
		if err != nil {
			return err
		}

		this.Path = strings.TrimSpace(text)

		if strings.HasPrefix(this.Path, "http") {
			this.Type = FILESPEC_TYPE_URL
		} else {
			this.Type = FILESPEC_TYPE_PATH
		}

		return nil
	}

	// If not a string, this should be an object.
	if !strings.HasPrefix(rawText, "{") {
		return fmt.Errorf("Could not deserialize FileSpec. Should be a JSON string or object, found '%s'.", rawText)
	}

	var values map[string]string
	err := json.Unmarshal(data, &values)
	if err != nil {
		return err
	}

	this.Type = FileSpecType(values["type"])
	this.Path = values["path"]
	this.Dest = values["dest"]
	this.Reference = values["reference"]
	this.Username = values["username"]
	this.Token = values["token"]

	return nil
}

// Parse a filespec from a string, but allow for strings to not be quoted.
// Returned FileSpec will be validated.
func ParseFileSpec(contents string) (*FileSpec, error) {
	contents = strings.TrimSpace(contents)
	if !strings.HasPrefix(contents, `"`) && !strings.HasPrefix(contents, "{") {
		contents = `"` + contents + `"`
	}

	var spec FileSpec
	err := util.JSONFromString(contents, &spec)
	if err != nil {
		return nil, err
	}

	err = spec.Validate()
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func GetEmptyFileSpec() *FileSpec {
	return &FileSpec{Type: FILESPEC_TYPE_EMPTY}
}

func GetNilFileSpec() *FileSpec {
	return &FileSpec{Type: FILESPEC_TYPE_NIL}
}

func GetPathFileSpec(content string) *FileSpec {
	return &FileSpec{Type: FILESPEC_TYPE_PATH, Path: content}
}

func (this *FileSpec) IsEmpty() bool {
	if this == nil {
		return true
	}

	return this.Type == FILESPEC_TYPE_EMPTY
}

func (this *FileSpec) IsPath() bool {
	if this == nil {
		return false
	}

	return this.Type == FILESPEC_TYPE_PATH
}

func (this *FileSpec) IsGit() bool {
	if this == nil {
		return false
	}

	return this.Type == FILESPEC_TYPE_GIT
}

func (this *FileSpec) IsURL() bool {
	if this == nil {
		return false
	}

	return this.Type == FILESPEC_TYPE_URL
}

func (this *FileSpec) IsNil() bool {
	if this == nil {
		return false
	}

	return this.Type == FILESPEC_TYPE_NIL
}

func (this *FileSpec) IsAbs() bool {
	return this.IsPath() && filepath.IsAbs(this.Path)
}

func (this *FileSpec) GetPath() string {
	return this.Path
}

// Get the path of this FileSpec's dest given the provided base dir.
// Note that specs may ignore the base dir if the current dest is absolute.
func (this *FileSpec) GetDest(baseDir string) string {
	if filepath.IsAbs(this.Dest) {
		return this.Dest
	}

	return filepath.Join(baseDir, this.Dest)
}

// Copy the target of this FileSpec in the specified location.
// If the FileSpec has a dest, then that will be the name of the resultant dirent within destDir.
// If the filespec is a path, then copy all matching dirents.
// If the filespec is a git repo, then ensure it is cloned/updated.
// Empty and Nil FileSpecs are no-ops.
// |onlyContents| applies to paths that are dirs and insists that only the contents of dir
// and not the base dir itself is copied.
// |baseDir| provides the relative base.
func (this *FileSpec) CopyTarget(baseDir string, destDir string, onlyContents bool) error {
	switch this.Type {
	case FILESPEC_TYPE_EMPTY, FILESPEC_TYPE_NIL:
		// no-op.
		return nil
	case FILESPEC_TYPE_PATH:
		return this.copyPaths(baseDir, destDir, onlyContents)
	case FILESPEC_TYPE_GIT:
		return this.copyGit(destDir)
	case FILESPEC_TYPE_URL:
		return this.downloadURL(destDir)
	default:
		return fmt.Errorf("Unknown filespec type: '%s'.", this.Type)
	}
}

func (this *FileSpec) copyPaths(baseDir string, destDir string, onlyContents bool) error {
	if !this.IsPath() {
		return fmt.Errorf("Cannot match targets: FileSpec must be a path.")
	}

	fileSpecPath := this.Path
	if !filepath.IsAbs(fileSpecPath) && (baseDir != "") {
		fileSpecPath = filepath.Join(baseDir, fileSpecPath)
	}

	destPath := ""
	if filepath.IsAbs(this.Dest) {
		destPath = this.Dest
	} else if onlyContents {
		if this.Dest == "" {
			destPath = destDir
		} else {
			destPath = filepath.Join(destDir, this.Dest)
		}
	} else {
		filename := this.Dest
		if filename == "" {
			filename = filepath.Base(this.Path)
		}

		destPath = filepath.Join(destDir, filename)
	}

	paths, err := filepath.Glob(fileSpecPath)
	if err != nil {
		return fmt.Errorf("Failed to resolve the path pattern '%s': '%w'.", this.Path, err)
	}

	if len(paths) == 0 {
		return fmt.Errorf("No targets found for the path '%s'.", this.Path)
	}

	// Ensure destPath is a directory if there are multiple paths.
	if len(paths) > 1 {
		if util.IsFile(destPath) {
			return fmt.Errorf("Cannot copy multiple targets into the existing file '%s'.", destDir)
		}

		if !util.PathExists(destPath) {
			err := util.MkDir(destPath)
			if err != nil {
				return fmt.Errorf("Failed to create a directory for the Filespec at path '%s': '%v'.", destPath, err)
			}
		}
	}

	// Loop over each matched path and copy it to the destination.
	for _, path := range paths {
		err := copyPath(path, destPath, onlyContents)
		if err != nil {
			return fmt.Errorf("Failed to copy target at path '%s': '%w'.", path, err)
		}
	}

	return nil
}

func copyPath(fileSpecPath string, destPath string, onlyContents bool) error {
	var err error
	if onlyContents {
		err = util.CopyDirContents(fileSpecPath, destPath)
	} else {
		err = util.CopyDirent(fileSpecPath, destPath, false)
	}

	if err != nil {
		return fmt.Errorf("Failed to copy path filespec '%s' to '%s': '%w'.", fileSpecPath, destPath, err)
	}

	return nil
}

func (this *FileSpec) copyGit(destDir string) error {
	destPath := this.GetDest(destDir)

	if util.PathExists(destPath) {
		err := util.RemoveDirent(destPath)
		if err != nil {
			return fmt.Errorf("Failed to remove existing destination for git FileSpec ('%s'): '%w'.", destPath, err)
		}
	}

	err := util.MkDir(filepath.Dir(destPath))
	if err != nil {
		return fmt.Errorf("Failed to make dir for git FileSpec ('%s'): '%w'.", destPath, err)
	}

	_, err = util.GitEnsureRepo(this.Path, destPath, true, this.Reference, this.Username, this.Token)
	return err
}

func (this *FileSpec) downloadURL(destDir string) error {
	destPath := this.GetDest(destDir)

	if util.PathExists(destPath) {
		err := util.RemoveDirent(destPath)
		if err != nil {
			return fmt.Errorf("Failed to remove existing destination for URL FileSpec ('%s'): '%w'.", destPath, err)
		}
	}

	err := util.MkDir(filepath.Dir(destDir))
	if err != nil {
		return fmt.Errorf("Failed to make dir for URL FileSpec ('%s'): '%w'.", destPath, err)
	}

	content, err := RawGet(this.Path)
	if err != nil {
		return err
	}

	err = util.WriteBinaryFile(content, destPath)
	if err != nil {
		return fmt.Errorf("Failed to write output '%s': '%w'.", destPath, err)
	}

	return nil
}

func getURLBaseName(uri string, removeExt bool) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("Failed to parse raw URL '%s': '%w'.", uri, err)
	}

	baseName := path.Base(parsedURL.Path)
	if baseName == "" {
		return "", fmt.Errorf("Could not find base name for URL: '%s'.", uri)
	}

	if removeExt {
		baseName = strings.TrimSuffix(baseName, path.Ext(baseName))
		if baseName == "" {
			return "", fmt.Errorf("Could not find base name for URL after removing extension: '%s'.", uri)
		}
	}

	return baseName, nil
}
