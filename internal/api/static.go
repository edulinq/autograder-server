package api

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed static
var staticDir embed.FS

func handleStatic(response http.ResponseWriter, request *http.Request) error {
	// Remove leading and trailing slashes.
	path := strings.Trim(request.URL.Path, "/")

	file, err := staticDir.Open(path)
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

	// embed guarentees that embed.FS.Open() on a file returns an io.ReadSeeker.
	readseeker, _ := file.(io.ReadSeeker)
	http.ServeContent(response, request, path, stat.ModTime(), readseeker)

	return nil
}
