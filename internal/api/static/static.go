package static

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/config"
)

//go:embed data
var staticDir embed.FS

// General function for opening our path objects.
// When called with a file (not dir) target, file must also implement io.Seeker.
type openFunc func(path string) (fs.File, error)

func Handle(response http.ResponseWriter, request *http.Request) error {
	// Remove leading and trailing slashes.
	path := strings.Trim(request.URL.Path, "/")

	root := config.WEB_STATIC_ROOT.Get()

	if root == "" {
		return serveEmbeddedDir(path, response, request)
	}

	return serveOutsideDir(path, root, response, request)
}

func serveEmbeddedDir(path string, response http.ResponseWriter, request *http.Request) error {
	open := func(path string) (fs.File, error) {
		return staticDir.Open(path)
	}

	return serve(path, response, request, open)
}

func serveOutsideDir(basepath string, root string, response http.ResponseWriter, request *http.Request) error {
	// Remove the static prefix or the common root path.
	// Trimming the static path should be fine, since they are both URL paths,
	// trimming the root is fragile and hacky at best.
	basepath = strings.TrimPrefix(basepath, "static/")
	basepath = strings.TrimPrefix("/"+basepath+"/", root)

	path := filepath.Join(root, basepath)

	open := func(path string) (fs.File, error) {
		return os.Open(path)
	}

	return serve(path, response, request, open)
}

func serve(path string, response http.ResponseWriter, request *http.Request, open openFunc) error {
	file, err := open(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(response, request)
			return nil
		}

		return fmt.Errorf("Error opening static file '%s': %w.", path, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Failed to stat static file (%s): %w.", path, err)
	}

	// Don't serve dirs. 301 to index.html.
	if stat.IsDir() {
		path = filepath.Join("/", path, "index.html")
		http.Redirect(response, request, path, 301)
		return nil
	}

	// Both os and embed guarentees that files opepened with os.Open() and embed.FS.Open() return an io.ReadSeeker.
	readseeker, _ := file.(io.ReadSeeker)
	http.ServeContent(response, request, path, stat.ModTime(), readseeker)

	return nil
}
