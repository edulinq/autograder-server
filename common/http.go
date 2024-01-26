package common

// Unless marked as "raw", all functions here assume that the respone body it textual.

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/util"
)

// A representation of an HTTP request.
type SavedHTTPRequest struct {
    URL string
    Method string
    RequestHeaders map[string][]string

    ResponseCode int
    ResponseHeaders map[string][]string
    ResponseBody string
}

// Get a binary response.
func RawGet(uri string) ([]byte, error) {
    response, err := http.Get(uri);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to perform GET on '%s': '%w'.", uri, err);
    }
    defer response.Body.Close();

    body, err := io.ReadAll(response.Body);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to read body from GET from '%s': '%w'.", uri, err);
    }

    return body, nil;
}

// Returns: (body, error)
func Get(uri string) (string, error) {
    body, _, err := GetWithHeaders(uri, make(map[string][]string));
    return body, err;
}

// Returns: (body, headers (response), error)
func GetWithHeaders(uri string, headers map[string][]string) (string, map[string][]string, error) {
    request, err := http.NewRequest("GET", uri, nil);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to create GET request on URL '%s': '%w'.", uri, err);
    }

    for key, values := range headers {
        for _, value := range values {
            request.Header.Add(key, value);
        }
    }

    return doRequest(uri, request, "GET", true);
}

// Returns: (body, error)
func Post(uri string, form map[string]string) (string, error) {
    body, _, err := PostWithHeaders(uri, form, make(map[string][]string));
    return body, err;
}

// Usually for testing.
// Returns: (body, error)
func PostNoCheck(uri string, form map[string]string) (string, error) {
    body, _, err := PostWithHeadersNoCheck(uri, form, make(map[string][]string));
    return body, err;
}

// Returns: (body, headers (response), error)
func PostWithHeaders(uri string, form map[string]string, headers map[string][]string) (string, map[string][]string, error) {
    return postPutWithHeaders("POST", uri, form, headers, true);
}

// Returns: (body, headers (response), error)
func PostWithHeadersNoCheck(uri string, form map[string]string, headers map[string][]string) (string, map[string][]string, error) {
    return postPutWithHeaders("POST", uri, form, headers, false);
}

// Returns: (body, error)
func Put(uri string, form map[string]string) (string, error) {
    body, _, err := PutWithHeaders(uri, form, make(map[string][]string));
    return body, err;
}

// Returns: (body, headers (response), error)
func PutWithHeaders(uri string, form map[string]string, headers map[string][]string) (string, map[string][]string, error) {
    return postPutWithHeaders("PUT", uri, form, headers, true);
}

func postPutWithHeaders(verb string, uri string, form map[string]string, headers map[string][]string, checkResult bool) (string, map[string][]string, error) {
    formValues := url.Values{};
    for key, value := range form {
        formValues.Set(key, value);
    }

    request, err := http.NewRequest(verb, uri, strings.NewReader(formValues.Encode()));
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to create %s request on URL '%s': '%w'.", verb, uri, err);
    }

    request.Header.Add("Content-Type", "application/x-www-form-urlencoded");

    for key, values := range headers {
        for _, value := range values {
            request.Header.Add(key, value);
        }
    }

    return doRequest(uri, request, verb, checkResult);
}

func PostFiles(uri string, form map[string]string, paths []string, checkResult bool) (string, error) {
    var buffer bytes.Buffer;

    // Create a new multipart writer with the buffer
    formWriter := multipart.NewWriter(&buffer);

    for key, value := range form {
        formWriter.WriteField(key, value);
    }

    for _, path := range paths {
        file, err := os.Open(path);
        if (err != nil) {
            return "", fmt.Errorf("Failed to open file '%s': '%w'.", path, err);
        }
        defer file.Close();

        filename := filepath.Base(path);

        fileWriter, err := formWriter.CreateFormFile(filename, filename);
        if (err != nil) {
            return "", fmt.Errorf("Failed to create form file for '%s': '%w'.", path, err);
        }

        _, err = io.Copy(fileWriter, file);
        if (err != nil) {
            return "", fmt.Errorf("Failed to copy file '%s' into form: '%w'.", path, err);
        }
    }

    formWriter.Close();

    request, err := http.NewRequest("POST", uri, &buffer);
    if (err != nil) {
        return "", fmt.Errorf("Failed to create POST request (with files) on URL '%s': '%w'.", uri, err);
    }

    request.Header.Add("Content-Type", formWriter.FormDataContentType());

    body, _, err := doRequest(uri, request, "POST", checkResult);
    return body, err;
}

// Returns: (body, headers (response), error)
func doRequest(uri string, request *http.Request, verb string, checkResult bool) (string, map[string][]string, error) {
    client := http.Client{}

    response, err := client.Do(request);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to perform %s request on URL '%s': '%w'.", verb, uri, err);
    }
    defer response.Body.Close();

    rawBody, err := io.ReadAll(response.Body);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to read body from %s on URL '%s': '%w'.", verb, uri, err);
    }
    body := string(rawBody);

    if (config.STORE_HTTP.Get() != "") {
        request := SavedHTTPRequest{
            URL: uri,
            Method: request.Method,
            RequestHeaders: request.Header,
            ResponseCode: response.StatusCode,
            ResponseHeaders: response.Header,
            ResponseBody: body,
        };

        err = writeRequest(&request);
        if (err != nil) {
            return "", nil, fmt.Errorf("Failed to save HTTP request '%s': '%w'.", uri, err);
        }
    }

    if (checkResult && (response.StatusCode != http.StatusOK)) {
        log.Error("Got a non-OK status.",
                log.NewAttr("code", response.StatusCode), log.NewAttr("body", body),
                log.NewAttr("headers", response.Header), log.NewAttr("url", uri));
        return "", nil, fmt.Errorf("Got a non-OK status code '%d' from %s on URL '%s': '%w'.", response.StatusCode, verb, uri, err);
    }

    return body, response.Header, nil;
}

func writeRequest(request *SavedHTTPRequest) error {
    baseDir := config.STORE_HTTP.Get();
    if (baseDir == "") {
        return fmt.Errorf("No base dir provided.");
    }

    err := util.MkDir(baseDir);
    if (err != nil) {
        return fmt.Errorf("Failed to create dir to store HTTP requests '%s': '%w'.", baseDir, err);
    }

    filename := fmt.Sprintf("%s.json", util.UUID());
    path := filepath.Join(baseDir, filename);

    err = util.ToJSONFileIndent(request, path);
    if (err != nil) {
        return fmt.Errorf("Failed to write JSON file '%s': '%w'.", path, err);
    }

    log.Debug("Saved HTTP request.", log.NewAttr("uri", request.URL), log.NewAttr("path", path));
    return nil;
}
