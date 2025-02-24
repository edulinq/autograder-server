package util

// Filespecs are specifications for file-like objects.
// They could be for a plain file/dir (just a path),
// or for something like a git repo.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/log"
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
	return this.ValidateFull(false)
}

func (this *FileSpec) String() string {
	if this == nil {
		return "<nil>"
	}

	value, err := ToJSON(this)
	if err != nil {
		log.Error("FileSpec could not marshal to JSON.", err)
		return "<error>"
	}

	return value
}

func (this *FileSpec) ValidateFull(onlyLocalPaths bool) error {
	if this == nil {
		return fmt.Errorf("File spec is nil.")
	}

	this.Type = FileSpecType(strings.ToLower(strings.TrimSpace(string(this.Type))))

	// Trim all fields.
	this.Path = strings.TrimSpace(this.Path)
	this.Dest = strings.TrimSpace(this.Dest)
	this.Reference = strings.TrimSpace(this.Reference)
	this.Username = strings.TrimSpace(this.Username)
	this.Token = strings.TrimSpace(this.Token)

	switch this.Type {
	case FILESPEC_TYPE_EMPTY, FILESPEC_TYPE_NIL:
		if (this.Path != "") || (this.Dest != "") || (this.Reference != "") || (this.Username != "") || (this.Token != "") {
			return fmt.Errorf("An empty/nil FileSpec should have no other fields set.")
		}
	case FILESPEC_TYPE_PATH:
		if (this.Reference != "") || (this.Username != "") || (this.Token != "") {
			return fmt.Errorf("An path FileSpec should not have reference, username, or token fields set.")
		}

		cleanPath, err := cleanFilePath(this.Path, false, onlyLocalPaths)
		if err != nil {
			return fmt.Errorf("Invalid path field for FileSpec: '%w'.", err)
		}

		this.Path = cleanPath

		_, err = filepath.Match(this.Path, "")
		if err != nil {
			return fmt.Errorf("Invalid path pattern '%s': '%w'.", this.Path, err)
		}

		cleanDest, err := cleanFilePath(this.Dest, true, onlyLocalPaths)
		if err != nil {
			return fmt.Errorf("Invalid dest field for FileSpec: '%w'.", err)
		}

		this.Dest = cleanDest
	case FILESPEC_TYPE_GIT:
		if this.Path == "" {
			return fmt.Errorf("A git FileSpec cannot have an empty path.")
		}

		cleanDest, err := cleanFilePath(this.Dest, true, onlyLocalPaths)
		if err != nil {
			return fmt.Errorf("Invalid dest field for FileSpec: '%w'.", err)
		}

		this.Dest = cleanDest

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

		cleanDest, err := cleanFilePath(this.Dest, true, onlyLocalPaths)
		if err != nil {
			return fmt.Errorf("Invalid dest field for FileSpec: '%w'.", err)
		}

		this.Dest = cleanDest

		if this.Dest == "" {
			this.Dest, err = getURLBaseName(this.Path, false)
			if err != nil {
				return fmt.Errorf("Failed to parse URL: '%w'.", err)
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
	err := JSONFromString(contents, &spec)
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
	return JoinIfNotAbs(this.Dest, baseDir)
}

// Copy the target of this FileSpec in the specified location.
// If the FileSpec has a dest, then that will be the name of the resultant dirent within destDir.
// If the filespec is a path, then copy all matching dirents.
// If the filespec is a git repo, then ensure it is cloned/updated.
// Empty and Nil FileSpecs are no-ops.
// |baseDir| provides the relative base.
func (this *FileSpec) CopyTarget(baseDir string, destDir string) error {
	switch this.Type {
	case FILESPEC_TYPE_EMPTY, FILESPEC_TYPE_NIL:
		// no-op.
		return nil
	case FILESPEC_TYPE_PATH:
		return this.copyPaths(baseDir, destDir)
	case FILESPEC_TYPE_GIT:
		return this.copyGit(destDir)
	case FILESPEC_TYPE_URL:
		return this.downloadURL(destDir)
	default:
		return fmt.Errorf("Unknown filespec type: '%s'.", this.Type)
	}
}

func (this *FileSpec) copyPaths(baseDir string, baseDestDir string) error {
	if !this.IsPath() {
		return fmt.Errorf("Cannot match targets: FileSpec must be a path.")
	}

	// Resolve relative paths.
	fileSpecPath := this.Path
	if !filepath.IsAbs(fileSpecPath) && (baseDir != "") {
		fileSpecPath = filepath.Join(baseDir, fileSpecPath)
	}

	// Resolve globs.
	paths, err := filepath.Glob(fileSpecPath)
	if err != nil {
		return fmt.Errorf("Failed to resolve the path pattern '%s': '%w'.", this.Path, err)
	}

	if len(paths) == 0 {
		return fmt.Errorf("No targets found for the path '%s'.", this.Path)
	}

	// If there are multiple paths, the dest cannot point to a file.
	destPath := this.GetDest(baseDestDir)
	if (len(paths) > 1) && (this.Dest != "") && IsFile(destPath) {
		return fmt.Errorf("Found multiple paths (via glob), but dest is a file. Dest must be a dir.")
	}

	// If there are multiple paths, make sure the dest dir already exists.
	if len(paths) > 1 {
		err := os.MkdirAll(destPath, 0755)
		if err != nil {
			return fmt.Errorf("Failed to make destination dir '%s': '%w'.", destPath, err)
		}
	}

	for _, path := range paths {
		// Note that CopyDirent() will handle when dest is a file or dir.
		err := CopyDirent(path, destPath, false)
		if err != nil {
			return fmt.Errorf("Failed to copy path '%s' to '%s': '%w'.", path, destPath, err)
		}
	}

	return nil
}

func (this *FileSpec) copyGit(destDir string) error {
	destPath := this.GetDest(destDir)

	if PathExists(destPath) {
		err := RemoveDirent(destPath)
		if err != nil {
			return fmt.Errorf("Failed to remove existing destination for git FileSpec ('%s'): '%w'.", destPath, err)
		}
	}

	err := MkDir(filepath.Dir(destPath))
	if err != nil {
		return fmt.Errorf("Failed to make dir for git FileSpec ('%s'): '%w'.", destPath, err)
	}

	_, err = GitEnsureRepo(this.Path, destPath, true, this.Reference, this.Username, this.Token)
	return err
}

func (this *FileSpec) downloadURL(destDir string) error {
	destPath := this.GetDest(destDir)

	if PathExists(destPath) {
		err := RemoveDirent(destPath)
		if err != nil {
			return fmt.Errorf("Failed to remove existing destination for URL FileSpec ('%s'): '%w'.", destPath, err)
		}
	}

	err := MkDir(filepath.Dir(destDir))
	if err != nil {
		return fmt.Errorf("Failed to make dir for URL FileSpec ('%s'): '%w'.", destPath, err)
	}

	content, err := RawGet(this.Path)
	if err != nil {
		return err
	}

	err = WriteBinaryFile(content, destPath)
	if err != nil {
		return fmt.Errorf("Failed to write output '%s': '%w'.", destPath, err)
	}

	return nil
}

// Copy over filespecs with ops.
// 1) Do pre-copy operations.
// 2) Copy.
// 3) Do post-copy operations.
func CopyFileSpecsWithOps(
	sourceDir string, destDir string, baseDir string, filespecs []*FileSpec,
	preOperations []*FileOperation, postOperations []*FileOperation) error {
	// Do pre ops.
	err := ExecFileOperations(preOperations, baseDir)
	if err != nil {
		return fmt.Errorf("Failed to do pre file operation: '%w'.", err)
	}

	// Copy files.
	for _, filespec := range filespecs {
		err = filespec.CopyTarget(sourceDir, destDir)
		if err != nil {
			return fmt.Errorf("Failed to handle FileSpec '%s': '%w'", filespec, err)
		}
	}

	// Do post ops.
	err = ExecFileOperations(postOperations, baseDir)
	if err != nil {
		return fmt.Errorf("Failed to do post file operation: '%w'.", err)
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

func cleanFilePath(rawPath string, allowEmpty bool, onlyLocalPaths bool) (string, error) {
	path := strings.TrimSpace(rawPath)

	if path == "" {
		if allowEmpty {
			return "", nil
		}

		return "", fmt.Errorf("File path cannot be empty.")
	}

	path = filepath.Clean(path)

	if onlyLocalPaths {
		if filepath.IsAbs(path) {
			return "", fmt.Errorf("File path '%s' is not allowed to be absolute.", rawPath)
		}

		if !filepath.IsLocal(path) {
			return "", fmt.Errorf("File path '%s' points outside of the its base directory.", rawPath)
		}
	}

	return path, nil
}
