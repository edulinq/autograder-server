package util

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"

    "github.com/rs/zerolog/log"
)

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

    if (checkResult && (response.StatusCode != http.StatusOK)) {
        log.Error().Int("code", response.StatusCode).Str("body", body).Any("headers", response.Header).Str("url", uri).Msg("Got a non-OK status.");
        return "", nil, fmt.Errorf("Got a non-OK status code '%d' from %s on URL '%s': '%w'.", response.StatusCode, verb, uri, err);
    }

    return body, response.Header, nil;
}
