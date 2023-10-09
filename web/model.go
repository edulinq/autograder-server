package web

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

const API_REQUEST_CONTENT_KEY = "content";
const MAX_FORM_MEM_SIZE_BYTES = 10 << 20  // 20 MB

type APIResponse struct {
    Success bool `json:"success"`
    HTTPStatus int `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Content any `json:"content"`
}

type APIRequest interface {
    io.Closer

    // Called after desieralization to clean/normalize.
    Clean() error
}

type BaseAPIRequest struct {
    Course string `json:"course"`
    User string `json:"user"`
    Pass string `json:"pass"`
}

func NewResponse(status int, content any) *APIResponse {
    response := APIResponse{
        Success: (status == http.StatusOK),
        HTTPStatus: status,
        Timestamp: time.Now(),
        Content: content,
    };

    return &response;
}

// A failure, but still return a 200.
func NewSoftFailureResponse(content any) *APIResponse {
    response := APIResponse{
        Success: false,
        HTTPStatus: http.StatusOK,
        Timestamp: time.Now(),
        Content: content,
    };

    return &response;
}

func (this *APIResponse) Send(response http.ResponseWriter) error {
    payload, err := util.ToJSON(this);
    if (err != nil) {
        return fmt.Errorf("Could not serialize API response: '%w'.", err);
    }

    response.WriteHeader(this.HTTPStatus);
    _, err = fmt.Fprint(response, payload);
    if (err != nil) {
        return fmt.Errorf("Could not write API response payload: '%w'.", err);
    }

    return nil;
}

func (this *BaseAPIRequest) Clean() error {
    var err error;
    this.Course, err = common.ValidateID(this.Course);
    if (err != nil) {
        return fmt.Errorf("Could not clean course ID ('%s'): '%w'.", this.Course, err);
    }

    return nil;
}

// The basic deserialization of an API request from an HTTP request.
// All requests should do this first.
// The |apiRequest| should be a pointer that we will decode JSON into.
func APIRequestFromPOST(apiRequest APIRequest, request *http.Request) error {
    var err error;

    if (strings.Contains(strings.Join(request.Header["Content-Type"], " "), "multipart/form-data")) {
        err = request.ParseMultipartForm(MAX_FORM_MEM_SIZE_BYTES);
        if (err != nil) {
            return fmt.Errorf("Improperly formatted POST submission: '%w'.", err);
        }
    }

    textContent := request.PostFormValue(API_REQUEST_CONTENT_KEY);
    if (textContent == "") {
        return fmt.Errorf("No JSON payload.");
    }

    err = json.Unmarshal([]byte(textContent), apiRequest);
    if (err != nil) {
        return fmt.Errorf("Improperly formatted JSON payload: '%w'.", err);
    }

    err = apiRequest.Clean();
    if (err != nil) {
        return fmt.Errorf("Could not clean API request: '%w'.", err);
    }

    return nil;
}

// Pull files off the HTTP request and store them in a temp directory.
// On success, the caller owns the temp directory.
func StoreRequestFiles(request *http.Request) (string, error) {
    if (len(request.MultipartForm.File) == 0) {
        return "", nil;
    }

    tempDir, err := os.MkdirTemp("", "api-request-files-");
    if (err != nil) {
        return "", fmt.Errorf("Failed to create temp api files directory: '%w'.", err);
    }

    // Use an inner function to help control the removal of the temp dir on error.
    innerFunc := func() error {
        for filename, _ := range request.MultipartForm.File {
            err = storeRequestFile(request, tempDir, filename);
            if (err != nil) {
                return err;
            }
        }

        return nil;
    }

    err = innerFunc();
    if (err != nil) {
        os.RemoveAll(tempDir);
        return "", err;
    }

    return tempDir, nil;
}

func storeRequestFile(request *http.Request, outDir string, filename string) error {
    inFile, _, err := request.FormFile(filename);
    if (err != nil) {
        return fmt.Errorf("Failed to access request file '%s': '%w'.", filename, err);
    }
    defer inFile.Close();

    outPath := filepath.Join(outDir, filename);

    outFile, err := os.Create(outPath);
    if (err != nil) {
        return fmt.Errorf("Failed to create output file '%s': '%w'.", outPath, err);
    }
    defer outFile.Close();

    _, err = io.Copy(outFile, inFile);
    if (err != nil) {
        return fmt.Errorf("Failed to copy contents of request file '%s': '%w'.", filename, err);
    }

    if (strings.HasSuffix(outPath, ".zip")) {
        err = util.Unzip(outPath, outDir);
        if (err != nil) {
            return fmt.Errorf("Failed to extract zip archive ('%s'): '%w'.", outPath, err);
        }

        os.Remove(outPath);
        if (err != nil) {
            return fmt.Errorf("Failed to remove extracted zip archive ('%s'): '%w'.", outPath, err);
        }
    }

    return nil;
}
